package codemapbench

import (
	"fmt"
	"io"
	"strings"
)

// FormatTextReport renders a deterministic human-readable benchmark report.
func FormatTextReport(report Report) string {
	report = canonicalReport(report)
	var output strings.Builder

	fmt.Fprintln(&output, "Codemap benchmark report")
	fmt.Fprintf(&output, "Schema: %d\n", ReportSchemaVersion)
	fmt.Fprintf(&output, "Seed: %s\n\n", report.Seed)
	fmt.Fprintf(&output, "Known links: %d\n", len(report.KnownLinks))
	fmt.Fprintf(&output, "Visible links: %d\n", len(report.VisibleLinks))
	fmt.Fprintf(&output, "Hidden links: %d\n", len(report.HiddenLinks))
	fmt.Fprintf(&output, "Recovered links: %d\n", len(report.RecoveredLinks))
	fmt.Fprintf(&output, "Missed links: %d\n", len(report.MissedLinks))
	fmt.Fprintf(&output, "Unmatched suggestions: %d\n", len(report.UnmatchedSuggestions))
	fmt.Fprintf(&output, "Already-linked suggestions: %d\n", len(report.AlreadyLinked))
	fmt.Fprintf(&output, "Duplicate suggestions: %d\n", len(report.DuplicateSuggestions))
	fmt.Fprintf(&output, "Invalid suggestions: %d\n", len(report.InvalidSuggestions))
	fmt.Fprintf(&output, "Raw suggestions: %d\n", report.RawSuggestionCount)
	fmt.Fprintf(&output, "Unique suggestions: %d\n", report.UniqueSuggestionCount)
	fmt.Fprintf(&output, "Precision: %.2f%%\n", report.Precision*100)
	fmt.Fprintf(&output, "Recall: %.2f%%\n", report.Recall*100)

	writeSuggestionSection(&output, "Recovered", report.RecoveredSuggestions)
	writeLinkSection(&output, "Missed", report.MissedLinks)
	writeSuggestionSection(&output, "Unmatched", report.UnmatchedSuggestions)
	writeSuggestionSection(&output, "Already linked", report.AlreadyLinked)
	writeSuggestionSection(&output, "Duplicates", report.DuplicateSuggestions)
	writeInvalidSection(&output, report.InvalidSuggestions)
	return output.String()
}

// WriteTextReport writes a deterministic human-readable benchmark report.
func WriteTextReport(writer io.Writer, report Report) error {
	_, err := io.WriteString(writer, FormatTextReport(report))
	return err
}

func writeLinkSection(output *strings.Builder, title string, links []Link) {
	if len(links) == 0 {
		return
	}
	fmt.Fprintf(output, "\n%s:\n", title)
	for _, link := range links {
		fmt.Fprintf(output, "- %s -> %s\n", link.Document, link.Target)
	}
}

func writeSuggestionSection(output *strings.Builder, title string, suggestions []Suggestion) {
	if len(suggestions) == 0 {
		return
	}
	fmt.Fprintf(output, "\n%s:\n", title)
	for _, suggestion := range suggestions {
		fmt.Fprintf(output, "- %s -> %s (score %.4f)\n", suggestion.Document, suggestion.Target, suggestion.Score)
		for _, evidence := range suggestion.Evidence {
			fmt.Fprintf(output, "  evidence: %s\n", evidence)
		}
	}
}

func writeInvalidSection(output *strings.Builder, suggestions []InvalidSuggestion) {
	if len(suggestions) == 0 {
		return
	}
	fmt.Fprintln(output, "\nInvalid:")
	for _, invalid := range suggestions {
		fmt.Fprintf(output, "- suggestion %d: %s", invalid.Index, invalid.Reason)
		if invalid.Suggestion.Document != "" || invalid.Suggestion.Target != "" {
			fmt.Fprintf(output, " (%s -> %s)", invalid.Suggestion.Document, invalid.Suggestion.Target)
		}
		fmt.Fprintln(output)
	}
}
