package markdown

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestFrontmatterIsInvisibleToMarkdownStructure(t *testing.T) {
	for name, source := range map[string]string{
		"yaml": "---\nsummary: A managed document.\n---\n# Real Title\n\nBody.\n",
		"toml": "+++\nsummary = \"A managed document.\"\n+++\n# Real Title\n\nBody.\n",
	} {
		t.Run(name, func(t *testing.T) {
			if got := FirstHeadingTitle(source); got != "Real Title" {
				t.Fatalf("title = %q", got)
			}
			updated := UpdateParent(source, "Parent index: [Docs](./README.md)", "Parent index")
			if !strings.Contains(updated, "# Real Title\n\nParent index: [Docs](./README.md)\n\nBody.") {
				t.Fatalf("parent inserted at wrong location:\n%s", updated)
			}
			if !strings.HasPrefix(updated, source[:4]) {
				t.Fatalf("frontmatter prefix changed:\n%s", updated)
			}
		})
	}
}

func TestGoldmarkIgnoresHeadingsInsideCodeFences(t *testing.T) {
	source := "# Real\n\n```md\n## Related Docs\n```\n\nTail\n"
	got := EnsureManaged(source, config.Default())
	if strings.Index(got, "## Direct Files") < strings.Index(got, "```md") {
		t.Fatalf("managed blocks anchored to fenced heading:\n%s", got)
	}
	if !strings.Contains(got, "```md\n## Related Docs\n```\n\nTail") {
		t.Fatal("unmanaged fenced content changed")
	}
}
func TestManagedReplacementPreservesOutsideBytes(t *testing.T) {
	source := "# Docs\n\n<!-- user comment -->  \n\n## Direct Files\n<!-- doc-ledger:files:start -->\nold\n<!-- doc-ledger:files:end -->\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n<!-- doc-ledger:folders:end -->\n\nTail  \n"
	got, err := ReplaceManaged(source, "files", []string{"- [a.md](a.md) - A."}, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "# Docs\n\n<!-- user comment -->  \n") || !strings.HasSuffix(got, "\n\nTail  \n") {
		t.Fatalf("unmanaged content changed: %q", got)
	}
}
func TestConfiguredIndexAndParentLinks(t *testing.T) {
	c := config.Default()
	c.IndexFile = "!README.md"
	c.Files.IndexFile = "!README.md"
	c.ParentLink.IndexedFiles = true
	root := filepath.Join("tmp", "docs")
	if got := DesiredParent(filepath.Join(root, "guide", "!README.md"), root, func(string) string { return "Docs" }, c); got != "Parent index: [Docs](../!README.md)" {
		t.Fatal(got)
	}
	if got := DesiredParent(filepath.Join(root, "page.md"), root, func(string) string { return "Docs" }, c); got != "Parent index: [Docs](./!README.md)" {
		t.Fatal(got)
	}
}
func TestTemplateAndDescriptions(t *testing.T) {
	c := config.Default()
	got := MakeTemplate(filepath.Join("tmp", "service-runbooks"), filepath.Join("tmp", "docs"), "Docs", "README.md", c)
	if !strings.Contains(got, "# Service Runbooks") || !strings.Contains(got, "Parent index: [Docs](../README.md)") {
		t.Fatal(got)
	}
	if DescriptionFromFile("draft-report.pdf", true, c) != "Stub: Draft Report documentation." {
		t.Fatal("description mismatch")
	}
}

func TestParentInsertionPreservesFinalNewlineState(t *testing.T) {
	for _, test := range []struct {
		name, source, want string
	}{
		{"absent", "# Page", "# Page\n\nParent index: [Docs](./README.md)"},
		{"present", "# Page\n", "# Page\n\nParent index: [Docs](./README.md)\n"},
		{"body_present", "# Page\n\nBody\n", "# Page\n\nParent index: [Docs](./README.md)\n\nBody\n"},
		{"body_absent", "# Page\n\nBody", "# Page\n\nParent index: [Docs](./README.md)\n\nBody"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := UpdateParent(test.source, "Parent index: [Docs](./README.md)", "Parent index")
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestMarkerLikeFenceContentIsNeverManaged(t *testing.T) {
	c := config.Default()
	source := "# Docs\n\n```md\n## Direct Files\n<!-- doc-ledger:files:start -->\nowned by user\n<!-- doc-ledger:files:end -->\n```\n\nTail\n"
	got := EnsureManaged(source, c)
	if !strings.Contains(got, "```md\n## Direct Files\n<!-- doc-ledger:files:start -->\nowned by user\n<!-- doc-ledger:files:end -->\n```") {
		t.Fatalf("fenced marker-like content changed:\n%s", got)
	}
	if strings.Count(got, "<!-- doc-ledger:files:start -->") != 2 {
		t.Fatalf("real managed section was not added separately:\n%s", got)
	}
}
