package codemapcorpus

import (
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
	dependencies []evidence.DependencyEdge
	symbols      []evidence.SymbolDeclaration
	err          error
}

type commitCollectionResult struct {
	commits []evidence.Commit
	err     error
}

func collectCorpusCollections(
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
		dependencies, symbols, err := collectSourceFacts(root, files)
		sources <- sourceCollectionResult{dependencies: dependencies, symbols: symbols, err: err}
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
		dependencies: sourceResult.dependencies,
		symbols:      sourceResult.symbols,
		commits:      commitResult.commits,
	}, nil
}
