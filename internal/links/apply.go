package links

import (
	"fmt"
	"os"
	"sort"

	"github.com/Lokee86/demon-docs/internal/textio"
)

func ApplyAndSave(plan *Plan) (int, error) {
	suppressions, err := ApplyGenerated(plan.Rewrites)
	if err != nil {
		return 0, err
	}
	plan.Suppressions = suppressions
	if err := refreshGeneratedSources(plan); err != nil {
		return 0, err
	}
	if err := Save(*plan); err != nil {
		return 0, err
	}
	return len(plan.Rewrites), nil
}

func refreshGeneratedSources(plan *Plan) error {
	for _, rewrite := range plan.Rewrites {
		document, err := textio.Read(rewrite.Path)
		if err != nil {
			return fmt.Errorf("read generated rewrite result %s: %w", rewrite.Path, err)
		}
		occurrences := parseMarkdownLinks(document.Text)
		indexes := sourceLinkIndexes(plan.Links.Links, rewrite.SourceFileID)
		searchFrom := 0
		for _, index := range indexes {
			record := &plan.Links.Links[index]
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
		for index := range plan.Files.Files {
			if plan.Files.Files[index].ID == rewrite.SourceFileID {
				plan.Files.Files[index].Fingerprint = fingerprint
				plan.Files.Files[index].Size = info.Size()
				plan.Files.Files[index].ModifiedUnixNano = info.ModTime().UnixNano()
				break
			}
		}
	}
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
