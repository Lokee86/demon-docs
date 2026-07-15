package parity_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func parityFixtures() []parityFixture {
	return []parityFixture{
		{name: "defaults", setup: setupDefaults},
		{name: "custom_index_headings_markers_drafts_non_markdown_editable", setup: setupCustomConfiguration},
		{name: "direct_to_stub_transition", setup: setupDirectToStub},
		{name: "stub_graduation", setup: setupStubGraduation},
		{name: "unique_file_move", setup: setupUniqueFileMove},
		{name: "ambiguous_file_move", setup: setupAmbiguousFileMove},
		{name: "unique_folder_move", setup: setupUniqueFolderMove},
		{name: "ambiguous_folder_move", setup: setupAmbiguousFolderMove},
		{name: "stale_entry_removal", setup: setupStaleEntryRemoval},
		{name: "malformed_managed_block", setup: setupMalformedManagedBlock},
	}
}

func setupDefaults(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/page.md", "# Page\n\nBody\n")
	writeFixtureText(t, project, "docs/guide/topic.md", "# Topic\n")
	writeFixtureText(t, project, "docs/stubs/idea.md", "# Idea\n")
}

func setupCustomConfiguration(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", `root = "docs"
index_file = "INDEX.md"
[markers]
prefix = "nav"
[parent_link]
label = "Up"
folder_indexes = true
indexed_files = true
[sections.files]
heading = "Pages"
[sections.stubs]
heading = "Ideas"
[sections.folders]
heading = "Areas"
[drafts]
folder = "_drafts"
description_prefix = "Draft: "
[files]
include_patterns = ["**/*.md", "**/*.mdx", "**/*.pdf"]
[editable]
parent_index_extensions = [".md", ".mdx"]
`)
	writeFixtureText(t, project, "docs/page.md", "# Page\n\nUp: [Docs](./INDEX.md)")
	writeFixtureText(t, project, "docs/component.mdx", "# Component\n\nUp: [Docs](./INDEX.md)")
	writeFixtureBytes(t, project, "docs/reference.pdf", []byte("%PDF-fixture\x00bytes"))
	writeFixtureText(t, project, "docs/_drafts/idea.md", "# Idea\n\nUp: [Docs](../INDEX.md)")
	writeFixtureText(t, project, "docs/guide/topic.md", "# Topic\n\nUp: [Guide](./INDEX.md)")
}

func setupDirectToStub(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/README.md", managedIndex("Docs", "- [report.md](report.md) - Carefully written report.", "", ""))
	writeFixtureText(t, project, "docs/stubs/report.md", "# Report\n")
}

func setupStubGraduation(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/README.md", managedIndex("Docs", "", "- [report.md](stubs/report.md) - Stub: Carefully written report.", ""))
	writeFixtureText(t, project, "docs/report.md", "# Report\n")
}

func setupUniqueFileMove(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/old/README.md", managedIndex("Old", "- [report.md](report.md) - Unique moved report.", "", ""))
	writeFixtureText(t, project, "docs/new/report.md", "# Report\n")
}

func setupAmbiguousFileMove(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/a/README.md", managedIndex("A", "- [report.md](report.md) - First stale report.", "", ""))
	writeFixtureText(t, project, "docs/b/README.md", managedIndex("B", "- [report.md](report.md) - Second stale report.", "", ""))
	writeFixtureText(t, project, "docs/c/report.md", "# Report\n")
}

func setupUniqueFolderMove(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/old/README.md", managedIndex("Old", "", "", "- [team](team/README.md) - Unique team description."))
	writeFixtureText(t, project, "docs/new/team/topic.md", "# Topic\n")
}

func setupAmbiguousFolderMove(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/a/README.md", managedIndex("A", "", "", "- [team](team/README.md) - First stale team."))
	writeFixtureText(t, project, "docs/b/README.md", managedIndex("B", "", "", "- [team](team/README.md) - Second stale team."))
	writeFixtureText(t, project, "docs/c/team/topic.md", "# Topic\n")
}

func setupStaleEntryRemoval(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/README.md", managedIndex("Docs", "- [gone.md](gone.md) - Removed page.\n- [keep.md](keep.md) - Kept page.", "", "- [gone](gone/README.md) - Removed folder."))
	writeFixtureText(t, project, "docs/keep.md", "# Keep\n")
}

func setupMalformedManagedBlock(t *testing.T, project string) {
	writeFixtureText(t, project, ".doc-ledger.toml", "root = \"docs\"\n")
	writeFixtureText(t, project, "docs/README.md", "# Docs\n\nIntro stays.\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n- [old.md](old.md) - Old.\n\n## Notes\n\nFollowing content must survive.")
	writeFixtureText(t, project, "docs/page.md", "# Page\n")
}

func managedIndex(title, files, stubs, folders string) string {
	return "# " + title + "\n\n" +
		"## Direct Files\n<!-- doc-ledger:files:start -->\n" + optionalManagedBody(files) + "<!-- doc-ledger:files:end -->\n\n" +
		"## Stub Files\n<!-- doc-ledger:stubs:start -->\n" + optionalManagedBody(stubs) + "<!-- doc-ledger:stubs:end -->\n\n" +
		"## Direct Folders\n<!-- doc-ledger:folders:start -->\n" + optionalManagedBody(folders) + "<!-- doc-ledger:folders:end -->"
}

func optionalManagedBody(body string) string {
	if body == "" {
		return ""
	}
	return "\n" + body + "\n"
}

func writeFixtureText(t *testing.T, project, relative, text string) {
	t.Helper()
	writeFixtureBytes(t, project, relative, []byte(nativeNewlines(text)))
}

func writeFixtureBytes(t *testing.T, project, relative string, data []byte) {
	t.Helper()
	path := filepath.Join(project, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func nativeNewlines(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(text, "\n", "\r\n")
	}
	return text
}
