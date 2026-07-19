package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstScanRecordsOnlyThenRepairsMovedNonMarkdownTarget(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[asset](assets/picture.png#preview)\n")
	writeTestFile(t, filepath.Join(root, "assets", "picture.png"), "image-data")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if !first.NeedsInitialization || len(first.Updates) != 0 {
		t.Fatalf("first scan should only establish state: %#v", first)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "media"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, "assets", "picture.png"), filepath.Join(root, "media", "picture.png")); err != nil {
		t.Fatal(err)
	}

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 1 {
		t.Fatalf("got %d updates, want 1; messages=%v", len(second.Updates), second.Messages)
	}
	if !strings.Contains(second.Updates[0].NewText, "media/picture.png#preview") {
		t.Fatalf("link was not repaired: %q", second.Updates[0].NewText)
	}
}

func TestDocignoreExcludesLinkTargets(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, ".docignore"), "ignored.bin\n")
	writeTestFile(t, filepath.Join(root, "README.md"), "[ignored](ignored.bin)\n")
	writeTestFile(t, filepath.Join(root, "ignored.bin"), "ignored")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Links.Links) != 0 {
		t.Fatalf("ignored link target was tracked: %#v", plan.Links.Links)
	}
}

func TestDirectoryTargetsKeepStableIdentity(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[assets](assets/)\n")
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Links.Links) != 1 || first.Links.Links[0].TargetFileID == "" {
		t.Fatalf("directory target identity was not recorded: %#v", first.Links.Links)
	}
	identity := first.Links.Links[0].TargetFileID
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Links.Links) != 1 || second.Links.Links[0].TargetFileID != identity {
		t.Fatalf("directory target identity changed: before=%q after=%#v", identity, second.Links.Links)
	}
}

func TestBrokenLinkGuessWaitsUntilAfterInitialScan(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[manual](old/manual.pdf)\n")
	writeTestFile(t, filepath.Join(root, "references", "manual.pdf"), "manual")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Updates) != 0 || first.Unresolved != 1 {
		t.Fatalf("initial scan guessed a repair: updates=%d unresolved=%d", len(first.Updates), first.Unresolved)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 1 || !strings.Contains(second.Updates[0].NewText, "references/manual.pdf") {
		t.Fatalf("second scan did not apply unique guess: %#v", second)
	}
}

func TestAmbiguousGuessIsLeftForTheUser(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[manual](old/manual.pdf)\n")
	writeTestFile(t, filepath.Join(root, "one", "manual.pdf"), "one")
	writeTestFile(t, filepath.Join(root, "two", "manual.pdf"), "two")

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
	if len(second.Updates) != 0 || second.Unresolved != 1 {
		t.Fatalf("ambiguous link should remain unchanged: %#v", second)
	}
	if len(second.Links.Links) != 1 || second.Links.Links[0].Status != "ambiguous" || len(second.Links.Links[0].Candidates) != 2 {
		t.Fatalf("ambiguous candidates were not recorded: %#v", second.Links.Links)
	}
}

func TestAbsoluteExternalTargetCanMoveIntoRepository(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	oldTarget := filepath.Join(external, "guide.bin")
	writeTestFile(t, oldTarget, "guide-data")
	writeTestFile(t, filepath.Join(root, "README.md"), "[guide](<"+filepath.ToSlash(oldTarget)+">)\n")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	newTarget := filepath.Join(root, "assets", "renamed-guide.bin")
	if err := os.MkdirAll(filepath.Dir(newTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldTarget, newTarget); err != nil {
		t.Fatal(err)
	}

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 1 {
		t.Fatalf("absolute external move was not repaired: %#v", second)
	}
	if !strings.Contains(second.Updates[0].NewText, filepath.ToSlash(newTarget)) {
		t.Fatalf("absolute link style was not preserved: %q", second.Updates[0].NewText)
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
