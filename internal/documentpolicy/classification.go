package documentpolicy

import (
	"fmt"
	"strings"
)

type nodeLocation struct {
	parent   *markdownSection
	parentID string
}

type documentClassification struct {
	byNode      map[*markdownSection]Section
	byID        map[string][]*markdownSection
	locations   map[*markdownSection]nodeLocation
	diagnostics []Diagnostic
}

func classifyDocument(roots []*markdownSection, current, previous Schema, hasPrevious bool) documentClassification {
	classification := documentClassification{
		byNode:    map[*markdownSection]Section{},
		byID:      map[string][]*markdownSection{},
		locations: map[*markdownSection]nodeLocation{},
	}
	previousByID := map[string]Section{}
	if hasPrevious {
		for _, section := range previous.Sections {
			previousByID[section.ID] = section
		}
	}
	var walk func([]*markdownSection, *markdownSection, string)
	walk = func(children []*markdownSection, parent *markdownSection, parentID string) {
		localDefinitions := childrenForParent(current.Sections, parentID)
		localMatches := groupMatches(children, localDefinitions, previousByID)
		for _, child := range children {
			classification.locations[child] = nodeLocation{parent: parent, parentID: parentID}
			definition, matched := localMatches.byNode[child]
			if !matched {
				candidates := matchingDefinitions(child.Heading, current.Sections, previousByID)
				switch len(candidates) {
				case 1:
					definition = candidates[0]
					matched = true
				case 2:
					classification.diagnostics = append(classification.diagnostics, Diagnostic{
						Section: child.Heading,
						Message: fmt.Sprintf("section heading is ambiguous between schema sections %q and %q", candidates[0].ID, candidates[1].ID),
						Options: []string{"repair manually"},
					})
				default:
					if len(candidates) > 2 {
						classification.diagnostics = append(classification.diagnostics, Diagnostic{
							Section: child.Heading,
							Message: "section heading is ambiguous across multiple schema sections",
							Options: []string{"repair manually"},
						})
					}
				}
			}
			if !matched {
				continue
			}
			classification.byNode[child] = definition
			classification.byID[definition.ID] = append(classification.byID[definition.ID], child)
			walk(child.Children, child, definition.ID)
		}
	}
	walk(roots, nil, "")
	return classification
}

func matchingDefinitions(heading string, definitions []Section, previousByID map[string]Section) []Section {
	var matches []Section
	for _, definition := range definitions {
		canonical := strings.EqualFold(strings.TrimSpace(heading), strings.TrimSpace(definition.Heading))
		alias := matchesAlias(heading, definition.Aliases)
		renamed := false
		if !canonical && !alias {
			if prior, ok := previousByID[definition.ID]; ok && !strings.EqualFold(prior.Heading, definition.Heading) {
				renamed = strings.EqualFold(strings.TrimSpace(heading), strings.TrimSpace(prior.Heading)) || matchesAlias(heading, prior.Aliases)
			}
		}
		if canonical || alias || renamed {
			matches = append(matches, definition)
		}
	}
	return matches
}

func manualConflictDiagnostics(roots []*markdownSection, classification documentClassification, schema Schema) []Diagnostic {
	diagnostics := append([]Diagnostic(nil), classification.diagnostics...)
	var walk func([]*markdownSection)
	walk = func(children []*markdownSection) {
		for _, child := range children {
			definition, matched := classification.byNode[child]
			if !matched {
				if strings.EqualFold(schema.UnknownSections, "manual") || schema.UnknownSections == "" {
					diagnostics = append(diagnostics, Diagnostic{
						Section: child.Heading,
						Message: "unknown human-authored section requires a document-specific schema or explicit deletion",
						Options: []string{"ignore", "delete", "repair manually"},
					})
				}
				continue
			}
			_ = definition
			walk(child.Children)
		}
	}
	walk(roots)
	for _, definition := range schema.Sections {
		nodes := classification.byID[definition.ID]
		if len(nodes) > 1 && !definition.AllowDuplicates && (strings.EqualFold(schema.DuplicateSections, "manual") || schema.DuplicateSections == "") {
			diagnostics = append(diagnostics, Diagnostic{
				Section: definition.Heading,
				Message: fmt.Sprintf("duplicate section has %d occurrences", len(nodes)),
				Options: []string{"merge", "delete an occurrence", "ignore", "repair manually"},
			})
		}
	}
	return diagnostics
}

func relocateKnownSections(document *markdownDocument, classification documentClassification, schema Schema, resolveDiagnostics bool) ([]Diagnostic, bool, bool) {
	type move struct {
		node       *markdownSection
		location   nodeLocation
		target     *markdownSection
		parentID   string
		parentName string
	}
	var diagnostics []Diagnostic
	var moves []move
	blocked := false
	for _, definition := range schema.Sections {
		for _, node := range classification.byID[definition.ID] {
			location := classification.locations[node]
			if location.parentID == definition.Parent {
				continue
			}
			target := (*markdownSection)(nil)
			if definition.Parent != "" {
				parents := classification.byID[definition.Parent]
				if len(parents) != 1 {
					diagnostics = append(diagnostics, Diagnostic{
						Section: node.Heading,
						Message: fmt.Sprintf("section belongs under %q but that parent is missing or duplicated", definition.Parent),
						Options: []string{"repair manually"},
					})
					blocked = true
					continue
				}
				target = parents[0]
			}
			parentName := "the document root"
			if target != nil {
				parentName = fmt.Sprintf("%q", target.Heading)
			}
			moves = append(moves, move{node: node, location: location, target: target, parentID: definition.Parent, parentName: parentName})
		}
	}
	resolved := resolveDiagnostics && !blocked
	for _, planned := range moves {
		diagnostics = append(diagnostics, Diagnostic{
			Section:  planned.node.Heading,
			Message:  fmt.Sprintf("section moved to schema parent %s", planned.parentName),
			Resolved: resolved,
		})
		removeSectionNode(document, planned.location.parent, planned.node)
		if planned.target == nil {
			document.Roots = append(document.Roots, planned.node)
		} else {
			planned.target.Children = append(planned.target.Children, planned.node)
		}
		classification.locations[planned.node] = nodeLocation{parent: planned.target, parentID: planned.parentID}
	}
	return diagnostics, resolved && len(moves) > 0, blocked
}

func removeSectionNode(document *markdownDocument, parent, node *markdownSection) {
	if parent == nil {
		document.Roots = removeNode(document.Roots, node)
		return
	}
	parent.Children = removeNode(parent.Children, node)
}

func removeNode(nodes []*markdownSection, target *markdownSection) []*markdownSection {
	for index, node := range nodes {
		if node != target {
			continue
		}
		copy(nodes[index:], nodes[index+1:])
		nodes[len(nodes)-1] = nil
		return nodes[:len(nodes)-1]
	}
	return nodes
}
