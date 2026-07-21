package reverseindex

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

func TestInventoryFoldersWithPreparationBoundsConcurrencyAndMergesByFolder(t *testing.T) {
	const count = reverseWorkerLimit * 2
	folders := make(map[string]struct{}, count)
	for index := 0; index < count; index++ {
		folders[fmt.Sprintf("folder-%02d", index)] = struct{}{}
	}

	var active atomic.Int32
	var maximum atomic.Int32
	var calls atomic.Int32
	folderFiles, existingManaged, err := inventoryFoldersWithPreparation("", config.Default(), nil, folders, facts{}, func(_ string, _ config.Config, _ *ignorepolicy.Hierarchy, folder string, _ facts) (inventoryFolderResult, error) {
		current := active.Add(1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		calls.Add(1)
		time.Sleep(time.Millisecond)
		active.Add(-1)
		return inventoryFolderResult{
			files:           []string{folder + "/file.go"},
			existingManaged: true,
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != count {
		t.Fatalf("preparation calls = %d, want %d", got, count)
	}
	if got := maximum.Load(); got > reverseWorkerLimit {
		t.Fatalf("maximum concurrency %d exceeds limit %d", got, reverseWorkerLimit)
	}
	if got := maximum.Load(); got < 2 {
		t.Fatalf("maximum concurrency %d did not demonstrate parallel work", got)
	}
	if len(folderFiles) != count || len(existingManaged) != count {
		t.Fatalf("merged folders = %d files, %d managed; want %d each", len(folderFiles), len(existingManaged), count)
	}
	for folder := range folders {
		if got := folderFiles[folder]; len(got) != 1 || got[0] != folder+"/file.go" {
			t.Fatalf("folder %q files = %#v", folder, got)
		}
	}
}

func TestInventoryFoldersWithPreparationReturnsFirstSortedError(t *testing.T) {
	first := errors.New("first sorted error")
	second := errors.New("second sorted error")
	folders := map[string]struct{}{"z-folder": {}, "a-folder": {}, "m-folder": {}}

	_, _, err := inventoryFoldersWithPreparation("", config.Default(), nil, folders, facts{}, func(_ string, _ config.Config, _ *ignorepolicy.Hierarchy, folder string, _ facts) (inventoryFolderResult, error) {
		switch folder {
		case "a-folder":
			time.Sleep(3 * time.Millisecond)
			return inventoryFolderResult{}, first
		case "m-folder":
			return inventoryFolderResult{}, second
		default:
			return inventoryFolderResult{}, nil
		}
	})
	if !errors.Is(err, first) {
		t.Fatalf("error = %v, want sorted first error %v", err, first)
	}
}
