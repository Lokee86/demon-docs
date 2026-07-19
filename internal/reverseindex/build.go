package reverseindex

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/config"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func Build(repositoryRoot, docsRoot string, c config.Config, format codemap.Format) (Plan, error) {
	repositoryRoot, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	docsRoot, err = filepath.Abs(docsRoot)
	if err != nil {
		return Plan{}, err
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
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
	for _, diagnostic := range dataset.Diagnostics {
		plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s:%d: %s", diagnostic.DocumentPath, diagnostic.Source.Line, diagnostic.Message))
	}
	for _, item := range dataset.Entries {
		paths := resolvedPaths(item.Resolution)
		if len(paths) == 0 {
			if item.Resolution.Status != codemap.ResolutionUnsupported {
				plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s:%d: %s target %s", item.Entry.DocumentPath, item.Entry.Source.Line, item.Resolution.Status, item.Entry.Target))
			}
			continue
		}
		for _, relative := range paths {
			accepted, targetErr := collected.addTarget(repositoryRoot, docsRoot, relative, item.Entry.DocumentPath, policy)
			if targetErr != nil {
				plan.Diagnostics = append(plan.Diagnostics, fmt.Sprintf("%s:%d: %s", item.Entry.DocumentPath, item.Entry.Source.Line, targetErr))
				continue
			}
			if accepted {
				plan.ReferenceCount++
			}
		}
	}

	folderFiles, existingManaged, err := inventoryFolders(repositoryRoot, docsRoot, c, policy, collected)
	if err != nil {
		return Plan{}, err
	}
	folders := map[string]struct{}{}
	for folder := range collected.eligibleFolder {
		folders[folder] = struct{}{}
	}
	for folder := range existingManaged {
		folders[folder] = struct{}{}
	}
	ordered := sortedFolders(folders)
	for _, folder := range ordered {
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
