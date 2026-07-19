package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/repository"
)

// ResolveRoots selects positional command paths when provided, otherwise the
// configured reverse-index roots. Positional relative paths are resolved from
// cwd; configured relative paths are resolved from the repository root.
func ResolveRoots(repositoryRoot, docsRoot, cwd string, commandPaths, configuredRoots []string) ([]string, error) {
	base := repositoryRoot
	selected := configuredRoots
	if len(commandPaths) > 0 {
		base = cwd
		selected = commandPaths
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no reverse-index roots configured; add [reverse_index].roots or pass a directory path")
	}

	hierarchy, err := ignorepolicy.LoadHierarchy(repositoryRoot)
	if err != nil {
		return nil, err
	}
	resolved := make([]string, 0, len(selected))
	for _, raw := range selected {
		if strings.TrimSpace(raw) == "" {
			return nil, fmt.Errorf("reverse-index root cannot be empty")
		}
		path := raw
		if !filepath.IsAbs(path) {
			path = filepath.Join(base, path)
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		path = filepath.Clean(path)
		info, statErr := os.Stat(path)
		if statErr != nil {
			return nil, fmt.Errorf("reverse-index root %s: %w", path, statErr)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("reverse-index root is not a directory: %s", path)
		}
		if path == filepath.Clean(repositoryRoot) {
			return nil, fmt.Errorf("repository root cannot be a reverse-index root: %s", path)
		}
		if !repository.Contains(repositoryRoot, path) {
			return nil, fmt.Errorf("reverse-index root is outside repository: %s", path)
		}
		if inside(path, docsRoot) {
			return nil, fmt.Errorf("reverse-index root overlaps docs root: %s", path)
		}
		if hasWorktreePart(repositoryRoot, path) {
			return nil, fmt.Errorf("reverse-index root is inside a worktree control directory: %s", path)
		}
		if err := hierarchy.LoadAncestors(filepath.Dir(path)); err != nil {
			return nil, err
		}
		ignored, err := hierarchy.Ignored(path, true)
		if err != nil {
			return nil, err
		}
		if ignored {
			return nil, fmt.Errorf("reverse-index root is ignored by .docignore: %s", path)
		}
		resolved = append(resolved, path)
	}

	sort.Slice(resolved, func(i, j int) bool {
		if len(resolved[i]) == len(resolved[j]) {
			return resolved[i] < resolved[j]
		}
		return len(resolved[i]) < len(resolved[j])
	})
	collapsed := make([]string, 0, len(resolved))
	for _, path := range resolved {
		covered := false
		for _, existing := range collapsed {
			if inside(path, existing) {
				covered = true
				break
			}
		}
		if !covered {
			collapsed = append(collapsed, path)
		}
	}
	sort.Strings(collapsed)
	return collapsed, nil
}

func insideAny(path string, roots []string) bool {
	for _, root := range roots {
		if inside(path, root) {
			return true
		}
	}
	return false
}
