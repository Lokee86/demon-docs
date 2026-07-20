package frontmatter

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSupportsYAMLAndTOMLAndPreservesBody(t *testing.T) {
	for name, source := range map[string]string{
		FormatYAML: "---\r\nauthor: Human\r\n---\r\n# Title\r\n\r\nBody\r\n",
		FormatTOML: "+++\nauthor = \"Human\"\n+++\n# Title\n\nBody\n",
	} {
		t.Run(name, func(t *testing.T) {
			document, err := Parse(source, []string{FormatYAML, FormatTOML})
			if err != nil {
				t.Fatal(err)
			}
			if document.Format != name || document.Values["author"] != "Human" {
				t.Fatalf("unexpected document: %+v", document)
			}
			wantBody := source[strings.Index(source, "# Title"):]
			if document.Body != wantBody {
				t.Fatalf("body changed:\nwant %q\n got %q", wantBody, document.Body)
			}
		})
	}
}

func TestParseRejectsMalformedAndDisallowedFrontmatter(t *testing.T) {
	for _, source := range []string{
		"---\nauthor: [\n---\nBody\n",
		"---\nauthor: one\nauthor: two\n---\nBody\n",
		"---\nauthor: one\nBody\n",
	} {
		if _, err := Parse(source, []string{FormatYAML, FormatTOML}); err == nil {
			t.Fatalf("expected parse failure for %q", source)
		}
	}
	if _, err := Parse("+++\nauthor = \"Human\"\n+++\nBody\n", []string{FormatYAML}); err == nil {
		t.Fatal("expected disallowed TOML failure")
	}
}

func TestRenderPreservesSelectedFormatAndSortsFields(t *testing.T) {
	values := map[string]any{"zeta": "last", "alpha": []any{"one", "two"}}
	for _, format := range []string{FormatYAML, FormatTOML} {
		rendered, err := Render(format, values, "Body\n")
		if err != nil {
			t.Fatal(err)
		}
		delimiter := "---"
		if format == FormatTOML {
			delimiter = "+++"
		}
		if !strings.HasPrefix(rendered, delimiter+"\n") || !strings.HasSuffix(rendered, delimiter+"\nBody\n") {
			t.Fatalf("wrong %s framing: %q", format, rendered)
		}
		if strings.Index(rendered, "alpha") > strings.Index(rendered, "zeta") {
			t.Fatalf("fields not sorted: %q", rendered)
		}
	}
}

func TestRenderTOMLRoundTripsNestedUnknownValues(t *testing.T) {
	values := map[string]any{
		"author": "Human",
		"extra":  map[string]any{"enabled": true, "labels": []any{"one", "two"}},
	}
	rendered, err := Render(FormatTOML, values, "Body\n")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(rendered, []string{FormatTOML})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed.Values, normalizeMap(values)) {
		t.Fatalf("nested TOML value changed:\nwant %#v\n got %#v\n%s", normalizeMap(values), parsed.Values, rendered)
	}
}
