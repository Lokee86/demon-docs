package documentpolicy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
	"github.com/Lokee86/demon-docs/internal/validationcache"
	"github.com/Lokee86/demon-docs/internal/validationworkers"
)

type documentSource struct {
	path           string
	relative       string
	data           []byte
	text           string
	contentHash    string
	formatIdentity string
	candidate      validationcache.Entry
	cacheHit       bool
	parsed         frontmatter.Document
	parseErr       error
	schemaName     string
	schemaErr      error
	bodyStart      int
	document       markdownDocument
}

type documentEvaluation struct {
	path           string
	relative       string
	data           []byte
	text           string
	contentHash    string
	formatIdentity string
	schemaName     string
	documentID     string
	documentType   string
	bodyStart      int
	document       markdownDocument
	current        Schema
	previous       Schema
	hasPrevious    bool
	result         enforcementResult
}

func loadDocumentSources(repoRoot string, files []string, cfg config.Config, cache *validationcache.Store, policyHash string, schemaHasher *validationcache.SchemaHasher) ([]documentSource, error) {
	sources := make([]documentSource, len(files))
	errors := validationworkers.Run(len(files), func(index int) error {
		path := files[index]
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read document format source %s: %w", path, err)
		}
		source := documentSource{
			path:        path,
			relative:    filepath.ToSlash(relative),
			data:        data,
			text:        string(data),
			contentHash: validationcache.ContentHash(data),
		}
		source.parsed, source.parseErr = frontmatter.Parse(source.text, cfg.Frontmatter.AllowedFormats)
		if source.parseErr != nil {
			sources[index] = source
			return nil
		}
		source.schemaName, source.schemaErr = selectSchema(source.relative, source.parsed.Values, cfg.Format)
		if source.schemaErr != nil || strings.TrimSpace(source.schemaName) == "" {
			sources[index] = source
			return nil
		}
		source.bodyStart = frontmatter.LeadingBlockEnd(source.text)
		source.document = parseMarkdown(source.text[source.bodyStart:])
		documentID, _ := source.parsed.Values["document_id"].(string)
		documentType, _ := source.parsed.Values["document_type"].(string)
		source.formatIdentity = formatIdentity(source.schemaName, documentID, documentType, source.document)
		if candidate, ok := cache.CandidateFormat(source.relative, source.formatIdentity, policyHash); ok {
			schemaHash := schemaHasher.Effective(source.schemaName, documentID)
			if entry, valid := cache.LookupFormat(source.relative, source.formatIdentity, policyHash, schemaHash); valid {
				source.cacheHit = true
				source.candidate = entry
				if entry.ContentSHA256 != source.contentHash {
					entry.ContentSHA256 = source.contentHash
					entry.FrontmatterClean = false
					cache.Merge(entry)
					source.candidate = entry
				}
			} else {
				source.candidate = candidate
			}
		}
		sources[index] = source
		return nil
	})
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}
	return sources, nil
}

func runDocumentEvaluations(evaluations []documentEvaluation, repair bool) {
	validationworkers.Run(len(evaluations), func(index int) error {
		evaluation := &evaluations[index]
		evaluation.result = enforceDocument(evaluation.document, evaluation.current, evaluation.previous, evaluation.hasPrevious, repair)
		return nil
	})
}
