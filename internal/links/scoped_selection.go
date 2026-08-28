package links

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func normalizeChangedPaths(paths []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		absolute = filepath.Clean(absolute)
		key := pathKey(absolute)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, absolute)
	}
	sort.Slice(result, func(i, j int) bool { return pathKey(result[i]) < pathKey(result[j]) })
	return result, nil
}

func scopedBatchContainsDirectory(root string, paths []string, previous FilesManifest) bool {
	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return true
		}
		for _, record := range previous.Files {
			if record.Kind == "directory" && pathKey(recordAbsolute(root, record)) == pathKey(path) {
				return true
			}
		}
	}
	return false
}

func affectedSourceIDs(root string, changedPaths []string, previousFiles FilesManifest, previousLinks LinksManifest, currentFiles FilesManifest) map[string]bool {
	changed := map[string]bool{}
	changedNames := map[string]bool{}
	for _, path := range changedPaths {
		changed[pathKey(path)] = true
		changedNames[potentialBasename(path)] = true
	}

	previousByID := fileRecordIndex(previousFiles)
	currentByID := fileRecordIndex(currentFiles)
	changedFileIDs := map[string]bool{}
	for _, manifest := range []FilesManifest{previousFiles, currentFiles} {
		for _, record := range manifest.Files {
			if changed[pathKey(recordAbsolute(root, record))] {
				changedFileIDs[record.ID] = true
			}
			for _, historical := range record.PathHistory {
				if changed[pathKey(recordAbsolute(root, FileRecord{Path: historical, Scope: record.Scope}))] {
					changedFileIDs[record.ID] = true
				}
			}
		}
	}

	affected := map[string]bool{}
	for id := range changedFileIDs {
		if sourceRecordIsMarkdown(previousByID[id]) || sourceRecordIsMarkdown(currentByID[id]) {
			affected[id] = true
		}
	}
	for _, link := range previousLinks.Links {
		if changedFileIDs[link.TargetFileID] || changedStoredPath(root, link.ResolvedPath, changed) || anyChangedCandidate(root, link.Candidates, changed) {
			affected[link.SourceFileID] = true
			continue
		}
		if link.TargetFileID == "" || link.Status == "broken" || link.Status == "ambiguous" {
			if changedNames[potentialBasename(link.RawPath)] {
				affected[link.SourceFileID] = true
			}
		}
	}
	return affected
}

func sourceRecordIsMarkdown(record *FileRecord) bool {
	return record != nil && record.Kind == "file" && isMarkdown(record.Path)
}

func changedStoredPath(root, stored string, changed map[string]bool) bool {
	if stored == "" {
		return false
	}
	return changed[pathKey(recordAbsolute(root, FileRecord{Path: stored, Scope: scopeForStoredPath(stored)}))]
}

func anyChangedCandidate(root string, candidates []string, changed map[string]bool) bool {
	for _, candidate := range candidates {
		if changedStoredPath(root, candidate, changed) {
			return true
		}
	}
	return false
}

func scopeForStoredPath(path string) string {
	if filepath.IsAbs(filepath.FromSlash(path)) {
		return "external"
	}
	return "repository"
}

func potentialBasename(path string) string {
	name := strings.ToLower(filepath.Base(filepath.FromSlash(path)))
	if strings.EqualFold(filepath.Ext(name), ".md") {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return name
}

func scopedBatchCoversInventoryChanges(root string, paths []string, previous, current FilesManifest) bool {
	changed := map[string]bool{}
	for _, path := range paths {
		changed[pathKey(path)] = true
	}
	previousByID := fileRecordIndex(previous)
	currentByID := fileRecordIndex(current)
	ids := map[string]bool{}
	for id := range previousByID {
		ids[id] = true
	}
	for id := range currentByID {
		ids[id] = true
	}
	for id := range ids {
		before := previousByID[id]
		after := currentByID[id]
		if before == nil {
			if after != nil && after.Present && !changed[pathKey(recordAbsolute(root, *after))] {
				return false
			}
			continue
		}
		if after == nil {
			if before.Present && !changed[pathKey(recordAbsolute(root, *before))] {
				return false
			}
			continue
		}
		beforePath := recordAbsolute(root, *before)
		afterPath := recordAbsolute(root, *after)
		pathChanged := before.Present != after.Present || before.Scope != after.Scope || before.Kind != after.Kind || pathKey(beforePath) != pathKey(afterPath)
		contentChanged := before.Fingerprint != after.Fingerprint || before.DocumentID != after.DocumentID
		if pathChanged {
			if before.Present && !changed[pathKey(beforePath)] {
				return false
			}
			if after.Present && !changed[pathKey(afterPath)] {
				return false
			}
			continue
		}
		if contentChanged && after.Present && !changed[pathKey(afterPath)] {
			return false
		}
	}
	return true
}
