package app

import (
	"encoding/json"
	"io"

	"github.com/Lokee86/demon-docs/internal/diagnostics"
	"github.com/Lokee86/demon-docs/internal/documentpolicy"
	"github.com/Lokee86/demon-docs/internal/frontmatter"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/model"
)

func writeDiagnosticReport(out io.Writer, command string, exitCode int, indexes model.ReconcileResult, frontmatterPlan frontmatter.Plan, formatPlan documentpolicy.Plan, plan links.Plan, orphanDocuments []string) error {
	items := make([]diagnostics.Diagnostic, 0, len(indexes.Diagnostics)+len(frontmatterPlan.Diagnostics)+len(formatPlan.Diagnostics)+len(plan.Diagnostics)+len(orphanDocuments))
	items = append(items, indexes.Diagnostics...)
	items = append(items, machineFrontmatterDiagnostics(frontmatterPlan.Diagnostics)...)
	items = append(items, machineFormatDiagnostics(formatPlan.Diagnostics)...)
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

func machineFormatDiagnostics(source []documentpolicy.Diagnostic) []diagnostics.Diagnostic {
	items := make([]diagnostics.Diagnostic, 0, len(source))
	for _, diagnostic := range source {
		severity := diagnostics.SeverityError
		if diagnostic.Warning {
			severity = diagnostics.SeverityWarning
		}
		items = append(items, diagnostics.Diagnostic{
			Code:      diagnostic.Code,
			Severity:  severity,
			Subsystem: "format",
			Message:   diagnostic.Message,
			Path:      diagnostic.Path,
			Section:   diagnostic.Section,
			Options:   diagnostic.Options,
		})
	}
	return items
}

func machineFrontmatterDiagnostics(source []frontmatter.Diagnostic) []diagnostics.Diagnostic {
	items := make([]diagnostics.Diagnostic, 0, len(source))
	for _, diagnostic := range source {
		severity := diagnostics.SeverityError
		if diagnostic.Warning {
			severity = diagnostics.SeverityWarning
		}
		items = append(items, diagnostics.Diagnostic{
			Code:      diagnostic.Code,
			Severity:  severity,
			Subsystem: "frontmatter",
			Message:   diagnostic.Message,
			Path:      diagnostic.Path,
			Field:     diagnostic.Field,
		})
	}
	return items
}
