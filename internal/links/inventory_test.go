package links

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInventoryRebuildPrefersPresentDuplicatePathRecord(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "target.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# Target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stored := storePath(root, path)
	inventory := &inventory{
		root: root,
		manifest: FilesManifest{Files: []FileRecord{
			{ID: "present", Path: stored, Scope: "repository", Kind: "file", Present: true},
			{ID: "historical", Path: stored, Scope: "repository", Kind: "file", Present: false},
		}},
	}
	inventory.rebuild()

	record, actual := inventory.exact(path)
	if record == nil {
		t.Fatal("present record was hidden by historical duplicate")
	}
	if record.ID != "present" {
		t.Fatalf("record ID = %q, want present", record.ID)
	}
	if filepath.Clean(actual) != filepath.Clean(path) {
		t.Fatalf("actual path = %q, want %q", actual, path)
	}

	resolved, actual, err := inventory.ensureTarget(path, "historical")
	if err != nil {
		t.Fatal(err)
	}
	if resolved == nil || resolved.ID != "present" {
		t.Fatalf("resolved record = %#v, want present record", resolved)
	}
	if filepath.Clean(actual) != filepath.Clean(path) {
		t.Fatalf("actual path = %q, want %q", actual, path)
	}
}
