package codemaparcana

import (
	"context"
	"fmt"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
)

type relationshipQueryClient struct {
	calls             []map[string]any
	truncateList      bool
	truncateNeighbors bool
}

func (client *relationshipQueryClient) query(_ context.Context, request map[string]any, target any) error {
	copy := map[string]any{}
	for key, value := range request {
		copy[key] = value
	}
	client.calls = append(client.calls, copy)
	op, _ := request["op"].(string)
	switch op {
	case "list_nodes":
		result := target.(*nodeList)
		*result = nodeList{Count: 1, Returned: 1, Nodes: []protocolNode{{NodeID: 1, Identity: "a", Kind: "function", Path: "src/a.go", Name: "Run"}}}
		if client.truncateList {
			result.Count, result.Truncated = 2, true
		}
	case "resolve_symbol":
		result := target.(*nodeList)
		*result = nodeList{Count: 1, Returned: 1, Nodes: []protocolNode{{NodeID: 1, Identity: "a", Kind: "function", Path: "src/a.go", Name: "Run", QualifiedName: "src/a.go::Run"}}}
	case "neighbors":
		result := target.(*neighborList)
		direction, _ := request["direction"].(string)
		if direction == "outgoing" {
			*result = neighborList{Count: 1, Returned: 1, Relationships: []protocolRelationship{{
				Relation: "calls", Node: protocolNode{NodeID: 2, Identity: "b", Kind: "function", Path: "src/b.go", Name: "Work"},
			}}}
			if client.truncateNeighbors {
				result.Count, result.Truncated = 2, true
			}
		} else {
			*result = neighborList{}
		}
	default:
		return fmt.Errorf("unexpected operation %q", op)
	}
	return nil
}

func (*relationshipQueryClient) Close() error { return nil }

func TestRelationshipProviderProjectsBoundedCrossFileArcanaEdges(t *testing.T) {
	root := t.TempDir()
	a := "package src\nfunc Run() {}\n"
	b := "package src\nfunc Work() {}\n"
	writeTestStateFile(t, root+"/src/a.go", a)
	writeTestStateFile(t, root+"/src/b.go", b)
	client := &relationshipQueryClient{}
	resolver := &Resolver{
		repositoryRoot: root,
		state: snapshotState{fileContentIDs: map[string]string{
			"src/a.go": fileContentID(a), "src/b.go": fileContentID(b),
		}},
		client: client,
	}
	edges, err := resolver.CollectRelationships(context.Background(), codemapcorpus.RelationshipRequest{
		RepositoryRoot:  root,
		RepositoryFiles: []string{"src/a.go", "src/b.go"},
		Seeds:           []codemapcorpus.RelationshipSeed{{Path: "src/a.go", Kind: "file"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 1 || edges[0].Source != "src/a.go" || edges[0].Target != "src/b.go" || edges[0].Relation != "calls" {
		t.Fatalf("relationships = %#v", edges)
	}
	if len(client.calls) != 3 || client.calls[0]["op"] != "list_nodes" {
		t.Fatalf("Arcana calls = %#v", client.calls)
	}
}

func TestRelationshipProviderDiscardsTruncatedSeedNeighborhood(t *testing.T) {
	for name, configure := range map[string]func(*relationshipQueryClient){
		"node list": func(client *relationshipQueryClient) { client.truncateList = true },
		"neighbors": func(client *relationshipQueryClient) { client.truncateNeighbors = true },
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			a := "package src\nfunc Run() {}\n"
			b := "package src\nfunc Work() {}\n"
			writeTestStateFile(t, root+"/src/a.go", a)
			writeTestStateFile(t, root+"/src/b.go", b)
			client := &relationshipQueryClient{}
			configure(client)
			resolver := &Resolver{
				repositoryRoot: root,
				state: snapshotState{fileContentIDs: map[string]string{
					"src/a.go": fileContentID(a), "src/b.go": fileContentID(b),
				}},
				client: client,
			}
			edges, err := resolver.CollectRelationships(context.Background(), codemapcorpus.RelationshipRequest{
				RepositoryRoot: root, RepositoryFiles: []string{"src/a.go", "src/b.go"},
				Seeds: []codemapcorpus.RelationshipSeed{{Path: "src/a.go", Kind: "file"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(edges) != 0 {
				t.Fatalf("truncated neighborhood leaked partial edges: %#v", edges)
			}
		})
	}
}

func TestRelationshipProviderUsesExactVerifiedSymbolSeed(t *testing.T) {
	root := t.TempDir()
	a := "package src\nfunc Run() {}\n"
	b := "package src\nfunc Work() {}\n"
	writeTestStateFile(t, root+"/src/a.go", a)
	writeTestStateFile(t, root+"/src/b.go", b)
	client := &relationshipQueryClient{}
	resolver := &Resolver{
		repositoryRoot: root,
		state: snapshotState{fileContentIDs: map[string]string{
			"src/a.go": fileContentID(a), "src/b.go": fileContentID(b),
		}},
		client: client,
	}
	_, err := resolver.CollectRelationships(context.Background(), codemapcorpus.RelationshipRequest{
		RepositoryRoot:  root,
		RepositoryFiles: []string{"src/a.go", "src/b.go"},
		Seeds: []codemapcorpus.RelationshipSeed{{
			Path: "src/a.go", Identity: "a", Kind: "function", Name: "Run", QualifiedName: "src/a.go::Run",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(client.calls) == 0 || client.calls[0]["op"] != "resolve_symbol" {
		t.Fatalf("symbol seed did not resolve exact Arcana node: %#v", client.calls)
	}
}
