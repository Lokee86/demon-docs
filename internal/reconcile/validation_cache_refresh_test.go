package reconcile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/ddrepo"
	"github.com/Lokee86/demon-docs/internal/validationcache"
)

func TestConvergeRefreshesValidationCacheAfterIndexRewrite(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	if _, err := ddrepo.Init(repositoryRoot); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(docsRoot, "INDEX.md")
	oldData := []byte("# Docs\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n<!-- doc-ledger:files:end -->\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n<!-- doc-ledger:folders:end -->\n")
	write(t, indexPath, string(oldData))
	write(t, filepath.Join(docsRoot, "page.md"), "# Page\n")

	entry := validationcache.Entry{
		Path:                      "docs/INDEX.md",
		ContentSHA256:             validationcache.ContentHash(oldData),
		EngineVersion:             validationcache.EngineVersion,
		FrontmatterIdentitySHA256: validationcache.Hash("frontmatter-source"),
		FrontmatterPolicyHash:     validationcache.Hash("frontmatter-policy"),
		FrontmatterSchemaHash:     validationcache.Hash("frontmatter-schema"),
		ImmutableSnapshotHash:     validationcache.Hash(nil),
		FormatIdentitySHA256:      validationcache.Hash("format-source"),
		FormatPolicyHash:          validationcache.Hash("format-policy"),
		FormatSchemaHash:          validationcache.Hash("format-schema"),
		FrontmatterClean:          true,
		FormatClean:               true,
	}
	store, err := validationcache.Open(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	store.Merge(entry)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	_, changed, err := ConvergeWithin(docsRoot, repositoryRoot, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if changed == 0 {
		t.Fatal("expected index rewrite")
	}
	newData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := validationcache.Open(repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	updated, ok := reopened.LookupPath(entry.Path)
	if !ok || updated.ContentSHA256 != validationcache.ContentHash(newData) {
		t.Fatal("index rewrite did not refresh the cached source state")
	}
	if !updated.FrontmatterClean || updated.FormatClean {
		t.Fatalf("index rewrite retained wrong validation surfaces: frontmatter=%t format=%t", updated.FrontmatterClean, updated.FormatClean)
	}
	if _, ok := reopened.LookupFrontmatter(entry.Path, entry.FrontmatterIdentitySHA256, entry.FrontmatterPolicyHash, entry.FrontmatterSchemaHash, entry.ImmutableSnapshotHash); !ok {
		t.Fatal("index rewrite did not retain the unaffected frontmatter result")
	}
	if _, ok := reopened.LookupFormat(entry.Path, entry.FormatIdentitySHA256, entry.FormatPolicyHash, entry.FormatSchemaHash); ok {
		t.Fatal("index rewrite retained the invalidated format result")
	}
}
