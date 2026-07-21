package links

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/repository"
)

type inventoryNode struct {
	abs, stored, kind, fingerprint, documentID string
	size, modified                             int64
}

type inventory struct {
	root     string
	policy   ignorepolicy.Policy
	manifest FilesManifest
	byAbs    map[string]int
	byFold   map[string][]int
	dirs     []string
}

func buildInventory(root string, previous FilesManifest) (*inventory, error) {
	policy, err := ignorepolicy.Load(root)
	if err != nil {
		return nil, err
	}
	previousByPath := map[string]int{}
	for index, record := range previous.Files {
		if record.Scope == "repository" {
			previousByPath[pathKey(record.Path)] = index
		}
	}
	var nodes []inventoryNode
	var dirs []string
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root {
			if entry.IsDir() && isNestedWorktreeDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			if entry.Type()&os.ModeSymlink != 0 {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
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
		if entry.IsDir() {
			dirs = append(dirs, filepath.Clean(path))
			nodes = append(nodes, inventoryNode{abs: filepath.Clean(path), stored: storePath(root, path), kind: "directory"})
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		stored := storePath(root, path)
		modified := info.ModTime().UnixNano()
		fingerprint := ""
		documentID := ""
		if previousIndex, ok := previousByPath[pathKey(stored)]; ok {
			previousRecord := previous.Files[previousIndex]
			if previousRecord.Present && previousRecord.Kind == "file" && previousRecord.Size == info.Size() && previousRecord.ModifiedUnixNano == modified {
				fingerprint = previousRecord.Fingerprint
				documentID = previousRecord.DocumentID
			}
		}
		if fingerprint == "" {
			fingerprint, err = fileFingerprint(path)
			if err != nil {
				return err
			}
		}
		if documentID == "" && isMarkdown(stored) {
			documentID = markdownDocumentID(path)
		}
		nodes = append(nodes, inventoryNode{abs: filepath.Clean(path), stored: stored, kind: "file", fingerprint: fingerprint, documentID: documentID, size: info.Size(), modified: modified})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan repository for link targets: %w", err)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].stored < nodes[j].stored })
	currentCounts := map[string]int{}
	currentDocumentCounts := map[string]int{}
	for _, node := range nodes {
		if node.fingerprint != "" {
			currentCounts[node.fingerprint]++
		}
		if node.documentID != "" {
			currentDocumentCounts[node.documentID]++
		}
	}
	previousByFingerprint := map[string][]int{}
	previousByDocumentID := map[string][]int{}
	for index, record := range previous.Files {
		if record.Scope == "repository" {
			previousByPath[pathKey(record.Path)] = index
		}
		if record.Kind == "file" && record.Fingerprint != "" {
			previousByFingerprint[record.Fingerprint] = append(previousByFingerprint[record.Fingerprint], index)
		}
		if record.DocumentID != "" {
			previousByDocumentID[record.DocumentID] = append(previousByDocumentID[record.DocumentID], index)
		}
	}
	used := map[int]bool{}
	matched := make([]int, len(nodes))
	for index := range matched {
		matched[index] = -1
	}
	for index, node := range nodes {
		if previousIndex, ok := previousByPath[pathKey(node.stored)]; ok && !used[previousIndex] {
			previousRecord := previous.Files[previousIndex]
			if node.documentID != "" && previousRecord.DocumentID != "" && node.documentID != previousRecord.DocumentID {
				continue
			}
			matched[index] = previousIndex
			used[previousIndex] = true
		}
	}
	for index, node := range nodes {
		if matched[index] >= 0 || node.documentID == "" || currentDocumentCounts[node.documentID] != 1 {
			continue
		}
		candidates := previousByDocumentID[node.documentID]
		available := -1
		for _, candidate := range candidates {
			if !used[candidate] {
				if available >= 0 {
					available = -1
					break
				}
				available = candidate
			}
		}
		if available >= 0 {
			matched[index] = available
			used[available] = true
		}
	}
	for index, node := range nodes {
		if matched[index] >= 0 || node.fingerprint == "" || currentCounts[node.fingerprint] != 1 {
			continue
		}
		candidates := previousByFingerprint[node.fingerprint]
		available := -1
		for _, candidate := range candidates {
			if !used[candidate] {
				if available >= 0 {
					available = -1
					break
				}
				available = candidate
			}
		}
		if available >= 0 {
			matched[index] = available
			used[available] = true
		}
	}
	manifest := FilesManifest{SchemaVersion: schemaVersion}
	for index, node := range nodes {
		var record FileRecord
		if matched[index] >= 0 {
			record = previous.Files[matched[index]]
			if record.Path != node.stored {
				record.PathHistory = appendUnique(record.PathHistory, record.Path)
			}
		} else {
			id, err := newFileID()
			if err != nil {
				return nil, fmt.Errorf("create file identity: %w", err)
			}
			record.ID = id
		}
		record.DocumentID = node.documentID
		record.Path = node.stored
		record.Scope = "repository"
		record.Kind = node.kind
		record.Present = true
		record.Fingerprint = node.fingerprint
		record.Size = node.size
		record.ModifiedUnixNano = node.modified
		manifest.Files = append(manifest.Files, record)
	}
	for index, record := range previous.Files {
		if used[index] {
			continue
		}
		if record.Scope == "external" {
			abs := filepath.FromSlash(record.Path)
			if info, err := os.Stat(abs); err == nil {
				record.Present = true
				record.Kind = kindFromInfo(info)
				record.Size = info.Size()
				record.ModifiedUnixNano = info.ModTime().UnixNano()
				if info.Mode().IsRegular() {
					record.Fingerprint, _ = fileFingerprint(abs)
				}
			} else {
				record.Present = false
			}
		} else {
			record.Present = false
		}
		manifest.Files = append(manifest.Files, record)
	}
	result := &inventory{root: filepath.Clean(root), policy: policy, manifest: manifest, dirs: dirs}
	result.rebuild()
	return result, nil
}

func isNestedWorktreeDirectory(name string) bool {
	return name == ".worktrees" || name == ".workingtrees"
}

func isNestedWorktreePath(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == "." || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	for _, segment := range strings.Split(filepath.Clean(relative), string(filepath.Separator)) {
		if isNestedWorktreeDirectory(segment) {
			return true
		}
	}
	return false
}

func pruneNestedWorktreeState(root string, files FilesManifest, links LinksManifest) (FilesManifest, LinksManifest) {
	excludedSources := map[string]bool{}
	keptFiles := FilesManifest{SchemaVersion: files.SchemaVersion}
	for _, record := range files.Files {
		if record.Scope == "repository" && isNestedWorktreePath(root, filepath.Join(root, filepath.FromSlash(record.Path))) {
			excludedSources[record.ID] = true
			continue
		}
		keptFiles.Files = append(keptFiles.Files, record)
	}
	keptLinks := LinksManifest{SchemaVersion: links.SchemaVersion}
	for _, record := range links.Links {
		if excludedSources[record.SourceFileID] {
			continue
		}
		keptLinks.Links = append(keptLinks.Links, record)
	}
	return keptFiles, keptLinks
}

func (i *inventory) rebuild() {
	i.byAbs = map[string]int{}
	i.byFold = map[string][]int{}
	for index, record := range i.manifest.Files {
		abs := recordAbsolute(i.root, record)
		key := pathKey(abs)
		if existing, ok := i.byAbs[key]; !ok || (!i.manifest.Files[existing].Present && record.Present) {
			i.byAbs[key] = index
		}
		fold := strings.ToLower(filepath.Clean(abs))
		i.byFold[fold] = append(i.byFold[fold], index)
	}
}

func (i *inventory) ignored(path string) (bool, error) {
	if !repository.Contains(i.root, path) {
		return false, nil
	}
	if isNestedWorktreePath(i.root, path) {
		return true, nil
	}
	isDirectory := false
	if info, err := os.Stat(path); err == nil {
		isDirectory = info.IsDir()
	}
	return i.policy.Ignored(path, isDirectory)
}

func (i *inventory) exact(path string) (*FileRecord, string) {
	clean := filepath.Clean(path)
	if index, ok := i.byAbs[pathKey(clean)]; ok && i.manifest.Files[index].Present {
		return &i.manifest.Files[index], recordAbsolute(i.root, i.manifest.Files[index])
	}
	candidates := i.byFold[strings.ToLower(clean)]
	if len(candidates) == 1 && i.manifest.Files[candidates[0]].Present {
		return &i.manifest.Files[candidates[0]], recordAbsolute(i.root, i.manifest.Files[candidates[0]])
	}
	return nil, ""
}

func (i *inventory) byID(id string) (*FileRecord, string) {
	for index := range i.manifest.Files {
		if i.manifest.Files[index].ID == id && i.manifest.Files[index].Present {
			return &i.manifest.Files[index], recordAbsolute(i.root, i.manifest.Files[index])
		}
	}
	return nil, ""
}

func (i *inventory) recordByID(id string) *FileRecord {
	for index := range i.manifest.Files {
		if i.manifest.Files[index].ID == id {
			return &i.manifest.Files[index]
		}
	}
	return nil
}

func (i *inventory) ensureTarget(path, preferredID string) (*FileRecord, string, error) {
	clean := filepath.Clean(path)
	if record, actual := i.exact(clean); record != nil {
		return record, actual, nil
	}
	info, err := os.Stat(clean)
	if err != nil {
		return nil, "", err
	}
	record := FileRecord{Path: storePath(i.root, clean), Scope: scopeFor(i.root, clean), Kind: kindFromInfo(info), Present: true, Size: info.Size()}
	if info.Mode().IsRegular() {
		record.Fingerprint, err = fileFingerprint(clean)
		if err != nil {
			return nil, "", err
		}
		if isMarkdown(record.Path) {
			record.DocumentID = markdownDocumentID(clean)
		}
	}
	if preferredID != "" {
		if previous := i.recordByID(preferredID); previous != nil {
			oldPath := previous.Path
			history := append([]string(nil), previous.PathHistory...)
			*previous = record
			previous.ID = preferredID
			previous.PathHistory = appendUnique(history, oldPath)
			i.rebuild()
			resolved, _ := i.exact(clean)
			return resolved, clean, nil
		}
	}
	id, err := newFileID()
	if err != nil {
		return nil, "", err
	}
	record.ID = id
	i.manifest.Files = append(i.manifest.Files, record)
	i.rebuild()
	resolved, actual := i.exact(clean)
	return resolved, actual, nil
}

func (i *inventory) candidates(base, kind string) []string {
	seen := map[string]bool{}
	var result []string
	for _, record := range i.manifest.Files {
		if !record.Present || record.Kind != kind {
			continue
		}
		abs := recordAbsolute(i.root, record)
		if strings.EqualFold(filepath.Base(abs), base) && !seen[pathKey(abs)] {
			seen[pathKey(abs)] = true
			result = append(result, abs)
		}
	}
	if kind == "directory" {
		for _, path := range i.dirs {
			if strings.EqualFold(filepath.Base(path), base) && !seen[pathKey(path)] {
				seen[pathKey(path)] = true
				result = append(result, path)
			}
		}
	}
	sort.Slice(result, func(a, b int) bool { return pathKey(result[a]) < pathKey(result[b]) })
	return result
}
