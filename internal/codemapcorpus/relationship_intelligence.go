package codemapcorpus

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

// RelationshipProvider supplies bounded semantic relationships for the exact
// authored targets visible to one analysis request. It cannot choose candidates
// or mutate documentation.
type RelationshipProvider interface {
	CollectRelationships(context.Context, RelationshipRequest) ([]evidence.RelationshipEdge, error)
}

type RelationshipRequest struct {
	RepositoryRoot  string
	RepositoryFiles []string
	Seeds           []RelationshipSeed
}

type RelationshipSeed struct {
	Path          string
	Identity      string
	Kind          string
	Name          string
	QualifiedName string
}

func relationshipSeeds(dataset codemap.Dataset) map[string][]RelationshipSeed {
	sets := map[string]map[string]RelationshipSeed{}
	for _, item := range dataset.Entries {
		if item.Resolution.Status != codemap.ResolutionResolved || item.Resolution.ResolvedPath == "" {
			continue
		}
		if item.Entry.Kind != codemap.TargetFile && item.Entry.Kind != codemap.TargetSymbol {
			continue
		}
		document := normalizePath(item.Entry.DocumentPath)
		path := normalizePath(item.Resolution.ResolvedPath)
		if document == "" || path == "" {
			continue
		}
		seed := RelationshipSeed{Path: path, Kind: "file"}
		if item.Entry.Kind == codemap.TargetSymbol {
			if item.Resolution.SemanticNode == nil {
				continue
			}
			node := item.Resolution.SemanticNode
			seed.Identity = strings.TrimSpace(node.Identity)
			seed.Kind = strings.TrimSpace(node.Kind)
			seed.Name = strings.TrimSpace(node.Name)
			seed.QualifiedName = strings.TrimSpace(node.QualifiedName)
		}
		if sets[document] == nil {
			sets[document] = map[string]RelationshipSeed{}
		}
		sets[document][relationshipSeedKey(seed)] = seed
	}
	result := make(map[string][]RelationshipSeed, len(sets))
	for document, values := range sets {
		for _, seed := range values {
			result[document] = append(result[document], seed)
		}
		sortRelationshipSeeds(result[document])
	}
	return result
}

func visibleRelationshipSeeds(seeds []RelationshipSeed, visibleTargets []string) []RelationshipSeed {
	visible := make(map[string]struct{}, len(visibleTargets))
	for _, target := range visibleTargets {
		if normalized := normalizePath(target); normalized != "" {
			visible[normalized] = struct{}{}
		}
	}
	result := make([]RelationshipSeed, 0, len(seeds))
	for _, seed := range seeds {
		if _, ok := visible[normalizePath(seed.Path)]; ok {
			result = append(result, seed)
		}
	}
	sortRelationshipSeeds(result)
	return result
}

func normalizeRelationshipEdges(request RelationshipRequest, edges []evidence.RelationshipEdge) ([]evidence.RelationshipEdge, error) {
	knownFiles := normalizedRepositoryFiles(request.RepositoryFiles)
	values := map[string]evidence.RelationshipEdge{}
	for _, edge := range edges {
		source := normalizePath(edge.Source)
		target := normalizePath(edge.Target)
		relation := strings.TrimSpace(edge.Relation)
		if source == "" || target == "" || relation == "" || source == target {
			continue
		}
		if _, ok := knownFiles[source]; !ok {
			return nil, fmt.Errorf("relationship source %q is not a repository file", edge.Source)
		}
		if _, ok := knownFiles[target]; !ok {
			return nil, fmt.Errorf("relationship target %q is not a repository file", edge.Target)
		}
		item := evidence.RelationshipEdge{Source: source, Target: target, Relation: relation}
		values[source+"\x00"+target+"\x00"+relation] = item
	}
	result := make([]evidence.RelationshipEdge, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i].Source + "\x00" + result[i].Target + "\x00" + result[i].Relation
		right := result[j].Source + "\x00" + result[j].Target + "\x00" + result[j].Relation
		return left < right
	})
	return result, nil
}

func relationshipSeedKey(seed RelationshipSeed) string {
	return seed.Path + "\x00" + seed.Identity + "\x00" + seed.QualifiedName
}

func sortRelationshipSeeds(values []RelationshipSeed) {
	sort.Slice(values, func(i, j int) bool { return relationshipSeedKey(values[i]) < relationshipSeedKey(values[j]) })
}
