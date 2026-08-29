package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/fsnotify/fsnotify"
)

func TestWatcherConvergesAfterLargeRenameEventBurst(t *testing.T) {
	const fileCount = 128
	root := t.TempDir()
	oldPaths := make([]string, fileCount)
	newPaths := make([]string, fileCount)
	var source strings.Builder
	source.WriteString("# Source\n\n")
	for index := 0; index < fileCount; index++ {
		oldName := fmt.Sprintf("old-%03d.md", index)
		newName := fmt.Sprintf("new-%03d.md", index)
		oldPaths[index] = filepath.Join(root, oldName)
		newPaths[index] = filepath.Join(root, newName)
		if err := os.WriteFile(oldPaths[index], []byte(fmt.Sprintf("# Target %03d\n", index)), 0o644); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&source, "- [target %03d](%s)\n", index, oldName)
	}
	sourcePath := filepath.Join(root, "source.md")
	if err := os.WriteFile(sourcePath, []byte(source.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	fake := newFakeWatcher()
	fake.events = make(chan fsnotify.Event, fileCount*2+32)
	installFakeWatcher(t, fake, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	zero := 0.0
	go func() {
		done <- RootSelected(ctx, root, root, config.Default(), Features{Indexes: true, Links: true}, &zero, false, nil)
	}()
	waitFor(t, 3*time.Second, func() bool { return fake.hasWatch(root) })

	for index := 0; index < fileCount; index++ {
		if err := os.Rename(oldPaths[index], newPaths[index]); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range oldPaths {
		fake.events <- fsnotify.Event{Name: path, Op: fsnotify.Rename}
	}
	for _, path := range newPaths {
		fake.events <- fsnotify.Event{Name: path, Op: fsnotify.Create}
	}

	waitFor(t, 20*time.Second, func() bool {
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			return false
		}
		text := string(data)
		for index := 0; index < fileCount; index++ {
			if strings.Contains(text, fmt.Sprintf("(old-%03d.md)", index)) || !strings.Contains(text, fmt.Sprintf("(new-%03d.md)", index)) {
				return false
			}
		}
		return true
	})
	select {
	case err := <-done:
		t.Fatalf("rename burst terminated watcher: %v", err)
	default:
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("watch did not finish queued burst reconciliation after cancellation")
	}
}
