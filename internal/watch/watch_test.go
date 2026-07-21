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

func TestInitialReconciliationRetriesTransientFilesystemRaces(t *testing.T) {
	attempts := 0
	var out bytes.Buffer
	err := runInitialUntilStable(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return errors.New("rewrite source changed before apply during move")
		}
		return nil
	}, time.Millisecond, &out)
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d", attempts)
	}
	if !strings.Contains(out.String(), "deferred stale initial reconciliation plan") {
		t.Fatalf("retry was not logged: %q", out.String())
	}
}

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
		done <- RootSelectedWithRunLock(context.Background(), root, root, config.Default(), Features{Indexes: true}, nil, true, nil, locker, nil)
	}()

	select {
	case <-locker.entered:
	case <-time.After(2 * time.Second):
		close(locker.release)
		t.Fatal("watch reconciliation did not acquire the shared run lock")
	}
	index := filepath.Join(root, "INDEX.md")
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
	index := filepath.Join(root, "INDEX.md")
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

func TestSelectedWatchRepairsLinksBeforeFrontmatterAndRefreshesFinalFingerprints(t *testing.T) {
	root := t.TempDir()
	writeWatchFile := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeWatchFile(filepath.Join(root, "docs", "source.md"), "[target](old/target.md)\n")
	writeWatchFile(filepath.Join(root, "docs", "old", "target.md"), "# Original target\n")
	writeWatchFile(filepath.Join(root, "docs", "decoy", "target.md"), "# Decoy target\n")

	c := config.Default()
	c.Root = "docs"
	c.Index.Enabled = true
	c.Links.Enabled = true
	c.Format.Enabled = false
	c.Frontmatter = config.Frontmatter{
		Enabled:        true,
		DefaultFormat:  "yaml",
		AllowedFormats: []string{"yaml"},
		UnknownFields:  "remove",
		Fields: map[string]config.FrontmatterField{
			"created": {Type: "date", Required: true, Immutable: true, Generated: true},
		},
	}
	docs := filepath.Join(root, "docs")
	if err := RootSelected(context.Background(), docs, root, c, Features{Links: true, TrackLinks: true}, nil, true, nil); err != nil {
		t.Fatal(err)
	}
	moved := filepath.Join(docs, "moved", "target.md")
	if err := os.MkdirAll(filepath.Dir(moved), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(docs, "old", "target.md"), moved); err != nil {
		t.Fatal(err)
	}
	features := Features{Indexes: true, Frontmatter: true, Links: true, TrackLinks: true}
	if err := RootSelected(context.Background(), docs, root, c, features, nil, true, nil); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(docs, "source.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "(moved/target.md)") {
		t.Fatalf("watch did not repair the link before frontmatter changed the target fingerprint:\n%s", source)
	}

	final := filepath.Join(docs, "final", "target.md")
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(moved, final); err != nil {
		t.Fatal(err)
	}
	if err := RootSelected(context.Background(), docs, root, c, Features{Links: true, TrackLinks: true}, nil, true, nil); err != nil {
		t.Fatal(err)
	}
	source, err = os.ReadFile(filepath.Join(docs, "source.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "(final/target.md)") {
		t.Fatalf("watch did not refresh final fingerprint state after frontmatter/index writes:\n%s", source)
	}
}

func TestWatchPrintsActualReconciliationMessages(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "INDEX.md"), []byte("[missing](missing.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := RootSelected(context.Background(), root, root, config.Default(), Features{Links: true}, nil, true, &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.Contains(text, "Broken link in INDEX.md") {
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
