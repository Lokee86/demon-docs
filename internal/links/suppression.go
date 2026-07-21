package links

import (
	"os"
	"path/filepath"
	"sort"
)

func mergeSuppressions(existing, generated []Suppression) []Suppression {
	bySource := make(map[string]Suppression, len(existing)+len(generated))
	for _, suppression := range existing {
		bySource[suppression.SourceFileID] = suppression
	}
	for _, suppression := range generated {
		bySource[suppression.SourceFileID] = suppression
	}
	result := make([]Suppression, 0, len(bySource))
	for _, suppression := range bySource {
		result = append(result, suppression)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SourceFileID < result[j].SourceFileID
	})
	return result
}

func ConsumePendingSuppression(repositoryRoot, path string) (bool, error) {
	path = filepath.Clean(path)
	records, err := LoadPendingSuppressions(repositoryRoot)
	if err != nil {
		return false, err
	}
	for _, record := range records {
		if pathKey(record.Path) != pathKey(path) {
			continue
		}
		data, readErr := os.ReadFile(path)
		matched := readErr == nil && sha256Digest(data) == record.ExpectedNewSHA256
		if err := DeletePendingSuppression(repositoryRoot, record.SourceFileID); err != nil {
			return false, err
		}
		return matched, nil
	}
	return false, nil
}
