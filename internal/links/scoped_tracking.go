package links

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/review"
)

// TrackSources refreshes link state for the supplied repository Markdown
// sources without rescanning or reparsing the rest of the repository. The
// existing file and link manifests are retained for every other source.
//
// This is intentionally tracking-only: it does not plan generated rewrites,
// append review history, or create watcher suppressions.
func TrackSources(repositoryRoot string, sourcePaths []string) (Plan, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	root = filepath.Clean(root)
	paths, err := scopedSourcePaths(root, sourcePaths)
	if err != nil {
		return Plan{}, err
	}

	files, previousLinks, initialized, err := loadState(root)
	if err != nil {
		return Plan{}, err
	}
	if !initialized {
		return Plan{RepositoryRoot: root, NeedsInitialization: true}, nil
	}
	files, previousLinks = pruneNestedWorktreeState(root, files, previousLinks)
	policy, err := ignorepolicy.Load(root)
	if err != nil {
		return Plan{}, err
	}
	suppressions, err := LoadPendingSuppressions(root)
	if err != nil {
		return Plan{}, err
	}

	inventory := &inventory{
		root:     root,
		policy:   policy,
		manifest: files,
	}
	inventory.rebuild()
	previousBySource := previousLinkIndex(previousLinks)
	selected := map[string]bool{}
	for _, path := range paths {
		stored := storePath(root, path)
		if record := inventory.fileByPath(stored); record != nil {
			selected[record.ID] = true
		}
	}

	plan := Plan{
		RepositoryRoot:      root,
		Initialized:         initialized,
		NeedsInitialization: !initialized,
		Files:               inventory.manifest,
		Links:               LinksManifest{SchemaVersion: schemaVersion},
		Suppressions:        suppressions,
	}
	for _, record := range previousLinks.Links {
		if !selected[record.SourceFileID] {
			plan.Links.Links = append(plan.Links.Links, record)
		}
	}

	for _, path := range paths {
		record, err := inventory.refreshScopedFile(path)
		if err != nil {
			return Plan{}, err
		}
		if record == nil || record.Kind != "file" || !record.Present || !isMarkdown(record.Path) {
			continue
		}
		previousRecords := previousBySource[record.ID]
		source := markdownSource{path: path, record: record}
		if err := reconcileMarkdownSource(&plan, inventory, source, previousRecords, initialized, false, review.Policy{}); err != nil {
			return Plan{}, err
		}
	}

	plan.Files = inventory.manifest
	sortManifests(&plan.Files, &plan.Links)
	return plan, nil
}

func scopedSourcePaths(root string, sourcePaths []string) ([]string, error) {
	seen := map[string]bool{}
	paths := make([]string, 0, len(sourcePaths))
	for _, sourcePath := range sourcePaths {
		path, err := filepath.Abs(sourcePath)
		if err != nil {
			return nil, err
		}
		path = filepath.Clean(path)
		if !repository.Contains(root, path) {
			return nil, fmt.Errorf("refusing to track source outside repository root: %s", path)
		}
		key := pathKey(path)
		if seen[key] {
			continue
		}
		seen[key] = true
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool { return pathKey(paths[i]) < pathKey(paths[j]) })
	return paths, nil
}

func (i *inventory) fileByPath(stored string) *FileRecord {
	for index := range i.manifest.Files {
		record := &i.manifest.Files[index]
		if record.Scope == "repository" && pathKey(record.Path) == pathKey(stored) {
			return record
		}
	}
	return nil
}

func (i *inventory) refreshScopedFile(path string) (*FileRecord, error) {
	stored := storePath(i.root, path)
	record := i.fileByPath(stored)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if record != nil {
			record.Present = false
		}
		return record, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat tracked source %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, nil
	}
	if record == nil {
		id, err := newFileID()
		if err != nil {
			return nil, fmt.Errorf("create scoped file identity %s: %w", stored, err)
		}
		record = &FileRecord{ID: id, Scope: "repository", Path: stored}
		i.manifest.Files = append(i.manifest.Files, *record)
		i.rebuild()
		record = i.fileByPath(stored)
	}
	record.Kind = "file"
	record.Present = true
	record.Size = info.Size()
	record.ModifiedUnixNano = info.ModTime().UnixNano()
	record.Fingerprint, err = fileFingerprint(path)
	if err != nil {
		return nil, fmt.Errorf("fingerprint tracked source %s: %w", path, err)
	}
	record.DocumentID = ""
	if isMarkdown(stored) {
		record.DocumentID = markdownDocumentID(path)
	}
	return record, nil
}
