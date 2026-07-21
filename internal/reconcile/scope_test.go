package reconcile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/model"
)

func TestApplyWithinRejectsOutsideWriteBeforeMutation(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "inside.md")
	outside := filepath.Join(t.TempDir(), "outside.md")
	result := model.ReconcileResult{Updates: []model.FileUpdate{
		{Path: inside, NewText: "inside"},
		{Path: outside, NewText: "outside"},
	}}

	if _, err := ApplyWithin(result, root); err == nil {
		t.Fatal("outside write was accepted")
	}
	if _, err := os.Stat(inside); !os.IsNotExist(err) {
		t.Fatal("inside update was written before boundary validation completed")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatal("outside update was written")
	}
}

func TestPrepareMissingWithinWritesPlannedContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "area", "INDEX.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	result := model.ReconcileResult{Updates: []model.FileUpdate{{
		Path:    path,
		NewText: "# Area\n\nPrepared index.\n",
	}}}

	if err := PrepareMissingWithin(result, root); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	if normalized != "# Area\n\nPrepared index.\n" {
		t.Fatalf("prepared content = %q", content)
	}
}

func TestPrepareMissingWithinDoesNotRecreateMovedDirectory(t *testing.T) {
	root := t.TempDir()
	oldDirectory := filepath.Join(root, "old")
	newDirectory := filepath.Join(root, "new")
	if err := os.Mkdir(oldDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldDirectory, newDirectory); err != nil {
		t.Fatal(err)
	}

	result := model.ReconcileResult{Updates: []model.FileUpdate{{
		Path:    filepath.Join(oldDirectory, "INDEX.md"),
		NewText: "# Old\n",
	}}}
	if err := PrepareMissingWithin(result, root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldDirectory); !os.IsNotExist(err) {
		t.Fatalf("stale source directory was recreated: %v", err)
	}
}

func TestApplyWithinDoesNotRecreateMovedDirectory(t *testing.T) {
	root := t.TempDir()
	oldDirectory := filepath.Join(root, "old")
	newDirectory := filepath.Join(root, "new")
	if err := os.Mkdir(oldDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldDirectory, newDirectory); err != nil {
		t.Fatal(err)
	}

	result := model.ReconcileResult{Updates: []model.FileUpdate{{
		Path:    filepath.Join(oldDirectory, "INDEX.md"),
		NewText: "# Old\n",
	}}}
	changed, err := ApplyWithin(result, root)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 0 {
		t.Fatalf("changed = %d, want 0", changed)
	}
	if _, err := os.Stat(oldDirectory); !os.IsNotExist(err) {
		t.Fatalf("stale source directory was recreated: %v", err)
	}
}

func TestApplyWithinSkipsChangedExistingFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "INDEX.md")
	if err := os.WriteFile(path, []byte("newer\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldText := "older\n"
	result := model.ReconcileResult{Updates: []model.FileUpdate{{
		Path:    path,
		OldText: &oldText,
		NewText: "planned\n",
	}}}

	changed, err := ApplyWithin(result, root)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 0 {
		t.Fatalf("changed = %d, want 0", changed)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "newer\n" {
		t.Fatalf("content = %q, want newer content preserved", content)
	}
}
