package documentpolicy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
)

func TestMarkdownIgnoresProtectedHeadings(t *testing.T) {
	source := "# Title\n\n```md\n## Fenced\n```\n\n> ## Quoted\n\n<div>\n## HTML\n</div>\n\n## Real\nBody\n"
	document := parseMarkdown(source)
	if len(document.Roots) != 1 || document.Roots[0].Heading != "Real" {
		t.Fatalf("roots = %#v", headings(document.Roots))
	}
	if document.render() != source {
		t.Fatalf("round trip changed source:\n%s", document.render())
	}
}

func TestMarkdownPreservesHashInsideHeadingText(t *testing.T) {
	document := parseMarkdown("# Title\n\n## C#\nLanguage notes.\n\n## Trimmed ##\nBody.\n")
	if len(document.Roots) != 2 || document.Roots[0].Heading != "C#" || document.Roots[1].Heading != "Trimmed" {
		t.Fatalf("headings = %#v", headings(document.Roots))
	}
}

func TestEnforcementReordersAndCreatesMissingSections(t *testing.T) {
	schema := Schema{Name: "test", Placeholder: "TODO", Sections: []Section{
		{ID: "purpose", Heading: "Purpose"},
		{ID: "overview", Heading: "Overview"},
	}}
	document := parseMarkdown("# Title\n\n## Overview\nText\n\n## Purpose\nWhy\n")
	result := enforceDocument(document, schema, Schema{}, false, true)
	if result.Blocked || !result.Changed {
		t.Fatalf("result = %#v", result)
	}
	rendered := result.Document.render()
	if strings.Index(rendered, "## Purpose") > strings.Index(rendered, "## Overview") {
		t.Fatalf("sections not reordered:\n%s", rendered)
	}
}

func TestEnforcementMovesSectionToConfiguredParent(t *testing.T) {
	schema := Schema{Name: "test", Placeholder: "TODO", Sections: []Section{
		{ID: "parent", Heading: "Parent"},
		{ID: "child", Heading: "Child", Parent: "parent"},
	}}
	document := parseMarkdown("# Title\n\n## Child\nAuthored child prose.\n\n## Parent\nParent prose.\n")
	result := enforceDocument(document, schema, Schema{}, false, true)
	if result.Blocked || !result.Changed {
		t.Fatalf("result = %#v", result)
	}
	rendered := result.Document.render()
	if !strings.Contains(rendered, "## Parent\nParent prose.\n### Child\nAuthored child prose.") {
		t.Fatalf("section was not moved beneath its configured parent:\n%s", rendered)
	}
}

func TestUnknownSectionBlocksFix(t *testing.T) {
	schema := Schema{Name: "test", Placeholder: "TODO", UnknownSections: "manual", Sections: []Section{{ID: "purpose", Heading: "Purpose"}}}
	document := parseMarkdown("# Title\n\n## Extra\nKeep me\n")
	result := enforceDocument(document, schema, Schema{}, false, true)
	if !result.Blocked || result.Changed {
		t.Fatalf("result = %#v", result)
	}
	if result.Document.render() != document.render() {
		t.Fatal("blocked fix changed document")
	}
}

func TestSchemaRenameUsesStableSectionID(t *testing.T) {
	previous := Schema{Name: "test", Sections: []Section{{ID: "decision", Heading: "Decision"}}}
	current := Schema{Name: "test", Sections: []Section{{ID: "decision", Heading: "Resolution"}}}
	document := parseMarkdown("# Title\n\n## Decision\nKeep prose.\n")
	result := enforceDocument(document, current, previous, true, true)
	if !strings.Contains(result.Document.render(), "## Resolution\nKeep prose.") {
		t.Fatalf("rename not propagated:\n%s", result.Document.render())
	}
}

func TestMergeSameListTypeDeduplicatesExactItems(t *testing.T) {
	first := &markdownSection{Heading: "Items", Lead: "\n- Alpha\n- Beta\n\n"}
	second := &markdownSection{Heading: "Items", Lead: "\n* Beta\n* Gamma\n\n"}
	mergeNodes(first, second, "\n")
	if got := strings.Count(first.Lead, "Beta"); got != 1 {
		t.Fatalf("Beta count = %d:\n%s", got, first.Lead)
	}
	if !strings.Contains(first.Lead, "Gamma") {
		t.Fatalf("missing appended item:\n%s", first.Lead)
	}
}

func TestCreatePreservesRequiredIdentityWhenFrontmatterEnforcementIsDisabled(t *testing.T) {
	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "docs")
	if err := os.Mkdir(docsRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	target := filepath.Join(docsRoot, "guide.md")
	if _, err := Create(repoRoot, docsRoot, cfg, "general", target, false, time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"document_id:", "document_type: general", `created: "2026-07-19"`, "# Guide"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("created document missing %q:\n%s", expected, text)
		}
	}
}

func TestSchemaValidationRejectsCyclesAndAmbiguousSiblingHeadings(t *testing.T) {
	base := Schema{Name: "test", Version: 1, UnknownSections: "manual", DuplicateSections: "manual"}

	cycle := base
	cycle.Sections = []Section{{ID: "a", Heading: "A", Parent: "b"}, {ID: "b", Heading: "B", Parent: "a"}}
	if err := ValidateSchema(cycle); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle validation error = %v", err)
	}

	ambiguous := base
	ambiguous.Sections = []Section{{ID: "a", Heading: "Alpha", Aliases: []string{"Shared"}}, {ID: "b", Heading: "Shared"}}
	if err := ValidateSchema(ambiguous); err == nil || !strings.Contains(err.Error(), "share heading") {
		t.Fatalf("ambiguous heading validation error = %v", err)
	}

	invalidAfter := base
	invalidAfter.Sections = []Section{{ID: "parent", Heading: "Parent"}, {ID: "child", Heading: "Child", Parent: "parent", After: "parent"}}
	if err := ValidateSchema(invalidAfter); err == nil || !strings.Contains(err.Error(), "not a sibling") {
		t.Fatalf("after validation error = %v", err)
	}
}

func TestSchemaSimilarityCountsStableSectionDefinitions(t *testing.T) {
	previous := Schema{Sections: []Section{{ID: "a", Heading: "A"}, {ID: "b", Heading: "B"}}}
	oneOfTwoChanged := Schema{Sections: []Section{{ID: "a", Heading: "Renamed"}, {ID: "b", Heading: "B"}}}
	if similarity := Similarity(previous, oneOfTwoChanged); similarity != 0.5 {
		t.Fatalf("one-of-two similarity = %v, want 0.5", similarity)
	}
	allChanged := Schema{Sections: []Section{{ID: "a", Heading: "One"}, {ID: "b", Heading: "Two"}}}
	if similarity := Similarity(previous, allChanged); similarity != 0 {
		t.Fatalf("all-changed similarity = %v, want 0", similarity)
	}
}

func TestDocumentSpecificSchemaRejectsUnsafeDocumentID(t *testing.T) {
	_, _, _, err := LoadDocumentSchema(t.TempDir(), config.Format{DocumentSchemaDir: ".ddocs/document-schemas"}, "../escape")
	if err == nil || !strings.Contains(err.Error(), "unsafe document_id") {
		t.Fatalf("unsafe document ID error = %v", err)
	}
}

func TestSchemaSelectionUsesMetadataBeforePathFallback(t *testing.T) {
	cfg := config.Format{
		DefaultSchema: "general",
		PathRules: []config.FormatPathRule{
			{Pattern: "**/README.md", Schema: "index"},
			{Pattern: "docs/planning/**", Schema: "planning"},
		},
	}
	selected, err := selectSchema("docs/planning/topic.md", map[string]any{"document_type": "service"}, cfg)
	if err != nil || selected != "service" {
		t.Fatalf("metadata selection = %q, %v", selected, err)
	}
	selected, err = selectSchema("docs/planning/topic.md", map[string]any{}, cfg)
	if err != nil || selected != "planning" {
		t.Fatalf("path fallback = %q, %v", selected, err)
	}
}

func TestMergeDifferentContentConcatenates(t *testing.T) {
	first := &markdownSection{Heading: "Notes", Lead: "\nFirst prose.\n\n"}
	second := &markdownSection{Heading: "Notes", Lead: "\n- Item\n\n"}
	mergeNodes(first, second, "\n")
	combined := first.Lead + first.Tail
	if !strings.Contains(combined, "First prose.") || !strings.Contains(combined, "- Item") {
		t.Fatalf("content not concatenated:\n%s", combined)
	}
}

func TestCodemapSchemaProviderPlacesRequiredServiceSection(t *testing.T) {
	repoRoot := t.TempDir()
	cfg := config.Default()
	cfg.Format.Enabled = true
	provider := CodemapSchemaProvider{
		RepositoryRoot: repoRoot,
		Config:         cfg,
		Headings:       cfg.Codemap.Headings,
	}
	source := "---\ndocument_type: service\n---\n# Service\n\n## Purpose\nWhy.\n\n## Protocol and API surfaces\nAPI.\n\n## Tests and verification\nTests.\n"
	placement, required, err := provider.CodemapSection("docs/services/example.md", source)
	if err != nil {
		t.Fatal(err)
	}
	if !required || placement.Heading != "Code map" || placement.Level != 2 {
		t.Fatalf("placement = %#v, required = %t", placement, required)
	}
	testsOffset := strings.Index(source, "## Tests and verification")
	if placement.Offset != testsOffset {
		t.Fatalf("offset = %d, want %d", placement.Offset, testsOffset)
	}
}

func TestCodemapSchemaProviderLeavesSchemaWithoutCodemapUnchanged(t *testing.T) {
	cfg := config.Default()
	cfg.Format.Enabled = true
	provider := CodemapSchemaProvider{
		RepositoryRoot: t.TempDir(),
		Config:         cfg,
		Headings:       cfg.Codemap.Headings,
	}
	placement, required, err := provider.CodemapSection("docs/guide.md", "---\ndocument_type: general\n---\n# Guide\n")
	if err != nil {
		t.Fatal(err)
	}
	if required || placement != (codemap.SectionPlacement{}) {
		t.Fatalf("placement = %#v, required = %t", placement, required)
	}
}

func headings(sections []*markdownSection) []string {
	result := make([]string, len(sections))
	for i, section := range sections {
		result[i] = section.Heading
	}
	return result
}
