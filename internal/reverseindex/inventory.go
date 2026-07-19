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

func inventoryFolders(repositoryRoot string, c config.Config, hierarchy *ignorepolicy.Hierarchy, folders map[string]struct{}, f facts) (map[string][]string, map[string]struct{}, error) {
	folderFiles := map[string][]string{}
	existingManaged := map[string]struct{}{}
	for _, folder := range sortedFolders(folders) {
		indexPath := filepath.Join(folder, c.IndexFile)
		if doc, err := textio.Read(indexPath); err == nil && strings.Contains(doc.Text, markerStart(c)) {
			existingManaged[folder] = struct{}{}
		} else if err != nil && !os.IsNotExist(err) {
			return nil, nil, err
		}

		entries, err := os.ReadDir(folder)
		if err != nil {
			return nil, nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() || entry.Name() == c.IndexFile || entry.Name() == ignorepolicy.FileName {
				continue
			}
			path := filepath.Join(folder, entry.Name())
			ignored, err := hierarchy.Ignored(path, false)
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
	return folderFiles, existingManaged, nil
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
