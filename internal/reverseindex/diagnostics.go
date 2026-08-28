package reverseindex

import (
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/diagnostics"
	"github.com/Lokee86/demon-docs/internal/model"
)

const (
	diagnosticTargetMissing      = "reverse_indexes.target_missing"
	diagnosticTargetOutsideRepo  = "reverse_indexes.target_outside_repository"
	diagnosticTargetKindMismatch = "reverse_indexes.target_kind_mismatch"
	diagnosticPatternMissing     = "reverse_indexes.pattern_missing"
	diagnosticTargetAmbiguous    = "reverse_indexes.target_ambiguous"
	diagnosticTargetUnresolved   = "reverse_indexes.target_unresolved"
	diagnosticTargetUnavailable  = "reverse_indexes.target_unavailable"
	diagnosticSymbolProjection   = "reverse_indexes.symbol_projection_error"
	diagnosticIndexMissing       = "reverse_indexes.index_missing"
	diagnosticIndexOutOfDate     = "reverse_indexes.index_out_of_date"
	diagnosticOrphanCodeFile     = "reverse_indexes.orphan_code_file"
)

func targetResolutionDiagnostic(entry codemap.Entry, record codemap.TargetRecord) diagnostics.Diagnostic {
	code := diagnosticTargetUnresolved
	switch record.Status {
	case codemap.ResolutionMissing:
		code = diagnosticTargetMissing
	case codemap.ResolutionOutsideRepo:
		code = diagnosticTargetOutsideRepo
	case codemap.ResolutionKindMismatch:
		code = diagnosticTargetKindMismatch
	case codemap.ResolutionPatternMissing:
		code = diagnosticPatternMissing
	case codemap.ResolutionAmbiguous:
		code = diagnosticTargetAmbiguous
	}
	return diagnostics.Diagnostic{Code: code, Severity: diagnostics.SeverityError, Subsystem: "reverse_indexes", Message: "Codemap target cannot be projected into the reverse index", Path: filepath.ToSlash(filepath.Clean(entry.DocumentPath)), Line: entry.Source.Line, Target: entry.Target, Candidates: append([]string(nil), record.Candidates...)}
}

func targetErrorDiagnostic(entry codemap.Entry, code, message string) diagnostics.Diagnostic {
	return diagnostics.Diagnostic{Code: code, Severity: diagnostics.SeverityError, Subsystem: "reverse_indexes", Message: message, Path: filepath.ToSlash(filepath.Clean(entry.DocumentPath)), Line: entry.Source.Line, Target: entry.Target}
}

func reverseIndexUpdateDiagnostics(repositoryRoot string, updates []model.FileUpdate) []diagnostics.Diagnostic {
	items := make([]diagnostics.Diagnostic, 0, len(updates))
	for _, update := range updates {
		code := diagnosticIndexOutOfDate
		message := "Reverse index is out of date"
		if update.OldText == nil {
			code = diagnosticIndexMissing
			message = "Reverse index is missing"
		}
		items = append(items, diagnostics.Diagnostic{Code: code, Severity: diagnostics.SeverityWarning, Subsystem: "reverse_indexes", Message: message, Path: reverseDiagnosticPath(repositoryRoot, update.Path)})
	}
	return items
}

func reverseOrphanDiagnostics(orphans []string) []diagnostics.Diagnostic {
	items := make([]diagnostics.Diagnostic, 0, len(orphans))
	for _, path := range orphans {
		items = append(items, diagnostics.Diagnostic{Code: diagnosticOrphanCodeFile, Severity: diagnostics.SeverityWarning, Subsystem: "reverse_indexes", Message: "In-scope code file has no resolved authored documentation target", Path: filepath.ToSlash(filepath.Clean(path))})
	}
	return items
}

func sortReverseDiagnostics(items []diagnostics.Diagnostic) {
	sort.Slice(items, func(i, j int) bool {
		left, right := items[i], items[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Target < right.Target
	})
}

func reverseDiagnosticPath(repositoryRoot, path string) string {
	relative, err := filepath.Rel(repositoryRoot, path)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	return filepath.ToSlash(filepath.Clean(relative))
}
