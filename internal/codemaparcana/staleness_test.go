package codemaparcana

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

type fakeStalenessClient struct {
	diff    semanticDiffResult
	results map[string]nodeList
}

func (client *fakeStalenessClient) query(_ context.Context, request map[string]any, target any) error {
	if request["op"] == "diff" {
		*target.(*semanticDiffResult) = client.diff
		return nil
	}
	*target.(*nodeList) = client.results[stalenessRequestKey(request)]
	return nil
}

func (*fakeStalenessClient) Close() error { return nil }

func stalenessRequestKey(request map[string]any) string {
	operation, _ := request["op"].(string)
	path, _ := request["path"].(string)
	name, _ := request["name"].(string)
	return operation + "|" + path + "|" + name
}

func testSnapshotDirectory(t *testing.T, root, snapshotID string) string {
	t.Helper()
	directory := filepath.Join(root, ".arcana", "snapshots", snapshotID[len("sha256:"):])
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"repository.manifest", "lexicon.snapshot"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func TestAnalyzeStalenessReportsMappedSemanticChanges(t *testing.T) {
	root := t.TempDir()
	previousID := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testSnapshotDirectory(t, root, previousID)
	currentRun := protocolNode{Key: "run", Identity: "current-run", Kind: "function", Path: "src/new.go", Name: "Run", QualifiedName: "pkg::Run"}
	previousRun := protocolNode{Key: "old-run", Identity: "old-run", Kind: "function", Path: "src/old.go", Name: "Run", QualifiedName: "pkg::Run"}
	currentStable := protocolNode{Key: "stable", Identity: "stable-current", Kind: "file", Path: "src/stable.go", Name: "stable.go"}
	previousStable := protocolNode{Key: "stable", Identity: "stable-old", Kind: "file", Path: "src/stable.go", Name: "stable.go"}
	previousGone := protocolNode{Key: "gone", Identity: "gone", Kind: "file", Path: "src/gone.go", Name: "gone.go"}

	currentClient := &fakeStalenessClient{}
	currentClient.diff.Nodes.Added = []protocolNode{currentRun}
	currentClient.diff.Nodes.Removed = []protocolNode{previousRun, previousGone}
	currentClient.diff.Nodes.MetadataChanged = []protocolNode{currentStable}
	currentClient.diff.Nodes.RelationshipChanged = []protocolNode{currentStable}
	historical := &fakeStalenessClient{results: map[string]nodeList{
		"resolve_symbol||Run":         {Count: 1, Returned: 1, Nodes: []protocolNode{previousRun}},
		"resolve_file|src/stable.go|": {Count: 1, Returned: 1, Nodes: []protocolNode{previousStable}},
		"resolve_file|src/gone.go|":   {Count: 1, Returned: 1, Nodes: []protocolNode{previousGone}},
	}}
	resolver := &Resolver{
		repositoryRoot: root,
		state:          snapshotState{id: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		client:         currentClient,
		openSnapshot:   func(context.Context, string) (queryClient, error) { return historical, nil },
	}
	entries := []codemap.DatasetEntry{
		{Entry: codemap.Entry{Target: "symbol:pkg::Run", Kind: codemap.TargetSymbol}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved, SemanticNode: &codemap.SemanticNode{Key: currentRun.Key, Identity: currentRun.Identity, Kind: currentRun.Kind, Path: currentRun.Path, Name: currentRun.Name, QualifiedName: currentRun.QualifiedName}}},
		{Entry: codemap.Entry{Target: "src/stable.go", Kind: codemap.TargetFile}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionResolved, ResolvedPath: "src/stable.go", SemanticNode: &codemap.SemanticNode{Key: currentStable.Key, Identity: currentStable.Identity, Kind: "file", Path: "src/stable.go", Name: "stable.go"}}},
		{Entry: codemap.Entry{Target: "src/gone.go", Kind: codemap.TargetFile}, Resolution: codemap.TargetRecord{Status: codemap.ResolutionMissing, ResolvedPath: "src/gone.go"}},
	}
	changes, err := resolver.AnalyzeStaleness(context.Background(), previousID, entries)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[codemap.SemanticChangeKind]bool{}
	for _, change := range changes {
		seen[change.Kind] = true
	}
	for _, kind := range []codemap.SemanticChangeKind{codemap.SemanticChangeMoved, codemap.SemanticChangeMetadata, codemap.SemanticChangeRelationships, codemap.SemanticChangeDisappeared} {
		if !seen[kind] {
			t.Fatalf("missing %s change: %#v", kind, changes)
		}
	}
}

func TestAnalyzeStalenessRecognizesStableKeyMoveForPathQualifiedSymbol(t *testing.T) {
	root := t.TempDir()
	previousID := "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	testSnapshotDirectory(t, root, previousID)
	contents := "package runtime\nfunc Run() {}\n"
	writeTestStateFile(t, filepath.Join(root, "src", "new.go"), contents)
	previous := protocolNode{Key: "run-key", Identity: "old-run", Kind: "function", Path: "src/old.go", Name: "Run", QualifiedName: "src/old.go::Run"}
	current := protocolNode{Key: "run-key", Identity: "new-run", Kind: "function", Path: "src/new.go", Name: "Run", QualifiedName: "src/new.go::Run"}
	currentClient := &fakeStalenessClient{}
	currentClient.diff.Nodes.MetadataChanged = []protocolNode{current}
	historical := &fakeStalenessClient{results: map[string]nodeList{
		"resolve_symbol|src/old.go|Run": {Count: 1, Returned: 1, Nodes: []protocolNode{previous}},
	}}
	resolver := &Resolver{
		repositoryRoot: root,
		state: snapshotState{
			id:             "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			fileContentIDs: map[string]string{"src/new.go": fileContentID(contents)},
		},
		client:       currentClient,
		openSnapshot: func(context.Context, string) (queryClient, error) { return historical, nil },
	}
	entries := []codemap.DatasetEntry{{
		Entry:      codemap.Entry{Target: "src/old.go#Run", Kind: codemap.TargetSymbol},
		Resolution: codemap.TargetRecord{Status: codemap.ResolutionMissing, ResolvedPath: "src/old.go"},
	}}
	changes, err := resolver.AnalyzeStaleness(context.Background(), previousID, entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) < 1 || changes[0].Kind != codemap.SemanticChangeMoved || changes[0].CurrentPath != "src/new.go" {
		t.Fatalf("stable-key move = %#v", changes)
	}
}
