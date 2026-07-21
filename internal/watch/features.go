package watch

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/fsnotify/fsnotify"
)

type Features struct {
	Indexes     bool
	Frontmatter bool
	Format      bool
	Links       bool
	TrackLinks  bool
	Reverse     bool
}

// validationBatchForEvent classifies the validation consequence of one
// relevant filesystem event. Link and index scheduling remains independent of
// this classification.
func validationBatchForEvent(event fsnotify.Event, c config.Config, policy ignorepolicy.Policy, docsRoot, repositoryRoot string, features Features, wasDirectory, isExternal bool) (string, bool, bool) {
	if !isExternal && !relevantSelectedWithPolicy(event.Name, c, policy, docsRoot, repositoryRoot, features, wasDirectory) {
		return "", false, false
	}
	schemaEvent := features.Format && formatSchemaEvent(event.Name, repositoryRoot, c, wasDirectory)
	if policy.IsControlFile(event.Name) || schemaEvent {
		return "", true, true
	}
	if isExternal || !repository.Contains(docsRoot, event.Name) {
		return "", false, true
	}
	removeOrRename := event.Op&(fsnotify.Remove|fsnotify.Rename) != 0
	if wasDirectory || removeOrRename {
		return "", true, true
	}
	if !strings.EqualFold(filepath.Ext(event.Name), ".md") {
		return "", false, true
	}
	ordinaryMarkdown := event.Op&(fsnotify.Create|fsnotify.Write) != 0
	if ordinaryMarkdown {
		info, err := os.Stat(event.Name)
		ordinaryMarkdown = err == nil && info.Mode().IsRegular()
	}
	if !ordinaryMarkdown {
		return "", true, true
	}
	return filepath.Clean(event.Name), false, true
}

func modelResult(updates []model.FileUpdate) model.ReconcileResult {
	return model.ReconcileResult{Updates: updates}
}

func relevantSelectedWithPolicy(path string, c config.Config, policy ignorepolicy.Policy, docsRoot, repositoryRoot string, features Features, wasDirectory bool) bool {
	if policy.IsControlFile(path) {
		return true
	}
	if features.Format && formatSchemaEvent(path, repositoryRoot, c, wasDirectory) {
		return true
	}
	if features.TrackLinks && repository.Contains(repositoryRoot, path) {
		ignored, err := policy.Ignored(path, wasDirectory)
		return err == nil && !ignored && !watchIgnored(path, c)
	}
	if (features.Frontmatter || features.Format) && repository.Contains(docsRoot, path) {
		ignored, err := policy.Ignored(path, wasDirectory)
		if err != nil || ignored || watchIgnored(path, c) {
			return false
		}
		if wasDirectory {
			return true
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return true
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			return true
		}
	}
	if (features.Indexes || features.Frontmatter || features.Format) && wasDirectory && repository.Contains(docsRoot, path) {
		ignored, err := policy.Ignored(path, true)
		return err == nil && !ignored && !watchIgnored(path, c)
	}
	if features.Indexes || features.Frontmatter || features.Format {
		return relevantWithPolicy(path, c, policy, docsRoot)
	}
	return false
}

func configuredFormatDirectories(repositoryRoot string, c config.Config) []string {
	seen := map[string]bool{}
	var result []string
	for _, configured := range []string{c.Format.SchemaDir, c.Format.DocumentSchemaDir} {
		configured = strings.TrimSpace(configured)
		if configured == "" {
			continue
		}
		directory := filepath.FromSlash(configured)
		if !filepath.IsAbs(directory) {
			directory = filepath.Join(repositoryRoot, directory)
		}
		directory = filepath.Clean(directory)
		if seen[directory] {
			continue
		}
		seen[directory] = true
		result = append(result, directory)
	}
	sort.Strings(result)
	return result
}

func formatSchemaEvent(path, repositoryRoot string, c config.Config, wasDirectory bool) bool {
	for _, directory := range configuredFormatDirectories(repositoryRoot, c) {
		if !repository.Contains(directory, path) {
			continue
		}
		if wasDirectory {
			return true
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return true
		}
		return strings.EqualFold(filepath.Ext(path), ".toml")
	}
	return false
}

func addFormatWatches(w eventWatcher, repositoryRoot string, c config.Config, watched map[string]bool) error {
	for _, configured := range configuredFormatDirectories(repositoryRoot, c) {
		directories := []string{nearestExistingDirectory(configured)}
		if info, err := os.Stat(configured); err == nil && info.IsDir() {
			directories = append(directories, nearestExistingDirectory(filepath.Dir(configured)))
		}
		for _, directory := range directories {
			if directory == "" || watched[directory] {
				continue
			}
			if err := w.Add(directory); err != nil {
				return err
			}
			watched[directory] = true
		}
	}
	return nil
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
