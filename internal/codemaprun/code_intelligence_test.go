package codemaprun

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

type runCodeIntelligenceProvider struct{}

func (runCodeIntelligenceProvider) Collect(context.Context, codemapcorpus.CodeIntelligenceRequest) (codemapcorpus.CodeIntelligenceFacts, error) {
	return codemapcorpus.CodeIntelligenceFacts{
		SymbolDeclarations: []evidence.SymbolDeclaration{{Path: "src/runtime.go", Symbol: "RemoteSymbol"}},
	}, nil
}

func TestBuildPassesCodeIntelligenceProviderIntoCorpus(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	writeFile(t, filepath.Join(docs, "runtime.md"), "# Runtime\n\nRemoteSymbol owns execution.\n\n## Code Map\n")
	writeFile(t, filepath.Join(root, "src", "runtime.go"), "package runtime\n")

	plan, err := Build(context.Background(), Options{
		RepositoryRoot:   root,
		DocsRoot:         docs,
		TargetFiles:      []string{filepath.Join(docs, "runtime.md")},
		Headings:         []string{"Code Map"},
		MarkerPrefix:     "ddocs",
		CodeIntelligence: runCodeIntelligenceProvider{},
	})
	if err != nil {
		t.Fatal(err)
	}
	document := plan.Documents[0]
	if len(document.Recommendations) != 1 || document.Recommendations[0].Tier != codemaprecommend.SuggestionTierHardLink {
		t.Fatalf("provider recommendation = %#v", document.Recommendations)
	}
	if len(document.Added) != 1 || document.Added[0] != "src/runtime.go" {
		t.Fatalf("provider hard link was not selected: %#v", document.Added)
	}
}
