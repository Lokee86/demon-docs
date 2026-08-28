package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/fsnotify/fsnotify"
)

type benchmarkCompletionWriter struct {
	mu        sync.Mutex
	buffer    string
	completed chan struct{}
}

func newBenchmarkCompletionWriter() *benchmarkCompletionWriter {
	return &benchmarkCompletionWriter{completed: make(chan struct{}, 32)}
}

func (w *benchmarkCompletionWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer += string(p)
	for {
		newline := strings.IndexByte(w.buffer, '\n')
		if newline < 0 {
			break
		}
		line := w.buffer[:newline]
		w.buffer = w.buffer[newline+1:]
		if strings.Contains(line, "ddocs watch updated ") {
			select {
			case w.completed <- struct{}{}:
			default:
			}
		}
	}
	return len(p), nil
}

func BenchmarkScopedWatcherSingleFileCreate(b *testing.B) {
	benchmarkWatcherSingleFileCreate(b, false)
}

func BenchmarkFullWatcherSingleFileCreate(b *testing.B) {
	benchmarkWatcherSingleFileCreate(b, true)
}

func benchmarkWatcherSingleFileCreate(b *testing.B, full bool) {
	root := b.TempDir()
	const folderCount = 128
	const filesPerFolder = 4
	for folderIndex := 0; folderIndex < folderCount; folderIndex++ {
		folder := filepath.Join(root, fmt.Sprintf("group-%03d", folderIndex))
		if err := os.MkdirAll(folder, 0o755); err != nil {
			b.Fatal(err)
		}
		for fileIndex := 0; fileIndex < filesPerFolder; fileIndex++ {
			path := filepath.Join(folder, fmt.Sprintf("document-%02d.md", fileIndex))
			if err := os.WriteFile(path, []byte("# Document\n"), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}

	fake := newFakeWatcher()
	originalCreate := createWatcher
	originalRecursive := useRecursiveTreeWatches
	useRecursiveTreeWatches = false
	createWatcher = func() (eventWatcher, error) { return fake, nil }
	defer func() {
		createWatcher = originalCreate
		useRecursiveTreeWatches = originalRecursive
	}()

	cfg := config.Default()
	completion := newBenchmarkCompletionWriter()
	zero := 0.0
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ready := make(chan struct{})
	go func() {
		done <- RootSelectedWithRunLock(
			ctx,
			root,
			root,
			cfg,
			Features{Indexes: true, Links: true, TrackLinks: true},
			&zero,
			false,
			completion,
			nil,
			func() error { close(ready); return nil },
		)
	}()
	select {
	case <-ready:
	case <-time.After(10 * time.Second):
		cancel()
		b.Fatal("watcher did not become ready")
	}
	// The watcher schedules one post-registration full handoff pass. Let that
	// converge before timing event-to-completion latency, then drain startup
	// completion signals.
	time.Sleep(250 * time.Millisecond)
drainCompletion:
	for {
		select {
		case <-completion.completed:
			continue
		default:
			break drainCompletion
		}
	}

	folder := filepath.Join(root, "group-064")
	index := filepath.Join(folder, cfg.IndexFile)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		name := fmt.Sprintf("created-%04d.md", iteration)
		path := filepath.Join(folder, name)
		if err := os.WriteFile(path, []byte("# Created\n"), 0o644); err != nil {
			b.Fatal(err)
		}
		if full {
			fake.errors <- fsnotify.ErrEventOverflow
		} else {
			fake.events <- fsnotify.Event{Name: path, Op: fsnotify.Create}
		}
		select {
		case <-completion.completed:
		case <-time.After(5 * time.Second):
			b.Fatalf("watcher did not complete reconciliation for %s", name)
		}
		data, err := os.ReadFile(index)
		if err != nil || !strings.Contains(string(data), "["+name+"]("+name+")") {
			b.Fatalf("watcher completed without reconciling %s", name)
		}
	}
	b.StopTimer()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			b.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		b.Fatal("watcher did not stop")
	}
}
