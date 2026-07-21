package links

import (
	"os"
	"regexp"
	"strings"

	"github.com/Lokee86/demon-docs/internal/frontmatter"
)

var documentIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func markdownDocumentID(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return markdownDocumentIDBytes(data)
}

func markdownDocumentIDBytes(data []byte) string {
	document, err := frontmatter.Parse(string(data), []string{frontmatter.FormatYAML, frontmatter.FormatTOML})
	if err != nil || !document.HasBlock {
		return ""
	}
	value, ok := document.Values["document_id"].(string)
	if !ok {
		return ""
	}
	value = strings.ToLower(strings.TrimSpace(value))
	if !documentIDPattern.MatchString(value) {
		return ""
	}
	return value
}
