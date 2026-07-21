package frontmatter

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/textio"
	"github.com/Lokee86/demon-docs/internal/validationcache"
	"github.com/Lokee86/demon-docs/internal/validationworkers"
)

type plannedSource struct {
	path        string
	relative    string
	document    textio.Document
	parsed      Document
	parseErr    error
	contentHash string
	cacheHit    bool
	cacheEntry  validationcache.Entry
}

func loadSources(repoRoot string, files []string, allowedFormats []string, cfg config.Config, immutable immutableIndex, cache *validationcache.Store, schemaHasher *validationcache.SchemaHasher, selected map[string]bool) ([]plannedSource, map[string]bool, error) {
	sources := make([]plannedSource, len(files))
	policyHash := validationcache.FrontmatterPolicyHash(cfg)
	errors := validationworkers.Run(len(files), func(index int) error {
		path := files[index]
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relativePath := filepath.ToSlash(relative)
		if selected != nil && !selected[validationcache.NormalizePath(path)] {
			entry, ok := cache.LookupPath(relativePath)
			if !ok || !entry.FrontmatterClean || entry.FrontmatterPolicyHash != policyHash {
				return validationcache.ErrScopedReuseUnavailable
			}
			sources[index] = plannedSource{
				path: path, relative: relativePath, contentHash: entry.ContentSHA256,
				cacheHit: true, cacheEntry: entry,
				parsed: Document{Values: map[string]any{"document_id": entry.DocumentID, "document_type": entry.DocumentType}, HasBlock: true},
			}
			return nil
		}
		document, err := textio.Read(path)
		if err != nil {
			return fmt.Errorf("read frontmatter source %s: %w", path, err)
		}
		source := plannedSource{
			path:        path,
			relative:    relativePath,
			document:    document,
			contentHash: validationcache.ContentHash(document.RawBytes()),
		}
		if candidate, ok := cache.Candidate(source.relative, source.contentHash, policyHash); ok && candidate.FrontmatterClean {
			schemaHash := schemaHasher.Effective(candidate.SchemaName, candidate.DocumentID)
			recorded := immutable.values(source.relative, map[string]any{"document_id": candidate.DocumentID}, true)
			immutableHash := validationcache.Hash(recorded)
			if entry, valid := cache.Lookup(source.relative, source.contentHash, policyHash, schemaHash, immutableHash); valid {
				source.cacheHit = true
				source.cacheEntry = entry
				source.parsed = Document{Values: map[string]any{"document_id": entry.DocumentID, "document_type": entry.DocumentType}, HasBlock: true}
			}
		}
		if !source.cacheHit {
			source.parsed, source.parseErr = Parse(document.Text, allowedFormats)
		}
		sources[index] = source
		return nil
	})
	for _, err := range errors {
		if err != nil {
			return nil, nil, err
		}
	}

	pathsByID := make(map[string][]string)
	for _, source := range sources {
		if id := sourceDocumentID(source); id != "" {
			pathsByID[id] = append(pathsByID[id], source.path)
		}
	}
	duplicates := make(map[string]bool)
	duplicateIDs := make(map[string]bool)
	for id, paths := range pathsByID {
		if len(paths) < 2 {
			continue
		}
		duplicateIDs[id] = true
		for _, path := range paths {
			duplicates[path] = true
		}
	}
	if selected != nil && len(duplicateIDs) > 0 {
		for _, source := range sources {
			if duplicateIDs[sourceDocumentID(source)] && !selected[validationcache.NormalizePath(source.path)] {
				return nil, nil, validationcache.ErrScopedReuseUnavailable
			}
		}
	}
	duplicateSources := make([]int, 0)
	for index := range sources {
		if selected == nil && sources[index].cacheHit && duplicateIDs[sourceDocumentID(sources[index])] {
			duplicateSources = append(duplicateSources, index)
		}
	}
	validationworkers.Run(len(duplicateSources), func(job int) error {
		index := duplicateSources[job]
		sources[index].parsed, sources[index].parseErr = Parse(sources[index].document.Text, allowedFormats)
		sources[index].cacheHit = false
		return nil
	})
	return sources, duplicates, nil
}

func scopedPathSet(repoRoot string, paths []string) map[string]bool {
	if paths == nil {
		return nil
	}
	selected := make(map[string]bool, len(paths))
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(repoRoot, path)
		}
		selected[validationcache.NormalizePath(path)] = true
	}
	return selected
}

func sourceDocumentID(source plannedSource) string {
	if source.cacheHit {
		return source.cacheEntry.DocumentID
	}
	if source.parseErr != nil {
		return ""
	}
	return documentID(source.parsed.Values)
}
