package documentpolicy

import (
	"fmt"
	"strings"
)

type enforcementResult struct {
	Document    markdownDocument
	Diagnostics []Diagnostic
	Changed     bool
	Blocked     bool
}

type sectionMatch struct {
	definition Section
	node       *markdownSection
	renamed    bool
	alias      bool
}

func enforceDocument(document markdownDocument, current, previous Schema, hasPrevious, repair bool) enforcementResult {
	working := cloneDocument(document)
	classification := classifyDocument(working.Roots, current, previous, hasPrevious)
	conflicts := manualConflictDiagnostics(working.Roots, classification, current)
	canApply := repair && len(conflicts) == 0
	relocationDiagnostics, relocationChanged, relocationBlocked := relocateKnownSections(&working, classification, current, canApply)
	blocked := len(conflicts) > 0 || relocationBlocked
	if relocationBlocked {
		canApply = false
	}
	outcome := applyChildren(working.Roots, "", 2, current, previous, hasPrevious, working.Newline, canApply)
	working.Roots = outcome.children
	diagnostics := append(conflicts, relocationDiagnostics...)
	diagnostics = append(diagnostics, outcome.diagnostics...)
	if blocked {
		return enforcementResult{Document: document, Diagnostics: diagnostics, Blocked: true}
	}
	return enforcementResult{Document: working, Diagnostics: diagnostics, Changed: relocationChanged || outcome.changed}
}

type childOutcome struct {
	children    []*markdownSection
	diagnostics []Diagnostic
	changed     bool
}

func applyChildren(children []*markdownSection, parentID string, expectedLevel int, current, previous Schema, hasPrevious bool, newline string, repair bool) childOutcome {
	definitions := childrenForParent(current.Sections, parentID)
	previousByID := map[string]Section{}
	if hasPrevious {
		for _, section := range previous.Sections {
			previousByID[section.ID] = section
		}
	}
	matches := groupMatches(children, definitions, previousByID)
	out := childOutcome{}
	unknown := make([]*markdownSection, 0)
	for _, child := range children {
		if _, ok := matches.byNode[child]; !ok {
			unknown = append(unknown, child)
		}
	}

	ordered := make([]*markdownSection, 0, len(children)+len(definitions))
	for _, definition := range definitions {
		nodes := append([]*markdownSection(nil), matches.byID[definition.ID]...)
		if len(nodes) == 0 {
			if definition.Optional {
				continue
			}
			resolved := repair
			out.diagnostics = append(out.diagnostics, Diagnostic{Section: definition.Heading, Message: "required section is missing; fix creates it with configured placeholder text", Resolved: resolved})
			if repair {
				placeholder := definition.Placeholder
				if placeholder == "" {
					placeholder = current.Placeholder
				}
				node := newSection(definition.Heading, expectedLevel, placeholder, newline)
				nodes = []*markdownSection{node}
				out.changed = true
			}
		}
		if len(nodes) > 1 && !definition.AllowDuplicates {
			switch strings.ToLower(strings.TrimSpace(current.DuplicateSections)) {
			case "merge":
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: definition.Heading, Message: fmt.Sprintf("merged %d duplicate sections", len(nodes)), Resolved: repair})
				if repair {
					merged := nodes[0]
					for _, duplicate := range nodes[1:] {
						mergeNodes(merged, duplicate, newline)
					}
					nodes = []*markdownSection{merged}
					out.changed = true
				}
			case "delete-first":
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: definition.Heading, Message: "duplicate policy deletes all but the last occurrence", Resolved: repair})
				if repair {
					nodes = nodes[len(nodes)-1:]
					out.changed = true
				}
			case "delete-last":
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: definition.Heading, Message: "duplicate policy deletes all but the first occurrence", Resolved: repair})
				if repair {
					nodes = nodes[:1]
					out.changed = true
				}
			case "keep", "allow":
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: definition.Heading, Message: "duplicate section accepted by shared schema policy", Warning: true})
			}
		}
		for _, node := range nodes {
			match := matches.details[node]
			if match.renamed {
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: node.Heading, Message: fmt.Sprintf("schema renamed section to %q", definition.Heading), Resolved: repair})
				if repair {
					replaceHeading(node, definition.Heading, newline)
					out.changed = true
				}
			} else if match.alias && definition.CanonicalizeAliases {
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: node.Heading, Message: fmt.Sprintf("alias canonicalizes to %q", definition.Heading), Resolved: repair})
				if repair {
					replaceHeading(node, definition.Heading, newline)
					out.changed = true
				}
			}
			if node.Level != expectedLevel {
				out.diagnostics = append(out.diagnostics, Diagnostic{Section: definition.Heading, Message: fmt.Sprintf("heading level is %d; schema requires %d", node.Level, expectedLevel), Resolved: repair})
				if repair {
					setHeadingLevel(node, expectedLevel, newline)
					out.changed = true
				}
			}
			childResult := applyChildren(node.Children, definition.ID, expectedLevel+1, current, previous, hasPrevious, newline, repair)
			node.Children = childResult.children
			out.diagnostics = append(out.diagnostics, childResult.diagnostics...)
			out.changed = out.changed || childResult.changed
			ordered = append(ordered, node)
		}
	}

	switch strings.ToLower(strings.TrimSpace(current.UnknownSections)) {
	case "delete":
		for _, node := range unknown {
			out.diagnostics = append(out.diagnostics, Diagnostic{Section: node.Heading, Message: "unknown section removed by configured policy", Resolved: repair})
		}
		if repair && len(unknown) > 0 {
			out.changed = true
		}
	default:
		ordered = append(ordered, unknown...)
	}
	if !sameNodeOrder(children, ordered) {
		out.diagnostics = append(out.diagnostics, Diagnostic{Message: parentOrderMessage(parentID), Resolved: repair})
		if repair {
			out.changed = true
		}
	}
	if repair {
		out.children = ordered
	} else {
		out.children = children
	}
	return out
}

type matchGroups struct {
	byID    map[string][]*markdownSection
	byNode  map[*markdownSection]Section
	details map[*markdownSection]sectionMatch
}

func groupMatches(children []*markdownSection, definitions []Section, previous map[string]Section) matchGroups {
	groups := matchGroups{byID: map[string][]*markdownSection{}, byNode: map[*markdownSection]Section{}, details: map[*markdownSection]sectionMatch{}}
	for _, child := range children {
		for _, definition := range definitions {
			alias := matchesAlias(child.Heading, definition.Aliases)
			canonical := strings.EqualFold(strings.TrimSpace(child.Heading), strings.TrimSpace(definition.Heading))
			renamed := false
			if !canonical && !alias && previous != nil {
				prior, ok := previous[definition.ID]
				renamed = ok && (strings.EqualFold(child.Heading, prior.Heading) || matchesAlias(child.Heading, prior.Aliases)) && !strings.EqualFold(prior.Heading, definition.Heading)
			}
			if !canonical && !alias && !renamed {
				continue
			}
			groups.byID[definition.ID] = append(groups.byID[definition.ID], child)
			groups.byNode[child] = definition
			groups.details[child] = sectionMatch{definition: definition, node: child, renamed: renamed, alias: alias}
			break
		}
	}
	return groups
}

func childrenForParent(sections []Section, parent string) []Section {
	var result []Section
	for _, section := range sections {
		if section.Parent == parent {
			result = append(result, section)
		}
	}
	return result
}

func matchesAlias(heading string, aliases []string) bool {
	for _, alias := range aliases {
		if strings.EqualFold(strings.TrimSpace(heading), strings.TrimSpace(alias)) {
			return true
		}
	}
	return false
}

func newSection(heading string, level int, placeholder, newline string) *markdownSection {
	lead := newline
	if placeholder != "" {
		lead += placeholder + newline
	}
	lead += newline
	return &markdownSection{Heading: heading, Level: level, HeadingText: strings.Repeat("#", level) + " " + heading + newline, Lead: lead}
}

func setHeadingLevel(section *markdownSection, level int, newline string) {
	section.Level = level
	section.HeadingText = strings.Repeat("#", level) + " " + section.Heading + newline
}

func cloneDocument(document markdownDocument) markdownDocument {
	copy := document
	copy.Roots = cloneSections(document.Roots)
	return copy
}

func cloneSections(sections []*markdownSection) []*markdownSection {
	result := make([]*markdownSection, len(sections))
	for i, section := range sections {
		copy := *section
		copy.Children = cloneSections(section.Children)
		result[i] = &copy
	}
	return result
}

func sameNodeOrder(left, right []*markdownSection) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func parentOrderMessage(parentID string) string {
	if parentID == "" {
		return "top-level sections are not in schema order"
	}
	return fmt.Sprintf("sections beneath %q are not in schema order", parentID)
}
