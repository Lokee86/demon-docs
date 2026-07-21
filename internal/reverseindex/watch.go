package reverseindex

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/fsnotify/fsnotify"
)

func Watch(ctx context.Context, repositoryRoot, docsRoot string, roots []string, c config.Config, format codemap.Format, debounce time.Duration, once bool, out io.Writer) error {
	return WatchWithRunLock(ctx, repositoryRoot, docsRoot, roots, c, format, debounce, once, out, nil, nil)
}

func WatchWithRunLock(ctx context.Context, repositoryRoot, docsRoot string, roots []string, c config.Config, format codemap.Format, debounce time.Duration, once bool, out io.Writer, runLock sync.Locker, ready func() error) error {
	if out == nil {
		out = io.Discard
	}
	run := func() error {
		if runLock != nil {
			runLock.Lock()
			defer runLock.Unlock()
		}
		plan, err := Build(repositoryRoot, docsRoot, roots, c, format)
		if err != nil {
			return err
		}
		changed, err := Apply(repositoryRoot, plan)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "ddocs watch --reverse updated %d file(s), %d diagnostic(s)\n", changed, len(plan.Diagnostics))
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
	defer closeAndDrainWatcher(watcher)
	watched := map[string]struct{}{}
	refresh := func() error {
		_, folders, err := discoverScopeFolders(repositoryRoot, roots)
		if err != nil {
			return err
		}
		for _, directory := range ancestorDirectories(repositoryRoot, roots) {
			folders[directory] = struct{}{}
		}
		for _, directory := range sortedFolders(folders) {
			if _, ok := watched[directory]; ok {
				continue
			}
			if err := watcher.Add(directory); err != nil {
				return err
			}
			watched[directory] = struct{}{}
		}
		return nil
	}
	if err := refresh(); err != nil {
		return err
	}
	if ready != nil {
		if err := ready(); err != nil {
			return fmt.Errorf("mark reverse-index watcher ready: %w", err)
		}
	}
	refreshRequests := make(chan struct{}, 1)
	refreshResults := make(chan error, 1)
	workerContext, stopWorker := context.WithCancel(ctx)
	defer stopWorker()
	go func() {
		for {
			select {
			case <-workerContext.Done():
				return
			case <-refreshRequests:
				err := refresh()
				select {
				case refreshResults <- err:
				case <-workerContext.Done():
					return
				}
			}
		}
	}()
	requestRefresh := func() {
		select {
		case refreshRequests <- struct{}{}:
		default:
		}
	}
	fmt.Fprintf(out, "ddocs watch --reverse watching %d root(s) pid=%d\n", len(roots), os.Getpid())

	var timer *time.Timer
	var timerChannel <-chan time.Time
	mark := func() {
		if debounce < 0 {
			debounce = 0
		}
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
		case err := <-refreshResults:
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
		case err, ok := <-watcher.Errors:
			if ctx.Err() != nil || !ok {
				return nil
			}
			if err != nil {
				return err
			}
		case event, ok := <-watcher.Events:
			if ctx.Err() != nil || !ok {
				return nil
			}
			relevant := filepath.Base(event.Name) == ignorepolicy.FileName || insideAny(event.Name, roots)
			if !relevant {
				continue
			}
			if filepath.Base(event.Name) == ignorepolicy.FileName {
				requestRefresh()
			} else if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() && insideAny(event.Name, roots) {
					requestRefresh()
				}
			}
			mark()
		case <-timerChannel:
			timerChannel = nil
			if ctx.Err() != nil {
				return nil
			}
			if err := run(); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
			if ctx.Err() != nil {
				return nil
			}
			requestRefresh()
		}
	}
}

func closeAndDrainWatcher(watcher *fsnotify.Watcher) {
	done := make(chan struct{})
	go func() {
		_ = watcher.Close()
		close(done)
	}()
	events := watcher.Events
	errors := watcher.Errors
	for {
		select {
		case <-done:
			return
		case _, ok := <-events:
			if !ok {
				events = nil
			}
		case _, ok := <-errors:
			if !ok {
				errors = nil
			}
		}
	}
}
