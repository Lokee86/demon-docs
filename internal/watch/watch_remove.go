package watch

import (
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/repository"
)

// removeWatchTree closes every watch rooted at path before forgetting it.
// Windows can retain inaccessible delete-pending directory entries when a
// parent tree is moved while descendant fsnotify handles remain open.
func removeWatchTree(w eventWatcher, path string, watched map[string]bool) {
	root := filepath.Clean(path)
	matches := make([]string, 0)
	for candidate := range watched {
		if repository.Contains(root, candidate) {
			matches = append(matches, candidate)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i]) > len(matches[j])
	})
	for _, candidate := range matches {
		_ = w.Remove(candidate)
		delete(watched, candidate)
	}
}
