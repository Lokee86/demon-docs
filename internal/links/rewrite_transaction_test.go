package links

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRollbackGeneratedRefusesToOverwriteNewerContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.md")
	before := []byte("before\n")
	after := []byte("after\n")
	newer := []byte("newer external edit\n")
	if err := os.WriteFile(path, before, 0o644); err != nil {
		t.Fatal(err)
	}
	rewrite, err := NewGeneratedRewriteBytes("file-1", path, before, after, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyGenerated([]GeneratedRewrite{rewrite}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, newer, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RollbackGenerated([]GeneratedRewrite{rewrite}); err == nil {
		t.Fatal("expected rollback to reject newer source content")
	}
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != string(newer) {
		t.Fatalf("rollback overwrote newer content: got %q want %q", current, newer)
	}
}
