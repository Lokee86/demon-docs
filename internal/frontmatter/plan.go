package frontmatter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/filetxn"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/validationcache"
)

type Plan struct {
	Updates     []model.FileUpdate
	Diagnostics []Diagnostic
	immutable   map[string]map[string]any
	rewrites    []filetxn.Rewrite
	cacheHits   int
}

func Build(repoRoot, docsRoot string, cfg config.Config, repair bool, now time.Time) (Plan, error) {
	if !cfg.Frontmatter.Enabled {
		return BuildWithValidationCache(repoRoot, docsRoot, cfg, repair, now, nil)
	}
	cache, err := validationcache.Open(repoRoot)
	if err != nil {
		return Plan{}, fmt.Errorf("open validation cache: %w", err)
	}
	plan, err := BuildWithValidationCache(repoRoot, docsRoot, cfg, repair, now, cache)
	if err != nil {
		return plan, err
	}
	if err := cache.Save(); err != nil {
		return plan, fmt.Errorf("save validation cache: %w", err)
	}
	return plan, nil
}

// BuildWithValidationCache builds one frontmatter plan against a caller-owned
// command cache. The caller publishes the cache after all parallel planners
// complete, keeping private-state writes outside the planning phase.
func BuildWithValidationCache(repoRoot, docsRoot string, cfg config.Config, repair bool, now time.Time, cache *validationcache.Store) (Plan, error) {
	plan := Plan{immutable: map[string]map[string]any{}}
	if !cfg.Frontmatter.Enabled {
		return plan, nil
	}
	if cache == nil {
		return plan, errors.New("frontmatter validation cache is required")
	}
	if err := ValidateConfig(cfg.Frontmatter); err != nil {
		return plan, err
	}
	files, err := markdownFiles(repoRoot, docsRoot)
	if err != nil {
		return plan, err
	}
	immutable := loadImmutableIndex(repoRoot)
	cache.Retain(relativeValidationPaths(repoRoot, files))
	schemaHasher := validationcache.NewSchemaHasher(repoRoot, cfg.Format)
	sources, duplicateExisting, err := loadSources(repoRoot, files, cfg.Frontmatter.AllowedFormats, cfg, immutable, cache, schemaHasher)
	if err != nil {
		return plan, err
	}
	duplicateOwners := selectDuplicateOwners(sources, immutable)
	usedIDs := collectDocumentIDs(sources)
	ids := map[string][]string{}
	for _, source := range sources {
		path := source.path
		relative := source.relative
		doc := source.document
		parsed := source.parsed
		if source.cacheHit {
			plan.cacheHits++
			if repair && len(source.cacheEntry.ImmutableValues) > 0 {
				current := immutable.values(relative, map[string]any{"document_id": source.cacheEntry.DocumentID}, true)
				if validationcache.Hash(current) != validationcache.Hash(source.cacheEntry.ImmutableValues) {
					plan.immutable[relative] = cloneValues(source.cacheEntry.ImmutableValues)
					cacheEntry := source.cacheEntry
					cacheEntry.ImmutableSnapshotHash = validationcache.Hash(cacheEntry.ImmutableValues)
					cache.Merge(cacheEntry)
				}
			}
			if id := sourceDocumentID(source); id != "" {
				ids[id] = append(ids[id], relative)
			}
			continue
		}
		if source.parseErr != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: source.parseErr.Error()})
			continue
		}
		if !parsed.HasBlock {
			parsed.Format = cfg.Frontmatter.DefaultFormat
		}
		frontmatterSchema, err := schemaForDocument(relative, parsed.Values, cfg)
		if err != nil {
			return plan, fmt.Errorf("resolve frontmatter defaults for %s: %w", relative, err)
		}
		currentID := documentID(parsed.Values)
		owner, duplicated := duplicateOwners[currentID]
		duplicateIDChanged := false
		allowIDHistory := !duplicateExisting[path] || duplicated && relative == owner
		recorded := immutable.values(relative, parsed.Values, allowIDHistory)
		if repair && duplicated && relative != owner {
			replacement, ok, replacementErr := generateUniqueDocumentID(frontmatterSchema, now, usedIDs)
			if replacementErr != nil {
				return plan, fmt.Errorf("repair duplicate document_id for %s: %w", relative, replacementErr)
			}
			if ok {
				parsed.Values = cloneValues(parsed.Values)
				parsed.Values["document_id"] = replacement
				recorded = cloneValues(recorded)
				delete(recorded, "document_id")
				usedIDs[replacement] = struct{}{}
				duplicateIDChanged = true
				plan.Diagnostics = append(plan.Diagnostics, Diagnostic{
					Path:     relative,
					Field:    "document_id",
					Message:  fmt.Sprintf("duplicate document_id %s also used by %s; fix assigned %s", currentID, owner, replacement),
					Resolved: true,
				})
			}
		}
		outcome := Evaluate(relative, parsed, frontmatterSchema, repair, recorded, now)
		if duplicateIDChanged {
			outcome.Changed = true
		}
		plan.Diagnostics = append(plan.Diagnostics, outcome.Diagnostics...)
		if id := documentID(outcome.Values); id != "" {
			ids[id] = append(ids[id], relative)
		}
		if len(outcome.Immutable) > 0 {
			plan.immutable[relative] = outcome.Immutable
		}
		if repair && (!parsed.HasBlock || outcome.Changed) {
			next, err := Render(parsed.Format, outcome.Values, parsed.Body)
			if err != nil {
				return plan, fmt.Errorf("render frontmatter %s: %w", path, err)
			}
			if next != doc.Text {
				old := doc.Text
				plan.Updates = append(plan.Updates, model.FileUpdate{Path: path, OldText: &old, NewText: next})
				plan.rewrites = append(plan.rewrites, filetxn.New(path, doc.Encode(old), doc.Encode(next)))
			}
		}
		if len(outcome.Diagnostics) == 0 && !duplicateExisting[path] && !outcome.Changed {
			immutableSnapshot := recorded
			if repair && len(outcome.Immutable) > 0 {
				immutableSnapshot = outcome.Immutable
			}
			cache.Merge(validationcache.Entry{
				Path:                  relative,
				ContentSHA256:         source.contentHash,
				EngineVersion:         validationcache.EngineVersion,
				FrontmatterPolicyHash: validationcache.FrontmatterPolicyHash(cfg),
				EffectiveSchemaHash:   selectedSchemaHash(schemaHasher, relative, parsed.Values, cfg),
				ImmutableSnapshotHash: validationcache.Hash(immutableSnapshot),
				DocumentID:            documentID(outcome.Values),
				DocumentType:          stringValue(outcome.Values["document_type"]),
				SchemaName:            selectedSchemaName(relative, parsed.Values, cfg),
				ImmutableValues:       cloneValues(outcome.Immutable),
				FrontmatterClean:      true,
			})
		}
	}
	for id, paths := range ids {
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		for _, path := range paths {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: path, Field: "document_id", Message: fmt.Sprintf("duplicate document_id %s also used by %s", id, strings.Join(otherPaths(paths, path), ", "))})
			if values := plan.immutable[path]; values != nil {
				delete(values, "document_id")
				if len(values) == 0 {
					delete(plan.immutable, path)
				}
			}
		}
	}
	sort.Slice(plan.Diagnostics, func(i, j int) bool {
		left, right := plan.Diagnostics[i], plan.Diagnostics[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Field != right.Field {
			return left.Field < right.Field
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

func Apply(repoRoot, docsRoot string, plan Plan) (int, error) {
	if len(plan.rewrites) != len(plan.Updates) {
		return 0, errors.New("frontmatter plan does not contain prepared rewrite transactions")
	}
	for _, rewrite := range plan.rewrites {
		if !repository.Contains(docsRoot, rewrite.Path()) {
			return 0, fmt.Errorf("refusing to write frontmatter outside docs root: %s", rewrite.Path())
		}
	}
	if _, err := filetxn.Apply(plan.rewrites); err != nil {
		return 0, fmt.Errorf("apply frontmatter rewrites: %w", err)
	}
	if err := writeImmutable(repoRoot, plan.immutable); err != nil {
		if rollbackErr := filetxn.Rollback(plan.rewrites); rollbackErr != nil {
			return 0, errors.Join(
				fmt.Errorf("save frontmatter immutable state: %w", err),
				fmt.Errorf("rollback frontmatter rewrites: %w", rollbackErr),
			)
		}
		return 0, fmt.Errorf("save frontmatter immutable state: %w; rewrites rolled back", err)
	}
	if err := validationcache.RefreshTransactions(
		repoRoot,
		plan.rewrites,
		validationcache.SurfaceFrontmatter|validationcache.SurfaceFormat,
	); err != nil {
		return len(plan.rewrites), fmt.Errorf("refresh validation cache after frontmatter rewrites: %w", err)
	}
	return len(plan.rewrites), nil
}

func markdownFiles(repoRoot, docsRoot string) ([]string, error) {
	policy, err := ignorepolicy.Load(repoRoot)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(docsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		ignored, err := policy.Ignored(path, entry.IsDir())
		if err != nil {
			return err
		}
		if ignored {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		left, right := files[i], files[j]
		if runtime.GOOS == "windows" {
			left, right = strings.ToLower(left), strings.ToLower(right)
		}
		return left < right
	})
	return files, nil
}

func schemaForDocument(relative string, values map[string]any, cfg config.Config) (config.Frontmatter, error) {
	schema := cfg.Frontmatter
	fields := make(map[string]config.FrontmatterField, len(schema.Fields))
	for name, definition := range schema.Fields {
		fields[name] = definition
	}
	if cfg.Format.Enabled {
		selectionValues := values
		if value, present := values["document_type"]; present {
			if !emptyValue(value) {
				selectionValues = nil
			} else {
				selectionValues = make(map[string]any, len(values)-1)
				for name, existing := range values {
					if name != "document_type" {
						selectionValues[name] = existing
					}
				}
			}
		}
		if selectionValues != nil {
			selected, err := config.SelectFormatSchema(relative, selectionValues, cfg.Format)
			if err != nil {
				return schema, err
			}
			if definition, ok := fields["document_type"]; ok {
				definition.Default = selected
				definition.DefaultFrom = ""
				fields["document_type"] = definition
			}
		}
	}
	if strings.EqualFold(filepath.Base(filepath.FromSlash(relative)), cfg.IndexFile) {
		if definition, ok := fields["author"]; ok && !hasConfiguredSource(definition, schema) {
			definition.Default = "TODO"
			definition.DefaultFrom = ""
			fields["author"] = definition
		}
		if definition, ok := fields["summary"]; ok && !hasConfiguredSource(definition, schema) {
			definition.Default = "Generated documentation folder index."
			definition.DefaultFrom = ""
			fields["summary"] = definition
		}
	}
	schema.Fields = fields
	return schema, nil
}

func selectedSchemaName(relative string, values map[string]any, cfg config.Config) string {
	if !cfg.Format.Enabled {
		return ""
	}
	selectionValues := values
	if value, present := values["document_type"]; present && emptyValue(value) {
		selectionValues = make(map[string]any, len(values)-1)
		for name, existing := range values {
			if name != "document_type" {
				selectionValues[name] = existing
			}
		}
	}
	name, err := config.SelectFormatSchema(relative, selectionValues, cfg.Format)
	if err != nil {
		return ""
	}
	return name
}

func selectedSchemaHash(schemaHasher *validationcache.SchemaHasher, relative string, values map[string]any, cfg config.Config) string {
	return schemaHasher.Effective(selectedSchemaName(relative, values, cfg), documentID(values))
}

func relativeValidationPaths(repoRoot string, files []string) []string {
	paths := make([]string, 0, len(files))
	for _, path := range files {
		relative, err := filepath.Rel(repoRoot, path)
		if err == nil {
			paths = append(paths, filepath.ToSlash(relative))
		}
	}
	return paths
}

func stringValue(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func otherPaths(paths []string, current string) []string {
	result := make([]string, 0, len(paths)-1)
	for _, path := range paths {
		if path != current {
			result = append(result, path)
		}
	}
	return result
}
