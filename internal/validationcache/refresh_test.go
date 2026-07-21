package validationcache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestRefreshPublishedRetainsUnaffectedValidationResult(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(filepath.Join(root, ".ddocs")); err != nil {
		t.Fatal(err)
	}
	oldData := []byte("# Guide\n\nOld body.\n")
	newData := []byte("# Guide\n\nNew body.\n")
	entry := refreshTestEntry(oldData)

	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(entry)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	if err := RefreshPublished(root, []PublishedRewrite{{
		Path:        filepath.Join(root, "docs", "guide.md"),
		OldData:     oldData,
		NewData:     newData,
		Invalidated: SurfaceFormat,
	}}); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	updated, ok := reopened.Lookup(entry.Path, ContentHash(newData), entry.FrontmatterPolicyHash, entry.EffectiveSchemaHash, entry.ImmutableSnapshotHash)
	if !ok {
		t.Fatal("published content hash did not replace the old cache identity")
	}
	if !updated.FrontmatterClean || updated.FormatClean {
		t.Fatalf("unexpected retained surfaces: frontmatter=%t format=%t", updated.FrontmatterClean, updated.FormatClean)
	}
	if _, ok := reopened.Lookup(entry.Path, ContentHash(oldData), entry.FrontmatterPolicyHash, entry.EffectiveSchemaHash, entry.ImmutableSnapshotHash); ok {
		t.Fatal("old content hash remained reachable after refresh")
	}
}

func TestRefreshPublishedPreservesAllValidationResultsForLinkRewrite(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(filepath.Join(root, ".ddocs")); err != nil {
		t.Fatal(err)
	}
	oldData := []byte("[Guide](old.md)\n")
	newData := []byte("[Guide](new.md)\n")
	entry := refreshTestEntry(oldData)
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(entry)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}
	if err := RefreshPublished(root, []PublishedRewrite{{Path: "docs/guide.md", OldData: oldData, NewData: newData}}); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	updated, ok := reopened.Lookup(entry.Path, ContentHash(newData), entry.FrontmatterPolicyHash, entry.EffectiveSchemaHash, entry.ImmutableSnapshotHash)
	if !ok || !updated.FrontmatterClean || !updated.FormatClean {
		t.Fatalf("link rewrite did not preserve clean validation results: %#v ok=%t", updated, ok)
	}
}

func TestRefreshPublishedDoesNotCarryForwardStaleEntry(t *testing.T) {
	root := t.TempDir()
	store := &Store{entries: map[string]Entry{}, dirty: map[string]Entry{}, deleted: map[string]bool{}}
	entry := refreshTestEntry([]byte("cached\n"))
	store.entries[NormalizePath(entry.Path)] = entry

	changed, err := store.RefreshPublished(root, filepath.Join(root, "docs", "guide.md"), []byte("different old bytes\n"), []byte("new bytes\n"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("stale cache entry was refreshed across unrelated content")
	}
	if store.entries[NormalizePath(entry.Path)].ContentSHA256 != entry.ContentSHA256 {
		t.Fatal("stale cache entry was mutated")
	}
}

func TestRefreshPublishedDeletesEntryWhenAllResultsAreInvalidated(t *testing.T) {
	root := t.TempDir()
	oldData := []byte("old\n")
	store := &Store{entries: map[string]Entry{}, dirty: map[string]Entry{}, deleted: map[string]bool{}}
	entry := refreshTestEntry(oldData)
	normalized := NormalizePath(entry.Path)
	store.entries[normalized] = entry

	changed, err := store.RefreshPublished(root, filepath.Join(root, "docs", "guide.md"), oldData, []byte("new\n"), SurfaceFrontmatter|SurfaceFormat)
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !store.deleted[normalized] {
		t.Fatal("fully invalidated cache entry was not deleted")
	}
	if _, ok := store.entries[normalized]; ok {
		t.Fatal("fully invalidated cache entry remained in memory")
	}
}

func refreshTestEntry(data []byte) Entry {
	return Entry{
		Path:                  "docs/guide.md",
		ContentSHA256:         ContentHash(data),
		EngineVersion:         EngineVersion,
		FrontmatterPolicyHash: Hash("frontmatter"),
		EffectiveSchemaHash:   Hash("schema"),
		ImmutableSnapshotHash: Hash(nil),
		DocumentID:            "guide",
		DocumentType:          "general",
		SchemaName:            "general",
		FrontmatterClean:      true,
		FormatClean:           true,
	}
}

func TestRepositoryRelativePathRejectsOutsideRewrite(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.md")
	if err := os.WriteFile(outside, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := repositoryRelativePath(root, outside); err == nil {
		t.Fatal("outside rewrite path was accepted")
	}
}
