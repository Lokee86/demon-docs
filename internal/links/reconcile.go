package links

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

type replacement struct {
	linkID     string
	start, end int
	oldValue   string
	newValue   string
}

func Reconcile(repositoryRoot string) (Plan, error) {
	return reconcileWithOptions(repositoryRoot, true)
}

// Track refreshes persistent file and link identity without planning user-file
// rewrites. It is used while automatic link maintenance is disabled.
func Track(repositoryRoot string) (Plan, error) {
	return reconcileWithOptions(repositoryRoot, false)
}

func reconcileWithOptions(repositoryRoot string, repair bool) (Plan, error) {
	started := time.Now()
	var timings ReconcileTimings
	plan, err := reconcile(repositoryRoot, repair, &timings)
	timings.Total = time.Since(started)
	return plan, err
}

func reconcileWithTimings(repositoryRoot string) (Plan, ReconcileTimings, error) {
	started := time.Now()
	var timings ReconcileTimings
	plan, err := reconcile(repositoryRoot, true, &timings)
	timings.Total = time.Since(started)
	return plan, timings, err
}

func reconcile(repositoryRoot string, repair bool, timings *ReconcileTimings) (Plan, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Plan{}, err
	}

	stateStarted := time.Now()
	previousFiles, previousLinks, initialized, err := loadState(root)
	if err == nil {
		previousFiles, previousLinks = pruneNestedWorktreeState(root, previousFiles, previousLinks)
	}
	timings.StateLoad = time.Since(stateStarted)
	if err != nil {
		return Plan{}, err
	}

	inventoryStarted := time.Now()
	inventory, err := buildInventory(root, previousFiles)
	timings.InventoryBuild = time.Since(inventoryStarted)
	if err != nil {
		return Plan{}, err
	}

	planningStarted := time.Now()
	plan := Plan{
		RepositoryRoot:      filepath.Clean(root),
		Initialized:         initialized,
		NeedsInitialization: !initialized,
		Files:               inventory.manifest,
		Links:               LinksManifest{SchemaVersion: schemaVersion},
	}
	if !initialized {
		plan.Messages = append(plan.Messages, "Link state is not initialized; this pass records a baseline and does not repair links.")
	}

	policy, err := review.LoadPolicy(root)
	if err != nil {
		return Plan{}, fmt.Errorf("load review policy: %w", err)
	}
	previousBySource := previousLinkIndex(previousLinks)
	previousByID := fileRecordIndex(previousFiles)
	currentByID := fileRecordIndex(inventory.manifest)
	internal := map[string]internalRewritePlan{}
	if repair {
		internal, err = buildInternalMoveRewrites(root, previousBySource, previousByID, currentByID, policy)
		if err != nil {
			return Plan{}, err
		}
	}

	for _, source := range markdownSources(inventory) {
		if rewrite, ok := internal[source.record.ID]; ok {
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
		previousSource := previousByID[source.record.ID]
		previousRecords := previousBySource[source.record.ID]
		if initialized && sourceUnchanged(previousSource, source.record) && previousSource.LinkParserVersion == linkParserVersion && recordsReusable(previousRecords) && !recordsReferenceChangedTarget(previousRecords, previousByID, currentByID) {
			for _, record := range previousRecords {
				record.SourcePath = source.record.Path
				plan.Links.Links = append(plan.Links.Links, record)
			}
			continue
		}
		if err := reconcileMarkdownSource(&plan, inventory, source, previousRecords, initialized, repair, policy); err != nil {
			return Plan{}, err
		}
	}
	plan.Files = inventory.manifest
	sortManifests(&plan.Files, &plan.Links)
	timings.Planning = time.Since(planningStarted)
	return plan, nil
}

type internalRewritePlan struct {
	rewrite    GeneratedRewrite
	update     model.FileUpdate
	records    []LinkRecord
	messages   []string
	unresolved int
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
	result := make(map[string]internalRewritePlan)
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
		sourcePath := recordAbsolute(root, *currentSource)
		document, err := textio.Read(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("read internal rewrite source %s: %w", sourcePath, err)
		}
		records := append([]LinkRecord(nil), previousRecords...)
		var replacements []replacement
		var messages []string
		unresolved := 0
		metadataChanged := false
		for index := range records {
			target := movedTargets[records[index].TargetFileID]
			if target == nil {
				records[index].SourcePath = currentSource.Path
				continue
			}
			targetPath := recordAbsolute(root, *target)
			_, style, local := resolveLocalTarget(records[index].RawPath, sourcePath, records[index].Angle)
			if !local {
				continue
			}
			newPath := renderTargetForSyntax(records[index].Syntax, records[index].RawPath, style, sourcePath, targetPath)
			if newPath == records[index].RawPath {
				records[index].SourcePath = currentSource.Path
				records[index].ResolvedPath = target.Path
				records[index].Status = "valid"
				metadataChanged = true
				continue
			}
			if state, decision := reviewRepairPolicy(policy, sourceID, currentSource.Path, records[index], records[index].RawPath, newPath, target.ID); state != review.MatchNone {
				records[index].SourcePath = currentSource.Path
				records[index].Candidates = []string{target.Path}
				records[index].Status = "blocked"
				label := "Blocked"
				if state == review.MatchStale {
					records[index].Status = "stale_block"
					label = "Stale blocked"
				}
				messages = append(messages, fmt.Sprintf("%s link repair in %s:%d: %s -> %s%s", label, currentSource.Path, records[index].Line, records[index].RawPath, newPath, reviewReason(decision.Reason)))
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
			records[index].SourcePath = currentSource.Path
			records[index].RawPath = newPath
			records[index].Target = newPath + records[index].Suffix
			records[index].ResolvedPath = target.Path
			records[index].Status = "moved"
			messages = append(messages, fmt.Sprintf("Repair link in %s:%d: %s -> %s", currentSource.Path, records[index].Line, replacements[len(replacements)-1].oldValue, newPath))
		}
		if len(replacements) == 0 {
			if metadataChanged || unresolved > 0 {
				result[sourceID] = internalRewritePlan{records: records, messages: messages, unresolved: unresolved}
			}
			continue
		}
		transformations := transformationsFor(replacements)
		rewrite, err := NewGeneratedRewrite(sourceID, sourcePath, document, transformations)
		if err != nil {
			return nil, err
		}
		updated := applyReplacements(document.Text, replacements)
		old := document.Text
		result[sourceID] = internalRewritePlan{
			rewrite:    rewrite,
			update:     model.FileUpdate{Path: sourcePath, OldText: &old, NewText: updated},
			records:    records,
			messages:   messages,
			unresolved: unresolved,
		}
	}
	return result, nil
}

func reconcileMarkdownSource(plan *Plan, inventory *inventory, source markdownSource, previousRecords []LinkRecord, initialized, repair bool, policy review.Policy) error {
	document, err := textio.Read(source.path)
	if err != nil {
		return fmt.Errorf("read Markdown source %s: %w", source.path, err)
	}
	parsed := parseMarkdownDocument(document.Text)
	source.record.LinkParserVersion = linkParserVersion
	var replacements []replacement
	ordinal := 0
	for _, found := range parsed.Links {
		resolved, style, local := resolveLocalTarget(found.RawPath, source.path, found.Angle)
		if !local {
			continue
		}
		ignored, err := inventory.ignored(resolved)
		if err != nil {
			return fmt.Errorf("evaluate link target ignore policy %s: %w", resolved, err)
		}
		if ignored {
			continue
		}
		originalTarget := found.RawPath + found.Suffix
		previous := findPreviousLink(previousRecords, ordinal, originalTarget, found.Syntax)
		record := LinkRecord{
			ID:            deterministicLinkID(source.record.ID, ordinal, found.Syntax, originalTarget),
			SourceFileID:  source.record.ID,
			SourcePath:    source.record.Path,
			Ordinal:       ordinal,
			Start:         found.Start,
			End:           found.End,
			Line:          found.Line,
			Column:        found.Column,
			Syntax:        found.Syntax,
			RawPath:       found.RawPath,
			Suffix:        found.Suffix,
			Angle:         found.Angle,
			Target:        originalTarget,
			ParserVersion: linkParserVersion,
		}
		if previous != nil && previous.ID != "" {
			record.ID = previous.ID
		}
		ordinal++
		targetRecord, actualPath := exactTargetForSyntax(inventory, resolved, found.Syntax)
		if targetRecord == nil {
			if _, statErr := os.Stat(resolved); statErr == nil {
				targetRecord, actualPath, err = inventory.ensureTarget(resolved, "")
				if err != nil {
					return fmt.Errorf("record link target %s: %w", resolved, err)
				}
			}
		}
		if targetRecord != nil {
			record.TargetFileID = targetRecord.ID
			record.ResolvedPath = storePath(inventory.root, actualPath)
			record.Status = "valid"
			if targetCaseMismatch(found.Syntax, resolved, actualPath) {
				record.Status = "case_mismatch"
				if initialized && repair {
					newPath := renderTargetForSyntax(found.Syntax, found.RawPath, style, source.path, actualPath)
					if state, decision := reviewRepairPolicy(policy, source.record.ID, source.record.Path, record, found.RawPath, newPath, targetRecord.ID); state != review.MatchNone {
						record.Status = "blocked"
						label := "Blocked"
						if state == review.MatchStale {
							record.Status = "stale_block"
							label = "Stale blocked"
						}
						record.Candidates = []string{storePath(inventory.root, actualPath)}
						plan.Unresolved++
						plan.Messages = append(plan.Messages, fmt.Sprintf("%s link repair in %s:%d: %s -> %s%s", label, source.record.Path, found.Line, found.RawPath, newPath, reviewReason(decision.Reason)))
					} else {
						replacements = append(replacements, replacement{record.ID, found.Start, found.End, found.RawPath, newPath})
						record.RawPath = newPath
						record.Target = newPath + found.Suffix
						plan.Messages = append(plan.Messages, fmt.Sprintf("Updated link case in %s:%d: %s -> %s", source.record.Path, found.Line, found.RawPath, newPath))
					}
				}
			}
			plan.Links.Links = append(plan.Links.Links, record)
			continue
		}

		preferredID := ""
		if previous != nil {
			preferredID = previous.TargetFileID
		}
		var candidates []string
		if initialized && preferredID != "" {
			if _, moved := inventory.byID(preferredID); moved != "" {
				candidates = []string{moved}
			}
		}
		if len(candidates) == 0 && (initialized || found.Syntax == "wiki") {
			candidates = candidatePathsForSyntax(inventory, resolved, preferredID, found.Syntax)
		}
		record.Candidates = displayPaths(inventory.root, candidates)
		switch len(candidates) {
		case 0:
			record.Status = "broken"
			plan.Unresolved++
			plan.Messages = append(plan.Messages, fmt.Sprintf("Broken link in %s:%d:%d: %s", source.record.Path, found.Line, found.Column, originalTarget))
		case 1:
			candidate := candidates[0]
			targetRecord, actualPath, err = inventory.ensureTarget(candidate, preferredID)
			if err != nil {
				return fmt.Errorf("record moved link target %s: %w", candidate, err)
			}
			record.TargetFileID = targetRecord.ID
			record.ResolvedPath = storePath(inventory.root, actualPath)
			newPath := renderTargetForSyntax(found.Syntax, found.RawPath, style, source.path, actualPath)
			if newPath == found.RawPath {
				record.Status = "valid"
			} else {
				record.Status = "moved"
				if repair {
					if state, decision := reviewRepairPolicy(policy, source.record.ID, source.record.Path, record, found.RawPath, newPath, targetRecord.ID); state != review.MatchNone {
						record.Status = "blocked"
						label := "Blocked"
						if state == review.MatchStale {
							record.Status = "stale_block"
							label = "Stale blocked"
						}
						plan.Unresolved++
						plan.Messages = append(plan.Messages, fmt.Sprintf("%s link repair in %s:%d: %s -> %s%s", label, source.record.Path, found.Line, found.RawPath, newPath, reviewReason(decision.Reason)))
					} else {
						replacements = append(replacements, replacement{record.ID, found.Start, found.End, found.RawPath, newPath})
						record.RawPath = newPath
						record.Target = newPath + found.Suffix
						plan.Messages = append(plan.Messages, fmt.Sprintf("Repair link in %s:%d: %s -> %s", source.record.Path, found.Line, found.RawPath, newPath))
					}
				}
			}
		default:
			record.Status = "ambiguous"
			plan.Unresolved++
			plan.Messages = append(plan.Messages, fmt.Sprintf("Ambiguous link in %s:%d:%d: %s; candidates: %s", source.record.Path, found.Line, found.Column, originalTarget, strings.Join(record.Candidates, ", ")))
		}
		plan.Links.Links = append(plan.Links.Links, record)
	}
	for _, missing := range parsed.UndefinedReferences {
		originalTarget := missing.Label
		previous := findPreviousLink(previousRecords, ordinal, originalTarget, "reference_use")
		record := LinkRecord{
			ID:            deterministicLinkID(source.record.ID, ordinal, "reference_use", originalTarget),
			SourceFileID:  source.record.ID,
			SourcePath:    source.record.Path,
			Ordinal:       ordinal,
			Start:         missing.Start,
			End:           missing.End,
			Line:          missing.Line,
			Column:        missing.Column,
			Syntax:        "reference_use",
			RawPath:       originalTarget,
			Target:        originalTarget,
			Status:        "undefined_reference",
			ParserVersion: linkParserVersion,
		}
		if previous != nil && previous.ID != "" {
			record.ID = previous.ID
		}
		ordinal++
		plan.Unresolved++
		plan.Messages = append(plan.Messages, fmt.Sprintf("Undefined reference label in %s:%d:%d: %s", source.record.Path, missing.Line, missing.Column, missing.Label))
		plan.Links.Links = append(plan.Links.Links, record)
	}
	if initialized && repair && len(replacements) > 0 {
		updated := applyReplacements(document.Text, replacements)
		if updated != document.Text {
			rewrite, err := NewGeneratedRewrite(source.record.ID, source.path, document, transformationsFor(replacements))
			if err != nil {
				return err
			}
			plan.Rewrites = append(plan.Rewrites, rewrite)
			old := document.Text
			plan.Updates = append(plan.Updates, model.FileUpdate{Path: source.path, OldText: &old, NewText: updated})
		}
	}
	return nil
}

func transformationsFor(replacements []replacement) []LinkTransformation {
	result := make([]LinkTransformation, 0, len(replacements))
	for _, item := range replacements {
		result = append(result, LinkTransformation{
			LinkID:         item.linkID,
			Start:          item.start,
			End:            item.end,
			OldDestination: item.oldValue,
			NewDestination: item.newValue,
		})
	}
	return result
}

func sourceUnchanged(previous, current *FileRecord) bool {
	return previous != nil && current != nil && previous.Present && current.Present &&
		previous.Scope == current.Scope && previous.Path == current.Path &&
		previous.Fingerprint != "" && previous.Fingerprint == current.Fingerprint
}

func recordsHaveRewriteMetadata(records []LinkRecord) bool {
	for _, record := range records {
		if record.ID == "" || record.End < record.Start || record.Target != record.RawPath+record.Suffix {
			return false
		}
	}
	return true
}

func recordsReferenceChangedTarget(records []LinkRecord, previousByID, currentByID map[string]*FileRecord) bool {
	for _, record := range records {
		if record.TargetFileID == "" {
			continue
		}
		previous := previousByID[record.TargetFileID]
		current := currentByID[record.TargetFileID]
		if previous == nil || current == nil || previous.Present != current.Present || previous.Scope != current.Scope || previous.Path != current.Path {
			return true
		}
	}
	return false
}

func recordsReusable(records []LinkRecord) bool {
	if !recordsHaveRewriteMetadata(records) {
		return false
	}
	for _, record := range records {
		if record.ParserVersion != linkParserVersion || record.Status != "valid" {
			return false
		}
	}
	return true
}

func fileRecordIndex(manifest FilesManifest) map[string]*FileRecord {
	result := make(map[string]*FileRecord, len(manifest.Files))
	for index := range manifest.Files {
		result[manifest.Files[index].ID] = &manifest.Files[index]
	}
	return result
}

func deterministicLinkID(sourceID string, ordinal int, syntax, target string) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%s\x00%s", sourceID, ordinal, syntax, target)))
	return hex.EncodeToString(digest[:16])
}

type markdownSource struct {
	path   string
	record *FileRecord
}

func markdownSources(inventory *inventory) []markdownSource {
	var result []markdownSource
	for index := range inventory.manifest.Files {
		record := &inventory.manifest.Files[index]
		if record.Scope != "repository" || !record.Present || record.Kind != "file" || !isMarkdown(record.Path) {
			continue
		}
		result = append(result, markdownSource{path: recordAbsolute(inventory.root, *record), record: record})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].record.Path < result[j].record.Path })
	return result
}

func isMarkdown(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown", ".mkd", ".mdx":
		return true
	default:
		return false
	}
}

func previousLinkIndex(manifest LinksManifest) map[string][]LinkRecord {
	result := map[string][]LinkRecord{}
	for _, record := range manifest.Links {
		result[record.SourceFileID] = append(result[record.SourceFileID], record)
	}
	for sourceID := range result {
		sort.Slice(result[sourceID], func(i, j int) bool { return result[sourceID][i].Ordinal < result[sourceID][j].Ordinal })
	}
	return result
}

func findPreviousLink(records []LinkRecord, ordinal int, target, syntax string) *LinkRecord {
	for index := range records {
		if records[index].Ordinal == ordinal && records[index].Target == target && records[index].Syntax == syntax {
			return &records[index]
		}
	}
	var match *LinkRecord
	for index := range records {
		if records[index].Target == target && records[index].Syntax == syntax {
			if match != nil {
				return nil
			}
			match = &records[index]
		}
	}
	return match
}

func candidatePaths(inventory *inventory, missingPath, preferredID string) []string {
	base := filepath.Base(missingPath)
	kind := "file"
	fingerprint := ""
	if preferred := inventory.recordByID(preferredID); preferred != nil {
		kind = preferred.Kind
		fingerprint = preferred.Fingerprint
	}
	candidates := inventory.candidates(base, kind)
	if fingerprint != "" {
		var exact []string
		for _, candidate := range candidates {
			if record, _ := inventory.exact(candidate); record != nil && record.Fingerprint == fingerprint {
				exact = append(exact, candidate)
			}
		}
		if len(exact) > 0 {
			candidates = exact
		}
	}
	if !strings.EqualFold(filepath.Dir(missingPath), inventory.root) {
		candidates = append(candidates, discoverExternalCandidates(missingPath, base, kind, fingerprint)...)
	}
	return rankPathAwareCandidates(inventory.root, missingPath, uniquePaths(candidates))
}

func displayPaths(root string, paths []string) []string {
	result := make([]string, len(paths))
	for index, path := range paths {
		result[index] = storePath(root, path)
	}
	sort.Strings(result)
	return result
}

func applyReplacements(source string, replacements []replacement) string {
	ordered := append([]replacement(nil), replacements...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].start > ordered[j].start })
	result := source
	for _, replacement := range ordered {
		result = result[:replacement.start] + replacement.newValue + result[replacement.end:]
	}
	return result
}
