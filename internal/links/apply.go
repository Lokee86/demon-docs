package links

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Lokee86/demon-docs/internal/textio"
)

func ApplyAndSave(plan *Plan) (int, error) {
	updates, _, err := applyAndSaveWithTimings(plan)
	return updates, err
}

func applyAndSaveWithTimings(plan *Plan) (int, ApplyTimings, error) {
	started := time.Now()
	var timings ApplyTimings
	updates, err := applyAndSave(plan, &timings)
	timings.Total = time.Since(started)
	return updates, timings, err
}

func applyAndSave(plan *Plan, timings *ApplyTimings) (int, error) {
	changeBatch, err := prepareGeneratedChanges(plan)
	if err != nil {
		return 0, err
	}

	rewriteStarted := time.Now()
	suppressions, err := ApplyGenerated(plan.Rewrites)
	timings.FilesystemRewrite = time.Since(rewriteStarted)
	if err != nil {
		return 0, err
	}

	refreshStarted := time.Now()
	if err := refreshGeneratedSources(plan); err != nil {
		timings.GeneratedSourceRefresh = time.Since(refreshStarted)
		if rollbackErr := RollbackGenerated(plan.Rewrites); rollbackErr != nil {
			return 0, errors.Join(err, fmt.Errorf("restore source files after generated rewrite verification failure: %w", rollbackErr))
		}
		return 0, err
	}
	timings.GeneratedSourceRefresh = time.Since(refreshStarted)

	if err := recordGeneratedChanges(plan, changeBatch); err != nil {
		if rollbackErr := RollbackGenerated(plan.Rewrites); rollbackErr != nil {
			return 0, errors.Join(err, fmt.Errorf("restore source files after review history failure: %w", rollbackErr))
		}
		return 0, err
	}
	plan.Suppressions = suppressions

	publicationStarted := time.Now()
	if err := Save(*plan); err != nil {
		timings.DdocsPublication = time.Since(publicationStarted)
		return 0, err
	}
	timings.DdocsPublication = time.Since(publicationStarted)

	return len(plan.Rewrites), nil
}

func refreshGeneratedSources(plan *Plan) error {
	results := make([]sourceRefreshResult, len(plan.Rewrites))
	errors := runLinkWorkers(len(plan.Rewrites), func(index int) error {
		return refreshGeneratedSource(plan.Links.Links, plan.Rewrites[index], &results[index])
	})
	for _, err := range errors {
		if err != nil {
			return err
		}
	}

	// Workers only produce detached copies. Merge in rewrite order so plan
	// mutation remains deterministic and race-free.
	for _, result := range results {
		for _, link := range result.links {
			plan.Links.Links[link.index] = link.record
		}
		for index := range plan.Files.Files {
			if plan.Files.Files[index].ID == result.sourceFileID {
				plan.Files.Files[index].Fingerprint = result.fingerprint
				plan.Files.Files[index].Size = result.size
				plan.Files.Files[index].ModifiedUnixNano = result.modifiedUnixNano
				break
			}
		}
	}
	return nil
}

type sourceRefreshResult struct {
	sourceFileID     string
	fingerprint      string
	size             int64
	modifiedUnixNano int64
	links            []refreshedLink
}

type refreshedLink struct {
	index  int
	record LinkRecord
}

func refreshGeneratedSource(records []LinkRecord, rewrite GeneratedRewrite, result *sourceRefreshResult) error {
	document, err := textio.Read(rewrite.Path)
	if err != nil {
		return fmt.Errorf("read generated rewrite result %s: %w", rewrite.Path, err)
	}
	occurrences := parseMarkdownLinks(document.Text)
	indexes := sourceLinkIndexes(records, rewrite.SourceFileID)
	result.links = make([]refreshedLink, 0, len(indexes))
	searchFrom := 0
	for _, index := range indexes {
		record := records[index]
		foundIndex := -1
		for occurrenceIndex := searchFrom; occurrenceIndex < len(occurrences); occurrenceIndex++ {
			occurrence := occurrences[occurrenceIndex]
			if occurrence.RawPath+occurrence.Suffix == record.Target {
				foundIndex = occurrenceIndex
				break
			}
		}
		if foundIndex < 0 {
			return fmt.Errorf("generated rewrite verification could not find link %s in %s", record.ID, rewrite.Path)
		}
		found := occurrences[foundIndex]
		record.Start = found.Start
		record.End = found.End
		record.Line = found.Line
		record.Column = found.Column
		record.Syntax = found.Syntax
		record.RawPath = found.RawPath
		record.Suffix = found.Suffix
		record.Angle = found.Angle
		result.links = append(result.links, refreshedLink{index: index, record: record})
		searchFrom = foundIndex + 1
	}
	fingerprint, err := fileFingerprint(rewrite.Path)
	if err != nil {
		return fmt.Errorf("fingerprint generated rewrite %s: %w", rewrite.Path, err)
	}
	info, err := os.Stat(rewrite.Path)
	if err != nil {
		return fmt.Errorf("stat generated rewrite %s: %w", rewrite.Path, err)
	}
	result.sourceFileID = rewrite.SourceFileID
	result.fingerprint = fingerprint
	result.size = info.Size()
	result.modifiedUnixNano = info.ModTime().UnixNano()
	return nil
}

func sourceLinkIndexes(records []LinkRecord, sourceFileID string) []int {
	var indexes []int
	for index := range records {
		if records[index].SourceFileID == sourceFileID {
			indexes = append(indexes, index)
		}
	}
	sort.Slice(indexes, func(i, j int) bool {
		return records[indexes[i]].Ordinal < records[indexes[j]].Ordinal
	})
	return indexes
}
