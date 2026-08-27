package codemapcorpus

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

const (
	DefaultMaxCommits        = 1000
	DefaultMaxPathsPerCommit = 200
)

type Options struct {
	MaxCommits        int
	MaxPathsPerCommit int
	CodeIntelligence  CodeIntelligenceProvider
}

type Corpus struct {
	RepositoryRoot            string
	RepositoryFiles           []string
	RepositoryPaths           []string
	Documents                 map[string]string
	TargetsByDocument         map[string][]string
	AuthoredTargetsByDocument map[string][]evidence.AuthoredTarget
	DependencyEdges           []evidence.DependencyEdge
	Commits                   []evidence.Commit
	RelatedDocuments          map[string][]evidence.RelatedDocument
	SymbolDeclarations        []evidence.SymbolDeclaration
}

func (c Corpus) KnownTargets(documentPath string) []string {
	return cloneStrings(c.TargetsByDocument[normalizePath(documentPath)])
}

func (c Corpus) Input(documentPath string, existingTargets []string) (evidence.Input, error) {
	documentPath = normalizePath(documentPath)
	text, ok := c.Documents[documentPath]
	if !ok {
		return evidence.Input{}, fmt.Errorf("document %s is not in the corpus", documentPath)
	}
	repositoryPaths := c.RepositoryPaths
	if len(repositoryPaths) == 0 {
		repositoryPaths = c.RepositoryFiles
	}
	return evidence.Input{
		DocumentPath:       documentPath,
		DocumentText:       text,
		RepositoryFiles:    repositoryPaths,
		ExistingTargets:    cloneStrings(existingTargets),
		AuthoredTargets:    visibleAuthoredTargets(c.AuthoredTargetsByDocument[documentPath], existingTargets),
		DependencyEdges:    c.DependencyEdges,
		Commits:            c.Commits,
		RelatedDocuments:   cloneRelated(c.RelatedDocuments[documentPath]),
		SymbolDeclarations: cloneSymbols(c.SymbolDeclarations),
	}, nil
}

func normalizeOptions(options Options) Options {
	if options.MaxCommits <= 0 {
		options.MaxCommits = DefaultMaxCommits
	}
	if options.MaxPathsPerCommit <= 0 {
		options.MaxPathsPerCommit = DefaultMaxPathsPerCommit
	}
	if options.CodeIntelligence == nil {
		options.CodeIntelligence = localCodeIntelligenceProvider{}
	}
	return options
}

func normalizePath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, `\`, "/"))
	value = strings.TrimPrefix(value, "./")
	if value == "" {
		return ""
	}
	clean := path.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return ""
	}
	return clean
}

func sortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}

func visibleAuthoredTargets(values []evidence.AuthoredTarget, visibleTargets []string) []evidence.AuthoredTarget {
	visible := map[string]struct{}{}
	for _, value := range visibleTargets {
		if normalized := normalizePath(value); normalized != "" {
			visible[normalized] = struct{}{}
		}
	}
	result := make([]evidence.AuthoredTarget, 0, len(values))
	for _, value := range values {
		item := evidence.AuthoredTarget{Target: value.Target, Kind: value.Kind}
		for _, resolved := range value.ResolvedTargets {
			normalized := normalizePath(resolved)
			if normalized == "" {
				continue
			}
			// Pattern and symbol entries are not exact-link holdout answers, so
			// their authored coverage remains visible while exact file/directory
			// targets honor the caller-provided visible target set.
			if value.Kind == evidence.AuthoredTargetPattern || value.Kind == evidence.AuthoredTargetSymbol {
				item.ResolvedTargets = append(item.ResolvedTargets, normalized)
				continue
			}
			if _, ok := visible[normalized]; ok {
				item.ResolvedTargets = append(item.ResolvedTargets, normalized)
			}
		}
		result = append(result, item)
	}
	return result
}

func cloneRelated(values []evidence.RelatedDocument) []evidence.RelatedDocument {
	result := make([]evidence.RelatedDocument, len(values))
	for index, value := range values {
		result[index] = evidence.RelatedDocument{Path: value.Path, Targets: cloneStrings(value.Targets)}
	}
	return result
}

func cloneSymbols(values []evidence.SymbolDeclaration) []evidence.SymbolDeclaration {
	return append([]evidence.SymbolDeclaration(nil), values...)
}
