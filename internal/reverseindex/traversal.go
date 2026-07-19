package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

func discoverScopeFolders(repositoryRoot string, roots []string) (*ignorepolicy.Hierarchy, map[string]struct{}, error) {
	hierarchy, err := ignorepolicy.LoadHierarchy(repositoryRoot)
	if err != nil {
		return nil, nil, err
	}
	folders := map[string]struct{}{}
	for _, root := range roots {
		if err := hierarchy.LoadAncestors(filepath.Dir(root)); err != nil {
			return nil, nil, err
		}
		ignored, err := hierarchy.Ignored(root, true)
		if err != nil {
			return nil, nil, err
		}
		if ignored {
			return nil, nil, fmt.Errorf("reverse-index root is ignored by .docignore: %s", root)
		}
		err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() {
				return nil
			}
			if path != root {
				if worktreeDirectory(entry.Name()) {
					return filepath.SkipDir
				}
				ignored, err := hierarchy.Ignored(path, true)
				if err != nil {
					return err
				}
				if ignored {
					return filepath.SkipDir
				}
			}
			if err := hierarchy.LoadDirectory(path); err != nil {
				return err
			}
			folders[path] = struct{}{}
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}
	return hierarchy, folders, nil
}

func ancestorDirectories(repositoryRoot string, roots []string) []string {
	values := map[string]struct{}{filepath.Clean(repositoryRoot): {}}
	for _, root := range roots {
		current := filepath.Dir(filepath.Clean(root))
		for {
			values[current] = struct{}{}
			if current == filepath.Clean(repositoryRoot) {
				break
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	return sortedFolders(values)
}
