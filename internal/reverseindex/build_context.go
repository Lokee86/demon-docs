package reverseindex

import (
	"context"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemaparcana"
	"github.com/Lokee86/demon-docs/internal/config"
)

func Build(repositoryRoot, docsRoot string, roots []string, c config.Config, format codemap.Format) (Plan, error) {
	return BuildContext(context.Background(), repositoryRoot, docsRoot, roots, c, format)
}

func BuildContext(ctx context.Context, repositoryRoot, docsRoot string, roots []string, c config.Config, format codemap.Format) (Plan, error) {
	resolver, _, err := codemaparcana.OpenCurrent(ctx, repositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	if resolver == nil {
		return buildWithResolver(ctx, repositoryRoot, docsRoot, roots, c, format, nil)
	}
	defer resolver.Close()
	return buildWithResolver(ctx, repositoryRoot, docsRoot, roots, c, format, resolver)
}
