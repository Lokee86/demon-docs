package validationcache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestOpenWithoutPrivateRepositoryDoesNotInitializeState(t *testing.T) {
	root := t.TempDir()
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(Entry{
		Path:                  "docs/guide.md",
		ContentSHA256:         ContentHash([]byte("content")),
		EngineVersion:         EngineVersion,
		FrontmatterPolicyHash: Hash("frontmatter"),
		EffectiveSchemaHash:   Hash("schema"),
		ImmutableSnapshotHash: Hash(nil),
		FrontmatterClean:      true,
	})
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".ddocs")); !os.IsNotExist(err) {
		t.Fatalf("validation cache initialized private state: %v", err)
	}
}

func TestRetainDeletesRecordsOutsideActiveScope(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	identity := func(path string) Entry {
		return Entry{
			Path:                  path,
			ContentSHA256:         ContentHash([]byte(path)),
			EngineVersion:         EngineVersion,
			FrontmatterPolicyHash: Hash("frontmatter"),
			EffectiveSchemaHash:   Hash("schema"),
			ImmutableSnapshotHash: Hash(nil),
			FrontmatterClean:      true,
		}
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(identity("docs/keep.md"))
	store.Merge(identity("docs/remove.md"))
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Retain([]string{"docs/keep.md"})
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	keep := identity("docs/keep.md")
	if _, ok := store.Lookup(keep.Path, keep.ContentSHA256, keep.FrontmatterPolicyHash, keep.EffectiveSchemaHash, keep.ImmutableSnapshotHash); !ok {
		t.Fatal("active cache record was removed")
	}
	removed := identity("docs/remove.md")
	if _, ok := store.Lookup(removed.Path, removed.ContentSHA256, removed.FrontmatterPolicyHash, removed.EffectiveSchemaHash, removed.ImmutableSnapshotHash); ok {
		t.Fatal("stale cache record remained reachable")
	}
}

func TestEntryRoundTripsAndSchemaSourceChangesInvalidate(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	format := config.Format{Enabled: true, SchemaDir: ".ddocs/schemas", DocumentSchemaDir: ".ddocs/document-schemas"}
	if err := os.MkdirAll(filepath.Join(root, ".ddocs", "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(root, ".ddocs", "schemas", "general.toml")
	if err := os.WriteFile(schemaPath, []byte("name = 'general'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	firstSchemaHash := EffectiveSchemaHash(root, format, "general", "")
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(Entry{
		Path:                  "docs/Guide.md",
		ContentSHA256:         ContentHash([]byte("content")),
		EngineVersion:         EngineVersion,
		FrontmatterPolicyHash: Hash("frontmatter"),
		EffectiveSchemaHash:   firstSchemaHash,
		ImmutableSnapshotHash: Hash(nil),
		FrontmatterClean:      true,
	})
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reopened.Lookup("docs/guide.md", ContentHash([]byte("content")), Hash("frontmatter"), firstSchemaHash, Hash(nil)); !ok {
		t.Fatal("stored validation entry did not round-trip through ddrepo")
	}
	if err := os.WriteFile(schemaPath, []byte("name = 'general'\ndescription = 'changed'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if EffectiveSchemaHash(root, format, "general", "") == firstSchemaHash {
		t.Fatal("schema source change did not change effective schema hash")
	}
}
