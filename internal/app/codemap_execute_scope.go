package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/repository"
)

func resolveCodemapRoot(scope repository.Scope, value string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root := value
	if !filepath.IsAbs(root) {
		root = filepath.Join(cwd, root)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if !repository.Contains(scope.DocsRoot, root) {
		return "", fmt.Errorf("codemap root is outside the configured docs root: %s", value)
	}
	if _, err := os.Stat(root); err != nil {
		return "", fmt.Errorf("codemap root does not exist: %s", root)
	}
	return root, nil
}

func codemapTargetFiles(repositoryRoot, root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !strings.EqualFold(filepath.Ext(root), ".md") {
			return nil, fmt.Errorf("codemap root file must be Markdown: %s", root)
		}
		return []string{root}, nil
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		return nil, err
	}
	files := []string{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root {
			ignored, err := policy.Ignored(path, entry.IsDir())
			if err != nil {
				return err
			}
			worktree := entry.IsDir() && (entry.Name() == ".worktrees" || entry.Name() == ".workingtrees")
			if ignored || worktree {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 && strings.EqualFold(filepath.Ext(path), ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
