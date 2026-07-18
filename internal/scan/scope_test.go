package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestTreeUsesRepositoryOwnedDocignore(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"keep.md", "ignore.md"} {
		if err := os.WriteFile(filepath.Join(docsRoot, name), []byte("# Test\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(repositoryRoot, ".docignore"), []byte("/docs/ignore.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tree, err := TreeWithIgnoreRoot(docsRoot, repositoryRoot, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	files := tree.Folders[docsRoot].DirectFiles
	if len(files) != 1 || filepath.Base(files[0]) != "keep.md" {
		t.Fatalf("unexpected files: %v", files)
	}
}

func TestTreeSkipsSymlinkedFiles(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "linked.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	tree, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if files := tree.Folders[root].DirectFiles; len(files) != 0 {
		t.Fatalf("symlinked file was indexed: %v", files)
	}
}
