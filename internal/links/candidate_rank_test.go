package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathAwareRepairNarrowsRepeatedIndexBasenames(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "docs", "services", "game-server", "integrations", "auth.md")
	writeTestFile(t, source, "[Player data](../../../player-data/!INDEX.md)\n")
	writeTestFile(t, filepath.Join(root, "docs", "services", "player-data", "!INDEX.md"), "# Player data\n")
	writeTestFile(t, filepath.Join(root, "docs", "planning", "services", "player-data", "!INDEX.md"), "# Planned player data\n")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if second.Unresolved != 0 || len(second.Updates) != 1 {
		t.Fatalf("path-aware repair failed: unresolved=%d updates=%d messages=%v", second.Unresolved, len(second.Updates), second.Messages)
	}
	if !strings.Contains(second.Updates[0].NewText, "../../player-data/!INDEX.md") {
		t.Fatalf("repair chose the wrong repeated basename: %q", second.Updates[0].NewText)
	}
}

func TestPathAwareRepairLeavesEqualPathMatchesAmbiguous(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "docs", "current", "source.md"), "[Index](../missing/!INDEX.md)\n")
	writeTestFile(t, filepath.Join(root, "docs", "one", "missing", "!INDEX.md"), "# One\n")
	writeTestFile(t, filepath.Join(root, "docs", "two", "missing", "!INDEX.md"), "# Two\n")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if second.Unresolved != 1 || len(second.Updates) != 0 {
		t.Fatalf("equal path matches should remain unresolved: %#v", second)
	}
	if len(second.Links.Links) != 1 || second.Links.Links[0].Status != "ambiguous" || len(second.Links.Links[0].Candidates) != 2 {
		t.Fatalf("equal candidates were not preserved: %#v", second.Links.Links)
	}
}

func TestFingerprintIdentityResolvesWithoutTargetFileIDBeforePathRanking(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "docs", "current", "source.md")
	originalTarget := filepath.Join(root, "docs", "missing", "!INDEX.md")
	writeTestFile(t, source, "[Index](../missing/!INDEX.md)\n")
	writeTestFile(t, originalTarget, "# Intended\nunique fingerprint payload\n")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if first.Unresolved != 0 || len(first.Links.Links) != 1 {
		t.Fatalf("baseline did not resolve exactly: %#v", first)
	}
	first.Links.Links[0].TargetFileID = ""
	if err := Save(first); err != nil {
		t.Fatal(err)
	}

	intended := filepath.Join(root, "docs", "one", "missing", "!INDEX.md")
	if err := os.MkdirAll(filepath.Dir(intended), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(originalTarget, intended); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "docs", "two", "missing", "!INDEX.md"), "# Decoy\ndifferent payload\n")

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if second.Unresolved != 0 || len(second.Updates) != 1 {
		t.Fatalf("fingerprint identity did not resolve adversarial move: unresolved=%d updates=%d messages=%v links=%#v", second.Unresolved, len(second.Updates), second.Messages, second.Links.Links)
	}
	if len(second.Links.Links) != 1 || second.Links.Links[0].TargetFileID == "" {
		t.Fatalf("recovered target identity was not persisted: %#v", second.Links.Links)
	}
	if !strings.Contains(second.Updates[0].NewText, "../one/missing/!INDEX.md") {
		t.Fatalf("fingerprint resolution chose the wrong target: %q", second.Updates[0].NewText)
	}
}
