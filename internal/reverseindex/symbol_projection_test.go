package reverseindex

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
)

type reverseSymbolResolver struct {
	resolution codemap.SemanticResolution
}

func (resolver reverseSymbolResolver) ResolveFile(context.Context, string) (codemap.SemanticResolution, error) {
	return codemap.SemanticResolution{Status: codemap.SemanticUnsupported}, nil
}

func (resolver reverseSymbolResolver) ResolveSymbol(context.Context, codemap.SymbolQuery) (codemap.SemanticResolution, error) {
	return resolver.resolution, nil
}

func TestBuildProjectsVerifiedAuthoredSymbolSeparatelyFromFileDocumentation(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "service")
	mustWrite(t, filepath.Join(docsRoot, "runtime.md"), "# Runtime Guide\n\n## Code map\n\n- `service/runtime.go#Run`\n")
	mustWrite(t, filepath.Join(codeRoot, "runtime.go"), "package service\n\nfunc Run() {}\n")

	resolver := reverseSymbolResolver{resolution: codemap.SemanticResolution{
		Status: codemap.SemanticResolved,
		Nodes: []codemap.SemanticNode{{
			Key: "run-key", Identity: "sha256:run", Kind: "function", Path: "service/runtime.go",
			Name: "Run", QualifiedName: "service/runtime.go::Run",
			Span: &codemap.SemanticSpan{Path: "service/runtime.go", StartLine: 3, StartColumn: 1, EndLine: 3, EndColumn: 14},
		}},
	}}
	plan, err := buildWithResolver(context.Background(), repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat(), resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || len(plan.Diagnostics) != 0 {
		t.Fatalf("updates=%d diagnostics=%v", len(plan.Updates), plan.Diagnostics)
	}
	text := plan.Updates[0].NewText
	for _, expected := range []string{
		"- [runtime.go](runtime.go)",
		"  - Symbol `service/runtime.go::Run` (function, line 3)",
		"    - [Runtime Guide](../docs/runtime.md)",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q:\n%s", expected, text)
		}
	}
	if strings.Contains(text, "\n  - [Runtime Guide](../docs/runtime.md)") {
		t.Fatalf("verified symbol was flattened into file documentation:\n%s", text)
	}
	if len(plan.Orphans) != 0 {
		t.Fatalf("verified symbol did not cover its backing file: %v", plan.Orphans)
	}
}

func TestBuildProjectsStandaloneVerifiedSymbolOntoBackingFile(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "service")
	mustWrite(t, filepath.Join(docsRoot, "runtime.md"), "# Runtime Guide\n\n## Code map\n\n- `symbol:service::Run`\n")
	mustWrite(t, filepath.Join(codeRoot, "runtime.go"), "package service\n\nfunc Run() {}\n")

	resolver := reverseSymbolResolver{resolution: codemap.SemanticResolution{
		Status: codemap.SemanticResolved,
		Nodes: []codemap.SemanticNode{{
			Key: "run-key", Identity: "sha256:run", Kind: "function", Path: "service/runtime.go",
			Name: "Run", QualifiedName: "service::Run",
			Span: &codemap.SemanticSpan{Path: "service/runtime.go", StartLine: 3, StartColumn: 1, EndLine: 3, EndColumn: 14},
		}},
	}}
	plan, err := buildWithResolver(context.Background(), repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat(), resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || !strings.Contains(plan.Updates[0].NewText, "Symbol `service::Run` (function, line 3)") {
		t.Fatalf("standalone symbol projection missing:\n%s", plan.Updates[0].NewText)
	}
	if len(plan.Orphans) != 0 {
		t.Fatalf("standalone symbol did not cover its backing file: %v", plan.Orphans)
	}
}

func TestBuildKeepsUnverifiedPathSymbolAsFileLevelFallback(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "service")
	mustWrite(t, filepath.Join(docsRoot, "runtime.md"), "# Runtime Guide\n\n## Code map\n\n- `service/runtime.go#Run`\n")
	mustWrite(t, filepath.Join(codeRoot, "runtime.go"), "package service\n")

	plan, err := buildWithResolver(context.Background(), repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat(), nil)
	if err != nil {
		t.Fatal(err)
	}
	text := plan.Updates[0].NewText
	if strings.Contains(text, "Symbol `") || !strings.Contains(text, "\n  - [Runtime Guide](../docs/runtime.md)") {
		t.Fatalf("unverified path symbol did not retain file-level fallback:\n%s", text)
	}
}

func TestBuildReportsAmbiguousStandaloneSymbolWhenCandidateIsInReverseScope(t *testing.T) {
	repositoryRoot := t.TempDir()
	docsRoot := filepath.Join(repositoryRoot, "docs")
	codeRoot := filepath.Join(repositoryRoot, "service")
	mustWrite(t, filepath.Join(docsRoot, "runtime.md"), "# Runtime Guide\n\n## Code map\n\n- `symbol:Run`\n")
	mustWrite(t, filepath.Join(codeRoot, "runtime.go"), "package service\n")

	resolver := reverseSymbolResolver{resolution: codemap.SemanticResolution{
		Status: codemap.SemanticAmbiguous,
		Nodes: []codemap.SemanticNode{
			{Key: "run-a", Identity: "a", Kind: "function", Path: "service/runtime.go", Name: "Run", QualifiedName: "service/runtime.go::Run"},
			{Key: "run-b", Identity: "b", Kind: "function", Path: "outside/runtime.go", Name: "Run", QualifiedName: "outside/runtime.go::Run"},
		},
	}}
	plan, err := buildWithResolver(context.Background(), repositoryRoot, docsRoot, []string{codeRoot}, config.Default(), codemap.DefaultFormat(), resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Diagnostics) != 1 || !strings.Contains(plan.Diagnostics[0], "ambiguous target symbol:Run") {
		t.Fatalf("diagnostics=%v", plan.Diagnostics)
	}
}
