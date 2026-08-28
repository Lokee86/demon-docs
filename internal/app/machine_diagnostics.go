package app

import (
	"encoding/json"
	"io"

	"github.com/Lokee86/demon-docs/internal/diagnostics"
	"github.com/Lokee86/demon-docs/internal/links"
)

func writeLinkDiagnosticReport(out io.Writer, command string, exitCode int, plan links.Plan, orphanDocuments []string) error {
	items := make([]diagnostics.Diagnostic, 0, len(plan.Diagnostics)+len(orphanDocuments))
	items = append(items, plan.Diagnostics...)
	for _, path := range orphanDocuments {
		items = append(items, diagnostics.Diagnostic{
			Code:      "links.orphan_document",
			Severity:  diagnostics.SeverityWarning,
			Subsystem: "links",
			Message:   "Managed Markdown document has no meaningful inbound link",
			Path:      path,
		})
	}
	status := "passed"
	if exitCode != 0 {
		status = "failed"
	}
	report := diagnostics.Report{
		SchemaVersion: diagnostics.SchemaVersion,
		Command:       command,
		Status:        status,
		ExitCode:      exitCode,
		Diagnostics:   items,
	}
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(report)
}
