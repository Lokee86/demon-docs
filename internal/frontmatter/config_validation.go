package frontmatter

import (
	"fmt"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
)

func ValidateConfig(schema config.Frontmatter) error {
	if !schema.Enabled {
		return nil
	}
	defaultFormat := normalizedFormat(schema.DefaultFormat)
	if defaultFormat == "" {
		return fmt.Errorf("frontmatter.default_format must be yaml or toml")
	}
	if len(schema.AllowedFormats) == 0 {
		return fmt.Errorf("frontmatter.allowed_formats must not be empty")
	}
	defaultAllowed := false
	for _, candidate := range schema.AllowedFormats {
		format := normalizedFormat(candidate)
		if format == "" {
			return fmt.Errorf("frontmatter.allowed_formats contains unsupported format %q", candidate)
		}
		defaultAllowed = defaultAllowed || format == defaultFormat
	}
	if !defaultAllowed {
		return fmt.Errorf("frontmatter.allowed_formats must contain the default format")
	}
	switch strings.ToLower(strings.TrimSpace(schema.UnknownFields)) {
	case "remove", "warn", "ignore":
	default:
		return fmt.Errorf("frontmatter.unknown_fields must be remove, warn, or ignore")
	}
	for name, field := range schema.Fields {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("frontmatter field name cannot be empty")
		}
		kind, err := validateConfiguredType(field.Type)
		if err != nil {
			return fmt.Errorf("frontmatter field %s: %w", name, err)
		}
		sources := 0
		if field.Default != nil {
			sources++
			if _, err := validateValue(kind, field.Default); err != nil {
				return fmt.Errorf("frontmatter field %s default: %w", name, err)
			}
		}
		if field.DefaultFrom != "" {
			sources++
			if field.DefaultFrom != "frontmatter.default_author" {
				return fmt.Errorf("frontmatter field %s has unsupported default_from %q", name, field.DefaultFrom)
			}
			if kind != "string" {
				return fmt.Errorf("frontmatter field %s default_from requires type string", name)
			}
		}
		if field.Generated {
			sources++
			if kind != "uuid" && kind != "date" {
				return fmt.Errorf("frontmatter field %s can only generate uuid or date values", name)
			}
		}
		if sources > 1 {
			return fmt.Errorf("frontmatter field %s has multiple value sources", name)
		}
	}
	for _, rule := range schema.Rules {
		condition, ok := schema.Fields[rule.WhenField]
		if !ok {
			return fmt.Errorf("frontmatter rule references unknown field %s", rule.WhenField)
		}
		if _, ok := schema.Fields[rule.Require]; !ok {
			return fmt.Errorf("frontmatter rule requires unknown field %s", rule.Require)
		}
		if _, err := validateValue(normalizedType(condition.Type), rule.Equals); err != nil {
			return fmt.Errorf("frontmatter rule for %s has invalid equals value: %w", rule.WhenField, err)
		}
	}
	return nil
}

func normalizedFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "yaml", "yml":
		return FormatYAML
	case "toml":
		return FormatTOML
	default:
		return ""
	}
}

func normalizedType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "bool":
		return "boolean"
	case "list", "strings", "string-list":
		return "string_list"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func validateConfiguredType(kind string) (string, error) {
	kind = normalizedType(kind)
	switch kind {
	case "string", "boolean", "integer", "number", "string_list", "date", "uuid":
		return kind, nil
	default:
		return kind, fmt.Errorf("unsupported type %q", kind)
	}
}
