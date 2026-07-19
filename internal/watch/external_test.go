package watch

import (
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/links"
)

func TestExternalWatchDirectoriesUseTargetParents(t *testing.T) {
	external := t.TempDir()
	manifest := links.FilesManifest{Files: []links.FileRecord{
		{Scope: "external", Path: filepath.ToSlash(filepath.Join(external, "asset.bin")), Kind: "file", Present: true},
		{Scope: "repository", Path: "docs/page.md", Kind: "file", Present: true},
	}}
	directories := externalWatchDirectories(manifest)
	if len(directories) != 1 || directories[0] != external {
		t.Fatalf("external directories = %#v, want [%q]", directories, external)
	}
	if !externalEvent(filepath.Join(external, "asset.bin"), map[string]bool{external: true}) {
		t.Fatal("external target event was not recognized")
	}
}
