package links

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

// RepairObservedRename applies a conservative same-directory file-rename repair
// from persisted link state. It returns handled=false without writing when the
// event cannot be proven safe, allowing the normal full reconciliation to run.
func RepairObservedRename(repositoryRoot, oldPath, newPath string) (bool, int, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return false, 0, err
	}
	root = filepath.Clean(root)
	oldPath, err = filepath.Abs(oldPath)
	if err != nil {
		return false, 0, err
	}
	newPath, err = filepath.Abs(newPath)
	if err != nil {
		return false, 0, err
	}
	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)
	if !repository.Contains(root, oldPath) || !repository.Contains(root, newPath) {
		return false, 0, nil
	}
	if pathKey(filepath.Dir(oldPath)) != pathKey(filepath.Dir(newPath)) || pathKey(oldPath) == pathKey(newPath) {
		return false, 0, nil
	}
	if _, err := os.Lstat(oldPath); err == nil || !os.IsNotExist(err) {
		return false, 0, nil
	}
	info, err := os.Lstat(newPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return false, 0, nil
	}

	previousFiles, previousLinks, initialized, err := loadState(root)
	if err != nil {
		return false, 0, err
	}
	if !initialized {
		return false, 0, nil
	}
	oldStored := storePath(root, oldPath)
	newStored := storePath(root, newPath)
	movedIndex := -1
	for index, record := range previousFiles.Files {
		if record.Scope == "repository" && record.Present && record.Kind == "file" && pathKey(record.Path) == pathKey(oldStored) {
			if movedIndex >= 0 {
				return false, 0, nil
			}
			movedIndex = index
		}
		if record.Scope == "repository" && record.Present && pathKey(record.Path) == pathKey(newStored) {
			return false, 0, nil
		}
	}
	if movedIndex < 0 || previousFiles.Files[movedIndex].Fingerprint == "" {
		return false, 0, nil
	}
	fingerprint, err := fileFingerprint(newPath)
	if err != nil {
		return false, 0, nil
	}
	movedPrevious := previousFiles.Files[movedIndex]
	if fingerprint != movedPrevious.Fingerprint {
		return false, 0, nil
	}
	for index, record := range previousFiles.Files {
		if index != movedIndex && record.Present && record.Kind == "file" && record.Fingerprint == fingerprint {
			return false, 0, nil
		}
	}

	currentFiles := FilesManifest{SchemaVersion: previousFiles.SchemaVersion, Files: append([]FileRecord(nil), previousFiles.Files...)}
	movedCurrent := &currentFiles.Files[movedIndex]
	movedCurrent.PathHistory = appendUnique(movedCurrent.PathHistory, movedCurrent.Path)
	movedCurrent.Path = newStored
	movedCurrent.Size = info.Size()
	movedCurrent.ModifiedUnixNano = info.ModTime().UnixNano()
	movedCurrent.Fingerprint = fingerprint
	movedCurrent.Present = true

	previousBySource := previousLinkIndex(previousLinks)
	previousByID := fileRecordIndex(previousFiles)
	currentByID := fileRecordIndex(currentFiles)
	for sourceID, records := range previousBySource {
		affected := false
		for _, record := range records {
			if record.TargetFileID == movedPrevious.ID {
				affected = true
				break
			}
		}
		if !affected {
			continue
		}
		previousSource := previousByID[sourceID]
		currentSource := currentByID[sourceID]
		if sourceID == movedPrevious.ID || !sourceUnchanged(previousSource, currentSource) || !recordsReusable(records) {
			return false, 0, nil
		}
		sourcePath := recordAbsolute(root, *previousSource)
		sourceInfo, statErr := os.Lstat(sourcePath)
		if statErr != nil || !sourceInfo.Mode().IsRegular() || sourceInfo.Mode()&os.ModeSymlink != 0 || sourceInfo.Size() != previousSource.Size {
			return false, 0, nil
		}
		sourceFingerprint, fingerprintErr := fileFingerprint(sourcePath)
		if fingerprintErr != nil || sourceFingerprint != previousSource.Fingerprint {
			return false, 0, nil
		}
	}

	policy, err := review.LoadPolicy(root)
	if err != nil {
		return false, 0, fmt.Errorf("load review policy: %w", err)
	}
	internal, err := buildInternalMoveRewrites(root, previousBySource, previousByID, currentByID, policy)
	if err != nil {
		return false, 0, err
	}
	for sourceID, rewrite := range internal {
		updated, err := refreshObservedRenameLabel(rewrite, filepath.Base(oldPath), filepath.Base(newPath))
		if err != nil {
			return false, 0, err
		}
		internal[sourceID] = updated
	}
	plan := Plan{
		RepositoryRoot: root,
		Initialized:    true,
		Files:          currentFiles,
		Links:          LinksManifest{SchemaVersion: schemaVersion},
	}
	for sourceID, records := range previousBySource {
		if rewrite, ok := internal[sourceID]; ok {
			if rewrite.rewrite.SourceFileID != "" {
				plan.Rewrites = append(plan.Rewrites, rewrite.rewrite)
			}
			if rewrite.update.Path != "" {
				plan.Updates = append(plan.Updates, rewrite.update)
			}
			plan.Links.Links = append(plan.Links.Links, rewrite.records...)
			plan.Messages = append(plan.Messages, rewrite.messages...)
			plan.Unresolved += rewrite.unresolved
			continue
		}
		for _, record := range records {
			if record.SourceFileID == movedPrevious.ID {
				record.SourcePath = newStored
			}
			plan.Links.Links = append(plan.Links.Links, record)
		}
	}
	sortManifests(&plan.Files, &plan.Links)
	changed, err := ApplyAndSave(&plan)
	if err != nil {
		return false, 0, err
	}
	return true, changed, nil
}

func refreshObservedRenameLabel(plan internalRewritePlan, oldName, newName string) (internalRewritePlan, error) {
	if plan.rewrite.SourceFileID == "" || len(plan.rewrite.Transformations) == 0 || oldName == newName {
		return plan, nil
	}
	document, err := textio.Read(plan.rewrite.Path)
	if err != nil {
		return plan, fmt.Errorf("read observed rename source %s: %w", plan.rewrite.Path, err)
	}
	transformations := append([]LinkTransformation(nil), plan.rewrite.Transformations...)
	for _, destination := range plan.rewrite.Transformations {
		start, end, ok := managedFilenameLabelRange(document.Text, destination.Start, oldName)
		if !ok {
			continue
		}
		transformations = append(transformations, LinkTransformation{
			LinkID:         destination.LinkID + ":label",
			Start:          start,
			End:            end,
			OldDestination: oldName,
			NewDestination: newName,
		})
	}
	if len(transformations) == len(plan.rewrite.Transformations) {
		return plan, nil
	}
	rewrite, err := NewGeneratedRewrite(plan.rewrite.SourceFileID, plan.rewrite.Path, document, transformations)
	if err != nil {
		return plan, err
	}
	newText, err := rewriteText(document.Text, transformations)
	if err != nil {
		return plan, err
	}
	oldText := document.Text
	plan.rewrite = rewrite
	plan.update.OldText = &oldText
	plan.update.NewText = newText
	return plan, nil
}

func managedFilenameLabelRange(source string, destinationStart int, oldName string) (int, int, bool) {
	if destinationStart < 2 || destinationStart > len(source) || source[destinationStart-2:destinationStart] != "](" {
		return 0, 0, false
	}
	lineStart := strings.LastIndex(source[:destinationStart], "\n") + 1
	openRelative := strings.LastIndex(source[lineStart:destinationStart-2], "[")
	if openRelative < 0 {
		return 0, 0, false
	}
	labelStart := lineStart + openRelative + 1
	labelEnd := destinationStart - 2
	if source[labelStart:labelEnd] != oldName || !insideManagedFileRegion(source, destinationStart) {
		return 0, 0, false
	}
	return labelStart, labelEnd, true
}

func insideManagedFileRegion(source string, offset int) bool {
	for _, section := range []string{"files", "stubs"} {
		startMarker := "<!-- doc-ledger:" + section + ":start -->"
		endMarker := "<!-- doc-ledger:" + section + ":end -->"
		start := strings.LastIndex(source[:offset], startMarker)
		if start < 0 {
			continue
		}
		endRelative := strings.Index(source[offset:], endMarker)
		if endRelative >= 0 {
			return true
		}
	}
	return false
}
