package documentpolicy

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

const schemaHistoryPrefix = "document-policy/schema/"

func loadSchemaHistory(repoRoot, name string) (Schema, bool, error) {
	if !privateStateExists(repoRoot) {
		return Schema{}, false, nil
	}
	repository, err := ddrepo.Open(repoRoot)
	if err != nil {
		return Schema{}, false, fmt.Errorf("open document-policy state: %w", err)
	}
	tx, err := repository.Begin()
	if err != nil {
		return Schema{}, false, err
	}
	fingerprint, err := tx.Read(schemaLatestKey(name))
	if errors.Is(err, ddrepo.ErrRecordAbsent) {
		return loadLegacySchemaHistory(tx, name)
	}
	if err != nil {
		return Schema{}, false, err
	}
	return readSchemaSnapshot(tx, name, strings.TrimSpace(string(fingerprint)))
}

func loadSchemaSnapshot(repoRoot, name, fingerprint string) (Schema, bool, error) {
	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return Schema{}, false, nil
	}
	if !validSchemaFingerprint(fingerprint) {
		return Schema{}, false, fmt.Errorf("invalid shared-schema fingerprint %q", fingerprint)
	}
	if !privateStateExists(repoRoot) {
		return Schema{}, false, nil
	}
	repository, err := ddrepo.Open(repoRoot)
	if err != nil {
		return Schema{}, false, fmt.Errorf("open document-policy state: %w", err)
	}
	tx, err := repository.Begin()
	if err != nil {
		return Schema{}, false, err
	}
	return readSchemaSnapshot(tx, name, fingerprint)
}

func loadLegacySchemaHistory(tx *ddrepo.Transaction, name string) (Schema, bool, error) {
	data, err := tx.Read(schemaHistoryPrefix + name)
	if errors.Is(err, ddrepo.ErrRecordAbsent) {
		return Schema{}, false, nil
	}
	if err != nil {
		return Schema{}, false, err
	}
	return decodeSchema(data, name)
}

func readSchemaSnapshot(tx *ddrepo.Transaction, name, fingerprint string) (Schema, bool, error) {
	if strings.TrimSpace(fingerprint) == "" {
		return Schema{}, false, nil
	}
	data, err := tx.Read(schemaSnapshotKey(name, fingerprint))
	if errors.Is(err, ddrepo.ErrRecordAbsent) {
		return Schema{}, false, nil
	}
	if err != nil {
		return Schema{}, false, err
	}
	return decodeSchema(data, name)
}

func decodeSchema(data []byte, name string) (Schema, bool, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return Schema{}, false, fmt.Errorf("decode prior document schema %q: %w", name, err)
	}
	return schema, true, nil
}

func saveSchemaHistory(repoRoot string, schemas map[string]Schema) error {
	if len(schemas) == 0 || !privateStateExists(repoRoot) {
		return nil
	}
	repository, err := ddrepo.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("open document-policy state: %w", err)
	}
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return repository.Transaction(func(tx *ddrepo.Transaction) error {
		for _, name := range names {
			schema := schemas[name]
			fingerprint := Fingerprint(schema)
			if err := writeSchemaSnapshot(tx, name, fingerprint, schema); err != nil {
				return err
			}
			if err := tx.Write(schemaLatestKey(name), []byte(fingerprint)); err != nil {
				return err
			}
		}
		return nil
	})
}

func saveSchemaSnapshot(repoRoot, name string, schema Schema) error {
	if !privateStateExists(repoRoot) {
		return nil
	}
	repository, err := ddrepo.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("open document-policy state: %w", err)
	}
	fingerprint := Fingerprint(schema)
	return repository.Transaction(func(tx *ddrepo.Transaction) error {
		return writeSchemaSnapshot(tx, name, fingerprint, schema)
	})
}

func writeSchemaSnapshot(tx *ddrepo.Transaction, name, fingerprint string, schema Schema) error {
	data, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	return tx.Write(schemaSnapshotKey(name, fingerprint), data)
}

func schemaLatestKey(name string) string {
	return schemaHistoryPrefix + name + "/latest"
}

func schemaSnapshotKey(name, fingerprint string) string {
	return schemaHistoryPrefix + name + "/snapshots/" + fingerprint
}

func validSchemaFingerprint(fingerprint string) bool {
	if len(fingerprint) != 64 {
		return false
	}
	_, err := hex.DecodeString(fingerprint)
	return err == nil
}

func privateStateExists(repoRoot string) bool {
	info, err := os.Stat(filepath.Join(repoRoot, ".ddocs", "HEAD"))
	return err == nil && !info.IsDir()
}
