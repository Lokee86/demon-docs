package documentpolicy

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Lokee86/demon-docs/internal/filetxn"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
)

func loadDocument(path string, allowedFormats []string) (string, frontmatter.Document, markdownDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", frontmatter.Document{}, markdownDocument{}, err
	}
	source := string(data)
	parsed, err := frontmatter.Parse(source, allowedFormats)
	if err != nil {
		return "", frontmatter.Document{}, markdownDocument{}, err
	}
	bodyStart := frontmatter.LeadingBlockEnd(source)
	return source, parsed, parseMarkdown(source[bodyStart:]), nil
}

func rewriteDocument(path, original, body string) error {
	bodyStart := frontmatter.LeadingBlockEnd(original)
	next := original[:bodyStart] + body
	if next == original {
		return nil
	}
	_, err := filetxn.Apply([]filetxn.Rewrite{filetxn.New(path, []byte(original), []byte(next))})
	return err
}

func locateSection(children []*markdownSection, parentID string, schema Schema, heading string) (string, []*markdownSection, Section, error) {
	definitions := childrenForParent(schema.Sections, parentID)
	var matches []*markdownSection
	known := Section{}
	for _, node := range children {
		definition := matchingDefinition(node.Heading, definitions)
		if strings.EqualFold(strings.TrimSpace(node.Heading), strings.TrimSpace(heading)) || definition.ID != "" && (strings.EqualFold(definition.Heading, heading) || matchesAlias(heading, definition.Aliases)) {
			matches = append(matches, node)
			if definition.ID != "" {
				known = definition
			}
		}
		if definition.ID != "" {
			childParent, childMatches, childKnown, err := locateSection(node.Children, definition.ID, schema, heading)
			if err != nil {
				return "", nil, Section{}, err
			}
			if len(childMatches) > 0 {
				if len(matches) > 0 {
					return "", nil, Section{}, fmt.Errorf("section %q is ambiguous across multiple parents", heading)
				}
				return childParent, childMatches, childKnown, nil
			}
		}
	}
	return parentID, matches, known, nil
}

func matchingDefinition(heading string, definitions []Section) Section {
	for _, definition := range definitions {
		if strings.EqualFold(heading, definition.Heading) || matchesAlias(heading, definition.Aliases) {
			return definition
		}
	}
	return Section{}
}

func findSiblingOccurrences(children []*markdownSection, heading string) ([]*markdownSection, []int) {
	var indexes []int
	for index, child := range children {
		if strings.EqualFold(strings.TrimSpace(child.Heading), strings.TrimSpace(heading)) {
			indexes = append(indexes, index)
		}
	}
	if len(indexes) > 0 {
		return children, indexes
	}
	for _, child := range children {
		if parent, found := findSiblingOccurrences(child.Children, heading); len(found) > 0 {
			return parent, found
		}
	}
	return nil, nil
}

func setSiblingSlice(document *markdownDocument, original, replacement []*markdownSection) {
	if sameSlice(document.Roots, original) {
		document.Roots = replacement
		return
	}
	replaceChildSlice(document.Roots, original, replacement)
}

func replaceChildSlice(nodes []*markdownSection, original, replacement []*markdownSection) bool {
	for _, node := range nodes {
		if sameSlice(node.Children, original) {
			node.Children = replacement
			return true
		}
		if replaceChildSlice(node.Children, original, replacement) {
			return true
		}
	}
	return false
}

func sameSlice(left, right []*markdownSection) bool {
	if len(left) != len(right) {
		return false
	}
	if len(left) == 0 {
		return left == nil && right == nil
	}
	return &left[0] == &right[0]
}

func lastSiblingID(sections []Section, parent string) string {
	last := ""
	for _, section := range sections {
		if section.Parent == parent {
			last = section.ID
		}
	}
	return last
}

func localSectionID(parent, heading string) string {
	slug := strings.ToLower(heading)
	slug = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, slug)
	slug = strings.Trim(strings.Join(strings.FieldsFunc(slug, func(r rune) bool { return r == '-' }), "-"), "-")
	if slug == "" {
		slug = "section"
	}
	sum := sha256.Sum256([]byte(parent + "\x00" + heading))
	return "local-" + slug + "-" + hex.EncodeToString(sum[:4])
}

func writeDocumentSchema(path string, schema DocumentSchema) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(schema); err != nil {
		return err
	}
	next := buffer.Bytes()
	current, err := os.ReadFile(path)
	if err == nil {
		_, err = filetxn.Apply([]filetxn.Rewrite{filetxn.New(path, current, next)})
		return err
	}
	if !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(next); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}
