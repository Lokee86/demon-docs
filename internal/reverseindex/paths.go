package reverseindex

import (
	"path/filepath"
	"strings"
)

func inside(path, root string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func hasWorktreePart(repositoryRoot, path string) bool {
	relative, err := filepath.Rel(repositoryRoot, path)
	if err != nil {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		if worktreeDirectory(part) {
			return true
		}
	}
	return false
}

func worktreeDirectory(name string) bool {
	return name == ".worktrees" || name == ".workingtrees"
}
