package links

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/review"
)

var ErrScopedReconciliationUnavailable = errors.New("scoped link reconciliation unavailable")

// ReconcileChangedPaths repairs only link sources affected by a complete batch
// of changed file paths. Inventory construction remains repository-authoritative
// so move identity and ambiguity decisions use the same evidence as a full pass.
func ReconcileChangedPaths(repositoryRoot string, changedPaths []string) (Plan, error) {
	return reconcileChangedPaths(repositoryRoot, changedPaths, true)
}

// TrackChangedPaths refreshes only link sources affected by a complete batch of
// changed file paths without planning user-file rewrites.
func TrackChangedPaths(repositoryRoot string, changedPaths []string) (Plan, error) {
	return reconcileChangedPaths(repositoryRoot, changedPaths, false)
}

func reconcileChangedPaths(repositoryRoot string, changedPaths []string, repair bool) (Plan, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	root = filepath.Clean(root)
	paths, err := normalizeChangedPaths(changedPaths)
	if err != nil {
		return Plan{}, err
	}
	if len(paths) == 0 {
		return Plan{}, ErrScopedReconciliationUnavailable
	}

	previousFiles, previousLinks, initialized, err := loadState(root)
	if err != nil {
		return Plan{}, err
	}
	if !initialized {
		return Plan{}, ErrScopedReconciliationUnavailable
	}
	previousFiles, previousLinks = pruneNestedWorktreeState(root, previousFiles, previousLinks)
	if scopedBatchContainsDirectory(root, paths, previousFiles) {
		return Plan{}, ErrScopedReconciliationUnavailable
	}

	inventory, err := buildInventory(root, previousFiles)
	if err != nil {
		return Plan{}, err
	}
	if !scopedBatchCoversInventoryChanges(root, paths, previousFiles, inventory.manifest) {
		return Plan{}, ErrScopedReconciliationUnavailable
	}
	previousFiles, previousLinks = collapseDocumentIdentityAliases(previousFiles, previousLinks, &inventory.manifest)
	inventory.rebuild()

	affected := affectedSourceIDs(root, paths, previousFiles, previousLinks, inventory.manifest)
	previousBySource := previousLinkIndex(previousLinks)
	currentByID := fileRecordIndex(inventory.manifest)
	policy, err := review.LoadPolicy(root)
	if err != nil {
		return Plan{}, fmt.Errorf("load review policy: %w", err)
	}

	plan := Plan{
		RepositoryRoot: root,
		Initialized:    true,
		Files:          inventory.manifest,
		Links:          LinksManifest{SchemaVersion: schemaVersion},
	}
	for _, record := range previousLinks.Links {
		if !affected[record.SourceFileID] {
			plan.Links.Links = append(plan.Links.Links, record)
		}
	}

	ids := make([]string, 0, len(affected))
	for id := range affected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		record := currentByID[id]
		if record == nil || !record.Present || record.Kind != "file" || !isMarkdown(record.Path) {
			continue
		}
		source := markdownSource{path: recordAbsolute(root, *record), record: record}
		if err := reconcileMarkdownSource(&plan, inventory, source, previousBySource[id], true, repair, policy); err != nil {
			return Plan{}, err
		}
	}

	plan.Files = inventory.manifest
	sortManifests(&plan.Files, &plan.Links)
	return plan, nil
}
