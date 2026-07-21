package reverseindex

import (
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/model"
)

func TestReconcileSelectedFoldersWithPreparationMergesUpdatesInSortedOrder(t *testing.T) {
	selected := map[string]struct{}{"z-folder": {}, "a-folder": {}, "m-folder": {}}
	folderFiles := map[string][]string{}

	updates, indexCount, err := reconcileSelectedFoldersWithPreparation("", config.Default(), folderFiles, facts{}, selected, func(_ string, _ config.Config, folder string, _ []string, _ facts) (folderReconciliationResult, error) {
		switch folder {
		case "z-folder":
			time.Sleep(3 * time.Millisecond)
		case "a-folder":
			time.Sleep(time.Millisecond)
		}
		return folderReconciliationResult{
			update:  model.FileUpdate{Path: folder, NewText: folder},
			changed: true,
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if indexCount != 3 {
		t.Fatalf("index count = %d, want 3", indexCount)
	}
	if len(updates) != 3 {
		t.Fatalf("updates = %#v, want 3 updates", updates)
	}
	for index, want := range []string{"a-folder", "m-folder", "z-folder"} {
		if updates[index].Path != want {
			t.Fatalf("update %d path = %q, want %q", index, updates[index].Path, want)
		}
	}
}

func TestReconcileSelectedFoldersWithPreparationReturnsFirstSortedError(t *testing.T) {
	first := errors.New("first sorted error")
	second := errors.New("second sorted error")
	selected := map[string]struct{}{"z-folder": {}, "a-folder": {}, "m-folder": {}}

	_, _, err := reconcileSelectedFoldersWithPreparation("", config.Default(), nil, facts{}, selected, func(_ string, _ config.Config, folder string, _ []string, _ facts) (folderReconciliationResult, error) {
		switch folder {
		case "a-folder":
			time.Sleep(3 * time.Millisecond)
			return folderReconciliationResult{}, first
		case "m-folder":
			return folderReconciliationResult{}, second
		default:
			return folderReconciliationResult{}, nil
		}
	})
	if !errors.Is(err, first) {
		t.Fatalf("error = %v, want sorted first error %v", err, first)
	}
}
