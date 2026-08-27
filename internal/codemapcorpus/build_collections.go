package codemapcorpus

import (
	"context"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

type corpusCollections struct {
	documents    map[string]string
	dependencies []evidence.DependencyEdge
	symbols      []evidence.SymbolDeclaration
	commits      []evidence.Commit
}

type documentCollectionResult struct {
	documents map[string]string
	err       error
}

type sourceCollectionResult struct {
	facts CodeIntelligenceFacts
	err   error
}

type commitCollectionResult struct {
	commits []evidence.Commit
	err     error
}

func collectCorpusCollections(
	ctx context.Context,
	root string,
	files []string,
	dataset codemap.Dataset,
	options Options,
) (corpusCollections, error) {
	documents := make(chan documentCollectionResult, 1)
	sources := make(chan sourceCollectionResult, 1)
	commits := make(chan commitCollectionResult, 1)

	go func() {
		items, err := loadDocuments(root, dataset)
		documents <- documentCollectionResult{documents: items, err: err}
	}()
	go func() {
		request := CodeIntelligenceRequest{RepositoryRoot: root, RepositoryFiles: cloneStrings(files)}
		facts, err := options.CodeIntelligence.Collect(ctx, request)
		if err == nil {
			facts, err = normalizeCodeIntelligence(request, facts)
		}
		sources <- sourceCollectionResult{facts: facts, err: err}
	}()
	go func() {
		items, err := collectCommits(root, files, options)
		commits <- commitCollectionResult{commits: items, err: err}
	}()

	documentResult := <-documents
	sourceResult := <-sources
	commitResult := <-commits
	for _, err := range []error{documentResult.err, sourceResult.err, commitResult.err} {
		if err != nil {
			return corpusCollections{}, err
		}
	}
	return corpusCollections{
		documents:    documentResult.documents,
		dependencies: sourceResult.facts.DependencyEdges,
		symbols:      sourceResult.facts.SymbolDeclarations,
		commits:      commitResult.commits,
	}, nil
}
