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
	schemaVersion = 2
)

// ErrScopedReuseUnavailable means that a scoped validation pass cannot prove
// the untouched active documents are represented by reusable clean entries.
var ErrScopedReuseUnavailable = errors.New("scoped validation cache reuse unavailable")

// Entry is the durable identity and clean-result state for one document.
// Frontmatter and format keep independent identity and policy fields so one
// subsystem can be reused when the other subsystem's source surface changes.
type Entry struct {
	SchemaVersion             int            `json:"schema_version"`
	Path                      string         `json:"path"`
	ContentSHA256             string         `json:"content_sha256"`
	EngineVersion             string         `json:"engine_version"`
	FrontmatterIdentitySHA256 string         `json:"frontmatter_identity_sha256,omitempty"`
	FrontmatterPolicyHash     string         `json:"frontmatter_policy_hash,omitempty"`
	FrontmatterSchemaHash     string         `json:"frontmatter_schema_hash,omitempty"`
	ImmutableSnapshotHash     string         `json:"immutable_snapshot_hash,omitempty"`
	FormatIdentitySHA256      string         `json:"format_identity_sha256,omitempty"`
	FormatPolicyHash          string         `json:"format_policy_hash,omitempty"`
	FormatSchemaHash          string         `json:"format_schema_hash,omitempty"`
	DocumentID                string         `json:"document_id,omitempty"`
	DocumentType              string         `json:"document_type,omitempty"`
	SchemaName                string         `json:"schema_name,omitempty"`
	ImmutableValues           map[string]any `json:"immutable_values,omitempty"`
	FrontmatterClean          bool           `json:"frontmatter_clean"`
	FormatClean               bool           `json:"format_clean"`
}

type Store struct {
	mu         sync.RWMutex
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

func (s *Store) LookupFrontmatter(path, identityHash, policyHash, schemaHash, immutableHash string) (Entry, bool) {
	entry, ok := s.CandidateFrontmatter(path, identityHash, policyHash)
	if !ok || entry.FrontmatterSchemaHash != schemaHash || entry.ImmutableSnapshotHash != immutableHash {
		return Entry{}, false
	}
	return entry, true
}

// CandidateFrontmatter returns a frontmatter identity and policy match before
// the caller re-reads the selected schema and immutable-state snapshot.
func (s *Store) CandidateFrontmatter(path, identityHash, policyHash string) (Entry, bool) {
	if s == nil {
		return Entry{}, false
	}
	normalized := NormalizePath(path)
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[normalized]
	if !ok || entry.SchemaVersion != schemaVersion || entry.Path != normalized ||
		entry.EngineVersion != EngineVersion || !entry.FrontmatterClean ||
		entry.FrontmatterIdentitySHA256 != identityHash || entry.FrontmatterPolicyHash != policyHash {
		return Entry{}, false
	}
	return cloneEntry(entry), true
}

// LookupPath returns the current cache entry for path without requiring a
// source identity. Scoped builders use it for documents they deliberately do
// not read. Callers must still verify the subsystem-specific clean flag and
// policy inputs that apply to their validation pass.
func (s *Store) LookupPath(path string) (Entry, bool) {
	if s == nil {
		return Entry{}, false
	}
	normalized := NormalizePath(path)
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[normalized]
	if !ok || entry.SchemaVersion != schemaVersion || entry.Path != normalized || entry.EngineVersion != EngineVersion {
		return Entry{}, false
	}
	return cloneEntry(entry), true
}

func (s *Store) LookupFormat(path, identityHash, policyHash, schemaHash string) (Entry, bool) {
	entry, ok := s.CandidateFormat(path, identityHash, policyHash)
	if !ok || entry.FormatSchemaHash != schemaHash {
		return Entry{}, false
	}
	return entry, true
}

// CandidateFormat returns a format identity and policy match before the caller
// re-reads the selected shared and document-specific schema sources.
func (s *Store) CandidateFormat(path, identityHash, policyHash string) (Entry, bool) {
	if s == nil {
		return Entry{}, false
	}
	normalized := NormalizePath(path)
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[normalized]
	if !ok || entry.SchemaVersion != schemaVersion || entry.Path != normalized ||
		entry.EngineVersion != EngineVersion || !entry.FormatClean ||
		entry.FormatIdentitySHA256 != identityHash || entry.FormatPolicyHash != policyHash {
		return Entry{}, false
	}
	return cloneEntry(entry), true
}

func (s *Store) Merge(entry Entry) {
	if s == nil || !entry.FrontmatterClean && !entry.FormatClean {
		return
	}
	entry.SchemaVersion = schemaVersion
	entry.Path = NormalizePath(entry.Path)
	entry.EngineVersion = EngineVersion
	entry = cloneEntry(entry)
	s.mu.Lock()
	defer s.mu.Unlock()

	previous, exists := s.entries[entry.Path]
	merged := Entry{
		SchemaVersion: schemaVersion,
		Path:          entry.Path,
		EngineVersion: EngineVersion,
	}
	if exists && previous.SchemaVersion == schemaVersion && previous.EngineVersion == EngineVersion {
		merged = cloneEntry(previous)
	}
	if entry.ContentSHA256 != "" {
		merged.ContentSHA256 = entry.ContentSHA256
	}
	if entry.FrontmatterClean {
		merged.FrontmatterIdentitySHA256 = entry.FrontmatterIdentitySHA256
		merged.FrontmatterPolicyHash = entry.FrontmatterPolicyHash
		merged.FrontmatterSchemaHash = entry.FrontmatterSchemaHash
		merged.ImmutableSnapshotHash = entry.ImmutableSnapshotHash
		merged.DocumentID = entry.DocumentID
		merged.DocumentType = entry.DocumentType
		merged.SchemaName = entry.SchemaName
		merged.ImmutableValues = cloneValues(entry.ImmutableValues)
		merged.FrontmatterClean = true
	}
	if entry.FormatClean {
		merged.FormatIdentitySHA256 = entry.FormatIdentitySHA256
		merged.FormatPolicyHash = entry.FormatPolicyHash
		merged.FormatSchemaHash = entry.FormatSchemaHash
		merged.DocumentID = entry.DocumentID
		merged.DocumentType = entry.DocumentType
		merged.SchemaName = entry.SchemaName
		merged.FormatClean = true
	}

	s.entries[entry.Path] = merged
	if exists && reflect.DeepEqual(previous, merged) {
		delete(s.dirty, entry.Path)
		return
	}
	s.dirty[entry.Path] = merged
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.repository == nil || len(s.dirty) == 0 && len(s.deleted) == 0 {
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

// FrontmatterPolicyHash covers frontmatter rules and only the format-selection
// inputs that can change the effective frontmatter schema. Format evaluation,
// schema locations, and invalidation thresholds belong to separate identities.
func FrontmatterPolicyHash(cfg config.Config) string {
	return Hash(struct {
		Frontmatter config.Frontmatter
		IndexFile   string
		Selection   struct {
			Enabled       bool
			DefaultSchema string
			PathRules     []config.FormatPathRule
		}
	}{
		Frontmatter: cfg.Frontmatter,
		IndexFile:   cfg.IndexFile,
		Selection: struct {
			Enabled       bool
			DefaultSchema string
			PathRules     []config.FormatPathRule
		}{cfg.Format.Enabled, cfg.Format.DefaultSchema, cfg.Format.PathRules},
	})
}

// FormatPolicyHash covers only configuration used while parsing frontmatter
// for schema selection and evaluating document-format rules.
func FormatPolicyHash(cfg config.Config) string {
	return Hash(struct {
		AllowedFormats []string
		Format         config.Format
	}{cfg.Frontmatter.AllowedFormats, cfg.Format})
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

func cloneEntry(entry Entry) Entry {
	entry.ImmutableValues = cloneValues(entry.ImmutableValues)
	return entry
}

func cloneValues(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
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
