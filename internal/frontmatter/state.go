package frontmatter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

type immutableRecord struct {
	Path       string         `json:"path"`
	DocumentID string         `json:"document_id,omitempty"`
	Values     map[string]any `json:"values"`
}

func readImmutable(repoRoot, relative string, values map[string]any, allowID bool) map[string]any {
	repository, err := ddrepo.Open(filepath.Join(repoRoot, ".ddocs"))
	if err != nil {
		return nil
	}
	tx, err := repository.Begin()
	if err != nil {
		return nil
	}
	if allowID {
		if id := documentID(values); id != "" {
			if record := readImmutableRecord(tx, idRecordName(id)); record != nil && record.DocumentID == id {
				return normalizeMap(record.Values)
			}
		}
	}
	record := readImmutableRecord(tx, pathRecordName(relative))
	if record == nil || filepath.ToSlash(record.Path) != filepath.ToSlash(relative) {
		return nil
	}
	return normalizeMap(record.Values)
}

func readImmutableRecord(tx *ddrepo.Transaction, name string) *immutableRecord {
	data, err := tx.Read(name)
	if err != nil {
		return nil
	}
	var record immutableRecord
	if json.Unmarshal(data, &record) != nil {
		return nil
	}
	return &record
}

func writeImmutable(repoRoot string, records map[string]map[string]any) error {
	if len(records) == 0 {
		return nil
	}
	storage := filepath.Join(repoRoot, ".ddocs")
	repository, err := ddrepo.Open(storage)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// go-git does not reliably wrap missing-repository errors.
			if _, statErr := os.Stat(filepath.Join(storage, "HEAD")); statErr == nil {
				return err
			}
		}
		repository, err = ddrepo.Init(storage)
		if err != nil {
			return err
		}
	}
	return repository.Transaction(func(tx *ddrepo.Transaction) error {
		for relative, values := range records {
			id := documentID(values)
			name := pathRecordName(relative)
			if id != "" {
				name = idRecordName(id)
			}
			data, err := json.Marshal(immutableRecord{
				Path:       filepath.ToSlash(relative),
				DocumentID: id,
				Values:     values,
			})
			if err != nil {
				return err
			}
			if err := tx.Write(name, data); err != nil {
				return err
			}
		}
		return nil
	})
}

func documentID(values map[string]any) string {
	value, ok := values["document_id"].(string)
	if !ok {
		return ""
	}
	value = strings.TrimSpace(value)
	if !uuidPattern.MatchString(value) {
		return ""
	}
	return strings.ToLower(value)
}

func pathRecordName(relative string) string {
	return "frontmatter-immutable-path-" + digestName(filepath.ToSlash(relative))
}

func idRecordName(id string) string {
	return "frontmatter-immutable-id-" + digestName(strings.ToLower(strings.TrimSpace(id)))
}

func digestName(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
