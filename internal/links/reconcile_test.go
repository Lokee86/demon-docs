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

func TestNestedWorktreeDirectoriesAreExcludedFromRepositoryInventory(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[nested](.worktrees/branch-a/nested.md)\n")
	writeTestFile(t, filepath.Join(root, ".worktrees", "branch-a", "nested.md"), "# Nested\n")
	writeTestFile(t, filepath.Join(root, ".workingtrees", "branch-b", "nested.md"), "# Nested\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range plan.Files.Files {
		if strings.HasPrefix(filepath.ToSlash(record.Path), ".worktrees/") ||
			strings.HasPrefix(filepath.ToSlash(record.Path), ".workingtrees/") {
			t.Fatalf("nested worktree path entered repository inventory: %#v", record)
		}
	}
	if len(plan.Links.Links) != 0 {
		t.Fatalf("nested worktree source or target entered link state: %#v", plan.Links.Links)
	}
}

func TestLinkedWorktreeCanBeTheRepositoryRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, ".worktrees", "branch-a")
	writeTestFile(t, filepath.Join(root, "README.md"), "# Linked checkout\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, record := range plan.Files.Files {
		if record.Path == "README.md" && record.Present {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("linked-worktree repository root was excluded: %#v", plan.Files.Files)
	}
}

func TestPruneNestedWorktreeStateDropsExcludedSourcesButKeepsAffectedAuthoredLinks(t *testing.T) {
	root := t.TempDir()
	files := FilesManifest{SchemaVersion: schemaVersion, Files: []FileRecord{
		{ID: "normal", Path: "README.md", Scope: "repository", Kind: "file", Present: true},
		{ID: "nested", Path: ".worktrees/branch-a/README.md", Scope: "repository", Kind: "file", Present: true},
	}}
	links := LinksManifest{SchemaVersion: schemaVersion, Links: []LinkRecord{
		{ID: "nested-source", SourceFileID: "nested", TargetFileID: "normal"},
		{ID: "normal-source", SourceFileID: "normal", TargetFileID: "nested"},
	}}

	keptFiles, keptLinks := pruneNestedWorktreeState(root, files, links)
	if len(keptFiles.Files) != 1 || keptFiles.Files[0].ID != "normal" {
		t.Fatalf("unexpected retained files: %#v", keptFiles.Files)
	}
	if len(keptLinks.Links) != 1 || keptLinks.Links[0].ID != "normal-source" {
		t.Fatalf("affected authored source link must remain to force reparsing: %#v", keptLinks.Links)
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

func TestHTMLAndWikiLinksRepairMovedTargets(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "<a href=\"docs/guide.md#part\">Guide</a>\n[[docs/guide|Guide]]\n")
	writeTestFile(t, filepath.Join(root, "docs", "guide.md"), "# Guide\n")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "manual"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, "docs", "guide.md"), filepath.Join(root, "manual", "guide.md")); err != nil {
		t.Fatal(err)
	}

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 1 {
		t.Fatalf("updates = %d, want 1; messages=%v", len(second.Updates), second.Messages)
	}
	updated := second.Updates[0].NewText
	if !strings.Contains(updated, "manual/guide.md#part") {
		t.Fatalf("HTML link was not repaired: %q", updated)
	}
	if !strings.Contains(updated, "[[manual/guide|Guide]]") {
		t.Fatalf("wiki link style was not preserved: %q", updated)
	}
}

func TestUndefinedReferenceLabelIsUnresolved(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[Guide][missing]\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	assertUndefinedReference(t, plan)
	if err := Save(plan); err != nil {
		t.Fatal(err)
	}
	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	assertUndefinedReference(t, second)
}

func TestBareWikiLinkResolvesUniqueRepositoryNote(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[[Design]]\n")
	writeTestFile(t, filepath.Join(root, "docs", "Design.md"), "# Design\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 0 || len(plan.Links.Links) != 1 || plan.Links.Links[0].Status != "valid" {
		t.Fatalf("bare wiki link was not resolved: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
}

func TestBareWikiFolderMoveUpdatesGraphWithoutRewritingSource(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[[Guide]]\n")
	writeTestFile(t, filepath.Join(root, "docs", "Guide.md"), "# Guide\n")

	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "manual"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, "docs", "Guide.md"), filepath.Join(root, "manual", "Guide.md")); err != nil {
		t.Fatal(err)
	}

	second, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Updates) != 0 || len(second.Links.Links) != 1 {
		t.Fatalf("folder-only wiki move should update graph only: %#v", second)
	}
	if second.Links.Links[0].ResolvedPath != "manual/Guide.md" || second.Links.Links[0].Status != "valid" {
		t.Fatalf("wiki graph target was not refreshed: %#v", second.Links.Links[0])
	}
}

func TestBareWikiLinkReportsAmbiguousRepositoryNotes(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[[Design]]\n")
	writeTestFile(t, filepath.Join(root, "one", "Design.md"), "# One\n")
	writeTestFile(t, filepath.Join(root, "two", "Design.md"), "# Two\n")

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Unresolved != 1 || len(plan.Links.Links) != 1 || plan.Links.Links[0].Status != "ambiguous" {
		t.Fatalf("ambiguous wiki link was not reported: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
}

func assertUndefinedReference(t *testing.T, plan Plan) {
	t.Helper()
	if plan.Unresolved != 1 || len(plan.Links.Links) != 1 {
		t.Fatalf("unexpected plan: unresolved=%d links=%#v messages=%v", plan.Unresolved, plan.Links.Links, plan.Messages)
	}
	if plan.Links.Links[0].Status != "undefined_reference" {
		t.Fatalf("status = %q", plan.Links.Links[0].Status)
	}
	found := false
	for _, message := range plan.Messages {
		if strings.Contains(message, "Undefined reference label") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing diagnostic: %v", plan.Messages)
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
