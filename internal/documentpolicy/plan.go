package documentpolicy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/filetxn"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
)

type Diagnostic struct {
	Path     string
	Section  string
	Message  string
	Options  []string
	Resolved bool
	Warning  bool
}

type Plan struct {
	Updates            []model.FileUpdate
	Diagnostics        []Diagnostic
	rewrites           []filetxn.Rewrite
	invalidatedSchemas map[string][]byte
	history            map[string]Schema
	blockedHistory     map[string]bool
	blockAllHistory    bool
	repositoryRoot     string
}

func Build(repoRoot, docsRoot string, cfg config.Config, repair bool) (Plan, error) {
	plan := Plan{
		invalidatedSchemas: map[string][]byte{},
		history:            map[string]Schema{},
		blockedHistory:     map[string]bool{},
		repositoryRoot:     repoRoot,
	}
	if !cfg.Format.Enabled {
		return plan, nil
	}
	if cfg.Format.InvalidationSimilarity < 0 || cfg.Format.InvalidationSimilarity > 1 {
		return plan, fmt.Errorf("format.invalidation_similarity must be between 0 and 1")
	}
	if strings.TrimSpace(cfg.Format.SchemaDir) == "" || strings.TrimSpace(cfg.Format.DocumentSchemaDir) == "" {
		return plan, fmt.Errorf("format.schema_dir and format.document_schema_dir must be non-empty")
	}
	for index, rule := range cfg.Format.PathRules {
		if strings.TrimSpace(rule.Pattern) == "" || strings.TrimSpace(rule.Schema) == "" {
			return plan, fmt.Errorf("format.path_rules[%d] requires pattern and schema", index)
		}
		if err := validatePathPattern(rule.Pattern); err != nil {
			return plan, fmt.Errorf("invalid format.path_rules[%d] pattern %q: %w", index, rule.Pattern, err)
		}
	}
	files, err := markdownFiles(repoRoot, docsRoot)
	if err != nil {
		return plan, err
	}
	for _, path := range files {
		relative, _ := filepath.Rel(repoRoot, path)
		relative = filepath.ToSlash(relative)
		data, err := os.ReadFile(path)
		if err != nil {
			return plan, err
		}
		source := string(data)
		parsed, err := frontmatter.Parse(source, cfg.Frontmatter.AllowedFormats)
		if err != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
			plan.blockAllHistory = true
			continue
		}
		schemaName, err := selectSchema(relative, parsed.Values, cfg.Format)
		if err != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
			plan.blockAllHistory = true
			continue
		}
		if schemaName == "" {
			continue
		}
		shared, _, err := LoadShared(repoRoot, cfg.Format, schemaName)
		if err != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
			continue
		}
		plan.history[schemaName] = shared
		previous, hasPrevious, err := loadSchemaHistory(repoRoot, schemaName)
		if err != nil {
			return plan, err
		}
		documentID, _ := parsed.Values["document_id"].(string)
		local := DocumentSchema{}
		localPath := ""
		localExists := false
		if strings.TrimSpace(documentID) != "" {
			local, localPath, localExists, err = LoadDocumentSchema(repoRoot, cfg.Format, documentID)
			if err != nil {
				plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
				plan.blockedHistory[schemaName] = true
				continue
			}
		}
		if localExists && local.DocumentID != "" && local.DocumentID != documentID {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: fmt.Sprintf("document-specific schema identifies document %q instead of %q", local.DocumentID, documentID)})
			plan.blockedHistory[schemaName] = true
			continue
		}
		if localExists && local.SharedSchema != "" && local.SharedSchema != schemaName {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: fmt.Sprintf("document-specific schema extends %q but metadata selects %q", local.SharedSchema, schemaName)})
			plan.blockedHistory[schemaName] = true
			continue
		}
		if localExists && cfg.Format.InvalidationSimilarity > 0 {
			currentFingerprint := Fingerprint(shared)
			accepted := Schema{}
			hasAccepted := false
			switch {
			case local.SharedFingerprint == currentFingerprint:
			case strings.TrimSpace(local.SharedFingerprint) == "":
				accepted, hasAccepted = previous, hasPrevious
			default:
				accepted, hasAccepted, err = loadSchemaSnapshot(repoRoot, schemaName, local.SharedFingerprint)
				if err != nil {
					plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
					plan.blockedHistory[schemaName] = true
					continue
				}
				if !hasAccepted && hasPrevious && Fingerprint(previous) == local.SharedFingerprint {
					accepted, hasAccepted = previous, true
				}
				if !hasAccepted {
					plan.Diagnostics = append(plan.Diagnostics, Diagnostic{
						Path:    relative,
						Message: fmt.Sprintf("accepted shared-schema snapshot %s is missing; document-specific exceptions cannot be evaluated safely", local.SharedFingerprint),
					})
					plan.blockedHistory[schemaName] = true
					continue
				}
			}
			if hasAccepted {
				if !hasPrevious {
					previous, hasPrevious = accepted, true
				}
				similarity := Similarity(accepted, shared)
				if similarity < cfg.Format.InvalidationSimilarity {
					plan.Diagnostics = append(plan.Diagnostics, Diagnostic{
						Path: relative, Message: fmt.Sprintf("document-specific schema invalidated because shared schema similarity is %.2f below %.2f", similarity, cfg.Format.InvalidationSimilarity), Resolved: repair,
					})
					if repair {
						backup, readErr := os.ReadFile(localPath)
						if readErr != nil {
							return plan, readErr
						}
						plan.invalidatedSchemas[localPath] = backup
						local = DocumentSchema{}
						localExists = false
					}
				}
			}
		}
		if localExists {
			shared = EffectiveSchema(shared, local)
			if err := ValidateSchema(shared); err != nil {
				plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: fmt.Sprintf("invalid effective document schema: %v", err)})
				plan.blockedHistory[schemaName] = true
				continue
			}
		}
		bodyStart := frontmatter.LeadingBlockEnd(source)
		body := source[bodyStart:]
		document := parseMarkdown(body)
		result := enforceDocument(document, shared, previous, hasPrevious, repair)
		for _, diagnostic := range result.Diagnostics {
			diagnostic.Path = relative
			plan.Diagnostics = append(plan.Diagnostics, diagnostic)
			if !diagnostic.Warning && !diagnostic.Resolved {
				plan.blockedHistory[schemaName] = true
			}
		}
		if repair && result.Changed && !result.Blocked {
			next := source[:bodyStart] + result.Document.render()
			if next != source {
				old := source
				plan.Updates = append(plan.Updates, model.FileUpdate{Path: path, OldText: &old, NewText: next})
				plan.rewrites = append(plan.rewrites, filetxn.New(path, data, []byte(next)))
			}
		}
	}
	sort.Slice(plan.Diagnostics, func(i, j int) bool {
		left, right := plan.Diagnostics[i], plan.Diagnostics[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Section != right.Section {
			return left.Section < right.Section
		}
		return left.Message < right.Message
	})
	return plan, nil
}

func (plan Plan) Failed() bool {
	for _, diagnostic := range plan.Diagnostics {
		if !diagnostic.Warning && !diagnostic.Resolved {
			return true
		}
	}
	return false
}

func Apply(plan Plan, docsRoot string) (int, error) {
	for _, rewrite := range plan.rewrites {
		if !repository.Contains(docsRoot, rewrite.Path()) {
			return 0, fmt.Errorf("refusing to write document format outside docs root: %s", rewrite.Path())
		}
	}
	if _, err := filetxn.Apply(plan.rewrites); err != nil {
		return 0, fmt.Errorf("apply document-format rewrites: %w", err)
	}
	removed := make([]string, 0, len(plan.invalidatedSchemas))
	for path := range plan.invalidatedSchemas {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			_ = filetxn.Rollback(plan.rewrites)
			return 0, fmt.Errorf("remove invalidated document schema %s: %w", path, err)
		}
		removed = append(removed, path)
	}
	history := make(map[string]Schema, len(plan.history))
	if !plan.blockAllHistory {
		for name, schema := range plan.history {
			if !plan.blockedHistory[name] {
				history[name] = schema
			}
		}
	}
	if err := saveSchemaHistory(plan.repositoryRoot, history); err != nil {
		for _, path := range removed {
			_ = os.MkdirAll(filepath.Dir(path), 0o755)
			_ = os.WriteFile(path, plan.invalidatedSchemas[path], 0o644)
		}
		if rollbackErr := filetxn.Rollback(plan.rewrites); rollbackErr != nil {
			return 0, errors.Join(err, rollbackErr)
		}
		return 0, err
	}
	return len(plan.rewrites) + len(removed), nil
}
