package reverseindex

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/textio"
)

var sourceExtensions = map[string]struct{}{
	".astro": {}, ".bash": {}, ".c": {}, ".cc": {}, ".cpp": {}, ".cs": {},
	".css": {}, ".erb": {}, ".ex": {}, ".exs": {}, ".gd": {}, ".gdshader": {},
	".go": {}, ".graphql": {}, ".gql": {}, ".h": {}, ".hpp": {}, ".html": {},
	".java": {}, ".js": {}, ".jsx": {}, ".kt": {}, ".kts": {}, ".lua": {},
	".php": {}, ".proto": {}, ".ps1": {}, ".py": {}, ".rake": {}, ".rb": {},
	".rs": {}, ".sass": {}, ".scala": {}, ".scss": {}, ".sh": {}, ".sql": {},
	".svelte": {}, ".swift": {}, ".ts": {}, ".tsx": {}, ".vue": {},
}

var sourceNames = map[string]struct{}{
	"Dockerfile": {}, "Gemfile": {}, "Makefile": {}, "Procfile": {}, "Rakefile": {},
}

func inventoryFolders(repositoryRoot, docsRoot string, c config.Config, policy ignorepolicy.Policy, f facts) (map[string][]string, map[string]struct{}, error) {
	folderFiles := map[string][]string{}
	for folder := range f.eligibleFolder {
		entries, err := os.ReadDir(folder)
		if err != nil {
			return nil, nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() || entry.Name() == c.IndexFile {
				continue
			}
			path := filepath.Join(folder, entry.Name())
			ignored, err := policy.Ignored(path, false)
			if err != nil {
				return nil, nil, err
			}
			if ignored || !indexableCodeFile(repositoryRoot, path, f.exactFiles) {
				continue
			}
			folderFiles[folder] = append(folderFiles[folder], path)
		}
		sort.Strings(folderFiles[folder])
	}
	existing, err := findManagedIndexes(repositoryRoot, docsRoot, c, policy)
	return folderFiles, existing, err
}

func indexableCodeFile(repositoryRoot, path string, exact map[string]struct{}) bool {
	relative, err := filepath.Rel(repositoryRoot, path)
	if err == nil {
		if _, ok := exact[filepath.ToSlash(filepath.Clean(relative))]; ok {
			return true
		}
	}
	if _, ok := sourceNames[filepath.Base(path)]; ok {
		return true
	}
	_, ok := sourceExtensions[strings.ToLower(filepath.Ext(path))]
	return ok
}

func findManagedIndexes(repositoryRoot, docsRoot string, c config.Config, policy ignorepolicy.Policy) (map[string]struct{}, error) {
	managed := map[string]struct{}{}
	err := filepath.WalkDir(repositoryRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != repositoryRoot {
			if inside(path, docsRoot) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() && worktreeDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			ignored, err := policy.Ignored(path, entry.IsDir())
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
		if entry.IsDir() || filepath.Base(path) != c.IndexFile {
			return nil
		}
		doc, err := textio.Read(path)
		if err == nil && strings.Contains(doc.Text, markerStart(c)) {
			managed[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	return managed, err
}
