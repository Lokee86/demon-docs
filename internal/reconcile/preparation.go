package reconcile

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/config"
	md "github.com/Lokee86/demon-docs/internal/markdown"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/validationworkers"
)

type folderPreparationResult struct {
	updates []model.FileUpdate
	matched []*model.IndexEntry
}

func prepareFolderResults(count int, prepare func(index int) (folderPreparationResult, error)) ([]model.FileUpdate, map[*model.IndexEntry]bool, error) {
	results := make([]folderPreparationResult, count)
	errors := validationworkers.Run(count, func(index int) error {
		result, err := prepare(index)
		if err == nil {
			results[index] = result
		}
		return err
	})

	updates := []model.FileUpdate{}
	matched := map[*model.IndexEntry]bool{}
	for index, err := range errors {
		if err != nil {
			return updates, matched, err
		}
		updates = append(updates, results[index].updates...)
		for _, entry := range results[index].matched {
			matched[entry] = true
		}
	}
	return updates, matched, nil
}

type treePreparationContext struct {
	root         string
	config       config.Config
	title        func(string) string
	texts        map[string]string
	indexExists  map[string]bool
	documents    map[string]string
	entries      map[string][]*model.IndexEntry
	crossFiles   map[string][]*model.IndexEntry
	fileCounts   map[string]int
	crossFolders map[string][]*model.IndexEntry
	folderCounts map[string]int
}

func (context treePreparationContext) prepare(folder *model.FolderInfo) (folderPreparationResult, error) {
	result := folderPreparationResult{}
	localMatched := map[*model.IndexEntry]bool{}
	if folder.IndexPath != "" {
		current := context.texts[folder.Path]
		desired := md.DesiredParent(folder.IndexPath, context.root, context.title, context.config)
		newText, err := updateSections(
			folder,
			md.UpdateParent(current, desired, context.config.ParentLink.Label),
			context.entries[folder.Path],
			context.crossFiles,
			context.fileCounts,
			context.crossFolders,
			context.folderCounts,
			localMatched,
			context.config,
		)
		if err != nil {
			return folderPreparationResult{}, fmt.Errorf("reconcile index %s: %w", folder.IndexPath, err)
		}
		if !context.indexExists[folder.Path] || newText != current {
			var old *string
			if context.indexExists[folder.Path] {
				copy := current
				old = &copy
			}
			result.updates = append(result.updates, model.FileUpdate{Path: folder.IndexPath, OldText: old, NewText: newText})
		}
	}

	if !folder.IsStubs {
		for _, path := range append(append([]string{}, folder.DirectFiles...), folder.StubFiles...) {
			if !config.IsParentEditable(path, context.config) {
				continue
			}
			current, ok := context.documents[path]
			if !ok {
				continue
			}
			desired := md.DesiredParent(path, context.root, context.title, context.config)
			next := md.UpdateParent(current, desired, context.config.ParentLink.Label)
			if next != current {
				copy := current
				result.updates = append(result.updates, model.FileUpdate{Path: path, OldText: &copy, NewText: next})
			}
		}
	}

	result.matched = make([]*model.IndexEntry, 0, len(localMatched))
	for entry := range localMatched {
		result.matched = append(result.matched, entry)
	}
	return result, nil
}

func parentTitleSources(root string, folders []*model.FolderInfo, texts, documents map[string]string, c config.Config) []string {
	result := []string{}
	rootIndex := filepath.Join(root, c.IndexFile)
	for _, folder := range folders {
		if folder.IndexPath != "" {
			if source, ok := texts[folder.Path]; ok {
				if title := parentTitleForRoot(folder.IndexPath, source, rootIndex, c.ParentLink.Label); title != "" {
					result = append(result, title)
				}
			}
		}
		for _, path := range append(append([]string{}, folder.DirectFiles...), folder.StubFiles...) {
			if !config.IsParentEditable(path, c) {
				continue
			}
			if source, ok := documents[path]; ok {
				if title := parentTitleForRoot(path, source, rootIndex, c.ParentLink.Label); title != "" {
					result = append(result, title)
				}
			}
		}
	}
	return result
}
