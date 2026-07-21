package links

import (
	"fmt"
	"sort"

	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

type internalRewritePlan struct {
	rewrite    GeneratedRewrite
	update     model.FileUpdate
	records    []LinkRecord
	messages   []string
	unresolved int
}

type internalRewriteJob struct {
	sourceID        string
	previousRecords []LinkRecord
	currentSource   *FileRecord
}

type internalRewriteResult struct {
	sourceID string
	plan     internalRewritePlan
	include  bool
}

func buildInternalMoveRewrites(root string, previousBySource map[string][]LinkRecord, previousByID, currentByID map[string]*FileRecord, policy review.Policy) (map[string]internalRewritePlan, error) {
	movedTargets := make(map[string]*FileRecord)
	for id, current := range currentByID {
		previous := previousByID[id]
		if previous == nil || !current.Present || (previous.Scope == current.Scope && previous.Path == current.Path) {
			continue
		}
		movedTargets[id] = current
	}

	jobs := make([]internalRewriteJob, 0, len(previousBySource))
	for sourceID, previousRecords := range previousBySource {
		previousSource := previousByID[sourceID]
		currentSource := currentByID[sourceID]
		if !sourceUnchanged(previousSource, currentSource) || !recordsReusable(previousRecords) {
			continue
		}
		hasMovedTarget := false
		for _, record := range previousRecords {
			if movedTargets[record.TargetFileID] != nil {
				hasMovedTarget = true
				break
			}
		}
		if !hasMovedTarget {
			continue
		}
		jobs = append(jobs, internalRewriteJob{
			sourceID:        sourceID,
			previousRecords: previousRecords,
			currentSource:   currentSource,
		})
	}
	sort.Slice(jobs, func(i, j int) bool {
		if jobs[i].currentSource.Path != jobs[j].currentSource.Path {
			return jobs[i].currentSource.Path < jobs[j].currentSource.Path
		}
		return jobs[i].sourceID < jobs[j].sourceID
	})

	results := make([]internalRewriteResult, len(jobs))
	errors := runLinkWorkers(len(jobs), func(index int) error {
		job := jobs[index]
		plan, include, err := buildInternalMoveRewrite(root, job, movedTargets, policy)
		results[index] = internalRewriteResult{sourceID: job.sourceID, plan: plan, include: include}
		return err
	})
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	result := make(map[string]internalRewritePlan, len(results))
	for _, prepared := range results {
		if prepared.include {
			result[prepared.sourceID] = prepared.plan
		}
	}
	return result, nil
}

func buildInternalMoveRewrite(root string, job internalRewriteJob, movedTargets map[string]*FileRecord, policy review.Policy) (internalRewritePlan, bool, error) {
	sourcePath := recordAbsolute(root, *job.currentSource)
	document, err := textio.Read(sourcePath)
	if err != nil {
		return internalRewritePlan{}, false, fmt.Errorf("read internal rewrite source %s: %w", sourcePath, err)
	}
	records := append([]LinkRecord(nil), job.previousRecords...)
	var replacements []replacement
	var messages []string
	unresolved := 0
	metadataChanged := false
	for index := range records {
		target := movedTargets[records[index].TargetFileID]
		if target == nil {
			records[index].SourcePath = job.currentSource.Path
			continue
		}
		targetPath := recordAbsolute(root, *target)
		_, style, local := resolveLocalTarget(records[index].RawPath, sourcePath, records[index].Angle)
		if !local {
			continue
		}
		newPath := renderTargetForSyntax(records[index].Syntax, records[index].RawPath, style, sourcePath, targetPath)
		if newPath == records[index].RawPath {
			records[index].SourcePath = job.currentSource.Path
			records[index].ResolvedPath = target.Path
			records[index].Status = "valid"
			metadataChanged = true
			continue
		}
		if state, decision := reviewRepairPolicy(policy, job.sourceID, job.currentSource.Path, records[index], records[index].RawPath, newPath, target.ID); state != review.MatchNone {
			records[index].SourcePath = job.currentSource.Path
			records[index].Candidates = []string{target.Path}
			records[index].Status = "blocked"
			label := "Blocked"
			if state == review.MatchStale {
				records[index].Status = "stale_block"
				label = "Stale blocked"
			}
			messages = append(messages, fmt.Sprintf("%s link repair in %s:%d: %s -> %s%s", label, job.currentSource.Path, records[index].Line, records[index].RawPath, newPath, reviewReason(decision.Reason)))
			unresolved++
			continue
		}
		replacements = append(replacements, replacement{
			linkID:   records[index].ID,
			start:    records[index].Start,
			end:      records[index].End,
			oldValue: records[index].RawPath,
			newValue: newPath,
		})
		records[index].SourcePath = job.currentSource.Path
		records[index].RawPath = newPath
		records[index].Target = newPath + records[index].Suffix
		records[index].ResolvedPath = target.Path
		records[index].Status = "moved"
		messages = append(messages, fmt.Sprintf("Repair link in %s:%d: %s -> %s", job.currentSource.Path, records[index].Line, replacements[len(replacements)-1].oldValue, newPath))
	}
	if len(replacements) == 0 {
		if metadataChanged || unresolved > 0 {
			return internalRewritePlan{records: records, messages: messages, unresolved: unresolved}, true, nil
		}
		return internalRewritePlan{}, false, nil
	}
	transformations := transformationsFor(replacements)
	rewrite, err := NewGeneratedRewrite(job.sourceID, sourcePath, document, transformations)
	if err != nil {
		if IsTransientFilesystemRace(err) {
			// Stored link offsets can lag behind the current source text even
			// when file identity metadata says the source is unchanged. Skip
			// the internal fast path so normal reconciliation reparses the
			// current document and rebuilds the repair from fresh offsets.
			return internalRewritePlan{}, false, nil
		}
		return internalRewritePlan{}, false, err
	}
	updated := applyReplacements(document.Text, replacements)
	old := document.Text
	return internalRewritePlan{
		rewrite:    rewrite,
		update:     model.FileUpdate{Path: sourcePath, OldText: &old, NewText: updated},
		records:    records,
		messages:   messages,
		unresolved: unresolved,
	}, true, nil
}
