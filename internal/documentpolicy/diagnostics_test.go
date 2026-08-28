package documentpolicy

import "testing"

func TestFormatDiagnosticCodesAreAssignedByEnforcement(t *testing.T) {
	schema := Schema{
		Sections:          []Section{{ID: "purpose", Heading: "Purpose"}},
		UnknownSections:   "manual",
		DuplicateSections: "manual",
	}
	document := markdownDocument{Newline: "\n", Roots: []*markdownSection{{Heading: "Appendix", Level: 2}}}
	result := enforceDocument(document, schema, Schema{}, false, false)
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected format diagnostics")
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == "" {
			t.Fatalf("format diagnostic missing stable code: %#v", diagnostic)
		}
	}
}
