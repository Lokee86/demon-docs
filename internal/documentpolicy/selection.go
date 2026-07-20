package documentpolicy

import (
	"os"
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
	return config.SelectFormatSchema(relative, values, cfg)
}

func validatePathPattern(pattern string) error {
	return config.ValidateFormatPathPattern(pattern)
}

func matchPath(pattern, name string) (bool, error) {
	return config.MatchFormatPath(pattern, name)
}
