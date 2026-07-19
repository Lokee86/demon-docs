package codemapcorpus

import (
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func collectRelatedDocuments(
	documents map[string]string,
	targets map[string][]string,
) map[string][]evidence.RelatedDocument {
	documentSet := make(map[string]struct{}, len(documents))
	for document := range documents {
		documentSet[document] = struct{}{}
	}
	related := make(map[string]map[string]struct{}, len(documents))
	for source, contents := range documents {
		for _, match := range markdownLinkPattern.FindAllStringSubmatch(contents, -1) {
			destination := markdownDestination(match[1])
			target := resolveDocumentLink(source, destination, documentSet)
			if target == "" || target == source {
				continue
			}
			addRelated(related, source, target)
			addRelated(related, target, source)
		}
	}

	result := make(map[string][]evidence.RelatedDocument, len(related))
	for document, paths := range related {
		ordered := sortedSet(paths)
		items := make([]evidence.RelatedDocument, 0, len(ordered))
		for _, relatedPath := range ordered {
			if len(targets[relatedPath]) == 0 {
				continue
			}
			items = append(items, evidence.RelatedDocument{
				Path:    relatedPath,
				Targets: cloneStrings(targets[relatedPath]),
			})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
		result[document] = items
	}
	return result
}

func markdownDestination(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "<") {
		if end := strings.Index(raw, ">"); end >= 0 {
			raw = raw[1:end]
		}
	} else if fields := strings.Fields(raw); len(fields) > 0 {
		raw = fields[0]
	}
	if index := strings.IndexAny(raw, "#?"); index >= 0 {
		raw = raw[:index]
	}
	return strings.TrimSpace(raw)
}

func resolveDocumentLink(source, destination string, documents map[string]struct{}) string {
	if destination == "" || strings.Contains(destination, "://") || strings.HasPrefix(destination, "#") {
		return ""
	}
	var candidate string
	if strings.HasPrefix(destination, "/") {
		candidate = normalizePath(strings.TrimPrefix(destination, "/"))
	} else {
		candidate = normalizePath(path.Join(path.Dir(source), destination))
	}
	if _, exists := documents[candidate]; exists {
		return candidate
	}
	return ""
}

func addRelated(related map[string]map[string]struct{}, source, target string) {
	if related[source] == nil {
		related[source] = map[string]struct{}{}
	}
	related[source][target] = struct{}{}
}
