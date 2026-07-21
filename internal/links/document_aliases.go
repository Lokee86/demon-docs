package links

import (
	"sort"
	"strconv"
)

// collapseDocumentIdentityAliases repairs private-state duplication created by
// interrupted or older reconciliations. When exactly one file with a given
// document_id is present, absent records carrying that same document_id are
// aliases of the live record rather than distinct documents.
func collapseDocumentIdentityAliases(previousFiles FilesManifest, previousLinks LinksManifest, current *FilesManifest) (FilesManifest, LinksManifest) {
	aliases := documentIdentityAliases(*current)
	if len(aliases) == 0 {
		return previousFiles, previousLinks
	}
	*current = remapFileAliases(*current, aliases)
	previousFiles = remapFileAliases(previousFiles, aliases)
	previousLinks = remapLinkAliases(previousLinks, aliases)
	return previousFiles, previousLinks
}

func documentIdentityAliases(files FilesManifest) map[string]string {
	groups := make(map[string][]FileRecord)
	for _, record := range files.Files {
		if record.Kind != "file" || record.DocumentID == "" || record.ID == "" {
			continue
		}
		groups[record.DocumentID] = append(groups[record.DocumentID], record)
	}
	aliases := make(map[string]string)
	for _, records := range groups {
		var present []FileRecord
		for _, record := range records {
			if record.Present {
				present = append(present, record)
			}
		}
		if len(present) != 1 {
			continue
		}
		canonical := present[0].ID
		for _, record := range records {
			if !record.Present && record.ID != canonical {
				aliases[record.ID] = canonical
			}
		}
	}
	return aliases
}

func remapFileAliases(files FilesManifest, aliases map[string]string) FilesManifest {
	result := FilesManifest{SchemaVersion: files.SchemaVersion}
	byID := make(map[string]int)
	for _, record := range files.Files {
		originalID := record.ID
		if canonical := aliases[record.ID]; canonical != "" {
			record.ID = canonical
		}
		index, exists := byID[record.ID]
		if !exists {
			byID[record.ID] = len(result.Files)
			result.Files = append(result.Files, record)
			continue
		}
		existing := result.Files[index]
		if record.Present && !existing.Present {
			record.PathHistory = mergeRecordHistory(record, existing)
			result.Files[index] = record
			continue
		}
		existing.PathHistory = mergeRecordHistory(existing, record)
		if existing.DocumentID == "" {
			existing.DocumentID = record.DocumentID
		}
		if existing.LinkParserVersion < record.LinkParserVersion {
			existing.LinkParserVersion = record.LinkParserVersion
		}
		if existing.ID == "" {
			existing.ID = originalID
		}
		result.Files[index] = existing
	}
	return result
}

func mergeRecordHistory(base, other FileRecord) []string {
	history := append([]string(nil), base.PathHistory...)
	for _, path := range append(append([]string(nil), other.PathHistory...), other.Path) {
		if path != "" && path != base.Path {
			history = appendUnique(history, path)
		}
	}
	sort.Strings(history)
	return history
}

func remapLinkAliases(links LinksManifest, aliases map[string]string) LinksManifest {
	result := LinksManifest{SchemaVersion: links.SchemaVersion}
	byKey := make(map[string]int)
	for _, record := range links.Links {
		if canonical := aliases[record.SourceFileID]; canonical != "" {
			record.SourceFileID = canonical
		}
		if canonical := aliases[record.TargetFileID]; canonical != "" {
			record.TargetFileID = canonical
		}
		key := record.SourceFileID + "\x00" + strconv.Itoa(record.Ordinal) + "\x00" + record.Syntax + "\x00" + record.Target
		if index, exists := byKey[key]; exists {
			if linkRecordQuality(record) > linkRecordQuality(result.Links[index]) {
				result.Links[index] = record
			}
			continue
		}
		byKey[key] = len(result.Links)
		result.Links = append(result.Links, record)
	}
	return result
}

func linkRecordQuality(record LinkRecord) int {
	score := 0
	if record.ID != "" {
		score++
	}
	if record.TargetFileID != "" {
		score += 2
	}
	if record.Status == "valid" || record.Status == "moved" {
		score += 4
	}
	return score
}
