package reconcile

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/model"
)

func plannedText(t *testing.T, result model.ReconcileResult, path string) string {
	t.Helper()
	for _, update := range result.Updates {
		if update.Path == path {
			return update.NewText
		}
	}
	t.Fatalf("no update planned for %s: %+v", path, result.Updates)
	return ""
}

func plannedUpdate(result model.ReconcileResult, path string) *model.FileUpdate {
	for i := range result.Updates {
		if result.Updates[i].Path == path {
			return &result.Updates[i]
		}
	}
	return nil
}

func managedIndex(title, files, stubs, folders string) string {
	return "# " + title + "\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n" + files + "<!-- doc-ledger:files:end -->\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n" + stubs + "<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n" + folders + "<!-- doc-ledger:folders:end -->"
}

func requireContains(t *testing.T, text string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(text, value) {
			t.Errorf("missing %q:\n%s", value, text)
		}
	}
}

func TestIndexCreationConfiguredNamesAndParentLinks(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	write(t, filepath.Join(root, "page.md"), "# Page")
	write(t, filepath.Join(root, "guide", "topic.md"), "# Topic")
	c := config.Default()
	c.IndexFile, c.Files.IndexFile = "!INDEX.md", "!INDEX.md"
	c.ParentLink.IndexedFiles = true
	result, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	rootIndex := filepath.Join(root, "!INDEX.md")
	childIndex := filepath.Join(root, "guide", "!INDEX.md")
	if update := plannedUpdate(result, rootIndex); update == nil || update.OldText != nil {
		t.Fatalf("missing root index creation: %+v", update)
	}
	if update := plannedUpdate(result, childIndex); update == nil || update.OldText != nil {
		t.Fatalf("missing child index creation: %+v", update)
	}
	requireContains(t, plannedText(t, result, rootIndex), "[page.md](page.md)", "[guide](guide/!INDEX.md)")
	requireContains(t, plannedText(t, result, childIndex), "Parent index: [Docs](../!INDEX.md)", "[topic.md](topic.md)")
	requireContains(t, plannedText(t, result, filepath.Join(root, "page.md")), "Parent index: [Docs](./!INDEX.md)")
}

func TestNonMarkdownAndConfiguredEditableExtensions(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	binary := filepath.Join(root, "diagram.png")
	write(t, binary, "\x89PNG\r\n\x1a\nbytes")
	mdx := filepath.Join(root, "page.mdx")
	write(t, mdx, "# Page")
	write(t, filepath.Join(root, "stubs", "draft.pdf"), "pdf bytes")
	c := config.Default()
	c.Files.IncludePatterns = []string{"**/*.md", "**/*.mdx", "**/*.png", "**/*.pdf"}
	c.Files.EditableParentIndexExtensions = []string{".md", ".mdx"}
	c.ParentLink.IndexedFiles = true
	result, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	index := plannedText(t, result, filepath.Join(root, "README.md"))
	requireContains(t, index, "[diagram.png](diagram.png)", "[page.mdx](page.mdx)", "[draft.pdf](stubs/draft.pdf)")
	if plannedUpdate(result, binary) != nil || plannedUpdate(result, filepath.Join(root, "stubs", "draft.pdf")) != nil {
		t.Fatal("non-editable included file was rewritten")
	}
	requireContains(t, plannedText(t, result, mdx), "Parent index: [Docs](./README.md)")
}

func TestCustomRenderingConfiguration(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	write(t, filepath.Join(root, "alpha.md"), "# Alpha")
	write(t, filepath.Join(root, "ideas", "draft.md"), "# Draft")
	write(t, filepath.Join(root, "guide", "topic.md"), "# Topic")
	c := config.Default()
	c.Markers.Prefix = "nav"
	c.Sections.FilesHeading, c.Sections.StubsHeading, c.Sections.FoldersHeading = "Pages", "Ideas", "Areas"
	c.Draft.Folder, c.Draft.DescriptionPrefix = "ideas", "Draft: "
	c.Description.FileTemplate, c.Description.FolderTemplate = "File: {title}.", "Folder: {title}."
	c.ParentLink.Label, c.ParentLink.IndexedFiles = "Up", true
	result, err := Tree(root, c)
	if err != nil {
		t.Fatal(err)
	}
	index := plannedText(t, result, filepath.Join(root, "README.md"))
	requireContains(t, index, "## Pages", "## Ideas", "## Areas", "<!-- nav:files:start -->", "File: Alpha.", "Draft: File: Draft.", "Folder: Guide.")
	requireContains(t, plannedText(t, result, filepath.Join(root, "alpha.md")), "Up: [Docs](./README.md)")
}
