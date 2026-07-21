package validationcache

import (
	"sync"
	"testing"
)

func TestStoreMergesIndependentSubsystemResultsConcurrently(t *testing.T) {
	store := &Store{
		entries: map[string]Entry{},
		dirty:   map[string]Entry{},
		deleted: map[string]bool{},
	}
	path := "docs/guide.md"
	frontmatterIdentity := ContentHash([]byte("frontmatter"))
	formatIdentity := ContentHash([]byte("whole-document"))
	frontmatterPolicy := Hash("frontmatter-policy")
	formatPolicy := Hash("format-policy")
	frontmatterSchema := Hash("frontmatter-schema")
	formatSchema := Hash("format-schema")
	immutable := Hash(nil)

	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		store.Merge(Entry{
			Path:                      path,
			ContentSHA256:             ContentHash([]byte("whole-document")),
			FrontmatterIdentitySHA256: frontmatterIdentity,
			FrontmatterPolicyHash:     frontmatterPolicy,
			FrontmatterSchemaHash:     frontmatterSchema,
			ImmutableSnapshotHash:     immutable,
			DocumentID:                "doc-1",
			DocumentType:              "guide",
			SchemaName:                "guide",
			ImmutableValues:           map[string]any{"document_id": "doc-1"},
			FrontmatterClean:          true,
		})
	}()
	go func() {
		defer wait.Done()
		store.Merge(Entry{
			Path:                 path,
			ContentSHA256:        ContentHash([]byte("whole-document")),
			FormatIdentitySHA256: formatIdentity,
			FormatPolicyHash:     formatPolicy,
			FormatSchemaHash:     formatSchema,
			DocumentID:           "doc-1",
			DocumentType:         "guide",
			SchemaName:           "guide",
			FormatClean:          true,
		})
	}()
	wait.Wait()

	frontmatterEntry, ok := store.LookupFrontmatter(path, frontmatterIdentity, frontmatterPolicy, frontmatterSchema, immutable)
	if !ok {
		t.Fatal("merged frontmatter cache entry was not reachable")
	}
	if frontmatterEntry.ImmutableValues["document_id"] != "doc-1" {
		t.Fatalf("frontmatter cache metadata was lost: %#v", frontmatterEntry.ImmutableValues)
	}
	if _, ok := store.LookupFormat(path, formatIdentity, formatPolicy, formatSchema); !ok {
		t.Fatal("merged format cache entry was not reachable")
	}
}

func TestSubsystemIdentityChangeDoesNotDiscardOtherCleanResult(t *testing.T) {
	store := &Store{
		entries: map[string]Entry{},
		dirty:   map[string]Entry{},
		deleted: map[string]bool{},
	}
	path := "docs/guide.md"
	frontmatter := frontmatterEntry(path, "frontmatter", "first-content")
	store.Merge(frontmatter)
	store.Merge(Entry{
		Path:                 path,
		ContentSHA256:        ContentHash([]byte("first-content")),
		FormatIdentitySHA256: ContentHash([]byte("first-content")),
		FormatPolicyHash:     Hash("format-policy"),
		FormatSchemaHash:     Hash("format-schema"),
		DocumentID:           "doc-1",
		SchemaName:           "guide",
		FormatClean:          true,
	})

	store.Merge(Entry{
		Path:                 path,
		ContentSHA256:        ContentHash([]byte("second-content")),
		FormatIdentitySHA256: ContentHash([]byte("second-content")),
		FormatPolicyHash:     Hash("format-policy"),
		FormatSchemaHash:     Hash("format-schema"),
		DocumentID:           "doc-1",
		SchemaName:           "guide",
		FormatClean:          true,
	})

	if _, ok := store.LookupFrontmatter(path, frontmatter.FrontmatterIdentitySHA256, frontmatter.FrontmatterPolicyHash, frontmatter.FrontmatterSchemaHash, frontmatter.ImmutableSnapshotHash); !ok {
		t.Fatal("format identity update discarded reusable frontmatter result")
	}
	if _, ok := store.LookupFormat(path, ContentHash([]byte("second-content")), Hash("format-policy"), Hash("format-schema")); !ok {
		t.Fatal("updated format result was not reachable")
	}
}
