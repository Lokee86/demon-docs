package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestScanIncludesExcludesAndConfiguredIndex(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"!README.md", "page.md", "skip.tmp", "guide/topic.md", "stubs/draft.md"} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c := config.Default()
	c.IndexFile = "!README.md"
	c.Files.IncludePatterns = []string{"**/*.md", "**/*.tmp"}
	c.Files.ExcludePatterns = []string{"**/*.tmp"}
	tree, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	f := tree.Folders[root]
	if len(f.DirectFiles) != 1 || filepath.Base(f.DirectFiles[0]) != "page.md" || len(f.StubFiles) != 1 || len(f.Subfolders) != 1 {
		t.Fatalf("unexpected root scan: %+v", f)
	}
	if _, ok := tree.Folders[filepath.Join(root, "stubs")]; !ok {
		t.Fatal("stub folder absent")
	}
}

func TestScanPrunesPermanentAndDocignorePaths(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{
		"page.md",
		"ignored.md",
		"generated/topic.md",
		".git/secret.md",
		".demon-docs/state.md",
		".obsidian/workspace.md",
		"logseq/config.md",
		"nested/.git/secret.md",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".docignore"), []byte("ignored.md\ngenerated/\n!.git/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tree, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	rootInfo := tree.Folders[root]
	if len(rootInfo.DirectFiles) != 1 || filepath.Base(rootInfo.DirectFiles[0]) != "page.md" {
		t.Fatalf("unexpected direct files: %v", rootInfo.DirectFiles)
	}
	for _, rel := range []string{"generated", ".git", ".demon-docs", ".obsidian", "logseq", "nested/.git"} {
		if _, ok := tree.Folders[filepath.Join(root, filepath.FromSlash(rel))]; ok {
			t.Fatalf("ignored folder was traversed: %s", rel)
		}
	}
}
