package reconcile

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestStaleEntryRemovalPreservesUnmanagedContent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	readme := filepath.Join(root, "INDEX.md")
	write(t, readme, managedIndex("Docs", "- [gone.md](gone.md) - Gone.\n", "- [stale.md](stubs/stale.md) - Stub: Gone.\n", "- [Gone](gone/INDEX.md) - Gone.\n")+"\n\n## Notes\n\nKeep this note.")
	result, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Messages) != 3 {
		t.Fatalf("messages=%v", result.Messages)
	}
	for _, section := range []string{"files", "folders", "stubs"} {
		found := false
		for _, message := range result.Messages {
			found = found || strings.Contains(message, "Removed stale "+section+" entry")
		}
		if !found {
			t.Errorf("missing stale %s message: %v", section, result.Messages)
		}
	}
	text := plannedText(t, result, readme)
	if strings.Contains(text, "gone.md") || strings.Contains(text, "stale.md") || strings.Contains(text, "gone/INDEX.md") {
		t.Fatalf("stale entry remained:\n%s", text)
	}
	requireContains(t, text, "## Notes", "Keep this note.")
}

func TestMalformedBlockRepairStopsBeforeFollowingSection(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	readme := filepath.Join(root, "INDEX.md")
	write(t, filepath.Join(root, "alpha.md"), "# Alpha")
	write(t, readme, "# Docs\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n- [alpha.md](alpha.md) - Custom alpha description.\n\n## Notes\n\nKeep this note.")
	result, err := Tree(root, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	text := plannedText(t, result, readme)
	if strings.Count(text, "<!-- doc-ledger:files:end -->") != 1 {
		t.Fatalf("missing or duplicate repaired marker:\n%s", text)
	}
	requireContains(t, text, "Custom alpha description.", "## Notes\n\nKeep this note.")
}

func TestParentLinksCanBeRemovedWithoutChangingBody(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	child := filepath.Join(root, "guide", "INDEX.md")
	write(t, filepath.Join(root, "INDEX.md"), "# Documentation")
	write(t, child, "# Guide\n\nParent index: [Wrong](../INDEX.md)\n\nGuide body")
	c := config.Default()
	c.ParentLink.FolderIndexes = false
	result, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	text := plannedText(t, result, child)
	if strings.Contains(text, "Parent index:") {
		t.Fatalf("disabled parent link remained:\n%s", text)
	}
	requireContains(t, text, "# Guide", "Guide body")
}

func TestExistingDescriptionsAndRootDisplayTitleRemainStable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	readme := filepath.Join(root, "INDEX.md")
	write(t, readme, managedIndex("Documentation", "- [alpha.md](alpha.md) - Custom alpha description.\n", "", ""))
	write(t, filepath.Join(root, "alpha.md"), "Parent index: [Docs](./INDEX.md)\n\nAlpha body")
	write(t, filepath.Join(root, "guide", "topic.md"), "# Topic")
	c := config.Default()
	c.Description.FileTemplate = "File: {title}."
	c.ParentLink.IndexedFiles = true
	result, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	index := plannedText(t, result, readme)
	requireContains(t, index, "Custom alpha description.")
	if strings.Contains(index, "File: Alpha.") {
		t.Fatalf("configured fallback replaced stable description:\n%s", index)
	}
	child := plannedText(t, result, filepath.Join(root, "guide", "INDEX.md"))
	requireContains(t, child, "Parent index: [Docs](../INDEX.md)")
	if strings.Contains(child, "[Documentation]") {
		t.Fatalf("root display title changed despite retained child title:\n%s", child)
	}
}
