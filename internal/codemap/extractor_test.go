package codemap

import "testing"

func TestExtractDefaultParsesSpaceRocksStyleCodeMap(t *testing.T) {
	source := "# Runtime\n\n## Code map\n\n### Runtime owners\n\n* `client/scripts/runtime/gameplay.gd`\n* `client/scripts/runtime/` - supporting runtime files\n\n### Boundaries\n\n* `services/game-server/internal/game/` owns authority\n\n## Tests\n\n* `ignored/test.gd`\n"

	result := ExtractDefault(`docs\\runtime.md`, source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", result.Diagnostics)
	}
	if len(result.Entries) != 3 {
		t.Fatalf("got %d entries, want 3: %#v", len(result.Entries), result.Entries)
	}

	first := result.Entries[0]
	if first.DocumentPath != "docs/runtime.md" || first.Target != "client/scripts/runtime/gameplay.gd" {
		t.Fatalf("unexpected first entry: %#v", first)
	}
	if first.Kind != TargetFile || first.Context != "Runtime owners" || first.Description != "" {
		t.Fatalf("unexpected first metadata: %#v", first)
	}
	if first.Source.Line != 7 || first.Source.Column != 4 || first.Source.EndColumn != 37 {
		t.Fatalf("unexpected first source span: %#v", first.Source)
	}

	second := result.Entries[1]
	if second.Kind != TargetDirectory || second.Description != "supporting runtime files" {
		t.Fatalf("unexpected second entry: %#v", second)
	}
	if result.Entries[2].Context != "Boundaries" {
		t.Fatalf("unexpected subsection context: %#v", result.Entries[2])
	}
}

func TestExtractSupportsRepositorySpecificHeading(t *testing.T) {
	source := "## Implementation paths\n\n- `src/feature.go`\n"
	result := Extract("docs/feature.md", source, Format{SectionHeadings: []string{"Implementation paths"}})
	if len(result.Entries) != 1 || result.Entries[0].Target != "src/feature.go" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestExtractReportsUnsupportedAuthoredListEntry(t *testing.T) {
	source := "## Code map\r\n\r\n- TODO: add paths later.\r\n"
	result := ExtractDefault("docs/future.md", source)
	if len(result.Entries) != 0 || len(result.Diagnostics) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	got := result.Diagnostics[0]
	if got.Code != "unparsed_entry" || got.Source.Line != 3 || got.RawLine != "- TODO: add paths later." {
		t.Fatalf("unexpected diagnostic: %#v", got)
	}
}

func TestExtractIgnoresFencesAndStopsAtPeerHeading(t *testing.T) {
	source := "## Code map\n\n```md\n- `ignored/in/fence.go`\n```\n\n- `src/kept.go`\n\n## Related docs\n\n- `ignored/after.go`\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 1 || result.Entries[0].Target != "src/kept.go" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestExtractCreatesOneEntryPerTargetOnAListLine(t *testing.T) {
	source := "## Code map\n\n- `src/a.go` and `src/b.go` implement the boundary.\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 2 {
		t.Fatalf("got %d entries, want 2: %#v", len(result.Entries), result.Entries)
	}
	for _, entry := range result.Entries {
		if entry.Description != "implement the boundary." {
			t.Fatalf("unexpected shared description: %#v", entry)
		}
	}
}

func TestExtractDoesNotTreatInlineSymbolsInDescriptionsAsPathTargets(t *testing.T) {
	source := "## Code map\n\n- `src/a.go` owns `Handler`.\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 1 || result.Entries[0].Target != "src/a.go" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestExtractAcceptsBarePrimarySymbolTargets(t *testing.T) {
	source := "## Code map\n\n- `DevtoolsWindowController` owns window lifecycle.\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Diagnostics) != 0 || len(result.Entries) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Entries[0].Target != "DevtoolsWindowController" || result.Entries[0].Kind != TargetSymbol {
		t.Fatalf("unexpected symbol entry: %#v", result.Entries[0])
	}
}
