package codemapcorpus

import (
	"context"
	"reflect"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

type stubRelationshipProvider struct {
	edges   []evidence.RelationshipEdge
	request RelationshipRequest
	calls   int
}

func (provider *stubRelationshipProvider) CollectRelationships(_ context.Context, request RelationshipRequest) ([]evidence.RelationshipEdge, error) {
	provider.calls++
	provider.request = request
	return append([]evidence.RelationshipEdge(nil), provider.edges...), nil
}

func TestInputContextCollectsRelationshipsFromVisibleExactSeed(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/map.md": "# Map\n", "src/a.go": "package src\n", "src/b.go": "package src\n",
	})
	provider := &stubRelationshipProvider{edges: []evidence.RelationshipEdge{
		{Source: "./src/a.go", Target: "src/b.go", Relation: " calls "},
	}}
	dataset := codemap.Dataset{
		Documents: []codemap.DocumentRecord{{Path: "docs/map.md"}},
		Entries: []codemap.DatasetEntry{{
			Entry:      codemap.Entry{DocumentPath: "docs/map.md", Target: "src/a.go", Kind: codemap.TargetFile},
			Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved, ResolvedPath: "src/a.go", Exists: true},
		}},
	}
	corpus, err := BuildContext(context.Background(), root, dataset, Options{RelationshipProvider: provider})
	if err != nil {
		t.Fatal(err)
	}
	input, err := corpus.InputContext(context.Background(), "docs/map.md", []string{"src/a.go"})
	if err != nil {
		t.Fatal(err)
	}
	if provider.calls != 1 || len(provider.request.Seeds) != 1 || provider.request.Seeds[0].Path != "src/a.go" {
		t.Fatalf("relationship request = %#v calls=%d", provider.request, provider.calls)
	}
	want := []evidence.RelationshipEdge{{Source: "src/a.go", Target: "src/b.go", Relation: "calls"}}
	if !reflect.DeepEqual(input.SemanticRelationships, want) {
		t.Fatalf("semantic relationships = %#v", input.SemanticRelationships)
	}
}

func TestInputContextDoesNotFetchRelationshipsFromHiddenHoldoutSeed(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"docs/map.md": "# Map\n", "src/a.go": "package src\n"})
	provider := &stubRelationshipProvider{}
	dataset := codemap.Dataset{
		Documents: []codemap.DocumentRecord{{Path: "docs/map.md"}},
		Entries: []codemap.DatasetEntry{{
			Entry:      codemap.Entry{DocumentPath: "docs/map.md", Target: "src/a.go", Kind: codemap.TargetFile},
			Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved, ResolvedPath: "src/a.go", Exists: true},
		}},
	}
	corpus, err := BuildContext(context.Background(), root, dataset, Options{RelationshipProvider: provider})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := corpus.InputContext(context.Background(), "docs/map.md", nil); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 0 {
		t.Fatalf("hidden holdout seed triggered %d relationship request(s)", provider.calls)
	}
}

func TestRelationshipSeedsPreserveVerifiedSymbolIdentity(t *testing.T) {
	dataset := codemap.Dataset{Entries: []codemap.DatasetEntry{{
		Entry: codemap.Entry{DocumentPath: "docs/map.md", Target: "src/a.go#Run", Kind: codemap.TargetSymbol},
		Resolution: codemap.TargetRecord{
			Status: codemap.ResolutionResolved, ResolvedPath: "src/a.go", Exists: true,
			SemanticNode: &codemap.SemanticNode{Identity: "sha256:node", Kind: "function", Path: "src/a.go", Name: "Run", QualifiedName: "src/a.go::Run"},
		},
	}}}
	seeds := relationshipSeeds(dataset)["docs/map.md"]
	if len(seeds) != 1 || seeds[0].Identity != "sha256:node" || seeds[0].Name != "Run" || seeds[0].Kind != "function" {
		t.Fatalf("symbol seeds = %#v", seeds)
	}
}
