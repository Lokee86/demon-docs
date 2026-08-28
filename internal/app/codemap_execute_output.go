package app

import (
	"fmt"
	"io"

	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/codemaprun"
)

func writeCodemapSummary(out io.Writer, plan codemaprun.Plan) {
	for _, document := range plan.Documents {
		if document.Changed {
			fmt.Fprintf(out, "%s: added=%d removed=%d adopted=%t created=%t\n", document.Path, len(document.Added), len(document.Removed), document.SectionFound && !document.SectionCreated, document.SectionCreated)
		}
		if len(document.SemanticChanges) > 0 {
			fmt.Fprintf(out, "%s: semantic_stale=%d\n", document.Path, len(document.SemanticChanges))
		}
	}
}

func writeCodemapInspection(out io.Writer, plan codemaprun.Plan) {
	for _, document := range plan.Documents {
		status := "missing"
		if document.SectionCreated {
			status = "schema-created"
		} else if document.SectionFound {
			status = "existing"
		}
		fmt.Fprintf(out, "%s\n  section: %s\n  changed: %t\n", document.Path, status, document.Changed)
		for _, item := range document.Recommendations {
			decision := "context"
			if item.Tier == codemaprecommend.SuggestionTierHardLink {
				decision = "add"
			}
			if item.Declined {
				decision = "declined"
			}
			fmt.Fprintf(out, "  %s %s score=%.3f tier=%s role=%s\n", decision, item.Target, item.Score, item.Tier, item.Role)
			for _, evidence := range item.Evidence {
				fmt.Fprintf(out, "    evidence: %s\n", evidence)
			}
		}
		for _, target := range document.Removed {
			fmt.Fprintf(out, "  remove %s\n", target)
		}
		for _, change := range document.SemanticChanges {
			fmt.Fprintf(out, "  semantic-stale %s kind=%s", change.Target, change.Kind)
			if change.PreviousPath != "" {
				fmt.Fprintf(out, " previous=%s", change.PreviousPath)
			}
			if change.CurrentPath != "" {
				fmt.Fprintf(out, " current=%s", change.CurrentPath)
			}
			fmt.Fprintln(out)
		}
	}
}
