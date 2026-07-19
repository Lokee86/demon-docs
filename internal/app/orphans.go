package app

import (
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/scan"
)

func findOrphanDocuments(scope repository.Scope, c config.Config, plan links.Plan) ([]string, error) {
	tree, err := scan.TreeWithIgnoreRoot(scope.DocsRoot, scope.RepositoryRoot, c)
	if err != nil {
		return nil, err
	}

	filesByPath := make(map[string]links.FileRecord, len(plan.Files.Files))
	excludedSources := make(map[string]struct{})
	for _, record := range plan.Files.Files {
		if record.Scope != "repository" || !record.Present || record.Kind != "file" {
			continue
		}
		filesByPath[orphanPathKey(record.Path)] = record
		absolute := filepath.Join(scope.RepositoryRoot, filepath.FromSlash(record.Path))
		if strings.EqualFold(filepath.Base(absolute), c.IndexFile) || isDraftDocument(scope.DocsRoot, absolute, c.Draft.Folder) {
			excludedSources[record.ID] = struct{}{}
		}
	}

	candidates := make(map[string]string)
	for _, folder := range tree.Folders {
		for _, path := range folder.DirectFiles {
			if !isOrphanMarkdown(path) || isDraftDocument(scope.DocsRoot, path, c.Draft.Folder) {
				continue
			}
			if record, ok := fileRecordForPath(scope.RepositoryRoot, path, filesByPath); ok {
				candidates[record.ID] = record.Path
			}
		}
	}

	hasInbound := make(map[string]bool, len(candidates))
	for _, record := range plan.Links.Links {
		if record.TargetFileID == "" || record.SourceFileID == record.TargetFileID {
			continue
		}
		if _, excluded := excludedSources[record.SourceFileID]; excluded {
			continue
		}
		if _, managed := candidates[record.TargetFileID]; managed {
			hasInbound[record.TargetFileID] = true
		}
	}

	orphans := make([]string, 0)
	for id, path := range candidates {
		if !hasInbound[id] {
			orphans = append(orphans, path)
		}
	}
	sort.Strings(orphans)
	return orphans, nil
}

func fileRecordForPath(repositoryRoot, path string, filesByPath map[string]links.FileRecord) (links.FileRecord, bool) {
	relative, err := filepath.Rel(repositoryRoot, path)
	if err != nil {
		return links.FileRecord{}, false
	}
	record, ok := filesByPath[orphanPathKey(filepath.ToSlash(relative))]
	return record, ok
}

func orphanPathKey(path string) string {
	key := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return key
}

func isOrphanMarkdown(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown", ".mkd", ".mdx":
		return true
	default:
		return false
	}
}

func isDraftDocument(docsRoot, path, draftFolder string) bool {
	if strings.TrimSpace(draftFolder) == "" {
		return false
	}
	relative, err := filepath.Rel(docsRoot, path)
	if err != nil {
		return false
	}
	clean := filepath.Clean(relative)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		if strings.EqualFold(part, draftFolder) {
			return true
		}
	}
	return false
}
