package codemapcorpus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

func Build(repositoryRoot string, dataset codemap.Dataset, options Options) (Corpus, error) {
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
	documents, err := loadDocuments(root, dataset)
	if err != nil {
		return Corpus{}, err
	}
	targets := resolvedTargets(dataset)
	dependencies, err := collectDependencies(root, files)
	if err != nil {
		return Corpus{}, err
	}
	symbols, err := collectSymbolDeclarations(root, files)
	if err != nil {
		return Corpus{}, err
	}
	commits, err := collectCommits(root, files, options)
	if err != nil {
		return Corpus{}, err
	}
	return Corpus{
		RepositoryRoot:     root,
		RepositoryFiles:    files,
		RepositoryPaths:    paths,
		Documents:          documents,
		TargetsByDocument:  targets,
		DependencyEdges:    dependencies,
		Commits:            commits,
		RelatedDocuments:   collectRelatedDocuments(documents, targets),
		SymbolDeclarations: symbols,
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
