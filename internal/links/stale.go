package links

import (
	"errors"
	"io/fs"
	"strings"
)

// IsTransientFilesystemRace reports planning or application failures caused by
// repository files moving or changing during a reconciliation pass. Callers may
// safely discard the stale plan and rebuild it from the current filesystem.
func IsTransientFilesystemRace(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, fs.ErrNotExist) {
		return true
	}
	message := err.Error()
	for _, fragment := range []string{
		"old destination mismatch",
		"rewrite source changed before apply",
		"document changed since format plan creation",
		"document changed since frontmatter plan creation",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}
