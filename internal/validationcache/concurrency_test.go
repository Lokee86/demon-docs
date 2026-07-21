package validationcache

import (
	"sync"
	"testing"
)

func TestStoreMergesSubsystemResultsConcurrently(t *testing.T) {
	store := &Store{
		entries: map[string]Entry{},
		dirty:   map[string]Entry{},
		deleted: map[string]bool{},
	}
	identity := Entry{
		Path:                  "docs/guide.md",
		ContentSHA256:         ContentHash([]byte("content")),
		EngineVersion:         EngineVersion,
		FrontmatterPolicyHash: Hash("frontmatter"),
		EffectiveSchemaHash:   Hash("schema"),
		ImmutableSnapshotHash: Hash(nil),
	}

	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		entry := identity
		entry.FrontmatterClean = true
		entry.ImmutableValues = map[string]any{"document_id": "doc-1"}
		store.Merge(entry)
	}()
	go func() {
		defer wait.Done()
		entry := identity
		entry.FormatClean = true
		store.Merge(entry)
	}()
	wait.Wait()

	entry, ok := store.Lookup(identity.Path, identity.ContentSHA256, identity.FrontmatterPolicyHash, identity.EffectiveSchemaHash, identity.ImmutableSnapshotHash)
	if !ok {
		t.Fatal("merged cache entry was not reachable")
	}
	if !entry.FrontmatterClean || !entry.FormatClean {
		t.Fatalf("parallel subsystem results were not merged: %#v", entry)
	}
	if entry.ImmutableValues["document_id"] != "doc-1" {
		t.Fatalf("frontmatter cache metadata was lost: %#v", entry.ImmutableValues)
	}
}
