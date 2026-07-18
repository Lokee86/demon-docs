package reconcile

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestMixedLineEndingsPreserveUnmanagedBytes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	readme := filepath.Join(root, "README.md")
	source := []byte("# Docs\r\n\nAuthor line with spaces  \r\n\r\n## Direct Files\n<!-- doc-ledger:files:start -->\r\n<!-- doc-ledger:files:end -->\n\n## Stub Files\r\n<!-- doc-ledger:stubs:start -->\n<!-- doc-ledger:stubs:end -->\r\n\r\n## Direct Folders\n<!-- doc-ledger:folders:start -->\r\n<!-- doc-ledger:folders:end -->\n\nTail without final newline")
	write(t, readme, string(source))
	write(t, filepath.Join(root, "alpha.md"), "# Alpha")
	result, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(result); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	for _, exact := range [][]byte{
		[]byte("# Docs\r\n\nAuthor line with spaces  \r\n\r\n"),
		[]byte("\n\nTail without final newline"),
	} {
		if !bytes.Contains(got, exact) {
			t.Errorf("unmanaged mixed-ending bytes changed; missing %q in %q", exact, got)
		}
	}
	if bytes.HasSuffix(got, []byte("\n")) || bytes.HasSuffix(got, []byte("\r")) {
		t.Fatalf("final newline state changed: %q", got)
	}
}
