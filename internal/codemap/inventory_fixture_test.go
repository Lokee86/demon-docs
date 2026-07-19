package codemap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fixtureManifest struct {
	Fixtures []struct {
		File        string `json:"file"`
		Placeholder bool   `json:"placeholder"`
		Entries     []struct {
			Group       string     `json:"group"`
			Target      string     `json:"target"`
			Kind        TargetKind `json:"kind"`
			Syntax      SyntaxKind `json:"syntax"`
			Description string     `json:"description"`
		} `json:"entries"`
	} `json:"fixtures"`
}

func TestExtractorMatchesMergedSpaceRocksInventoryFixtures(t *testing.T) {
	fixtureRoot := filepath.Join("..", "..", "research", "codemap-inventory", "fixtures")
	encoded, err := os.ReadFile(filepath.Join(fixtureRoot, "expected.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest fixtureManifest
	if err := json.Unmarshal(encoded, &manifest); err != nil {
		t.Fatal(err)
	}

	for _, fixture := range manifest.Fixtures {
		fixture := fixture
		t.Run(fixture.File, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join(fixtureRoot, fixture.File))
			if err != nil {
				t.Fatal(err)
			}
			result := ExtractDefault("docs/"+fixture.File, string(source))
			if len(result.Diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", result.Diagnostics)
			}
			if len(result.Entries) != len(fixture.Entries) {
				t.Fatalf("got %d entries, want %d: %#v", len(result.Entries), len(fixture.Entries), result.Entries)
			}
			for index, expected := range fixture.Entries {
				got := result.Entries[index]
				if got.Context != expected.Group || got.Target != expected.Target || got.Kind != expected.Kind || got.Syntax != expected.Syntax || got.Description != expected.Description {
					t.Errorf("entry %d = %#v, want group=%q target=%q kind=%q syntax=%q description=%q", index, got, expected.Group, expected.Target, expected.Kind, expected.Syntax, expected.Description)
				}
			}
		})
	}
}
