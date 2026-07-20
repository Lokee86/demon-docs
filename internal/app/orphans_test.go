package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func TestFindOrphanDocumentsIgnoresIndexesDraftsAndSelfLinks(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	for _, path := range []string{
		"INDEX.md",
		"docs/INDEX.md",
		"docs/source.md",
		"docs/linked.md",
		"docs/orphan.md",
		"docs/self-only.md",
		"docs/reference.pdf",
		"docs/stubs/draft.md",
		"docs/stubs/nested/deep.md",
	} {
		absolute := filepath.Join(repositoryRoot, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absolute, []byte("# Test\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files := []links.FileRecord{
		{ID: "root-index", Path: "INDEX.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "docs-index", Path: "docs/INDEX.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "source", Path: "docs/source.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "linked", Path: "docs/linked.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "orphan", Path: "docs/orphan.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "self-only", Path: "docs/self-only.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "reference", Path: "docs/reference.pdf", Scope: "repository", Kind: "file", Present: true},
		{ID: "draft", Path: "docs/stubs/draft.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "nested-draft", Path: "docs/stubs/nested/deep.md", Scope: "repository", Kind: "file", Present: true},
	}
	plan := links.Plan{
		Files: links.FilesManifest{Files: files},
		Links: links.LinksManifest{Links: []links.LinkRecord{
			{SourceFileID: "source", TargetFileID: "linked", Status: "valid"},
			{SourceFileID: "linked", TargetFileID: "source", Status: "valid"},
			{SourceFileID: "docs-index", TargetFileID: "orphan", Status: "valid"},
			{SourceFileID: "root-index", TargetFileID: "orphan", Status: "valid"},
			{SourceFileID: "draft", TargetFileID: "orphan", Status: "valid"},
			{SourceFileID: "nested-draft", TargetFileID: "orphan", Status: "valid"},
			{SourceFileID: "self-only", TargetFileID: "self-only", Status: "valid"},
		}},
	}

	orphans, err := findOrphanDocuments(repository.Scope{RepositoryRoot: repositoryRoot, DocsRoot: docsRoot}, config.Default(), plan)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"docs/orphan.md", "docs/self-only.md"}
	if !reflect.DeepEqual(orphans, want) {
		t.Fatalf("orphans=%v want=%v", orphans, want)
	}
}
