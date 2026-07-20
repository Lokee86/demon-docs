package frontmatter

import "github.com/Lokee86/demon-docs/internal/config"

func schema() config.Frontmatter {
	return config.Frontmatter{
		Enabled:        true,
		DefaultFormat:  FormatYAML,
		AllowedFormats: []string{FormatYAML, FormatTOML},
		DefaultAuthor:  "Demon Docs",
		UnknownFields:  "remove",
		Fields: map[string]config.FrontmatterField{
			"author":        {Type: "string", Required: true, DefaultFrom: "frontmatter.default_author"},
			"created":       {Type: "date", Required: true, Immutable: true, Generated: true},
			"document_id":   {Type: "uuid", Required: true, Immutable: true, Generated: true},
			"document_type": {Type: "string", Required: true, Default: "general"},
			"summary":       {Type: "string", Required: true},
		},
	}
}

func completeValues() map[string]any {
	return map[string]any{
		"author":        "Human",
		"created":       "2026-07-20",
		"document_id":   "11111111-2222-4333-8444-555555555555",
		"document_type": "guide",
		"summary":       "Existing",
	}
}

func hasUnresolved(diagnostics []Diagnostic, field string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Field == field && !diagnostic.Warning && !diagnostic.Resolved {
			return true
		}
	}
	return false
}
