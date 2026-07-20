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
)

type Plan struct {
	Updates     []model.FileUpdate
	Diagnostics []Diagnostic
	immutable   map[string]map[string]any
	rewrites    []filetxn.Rewrite
}

func Build(repoRoot, docsRoot string, cfg config.Config, repair bool, now time.Time) (Plan, error) {
	plan := Plan{immutable: map[string]map[string]any{}}
	if !cfg.Frontmatter.Enabled {
		return plan, nil
	}
	if err := ValidateConfig(cfg.Frontmatter); err != nil {
		return plan, err
	}
	files, err := markdownFiles(repoRoot, docsRoot)
	if err != nil {
		return plan, err
	}
	sources, duplicateExisting, err := loadSources(repoRoot, files, cfg.Frontmatter.AllowedFormats)
	if err != nil {
		return plan, err
	}
	immutable := loadImmutableIndex(repoRoot)
	ids := map[string][]string{}
	for _, source := range sources {
		path := source.path
		relative := source.relative
		doc := source.document
		parsed := source.parsed
		if source.parseErr != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: source.parseErr.Error()})
			continue
		}
		if !parsed.HasBlock {
			parsed.Format = cfg.Frontmatter.DefaultFormat
		}
		recorded := immutable.values(relative, parsed.Values, !duplicateExisting[path])
		outcome := Evaluate(relative, parsed, cfg.Frontmatter, repair, recorded, now)
		plan.Diagnostics = append(plan.Diagnostics, outcome.Diagnostics...)
		if value, ok := outcome.Values["document_id"].(string); ok && strings.TrimSpace(value) != "" {
			ids[value] = append(ids[value], relative)
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

func otherPaths(paths []string, current string) []string {
	result := make([]string, 0, len(paths)-1)
	for _, path := range paths {
		if path != current {
			result = append(result, path)
		}
	}
	return result
}
