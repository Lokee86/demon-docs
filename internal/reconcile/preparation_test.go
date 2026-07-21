package reconcile

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/model"
)

func TestPrepareFolderResultsBoundsConcurrencyAndMergesInIndexOrder(t *testing.T) {
	const count = 32
	var active atomic.Int32
	var maximum atomic.Int32

	updates, matched, err := prepareFolderResults(count, func(index int) (folderPreparationResult, error) {
		current := active.Add(1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		if index%2 == 0 {
			time.Sleep(2 * time.Millisecond)
		}
		active.Add(-1)
		entry := &model.IndexEntry{OriginalLine: fmt.Sprintf("entry-%02d", index)}
		return folderPreparationResult{
			updates: []model.FileUpdate{{Path: fmt.Sprintf("folder-%02d", index)}},
			matched: []*model.IndexEntry{entry},
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := maximum.Load(); got > 16 {
		t.Fatalf("maximum concurrency %d exceeds worker limit 16", got)
	} else if got < 2 {
		t.Fatalf("maximum concurrency %d did not demonstrate parallel preparation", got)
	}
	if len(updates) != count || len(matched) != count {
		t.Fatalf("updates=%d matched=%d, want %d each", len(updates), len(matched), count)
	}
	for index, update := range updates {
		want := fmt.Sprintf("folder-%02d", index)
		if update.Path != want {
			t.Fatalf("update %d path = %q, want %q", index, update.Path, want)
		}
	}
}

func TestPrepareFolderResultsReturnsFirstIndexedErrorWithPriorUpdates(t *testing.T) {
	first := errors.New("first indexed error")
	second := errors.New("second indexed error")
	updates, _, err := prepareFolderResults(4, func(index int) (folderPreparationResult, error) {
		switch index {
		case 0:
			time.Sleep(2 * time.Millisecond)
			return folderPreparationResult{updates: []model.FileUpdate{{Path: "before"}}}, nil
		case 1:
			time.Sleep(3 * time.Millisecond)
			return folderPreparationResult{}, first
		case 2:
			return folderPreparationResult{}, second
		default:
			return folderPreparationResult{updates: []model.FileUpdate{{Path: "after"}}}, nil
		}
	})
	if !errors.Is(err, first) {
		t.Fatalf("error = %v, want %v", err, first)
	}
	if len(updates) != 1 || updates[0].Path != "before" {
		t.Fatalf("partial updates = %#v, want only prior indexed update", updates)
	}
}
