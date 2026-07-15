package reconcile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/doc-ledger/internal/config"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
func TestFixIsIdempotentAndUsesConfiguredIndexEverywhere(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "page.md"), "# Page\n")
	write(t, filepath.Join(root, "guide", "topic.md"), "# Topic\n")
	write(t, filepath.Join(root, "stubs", "draft.md"), "# Draft\n")
	c := config.Default()
	c.IndexFile = "!README.md"
	c.Files.IndexFile = "!README.md"
	c.ParentLink.IndexedFiles = true
	first, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Updates) != 5 {
		t.Fatalf("updates=%d: %+v", len(first.Updates), first.Updates)
	}
	if _, err := Apply(first); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "README.md")); !os.IsNotExist(err) {
		t.Fatal("silently created README.md")
	}
	rootIndex, err := os.ReadFile(filepath.Join(root, "!README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rootIndex), "guide/!README.md") {
		t.Fatal(string(rootIndex))
	}
	child, err := os.ReadFile(filepath.Join(root, "guide", "!README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(child), "../!README.md") {
		t.Fatal(string(child))
	}
	second, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 0 {
		t.Fatalf("not idempotent: %+v", second.Updates)
	}
}
func TestCheckPlanningDoesNotMutate(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "page.md")
	write(t, path, "# Page\n")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Updates) == 0 {
		t.Fatal("expected drift")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("planning mutated file")
	}
	if _, err := os.Stat(filepath.Join(root, "README.md")); !os.IsNotExist(err) {
		t.Fatal("planning created index")
	}
}
func TestPreservesCRLFUnmanagedContentAndFinalNewline(t *testing.T) {
	root := t.TempDir()
	readme := filepath.Join(root, "README.md")
	source := "# Docs\r\n\r\nUser text  \r\n\r\n## Direct Files\r\n<!-- doc-ledger:files:start -->\r\n<!-- doc-ledger:files:end -->\r\n\r\n## Stub Files\r\n<!-- doc-ledger:stubs:start -->\r\n<!-- doc-ledger:stubs:end -->\r\n\r\n## Direct Folders\r\n<!-- doc-ledger:folders:start -->\r\n<!-- doc-ledger:folders:end -->\r\n\r\nTail\r\n"
	write(t, readme, source)
	write(t, filepath.Join(root, "a.md"), "# A\r\n")
	result, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(result); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Fatal("introduced LF into CRLF file")
	}
	if !strings.Contains(got, "User text  \r\n") || !strings.HasSuffix(got, "Tail\r\n") {
		t.Fatalf("unmanaged bytes/final newline changed: %q", got)
	}
}
func TestDescriptionMovesAndStaleMessages(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "README.md"), "# Docs\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n- [gone.md](gone.md) - Gone.\n<!-- doc-ledger:files:end -->\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n- [foo.md](stubs/foo.md) - Stub: custom details.\n<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n<!-- doc-ledger:folders:end -->\n")
	write(t, filepath.Join(root, "foo.md"), "# Foo\n")
	result, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Messages) != 1 || !strings.Contains(result.Messages[0], "gone.md") {
		t.Fatalf("messages=%v", result.Messages)
	}
	index := result.Updates[0].NewText
	if !strings.Contains(index, "- [foo.md](foo.md) - Custom details.") {
		t.Fatal(index)
	}
}
