package frontmatter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/textio"
)

type Plan struct {
	Updates     []model.FileUpdate
	Diagnostics []Diagnostic
	immutable   map[string]map[string]any
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
	duplicateExisting, err := duplicateExistingIDs(files, cfg.Frontmatter)
	if err != nil {
		return plan, err
	}
	ids := map[string][]string{}
	for _, path := range files {
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return plan, err
		}
		relative = filepath.ToSlash(relative)
		doc, err := textio.Read(path)
		if err != nil {
			return plan, fmt.Errorf("read frontmatter source %s: %w", path, err)
		}
		parsed, err := Parse(doc.Text, cfg.Frontmatter.AllowedFormats)
		if err != nil {
			plan.Diagnostics = append(plan.Diagnostics, Diagnostic{Path: relative, Message: err.Error()})
			continue
		}
		if !parsed.HasBlock {
			parsed.Format = cfg.Frontmatter.DefaultFormat
		}
		recorded := readImmutable(repoRoot, relative, parsed.Values, !duplicateExisting[path])
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
	changed := 0
	for _, update := range plan.Updates {
		if !repository.Contains(docsRoot, update.Path) {
			return changed, fmt.Errorf("refusing to write frontmatter outside docs root: %s", update.Path)
		}
		document, err := textio.Read(update.Path)
		if err != nil {
			return changed, err
		}
		if update.OldText != nil && document.Text != *update.OldText {
			return changed, fmt.Errorf("frontmatter source changed before apply: %s", update.Path)
		}
		if err := atomicWrite(update.Path, document.Encode(update.NewText)); err != nil {
			return changed, err
		}
		changed++
	}
	if err := writeImmutable(repoRoot, plan.immutable); err != nil {
		return changed, fmt.Errorf("save frontmatter immutable state: %w", err)
	}
	return changed, nil
}

func duplicateExistingIDs(files []string, schema config.Frontmatter) (map[string]bool, error) {
	pathsByID := map[string][]string{}
	for _, path := range files {
		document, err := textio.Read(path)
		if err != nil {
			return nil, err
		}
		parsed, err := Parse(document.Text, schema.AllowedFormats)
		if err != nil {
			continue
		}
		if id := documentID(parsed.Values); id != "" {
			pathsByID[id] = append(pathsByID[id], path)
		}
	}
	duplicates := map[string]bool{}
	for _, paths := range pathsByID {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			duplicates[path] = true
		}
	}
	return duplicates, nil
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

func atomicWrite(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".frontmatter-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return atomicReplace(name, path)
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
