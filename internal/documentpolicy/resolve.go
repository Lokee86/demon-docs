package documentpolicy

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
)

func IgnoreSection(repoRoot string, cfg config.Config, path, heading string) (string, error) {
	source, parsed, document, err := loadDocument(path, cfg.Frontmatter.AllowedFormats)
	if err != nil {
		return "", err
	}
	documentID, _ := parsed.Values["document_id"].(string)
	if strings.TrimSpace(documentID) == "" {
		return "", fmt.Errorf("document requires document_id metadata before a document-specific schema can be created")
	}
	relative, _ := filepath.Rel(repoRoot, path)
	schemaName, err := selectSchema(filepath.ToSlash(relative), parsed.Values, cfg.Format)
	if err != nil {
		return "", err
	}
	shared, _, err := LoadShared(repoRoot, cfg.Format, schemaName)
	if err != nil {
		return "", err
	}
	local, localPath, exists, err := LoadDocumentSchema(repoRoot, cfg.Format, documentID)
	if err != nil {
		return "", err
	}
	if !exists {
		local = DocumentSchema{Version: 1, DocumentID: documentID, SharedSchema: schemaName}
	}
	if local.DocumentID != "" && local.DocumentID != documentID {
		return "", fmt.Errorf("document-specific schema identifies document %q, not %q", local.DocumentID, documentID)
	}
	if local.SharedSchema != "" && local.SharedSchema != schemaName {
		return "", fmt.Errorf("document-specific schema extends %q, not %q", local.SharedSchema, schemaName)
	}
	effective := EffectiveSchema(shared, local)
	if err := ValidateSchema(effective); err != nil {
		return "", fmt.Errorf("invalid effective document schema: %w", err)
	}
	parentID, nodes, known, err := locateSection(document.Roots, "", effective, heading)
	if err != nil {
		return "", err
	}
	if len(nodes) == 0 {
		return "", fmt.Errorf("section %q not found in %s", heading, path)
	}
	if known.ID != "" {
		override := Section{ID: known.ID, Heading: known.Heading, AllowDuplicates: true}
		local.Sections = insertSection(local.Sections, override)
	} else {
		section := Section{
			ID:              localSectionID(parentID, nodes[0].Heading),
			Heading:         nodes[0].Heading,
			Parent:          parentID,
			After:           lastSiblingID(effective.Sections, parentID),
			AllowDuplicates: len(nodes) > 1,
		}
		local.Sections = insertSection(local.Sections, section)
	}
	local.SharedSchema = schemaName
	local.SharedFingerprint = Fingerprint(shared)
	if err := ValidateSchema(EffectiveSchema(shared, local)); err != nil {
		return "", fmt.Errorf("document-specific schema update would be invalid: %w", err)
	}
	if err := saveSchemaSnapshot(repoRoot, schemaName, shared); err != nil {
		return "", err
	}
	if err := writeDocumentSchema(localPath, local); err != nil {
		return "", err
	}
	_ = source
	return localPath, nil
}

func MergeSections(path, heading string, allowedFormats []string) (int, error) {
	source, _, document, err := loadDocument(path, allowedFormats)
	if err != nil {
		return 0, err
	}
	parent, indexes := findSiblingOccurrences(document.Roots, heading)
	if len(indexes) < 2 {
		return 0, fmt.Errorf("section %q does not have duplicate sibling occurrences", heading)
	}
	first := parent[indexes[0]]
	for _, index := range indexes[1:] {
		mergeNodes(first, parent[index], document.Newline)
	}
	remove := map[int]bool{}
	for _, index := range indexes[1:] {
		remove[index] = true
	}
	filtered := parent[:0]
	for index, node := range parent {
		if !remove[index] {
			filtered = append(filtered, node)
		}
	}
	setSiblingSlice(&document, parent, filtered)
	if err := rewriteDocument(path, source, document.render()); err != nil {
		return 0, err
	}
	return len(indexes), nil
}

func DeleteSection(path, heading string, occurrence int, allowedFormats []string) error {
	if occurrence < 1 {
		return fmt.Errorf("occurrence must be at least 1")
	}
	source, _, document, err := loadDocument(path, allowedFormats)
	if err != nil {
		return err
	}
	parent, indexes := findSiblingOccurrences(document.Roots, heading)
	if occurrence > len(indexes) {
		return fmt.Errorf("section %q occurrence %d not found", heading, occurrence)
	}
	remove := indexes[occurrence-1]
	filtered := append([]*markdownSection(nil), parent[:remove]...)
	filtered = append(filtered, parent[remove+1:]...)
	setSiblingSlice(&document, parent, filtered)
	return rewriteDocument(path, source, document.render())
}
