package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/reconcile"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/scan"
	"github.com/fsnotify/fsnotify"
)

type eventWatcher interface {
	Add(string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

type fsnotifyWatcher struct{ *fsnotify.Watcher }

func (w fsnotifyWatcher) Events() <-chan fsnotify.Event { return w.Watcher.Events }
func (w fsnotifyWatcher) Errors() <-chan error          { return w.Watcher.Errors }

var createWatcher = func() (eventWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return fsnotifyWatcher{w}, nil
}

func Root(ctx context.Context, root string, c config.Config, debounce *float64, once bool, out io.Writer) error {
	return RootWithIgnoreRoot(ctx, root, root, c, debounce, once, out)
}

func RootWithIgnoreRoot(ctx context.Context, root, ignoreRoot string, c config.Config, debounce *float64, once bool, out io.Writer) error {
	return RootSelected(ctx, root, ignoreRoot, c, Features{Indexes: true}, debounce, once, out)
}

func RootSelected(ctx context.Context, docsRoot, repositoryRoot string, c config.Config, features Features, debounce *float64, once bool, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}
	if !features.Indexes && !features.Links {
		features = Features{Indexes: true, Links: true}
	}
	seconds := c.Watch.DebounceSeconds
	if debounce != nil {
		seconds = *debounce
	}
	watchRoot := docsRoot
	if features.Links {
		watchRoot = repositoryRoot
	}
	fmt.Fprintf(out, "%s ddocs watch watching %s pid=%d\n", timestamp(), watchRoot, os.Getpid())
	var watcher eventWatcher
	externalWatched := map[string]bool{}
	var externalDirectories []string
	run := func() error {
		changed := 0
		messages := 0
		unresolved := 0
		if features.Indexes {
			result, err := reconcile.TreeWithIgnoreRoot(docsRoot, repositoryRoot, c)
			if err != nil {
				return err
			}
			count, err := reconcile.ApplyWithin(result, docsRoot)
			if err != nil {
				return err
			}
			changed += count
			messages += len(result.Messages)
		}
		if features.Links {
			plan, err := links.Reconcile(repositoryRoot)
			if err != nil {
				return err
			}
			count, err := links.ApplyAndSave(&plan)
			if err != nil {
				return err
			}
			changed += count
			messages += len(plan.Messages)
			unresolved = plan.Unresolved
			externalDirectories = externalWatchDirectories(plan.Files)
			if watcher != nil {
				if err := addExternalWatches(watcher, externalDirectories, externalWatched); err != nil {
					return fmt.Errorf("watch external link targets: %w", err)
				}
			}
		}
		fmt.Fprintf(out, "%s ddocs watch updated %d file(s)\n", timestamp(), changed)
		if messages > 0 {
			fmt.Fprintf(out, "%s ddocs watch reconciliation messages: %d\n", timestamp(), messages)
		}
		if unresolved > 0 {
			fmt.Fprintf(out, "%s ddocs watch unresolved links: %d\n", timestamp(), unresolved)
		}
		return nil
	}
	if err := run(); err != nil {
		return err
	}
	if once {
		return nil
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		return err
	}
	w, err := createWatcher()
	if err != nil {
		return fmt.Errorf("create watcher for %s: %w", watchRoot, err)
	}
	watcher = w
	defer w.Close()
	watchedDirs := map[string]bool{}
	if filepath.Clean(repositoryRoot) != filepath.Clean(watchRoot) {
		if err := w.Add(repositoryRoot); err != nil {
			return fmt.Errorf("watch repository root %s: %w", repositoryRoot, err)
		}
	}
	if err := addTree(w, watchRoot, watchRoot, c, policy, watchedDirs); err != nil {
		return err
	}
	if err := addExternalWatches(w, externalDirectories, externalWatched); err != nil {
		return fmt.Errorf("watch external link targets: %w", err)
	}
	scheduler := NewScheduler(run, time.Duration(seconds*float64(time.Second)))
	interval := 100 * time.Millisecond
	if seconds > 0 {
		interval = time.Duration(seconds * float64(time.Second) / 2)
		if interval > 250*time.Millisecond {
			interval = 250 * time.Millisecond
		}
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-w.Errors():
			if !ok {
				return nil
			}
			if err != nil {
				return fmt.Errorf("watch %s: %w", watchRoot, err)
			}
		case event, ok := <-w.Events():
			if !ok {
				return nil
			}
			if features.Links && repository.Contains(repositoryRoot, event.Name) {
				suppressed, err := links.ConsumePendingSuppression(repositoryRoot, event.Name)
				if err != nil {
					return fmt.Errorf("consume generated link rewrite event: %w", err)
				}
				if suppressed {
					continue
				}
			}
			isExternal := features.Links && externalEvent(event.Name, externalWatched)
			if policy.IsControlFile(event.Name) {
				updated, err := ignorepolicy.Load(repositoryRoot)
				if err != nil {
					return err
				}
				policy = updated
				if err := addTree(w, watchRoot, watchRoot, c, policy, watchedDirs); err != nil {
					return err
				}
				scheduler.MarkChanged()
				continue
			}
			if event.Op&fsnotify.Create != 0 && repository.Contains(watchRoot, event.Name) {
				if st, err := os.Stat(event.Name); err == nil && st.IsDir() {
					ignored, err := policy.Ignored(event.Name, true)
					if err != nil {
						return err
					}
					if !ignored && !watchIgnored(event.Name, c) {
						if err := addTree(w, watchRoot, event.Name, c, policy, watchedDirs); err != nil {
							return err
						}
					}
				}
			}
			wasDirectory := watchedDirs[event.Name]
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
				if wasDirectory {
					delete(watchedDirs, event.Name)
				}
				if externalWatched[event.Name] {
					delete(externalWatched, event.Name)
				}
			}
			relevant := isExternal || relevantSelectedWithPolicy(event.Name, c, policy, docsRoot, repositoryRoot, features, wasDirectory)
			if relevant {
				scheduler.MarkChanged()
			}
		case <-ticker.C:
			if _, err := scheduler.RunIfPending(); err != nil {
				return err
			}
		}
	}
}

func addTree(w eventWatcher, root, start string, c config.Config, policy ignorepolicy.Policy, watched map[string]bool) error {
	return filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root {
			ignored, err := policy.Ignored(path, true)
			if err != nil {
				return err
			}
			if ignored || watchIgnored(path, c) {
				return filepath.SkipDir
			}
		}
		if watched[path] {
			return nil
		}
		if err := w.Add(path); err != nil {
			return fmt.Errorf("watch directory %s: %w", path, err)
		}
		watched[path] = true
		return nil
	})
}

func Relevant(path string, c config.Config, root string) bool {
	return RelevantWithIgnoreRoot(path, c, root, root)
}

func RelevantWithIgnoreRoot(path string, c config.Config, root, ignoreRoot string) bool {
	policy, err := ignorepolicy.Load(ignoreRoot)
	return err == nil && relevantWithPolicy(path, c, policy, root)
}

func relevantWithPolicy(path string, c config.Config, policy ignorepolicy.Policy, root string) bool {
	if policy.IsControlFile(path) {
		return true
	}
	if !repository.Contains(root, path) {
		return false
	}
	ignored, err := policy.Ignored(path, false)
	if err != nil || ignored || watchIgnored(path, c) {
		return false
	}
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		ignored, err := policy.Ignored(path, true)
		return err == nil && !ignored
	}
	ok, err := scan.IsIndexable(root, path, c)
	return err == nil && ok
}

func watchIgnored(path string, c config.Config) bool {
	clean := filepath.Clean(path)
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		for _, name := range c.Watch.IgnoredDirs {
			if part == name {
				return true
			}
		}
	}
	name := filepath.Base(clean)
	if strings.HasPrefix(name, ".#") {
		return true
	}
	for _, suffix := range c.Watch.IgnoredSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func timestamp() string { return time.Now().Format("2006-01-02T15:04:05") }
