package codemap

import "testing"

func TestExtractDefaultParsesSpaceRocksStyleCodeMap(t *testing.T) {
	source := "# Runtime\n\n## Code map\n\n### Runtime owners\n\n* `client/scripts/runtime/gameplay.gd`\n* `client/scripts/runtime/` - supporting runtime files\n\n### Boundaries\n\n* `services/game-server/internal/game/` owns authority\n\n## Tests\n\n* `ignored/test.gd`\n"

	result := ExtractDefault(`docs\\runtime.md`, source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", result.Diagnostics)
	}
	if result.SectionCount != 1 {
		t.Fatalf("got %d codemap sections, want 1", result.SectionCount)
	}
	if len(result.Entries) != 3 {
		t.Fatalf("got %d entries, want 3: %#v", len(result.Entries), result.Entries)
	}

	first := result.Entries[0]
	if first.DocumentPath != "docs/runtime.md" || first.Target != "client/scripts/runtime/gameplay.gd" {
		t.Fatalf("unexpected first entry: %#v", first)
	}
	if first.Kind != TargetFile || first.Syntax != SyntaxBullet || first.Heading != "Code map" || first.Context != "Runtime owners" || first.Description != "" {
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

func TestExtractIgnoresProseOnlyBoundaryBullets(t *testing.T) {
	source := "## Code map\r\n\r\n- explain the boundary later.\r\n"
	result := ExtractDefault("docs/future.md", source)
	if len(result.Entries) != 0 || len(result.Diagnostics) != 0 || result.SectionCount != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestExtractTreatsTODOOnlyMapAsEmpty(t *testing.T) {
	result := ExtractDefault("docs/future.md", "## Code map\n\n- TODO: add paths later.\n")
	if len(result.Entries) != 0 || len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected placeholder result: %#v", result)
	}
}

func TestExtractSkipsFencesOutsideCodeMapAndStopsAtPeerHeading(t *testing.T) {
	source := "```md\n## Code map\n- `ignored/in/fence.go`\n```\n\n## Code map\n\n- `src/kept.go`\n\n## Related docs\n\n- `ignored/after.go`\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 1 || result.Entries[0].Target != "src/kept.go" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestExtractTreatsOnlyFirstBulletCodeSpanAsTarget(t *testing.T) {
	source := "## Code map\n\n- `src/a.go` creates `Handler` and calls `src/b.go`.\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 1 || result.Entries[0].Target != "src/a.go" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Entries[0].Description != "creates `Handler` and calls `src/b.go`." {
		t.Fatalf("unexpected description: %#v", result.Entries[0])
	}
}

func TestExtractPreservesBareSymbolBullet(t *testing.T) {
	source := "## Code map\n\n- `Handler` owns orchestration.\n- `math/rand.Rand` provides the random source.\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 2 || result.Entries[0].Target != "Handler" || result.Entries[0].Kind != TargetSymbol || result.Entries[1].Kind != TargetSymbol {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestExtractSupportsEqualsDescriptionPairs(t *testing.T) {
	source := "## Code map\n\n```text\nsrc/runtime.go\n= owns runtime behavior\n```\n"
	result := ExtractDefault("docs/a.md", source)
	if len(result.Entries) != 1 || result.Entries[0].Syntax != SyntaxFencedEquals || result.Entries[0].Description != "owns runtime behavior" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
