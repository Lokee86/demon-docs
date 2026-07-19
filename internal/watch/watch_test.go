package watch

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

type blockingLocker struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newBlockingLocker() *blockingLocker {
	return &blockingLocker{entered: make(chan struct{}), release: make(chan struct{})}
}

func (l *blockingLocker) Lock() {
	l.once.Do(func() { close(l.entered) })
	<-l.release
}

func (l *blockingLocker) Unlock() {}

func TestSchedulerDebouncesAndRunsFollowup(t *testing.T) {
	now := time.Unix(0, 0)
	runs := 0
	var scheduler *Scheduler
	scheduler = NewScheduler(func() error {
		runs++
		if runs == 1 {
			scheduler.MarkChanged()
		}
		return nil
	}, 500*time.Millisecond)
	scheduler.now = func() time.Time { return now }
	scheduler.MarkChanged()
	now = now.Add(100 * time.Millisecond)
	if ran, _ := scheduler.RunIfPending(); ran {
		t.Fatal("ran before debounce")
	}
	now = now.Add(500 * time.Millisecond)
	if ran, err := scheduler.RunIfPending(); !ran || err != nil {
		t.Fatal("first run missing")
	}
	now = now.Add(500 * time.Millisecond)
	if ran, err := scheduler.RunIfPending(); !ran || err != nil {
		t.Fatal("follow-up run missing")
	}
	if runs != 2 {
		t.Fatalf("runs=%d", runs)
	}
}

func TestRootSelectedWithRunLockSerializesReconciliation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "page.md"), []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	locker := newBlockingLocker()
	done := make(chan error, 1)
	go func() {
		done <- RootSelectedWithRunLock(context.Background(), root, root, config.Default(), Features{Indexes: true}, nil, true, nil, locker)
	}()

	select {
	case <-locker.entered:
	case <-time.After(2 * time.Second):
		close(locker.release)
		t.Fatal("watch reconciliation did not acquire the shared run lock")
	}
	index := filepath.Join(root, "README.md")
	_, statErr := os.Stat(index)
	wroteBeforeRelease := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		close(locker.release)
		t.Fatal(statErr)
	}
	close(locker.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watch reconciliation did not finish after the run lock was released")
	}
	if wroteBeforeRelease {
		t.Fatal("watch reconciliation wrote files before the shared run lock was released")
	}
	if _, err := os.Stat(index); err != nil {
		t.Fatalf("watch reconciliation did not write the index after release: %v", err)
	}
}

func TestWatchConvergesWithoutSelfWriteLoop(t *testing.T) {
	root := t.TempDir()
	c := config.Default()
	c.Watch.DebounceSeconds = 0.05
	c.ParentLink.IndexedFiles = true
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Root(ctx, root, c, nil, false, nil) }()
	index := filepath.Join(root, "README.md")
	waitFor(t, 3*time.Second, func() bool { _, err := os.Stat(index); return err == nil })
	time.Sleep(150 * time.Millisecond)
	page := filepath.Join(root, "page.md")
	if err := os.WriteFile(page, []byte("# Page\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 3*time.Second, func() bool {
		data, err := os.ReadFile(index)
		return err == nil && strings.Contains(string(data), "[page.md](page.md)")
	})
	info, err := os.Stat(index)
	if err != nil {
		t.Fatal(err)
	}
	mtime := info.ModTime()
	time.Sleep(400 * time.Millisecond)
	info, err = os.Stat(index)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(mtime) {
		t.Fatal("watcher continued rewriting its own output")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not shut down")
	}
}

func TestWatchPrintsActualReconciliationMessages(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("[missing](missing.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RootSelected(context.Background(), root, root, config.Default(), Features{Links: true}, nil, true, &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(text, "Broken link in README.md") {
		t.Fatalf("watch output did not include the diagnostic: %q", text)
	}
	if strings.Contains(text, "reconciliation messages:") {
		t.Fatalf("watch output still used the opaque count: %q", text)
	}
}

func waitFor(t *testing.T, timeout time.Duration, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition did not become ready")
}
