package reverseindex

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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

const reverseWorkerLimit = 16

type inventoryFolderResult struct {
	files           []string
	existingManaged bool
}

type inventoryFolderPreparation func(repositoryRoot string, c config.Config, hierarchy *ignorepolicy.Hierarchy, folder string, f facts) (inventoryFolderResult, error)

func inventoryFolders(repositoryRoot string, c config.Config, hierarchy *ignorepolicy.Hierarchy, folders map[string]struct{}, f facts) (map[string][]string, map[string]struct{}, error) {
	return inventoryFoldersWithPreparation(repositoryRoot, c, hierarchy, folders, f, prepareInventoryFolder)
}

func inventoryFoldersWithPreparation(repositoryRoot string, c config.Config, hierarchy *ignorepolicy.Hierarchy, folders map[string]struct{}, f facts, prepare inventoryFolderPreparation) (map[string][]string, map[string]struct{}, error) {
	ordered := sortedFolders(folders)
	results := make([]inventoryFolderResult, len(ordered))
	errors := runReverseWorkers(len(ordered), func(index int) error {
		result, err := prepare(repositoryRoot, c, hierarchy, ordered[index], f)
		if err == nil {
			results[index] = result
		}
		return err
	})

	folderFiles := map[string][]string{}
	existingManaged := map[string]struct{}{}
	for index, err := range errors {
		if err != nil {
			return nil, nil, err
		}
		folder := ordered[index]
		if len(results[index].files) > 0 {
			folderFiles[folder] = results[index].files
		}
		if results[index].existingManaged {
			existingManaged[folder] = struct{}{}
		}
	}
	return folderFiles, existingManaged, nil
}

func prepareInventoryFolder(repositoryRoot string, c config.Config, hierarchy *ignorepolicy.Hierarchy, folder string, f facts) (inventoryFolderResult, error) {
	result := inventoryFolderResult{}
	indexPath := filepath.Join(folder, c.IndexFile)
	if doc, err := textio.Read(indexPath); err == nil && strings.Contains(doc.Text, markerStart(c)) {
		result.existingManaged = true
	} else if err != nil && !os.IsNotExist(err) {
		return inventoryFolderResult{}, err
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		return inventoryFolderResult{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() || entry.Name() == c.IndexFile || entry.Name() == ignorepolicy.FileName {
			continue
		}
		path := filepath.Join(folder, entry.Name())
		ignored, err := hierarchy.Ignored(path, false)
		if err != nil {
			return inventoryFolderResult{}, err
		}
		if ignored || !indexableCodeFile(repositoryRoot, path, f.exactFiles) {
			continue
		}
		result.files = append(result.files, path)
	}
	sort.Strings(result.files)
	return result, nil
}

func runReverseWorkers(count int, work func(index int) error) []error {
	if count == 0 {
		return nil
	}
	workerCount := count
	if workerCount > reverseWorkerLimit {
		workerCount = reverseWorkerLimit
	}

	errors := make([]error, count)
	jobs := make(chan int)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for worker := 0; worker < workerCount; worker++ {
		go func() {
			defer workers.Done()
			for index := range jobs {
				errors[index] = work(index)
			}
		}()
	}
	for index := 0; index < count; index++ {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	return errors
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

func orphanFiles(repositoryRoot string, folderFiles map[string][]string, f facts) []string {
	orphans := []string{}
	for _, files := range folderFiles {
		for _, path := range files {
			relative, err := filepath.Rel(repositoryRoot, path)
			if err != nil {
				continue
			}
			relative = filepath.ToSlash(filepath.Clean(relative))
			if len(f.fileDocs[relative]) == 0 {
				orphans = append(orphans, relative)
			}
		}
	}
	sort.Strings(orphans)
	return orphans
}
