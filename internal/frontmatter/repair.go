package frontmatter

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

func hasConfiguredSource(field config.FrontmatterField, schema config.Frontmatter) bool {
	if field.Default != nil || field.Generated {
		return true
	}
	return field.DefaultFrom == "frontmatter.default_author" && strings.TrimSpace(schema.DefaultAuthor) != ""
}

func replacementValue(field config.FrontmatterField, schema config.Frontmatter, prior any, hasPrior bool, now time.Time) (any, bool, error) {
	kind := normalizedType(field.Type)
	if field.Immutable && hasPrior {
		return prior, true, nil
	}
	if field.Default != nil {
		value, err := validateValue(kind, field.Default)
		return value, err == nil, err
	}
	if field.DefaultFrom == "frontmatter.default_author" && strings.TrimSpace(schema.DefaultAuthor) != "" {
		return schema.DefaultAuthor, true, nil
	}
	if !field.Generated {
		return nil, false, nil
	}
	switch kind {
	case "uuid":
		value, err := newUUIDv7(now.UTC())
		if err != nil {
			return nil, false, fmt.Errorf("generate UUID: %w", err)
		}
		return value, true, nil
	case "date":
		return now.Format("2006-01-02"), true, nil
	default:
		return nil, false, fmt.Errorf("unsupported generated field type %q", field.Type)
	}
}

func (out *Outcome) add(path, field, message string, warning, resolved bool) {
	out.Diagnostics = append(out.Diagnostics, Diagnostic{Path: path, Field: field, Message: message, Warning: warning, Resolved: resolved})
}

func cloneValues(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func emptyValue(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func equalValue(left, right any) bool {
	return reflect.DeepEqual(normalize(left), normalize(right))
}
