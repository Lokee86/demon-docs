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
		Path:                  "docs/INDEX.md",
		ContentSHA256:         validationcache.ContentHash(oldData),
		EngineVersion:         validationcache.EngineVersion,
		FrontmatterPolicyHash: validationcache.Hash("frontmatter"),
		EffectiveSchemaHash:   validationcache.Hash("schema"),
		ImmutableSnapshotHash: validationcache.Hash(nil),
		FrontmatterClean:      true,
		FormatClean:           true,
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
	updated, ok := reopened.Lookup(
		entry.Path,
		validationcache.ContentHash(newData),
		entry.FrontmatterPolicyHash,
		entry.EffectiveSchemaHash,
		entry.ImmutableSnapshotHash,
	)
	if !ok {
		t.Fatal("index rewrite did not refresh the cached content identity")
	}
	if !updated.FrontmatterClean || updated.FormatClean {
		t.Fatalf("index rewrite retained wrong validation surfaces: frontmatter=%t format=%t", updated.FrontmatterClean, updated.FormatClean)
	}
}
