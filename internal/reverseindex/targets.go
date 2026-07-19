package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/codemap"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

func (f facts) addTarget(repositoryRoot, docsRoot, relative, document string, policy ignorepolicy.Policy) (bool, error) {
	relative = filepath.ToSlash(filepath.Clean(relative))
	full := filepath.Join(repositoryRoot, filepath.FromSlash(relative))
	info, err := os.Stat(full)
	if err != nil {
		return false, fmt.Errorf("resolved target unavailable: %s", relative)
	}
	accepted, err := eligibleTarget(repositoryRoot, docsRoot, full, info.IsDir(), policy)
	if err != nil || !accepted {
		return false, err
	}
	if info.IsDir() {
		addReference(f.folderDocs, relative, document)
		f.eligibleFolder[full] = struct{}{}
		return true, nil
	}
	parent := filepath.Dir(full)
	if filepath.Clean(parent) == filepath.Clean(repositoryRoot) {
		return false, nil
	}
	addReference(f.fileDocs, relative, document)
	f.exactFiles[relative] = struct{}{}
	f.eligibleFolder[parent] = struct{}{}
	return true, nil
}

func eligibleTarget(repositoryRoot, docsRoot, path string, isDir bool, policy ignorepolicy.Policy) (bool, error) {
	if filepath.Clean(path) == filepath.Clean(repositoryRoot) || inside(path, docsRoot) {
		return false, nil
	}
	if hasWorktreePart(repositoryRoot, path) {
		return false, nil
	}
	ignored, err := policy.Ignored(path, isDir)
	if err != nil {
		return false, err
	}
	return !ignored, nil
}

func resolvedPaths(record codemap.TargetRecord) []string {
	switch record.Status {
	case codemap.ResolutionResolved, codemap.ResolutionSymbolUnverified:
		if record.ResolvedPath != "" {
			return []string{record.ResolvedPath}
		}
	case codemap.ResolutionPatternResolved:
		result := make([]string, 0, len(record.Matches))
		for _, match := range record.Matches {
			result = append(result, match.Path)
		}
		return result
	}
	return nil
}
