package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyGeneratedPreflightFailurePreventsAllWrites(t *testing.T) {
	root := t.TempDir()
	createBenchmarkRepository(t, root, 32)

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, "asset-a.bin"), filepath.Join(root, "asset-moved.bin")); err != nil {
		t.Fatal(err)
	}
	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) < 2 {
		t.Fatalf("rewrites = %d, want at least 2", len(plan.Rewrites))
	}

	unchangedPath := plan.Rewrites[0].Path
	changedPath := plan.Rewrites[len(plan.Rewrites)-1].Path
	userEdit := "user edit after reconciliation\n"
	if err := os.WriteFile(changedPath, []byte(userEdit), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyGenerated(plan.Rewrites); err == nil {
		t.Fatal("preflight accepted a changed source")
	}

	unchanged, err := os.ReadFile(unchangedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unchanged), "../asset-a.bin") {
		t.Fatalf("another source was written before the preflight barrier: %q", unchanged)
	}
	changed, err := os.ReadFile(changedPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(changed) != userEdit {
		t.Fatalf("changed source was overwritten: %q", changed)
	}
}

func TestApplyGeneratedPreservesSuppressionOrder(t *testing.T) {
	root := t.TempDir()
	createBenchmarkRepository(t, root, 32)

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, "asset-a.bin"), filepath.Join(root, "asset-moved.bin")); err != nil {
		t.Fatal(err)
	}
	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}

	suppressions, err := ApplyGenerated(plan.Rewrites)
	if err != nil {
		t.Fatal(err)
	}
	if len(suppressions) != len(plan.Rewrites) {
		t.Fatalf("suppressions = %d, rewrites = %d", len(suppressions), len(plan.Rewrites))
	}
	for index := range suppressions {
		if suppressions[index].SourceFileID != plan.Rewrites[index].SourceFileID || suppressions[index].Path != plan.Rewrites[index].Path {
			t.Fatalf("suppression %d is out of order: %#v for rewrite %#v", index, suppressions[index], plan.Rewrites[index])
		}
	}
}
