package reconcile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestScopedConvergeUpdatesBothSidesOfFileMove(t *testing.T) {
	root := t.TempDir()
	c := config.Default()
	oldPath := filepath.Join(root, "one", "topic.md")
	newPath := filepath.Join(root, "two", "topic.md")
	write(t, oldPath, "# Topic\n")
	write(t, filepath.Join(root, "two", "keep.md"), "# Keep\n")
	if _, _, err := ConvergeWithin(root, root, c); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ConvergeScopedWithin(root, root, c, []string{oldPath, newPath}); err != nil {
		t.Fatal(err)
	}
	oldIndex, err := os.ReadFile(filepath.Join(root, "one", c.IndexFile))
	if err != nil {
		t.Fatal(err)
	}
	newIndex, err := os.ReadFile(filepath.Join(root, "two", c.IndexFile))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(oldIndex), "topic.md") || !strings.Contains(string(newIndex), "topic.md") {
		t.Fatalf("scoped move did not update both folders: old=%q new=%q", oldIndex, newIndex)
	}
}

func TestScopedTreePlansOnlyAffectedIndexFolders(t *testing.T) {
	root := t.TempDir()
	c := config.Default()
	one := filepath.Join(root, "one")
	two := filepath.Join(root, "two")
	if err := os.MkdirAll(one, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(two, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(one, "a.md"), "# A\n")
	write(t, filepath.Join(two, "b.md"), "# B\n")
	if _, _, err := ConvergeWithin(root, root, c); err != nil {
		t.Fatal(err)
	}

	stalePath := filepath.Join(two, "b.md")
	if err := os.Remove(stalePath); err != nil {
		t.Fatal(err)
	}
	changedPath := filepath.Join(one, "c.md")
	write(t, changedPath, "# C\n")

	scoped, err := TreeScopedWithIgnoreRoot(root, root, c, []string{changedPath})
	if err != nil {
		t.Fatal(err)
	}
	for _, update := range scoped.Updates {
		if filepath.Clean(update.Path) == filepath.Clean(filepath.Join(two, c.IndexFile)) {
			t.Fatalf("scoped reconciliation touched unrelated stale index: %#v", scoped.Updates)
		}
	}

	full, err := TreeWithIgnoreRoot(root, root, c)
	if err != nil {
		t.Fatal(err)
	}
	foundStale := false
	for _, update := range full.Updates {
		if filepath.Clean(update.Path) == filepath.Clean(filepath.Join(two, c.IndexFile)) {
			foundStale = true
			break
		}
	}
	if !foundStale {
		t.Fatalf("fixture did not contain unrelated stale index: %#v", full.Updates)
	}
}
