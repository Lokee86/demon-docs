package codemap

import (
	"context"
	"path/filepath"
	"testing"
)

type fakeTargetResolver struct {
	file   func(string) SemanticResolution
	symbol func(SymbolQuery) SemanticResolution
}

func (resolver fakeTargetResolver) ResolveFile(_ context.Context, path string) (SemanticResolution, error) {
	if resolver.file == nil {
		return SemanticResolution{Status: SemanticUnsupported}, nil
	}
	return resolver.file(path), nil
}

func (resolver fakeTargetResolver) ResolveSymbol(_ context.Context, query SymbolQuery) (SemanticResolution, error) {
	if resolver.symbol == nil {
		return SemanticResolution{Status: SemanticUnsupported}, nil
	}
	return resolver.symbol(query), nil
}

func TestBuildDatasetContextResolvesPathQualifiedSymbol(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/main.go", "package main\nfunc Run() {}\n")
	writeFixture(t, repository, "docs/guide.md", "## Code map\n\n- `src/main.go#Run`\n")
	resolver := fakeTargetResolver{symbol: func(query SymbolQuery) SemanticResolution {
		if query.Path != "src/main.go" || query.Name != "Run" {
			t.Fatalf("unexpected symbol query: %#v", query)
		}
		return SemanticResolution{Status: SemanticResolved, Nodes: []SemanticNode{{
			Identity: "sha256:node", Kind: "function", Path: "src/main.go", Name: "Run", QualifiedName: "src/main.go::Run",
		}}}
	}}

	dataset, err := BuildDatasetContext(context.Background(), repository, filepath.Join(repository, "docs"), DefaultFormat(), resolver)
	if err != nil {
		t.Fatal(err)
	}
	resolution := dataset.Entries[0].Resolution
	if resolution.Status != ResolutionResolved || resolution.SemanticNode == nil || resolution.SemanticNode.Identity != "sha256:node" {
		t.Fatalf("unexpected semantic resolution: %#v", resolution)
	}
}

func TestBuildDatasetContextResolvesStandaloneSymbol(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/runtime.go", "package src\nfunc Start() {}\n")
	writeFixture(t, repository, "docs/guide.md", "## Code map\n\n- `symbol:Start`\n")
	resolver := fakeTargetResolver{symbol: func(query SymbolQuery) SemanticResolution {
		if query.Name != "Start" || query.Path != "" {
			t.Fatalf("unexpected symbol query: %#v", query)
		}
		return SemanticResolution{Status: SemanticResolved, Nodes: []SemanticNode{{
			Identity: "sha256:start", Kind: "function", Path: "src/runtime.go", Name: "Start",
		}}}
	}}

	dataset, err := BuildDatasetContext(context.Background(), repository, filepath.Join(repository, "docs"), DefaultFormat(), resolver)
	if err != nil {
		t.Fatal(err)
	}
	resolution := dataset.Entries[0].Resolution
	if resolution.Status != ResolutionResolved || resolution.ResolvedPath != "src/runtime.go" || !resolution.Exists {
		t.Fatalf("unexpected standalone resolution: %#v", resolution)
	}
}

func TestBuildDatasetContextPreservesExplicitSemanticOutcomes(t *testing.T) {
	for _, test := range []struct {
		name   string
		status SemanticResolutionStatus
		want   ResolutionStatus
	}{
		{"missing", SemanticMissing, ResolutionMissing},
		{"ambiguous", SemanticAmbiguous, ResolutionAmbiguous},
		{"unsupported", SemanticUnsupported, ResolutionUnsupported},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := t.TempDir()
			writeFixture(t, repository, "src/a.go", "package src\n")
			writeFixture(t, repository, "src/b.go", "package src\n")
			writeFixture(t, repository, "docs/guide.md", "## Code map\n\n- `symbol:Thing`\n")
			resolver := fakeTargetResolver{symbol: func(SymbolQuery) SemanticResolution {
				result := SemanticResolution{Status: test.status}
				if test.status == SemanticAmbiguous {
					result.Nodes = []SemanticNode{{Path: "src/a.go", Name: "Thing"}, {Path: "src/b.go", Name: "Thing"}}
				}
				return result
			}}
			dataset, err := BuildDatasetContext(context.Background(), repository, filepath.Join(repository, "docs"), DefaultFormat(), resolver)
			if err != nil {
				t.Fatal(err)
			}
			if got := dataset.Entries[0].Resolution.Status; got != test.want {
				t.Fatalf("status = %s, want %s", got, test.want)
			}
		})
	}
}

func TestBuildDatasetContextAttachesArcanaFileIdentityWithoutChangingFilesystemTruth(t *testing.T) {
	repository := t.TempDir()
	writeFixture(t, repository, "src/main.go", "package main\n")
	writeFixture(t, repository, "docs/guide.md", "## Code map\n\n- `src/main.go`\n")
	resolver := fakeTargetResolver{file: func(path string) SemanticResolution {
		if path != "src/main.go" {
			t.Fatalf("file query = %q", path)
		}
		return SemanticResolution{Status: SemanticResolved, Nodes: []SemanticNode{{
			Identity: "sha256:file", Kind: "file", Path: path, Name: path, QualifiedName: path,
		}}}
	}}
	dataset, err := BuildDatasetContext(context.Background(), repository, filepath.Join(repository, "docs"), DefaultFormat(), resolver)
	if err != nil {
		t.Fatal(err)
	}
	resolution := dataset.Entries[0].Resolution
	if resolution.Status != ResolutionResolved || resolution.SemanticNode == nil || resolution.SemanticNode.Identity != "sha256:file" {
		t.Fatalf("unexpected file resolution: %#v", resolution)
	}
}
