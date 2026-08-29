//go:build windows

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
)

func TestWindowsWatcherConvergesAfterRealRenameBurst(t *testing.T) {
	const fileCount = 96
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

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ready := make(chan struct{})
	zero := 0.0
	go func() {
		done <- RootSelectedWithRunLock(ctx, root, root, config.Default(), Features{Links: true}, &zero, false, nil, nil, func() error {
			close(ready)
			return nil
		})
	}()
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("Windows watcher did not become ready")
	}

	for index := 0; index < fileCount; index++ {
		if err := os.Rename(oldPaths[index], newPaths[index]); err != nil {
			cancel()
			t.Fatal(err)
		}
	}

	waitFor(t, 30*time.Second, func() bool {
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
		t.Fatalf("real Windows rename burst terminated watcher: %v", err)
	default:
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Windows watcher did not stop after burst test")
	}
}
