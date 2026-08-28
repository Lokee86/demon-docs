package codemaparcana

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

type Resolver struct {
	repositoryRoot string
	state          snapshotState
	client         queryClient
	openSnapshot   func(context.Context, string) (queryClient, error)

	historyMu         sync.Mutex
	historicalClients map[string]queryClient
	stalenessMu       sync.Mutex
	stalenessDiffs    map[string]semanticDiffResult
}

func OpenCurrent(ctx context.Context, repositoryRoot string) (*Resolver, Availability, error) {
	state, availability := discoverSnapshot(repositoryRoot)
	if !availability.Available {
		return nil, availability, nil
	}
	command, ok := findArcanaCommand()
	if !ok {
		availability.Available = false
		availability.Reason = "Arcana query executable is unavailable"
		return nil, availability, nil
	}
	resolver, err := openWithState(ctx, repositoryRoot, command, state)
	if err != nil {
		availability.Available = false
		availability.Reason = err.Error()
		return nil, availability, nil
	}
	return resolver, availability, nil
}

func OpenCurrentWithCommand(ctx context.Context, repositoryRoot, command string) (*Resolver, Availability, error) {
	state, availability := discoverSnapshot(repositoryRoot)
	if !availability.Available {
		return nil, availability, nil
	}
	resolver, err := openWithState(ctx, repositoryRoot, command, state)
	return resolver, availability, err
}

func openWithState(ctx context.Context, repositoryRoot, command string, state snapshotState) (*Resolver, error) {
	client, err := startProtocolClient(ctx, command, state.directory)
	if err != nil {
		return nil, fmt.Errorf("open Arcana snapshot %s: %w", state.id, err)
	}
	return &Resolver{
		repositoryRoot: repositoryRoot,
		state:          state,
		client:         client,
		openSnapshot: func(ctx context.Context, directory string) (queryClient, error) {
			return startProtocolClient(ctx, command, directory)
		},
		historicalClients: map[string]queryClient{},
		stalenessDiffs:    map[string]semanticDiffResult{},
	}, nil
}

func (resolver *Resolver) ResolveFile(ctx context.Context, path string) (codemap.SemanticResolution, error) {
	path = normalizeRepositoryPath(path)
	if path == "" || !resolver.pathCurrent(path) {
		return unsupported(), nil
	}
	var result nodeList
	if err := resolver.client.query(ctx, map[string]any{"op": "resolve_file", "path": path, "limit": 10000}, &result); err != nil {
		return codemap.SemanticResolution{}, err
	}
	return resolver.nodeResolution(result, ""), nil
}

func (resolver *Resolver) ResolveSymbol(ctx context.Context, query codemap.SymbolQuery) (codemap.SemanticResolution, error) {
	name := strings.TrimSpace(query.Name)
	if name == "" {
		return unsupported(), nil
	}
	path := normalizeRepositoryPath(query.Path)
	if query.Path != "" {
		if path == "" || !resolver.pathCurrent(path) {
			return unsupported(), nil
		}
	} else if !resolver.state.globalCurrent {
		return unsupported(), nil
	}
	request := map[string]any{"op": "resolve_symbol", "name": name, "limit": 10000}
	if path != "" {
		request["path"] = path
	}
	var result nodeList
	if err := resolver.client.query(ctx, request, &result); err != nil {
		return codemap.SemanticResolution{}, err
	}
	return resolver.nodeResolution(result, strings.TrimSpace(query.QualifiedName)), nil
}

func (resolver *Resolver) nodeResolution(result nodeList, qualifiedName string) codemap.SemanticResolution {
	if result.Truncated || result.Returned != len(result.Nodes) {
		return codemap.SemanticResolution{Status: codemap.SemanticAmbiguous, Nodes: semanticNodes(result.Nodes, qualifiedName, resolver.state.fileContentIDs)}
	}
	nodes := semanticNodes(result.Nodes, qualifiedName, resolver.state.fileContentIDs)
	switch len(nodes) {
	case 0:
		return codemap.SemanticResolution{Status: codemap.SemanticMissing}
	case 1:
		return codemap.SemanticResolution{Status: codemap.SemanticResolved, Nodes: nodes}
	default:
		return codemap.SemanticResolution{Status: codemap.SemanticAmbiguous, Nodes: nodes}
	}
}

func semanticNodes(nodes []protocolNode, qualifiedName string, files map[string]string) []codemap.SemanticNode {
	result := make([]codemap.SemanticNode, 0, len(nodes))
	for _, node := range nodes {
		path := normalizeRepositoryPath(node.Path)
		if path == "" {
			continue
		}
		if _, ok := files[path]; !ok {
			continue
		}
		if qualifiedName != "" && node.QualifiedName != qualifiedName {
			continue
		}
		item := codemap.SemanticNode{
			Key: node.Key, Identity: node.Identity, Kind: node.Kind, Path: path,
			Name: node.Name, QualifiedName: node.QualifiedName,
		}
		if node.Span != nil {
			item.Span = &codemap.SemanticSpan{
				Path: normalizeRepositoryPath(node.Span.Path), StartLine: node.Span.StartLine,
				StartColumn: node.Span.StartColumn, EndLine: node.Span.EndLine, EndColumn: node.Span.EndColumn,
			}
		}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		if result[i].QualifiedName != result[j].QualifiedName {
			return result[i].QualifiedName < result[j].QualifiedName
		}
		return result[i].Identity < result[j].Identity
	})
	return result
}

func (resolver *Resolver) pathCurrent(path string) bool {
	expected, ok := resolver.state.fileContentIDs[path]
	if !ok {
		return false
	}
	fullPath := filepath.Join(resolver.repositoryRoot, filepath.FromSlash(path))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return false
	}
	sum := sha256.Sum256(data)
	actual := "sha256:" + hex.EncodeToString(sum[:])
	return actual == expected
}

func unsupported() codemap.SemanticResolution {
	return codemap.SemanticResolution{Status: codemap.SemanticUnsupported}
}
