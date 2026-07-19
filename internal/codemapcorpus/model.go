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
}

type Corpus struct {
	RepositoryRoot    string
	RepositoryFiles   []string
	RepositoryPaths   []string
	Documents         map[string]string
	TargetsByDocument map[string][]string
	DependencyEdges   []evidence.DependencyEdge
	Commits           []evidence.Commit
	RelatedDocuments  map[string][]evidence.RelatedDocument
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
		DocumentPath:     documentPath,
		DocumentText:     text,
		RepositoryFiles:  repositoryPaths,
		ExistingTargets:  cloneStrings(existingTargets),
		DependencyEdges:  c.DependencyEdges,
		Commits:          c.Commits,
		RelatedDocuments: cloneRelated(c.RelatedDocuments[documentPath]),
	}, nil
}

func normalizeOptions(options Options) Options {
	if options.MaxCommits <= 0 {
		options.MaxCommits = DefaultMaxCommits
	}
	if options.MaxPathsPerCommit <= 0 {
		options.MaxPathsPerCommit = DefaultMaxPathsPerCommit
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

func cloneRelated(values []evidence.RelatedDocument) []evidence.RelatedDocument {
	result := make([]evidence.RelatedDocument, len(values))
	for index, value := range values {
		result[index] = evidence.RelatedDocument{Path: value.Path, Targets: cloneStrings(value.Targets)}
	}
	return result
}
