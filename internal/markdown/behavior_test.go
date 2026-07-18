package markdown

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestManagedSectionMigrationAndPlacement(t *testing.T) {
	c := config.Default()
	legacy := "# Docs\n\n## Top-Level Files\n\n- [a.md](a.md) - Keep me.\n\n## Top-Level Folders\n\n- [guide](guide/README.md) - Keep folder.\n\n## Notes\n\nUser notes."
	got := EnsureManaged(legacy, c)
	for _, want := range []string{"## Direct Files", MarkerStart("doc-ledger", "files"), "- [a.md](a.md) - Keep me.", "## Stub Files", "## Direct Folders", "- [guide](guide/README.md) - Keep folder."} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
	if strings.Index(got, "## Stub Files") > strings.Index(got, "## Notes") {
		t.Fatal("managed section inserted after Notes")
	}
	if strings.Count(EnsureManaged(got, c), MarkerStart("doc-ledger", "files")) != 1 {
		t.Fatal("managed migration is not idempotent")
	}
}

func TestParseEntriesStopsAtUnmanagedHeading(t *testing.T) {
	c := config.Default()
	source := "## Direct Files\n<!-- doc-ledger:files:start -->\n- [a.md](a.md) - A.\n<!-- doc-ledger:files:end -->\n## Other\n- [outside.md](outside.md) - Outside.\n"
	entries := ParseEntries("README.md", source, c)
	if len(entries) != 1 || entries[0].LinkTarget != "a.md" || entries[0].Description != "A." {
		t.Fatalf("entries=%+v", entries)
	}
}

func TestCustomMarkersAndHeadings(t *testing.T) {
	c := config.Default()
	c.Markers.Prefix = "nav"
	c.Sections.FilesHeading = "Pages"
	c.Sections.StubsHeading = "Ideas"
	c.Sections.FoldersHeading = "Areas"
	got := EnsureManaged("# Docs", c)
	for _, want := range []string{"## Pages", "<!-- nav:files:start -->", "## Ideas", "## Areas"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q", want)
		}
	}
	got, err := ReplaceManaged(got, "files", []string{"- [x.md](x.md) - X."}, c)
	if err != nil {
		t.Fatal(err)
	}
	entries := ParseEntries("INDEX.md", got, c)
	if len(entries) != 1 || entries[0].Section != "files" {
		t.Fatalf("entries=%+v", entries)
	}
}

func TestUpdateParentInsertReplaceAndRemove(t *testing.T) {
	inserted := UpdateParent("# Title\n\nIntro\n", "Parent index: [Docs](./README.md)", "Parent index")
	if inserted != "# Title\n\nParent index: [Docs](./README.md)\n\nIntro\n" {
		t.Fatalf("insert=%q", inserted)
	}
	replaced := UpdateParent(inserted, "Parent index: [Guide](../README.md)", "Parent index")
	if !strings.Contains(replaced, "[Guide](../README.md)") || strings.Contains(replaced, "[Docs]") {
		t.Fatal(replaced)
	}
	removed := UpdateParent(replaced, "", "Parent index")
	if strings.Contains(removed, "Parent index:") || removed != "# Title\n\nIntro\n" {
		t.Fatalf("remove=%q", removed)
	}
	if got := UpdateParent("Intro only", "Parent index: [Docs](./README.md)", "Parent index"); got != "Parent index: [Docs](./README.md)\n\nIntro only" {
		t.Fatalf("no heading=%q", got)
	}
}

func TestUpdateParentIgnoresFencedParentLinkCandidates(t *testing.T) {
	fenced := "```markdown\nParent index: [Example](./README.md)\n```"
	source := "# Title\n\n" + fenced + "\n\nIntro\n"

	inserted := UpdateParent(source, "Parent index: [Docs](./README.md)", "Parent index")
	if !strings.Contains(inserted, fenced) || strings.Count(inserted, "Parent index:") != 2 {
		t.Fatalf("fenced candidate was replaced during insertion:\n%s", inserted)
	}
	if !strings.Contains(inserted, "# Title\n\nParent index: [Docs](./README.md)\n\n```markdown") {
		t.Fatalf("structural parent link was not inserted after heading:\n%s", inserted)
	}

	withStructural := "# Title\n\nParent index: [Old](../README.md)\n\n" + fenced + "\n"
	replaced := UpdateParent(withStructural, "Parent index: [New](../INDEX.md)", "Parent index")
	if strings.Contains(replaced, "[Old]") || !strings.Contains(replaced, "Parent index: [New](../INDEX.md)") || !strings.Contains(replaced, fenced) {
		t.Fatalf("replacement crossed the fenced-code boundary:\n%s", replaced)
	}

	removed := UpdateParent(source, "", "Parent index")
	if removed != source {
		t.Fatalf("fenced candidate was removed:\nwant %q\n got %q", source, removed)
	}
}

func TestTemplateFeatureTogglesAndConfiguredParent(t *testing.T) {
	c := config.Default()
	c.IndexFile = "INDEX.md"
	c.Files.IndexFile = "INDEX.md"
	c.Template.IncludeOwnership = false
	c.Template.IncludeDoesNotBelong = false
	c.Template.IncludeRelatedDocs = false
	c.Template.IncludeNotes = false
	got := MakeTemplate(filepath.Join("docs", "guide"), "docs", "Docs", "INDEX.md", c)
	for _, absent := range []string{"## Ownership", "## Does Not Belong", "## Related Docs", "## Notes"} {
		if strings.Contains(got, absent) {
			t.Fatalf("unexpected %s", absent)
		}
	}
	if !strings.Contains(got, "Parent index: [Docs](../INDEX.md)") {
		t.Fatal(got)
	}
	for _, managed := range []string{"Direct Files", "Stub Files", "Direct Folders"} {
		if !strings.Contains(got, managed) {
			t.Fatalf("missing %s", managed)
		}
	}
}

func TestTitlesAndRootTitleFallbacks(t *testing.T) {
	if got := TitleFromName("service_RUNBOOKS"); got != "Service Runbooks" {
		t.Fatal(got)
	}
	if got := FolderTitle("fallback-name", "# Explicit Title\n"); got != "Explicit Title" {
		t.Fatal(got)
	}
	if got := ManagedRootTitle("docs", "# Documentation\n", []string{"Docs", "Docs"}); got != "Docs" {
		t.Fatal(got)
	}
	if got := ManagedRootTitle("docs", "# Documentation\n", nil); got != "Documentation" {
		t.Fatal(got)
	}
}
