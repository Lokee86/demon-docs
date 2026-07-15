package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/doc-ledger/internal/config"
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
