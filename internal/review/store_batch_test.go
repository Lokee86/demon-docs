package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestStoreAppendBatchPublishesOneCommitAndExpandsHistory(t *testing.T) {
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
	if stored[0].CommitHash != stored[1].CommitHash {
		t.Fatalf("batch used multiple commits: %s != %s", stored[0].CommitHash, stored[1].CommitHash)
	}

	ref, err := store.repository.Storer.Reference(reviewReference)
	if err != nil {
		t.Fatal(err)
	}
	commit, err := object.GetCommit(store.repository.Storer, ref.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if len(commit.ParentHashes) != 0 {
		t.Fatalf("first batch commit has %d parents, want 0", len(commit.ParentHashes))
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

	latest, err := store.History(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(latest) != 1 || latest[0].Change == nil || latest[0].Change.ID != second.ID {
		t.Fatalf("limited history = %#v, want newest batch event", latest)
	}
	found, err := store.Find(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Change == nil || found.Change.ID != first.ID || string(found.Before) != "one-old" {
		t.Fatalf("find returned %#v", found)
	}
}

func TestStoreAppendBatchPreservesNilAndEmptySnapshots(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	first := Change{ID: "ch-empty-after", RunID: "run-test", SourceFileID: "file-1", SourcePath: "one.md"}
	second := Change{ID: "ch-empty-before", RunID: "run-test", SourceFileID: "file-2", SourcePath: "two.md"}
	if _, err := store.AppendBatch([]AppendRequest{
		{Event: Event{Type: EventChange, Change: &first}, Before: nil, After: []byte{}},
		{Event: Event{Type: EventChange, Change: &second}, Before: []byte{}, After: nil},
	}); err != nil {
		t.Fatal(err)
	}

	history, err := store.History(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 {
		t.Fatalf("history length = %d, want 2", len(history))
	}
	if history[0].Before == nil || len(history[0].Before) != 0 || history[0].After != nil {
		t.Fatalf("second snapshots lost nil/empty distinction: before=%#v after=%#v", history[0].Before, history[0].After)
	}
	if history[1].Before != nil || history[1].After == nil || len(history[1].After) != 0 {
		t.Fatalf("first snapshots lost nil/empty distinction: before=%#v after=%#v", history[1].Before, history[1].After)
	}
}

func TestStoreAppendBatchUsesConstantLooseObjectCount(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	before := countLooseObjects(t, root)

	requests := make([]AppendRequest, 240)
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	for index := range requests {
		change := Change{
			ID:           fmt.Sprintf("ch-%03d", index),
			RunID:        "run-large",
			SourceFileID: fmt.Sprintf("file-%03d", index),
			SourcePath:   fmt.Sprintf("docs/file-%03d.md", index),
			AppliedAt:    now,
		}
		requests[index] = AppendRequest{
			Event:  Event{Type: EventChange, Time: now, Change: &change},
			Before: []byte(fmt.Sprintf("before-%03d", index)),
			After:  []byte(fmt.Sprintf("after-%03d", index)),
		}
	}
	stored, err := store.AppendBatch(requests)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != len(requests) {
		t.Fatalf("stored %d events, want %d", len(stored), len(requests))
	}
	after := countLooseObjects(t, root)
	if added := after - before; added != 3 {
		t.Fatalf("240-change batch added %d loose objects, want 3", added)
	}
}

func TestStoreHistoryReadsLegacyPerEventCommit(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	change := Change{ID: "ch-legacy", RunID: "run-legacy", SourceFileID: "file-legacy", SourcePath: "legacy.md", AppliedAt: now}
	event := Event{SchemaVersion: SchemaVersion, ID: "ev-legacy", Type: EventChange, Time: now, Change: &change}
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	eventHash, err := writeBlob(store.repository, payload)
	if err != nil {
		t.Fatal(err)
	}
	beforeHash, err := writeBlob(store.repository, []byte("legacy-old"))
	if err != nil {
		t.Fatal(err)
	}
	afterHash, err := writeBlob(store.repository, []byte("legacy-new"))
	if err != nil {
		t.Fatal(err)
	}
	entries := []object.TreeEntry{
		{Name: "event.json", Mode: filemode.Regular, Hash: eventHash},
		{Name: "before", Mode: filemode.Regular, Hash: beforeHash},
		{Name: "after", Mode: filemode.Regular, Hash: afterHash},
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	treeHash, err := writeTree(store.repository, entries)
	if err != nil {
		t.Fatal(err)
	}
	commit := object.Commit{
		Author:    object.Signature{Name: "Demon Docs", Email: "ddocs@local", When: now},
		Committer: object.Signature{Name: "Demon Docs", Email: "ddocs@local", When: now},
		Message:   eventMessage(event),
		TreeHash:  treeHash,
	}
	encoded := store.repository.Storer.NewEncodedObject()
	if err := commit.Encode(encoded); err != nil {
		t.Fatal(err)
	}
	commitHash, err := store.repository.Storer.SetEncodedObject(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.repository.Storer.SetReference(plumbing.NewHashReference(reviewReference, commitHash)); err != nil {
		t.Fatal(err)
	}

	history, err := store.History(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Change == nil || history[0].Change.ID != change.ID {
		t.Fatalf("legacy history = %#v", history)
	}
	if string(history[0].Before) != "legacy-old" || string(history[0].After) != "legacy-new" {
		t.Fatalf("legacy snapshots were not retained: %#v", history[0])
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

func countLooseObjects(t *testing.T, root string) int {
	t.Helper()
	count := 0
	err := filepath.WalkDir(filepath.Join(root, ".ddocs", "objects"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return count
}
