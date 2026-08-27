package codemap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func resolveStandaloneSymbol(
	ctx context.Context,
	repositoryRoot string,
	entry Entry,
	resolver TargetResolver,
	contentCache *targetContentCache,
) (TargetRecord, error) {
	if resolver == nil {
		return TargetRecord{Status: ResolutionUnsupported}, nil
	}
	query, ok := standaloneSymbolQuery(entry.Target)
	if !ok {
		return TargetRecord{Status: ResolutionUnsupported}, nil
	}
	resolution, err := resolver.ResolveSymbol(ctx, query)
	if err != nil {
		return TargetRecord{}, err
	}
	return semanticTargetRecord(repositoryRoot, resolution, contentCache)
}

func resolvePathSymbol(
	ctx context.Context,
	record TargetRecord,
	target string,
	resolver TargetResolver,
) (TargetRecord, error) {
	if resolver == nil {
		record.Status = ResolutionSymbolUnverified
		return record, nil
	}
	name := pathSymbolName(target)
	if name == "" {
		record.Status = ResolutionUnsupported
		return record, nil
	}
	resolution, err := resolver.ResolveSymbol(ctx, SymbolQuery{Name: name, Path: record.ResolvedPath})
	if err != nil {
		return TargetRecord{}, err
	}
	switch resolution.Status {
	case SemanticResolved:
		if len(resolution.Nodes) != 1 {
			return TargetRecord{}, fmt.Errorf("semantic resolver returned %d nodes for resolved symbol %q", len(resolution.Nodes), target)
		}
		record.Status = ResolutionResolved
		record.SemanticNode = cloneSemanticNode(resolution.Nodes[0])
	case SemanticMissing:
		record.Status = ResolutionMissing
		record.Exists = false
	case SemanticAmbiguous:
		record.Status = ResolutionAmbiguous
		record.Exists = false
		setSemanticCandidates(&record, resolution.Nodes)
	case SemanticUnsupported:
		record.Status = ResolutionSymbolUnverified
	default:
		return TargetRecord{}, fmt.Errorf("semantic resolver returned unknown status %q", resolution.Status)
	}
	return record, nil
}

func attachFileSemantic(ctx context.Context, record TargetRecord, resolver TargetResolver) (TargetRecord, error) {
	if resolver == nil || record.ResolvedPath == "" {
		return record, nil
	}
	resolution, err := resolver.ResolveFile(ctx, record.ResolvedPath)
	if err != nil {
		return TargetRecord{}, err
	}
	if resolution.Status == SemanticResolved && len(resolution.Nodes) == 1 {
		record.SemanticNode = cloneSemanticNode(resolution.Nodes[0])
	} else if resolution.Status == SemanticAmbiguous {
		setSemanticCandidates(&record, resolution.Nodes)
	}
	return record, nil
}

func semanticTargetRecord(repositoryRoot string, resolution SemanticResolution, contentCache *targetContentCache) (TargetRecord, error) {
	switch resolution.Status {
	case SemanticUnsupported:
		return TargetRecord{Status: ResolutionUnsupported}, nil
	case SemanticMissing:
		return TargetRecord{Status: ResolutionMissing}, nil
	case SemanticAmbiguous:
		record := TargetRecord{Status: ResolutionAmbiguous}
		setSemanticCandidates(&record, resolution.Nodes)
		return record, nil
	case SemanticResolved:
		if len(resolution.Nodes) != 1 {
			return TargetRecord{}, fmt.Errorf("semantic resolver returned %d nodes for resolved target", len(resolution.Nodes))
		}
		node := resolution.Nodes[0]
		fullPath := filepath.Join(repositoryRoot, filepath.FromSlash(node.Path))
		if !within(repositoryRoot, fullPath) {
			return TargetRecord{}, fmt.Errorf("semantic resolver returned path outside repository: %s", node.Path)
		}
		info, err := os.Stat(fullPath)
		if err != nil {
			return TargetRecord{}, fmt.Errorf("stat semantic target %s: %w", node.Path, err)
		}
		if info.IsDir() {
			return TargetRecord{}, fmt.Errorf("semantic symbol target is a directory: %s", node.Path)
		}
		hash, err := contentCache.hashFile(fullPath)
		if err != nil {
			return TargetRecord{}, err
		}
		return TargetRecord{
			Status:       ResolutionResolved,
			ResolvedPath: filepath.ToSlash(filepath.Clean(node.Path)),
			Exists:       true,
			Size:         info.Size(),
			SHA256:       hash,
			SemanticNode: cloneSemanticNode(node),
		}, nil
	default:
		return TargetRecord{}, fmt.Errorf("semantic resolver returned unknown status %q", resolution.Status)
	}
}

func standaloneSymbolQuery(target string) (SymbolQuery, bool) {
	value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(target), "symbol:"))
	if value == "" || strings.Contains(value, "#") {
		return SymbolQuery{}, false
	}
	if index := strings.LastIndex(value, "::"); index >= 0 {
		name := strings.TrimSpace(value[index+2:])
		if name == "" {
			return SymbolQuery{}, false
		}
		return SymbolQuery{Name: name, QualifiedName: value}, true
	}
	return SymbolQuery{Name: value}, true
}

func pathSymbolName(target string) string {
	if index := strings.Index(target, "#"); index >= 0 {
		return strings.TrimSpace(target[index+1:])
	}
	if index := strings.LastIndex(target, "::"); index >= 0 {
		return strings.TrimSpace(target[index+2:])
	}
	return ""
}

func setSemanticCandidates(record *TargetRecord, nodes []SemanticNode) {
	record.SemanticCandidates = cloneSemanticNodes(nodes)
	set := map[string]struct{}{}
	for _, node := range record.SemanticCandidates {
		label := node.QualifiedName
		if label == "" {
			label = node.Path + "#" + node.Name
		}
		if label != "" {
			set[label] = struct{}{}
		}
	}
	for label := range set {
		record.Candidates = append(record.Candidates, label)
	}
	sort.Strings(record.Candidates)
}

func cloneSemanticNodes(nodes []SemanticNode) []SemanticNode {
	result := make([]SemanticNode, len(nodes))
	for index, node := range nodes {
		result[index] = *cloneSemanticNode(node)
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

func cloneSemanticNode(node SemanticNode) *SemanticNode {
	copy := node
	if node.Span != nil {
		span := *node.Span
		copy.Span = &span
	}
	return &copy
}
