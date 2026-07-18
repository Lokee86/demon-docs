package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/reconcile"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/scan"
	"github.com/fsnotify/fsnotify"
)

type Scheduler struct {
	mu       sync.Mutex
	run      func() error
	debounce time.Duration
	pending  int
	running  bool
	last     time.Time
	now      func() time.Time
}

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

func NewScheduler(run func() error, debounce time.Duration) *Scheduler {
	return &Scheduler{run: run, debounce: debounce, now: time.Now}
}
func (s *Scheduler) MarkChanged() { s.mu.Lock(); defer s.mu.Unlock(); s.pending++; s.last = s.now() }
func (s *Scheduler) RunIfPending() (bool, error) {
	s.mu.Lock()
	if s.running || s.pending == 0 || (s.debounce > 0 && s.now().Sub(s.last) < s.debounce) {
		s.mu.Unlock()
		return false, nil
	}
	s.running = true
	s.pending = 0
	s.mu.Unlock()
	err := s.run()
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	return true, err
}

func Root(ctx context.Context, root string, c config.Config, debounce *float64, once bool, out io.Writer) error {
	return RootWithIgnoreRoot(ctx, root, root, c, debounce, once, out)
}

func RootWithIgnoreRoot(ctx context.Context, root, ignoreRoot string, c config.Config, debounce *float64, once bool, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}
	seconds := c.Watch.DebounceSeconds
	if debounce != nil {
		seconds = *debounce
	}
	fmt.Fprintf(out, "%s ddocs watch watching %s pid=%d\n", timestamp(), root, os.Getpid())
	run := func() error {
		result, err := reconcile.TreeWithIgnoreRoot(root, ignoreRoot, c)
		if err != nil {
			return err
		}
		changed, err := reconcile.ApplyWithin(result, root)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s ddocs watch updated %d file(s)\n", timestamp(), changed)
		if len(result.Messages) > 0 {
			fmt.Fprintf(out, "%s ddocs watch reconciliation messages: %d\n", timestamp(), len(result.Messages))
		}
		return nil
	}
	if err := run(); err != nil {
		return err
	}
	if once {
		return nil
	}
	policy, err := ignorepolicy.Load(ignoreRoot)
	if err != nil {
		return err
	}
	w, err := createWatcher()
	if err != nil {
		return fmt.Errorf("create watcher for %s: %w", root, err)
	}
	defer w.Close()
	watchedDirs := map[string]bool{}
	if filepath.Clean(ignoreRoot) != filepath.Clean(root) {
		if err := w.Add(ignoreRoot); err != nil {
			return fmt.Errorf("watch repository root %s: %w", ignoreRoot, err)
		}
	}
	if err := addTree(w, root, root, c, policy, watchedDirs); err != nil {
		return err
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
				return fmt.Errorf("watch %s: %w", root, err)
			}
		case event, ok := <-w.Events():
			if !ok {
				return nil
			}
			if policy.IsControlFile(event.Name) {
				updated, err := ignorepolicy.Load(ignoreRoot)
				if err != nil {
					return err
				}
				policy = updated
				if err := addTree(w, root, root, c, policy, watchedDirs); err != nil {
					return err
				}
				scheduler.MarkChanged()
				continue
			}
			if event.Op&fsnotify.Create != 0 && repository.Contains(root, event.Name) {
				if st, err := os.Stat(event.Name); err == nil && st.IsDir() {
					ignored, err := policy.Ignored(event.Name, true)
					if err != nil {
						return err
					}
					if !ignored && !watchIgnored(event.Name, c) {
						if err := addTree(w, root, event.Name, c, policy, watchedDirs); err != nil {
							return err
						}
					}
				}
			}
			wasDirectory := watchedDirs[event.Name]
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 && wasDirectory {
				delete(watchedDirs, event.Name)
			}
			relevant := false
			if wasDirectory {
				ignored, err := policy.Ignored(event.Name, true)
				if err != nil {
					return err
				}
				relevant = !ignored && !watchIgnored(event.Name, c)
			} else {
				relevant = relevantWithPolicy(event.Name, c, policy, root)
			}
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
