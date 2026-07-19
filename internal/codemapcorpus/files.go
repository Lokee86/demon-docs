package codemapcorpus

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	git "github.com/go-git/go-git/v5"
)

func repositoryFiles(root string) ([]string, error) {
	policy, err := ignorepolicy.Load(root)
	if err != nil {
		return nil, err
	}
	if files, ok, err := gitCLIRepositoryFiles(root, policy); ok || err != nil {
		return files, err
	}
	if files, ok, err := indexedRepositoryFiles(root, policy); ok || err != nil {
		return files, err
	}
	return walkedRepositoryFiles(root, policy)
}

func indexedRepositoryFiles(root string, policy ignorepolicy.Policy) ([]string, bool, error) {
	repository, err := git.PlainOpen(root)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	index, err := repository.Storer.Index()
	if err != nil {
		return nil, true, err
	}
	files := map[string]struct{}{}
	for _, entry := range index.Entries {
		relative := normalizePath(entry.Name)
		if relative == "" {
			continue
		}
		fullPath := joinRepositoryPath(root, relative)
		ignored, err := policy.Ignored(fullPath, false)
		if err != nil {
			return nil, true, err
		}
		info, err := os.Lstat(fullPath)
		if os.IsNotExist(err) || ignored {
			continue
		}
		if err != nil {
			return nil, true, err
		}
		if info.Mode().IsRegular() {
			files[relative] = struct{}{}
		}
	}
	return sortedSet(files), true, nil
}

func walkedRepositoryFiles(root string, policy ignorepolicy.Policy) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath != root && entry.IsDir() && hasGitMarker(filePath) {
			return filepath.SkipDir
		}
		if filePath != root {
			ignored, err := policy.Ignored(filePath, entry.IsDir())
			if err != nil {
				return err
			}
			if ignored {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return nil
		}
		relative, err := repositoryRelative(root, filePath)
		if err != nil {
			return err
		}
		files = append(files, relative)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func hasGitMarker(directory string) bool {
	_, err := os.Lstat(filepath.Join(directory, ".git"))
	return err == nil
}

func joinRepositoryPath(root, relative string) string {
	return filepath.Join(root, filepath.FromSlash(relative))
}

func repositoryRelative(root, filePath string) (string, error) {
	relative, err := filepath.Rel(root, filePath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Clean(relative)), nil
}

func within(root, candidate string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
