package pathutil

import (
	"path/filepath"
	"strings"
)

func Resolve(root, cwd string) (string, error) {
	if filepath.IsAbs(root) {
		return filepath.Abs(root)
	}
	if cwd == "" {
		return filepath.Abs(root)
	}
	return filepath.Abs(filepath.Join(cwd, root))
}
func Relative(target, base string) (string, error) {
	target = stripExtended(target)
	base = stripExtended(base)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}
func stripExtended(path string) string { return strings.TrimPrefix(path, `\\?\`) }
