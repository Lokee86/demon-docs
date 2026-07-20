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

type immutableIndex struct {
	byID   map[string]immutableRecord
	byPath map[string]immutableRecord
}

func loadImmutableIndex(repoRoot string) immutableIndex {
	index := immutableIndex{
		byID:   make(map[string]immutableRecord),
		byPath: make(map[string]immutableRecord),
	}
	repository, err := ddrepo.Open(filepath.Join(repoRoot, ".ddocs"))
	if err != nil {
		return index
	}
	tx, err := repository.Begin()
	if err != nil {
		return index
	}
	idNames, err := tx.Names("frontmatter-immutable-id-")
	if err == nil {
		for _, name := range idNames {
			record := readImmutableRecord(tx, name)
			if record != nil && record.DocumentID != "" {
				index.byID[strings.ToLower(strings.TrimSpace(record.DocumentID))] = *record
			}
		}
	}
	pathNames, err := tx.Names("frontmatter-immutable-path-")
	if err == nil {
		for _, name := range pathNames {
			record := readImmutableRecord(tx, name)
			if record != nil && record.Path != "" {
				index.byPath[filepath.ToSlash(record.Path)] = *record
			}
		}
	}
	return index
}

func (index immutableIndex) values(relative string, values map[string]any, allowID bool) map[string]any {
	if allowID {
		if id := documentID(values); id != "" {
			if record, ok := index.byID[id]; ok && strings.EqualFold(record.DocumentID, id) {
				return normalizeMap(record.Values)
			}
		}
	}
	record, ok := index.byPath[filepath.ToSlash(relative)]
	if !ok || filepath.ToSlash(record.Path) != filepath.ToSlash(relative) {
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
