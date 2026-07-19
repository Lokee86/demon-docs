package links

import (
	"os"
	"path/filepath"
)

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
