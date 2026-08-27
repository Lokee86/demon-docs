package codemaprecommend

import (
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestClassifySuggestionRole(t *testing.T) {
	tests := []struct {
		name   string
		target string
		items  []evidence.Evidence
		want   SuggestionRole
	}{
		{
			name:   "verification by path",
			target: "internal/runtime_test.go",
			items:  []evidence.Evidence{{Kind: evidence.KindDeclaredSymbolMention, Detail: "Runtime"}},
			want:   SuggestionRoleVerification,
		},
		{
			name:   "verification by incoming tests edge",
			target: "internal/runtime_verifier.go",
			items:  []evidence.Evidence{{Kind: evidence.KindSemanticRelationship, Detail: "inbound:tests"}},
			want:   SuggestionRoleVerification,
		},
		{
			name:   "interface boundary by implemented target",
			target: "internal/runtime_contract.go",
			items: []evidence.Evidence{
				{Kind: evidence.KindExactPathMention, Detail: "internal/runtime_contract.go"},
				{Kind: evidence.KindSemanticRelationship, Detail: "outbound:implements"},
			},
			want: SuggestionRoleInterfaceBoundary,
		},
		{
			name:   "implementation side of implements edge",
			target: "internal/runtime.go",
			items:  []evidence.Evidence{{Kind: evidence.KindSemanticRelationship, Detail: "inbound:implements"}},
			want:   SuggestionRoleSupportingImplementation,
		},
		{
			name:   "direct symbol is primary",
			target: "internal/runtime.go",
			items:  []evidence.Evidence{{Kind: evidence.KindDeclaredSymbolMention, Detail: "Runtime"}},
			want:   SuggestionRolePrimaryImplementation,
		},
		{
			name:   "dependency is supporting",
			target: "internal/storage.go",
			items:  []evidence.Evidence{{Kind: evidence.KindDependencyNeighbor, Detail: "outbound:go_import"}},
			want:   SuggestionRoleSupportingImplementation,
		},
		{
			name:   "history only is context",
			target: "internal/old.go",
			items:  []evidence.Evidence{{Kind: evidence.KindGitDocumentCoChange, Count: 2}},
			want:   SuggestionRoleContext,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := classifySuggestionRole(test.target, test.items); got != test.want {
				t.Fatalf("role = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRoleClassificationDoesNotChangeTierOrScore(t *testing.T) {
	candidate := evidence.Candidate{
		Path: "internal/runtime.go",
		Evidence: []evidence.Evidence{
			{Kind: evidence.KindDependencyNeighbor, Source: "internal/app.go", Detail: "outbound:go_import", Count: 7},
			{Kind: evidence.KindSemanticRelationship, Source: "internal/app.go", Detail: "inbound:implements", Count: 1},
		},
	}
	items := SuggestionsFromEvidence("docs/runtime.md", []evidence.Candidate{candidate})
	if len(items) != 1 {
		t.Fatalf("suggestions = %#v", items)
	}
	if items[0].Role != SuggestionRoleSupportingImplementation {
		t.Fatalf("role = %q", items[0].Role)
	}
	if items[0].Tier != SuggestionTierContext {
		t.Fatalf("classification changed tier: %#v", items[0])
	}
	if items[0].Score <= HardLinkDependencyMinimumScore {
		t.Fatalf("test no longer exercises semantic-score threshold guard: %#v", items[0])
	}
}
