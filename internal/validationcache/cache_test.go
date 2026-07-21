package validationcache

import (
	"os"
	"path/filepath"
	"runtime"
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
	store.Merge(frontmatterEntry("docs/guide.md", "frontmatter", "content"))
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
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(frontmatterEntry("docs/keep.md", "keep-frontmatter", "keep-content"))
	store.Merge(frontmatterEntry("docs/remove.md", "remove-frontmatter", "remove-content"))
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
	keep := frontmatterEntry("docs/keep.md", "keep-frontmatter", "keep-content")
	if _, ok := store.LookupFrontmatter(keep.Path, keep.FrontmatterIdentitySHA256, keep.FrontmatterPolicyHash, keep.FrontmatterSchemaHash, keep.ImmutableSnapshotHash); !ok {
		t.Fatal("active cache record was removed")
	}
	removed := frontmatterEntry("docs/remove.md", "remove-frontmatter", "remove-content")
	if _, ok := store.LookupFrontmatter(removed.Path, removed.FrontmatterIdentitySHA256, removed.FrontmatterPolicyHash, removed.FrontmatterSchemaHash, removed.ImmutableSnapshotHash); ok {
		t.Fatal("stale cache record remained reachable")
	}
}

func TestMergeDoesNotDirtyAnUnchangedEntry(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	entry := frontmatterEntry("docs/guide.md", "frontmatter", "content")
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(entry)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(entry)
	if len(store.dirty) != 0 {
		t.Fatalf("unchanged cache entry was marked dirty: %#v", store.dirty)
	}
}

func TestSchemaHasherMemoizesOneValidationPass(t *testing.T) {
	root := t.TempDir()
	format := config.Format{Enabled: true, SchemaDir: "schemas", DocumentSchemaDir: "document-schemas"}
	if err := os.MkdirAll(filepath.Join(root, "schemas"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "schemas", "general.toml")
	if err := os.WriteFile(path, []byte("name = 'general'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hasher := NewSchemaHasher(root, format)
	first := hasher.Effective("general", "")
	if err := os.WriteFile(path, []byte("name = 'general'\ndescription = 'changed'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if second := hasher.Effective("general", ""); second != first {
		t.Fatalf("one validation pass observed inconsistent schema hashes: %s != %s", second, first)
	}
	if refreshed := NewSchemaHasher(root, format).Effective("general", ""); refreshed == first {
		t.Fatal("new validation pass did not observe changed schema source")
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
	entry := frontmatterEntry("docs/Guide.md", "frontmatter", "content")
	entry.FrontmatterSchemaHash = firstSchemaHash
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(entry)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	lookupPath := "docs/Guide.md"
	if runtime.GOOS == "windows" {
		lookupPath = "docs/guide.md"
	}
	if _, ok := reopened.LookupFrontmatter(lookupPath, entry.FrontmatterIdentitySHA256, entry.FrontmatterPolicyHash, firstSchemaHash, entry.ImmutableSnapshotHash); !ok {
		t.Fatal("stored validation entry did not round-trip through ddrepo")
	}
	if err := os.WriteFile(schemaPath, []byte("name = 'general'\ndescription = 'changed'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if EffectiveSchemaHash(root, format, "general", "") == firstSchemaHash {
		t.Fatal("schema source change did not change effective schema hash")
	}
}

func frontmatterEntry(path, frontmatter, content string) Entry {
	return Entry{
		Path:                      path,
		ContentSHA256:             ContentHash([]byte(content)),
		EngineVersion:             EngineVersion,
		FrontmatterIdentitySHA256: ContentHash([]byte(frontmatter)),
		FrontmatterPolicyHash:     Hash("frontmatter-policy"),
		FrontmatterSchemaHash:     Hash("schema"),
		ImmutableSnapshotHash:     Hash(nil),
		FrontmatterClean:          true,
	}
}
