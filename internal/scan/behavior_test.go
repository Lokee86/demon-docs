package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/doc-ledger/internal/config"
)

func TestNestedPatternsDraftsAndIndexExclusion(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"README.md", "root.md", "guide/README.md", "guide/deep/setup.md", "ignored/deep/skip.md", "stubs/draft.md", "stubs/nested/not-direct.md"} {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c := config.Default()
	c.Files.ExcludePatterns = []string{"ignored/**/*.md"}
	tree, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if got := tree.Folders[root].DirectFiles; len(got) != 1 || filepath.Base(got[0]) != "root.md" {
		t.Fatalf("root=%v", got)
	}
	if got := tree.Folders[root].StubFiles; len(got) != 1 || filepath.Base(got[0]) != "draft.md" {
		t.Fatalf("stubs=%v", got)
	}
	if len(tree.Folders[filepath.Join(root, "guide")].DirectFiles) != 0 {
		t.Fatal("index was indexed")
	}
	if len(tree.Folders[filepath.Join(root, "guide", "deep")].DirectFiles) != 1 {
		t.Fatal("nested include failed")
	}
	if len(tree.Folders[filepath.Join(root, "ignored", "deep")].DirectFiles) != 0 {
		t.Fatal("nested exclude failed")
	}
	if nested, ok := tree.Folders[filepath.Join(root, "stubs", "nested")]; !ok || len(nested.DirectFiles) != 1 {
		t.Fatal("scanner did not preserve nested draft-folder behavior")
	}
}

func TestIndexablePatternCases(t *testing.T) {
	root := t.TempDir()
	c := config.Default()
	cases := []struct {
		rel  string
		want bool
	}{{"page.md", true}, {"deep/page.md", true}, {"README.md", false}, {"page.txt", false}}
	for _, tc := range cases {
		got, err := IsIndexable(root, filepath.Join(root, filepath.FromSlash(tc.rel)), c)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("%s=%t want %t", tc.rel, got, tc.want)
		}
	}
}
