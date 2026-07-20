package codemap

import (
	"fmt"
	"sort"
	"strings"
)

// SectionPlacement lets the document-format schema place a missing codemap
// section without coupling codemap execution to one schema implementation.
type SectionPlacement struct {
	Heading string
	Level   int
	Offset  int
}

// SectionSchema reports whether a file type requires a codemap section and,
// when it does, where that section belongs. Existing sections bypass it.
type SectionSchema interface {
	CodemapSection(documentPath, source string) (SectionPlacement, bool, error)
}

type ManagedUpdate struct {
	AddTargets    []string
	RemoveTargets []string
}

type ManagedResult struct {
	Text           string
	SectionFound   bool
	SectionCreated bool
	Added          []string
	Removed        []string
}

// ReconcileManaged adopts the complete configured codemap section, removes
// only explicitly selected targets, and appends missing targets inside one
// codemap-specific managed region. It never creates a file.
func ReconcileManaged(documentPath, source string, format Format, markerPrefix string, update ManagedUpdate, schema SectionSchema) (ManagedResult, error) {
	span, found, err := locateSection(source, format)
	if err != nil {
		return ManagedResult{}, err
	}
	created := false
	if !found {
		if schema == nil {
			return ManagedResult{Text: source}, nil
		}
		placement, required, err := schema.CodemapSection(documentPath, source)
		if err != nil {
			return ManagedResult{}, err
		}
		if !required {
			return ManagedResult{Text: source}, nil
		}
		source, format, err = insertSchemaSection(source, format, placement)
		if err != nil {
			return ManagedResult{}, err
		}
		span, found, err = locateSection(source, format)
		if err != nil || !found {
			if err == nil {
				err = fmt.Errorf("schema-created codemap section could not be located")
			}
			return ManagedResult{}, err
		}
		created = true
	}

	removed := []string{}
	if removeSet := normalizedTargetSet(update.RemoveTargets); len(removeSet) > 0 {
		source, removed = removeEntryLines(documentPath, source, format, removeSet)
		span, _, err = locateSection(source, format)
		if err != nil {
			return ManagedResult{}, err
		}
	}

	startMarker := fmt.Sprintf("<!-- %s:codemap:start -->", markerPrefix)
	endMarker := fmt.Sprintf("<!-- %s:codemap:end -->", markerPrefix)
	body, err := removeManagedMarkers(source[span.bodyStart:span.bodyEnd], startMarker, endMarker)
	if err != nil {
		return ManagedResult{}, err
	}

	existing := Extract(documentPath, source, format)
	existingTargets := make(map[string]struct{}, len(existing.Entries))
	for _, entry := range existing.Entries {
		existingTargets[normalizeTarget(entry.Target)] = struct{}{}
	}
	additions := missingTargets(update.AddTargets, existingTargets)
	content := strings.Trim(body, "\n")
	if len(additions) > 0 {
		content = appendTargetsInSection(content, additions, existing.Entries)
	}

	updated := replaceSectionBody(source, span, startMarker, endMarker, content)
	return ManagedResult{
		Text:           updated,
		SectionFound:   true,
		SectionCreated: created,
		Added:          additions,
		Removed:        removed,
	}, nil
}

func missingTargets(values []string, existing map[string]struct{}) []string {
	additions := make([]string, 0, len(values))
	for target := range normalizedTargetSet(values) {
		if _, present := existing[target]; !present {
			additions = append(additions, target)
		}
	}
	sort.Strings(additions)
	return additions
}
