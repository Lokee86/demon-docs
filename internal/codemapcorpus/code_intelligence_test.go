package codemapcorpus

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

type stubCodeIntelligenceProvider struct {
	facts       CodeIntelligenceFacts
	err         error
	request     CodeIntelligenceRequest
	contextSeen bool
}

type providerContextKey struct{}

func (provider *stubCodeIntelligenceProvider) Collect(ctx context.Context, request CodeIntelligenceRequest) (CodeIntelligenceFacts, error) {
	provider.request = request
	provider.contextSeen = ctx.Value(providerContextKey{}) == "expected"
	return provider.facts, provider.err
}

func TestBuildContextUsesInjectedCodeIntelligenceProvider(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"docs/map.md": "# Map\n",
		"src/a.go":    "package src\n\ntype LocalOnly struct{}\n",
		"src/b.go":    "package src\n",
	})
	provider := &stubCodeIntelligenceProvider{facts: CodeIntelligenceFacts{
		DependencyEdges: []evidence.DependencyEdge{
			{Source: "./src/a.go", Target: "src/b.go", Relation: " calls "},
			{Source: "src/a.go", Target: "src/b.go", Relation: "calls"},
		},
		SymbolDeclarations: []evidence.SymbolDeclaration{
			{Path: "./src/a.go", Symbol: " Runner "},
			{Path: "src/a.go", Symbol: "Runner"},
		},
	}}
	ctx := context.WithValue(context.Background(), providerContextKey{}, "expected")
	corpus, err := BuildContext(ctx, root, codemap.Dataset{
		Documents: []codemap.DocumentRecord{{Path: "docs/map.md"}},
	}, Options{CodeIntelligence: provider})
	if err != nil {
		t.Fatal(err)
	}
	if !provider.contextSeen {
		t.Fatal("provider did not receive caller context")
	}
	if provider.request.RepositoryRoot == "" || !contains(provider.request.RepositoryFiles, "src/a.go") {
		t.Fatalf("provider request = %#v", provider.request)
	}
	wantDependencies := []evidence.DependencyEdge{{Source: "src/a.go", Target: "src/b.go", Relation: "calls"}}
	if !reflect.DeepEqual(corpus.DependencyEdges, wantDependencies) {
		t.Fatalf("dependency facts = %#v", corpus.DependencyEdges)
	}
	wantSymbols := []evidence.SymbolDeclaration{{Path: "src/a.go", Symbol: "Runner"}}
	if !reflect.DeepEqual(corpus.SymbolDeclarations, wantSymbols) {
		t.Fatalf("symbol facts = %#v", corpus.SymbolDeclarations)
	}
}

func TestBuildContextRejectsProviderFactsOutsideRepositoryFiles(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"docs/map.md": "# Map\n", "src/a.go": "package src\n"})
	provider := &stubCodeIntelligenceProvider{facts: CodeIntelligenceFacts{
		SymbolDeclarations: []evidence.SymbolDeclaration{{Path: "../other.go", Symbol: "Other"}},
	}}
	_, err := BuildContext(context.Background(), root, codemap.Dataset{
		Documents: []codemap.DocumentRecord{{Path: "docs/map.md"}},
	}, Options{CodeIntelligence: provider})
	if err == nil || !strings.Contains(err.Error(), "code-intelligence symbol path") {
		t.Fatalf("expected provider path rejection, got %v", err)
	}
}

func TestBuildContextPropagatesProviderError(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"docs/map.md": "# Map\n"})
	providerErr := errors.New("provider unavailable")
	_, err := BuildContext(context.Background(), root, codemap.Dataset{
		Documents: []codemap.DocumentRecord{{Path: "docs/map.md"}},
	}, Options{CodeIntelligence: &stubCodeIntelligenceProvider{err: providerErr}})
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestBuildContextHonorsCancellationBeforeCollection(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := BuildContext(ctx, root, codemap.Dataset{}, Options{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
