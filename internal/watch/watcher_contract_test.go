package watch

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/fsnotify/fsnotify"
)

type fakeWatcher struct {
	events  chan fsnotify.Event
	errors  chan error
	mu      sync.Mutex
	added   []string
	removed []string
	failAdd func(string) error
}

func newFakeWatcher() *fakeWatcher {
	return &fakeWatcher{events: make(chan fsnotify.Event, 32), errors: make(chan error, 4)}
}
func (w *fakeWatcher) Add(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	path = filepath.Clean(path)
	w.added = append(w.added, path)
	if w.failAdd != nil {
		return w.failAdd(path)
	}
	return nil
}
func (w *fakeWatcher) Remove(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.removed = append(w.removed, filepath.Clean(path))
	return nil
}
func (w *fakeWatcher) Close() error                  { return nil }
func (w *fakeWatcher) Events() <-chan fsnotify.Event { return w.events }
func (w *fakeWatcher) Errors() <-chan error          { return w.errors }
func (w *fakeWatcher) setAddFailure(fail func(string) error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.failAdd = fail
}

func (w *fakeWatcher) hasWatch(path string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	path = filepath.Clean(path)
	for _, added := range w.added {
		if added == path {
			return true
		}
	}
	return false
}

func (w *fakeWatcher) removedWatch(path string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	path = filepath.Clean(path)
	for _, removed := range w.removed {
		if removed == path {
			return true
		}
	}
	return false
}

func installFakeWatcher(t *testing.T, fake *fakeWatcher, beforeCreate func()) {
	t.Helper()
	original := createWatcher
	originalRecursive := useRecursiveTreeWatches
	useRecursiveTreeWatches = false
	createWatcher = func() (eventWatcher, error) {
		if beforeCreate != nil {
			beforeCreate()
		}
		return fake, nil
	}
	t.Cleanup(func() {
		createWatcher = original
		useRecursiveTreeWatches = originalRecursive
	})
}

func startFakeWatch(t *testing.T, root string, c config.Config, debounce *float64, fake *fakeWatcher) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Root(ctx, root, c, debounce, false, nil) }()
	waitFor(t, 2*time.Second, func() bool { return fake.hasWatch(root) })
	return cancel, done
}

func stopFakeWatch(t *testing.T, cancel context.CancelFunc, done <-chan error) {
	t.Helper()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watch did not stop after cancellation")
	}
}

func TestInitialFixCompletesBeforeObserverCreation(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, func() {
		if _, err := os.Stat(filepath.Join(root, "INDEX.md")); err != nil {
			t.Fatalf("observer created before initial fix: %v", err)
		}
	})
	cancel, done := startFakeWatch(t, root, config.Default(), nil, fake)
	stopFakeWatch(t, cancel, done)
}

func TestWatcherRecoversFromEventOverflowWithFullReconciliation(t *testing.T) {
	root := t.TempDir()
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	zero := 0.0
	cancel, done := startFakeWatch(t, root, config.Default(), &zero, fake)

	page := filepath.Join(root, "overflow-created.md")
	if err := os.WriteFile(page, []byte("# Overflow Created\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake.errors <- fsnotify.ErrEventOverflow
	waitFor(t, 3*time.Second, func() bool {
		data, err := os.ReadFile(filepath.Join(root, "INDEX.md"))
		return err == nil && strings.Contains(string(data), "[overflow-created.md](overflow-created.md)")
	})
	select {
	case err := <-done:
		t.Fatalf("overflow terminated watcher: %v", err)
	default:
	}
	stopFakeWatch(t, cancel, done)
}

func TestWatcherReportsObserverErrors(t *testing.T) {
	root := t.TempDir()
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	_, done := startFakeWatch(t, root, config.Default(), nil, fake)
	fake.errors <- errors.New("observer failed")
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "observer failed") || !strings.Contains(err.Error(), "watch ") {
			t.Fatalf("unexpected observer error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("observer error was not returned")
	}
}

func TestWatcherHandlesFileRenameSourceAndDestination(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "old.md")
	newPath := filepath.Join(root, "new.md")
	if err := os.WriteFile(oldPath, []byte("# Old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	zero := 0.0
	cancel, done := startFakeWatch(t, root, config.Default(), &zero, fake)
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: oldPath, Op: fsnotify.Rename}
	fake.events <- fsnotify.Event{Name: newPath, Op: fsnotify.Create}
	index := filepath.Join(root, "INDEX.md")
	waitFor(t, 3*time.Second, func() bool {
		data, err := os.ReadFile(index)
		return err == nil && strings.Contains(string(data), "[new.md](new.md)") && !strings.Contains(string(data), "[old.md](old.md)")
	})
	stopFakeWatch(t, cancel, done)
}

func TestWatcherAttemptsObservedRenameBeforeDebouncedValidation(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "old.md")
	newPath := filepath.Join(root, "new.md")
	if err := os.WriteFile(oldPath, []byte("# Old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	originalRepair := repairObservedRename
	called := make(chan [2]string, 1)
	repairObservedRename = func(_ string, oldName, newName string) (bool, int, error) {
		called <- [2]string{filepath.Clean(oldName), filepath.Clean(newName)}
		return true, 1, nil
	}
	t.Cleanup(func() { repairObservedRename = originalRepair })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	longDebounce := 10.0
	go func() {
		done <- RootSelected(ctx, root, root, config.Default(), Features{Links: true}, &longDebounce, false, nil)
	}()
	waitFor(t, 2*time.Second, func() bool { return fake.hasWatch(root) })
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: oldPath, Op: fsnotify.Rename}
	fake.events <- fsnotify.Event{Name: newPath, Op: fsnotify.Create}
	select {
	case pair := <-called:
		if pair != [2]string{filepath.Clean(oldPath), filepath.Clean(newPath)} {
			t.Fatalf("observed rename=%v", pair)
		}
	case <-time.After(time.Second):
		t.Fatal("observed rename repair waited for the debounce interval")
	}
	stopFakeWatch(t, cancel, done)
}

func TestWatcherLimitsImmediateRenameRepairDuringBurst(t *testing.T) {
	root := t.TempDir()
	oldPaths := []string{
		filepath.Join(root, "old-one.md"),
		filepath.Join(root, "old-two.md"),
		filepath.Join(root, "old-three.md"),
	}
	newPaths := []string{
		filepath.Join(root, "new-one.md"),
		filepath.Join(root, "new-two.md"),
		filepath.Join(root, "new-three.md"),
	}
	for _, path := range oldPaths {
		if err := os.WriteFile(path, []byte("# Page\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	originalRepair := repairObservedRename
	called := make(chan [2]string, len(oldPaths))
	repairObservedRename = func(_ string, oldName, newName string) (bool, int, error) {
		called <- [2]string{filepath.Clean(oldName), filepath.Clean(newName)}
		return true, 1, nil
	}
	t.Cleanup(func() { repairObservedRename = originalRepair })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	longDebounce := 10.0
	go func() {
		done <- RootSelected(ctx, root, root, config.Default(), Features{Links: true}, &longDebounce, false, nil)
	}()
	waitFor(t, 2*time.Second, func() bool { return fake.hasWatch(root) })
	for index := range oldPaths {
		if err := os.Rename(oldPaths[index], newPaths[index]); err != nil {
			t.Fatal(err)
		}
		fake.events <- fsnotify.Event{Name: oldPaths[index], Op: fsnotify.Rename}
		fake.events <- fsnotify.Event{Name: newPaths[index], Op: fsnotify.Create}
	}
	select {
	case pair := <-called:
		want := [2]string{filepath.Clean(oldPaths[0]), filepath.Clean(newPaths[0])}
		if pair != want {
			t.Fatalf("first observed rename=%v want=%v", pair, want)
		}
	case <-time.After(time.Second):
		t.Fatal("first observed rename was not repaired immediately")
	}
	time.Sleep(100 * time.Millisecond)
	select {
	case pair := <-called:
		t.Fatalf("rename burst used another synchronous repair: %v", pair)
	default:
	}
	stopFakeWatch(t, cancel, done)
}

func TestWatcherAddsNewNestedDirectories(t *testing.T) {
	root := t.TempDir()
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	zero := 0.0
	cancel, done := startFakeWatch(t, root, config.Default(), &zero, fake)
	nested := filepath.Join(root, "new", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: filepath.Join(root, "new"), Op: fsnotify.Create}
	waitFor(t, 2*time.Second, func() bool { return fake.hasWatch(nested) })
	topic := filepath.Join(nested, "topic.md")
	if err := os.WriteFile(topic, []byte("# Topic\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: topic, Op: fsnotify.Create}
	waitFor(t, 3*time.Second, func() bool {
		data, err := os.ReadFile(filepath.Join(nested, "INDEX.md"))
		return err == nil && strings.Contains(string(data), "[topic.md](topic.md)")
	})
	stopFakeWatch(t, cancel, done)
}

func TestWatcherReloadsDocignoreAndAddsNewlyVisibleDirectories(t *testing.T) {
	root := t.TempDir()
	ignoredDir := filepath.Join(root, "ignored")
	if err := os.MkdirAll(ignoredDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignoredDir, "topic.md"), []byte("# Topic\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ignorePath := filepath.Join(root, ".docignore")
	if err := os.WriteFile(ignorePath, []byte("ignored/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	zero := 0.0
	cancel, done := startFakeWatch(t, root, config.Default(), &zero, fake)
	if fake.hasWatch(ignoredDir) {
		t.Fatal("ignored directory was watched before .docignore changed")
	}

	if err := os.WriteFile(ignorePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: ignorePath, Op: fsnotify.Write}
	waitFor(t, 2*time.Second, func() bool { return fake.hasWatch(ignoredDir) })
	waitFor(t, 3*time.Second, func() bool {
		data, err := os.ReadFile(filepath.Join(root, "INDEX.md"))
		return err == nil && strings.Contains(string(data), "ignored/INDEX.md")
	})
	stopFakeWatch(t, cancel, done)
}

func TestWatcherReturnsDynamicDirectoryAddErrors(t *testing.T) {
	root := t.TempDir()
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	_, done := startFakeWatch(t, root, config.Default(), nil, fake)

	nested := filepath.Join(root, "new", "nested")
	fake.setAddFailure(func(path string) error {
		if path == filepath.Clean(nested) {
			return errors.New("add nested watch failed")
		}
		return nil
	})
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: filepath.Join(root, "new"), Op: fsnotify.Create}

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "watch directory "+nested) || !strings.Contains(err.Error(), "add nested watch failed") {
			t.Fatalf("unexpected dynamic add error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dynamic directory add error was not returned")
	}
}

func TestWatcherReconcilesDeletedWatchedDirectory(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "guide")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	zero := 0.0
	cancel, done := startFakeWatch(t, root, config.Default(), &zero, fake)
	if err := os.RemoveAll(child); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: child, Op: fsnotify.Remove}
	index := filepath.Join(root, "INDEX.md")
	waitFor(t, 3*time.Second, func() bool {
		data, err := os.ReadFile(index)
		return err == nil && !bytes.Contains(data, []byte("guide/INDEX.md"))
	})
	stopFakeWatch(t, cancel, done)
}

func TestWatcherRemovesDescendantWatchesWhenDirectoryMoves(t *testing.T) {
	root := t.TempDir()
	oldTree := filepath.Join(root, "old")
	nested := filepath.Join(oldTree, "nested", "deep")
	newTree := filepath.Join(root, "new")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	longDebounce := 10.0
	cancel, done := startFakeWatch(t, root, config.Default(), &longDebounce, fake)
	waitFor(t, 2*time.Second, func() bool { return fake.hasWatch(nested) })

	if err := os.Rename(oldTree, newTree); err != nil {
		t.Fatal(err)
	}
	fake.events <- fsnotify.Event{Name: oldTree, Op: fsnotify.Rename}
	fake.events <- fsnotify.Event{Name: newTree, Op: fsnotify.Create}

	waitFor(t, 2*time.Second, func() bool {
		return fake.removedWatch(oldTree) &&
			fake.removedWatch(filepath.Join(oldTree, "nested")) &&
			fake.removedWatch(nested)
	})
	waitFor(t, 2*time.Second, func() bool {
		return fake.hasWatch(filepath.Join(newTree, "nested", "deep"))
	})
	stopFakeWatch(t, cancel, done)
}

func TestExplicitDebounceOverrideWins(t *testing.T) {
	root := t.TempDir()
	fake := newFakeWatcher()
	installFakeWatcher(t, fake, nil)
	c := config.Default()
	c.Watch.DebounceSeconds = 10
	zero := 0.0
	cancel, done := startFakeWatch(t, root, c, &zero, fake)
	page := filepath.Join(root, "fast.md")
	if err := os.WriteFile(page, []byte("# Fast\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	fake.events <- fsnotify.Event{Name: page, Op: fsnotify.Create}
	waitFor(t, time.Second, func() bool {
		data, err := os.ReadFile(filepath.Join(root, "INDEX.md"))
		return err == nil && strings.Contains(string(data), "[fast.md](fast.md)")
	})
	if time.Since(started) >= time.Second {
		t.Fatal("configured debounce was used instead of explicit override")
	}
	stopFakeWatch(t, cancel, done)
}
