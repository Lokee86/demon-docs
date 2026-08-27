package codemapcorpus

import (
	"context"
	"fmt"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func (c Corpus) InputContext(ctx context.Context, documentPath string, existingTargets []string) (evidence.Input, error) {
	if err := ctx.Err(); err != nil {
		return evidence.Input{}, err
	}
	documentPath = normalizePath(documentPath)
	text, ok := c.Documents[documentPath]
	if !ok {
		return evidence.Input{}, fmt.Errorf("document %s is not in the corpus", documentPath)
	}
	repositoryPaths := c.RepositoryPaths
	if len(repositoryPaths) == 0 {
		repositoryPaths = c.RepositoryFiles
	}
	input := evidence.Input{
		DocumentPath:       documentPath,
		DocumentText:       text,
		RepositoryFiles:    repositoryPaths,
		ExistingTargets:    cloneStrings(existingTargets),
		AuthoredTargets:    visibleAuthoredTargets(c.AuthoredTargetsByDocument[documentPath], existingTargets),
		DependencyEdges:    c.DependencyEdges,
		Commits:            c.Commits,
		RelatedDocuments:   cloneRelated(c.RelatedDocuments[documentPath]),
		SymbolDeclarations: cloneSymbols(c.SymbolDeclarations),
	}
	if c.relationshipProvider == nil {
		return input, nil
	}
	seeds := visibleRelationshipSeeds(c.relationshipSeeds[documentPath], existingTargets)
	if len(seeds) == 0 {
		return input, nil
	}
	request := RelationshipRequest{
		RepositoryRoot: c.RepositoryRoot, RepositoryFiles: cloneStrings(c.RepositoryFiles), Seeds: seeds,
	}
	edges, err := c.relationshipProvider.CollectRelationships(ctx, request)
	if err != nil {
		return evidence.Input{}, err
	}
	input.SemanticRelationships, err = normalizeRelationshipEdges(request, edges)
	if err != nil {
		return evidence.Input{}, err
	}
	return input, nil
}
