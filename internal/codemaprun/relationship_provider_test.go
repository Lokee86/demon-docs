package codemaprun

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

type runRelationshipProvider struct{}

func (runRelationshipProvider) CollectRelationships(context.Context, codemapcorpus.RelationshipRequest) ([]evidence.RelationshipEdge, error) {
	return []evidence.RelationshipEdge{{Source: "src/runtime.go", Target: "worker/worker.go", Relation: "calls"}}, nil
}

func TestBuildPassesRelationshipProviderIntoPerDocumentEvidence(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\n## Code Map\n\n- `src/runtime.go`\n")
	writeFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n")
	writeFile(t, filepath.Join(root, "worker", "worker.go"), "package worker\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot:       root,
		DocsRoot:             docs,
		TargetFiles:          []string{filepath.Join(docs, "runtime.md")},
		Headings:             []string{"Code Map"},
		MarkerPrefix:         "ddocs",
		RelationshipProvider: runRelationshipProvider{},
	})
	if err != nil {
		t.Fatal(err)
	}
	document := plan.Documents[0]
	if len(document.Added) != 0 {
		t.Fatalf("semantic-only relationship was auto-written: %#v", document.Added)
	}
	found := false
	for _, recommendation := range document.Recommendations {
		if recommendation.Target != "worker/worker.go" {
			continue
		}
		found = recommendation.Tier == codemaprecommend.SuggestionTierContext
		for _, detail := range recommendation.Evidence {
			if strings.Contains(detail, string(evidence.KindSemanticRelationship)) {
				return
			}
		}
	}
	if !found {
		t.Fatalf("semantic context recommendation missing: %#v", document.Recommendations)
	}
	t.Fatal("semantic relationship evidence detail missing")
}
