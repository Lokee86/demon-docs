package links

import "github.com/Lokee86/demon-docs/internal/diagnostics"

const (
	diagnosticStateUninitialized = "links.state_uninitialized"
	diagnosticBroken             = "links.broken"
	diagnosticFragmentMissing    = "links.fragment_missing"
	diagnosticAmbiguous          = "links.ambiguous"
	diagnosticUndefinedReference = "links.undefined_reference"
	diagnosticRepair             = "links.repair"
	diagnosticCaseRepair         = "links.case_repair"
	diagnosticRepairBlocked      = "links.repair_blocked"
	diagnosticRepairBlockStale   = "links.repair_block_stale"
	diagnosticRepairSelected     = "links.repair_selected"
)

func addDiagnostic(plan *Plan, human string, diagnostic diagnostics.Diagnostic) {
	plan.Messages = append(plan.Messages, human)
	plan.Diagnostics = append(plan.Diagnostics, diagnostic)
}

func linkDiagnostic(code, severity, message, path string, line, column int, target string) diagnostics.Diagnostic {
	return diagnostics.Diagnostic{
		Code:      code,
		Severity:  severity,
		Subsystem: "links",
		Message:   message,
		Path:      path,
		Line:      line,
		Column:    column,
		Target:    target,
	}
}
