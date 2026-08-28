package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopedReconcileRecoversFingerprintWithoutTargetFileID(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "docs", "source.md")
	unrelated := filepath.Join(root, "docs", "unrelated.md")
	oldTarget := filepath.Join(root, "docs", "old", "guide.md")
	writeTestFile(t, source, "[Guide](old/guide.md)\n")
	writeTestFile(t, unrelated, "# Unrelated\n")
	writeTestFile(t, oldTarget, "# Guide\nunique fingerprint payload\n")
	writeTestFile(t, filepath.Join(root, "docs", "other", "guide.md"), "# Decoy\ndifferent content\n")
	if err := os.MkdirAll(filepath.Join(root, "docs", "archive"), 0o755); err != nil {
		t.Fatal(err)
	}

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if baseline.Unresolved != 0 || len(baseline.Links.Links) != 1 {
		t.Fatalf("baseline=%#v", baseline)
	}
	baseline.Links.Links[0].TargetFileID = ""
	if err := Save(baseline); err != nil {
		t.Fatal(err)
	}

	newTarget := filepath.Join(root, "docs", "archive", "guide.md")
	if err := os.Rename(oldTarget, newTarget); err != nil {
		t.Fatal(err)
	}

	plan, err := ReconcileChangedPaths(root, []string{oldTarget, newTarget})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 0 || len(plan.Updates) != 1 {
		t.Fatalf("scoped fingerprint repair failed: unresolved=%d updates=%d messages=%v", plan.Unresolved, len(plan.Updates), plan.Messages)
	}
	if !strings.Contains(plan.Updates[0].NewText, "archive/guide.md") {
		t.Fatalf("wrong repair: %q", plan.Updates[0].NewText)
	}
	if len(plan.Links.Links) != 1 || plan.Links.Links[0].TargetFileID == "" {
		t.Fatalf("target identity was not restored: %#v", plan.Links.Links)
	}
}

func TestScopedReconcileFallsBackWhenBatchMissesFilesystemChange(t *testing.T) {
	root := t.TempDir()
	changed := filepath.Join(root, "docs", "changed.md")
	missed := filepath.Join(root, "docs", "missed.md")
	writeTestFile(t, changed, "# Changed\n")
	writeTestFile(t, missed, "# Missed\n")
	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(baseline); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, changed, "# Changed again\n")
	writeTestFile(t, missed, "# Missed again\n")
	if _, err := ReconcileChangedPaths(root, []string{changed}); err != ErrScopedReconciliationUnavailable {
		t.Fatalf("incomplete batch error=%v", err)
	}
}

func TestScopedReconcileFallsBackForDirectoryBatch(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "docs")
	writeTestFile(t, filepath.Join(directory, "source.md"), "# Source\n")
	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(baseline); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileChangedPaths(root, []string{directory}); err != ErrScopedReconciliationUnavailable {
		t.Fatalf("directory batch error=%v", err)
	}
}
