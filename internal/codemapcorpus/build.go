package codemapcorpus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

func Build(repositoryRoot string, dataset codemap.Dataset, options Options) (Corpus, error) {
	return BuildContext(context.Background(), repositoryRoot, dataset, options)
}

func BuildContext(ctx context.Context, repositoryRoot string, dataset codemap.Dataset, options Options) (Corpus, error) {
	if err := ctx.Err(); err != nil {
		return Corpus{}, err
	}
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Corpus{}, err
	}
	options = normalizeOptions(options)
	files, err := repositoryFiles(root)
	if err != nil {
		return Corpus{}, err
	}
	paths := repositoryPaths(files)
	targets := resolvedTargets(dataset)
	authoredTargets := authoredTargets(dataset)
	collections, err := collectCorpusCollections(ctx, root, files, dataset, options)
	if err != nil {
		return Corpus{}, err
	}
	return Corpus{
		RepositoryRoot:            root,
		RepositoryFiles:           files,
		RepositoryPaths:           paths,
		Documents:                 collections.documents,
		TargetsByDocument:         targets,
		AuthoredTargetsByDocument: authoredTargets,
		DependencyEdges:           collections.dependencies,
		Commits:                   collections.commits,
		RelatedDocuments:          collectRelatedDocuments(collections.documents, directFileTargets(authoredTargets)),
		SymbolDeclarations:        collections.symbols,
	}, nil
}

func loadDocuments(root string, dataset codemap.Dataset) (map[string]string, error) {
	documents := make(map[string]string, len(dataset.Documents))
	for _, record := range dataset.Documents {
		documentPath := normalizePath(record.Path)
		if documentPath == "" {
			return nil, fmt.Errorf("invalid document path %q", record.Path)
		}
		fullPath := filepath.Join(root, filepath.FromSlash(documentPath))
		if !within(root, fullPath) {
			return nil, fmt.Errorf("document %s is outside repository", documentPath)
		}
		contents, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("read document %s: %w", documentPath, err)
		}
		documents[documentPath] = string(contents)
	}
	return documents, nil
}

func authoredTargets(dataset codemap.Dataset) map[string][]evidence.AuthoredTarget {
	result := map[string][]evidence.AuthoredTarget{}
	for _, item := range dataset.Entries {
		document := normalizePath(item.Entry.DocumentPath)
		if document == "" {
			continue
		}
		authored := evidence.AuthoredTarget{
			Target: normalizeAuthoredTarget(item.Entry.Target),
			Kind:   authoredTargetKind(item.Entry.Kind),
		}
		switch item.Resolution.Status {
		case codemap.ResolutionResolved, codemap.ResolutionSymbolUnverified:
			if target := normalizePath(item.Resolution.ResolvedPath); target != "" {
				authored.ResolvedTargets = []string{target}
			}
		case codemap.ResolutionPatternResolved:
			for _, match := range item.Resolution.Matches {
				if target := normalizePath(match.Path); target != "" {
					authored.ResolvedTargets = append(authored.ResolvedTargets, target)
				}
			}
		}
		result[document] = append(result[document], authored)
	}
	return result
}

func authoredTargetKind(kind codemap.TargetKind) evidence.AuthoredTargetKind {
	switch kind {
	case codemap.TargetFile:
		return evidence.AuthoredTargetFile
	case codemap.TargetDirectory:
		return evidence.AuthoredTargetDirectory
	case codemap.TargetGlob:
		return evidence.AuthoredTargetPattern
	case codemap.TargetSymbol:
		return evidence.AuthoredTargetSymbol
	default:
		return evidence.AuthoredTargetUnknown
	}
}

func normalizeAuthoredTarget(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func directFileTargets(targets map[string][]evidence.AuthoredTarget) map[string][]string {
	result := make(map[string][]string, len(targets))
	for document, values := range targets {
		set := map[string]struct{}{}
		for _, value := range values {
			if value.Kind != evidence.AuthoredTargetFile {
				continue
			}
			for _, resolved := range value.ResolvedTargets {
				if normalized := normalizePath(resolved); normalized != "" {
					set[normalized] = struct{}{}
				}
			}
		}
		result[document] = sortedSet(set)
	}
	return result
}

func resolvedTargets(dataset codemap.Dataset) map[string][]string {
	sets := map[string]map[string]struct{}{}
	for _, item := range dataset.Entries {
		document := normalizePath(item.Entry.DocumentPath)
		if document == "" {
			continue
		}
		if sets[document] == nil {
			sets[document] = map[string]struct{}{}
		}
		switch item.Resolution.Status {
		case codemap.ResolutionResolved, codemap.ResolutionSymbolUnverified, codemap.ResolutionKindMismatch:
			if target := normalizePath(item.Resolution.ResolvedPath); target != "" {
				sets[document][target] = struct{}{}
			}
		case codemap.ResolutionPatternResolved:
			for _, match := range item.Resolution.Matches {
				if target := normalizePath(match.Path); target != "" {
					sets[document][target] = struct{}{}
				}
			}
		}
	}
	result := make(map[string][]string, len(sets))
	for document, targets := range sets {
		result[document] = sortedSet(targets)
	}
	return result
}
