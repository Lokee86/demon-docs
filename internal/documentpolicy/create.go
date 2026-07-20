package documentpolicy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func Create(repoRoot, docsRoot string, cfg config.Config, schemaName, target string, force bool, now time.Time) (string, error) {
	schema, _, err := LoadShared(repoRoot, cfg.Format, schemaName)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(repoRoot, filepath.FromSlash(target))
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if !repository.Contains(docsRoot, target) {
		return "", fmt.Errorf("new document must be inside docs root: %s", target)
	}
	if err := validateCreateParent(docsRoot, filepath.Dir(target)); err != nil {
		return "", err
	}
	if info, statErr := os.Lstat(target); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("refusing to overwrite symbolic-link target: %s", target)
		}
		if info.IsDir() {
			return "", fmt.Errorf("target is a directory: %s", target)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("target is not a regular file: %s", target)
		}
		if !force {
			return "", fmt.Errorf("target already exists: %s", target)
		}
	} else if !os.IsNotExist(statErr) {
		return "", statErr
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	title := titleFromPath(target, cfg.IndexFile)
	body := renderTemplateBody(schema, title, target, docsRoot, cfg)
	values := cloneMap(schema.Frontmatter.Values)
	values["document_type"] = schema.Name
	if _, ok := values["document_id"]; !ok {
		documentID, err := frontmatter.NewUUIDv7(now.UTC())
		if err != nil {
			return "", fmt.Errorf("generate document ID: %w", err)
		}
		values["document_id"] = documentID
	}
	if _, ok := values["created"]; !ok {
		values["created"] = now.Format("2006-01-02")
	}
	if _, ok := values["summary"]; !ok {
		values["summary"] = schema.Placeholder
	}
	if _, ok := values["author"]; !ok {
		author := cfg.Frontmatter.DefaultAuthor
		if strings.TrimSpace(author) == "" {
			author = "TODO"
		}
		values["author"] = author
	}
	finalValues := values
	if cfg.Frontmatter.Enabled {
		outcome := frontmatter.Evaluate(filepath.ToSlash(target), frontmatter.Document{Values: values, Body: body}, cfg.Frontmatter, true, nil, now)
		if outcomeHasFailure(outcome) {
			return "", fmt.Errorf("schema %q cannot create valid frontmatter with the current frontmatter policy", schemaName)
		}
		finalValues = outcome.Values
	}
	if documentType, ok := finalValues["document_type"].(string); !ok || strings.TrimSpace(documentType) != schema.Name {
		return "", fmt.Errorf("frontmatter policy must preserve document_type = %q during creation", schema.Name)
	}
	if documentID, ok := finalValues["document_id"].(string); !ok || strings.TrimSpace(documentID) == "" {
		return "", fmt.Errorf("frontmatter policy must preserve document_id during creation")
	}
	format := strings.ToLower(strings.TrimSpace(schema.Frontmatter.Format))
	if format == "" {
		format = strings.ToLower(strings.TrimSpace(cfg.Frontmatter.DefaultFormat))
	}
	if !containsFold(cfg.Frontmatter.AllowedFormats, format) {
		return "", fmt.Errorf("schema %q requests frontmatter format %q, which is not allowed", schemaName, format)
	}
	source, err := frontmatter.Render(format, finalValues, body)
	if err != nil {
		return "", err
	}
	changed, err := writeTransactionalFile(target, []byte(source), force)
	if err != nil {
		return "", err
	}
	if !changed && !force {
		return "", fmt.Errorf("target already exists: %s", target)
	}
	return target, nil
}

func renderTemplateBody(schema Schema, title, target, docsRoot string, cfg config.Config) string {
	newline := "\n"
	document := markdownDocument{Newline: newline}
	renderedTitle := strings.ReplaceAll(schema.Document.Title, "{title}", title)
	document.Prefix = "# " + renderedTitle + newline
	if schema.Document.ParentLink {
		if label, link, ok := parentIndex(target, docsRoot, cfg.IndexFile); ok {
			document.Prefix += newline + cfg.ParentLink.Label + ": [" + label + "](" + link + ")" + newline
		}
	}
	document.Prefix += newline
	nodes := map[string]*markdownSection{}
	for _, section := range schema.Sections {
		level := 2
		if section.Parent != "" {
			if parent := nodes[section.Parent]; parent != nil {
				level = parent.Level + 1
			}
		}
		placeholder := section.Placeholder
		if placeholder == "" {
			placeholder = schema.Placeholder
		}
		node := newSection(section.Heading, level, placeholder, newline)
		nodes[section.ID] = node
		if section.Parent == "" || nodes[section.Parent] == nil {
			document.Roots = append(document.Roots, node)
		} else {
			nodes[section.Parent].Children = append(nodes[section.Parent].Children, node)
		}
	}
	return document.render()
}

func parentIndex(target, docsRoot, indexFile string) (string, string, bool) {
	directory := filepath.Dir(target)
	index := filepath.Join(directory, indexFile)
	labelDirectory := directory
	if samePath(target, index) {
		directory = filepath.Dir(directory)
		if !repository.Contains(docsRoot, directory) && directory != docsRoot {
			return "", "", false
		}
		index = filepath.Join(directory, indexFile)
		labelDirectory = directory
	}
	relative, err := filepath.Rel(filepath.Dir(target), index)
	if err != nil {
		return "", "", false
	}
	label := titleWords(filepath.Base(labelDirectory))
	if labelDirectory == docsRoot {
		label = titleWords(filepath.Base(docsRoot))
	}
	return label, filepath.ToSlash(relative), true
}

func titleFromPath(path, indexFile string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if strings.EqualFold(filepath.Base(path), indexFile) {
		base = filepath.Base(filepath.Dir(path))
	}
	return titleWords(base)
}

func titleWords(value string) string {
	value = strings.NewReplacer("-", " ", "_", " ").Replace(value)
	words := strings.Fields(value)
	for i, word := range words {
		runes := []rune(strings.ToLower(word))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func cloneMap(values map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range values {
		result[key] = value
	}
	return result
}

func samePath(left, right string) bool {
	left, _ = filepath.Abs(left)
	right, _ = filepath.Abs(right)
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func validateCreateParent(docsRoot, parent string) error {
	resolvedRoot, err := filepath.EvalSymlinks(docsRoot)
	if err != nil {
		return fmt.Errorf("resolve docs root: %w", err)
	}
	existing := parent
	for {
		if _, err := os.Lstat(existing); err == nil {
			break
		} else if !os.IsNotExist(err) {
			return err
		}
		next := filepath.Dir(existing)
		if next == existing {
			return fmt.Errorf("cannot resolve existing parent for %s", parent)
		}
		existing = next
	}
	resolvedParent, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return fmt.Errorf("resolve document parent: %w", err)
	}
	if !repository.Contains(resolvedRoot, resolvedParent) {
		return fmt.Errorf("document parent resolves outside docs root: %s", parent)
	}
	return nil
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}

func outcomeHasFailure(outcome frontmatter.Outcome) bool {
	for _, diagnostic := range outcome.Diagnostics {
		if !diagnostic.Warning && !diagnostic.Resolved {
			return true
		}
	}
	return false
}
