package codemap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

const DatasetSchemaVersion = 1

type ResolutionStatus string

const (
	ResolutionResolved         ResolutionStatus = "resolved"
	ResolutionMissing          ResolutionStatus = "missing"
	ResolutionOutsideRepo      ResolutionStatus = "outside_repository"
	ResolutionKindMismatch     ResolutionStatus = "kind_mismatch"
	ResolutionSymbolUnverified ResolutionStatus = "symbol_unverified"
	ResolutionPatternResolved  ResolutionStatus = "pattern_resolved"
	ResolutionPatternMissing   ResolutionStatus = "pattern_missing"
	ResolutionUnsupported      ResolutionStatus = "unsupported"
)

type DocumentRecord struct {
	Path            string `json:"path"`
	Size            int64  `json:"size"`
	SHA256          string `json:"sha256"`
	EntryCount      int    `json:"entry_count"`
	DiagnosticCount int    `json:"diagnostic_count"`
}

type TargetMatch struct {
	Path   string `json:"path"`
	IsDir  bool   `json:"is_directory"`
	Size   int64  `json:"size,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

type TargetRecord struct {
	Status       ResolutionStatus `json:"status"`
	ResolvedPath string           `json:"resolved_path,omitempty"`
	Exists       bool             `json:"exists"`
	Size         int64            `json:"size,omitempty"`
	SHA256       string           `json:"sha256,omitempty"`
	Matches      []TargetMatch    `json:"matches,omitempty"`
}

type DatasetEntry struct {
	Entry      Entry        `json:"entry"`
	Resolution TargetRecord `json:"resolution"`
}

type Dataset struct {
	SchemaVersion int              `json:"schema_version"`
	Documents     []DocumentRecord `json:"documents"`
	Entries       []DatasetEntry   `json:"entries"`
	Diagnostics   []Diagnostic     `json:"diagnostics"`
}

// BuildDataset scans Markdown documents under docsRoot, extracts authored code
// maps, and resolves their targets against repositoryRoot. Output ordering and
// hashes depend only on repository content and the supplied format.
func BuildDataset(repositoryRoot, docsRoot string, format Format) (Dataset, error) {
	repositoryRoot, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Dataset{}, err
	}
	docsRoot, err = filepath.Abs(docsRoot)
	if err != nil {
		return Dataset{}, err
	}
	if !within(repositoryRoot, docsRoot) {
		return Dataset{}, fmt.Errorf("docs root %s is outside repository root %s", docsRoot, repositoryRoot)
	}
	if format.TargetBase == "" {
		format.TargetBase = TargetBaseRepository
	}
	policy, err := ignorepolicy.Load(repositoryRoot)
	if err != nil {
		return Dataset{}, err
	}

	dataset := Dataset{SchemaVersion: DatasetSchemaVersion}
	err = filepath.WalkDir(docsRoot, func(filePath string, item os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath != docsRoot {
			ignored, err := policy.Ignored(filePath, item.IsDir())
			if err != nil {
				return err
			}
			if ignored {
				if item.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if item.IsDir() || item.Type()&os.ModeSymlink != 0 || !strings.EqualFold(filepath.Ext(filePath), ".md") {
			return nil
		}

		source, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		documentPath, err := repositoryRelative(repositoryRoot, filePath)
		if err != nil {
			return err
		}
		extracted := Extract(documentPath, string(source), format)
		dataset.Documents = append(dataset.Documents, DocumentRecord{
			Path:            documentPath,
			Size:            int64(len(source)),
			SHA256:          digest(source),
			EntryCount:      len(extracted.Entries),
			DiagnosticCount: len(extracted.Diagnostics),
		})
		dataset.Diagnostics = append(dataset.Diagnostics, extracted.Diagnostics...)
		for _, entry := range extracted.Entries {
			resolution, err := resolveTarget(repositoryRoot, documentPath, entry, format)
			if err != nil {
				return err
			}
			dataset.Entries = append(dataset.Entries, DatasetEntry{Entry: entry, Resolution: resolution})
		}
		return nil
	})
	if err != nil {
		return Dataset{}, err
	}
	sortDataset(&dataset)
	return dataset, nil
}

func MarshalDataset(dataset Dataset) ([]byte, error) {
	encoded, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func ExportDataset(outputPath string, dataset Dataset) error {
	encoded, err := MarshalDataset(dataset)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, encoded, 0o644)
}

func resolveTarget(repositoryRoot, documentPath string, entry Entry, format Format) (TargetRecord, error) {
	if entry.Kind == TargetSymbol && !strings.Contains(entry.Target, "#") && !strings.Contains(entry.Target, "::") {
		return TargetRecord{Status: ResolutionUnsupported}, nil
	}
	baseTarget, hasSymbol := targetFilePart(entry.Target)
	if baseTarget == "" {
		return TargetRecord{Status: ResolutionUnsupported}, nil
	}
	base := repositoryRoot
	if format.TargetBase == TargetBaseDocument {
		base = filepath.Dir(filepath.Join(repositoryRoot, filepath.FromSlash(documentPath)))
	}
	candidate := filepath.Clean(filepath.Join(base, filepath.FromSlash(baseTarget)))
	if filepath.IsAbs(filepath.FromSlash(baseTarget)) {
		candidate = filepath.Clean(filepath.FromSlash(baseTarget))
	}
	if !within(repositoryRoot, candidate) {
		return TargetRecord{Status: ResolutionOutsideRepo}, nil
	}
	resolvedPath, err := repositoryRelative(repositoryRoot, candidate)
	if err != nil {
		return TargetRecord{}, err
	}
	if hasPattern(baseTarget) {
		return resolvePattern(repositoryRoot, candidate, resolvedPath)
	}
	info, err := os.Stat(candidate)
	if os.IsNotExist(err) {
		return TargetRecord{Status: ResolutionMissing, ResolvedPath: resolvedPath}, nil
	}
	if err != nil {
		return TargetRecord{}, err
	}
	record := TargetRecord{ResolvedPath: resolvedPath, Exists: true, Size: info.Size()}
	if !info.IsDir() {
		contents, err := os.ReadFile(candidate)
		if err != nil {
			return TargetRecord{}, err
		}
		record.SHA256 = digest(contents)
	}
	if hasSymbol {
		record.Status = ResolutionSymbolUnverified
		return record, nil
	}
	if entry.Kind == TargetDirectory && !info.IsDir() || entry.Kind == TargetFile && info.IsDir() {
		record.Status = ResolutionKindMismatch
		return record, nil
	}
	record.Status = ResolutionResolved
	return record, nil
}

func resolvePattern(repositoryRoot, patternPath, resolvedPattern string) (TargetRecord, error) {
	matches, err := filepath.Glob(patternPath)
	if err != nil {
		return TargetRecord{Status: ResolutionUnsupported, ResolvedPath: resolvedPattern}, nil
	}
	sort.Strings(matches)
	record := TargetRecord{
		Status:       ResolutionPatternMissing,
		ResolvedPath: resolvedPattern,
		Exists:       len(matches) > 0,
	}
	for _, match := range matches {
		if !within(repositoryRoot, match) {
			continue
		}
		info, err := os.Stat(match)
		if err != nil {
			return TargetRecord{}, err
		}
		relative, err := repositoryRelative(repositoryRoot, match)
		if err != nil {
			return TargetRecord{}, err
		}
		item := TargetMatch{Path: relative, IsDir: info.IsDir(), Size: info.Size()}
		if !info.IsDir() {
			contents, err := os.ReadFile(match)
			if err != nil {
				return TargetRecord{}, err
			}
			item.SHA256 = digest(contents)
		}
		record.Matches = append(record.Matches, item)
	}
	if len(record.Matches) > 0 {
		record.Status = ResolutionPatternResolved
		record.Exists = true
	}
	return record, nil
}

func hasPattern(target string) bool {
	return strings.ContainsAny(target, "*?[")
}

func targetFilePart(target string) (string, bool) {
	if strings.HasPrefix(target, "symbol:") {
		return "", true
	}
	if index := strings.Index(target, "#"); index > 0 {
		return target[:index], true
	}
	if index := strings.Index(target, "::"); index > 0 {
		left := target[:index]
		if strings.Contains(left, "/") || filepath.Ext(left) != "" {
			return left, true
		}
		return "", true
	}
	return target, false
}

func repositoryRelative(root, filePath string) (string, error) {
	relative, err := filepath.Rel(root, filePath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Clean(relative)), nil
}

func within(root, candidate string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func digest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}

func sortDataset(dataset *Dataset) {
	sort.Slice(dataset.Documents, func(i, j int) bool {
		return dataset.Documents[i].Path < dataset.Documents[j].Path
	})
	sort.Slice(dataset.Entries, func(i, j int) bool {
		left, right := dataset.Entries[i].Entry, dataset.Entries[j].Entry
		if left.DocumentPath != right.DocumentPath {
			return left.DocumentPath < right.DocumentPath
		}
		if left.Source.Line != right.Source.Line {
			return left.Source.Line < right.Source.Line
		}
		if left.Source.Column != right.Source.Column {
			return left.Source.Column < right.Source.Column
		}
		return left.Target < right.Target
	})
	sort.Slice(dataset.Diagnostics, func(i, j int) bool {
		left, right := dataset.Diagnostics[i], dataset.Diagnostics[j]
		if left.DocumentPath != right.DocumentPath {
			return left.DocumentPath < right.DocumentPath
		}
		if left.Source.Line != right.Source.Line {
			return left.Source.Line < right.Source.Line
		}
		return left.Source.Column < right.Source.Column
	})
}
