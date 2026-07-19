package evidence

import (
	"path"
	"sort"
	"strings"
)

type Kind string

const (
	KindExactPathMention      Kind = "exact_path_mention"
	KindUniqueBasenameMention Kind = "unique_basename_mention"
	KindSiblingTarget         Kind = "sibling_of_existing_target"
	KindTestCounterpart       Kind = "test_counterpart"
	KindDependencyNeighbor    Kind = "dependency_neighbor"
	KindGitDocumentCoChange   Kind = "git_cochange_with_document"
	KindGitTargetCoChange     Kind = "git_cochange_with_existing_target"
	KindRelatedDocumentTarget Kind = "related_document_target"
)

type DependencyEdge struct {
	Source   string
	Target   string
	Relation string
}

type Commit struct {
	ID    string
	Paths []string
}

type RelatedDocument struct {
	Path    string
	Targets []string
}

type Input struct {
	DocumentPath     string
	DocumentText     string
	RepositoryFiles  []string
	ExistingTargets  []string
	DependencyEdges  []DependencyEdge
	Commits          []Commit
	RelatedDocuments []RelatedDocument
}

type Evidence struct {
	Kind   Kind
	Source string
	Detail string
	Count  int
}

type Candidate struct {
	Path        string
	Evidence    []Evidence
	Fingerprint string
}

func normalizePath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, `\`, "/"))
	value = strings.TrimPrefix(value, "./")
	if value == "" {
		return ""
	}
	directory := strings.HasSuffix(value, "/")
	clean := path.Clean(value)
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." || strings.HasPrefix(clean, "/") {
		return ""
	}
	if directory {
		clean += "/"
	}
	return clean
}

func normalizedSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if normalized := normalizePath(value); normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
