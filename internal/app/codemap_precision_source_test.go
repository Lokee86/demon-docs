package app

import (
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func TestFilterCodemapDatasetExcludesDocumentPrefixes(t *testing.T) {
	dataset := codemap.Dataset{
		SchemaVersion: codemap.DatasetSchemaVersion,
		Documents: []codemap.DocumentRecord{
			{Path: "docs/keep.md"},
			{Path: ".worktrees/branch/docs/drop.md"},
		},
		Entries: []codemap.DatasetEntry{
			{Entry: codemap.Entry{DocumentPath: "docs/keep.md"}},
			{Entry: codemap.Entry{DocumentPath: ".worktrees/branch/docs/drop.md"}},
		},
		Diagnostics: []codemap.Diagnostic{
			{DocumentPath: "docs/keep.md"},
			{DocumentPath: ".worktrees/branch/docs/drop.md"},
		},
	}

	filtered := filterCodemapDataset(dataset, []string{`.worktrees\`})
	if len(filtered.Documents) != 1 || filtered.Documents[0].Path != "docs/keep.md" {
		t.Fatalf("documents = %#v", filtered.Documents)
	}
	if len(filtered.Entries) != 1 || filtered.Entries[0].Entry.DocumentPath != "docs/keep.md" {
		t.Fatalf("entries = %#v", filtered.Entries)
	}
	if len(filtered.Diagnostics) != 1 || filtered.Diagnostics[0].DocumentPath != "docs/keep.md" {
		t.Fatalf("diagnostics = %#v", filtered.Diagnostics)
	}
}
