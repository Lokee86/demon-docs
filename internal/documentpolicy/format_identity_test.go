package documentpolicy

import "testing"

func TestFormatIdentityIgnoresUnvalidatedContent(t *testing.T) {
	left := parseMarkdown("# Original title\n\n## Purpose\n\nOriginal prose with [a link](old.md).\n\n```md\n## Protected heading\n```\n\n### Child\n\nBody.\n")
	right := parseMarkdown("# Changed title\n\n## Purpose\n\nDifferent prose with [another link](new.md).\n\n```md\n## Different protected heading\n```\n\n### Child\n\nEntirely different body.\n")
	leftHash := formatIdentity("general", "doc-1", "general", left)
	rightHash := formatIdentity("general", "doc-1", "general", right)
	if leftHash != rightHash {
		t.Fatal("ordinary body content or protected headings changed format identity")
	}
}

func TestFormatIdentityTracksSelectionMetadataAndHeadingStructure(t *testing.T) {
	baseDocument := parseMarkdown("## Alpha\n\n### Child\n\n## Alpha\n")
	base := formatIdentity("general", "doc-1", "general", baseDocument)
	tests := map[string]string{
		"selected schema": formatIdentity("planning", "doc-1", "general", baseDocument),
		"document ID":     formatIdentity("general", "doc-2", "general", baseDocument),
		"document type":   formatIdentity("general", "doc-1", "planning", baseDocument),
		"heading text":    formatIdentity("general", "doc-1", "general", parseMarkdown("## Changed\n\n### Child\n\n## Alpha\n")),
		"heading level":   formatIdentity("general", "doc-1", "general", parseMarkdown("## Alpha\n\n## Child\n\n## Alpha\n")),
		"heading order":   formatIdentity("general", "doc-1", "general", parseMarkdown("## Alpha\n\n## Alpha\n\n### Child\n")),
		"duplicate count": formatIdentity("general", "doc-1", "general", parseMarkdown("## Alpha\n\n### Child\n")),
	}
	for name, candidate := range tests {
		if candidate == base {
			t.Fatalf("%s did not change format identity", name)
		}
	}
}
