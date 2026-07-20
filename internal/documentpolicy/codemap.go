package documentpolicy

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
)

// CodemapSchemaProvider adapts document-format schemas to explicit codemap
// execution. A missing codemap section is created only when the selected
// effective schema contains a required codemap section.
type CodemapSchemaProvider struct {
	RepositoryRoot string
	Config         config.Config
	Headings       []string
}

func (provider CodemapSchemaProvider) CodemapSection(documentPath, source string) (codemap.SectionPlacement, bool, error) {
	if !provider.Config.Format.Enabled {
		return codemap.SectionPlacement{}, false, nil
	}
	parsed, err := frontmatter.Parse(source, provider.Config.Frontmatter.AllowedFormats)
	if err != nil {
		return codemap.SectionPlacement{}, false, err
	}
	relative := filepath.ToSlash(filepath.Clean(documentPath))
	schemaName, err := selectSchema(relative, parsed.Values, provider.Config.Format)
	if err != nil {
		return codemap.SectionPlacement{}, false, err
	}
	if schemaName == "" {
		return codemap.SectionPlacement{}, false, nil
	}
	schema, _, err := LoadShared(provider.RepositoryRoot, provider.Config.Format, schemaName)
	if err != nil {
		return codemap.SectionPlacement{}, false, err
	}
	if documentID, _ := parsed.Values["document_id"].(string); strings.TrimSpace(documentID) != "" {
		local, _, exists, err := LoadDocumentSchema(provider.RepositoryRoot, provider.Config.Format, documentID)
		if err != nil {
			return codemap.SectionPlacement{}, false, err
		}
		if exists {
			if local.DocumentID != "" && local.DocumentID != documentID {
				return codemap.SectionPlacement{}, false, fmt.Errorf("document-specific schema identifies document %q instead of %q", local.DocumentID, documentID)
			}
			if local.SharedSchema != "" && local.SharedSchema != schemaName {
				return codemap.SectionPlacement{}, false, fmt.Errorf("document-specific schema extends %q but metadata selects %q", local.SharedSchema, schemaName)
			}
			schema = EffectiveSchema(schema, local)
			if err := ValidateSchema(schema); err != nil {
				return codemap.SectionPlacement{}, false, fmt.Errorf("invalid effective document schema: %w", err)
			}
		}
	}

	sectionIndex := codemapSectionIndex(schema.Sections, provider.Headings)
	if sectionIndex < 0 || schema.Sections[sectionIndex].Optional {
		return codemap.SectionPlacement{}, false, nil
	}
	section := schema.Sections[sectionIndex]
	level, err := schemaSectionLevel(schema.Sections, section)
	if err != nil {
		return codemap.SectionPlacement{}, false, err
	}
	bodyStart := frontmatter.LeadingBlockEnd(source)
	offset, err := schemaInsertionOffset(source[bodyStart:], schema.Sections, sectionIndex)
	if err != nil {
		return codemap.SectionPlacement{}, false, err
	}
	return codemap.SectionPlacement{Heading: section.Heading, Level: level, Offset: bodyStart + offset}, true, nil
}

func codemapSectionIndex(sections []Section, headings []string) int {
	accepted := map[string]bool{"code-map": true, "codemap": true}
	for _, heading := range headings {
		accepted[normalizedHeading(heading)] = true
	}
	for index, section := range sections {
		if accepted[normalizedHeading(section.ID)] || accepted[normalizedHeading(section.Heading)] {
			return index
		}
		for _, alias := range section.Aliases {
			if accepted[normalizedHeading(alias)] {
				return index
			}
		}
	}
	return -1
}

func normalizedHeading(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func schemaSectionLevel(sections []Section, section Section) (int, error) {
	level := 2
	seen := map[string]bool{section.ID: true}
	parent := section.Parent
	for parent != "" {
		if seen[parent] {
			return 0, fmt.Errorf("document schema contains a parent cycle at section %q", section.ID)
		}
		seen[parent] = true
		found := false
		for _, candidate := range sections {
			if candidate.ID == parent {
				parent = candidate.Parent
				level++
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("document schema section %q references missing parent %q", section.ID, parent)
		}
	}
	if level > 6 {
		return 0, fmt.Errorf("document schema section %q would require heading level %d", section.ID, level)
	}
	return level, nil
}

func schemaInsertionOffset(body string, sections []Section, targetIndex int) (int, error) {
	records := scanHeadings(body)
	target := sections[targetIndex]
	siblings := make([]int, 0)
	for index, section := range sections {
		if section.Parent == target.Parent {
			siblings = append(siblings, index)
		}
	}
	position := -1
	for index, sibling := range siblings {
		if sibling == targetIndex {
			position = index
			break
		}
	}
	if position < 0 {
		return 0, fmt.Errorf("document schema section %q is missing from its sibling order", target.ID)
	}
	for _, sibling := range siblings[position+1:] {
		if record, found, err := uniqueSectionRecord(records, sections[sibling]); err != nil {
			return 0, err
		} else if found {
			return record.start, nil
		}
	}
	for index := position - 1; index >= 0; index-- {
		if record, found, err := uniqueSectionRecord(records, sections[siblings[index]]); err != nil {
			return 0, err
		} else if found {
			return subtreeEnd(records, record, len(body)), nil
		}
	}
	if target.Parent != "" {
		for _, section := range sections {
			if section.ID != target.Parent {
				continue
			}
			record, found, err := uniqueSectionRecord(records, section)
			if err != nil {
				return 0, err
			}
			if !found {
				return 0, fmt.Errorf("cannot place schema section %q because parent %q is missing", target.ID, target.Parent)
			}
			return subtreeEnd(records, record, len(body)), nil
		}
	}
	return len(body), nil
}

func uniqueSectionRecord(records []headingRecord, section Section) (headingRecord, bool, error) {
	accepted := map[string]bool{strings.ToLower(strings.TrimSpace(section.Heading)): true}
	for _, alias := range section.Aliases {
		accepted[strings.ToLower(strings.TrimSpace(alias))] = true
	}
	var match headingRecord
	found := false
	for _, record := range records {
		if !accepted[strings.ToLower(strings.TrimSpace(record.heading))] {
			continue
		}
		if found {
			return headingRecord{}, false, fmt.Errorf("document contains duplicate schema section %q", section.Heading)
		}
		match, found = record, true
	}
	return match, found, nil
}

func subtreeEnd(records []headingRecord, parent headingRecord, bodyLength int) int {
	for _, record := range records {
		if record.start > parent.start && record.level <= parent.level {
			return record.start
		}
	}
	return bodyLength
}
