package codemapbench

import (
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestSuggestionsFromEvidenceRejectsUnownedDependencyLockfiles(t *testing.T) {
	candidates := []evidence.Candidate{
		{
			Path: "Cargo.lock",
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindExactPathMention,
				Detail: "Cargo.lock",
				Count:  2,
			}},
		},
		{
			Path: "vendor/checksums.lock",
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindDependencyNeighbor,
				Source: "src/runtime.go",
				Detail: "outbound:generated_dependency",
				Count:  1,
			}},
		},
	}

	suggestions := SuggestionsFromEvidence("docs/runtime.md", candidates)
	if len(suggestions) != 1 || suggestions[0].Target != "vendor/checksums.lock" {
		t.Fatalf("unexpected lockfile suggestions: %#v", suggestions)
	}
}

func TestSuggestionsFromEvidenceRejectsWeakNestedContentMatches(t *testing.T) {
	candidates := []evidence.Candidate{
		{
			Path: "genesis/assets/meshes/camera/",
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindUniqueBasenameMention,
				Detail: "camera",
				Count:  1,
			}},
		},
		{
			Path: "genesis/assets/",
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindExactPathMention,
				Detail: "genesis/assets/",
				Count:  1,
			}},
		},
		{
			Path: "examples/plugins/hello-world/main.go",
			Evidence: []evidence.Evidence{
				{Kind: evidence.KindUniqueBasenameMention, Detail: "main.go", Count: 1},
				{Kind: evidence.KindDeclaredSymbolMention, Detail: "HelloWorldPlugin", Count: 1},
			},
		},
	}

	suggestions := SuggestionsFromEvidence("docs/architecture.md", candidates)
	byTarget := make(map[string]Suggestion, len(suggestions))
	for _, suggestion := range suggestions {
		byTarget[suggestion.Target] = suggestion
	}
	if _, exists := byTarget["genesis/assets/meshes/camera/"]; exists {
		t.Fatalf("weak nested asset match survived: %#v", suggestions)
	}
	for _, target := range []string{"genesis/assets/", "examples/plugins/hello-world/main.go"} {
		if _, exists := byTarget[target]; !exists {
			t.Fatalf("supported content target %q was removed: %#v", target, suggestions)
		}
	}
}

func TestSuggestionsFromEvidenceRejectsWeakWorkflowInfrastructureMatches(t *testing.T) {
	candidates := []evidence.Candidate{
		{
			Path: ".github/workflows/scripts/",
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindUniqueBasenameMention,
				Detail: "scripts",
				Count:  1,
			}},
		},
		{
			Path: ".github/workflows/",
			Evidence: []evidence.Evidence{{
				Kind:   evidence.KindExactPathMention,
				Detail: ".github/workflows/",
				Count:  1,
			}},
		},
	}

	suggestions := SuggestionsFromEvidence("AGENTS.md", candidates)
	if len(suggestions) != 1 || suggestions[0].Target != ".github/workflows/" {
		t.Fatalf("unexpected workflow suggestions: %#v", suggestions)
	}
}
