package review

import (
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestStoreAppendBatchPublishesCompleteHistoryChain(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	first := Change{ID: "ch-first", RunID: "run-test", SourceFileID: "file-1", SourcePath: "one.md", AppliedAt: now}
	second := Change{ID: "ch-second", RunID: "run-test", SourceFileID: "file-2", SourcePath: "two.md", AppliedAt: now.Add(time.Second)}
	stored, err := store.AppendBatch([]AppendRequest{
		{Event: Event{ID: "ev-first", Type: EventChange, Time: first.AppliedAt, Change: &first}, Before: []byte("one-old"), After: []byte("one-new")},
		{Event: Event{ID: "ev-second", Type: EventChange, Time: second.AppliedAt, Change: &second}, Before: []byte("two-old"), After: []byte("two-new")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 2 || stored[0].CommitHash == "" || stored[1].CommitHash == "" {
		t.Fatalf("unexpected stored batch: %#v", stored)
	}

	history, err := store.History(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("history length = %d, want 2", len(history))
	}
	if history[0].Change == nil || history[0].Change.ID != second.ID || history[1].Change == nil || history[1].Change.ID != first.ID {
		t.Fatalf("unexpected history order: %#v", history)
	}
	if string(history[0].Before) != "two-old" || string(history[1].After) != "one-new" {
		t.Fatalf("batch blobs were not retained: %#v", history)
	}
}

func TestStoreAppendBatchPreflightFailurePublishesNothing(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	change := Change{ID: "ch-valid", RunID: "run-test", SourceFileID: "file-1", SourcePath: "one.md"}
	_, err = store.AppendBatch([]AppendRequest{
		{Event: Event{Type: EventChange, Change: &change}},
		{Event: Event{SchemaVersion: SchemaVersion + 1, Type: EventChange, Change: &change}},
	})
	if err == nil {
		t.Fatal("expected batch validation failure")
	}
	history, historyErr := store.History(0)
	if historyErr != nil {
		t.Fatal(historyErr)
	}
	if len(history) != 0 {
		t.Fatalf("invalid batch published partial history: %#v", history)
	}
}
