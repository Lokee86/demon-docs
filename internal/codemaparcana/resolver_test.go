package codemaparcana

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

type fakeQueryClient struct {
	calls  []map[string]any
	result nodeList
}

func (client *fakeQueryClient) query(_ context.Context, request map[string]any, target any) error {
	copy := map[string]any{}
	for key, value := range request {
		copy[key] = value
	}
	client.calls = append(client.calls, copy)
	if output, ok := target.(*nodeList); ok {
		*output = client.result
	}
	return nil
}

func (*fakeQueryClient) Close() error { return nil }

func TestResolverRequiresCurrentLexiconContentForPathQueries(t *testing.T) {
	root := t.TempDir()
	contents := "package main\nfunc Run() {}\n"
	path := "src/main.go"
	writeTestStateFile(t, filepath.Join(root, filepath.FromSlash(path)), contents)
	client := &fakeQueryClient{result: nodeList{Count: 1, Returned: 1, Nodes: []protocolNode{{
		Identity: "sha256:node", Kind: "function", Path: path, Name: "Run", QualifiedName: path + "::Run",
	}}}}
	resolver := &Resolver{
		repositoryRoot: root,
		state:          snapshotState{fileContentIDs: map[string]string{path: fileContentID(contents)}},
		client:         client,
	}

	resolution, err := resolver.ResolveSymbol(context.Background(), codemap.SymbolQuery{Name: "Run", Path: path})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != codemap.SemanticResolved || len(resolution.Nodes) != 1 {
		t.Fatalf("unexpected resolution: %#v", resolution)
	}
	if len(client.calls) != 1 || client.calls[0]["op"] != "resolve_symbol" || client.calls[0]["path"] != path {
		t.Fatalf("unexpected Arcana calls: %#v", client.calls)
	}

	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path)), []byte(contents+"// dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolution, err = resolver.ResolveSymbol(context.Background(), codemap.SymbolQuery{Name: "Run", Path: path})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != codemap.SemanticUnsupported || len(client.calls) != 1 {
		t.Fatalf("stale path was queried: resolution=%#v calls=%#v", resolution, client.calls)
	}
}

func TestResolverGlobalSymbolRequiresRepositoryCurrentEvidence(t *testing.T) {
	path := "src/main.go"
	client := &fakeQueryClient{result: nodeList{Count: 1, Returned: 1, Nodes: []protocolNode{{
		Identity: "sha256:node", Kind: "function", Path: path, Name: "Run", QualifiedName: path + "::Run",
	}}}}
	resolver := &Resolver{state: snapshotState{fileContentIDs: map[string]string{path: "sha256:any"}}, client: client}
	resolution, err := resolver.ResolveSymbol(context.Background(), codemap.SymbolQuery{Name: "Run"})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != codemap.SemanticUnsupported || len(client.calls) != 0 {
		t.Fatalf("global stale resolution was queried: %#v calls=%#v", resolution, client.calls)
	}

	resolver.state.globalCurrent = true
	resolution, err = resolver.ResolveSymbol(context.Background(), codemap.SymbolQuery{Name: "Run"})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != codemap.SemanticResolved || len(client.calls) != 1 {
		t.Fatalf("current global resolution failed: %#v calls=%#v", resolution, client.calls)
	}
}

func TestResolverFiltersQualifiedNameAndReportsAmbiguity(t *testing.T) {
	files := map[string]string{"a.go": "sha256:a", "b.go": "sha256:b"}
	client := &fakeQueryClient{result: nodeList{Count: 2, Returned: 2, Nodes: []protocolNode{
		{Identity: "one", Kind: "function", Path: "a.go", Name: "Run", QualifiedName: "a.go::Run"},
		{Identity: "two", Kind: "function", Path: "b.go", Name: "Run", QualifiedName: "b.go::Run"},
	}}}
	resolver := &Resolver{state: snapshotState{fileContentIDs: files, globalCurrent: true}, client: client}
	resolution, err := resolver.ResolveSymbol(context.Background(), codemap.SymbolQuery{Name: "Run"})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != codemap.SemanticAmbiguous || len(resolution.Nodes) != 2 {
		t.Fatalf("expected ambiguity, got %#v", resolution)
	}
	resolution, err = resolver.ResolveSymbol(context.Background(), codemap.SymbolQuery{Name: "Run", QualifiedName: "b.go::Run"})
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Status != codemap.SemanticResolved || len(resolution.Nodes) != 1 || resolution.Nodes[0].Path != "b.go" {
		t.Fatalf("qualified resolution = %#v", resolution)
	}
}

func fileContentID(contents string) string {
	sum := sha256.Sum256([]byte(contents))
	return "sha256:" + hex.EncodeToString(sum[:])
}
