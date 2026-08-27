package codemaprun

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

type runTargetResolver struct{}

func (runTargetResolver) ResolveFile(context.Context, string) (codemap.SemanticResolution, error) {
	return codemap.SemanticResolution{Status: codemap.SemanticUnsupported}, nil
}

func (runTargetResolver) ResolveSymbol(_ context.Context, query codemap.SymbolQuery) (codemap.SemanticResolution, error) {
	if query.Name != "RemoteSymbol" {
		return codemap.SemanticResolution{Status: codemap.SemanticMissing}, nil
	}
	return codemap.SemanticResolution{Status: codemap.SemanticResolved, Nodes: []codemap.SemanticNode{{
		Identity: "sha256:remote", Kind: "function", Path: "src/runtime.go", Name: "RemoteSymbol",
	}}}, nil
}

func TestBuildUsesTargetResolverForAuthoredSymbolCoverage(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\nRemoteSymbol owns execution.\n\n## Code Map\n\n- `symbol:RemoteSymbol`\n")
	writeFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n\nfunc RemoteSymbol() {}\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot: root,
		DocsRoot:       docs,
		TargetFiles:    []string{filepath.Join(docs, "runtime.md")},
		Headings:       []string{"Code Map"},
		MarkerPrefix:   "ddocs",
		TargetResolver: runTargetResolver{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := plan.Documents[0].Recommendations; len(got) != 0 {
		t.Fatalf("resolved authored symbol did not suppress backing-file suggestion: %#v", got)
	}
}
