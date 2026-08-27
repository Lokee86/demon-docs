package codemapcorpus

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

// CodeIntelligenceProvider supplies deterministic repository-local semantic
// facts. Demon Docs owns candidate policy; providers only report facts.
type CodeIntelligenceProvider interface {
	Collect(context.Context, CodeIntelligenceRequest) (CodeIntelligenceFacts, error)
}

type CodeIntelligenceRequest struct {
	RepositoryRoot  string
	RepositoryFiles []string
}

type CodeIntelligenceFacts struct {
	DependencyEdges    []evidence.DependencyEdge
	SymbolDeclarations []evidence.SymbolDeclaration
}

type localCodeIntelligenceProvider struct{}

func (localCodeIntelligenceProvider) Collect(ctx context.Context, request CodeIntelligenceRequest) (CodeIntelligenceFacts, error) {
	if err := ctx.Err(); err != nil {
		return CodeIntelligenceFacts{}, err
	}
	dependencies, symbols, err := collectSourceFacts(request.RepositoryRoot, request.RepositoryFiles)
	if err != nil {
		return CodeIntelligenceFacts{}, err
	}
	if err := ctx.Err(); err != nil {
		return CodeIntelligenceFacts{}, err
	}
	return CodeIntelligenceFacts{DependencyEdges: dependencies, SymbolDeclarations: symbols}, nil
}

func normalizeCodeIntelligence(request CodeIntelligenceRequest, facts CodeIntelligenceFacts) (CodeIntelligenceFacts, error) {
	knownFiles := normalizedRepositoryFiles(request.RepositoryFiles)
	dependencies := map[string]evidence.DependencyEdge{}
	for _, item := range facts.DependencyEdges {
		sourceRaw := strings.TrimSpace(item.Source)
		targetRaw := strings.TrimSpace(item.Target)
		relation := strings.TrimSpace(item.Relation)
		if sourceRaw == "" || targetRaw == "" || relation == "" {
			continue
		}
		source := normalizePath(sourceRaw)
		target := normalizePath(targetRaw)
		if source == "" {
			return CodeIntelligenceFacts{}, fmt.Errorf("invalid code-intelligence dependency source %q", item.Source)
		}
		if target == "" {
			return CodeIntelligenceFacts{}, fmt.Errorf("invalid code-intelligence dependency target %q", item.Target)
		}
		if source == target {
			continue
		}
		if _, ok := knownFiles[source]; !ok {
			return CodeIntelligenceFacts{}, fmt.Errorf("code-intelligence dependency source %q is not a repository file", item.Source)
		}
		if _, ok := knownFiles[target]; !ok {
			return CodeIntelligenceFacts{}, fmt.Errorf("code-intelligence dependency target %q is not a repository file", item.Target)
		}
		normalized := evidence.DependencyEdge{Source: source, Target: target, Relation: relation}
		dependencies[source+"\x00"+target+"\x00"+relation] = normalized
	}

	symbols := map[string]evidence.SymbolDeclaration{}
	for _, item := range facts.SymbolDeclarations {
		pathRaw := strings.TrimSpace(item.Path)
		symbol := strings.TrimSpace(item.Symbol)
		if pathRaw == "" || symbol == "" {
			continue
		}
		file := normalizePath(pathRaw)
		if file == "" {
			return CodeIntelligenceFacts{}, fmt.Errorf("invalid code-intelligence symbol path %q", item.Path)
		}
		if _, ok := knownFiles[file]; !ok {
			return CodeIntelligenceFacts{}, fmt.Errorf("code-intelligence symbol path %q is not a repository file", item.Path)
		}
		normalized := evidence.SymbolDeclaration{Path: file, Symbol: symbol}
		symbols[file+"\x00"+symbol] = normalized
	}

	return CodeIntelligenceFacts{
		DependencyEdges:    sortedDependencyFacts(dependencies),
		SymbolDeclarations: sortedSymbolFacts(symbols),
	}, nil
}

func normalizedRepositoryFiles(files []string) map[string]struct{} {
	result := make(map[string]struct{}, len(files))
	for _, file := range files {
		if normalized := normalizePath(file); normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}

func sortedDependencyFacts(values map[string]evidence.DependencyEdge) []evidence.DependencyEdge {
	result := make([]evidence.DependencyEdge, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i].Source + "\x00" + result[i].Target + "\x00" + result[i].Relation
		right := result[j].Source + "\x00" + result[j].Target + "\x00" + result[j].Relation
		return left < right
	})
	return result
}

func sortedSymbolFacts(values map[string]evidence.SymbolDeclaration) []evidence.SymbolDeclaration {
	result := make([]evidence.SymbolDeclaration, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path != result[j].Path {
			return result[i].Path < result[j].Path
		}
		return result[i].Symbol < result[j].Symbol
	})
	return result
}
