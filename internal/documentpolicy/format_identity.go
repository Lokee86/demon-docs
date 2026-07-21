package documentpolicy

import (
	"strings"

	"github.com/Lokee86/demon-docs/internal/validationcache"
)

const formatIdentityVersion = "format-identity-v1"

type formatIdentityInput struct {
	Version       string
	EngineVersion string
	SchemaName    string
	DocumentID    string
	DocumentType  string
	Headings      []formatHeadingIdentity
}

type formatHeadingIdentity struct {
	Text     string
	Level    int
	Children []formatHeadingIdentity
}

// formatIdentity fingerprints only the document metadata and heading tree that
// format enforcement evaluates. Prose, links, code blocks, and section body
// content do not affect the identity.
func formatIdentity(schemaName, documentID, documentType string, document markdownDocument) string {
	return validationcache.Hash(formatIdentityInput{
		Version:       formatIdentityVersion,
		EngineVersion: validationcache.EngineVersion,
		SchemaName:    strings.TrimSpace(schemaName),
		DocumentID:    strings.TrimSpace(documentID),
		DocumentType:  strings.TrimSpace(documentType),
		Headings:      formatHeadingIdentities(document.Roots),
	})
}

func formatHeadingIdentities(sections []*markdownSection) []formatHeadingIdentity {
	if len(sections) == 0 {
		return nil
	}
	result := make([]formatHeadingIdentity, len(sections))
	for index, section := range sections {
		result[index] = formatHeadingIdentity{
			Text:     section.Heading,
			Level:    section.Level,
			Children: formatHeadingIdentities(section.Children),
		}
	}
	return result
}
