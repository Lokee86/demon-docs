package codemapsemantic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

const (
	SchemaVersion = 1
	recordPrefix  = "codemap-semantic/"
)

type Baseline struct {
	SchemaVersion  int    `json:"schema_version"`
	Document       string `json:"document"`
	DocumentSHA256 string `json:"document_sha256"`
	SnapshotID     string `json:"snapshot_id"`
}

func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func LoadMany(repositoryRoot string, documents []string) (map[string]Baseline, error) {
	result := map[string]Baseline{}
	if _, err := os.Stat(filepath.Join(repositoryRoot, ".ddocs")); os.IsNotExist(err) {
		return result, nil
	} else if err != nil {
		return nil, err
	}
	repository, err := ddrepo.Open(repositoryRoot)
	if err != nil {
		return nil, fmt.Errorf("open codemap semantic state: %w", err)
	}
	tx, err := repository.Begin()
	if err != nil {
		return nil, err
	}
	for _, document := range documents {
		payload, err := tx.Read(recordPrefix + document)
		if errors.Is(err, ddrepo.ErrRecordAbsent) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var baseline Baseline
		if err := json.Unmarshal(payload, &baseline); err != nil {
			return nil, fmt.Errorf("decode codemap semantic baseline %s: %w", document, err)
		}
		if baseline.SchemaVersion != SchemaVersion || baseline.Document != document || baseline.SnapshotID == "" || baseline.DocumentSHA256 == "" {
			return nil, fmt.Errorf("invalid codemap semantic baseline for %s", document)
		}
		result[document] = baseline
	}
	return result, nil
}

func SaveAll(repositoryRoot string, baselines []Baseline) error {
	if len(baselines) == 0 {
		return nil
	}
	repository, err := openOrInit(repositoryRoot)
	if err != nil {
		return err
	}
	return repository.TransactionRetry(3, func(tx *ddrepo.Transaction) error {
		for _, baseline := range baselines {
			baseline.SchemaVersion = SchemaVersion
			payload, err := json.Marshal(baseline)
			if err != nil {
				return err
			}
			if err := tx.Write(recordPrefix+baseline.Document, payload); err != nil {
				return err
			}
		}
		return nil
	})
}

func openOrInit(repositoryRoot string) (*ddrepo.Repository, error) {
	_, err := os.Stat(filepath.Join(repositoryRoot, ".ddocs"))
	if os.IsNotExist(err) {
		return ddrepo.Init(repositoryRoot)
	}
	if err != nil {
		return nil, err
	}
	return ddrepo.Open(repositoryRoot)
}
