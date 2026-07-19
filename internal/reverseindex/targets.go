package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

func (f facts) addTarget(repositoryRoot string, roots []string, folders map[string]struct{}, hierarchy *ignorepolicy.Hierarchy, relative, document string) (bool, error) {
	relative = filepath.ToSlash(filepath.Clean(relative))
	full := filepath.Join(repositoryRoot, filepath.FromSlash(relative))
	if !insideAny(full, roots) {
		return false, nil
	}
	info, err := os.Stat(full)
	if err != nil {
		return false, fmt.Errorf("resolved target unavailable: %s", relative)
	}
	ignored, err := hierarchy.Ignored(full, info.IsDir())
	if err != nil {
		return false, err
	}
	if ignored {
		return false, nil
	}
	if info.IsDir() {
		if _, ok := folders[full]; !ok {
			return false, nil
		}
		addReference(f.folderDocs, relative, document)
		return true, nil
	}
	parent := filepath.Dir(full)
	if _, ok := folders[parent]; !ok {
		return false, nil
	}
	addReference(f.fileDocs, relative, document)
	f.exactFiles[relative] = struct{}{}
	return true, nil
}

func entryPotentiallyInScope(repositoryRoot string, roots []string, entry codemap.Entry, format codemap.Format) bool {
	target := entry.Target
	if entry.Kind == codemap.TargetSymbol && !strings.Contains(target, "#") && !strings.Contains(target, "::") {
		return false
	}
	if at := strings.Index(target, "#"); at >= 0 {
		target = target[:at]
	}
	if at := strings.Index(target, "::"); at >= 0 {
		target = target[:at]
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	bases := []string{repositoryRoot}
	if format.TargetBase == codemap.TargetBaseDocument {
		bases[0] = filepath.Join(repositoryRoot, filepath.Dir(filepath.FromSlash(entry.DocumentPath)))
	}
	for _, root := range format.TargetRoots {
		bases = append(bases, filepath.Join(repositoryRoot, filepath.FromSlash(root)))
	}
	for _, base := range bases {
		candidate := target
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(base, filepath.FromSlash(candidate))
		}
		if insideAny(filepath.Clean(candidate), roots) {
			return true
		}
	}
	return false
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
