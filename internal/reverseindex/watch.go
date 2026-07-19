package reverseindex

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/fsnotify/fsnotify"
)

func Watch(ctx context.Context, repositoryRoot, docsRoot string, c config.Config, format codemap.Format, debounce time.Duration, once bool, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}
	run := func() error {
		plan, err := Build(repositoryRoot, docsRoot, c, format)
		if err != nil {
			return err
		}
		changed, err := Apply(repositoryRoot, plan)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "ddocs reverse-index watch updated %d file(s), %d diagnostic(s)\n", changed, len(plan.Diagnostics))
		return nil
	}
	if err := run(); err != nil {
		return err
	}
	if once {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		return err
	}
	watched := map[string]struct{}{}
	addTree := func(start string) error {
		return filepath.WalkDir(start, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() {
				return nil
			}
			if path != repositoryRoot {
				if worktreeDirectory(entry.Name()) {
					return filepath.SkipDir
				}
				ignored, ignoreErr := policy.Ignored(path, true)
				if ignoreErr != nil {
					return ignoreErr
				}
				if ignored {
					return filepath.SkipDir
				}
			}
			if _, ok := watched[path]; ok {
				return nil
			}
			if err := watcher.Add(path); err != nil {
				return err
			}
			watched[path] = struct{}{}
			return nil
		})
	}
	if err := addTree(repositoryRoot); err != nil {
		return err
	}
	fmt.Fprintf(out, "ddocs reverse-index watch watching %s pid=%d\n", repositoryRoot, os.Getpid())
	var timer *time.Timer
	var timerChannel <-chan time.Time
	mark := func() {
		if timer == nil {
			timer = time.NewTimer(debounce)
			timerChannel = timer.C
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(debounce)
		timerChannel = timer.C
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-watcher.Errors:
			if err != nil {
				return err
			}
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = addTree(event.Name)
				}
			}
			mark()
		case <-timerChannel:
			timerChannel = nil
			if err := run(); err != nil {
				return err
			}
		}
	}
}
