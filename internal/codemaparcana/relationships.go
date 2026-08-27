package codemaparcana

import (
	"context"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

const (
	maxRelationshipNodesPerSeed = 128
	maxRelationshipNeighbors    = 128
)

var codemapRelations = []string{
	"calls",
	"imports",
	"depends-on",
	"implements",
	"extends",
	"overrides",
	"uses-trait",
	"includes",
	"tests",
}

type neighborList struct {
	Count         int                    `json:"count"`
	Returned      int                    `json:"returned"`
	Truncated     bool                   `json:"truncated"`
	Relationships []protocolRelationship `json:"relationships"`
}

type protocolRelationship struct {
	Relation string       `json:"relation"`
	Node     protocolNode `json:"node"`
}

func (resolver *Resolver) CollectRelationships(ctx context.Context, request codemapcorpus.RelationshipRequest) ([]evidence.RelationshipEdge, error) {
	known := make(map[string]struct{}, len(request.RepositoryFiles))
	for _, file := range request.RepositoryFiles {
		if path := normalizeRepositoryPath(file); path != "" {
			known[path] = struct{}{}
		}
	}
	values := map[string]evidence.RelationshipEdge{}
	for _, seed := range request.Seeds {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := normalizeRepositoryPath(seed.Path)
		if path == "" || !resolver.pathCurrent(path) {
			continue
		}
		edges, complete, err := resolver.relationshipsForSeed(ctx, seed, path, known)
		if err != nil {
			return nil, err
		}
		if !complete {
			continue
		}
		for _, edge := range edges {
			values[edge.Source+"\x00"+edge.Target+"\x00"+edge.Relation] = edge
		}
	}
	result := make([]evidence.RelationshipEdge, 0, len(values))
	for _, edge := range values {
		result = append(result, edge)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i].Source + "\x00" + result[i].Target + "\x00" + result[i].Relation
		right := result[j].Source + "\x00" + result[j].Target + "\x00" + result[j].Relation
		return left < right
	})
	return result, nil
}

func (resolver *Resolver) relationshipsForSeed(ctx context.Context, seed codemapcorpus.RelationshipSeed, path string, known map[string]struct{}) ([]evidence.RelationshipEdge, bool, error) {
	nodes, complete, err := resolver.relationshipNodes(ctx, seed, path)
	if err != nil || !complete {
		return nil, complete, err
	}
	var result []evidence.RelationshipEdge
	for _, node := range nodes {
		for _, direction := range []string{"outgoing", "incoming"} {
			var neighbors neighborList
			request := map[string]any{
				"op": "neighbors", "node_id": node.NodeID, "direction": direction,
				"relations": codemapRelations, "limit": maxRelationshipNeighbors,
			}
			if err := resolver.client.query(ctx, request, &neighbors); err != nil {
				return nil, false, err
			}
			if neighbors.Truncated || neighbors.Count != neighbors.Returned || neighbors.Returned != len(neighbors.Relationships) {
				return nil, false, nil
			}
			for _, relationship := range neighbors.Relationships {
				neighborPath := normalizeRepositoryPath(relationship.Node.Path)
				if neighborPath == "" || neighborPath == path || !resolver.pathCurrent(neighborPath) {
					continue
				}
				if _, ok := known[neighborPath]; !ok {
					continue
				}
				relation := strings.TrimSpace(relationship.Relation)
				if relation == "" {
					continue
				}
				if direction == "outgoing" {
					result = append(result, evidence.RelationshipEdge{Source: path, Target: neighborPath, Relation: relation})
				} else {
					result = append(result, evidence.RelationshipEdge{Source: neighborPath, Target: path, Relation: relation})
				}
			}
		}
	}
	return result, true, nil
}

func (resolver *Resolver) relationshipNodes(ctx context.Context, seed codemapcorpus.RelationshipSeed, path string) ([]protocolNode, bool, error) {
	if seed.Kind != "" && seed.Kind != "file" && (seed.Identity != "" || seed.Name != "") {
		request := map[string]any{"op": "resolve_symbol", "name": seed.Name, "path": path, "limit": 10000}
		var result nodeList
		if err := resolver.client.query(ctx, request, &result); err != nil {
			return nil, false, err
		}
		nodes := filterRelationshipSymbolNodes(result.Nodes, seed)
		return nodes, !result.Truncated && result.Count == result.Returned && result.Returned == len(result.Nodes) && len(nodes) <= maxRelationshipNodesPerSeed, nil
	}
	var result nodeList
	if err := resolver.client.query(ctx, map[string]any{"op": "list_nodes", "path_prefix": path, "limit": 10000}, &result); err != nil {
		return nil, false, err
	}
	if result.Truncated || result.Count != result.Returned || result.Returned != len(result.Nodes) {
		return nil, false, nil
	}
	nodes := make([]protocolNode, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		if normalizeRepositoryPath(node.Path) == path && relationshipCapableKind(node.Kind) {
			nodes = append(nodes, node)
		}
	}
	if len(nodes) > maxRelationshipNodesPerSeed {
		return nil, false, nil
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].NodeID < nodes[j].NodeID })
	return nodes, true, nil
}

func filterRelationshipSymbolNodes(nodes []protocolNode, seed codemapcorpus.RelationshipSeed) []protocolNode {
	result := make([]protocolNode, 0, len(nodes))
	for _, node := range nodes {
		if seed.Identity != "" && node.Identity != seed.Identity {
			continue
		}
		if seed.QualifiedName != "" && node.QualifiedName != seed.QualifiedName {
			continue
		}
		if relationshipCapableKind(node.Kind) {
			result = append(result, node)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].NodeID < result[j].NodeID })
	return result
}

func relationshipCapableKind(kind string) bool {
	switch kind {
	case "file", "module", "namespace", "type", "interface", "trait", "function", "method", "constructor", "test", "import", "export", "http-endpoint", "message-channel", "process", "cli-command", "protocol", "state-path":
		return true
	default:
		return false
	}
}
