package documentpolicy

import (
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

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
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(path), ".md") {
			files = append(files, path)
		}
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

func selectSchema(relative string, values map[string]any, cfg config.Format) (string, error) {
	if raw, ok := values["document_type"]; ok {
		name, ok := raw.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return "", fmt.Errorf("document_type metadata must be a non-empty string")
		}
		return strings.TrimSpace(name), nil
	}
	for _, rule := range cfg.PathRules {
		matched, err := matchPath(rule.Pattern, relative)
		if err != nil {
			return "", fmt.Errorf("invalid format path pattern %q: %w", rule.Pattern, err)
		}
		if matched {
			return rule.Schema, nil
		}
	}
	return cfg.DefaultSchema, nil
}

func validatePathPattern(pattern string) error {
	pattern = strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(pattern)), "./")
	for _, segment := range strings.Split(pattern, "/") {
		if segment == "**" {
			continue
		}
		if _, err := pathpkg.Match(segment, ""); err != nil {
			return err
		}
	}
	return nil
}

func matchPath(pattern, name string) (bool, error) {
	pattern = strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(pattern)), "./")
	name = strings.TrimPrefix(filepath.ToSlash(name), "./")
	return matchPathSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchPathSegments(pattern, name []string) (bool, error) {
	if len(pattern) == 0 {
		return len(name) == 0, nil
	}
	if pattern[0] == "**" {
		for len(pattern) > 1 && pattern[1] == "**" {
			pattern = pattern[1:]
		}
		if len(pattern) == 1 {
			return true, nil
		}
		for consumed := 0; consumed <= len(name); consumed++ {
			matched, err := matchPathSegments(pattern[1:], name[consumed:])
			if err != nil || matched {
				return matched, err
			}
		}
		return false, nil
	}
	if len(name) == 0 {
		return false, nil
	}
	matched, err := pathpkg.Match(pattern[0], name[0])
	if err != nil || !matched {
		return matched, err
	}
	return matchPathSegments(pattern[1:], name[1:])
}
