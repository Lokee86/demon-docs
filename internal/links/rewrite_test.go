package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalMoveRewritePersistsGraphAndSuppressesWatcherEvent(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "README.md")
	oldTarget := filepath.Join(root, "assets", "guide.pdf")
	writeTestFile(t, sourcePath, "[guide](assets/guide.pdf#start)\n")
	writeTestFile(t, oldTarget, "guide")

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	newTarget := filepath.Join(root, "manuals", "guide.pdf")
	if err := os.MkdirAll(filepath.Dir(newTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldTarget, newTarget); err != nil {
		t.Fatal(err)
	}

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) != 1 {
		t.Fatalf("generated rewrites = %d, want 1", len(plan.Rewrites))
	}
	if len(plan.Rewrites[0].Transformations) != 1 {
		t.Fatalf("transformations = %#v", plan.Rewrites[0].Transformations)
	}
	if _, err := ApplyAndSave(&plan); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); !strings.Contains(got, "manuals/guide.pdf#start") {
		t.Fatalf("source was not rewritten: %q", got)
	}
	matched, err := ConsumePendingSuppression(root, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatal("internally generated event was not suppressed")
	}
	matched, err = ConsumePendingSuppression(root, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if matched {
		t.Fatal("suppression was not consumed")
	}

	_, graph, initialized, err := loadState(root)
	if err != nil {
		t.Fatal(err)
	}
	if !initialized || len(graph.Links) != 1 || graph.Links[0].RawPath != "manuals/guide.pdf" {
		t.Fatalf("persisted graph = %#v initialized=%v", graph.Links, initialized)
	}
}

func TestInternalMoveRewriteReportsUnresolvedLinksInTheSameSource(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "README.md")
	oldTarget := filepath.Join(root, "old", "asset.bin")
	writeTestFile(t, sourcePath, "[asset](old/asset.bin)\n[missing](missing/nowhere.bin)\n")
	writeTestFile(t, oldTarget, "asset")

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	newTarget := filepath.Join(root, "new", "asset.bin")
	if err := os.MkdirAll(filepath.Dir(newTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldTarget, newTarget); err != nil {
		t.Fatal(err)
	}

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) != 1 {
		t.Fatalf("generated rewrites = %d, want 1", len(plan.Rewrites))
	}
	if plan.Unresolved != 1 {
		t.Fatalf("unresolved = %d, want 1; messages=%v", plan.Unresolved, plan.Messages)
	}
	foundBroken := false
	for _, message := range plan.Messages {
		if strings.Contains(message, "Broken link") && strings.Contains(message, "missing/nowhere.bin") {
			foundBroken = true
			break
		}
	}
	if !foundBroken {
		t.Fatalf("missing unresolved diagnostic: %v", plan.Messages)
	}
}

func TestInternalMoveRewriteRejectsConcurrentUserEdit(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "README.md")
	oldTarget := filepath.Join(root, "old", "asset.bin")
	writeTestFile(t, sourcePath, "[asset](old/asset.bin)\n")
	writeTestFile(t, oldTarget, "asset")

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	newTarget := filepath.Join(root, "new", "asset.bin")
	if err := os.MkdirAll(filepath.Dir(newTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldTarget, newTarget); err != nil {
		t.Fatal(err)
	}
	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) != 1 {
		t.Fatalf("generated rewrites = %d, want 1", len(plan.Rewrites))
	}
	userEdit := "User changed this before Demon Docs wrote it.\n"
	if err := os.WriteFile(sourcePath, []byte(userEdit), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&plan); err == nil {
		t.Fatal("concurrent user edit was not rejected")
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != userEdit {
		t.Fatalf("concurrent user edit was overwritten: %q", data)
	}
}

func TestUserEditedMarkdownReplacesStoredOutgoingLinks(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "README.md")
	writeTestFile(t, sourcePath, "[one](one.bin)\n")
	writeTestFile(t, filepath.Join(root, "one.bin"), "one")
	writeTestFile(t, filepath.Join(root, "two.bin"), "two")

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("[two](two.bin)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) != 0 {
		t.Fatalf("user-authored valid edit produced generated rewrites: %d", len(plan.Rewrites))
	}
	if _, err := ApplyAndSave(&plan); err != nil {
		t.Fatal(err)
	}
	_, graph, _, err := loadState(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Links) != 1 || graph.Links[0].Target != "two.bin" {
		t.Fatalf("stored outgoing links = %#v", graph.Links)
	}
}
