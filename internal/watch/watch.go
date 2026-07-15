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

	"github.com/Lokee86/doc-ledger/internal/config"
	"github.com/Lokee86/doc-ledger/internal/reconcile"
	"github.com/Lokee86/doc-ledger/internal/scan"
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
	if out == nil {
		out = io.Discard
	}
	seconds := c.Watch.DebounceSeconds
	if debounce != nil {
		seconds = *debounce
	}
	fmt.Fprintf(out, "%s doc-ledger watch watching %s pid=%d\n", timestamp(), root, os.Getpid())
	run := func() error {
		result, err := reconcile.Tree(root, c)
		if err != nil {
			return err
		}
		changed, err := reconcile.Apply(result)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s doc-ledger watch updated %d file(s)\n", timestamp(), changed)
		if len(result.Messages) > 0 {
			fmt.Fprintf(out, "%s doc-ledger watch reconciliation messages: %d\n", timestamp(), len(result.Messages))
		}
		return nil
	}
	if err := run(); err != nil {
		return err
	}
	if once {
		return nil
	}
	w, err := createWatcher()
	if err != nil {
		return fmt.Errorf("create watcher for %s: %w", root, err)
	}
	defer w.Close()
	watchedDirs := map[string]bool{}
	if err := addTree(w, root, c, watchedDirs); err != nil {
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
			if event.Op&fsnotify.Create != 0 {
				if st, err := os.Stat(event.Name); err == nil && st.IsDir() {
					if err := addTree(w, event.Name, c, watchedDirs); err != nil {
						return err
					}
				}
			}
			wasDirectory := watchedDirs[event.Name]
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 && wasDirectory {
				delete(watchedDirs, event.Name)
			}
			if wasDirectory || Relevant(event.Name, c, root) {
				scheduler.MarkChanged()
			}
		case <-ticker.C:
			if _, err := scheduler.RunIfPending(); err != nil {
				return err
			}
		}
	}
}
func addTree(w eventWatcher, root string, c config.Config, watched map[string]bool) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && ignored(path, c) {
			return filepath.SkipDir
		}
		if err := w.Add(path); err != nil {
			return fmt.Errorf("watch directory %s: %w", path, err)
		}
		watched[path] = true
		return nil
	})
}
func Relevant(path string, c config.Config, root string) bool {
	if ignored(path, c) {
		return false
	}
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		return true
	}
	ok, err := scan.IsIndexable(root, path, c)
	return err == nil && ok
}
func ignored(path string, c config.Config) bool {
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
