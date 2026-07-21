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
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/validationcache"
)

type Diagnostic struct {
	Path     string
	Section  string
	Message  string
	Options  []string
	Resolved bool
	Warning  bool
}

type schemaHistoryResult struct {
	schema Schema
	exists bool
	err    error
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
	cacheHits          int
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
	cache, err := validationcache.Open(repoRoot)
	if err != nil {
		return plan, fmt.Errorf("open validation cache: %w", err)
	}
	activePaths := make([]string, 0, len(files))
	for _, path := range files {
		relative, relativeErr := filepath.Rel(repoRoot, path)
		if relativeErr == nil {
			activePaths = append(activePaths, filepath.ToSlash(relative))
		}
	}
	cache.Retain(activePaths)
	policyHash := validationcache.FrontmatterPolicyHash(cfg)
	schemaHasher := validationcache.NewSchemaHasher(repoRoot, cfg.Format)
	sharedSchemas := map[string]Schema{}
	sharedErrors := map[string]error{}
	sharedLoaded := map[string]bool{}
	loadSharedOnce := func(name string) (Schema, error) {
		if sharedLoaded[name] {
			return sharedSchemas[name], sharedErrors[name]
		}
		sharedLoaded[name] = true
		schema, _, loadErr := LoadShared(repoRoot, cfg.Format, name)
		sharedSchemas[name] = schema
		sharedErrors[name] = loadErr
		return schema, loadErr
	}
	historyResults := map[string]schemaHistoryResult{}
	loadHistoryOnce := func(name string) (Schema, bool, error) {
		if result, ok := historyResults[name]; ok {
			return result.schema, result.exists, result.err
		}
		schema, exists, historyErr := loadSchemaHistory(repoRoot, name)
		historyResults[name] = schemaHistoryResult{schema: schema, exists: exists, err: historyErr}
		return schema, exists, historyErr
	}
	sources, err := loadDocumentSources(repoRoot, files, cfg, cache, policyHash, schemaHasher)
	if err != nil {
		return plan, err
	}
	evaluations := make([]documentEvaluation, 0, len(sources))
	for _, source := range sources {
		relative := source.relative
		if source.cacheHit {
			shared, loadErr := loadSharedOnce(source.candidate.SchemaName)
			if loadErr != nil {
				plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: loadErr.Error()})
				continue
			}
			plan.history[source.candidate.SchemaName] = shared
			if _, _, historyErr := loadHistoryOnce(source.candidate.SchemaName); historyErr != nil {
				return plan, historyErr
			}
			plan.cacheHits++
			continue
		}
		if source.parseErr != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: source.parseErr.Error()})
			plan.blockAllHistory = true
			continue
		}
		if source.schemaErr != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: source.schemaErr.Error()})
			plan.blockAllHistory = true
			continue
		}
		schemaName := source.schemaName
		if schemaName == "" {
			continue
		}
		shared, err := loadSharedOnce(schemaName)
		if err != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
			continue
		}
		plan.history[schemaName] = shared
		previous, hasPrevious, err := loadHistoryOnce(schemaName)
		if err != nil {
			return plan, err
		}
		documentID, _ := source.parsed.Values["document_id"].(string)
		documentType, _ := source.parsed.Values["document_type"].(string)
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
		evaluations = append(evaluations, documentEvaluation{
			path:         source.path,
			relative:     relative,
			data:         source.data,
			text:         source.text,
			contentHash:  source.contentHash,
			candidate:    source.candidate,
			schemaName:   schemaName,
			documentID:   documentID,
			documentType: documentType,
			bodyStart:    source.bodyStart,
			document:     source.document,
			current:      shared,
			previous:     previous,
			hasPrevious:  hasPrevious,
		})
	}

	runDocumentEvaluations(evaluations, repair)
	for _, evaluation := range evaluations {
		result := evaluation.result
		for _, diagnostic := range result.Diagnostics {
			diagnostic.Path = evaluation.relative
			plan.Diagnostics = append(plan.Diagnostics, diagnostic)
			if !diagnostic.Warning && !diagnostic.Resolved {
				plan.blockedHistory[evaluation.schemaName] = true
			}
		}
		if repair && result.Changed && !result.Blocked {
			next := evaluation.text[:evaluation.bodyStart] + result.Document.render()
			if next != evaluation.text {
				old := evaluation.text
				plan.Updates = append(plan.Updates, model.FileUpdate{Path: evaluation.path, OldText: &old, NewText: next})
				plan.rewrites = append(plan.rewrites, filetxn.New(evaluation.path, evaluation.data, []byte(next)))
			}
		}
		if len(result.Diagnostics) == 0 && !result.Changed {
			cache.Merge(validationcache.Entry{
				Path:                  evaluation.relative,
				ContentSHA256:         evaluation.contentHash,
				EngineVersion:         validationcache.EngineVersion,
				FrontmatterPolicyHash: policyHash,
				EffectiveSchemaHash:   schemaHasher.Effective(evaluation.schemaName, evaluation.documentID),
				ImmutableSnapshotHash: evaluation.candidate.ImmutableSnapshotHash,
				DocumentID:            strings.TrimSpace(evaluation.documentID),
				DocumentType:          strings.TrimSpace(evaluation.documentType),
				SchemaName:            evaluation.schemaName,
				FormatClean:           true,
			})
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
	if err := cache.Save(); err != nil {
		return plan, fmt.Errorf("save validation cache: %w", err)
	}
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
