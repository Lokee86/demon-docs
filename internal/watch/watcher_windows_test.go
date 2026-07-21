//go:build windows

package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWindowsRecursiveWatcherAllowsNestedTreeMove(t *testing.T) {
	root := t.TempDir()
	oldTree := filepath.Join(root, "old")
	newTree := filepath.Join(root, "new")
	oldFile := filepath.Join(oldTree, "nested", "deep", "file.md")
	newFile := filepath.Join(newTree, "nested", "deep", "file.md")
	if err := os.MkdirAll(filepath.Dir(oldFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldFile, []byte("test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	watcher, err := newEventWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			t.Errorf("close watcher: %v", err)
		}
	}()
	if err := watcher.Add(root); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "watch-ready.tmp")
	if err := os.WriteFile(marker, []byte("ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readyDeadline := time.After(3 * time.Second)
	for {
		select {
		case event, ok := <-watcher.Events():
			if !ok {
				t.Fatal("watcher events closed before registration was observed")
			}
			if filepath.Clean(event.Name) == filepath.Clean(marker) {
				goto watcherReady
			}
		case err := <-watcher.Errors():
			if err != nil {
				t.Fatalf("watcher error before move: %v", err)
			}
		case <-readyDeadline:
			t.Fatal("watcher did not observe registration marker")
		}
	}

watcherReady:
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldTree, newTree); err != nil {
		t.Fatalf("move nested watched tree: %v", err)
	}
	if _, err := os.Stat(oldTree); !os.IsNotExist(err) {
		t.Fatalf("old tree still exists after move: %v", err)
	}
	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}

	var sawOldRename bool
	var sawNewCreate bool
	deadline := time.After(3 * time.Second)
	for !sawOldRename || !sawNewCreate {
		select {
		case event, ok := <-watcher.Events():
			if !ok {
				t.Fatal("watcher events closed before move events arrived")
			}
			if filepath.Clean(event.Name) == filepath.Clean(oldTree) && event.Op&fsnotify.Rename != 0 {
				sawOldRename = true
			}
			if filepath.Clean(event.Name) == filepath.Clean(newTree) && event.Op&fsnotify.Create != 0 {
				sawNewCreate = true
			}
		case err := <-watcher.Errors():
			if err != nil {
				t.Fatalf("watcher error: %v", err)
			}
		case <-deadline:
			t.Fatalf("move events missing: old rename=%t new create=%t", sawOldRename, sawNewCreate)
		}
	}
}
