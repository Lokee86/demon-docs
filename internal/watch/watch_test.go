package watch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

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
