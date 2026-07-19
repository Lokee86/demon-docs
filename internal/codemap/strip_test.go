package codemap

import (
	"strings"
	"testing"
)

func TestStripAuthoredSectionsRemovesMapThroughNextPeerHeading(t *testing.T) {
	source := "# Title\n\nIntro.\n\n## Code map\n\n```text\nsrc/hidden.go\n```\n\n### Notes\nStill map prose.\n\n## Behavior\nKeep src/visible.go here.\n"
	stripped := StripAuthoredSections(source, DefaultFormat())
	if strings.Contains(stripped, "src/hidden.go") || strings.Contains(stripped, "Still map prose") {
		t.Fatalf("map content leaked:\n%s", stripped)
	}
	if !strings.Contains(stripped, "Intro.") || !strings.Contains(stripped, "## Behavior") || !strings.Contains(stripped, "src/visible.go") {
		t.Fatalf("non-map content was removed:\n%s", stripped)
	}
	if strings.Count(stripped, "\n") != strings.Count(source, "\n") {
		t.Fatalf("line positions changed")
	}
}

func TestStripAuthoredSectionsIgnoresHeadingInsideFence(t *testing.T) {
	source := "# Title\n\n```markdown\n## Code map\nsrc/example.go\n```\n\nBody.\n"
	stripped := StripAuthoredSections(source, DefaultFormat())
	if stripped != source {
		t.Fatalf("fenced example changed:\n%s", stripped)
	}
}

func TestStripAuthoredSectionsUsesConfiguredAliases(t *testing.T) {
	source := "## Implementation files\n\nsrc/hidden.go\n\n## Next\nKeep.\n"
	format := DefaultFormat()
	format.SectionHeadings = []string{"Implementation files"}
	stripped := StripAuthoredSections(source, format)
	if strings.Contains(stripped, "src/hidden.go") || !strings.Contains(stripped, "Keep.") {
		t.Fatalf("unexpected result:\n%s", stripped)
	}
}
