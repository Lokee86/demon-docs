package reconcile

import (
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/diagnostics"
	"github.com/Lokee86/demon-docs/internal/model"
)

const (
	diagnosticIndexMissing    = "indexes.missing"
	diagnosticIndexOutdated   = "indexes.out_of_date"
	diagnosticIndexStaleEntry = "indexes.stale_entry"
)

func indexUpdateDiagnostics(repositoryRoot string, updates []model.FileUpdate) []diagnostics.Diagnostic {
	items := make([]diagnostics.Diagnostic, 0, len(updates))
	for _, update := range updates {
		code := diagnosticIndexOutdated
		message := "Documentation index is out of date"
		if update.OldText == nil {
			code = diagnosticIndexMissing
			message = "Documentation index is missing"
		}
		items = append(items, diagnostics.Diagnostic{
			Code:      code,
			Severity:  diagnostics.SeverityWarning,
			Subsystem: "indexes",
			Message:   message,
			Path:      diagnosticPath(repositoryRoot, update.Path),
		})
	}
	return items
}

func staleIndexDiagnostic(repositoryRoot, indexPath, section, target string) diagnostics.Diagnostic {
	return diagnostics.Diagnostic{
		Code:      diagnosticIndexStaleEntry,
		Severity:  diagnostics.SeverityWarning,
		Subsystem: "indexes",
		Message:   "Documentation index contains a stale managed entry",
		Path:      diagnosticPath(repositoryRoot, indexPath),
		Section:   section,
		Target:    target,
	}
}

func diagnosticPath(repositoryRoot, path string) string {
	relative, err := filepath.Rel(repositoryRoot, path)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	return filepath.ToSlash(filepath.Clean(relative))
}
