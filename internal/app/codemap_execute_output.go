package app

import (
	"fmt"
	"io"

	"github.com/Lokee86/demon-docs/internal/codemaprun"
)

func writeCodemapSummary(out io.Writer, plan codemaprun.Plan) {
	for _, document := range plan.Documents {
		if !document.Changed {
			continue
		}
		fmt.Fprintf(out, "%s: added=%d removed=%d adopted=%t created=%t\n", document.Path, len(document.Added), len(document.Removed), document.SectionFound && !document.SectionCreated, document.SectionCreated)
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
			decision := "add"
			if item.Declined {
				decision = "declined"
			}
			fmt.Fprintf(out, "  %s %s score=%.3f tier=%s\n", decision, item.Target, item.Score, item.Tier)
			for _, evidence := range item.Evidence {
				fmt.Fprintf(out, "    evidence: %s\n", evidence)
			}
		}
		for _, target := range document.Removed {
			fmt.Fprintf(out, "  remove %s\n", target)
		}
	}
}
