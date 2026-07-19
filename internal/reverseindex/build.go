package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func Build(repositoryRoot, docsRoot string, roots []string, c config.Config, format codemap.Format) (Plan, error) {
	repositoryRoot, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	docsRoot, err = filepath.Abs(docsRoot)
	if err != nil {
		return Plan{}, err
	}
	if len(roots) == 0 {
		return Plan{}, fmt.Errorf("no reverse-index roots selected")
	}
	hierarchy, folders, err := discoverScopeFolders(repositoryRoot, roots)
	if err != nil {
		return Plan{}, err
	}
	dataset, err := codemap.BuildDataset(repositoryRoot, docsRoot, format)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{}
	collected := newFacts()
	for _, document := range dataset.Documents {
		collected.titles[document.Path] = documentTitle(repositoryRoot, document.Path)
	}
	for _, item := range dataset.Entries {
		paths := resolvedPaths(item.Resolution)
		if len(paths) == 0 {
			if item.Resolution.Status != codemap.ResolutionUnsupported && entryPotentiallyInScope(repositoryRoot, roots, item.Entry, format) {
				plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s:%d: %s target %s", item.Entry.DocumentPath, item.Entry.Source.Line, item.Resolution.Status, item.Entry.Target))
			}
			continue
		}
		for _, relative := range paths {
			accepted, targetErr := collected.addTarget(repositoryRoot, roots, folders, hierarchy, relative, item.Entry.DocumentPath)
			if targetErr != nil {
				plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s:%d: %s", item.Entry.DocumentPath, item.Entry.Source.Line, targetErr))
				continue
			}
			if accepted {
				plan.ReferenceCount++
			}
		}
	}

	folderFiles, existingManaged, err := inventoryFolders(repositoryRoot, c, hierarchy, folders, collected)
	if err != nil {
		return Plan{}, err
	}
	selected := map[string]struct{}{}
	for folder, files := range folderFiles {
		if len(files) > 0 {
			selected[folder] = struct{}{}
		}
	}
	for relative := range collected.folderDocs {
		selected[filepath.Join(repositoryRoot, filepath.FromSlash(relative))] = struct{}{}
	}
	for folder := range existingManaged {
		selected[folder] = struct{}{}
	}

	for _, folder := range sortedFolders(selected) {
		indexPath := filepath.Join(folder, c.IndexFile)
		block := renderBlock(repositoryRoot, indexPath, folder, folderFiles[folder], collected, c)
		update, changed, reconcileErr := reconcileIndex(indexPath, folder, block, c)
		if reconcileErr != nil {
			return Plan{}, reconcileErr
		}
		if changed {
			plan.Updates = append(plan.Updates, update)
		}
		plan.IndexCount++
	}
	sort.Strings(plan.Diagnostics)
	return plan, nil
}

func reconcileIndex(indexPath, folder, block string, c config.Config) (model.FileUpdate, bool, error) {
	current := ""
	doc, readErr := textio.Read(indexPath)
	if readErr == nil {
		current = doc.Text
	} else if !os.IsNotExist(readErr) {
		return model.FileUpdate{}, false, fmt.Errorf("read reverse index %s: %w", indexPath, readErr)
	}
	next, err := replaceManaged(current, block, folder, c)
	if err != nil {
		return model.FileUpdate{}, false, fmt.Errorf("reconcile reverse index %s: %w", indexPath, err)
	}
	if next == current {
		return model.FileUpdate{}, false, nil
	}
	var old *string
	if readErr == nil {
		copy := current
		old = &copy
	}
	return model.FileUpdate{Path: indexPath, OldText: old, NewText: next}, true, nil
}

func sortedFolders(folders map[string]struct{}) []string {
	ordered := make([]string, 0, len(folders))
	for folder := range folders {
		ordered = append(ordered, folder)
	}
	sort.Strings(ordered)
	return ordered
}
