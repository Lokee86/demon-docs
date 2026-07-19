package watch

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
)

type Features struct {
	Indexes bool
	Links   bool
	Reverse bool
}

func modelResult(updates []model.FileUpdate) model.ReconcileResult {
	return model.ReconcileResult{Updates: updates}
}

func relevantSelectedWithPolicy(path string, c config.Config, policy ignorepolicy.Policy, docsRoot, repositoryRoot string, features Features, wasDirectory bool) bool {
	if policy.IsControlFile(path) {
		return true
	}
	if features.Links && repository.Contains(repositoryRoot, path) {
		ignored, err := policy.Ignored(path, wasDirectory)
		return err == nil && !ignored && !watchIgnored(path, c)
	}
	if features.Indexes && wasDirectory && repository.Contains(docsRoot, path) {
		ignored, err := policy.Ignored(path, true)
		return err == nil && !ignored && !watchIgnored(path, c)
	}
	if features.Indexes {
		return relevantWithPolicy(path, c, policy, docsRoot)
	}
	return false
}

func externalWatchDirectories(manifest links.FilesManifest) []string {
	seen := map[string]bool{}
	var result []string
	for _, record := range manifest.Files {
		if record.Scope != "external" {
			continue
		}
		directory := nearestExistingDirectory(filepath.Dir(filepath.FromSlash(record.Path)))
		if directory == "" || seen[directory] {
			continue
		}
		seen[directory] = true
		result = append(result, directory)
	}
	sort.Strings(result)
	return result
}

func nearestExistingDirectory(path string) string {
	current := filepath.Clean(path)
	for {
		if info, err := os.Stat(current); err == nil && info.IsDir() {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func addExternalWatches(w eventWatcher, directories []string, watched map[string]bool) error {
	for _, directory := range directories {
		if watched[directory] {
			continue
		}
		if err := w.Add(directory); err != nil {
			return err
		}
		watched[directory] = true
	}
	return nil
}

func externalEvent(path string, watched map[string]bool) bool {
	for directory := range watched {
		if repository.Contains(directory, path) {
			return true
		}
	}
	return false
}
