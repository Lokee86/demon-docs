package documentpolicy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Lokee86/demon-docs/internal/config"
)

type DocumentTemplate struct {
	Title      string `toml:"title"`
	ParentLink bool   `toml:"parent_link"`
}

type FrontmatterTemplate struct {
	Format string         `toml:"format"`
	Values map[string]any `toml:"values"`
}

type Section struct {
	ID                  string   `toml:"id"`
	Heading             string   `toml:"heading"`
	Parent              string   `toml:"parent"`
	After               string   `toml:"after"`
	Placeholder         string   `toml:"placeholder"`
	Aliases             []string `toml:"aliases"`
	Optional            bool     `toml:"optional"`
	AllowDuplicates     bool     `toml:"allow_duplicates"`
	CanonicalizeAliases bool     `toml:"canonicalize_aliases"`
}

type Schema struct {
	Version           int                 `toml:"version"`
	Name              string              `toml:"name"`
	Description       string              `toml:"description"`
	Placeholder       string              `toml:"placeholder"`
	UnknownSections   string              `toml:"unknown_sections"`
	DuplicateSections string              `toml:"duplicate_sections"`
	Document          DocumentTemplate    `toml:"document"`
	Frontmatter       FrontmatterTemplate `toml:"frontmatter"`
	Sections          []Section           `toml:"sections"`
}

type DocumentSchema struct {
	Version           int       `toml:"version"`
	DocumentID        string    `toml:"document_id"`
	SharedSchema      string    `toml:"shared_schema"`
	SharedFingerprint string    `toml:"shared_fingerprint"`
	Sections          []Section `toml:"sections"`
}

func LoadShared(repoRoot string, cfg config.Format, name string) (Schema, string, error) {
	name = strings.TrimSpace(name)
	if !safeSchemaName(name) {
		return Schema{}, "", fmt.Errorf("unsafe document schema name %q", name)
	}
	path := filepath.Join(resolveDir(repoRoot, cfg.SchemaDir), name+".toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return Schema{}, path, err
		}
		if _, dirErr := os.Stat(filepath.Dir(path)); dirErr == nil {
			return Schema{}, path, fmt.Errorf("document schema %q not found at %s", name, path)
		} else if !os.IsNotExist(dirErr) {
			return Schema{}, path, dirErr
		}
		text, ok := BuiltinSchemas()[name]
		if !ok {
			return Schema{}, path, fmt.Errorf("document schema %q not found at %s", name, path)
		}
		data = []byte(text)
	}
	var schema Schema
	if _, err := toml.Decode(string(data), &schema); err != nil {
		return Schema{}, path, fmt.Errorf("parse document schema %s: %w", path, err)
	}
	if schema.Name == "" {
		schema.Name = name
	} else if schema.Name != name {
		return Schema{}, path, fmt.Errorf("document schema %s declares name %q instead of %q", path, schema.Name, name)
	}
	if schema.Placeholder == "" {
		schema.Placeholder = "TODO"
	}
	if schema.UnknownSections == "" {
		schema.UnknownSections = "manual"
	}
	if schema.DuplicateSections == "" {
		schema.DuplicateSections = "manual"
	}
	if schema.Document.Title == "" {
		schema.Document.Title = "{title}"
	}
	if schema.Frontmatter.Values == nil {
		schema.Frontmatter.Values = map[string]any{}
	}
	if err := ValidateSchema(schema); err != nil {
		return Schema{}, path, fmt.Errorf("document schema %s: %w", path, err)
	}
	return schema, path, nil
}

func LoadDocumentSchema(repoRoot string, cfg config.Format, documentID string) (DocumentSchema, string, bool, error) {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" || documentID == "." || documentID == ".." || filepath.Base(documentID) != documentID || strings.ContainsAny(documentID, `/\\`) {
		return DocumentSchema{}, "", false, fmt.Errorf("unsafe document_id for document-specific schema: %q", documentID)
	}
	path := filepath.Join(resolveDir(repoRoot, cfg.DocumentSchemaDir), documentID+".toml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DocumentSchema{}, path, false, nil
	}
	if err != nil {
		return DocumentSchema{}, path, false, err
	}
	var schema DocumentSchema
	if _, err := toml.Decode(string(data), &schema); err != nil {
		return DocumentSchema{}, path, true, fmt.Errorf("parse document-specific schema %s: %w", path, err)
	}
	if schema.Version < 0 || schema.Version > 1 {
		return DocumentSchema{}, path, true, fmt.Errorf("document-specific schema %s has unsupported version %d", path, schema.Version)
	}
	if schema.SharedFingerprint != "" && !validSchemaFingerprint(schema.SharedFingerprint) {
		return DocumentSchema{}, path, true, fmt.Errorf("document-specific schema %s has invalid shared_fingerprint", path)
	}
	return schema, path, true, nil
}

func EffectiveSchema(shared Schema, local DocumentSchema) Schema {
	result := shared
	for _, section := range local.Sections {
		result.Sections = insertSection(result.Sections, section)
	}
	return result
}

func insertSection(sections []Section, section Section) []Section {
	for i := range sections {
		if sections[i].ID != section.ID {
			continue
		}
		merged := sections[i]
		if section.Heading != "" {
			merged.Heading = section.Heading
		}
		if section.Parent != "" {
			merged.Parent = section.Parent
		}
		if section.Placeholder != "" {
			merged.Placeholder = section.Placeholder
		}
		merged.Aliases = appendUnique(merged.Aliases, section.Aliases...)
		merged.AllowDuplicates = merged.AllowDuplicates || section.AllowDuplicates
		merged.CanonicalizeAliases = merged.CanonicalizeAliases || section.CanonicalizeAliases
		sections[i] = merged
		return sections
	}
	if section.After != "" {
		for i := range sections {
			if sections[i].ID == section.After {
				sections = append(sections, Section{})
				copy(sections[i+2:], sections[i+1:])
				sections[i+1] = section
				return sections
			}
		}
	}
	return append(sections, section)
}

func safeSchemaName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name && !strings.ContainsAny(name, `/\\`)
}

func resolveDir(repoRoot, configured string) string {
	if filepath.IsAbs(configured) {
		return filepath.Clean(configured)
	}
	return filepath.Join(repoRoot, filepath.FromSlash(configured))
}

func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[strings.ToLower(strings.TrimSpace(value))] = true
	}
	for _, value := range additions {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		values = append(values, value)
	}
	return values
}
