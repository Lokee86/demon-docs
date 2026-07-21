package validationcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

const (
	EngineVersion = "validation-engine-v1"
	prefix        = "validation-cache/"
	schemaVersion = 1
)

// Entry is the durable identity and clean-result state for one document.
// The two clean flags are merged by frontmatter and document-policy builders.
type Entry struct {
	SchemaVersion         int            `json:"schema_version"`
	Path                  string         `json:"path"`
	ContentSHA256         string         `json:"content_sha256"`
	EngineVersion         string         `json:"engine_version"`
	FrontmatterPolicyHash string         `json:"frontmatter_policy_hash"`
	EffectiveSchemaHash   string         `json:"effective_schema_hash"`
	ImmutableSnapshotHash string         `json:"immutable_snapshot_hash"`
	DocumentID            string         `json:"document_id,omitempty"`
	DocumentType          string         `json:"document_type,omitempty"`
	SchemaName            string         `json:"schema_name,omitempty"`
	ImmutableValues       map[string]any `json:"immutable_values,omitempty"`
	FrontmatterClean      bool           `json:"frontmatter_clean"`
	FormatClean           bool           `json:"format_clean"`
}

type Store struct {
	repository *ddrepo.Repository
	entries    map[string]Entry
	dirty      map[string]Entry
	deleted    map[string]bool
}

type SchemaHasher struct {
	repoRoot string
	format   config.Format
	mu       sync.Mutex
	hashes   map[string]string
}

func NewSchemaHasher(repoRoot string, format config.Format) *SchemaHasher {
	return &SchemaHasher{repoRoot: repoRoot, format: format, hashes: map[string]string{}}
}

func (h *SchemaHasher) Effective(schemaName, documentID string) string {
	if h == nil {
		return ""
	}
	key := strings.TrimSpace(schemaName) + "\x00" + strings.TrimSpace(documentID)
	h.mu.Lock()
	if hash, ok := h.hashes[key]; ok {
		h.mu.Unlock()
		return hash
	}
	h.mu.Unlock()

	hash := effectiveSchemaHash(h.repoRoot, h.format, schemaName, documentID)
	h.mu.Lock()
	if existing, ok := h.hashes[key]; ok {
		h.mu.Unlock()
		return existing
	}
	h.hashes[key] = hash
	h.mu.Unlock()
	return hash
}

func Open(repoRoot string) (*Store, error) {
	store := &Store{entries: map[string]Entry{}, dirty: map[string]Entry{}, deleted: map[string]bool{}}
	privateState := filepath.Join(repoRoot, ".ddocs")
	repository, err := ddrepo.Open(privateState)
	if err != nil {
		if _, statErr := os.Stat(privateState); errors.Is(statErr, os.ErrNotExist) || missingPrivateState(repoRoot) {
			return store, nil
		}
		return nil, err
	}
	store.repository = repository
	tx, err := repository.Begin()
	if err != nil {
		return nil, err
	}
	names, err := tx.Names(prefix)
	if err != nil {
		return nil, err
	}
	for _, name := range names {
		data, err := tx.Read(name)
		if err != nil {
			return nil, err
		}
		var entry Entry
		if json.Unmarshal(data, &entry) != nil || entry.SchemaVersion != schemaVersion || entry.Path == "" {
			continue
		}
		store.entries[entry.Path] = entry
	}
	return store, nil
}

func (s *Store) Lookup(path, contentHash, frontmatterPolicyHash, schemaHash, immutableHash string) (Entry, bool) {
	if s == nil {
		return Entry{}, false
	}
	entry, ok := s.entries[NormalizePath(path)]
	if !ok || entry.SchemaVersion != schemaVersion || entry.Path != NormalizePath(path) ||
		entry.ContentSHA256 != contentHash || entry.EngineVersion != EngineVersion ||
		entry.FrontmatterPolicyHash != frontmatterPolicyHash || entry.EffectiveSchemaHash != schemaHash ||
		entry.ImmutableSnapshotHash != immutableHash {
		return Entry{}, false
	}
	return entry, true
}

// Candidate returns a path/content/version/policy match before the caller has
// re-read the selected schema. It is used to recover the selection metadata
// needed to compute the current effective schema hash without parsing Markdown.
func (s *Store) Candidate(path, contentHash, frontmatterPolicyHash string) (Entry, bool) {
	if s == nil {
		return Entry{}, false
	}
	entry, ok := s.entries[NormalizePath(path)]
	if !ok || entry.SchemaVersion != schemaVersion || entry.Path != NormalizePath(path) ||
		entry.ContentSHA256 != contentHash || entry.EngineVersion != EngineVersion ||
		entry.FrontmatterPolicyHash != frontmatterPolicyHash {
		return Entry{}, false
	}
	return entry, true
}

func (s *Store) Merge(entry Entry) {
	if s == nil || !entry.FrontmatterClean && !entry.FormatClean {
		return
	}
	entry.SchemaVersion = schemaVersion
	entry.Path = NormalizePath(entry.Path)
	previous, exists := s.entries[entry.Path]
	if exists && sameIdentity(previous, entry) {
		entry.FrontmatterClean = entry.FrontmatterClean || previous.FrontmatterClean
		entry.FormatClean = entry.FormatClean || previous.FormatClean
		if entry.DocumentID == "" {
			entry.DocumentID = previous.DocumentID
		}
		if entry.DocumentType == "" {
			entry.DocumentType = previous.DocumentType
		}
		if entry.SchemaName == "" {
			entry.SchemaName = previous.SchemaName
		}
		if entry.ImmutableValues == nil {
			entry.ImmutableValues = previous.ImmutableValues
		}
	}
	s.entries[entry.Path] = entry
	if exists && reflect.DeepEqual(previous, entry) {
		delete(s.dirty, entry.Path)
		return
	}
	s.dirty[entry.Path] = entry
}

// Retain drops cache records for documents that are no longer in the active
// validation scope. Deletion is published with the next Save call.
func (s *Store) Retain(paths []string) {
	if s == nil {
		return
	}
	active := make(map[string]bool, len(paths))
	for _, path := range paths {
		active[NormalizePath(path)] = true
	}
	for path := range s.entries {
		if active[path] {
			continue
		}
		delete(s.entries, path)
		delete(s.dirty, path)
		s.deleted[path] = true
	}
}

func (s *Store) Save() error {
	if s == nil || s.repository == nil || len(s.dirty) == 0 && len(s.deleted) == 0 {
		return nil
	}
	deleted := make([]string, 0, len(s.deleted))
	for path := range s.deleted {
		deleted = append(deleted, path)
	}
	paths := make([]string, 0, len(s.dirty))
	for path := range s.dirty {
		paths = append(paths, path)
	}
	sort.Strings(deleted)
	sort.Strings(paths)
	if err := s.repository.TransactionRetry(4, func(tx *ddrepo.Transaction) error {
		for _, path := range deleted {
			if err := tx.Delete(recordName(path)); err != nil {
				return err
			}
		}
		for _, path := range paths {
			data, err := json.Marshal(s.dirty[path])
			if err != nil {
				return err
			}
			if err := tx.Write(recordName(path), data); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	s.dirty = map[string]Entry{}
	s.deleted = map[string]bool{}
	return nil
}

func Hash(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// FrontmatterPolicyHash covers frontmatter rules and the format-selection and
// index defaults that can change the effective frontmatter schema.
func FrontmatterPolicyHash(cfg config.Config) string {
	return Hash(struct {
		Frontmatter config.Frontmatter
		IndexFile   string
		Format      config.Format
	}{cfg.Frontmatter, cfg.IndexFile, cfg.Format})
}

func ContentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func NormalizePath(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	return path
}

// EffectiveSchemaHash fingerprints the selected shared and document-specific
// schema sources. Missing shared sources are built-in engine inputs and are
// invalidated by EngineVersion; missing local sources are explicit selections.
func EffectiveSchemaHash(repoRoot string, format config.Format, schemaName, documentID string) string {
	return effectiveSchemaHash(repoRoot, format, schemaName, documentID)
}

func effectiveSchemaHash(repoRoot string, format config.Format, schemaName, documentID string) string {
	if !format.Enabled || strings.TrimSpace(schemaName) == "" {
		return Hash("format-disabled")
	}
	sharedPath := resolveSchemaPath(repoRoot, format.SchemaDir, schemaName+".toml")
	localPath := resolveSchemaPath(repoRoot, format.DocumentSchemaDir, strings.TrimSpace(documentID)+".toml")
	shared := readSchemaSource(sharedPath, "builtin:"+strings.TrimSpace(schemaName))
	local := "none"
	if strings.TrimSpace(documentID) != "" {
		local = readSchemaSource(localPath, "missing-local:"+strings.TrimSpace(documentID))
	}
	return Hash(struct {
		SchemaName string
		DocumentID string
		Shared     string
		Local      string
	}{strings.TrimSpace(schemaName), strings.TrimSpace(documentID), shared, local})
}

func sameIdentity(left, right Entry) bool {
	return left.Path == NormalizePath(right.Path) && left.ContentSHA256 == right.ContentSHA256 &&
		left.EngineVersion == right.EngineVersion && left.FrontmatterPolicyHash == right.FrontmatterPolicyHash &&
		left.EffectiveSchemaHash == right.EffectiveSchemaHash && left.ImmutableSnapshotHash == right.ImmutableSnapshotHash
}

func recordName(path string) string { return prefix + Hash(NormalizePath(path)) }

func resolveSchemaPath(repoRoot, configured, name string) string {
	if filepath.IsAbs(configured) {
		return filepath.Join(configured, name)
	}
	return filepath.Join(repoRoot, filepath.FromSlash(configured), name)
}

func readSchemaSource(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return ContentHash(data)
}

func missingPrivateState(repoRoot string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, ".ddocs", "HEAD"))
	return errors.Is(err, os.ErrNotExist)
}
