package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepairObservedRenameRewritesInboundLinksFromState(t *testing.T) {
	root := t.TempDir()
	oldPath := filepath.Join(root, "target.md")
	newPath := filepath.Join(root, "renamed.md")
	writeTestFile(t, oldPath, "# Target\n")
	writeTestFile(t, filepath.Join(root, "one.md"), "[Target](target.md)\n")
	writeTestFile(t, filepath.Join(root, "two.md"), "[Target](target.md#section)\n")
	writeTestFile(t, filepath.Join(root, "!INDEX.md"), "# Index\n\n<!-- doc-ledger:files:start -->\n- [target.md](target.md) - Target documentation.\n<!-- doc-ledger:files:end -->\n")
	initializeRenameState(t, root)
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}

	handled, changed, err := RepairObservedRename(root, oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("rename was not handled")
	}
	if changed != 3 {
		t.Fatalf("changed=%d want=3", changed)
	}
	for _, name := range []string{"one.md", "two.md", "!INDEX.md"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, "target.md") || !strings.Contains(text, "renamed.md") {
			t.Fatalf("%s was not rewritten: %q", name, text)
		}
	}
	indexText, err := os.ReadFile(filepath.Join(root, "!INDEX.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(indexText), "[renamed.md](renamed.md)") {
		t.Fatalf("managed index label did not follow rename: %q", string(indexText))
	}
	followup, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(followup.Rewrites) != 0 || len(followup.Updates) != 0 {
		t.Fatalf("follow-up rewrites=%d updates=%d", len(followup.Rewrites), len(followup.Updates))
	}
}

func TestRepairObservedRenameFallsBackForChangedContent(t *testing.T) {
	root, oldPath := initializedRenameFixture(t)
	newPath := filepath.Join(root, "renamed.md")
	if err := os.WriteFile(oldPath, []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	handled, changed, err := RepairObservedRename(root, oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if handled || changed != 0 {
		t.Fatalf("handled=%v changed=%d", handled, changed)
	}
}

func TestRepairObservedRenameFallsBackForChangedSource(t *testing.T) {
	root, oldPath := initializedRenameFixture(t)
	newPath := filepath.Join(root, "renamed.md")
	if err := os.WriteFile(filepath.Join(root, "source.md"), []byte("prefix [Target](target.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	handled, changed, err := RepairObservedRename(root, oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if handled || changed != 0 {
		t.Fatalf("handled=%v changed=%d", handled, changed)
	}
}

func TestRepairObservedRenameFallsBackAcrossDirectories(t *testing.T) {
	root, oldPath := initializedRenameFixture(t)
	if err := os.Mkdir(filepath.Join(root, "other"), 0o755); err != nil {
		t.Fatal(err)
	}
	newPath := filepath.Join(root, "other", "renamed.md")
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
	handled, changed, err := RepairObservedRename(root, oldPath, newPath)
	if err != nil {
		t.Fatal(err)
	}
	if handled || changed != 0 {
		t.Fatalf("handled=%v changed=%d", handled, changed)
	}
}

func initializedRenameFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	oldPath := filepath.Join(root, "target.md")
	writeTestFile(t, oldPath, "# Target\n")
	writeTestFile(t, filepath.Join(root, "source.md"), "[Target](target.md)\n")
	initializeRenameState(t, root)
	return root, oldPath
}

func initializeRenameState(t *testing.T, root string) {
	t.Helper()
	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	ready, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&ready); err != nil {
		t.Fatal(err)
	}
}
