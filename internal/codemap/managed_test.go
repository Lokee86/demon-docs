package codemap

import (
	"strings"
	"testing"
)

type fixedSchema struct {
	placement SectionPlacement
	required  bool
}

func (schema fixedSchema) CodemapSection(string, string) (SectionPlacement, bool, error) {
	return schema.placement, schema.required, nil
}

func TestReconcileManagedAdoptsWholeExistingSection(t *testing.T) {
	source := "# Runtime\n\n## Implementation Map\n\n- `src/existing.go`\n\nNotes about ownership.\n\n## Notes\n\nKeep me.\n"
	format := DefaultFormat()
	format.SectionHeadings = []string{"Implementation Map"}
	result, err := ReconcileManaged("docs/runtime.md", source, format, "ddocs", ManagedUpdate{AddTargets: []string{"src/new.go"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Implementation Map\n\n<!-- ddocs:codemap:start -->",
		"- `src/existing.go`",
		"Notes about ownership.",
		"- `src/new.go`",
		"<!-- ddocs:codemap:end -->\n\n## Notes",
	} {
		if !strings.Contains(result.Text, want) {
			t.Fatalf("missing %q:\n%s", want, result.Text)
		}
	}
	if len(result.Added) != 1 || result.Added[0] != "src/new.go" || result.SectionCreated {
		t.Fatalf("unexpected result: %#v", result)
	}
	second, err := ReconcileManaged("docs/runtime.md", result.Text, format, "ddocs", ManagedUpdate{AddTargets: []string{"src/new.go"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Text != result.Text || len(second.Added) != 0 {
		t.Fatalf("reconcile was not idempotent:\n%s", second.Text)
	}
}

func TestReconcileManagedSkipsMissingSectionWithoutSchema(t *testing.T) {
	source := "# Runtime\n\nNo map.\n"
	result, err := ReconcileManaged("docs/runtime.md", source, DefaultFormat(), "ddocs", ManagedUpdate{AddTargets: []string{"src/new.go"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != source || result.SectionFound {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestReconcileManagedCreatesOnlySchemaRequiredSection(t *testing.T) {
	source := "# Runtime\n\nIntro.\n"
	schema := fixedSchema{required: true, placement: SectionPlacement{Heading: "Source Guide", Level: 2, Offset: len(source)}}
	result, err := ReconcileManaged("docs/runtime.md", source, DefaultFormat(), "ddocs", ManagedUpdate{AddTargets: []string{"src/new.go"}}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if !result.SectionCreated || !strings.Contains(result.Text, "## Source Guide\n\n<!-- ddocs:codemap:start -->") {
		t.Fatalf("schema section was not created:\n%s", result.Text)
	}
}

func TestReconcileManagedPreservesFencedCodemapStyle(t *testing.T) {
	source := "# Runtime\n\n## Code Map\n\n```text\nsrc/existing.go\n```\n\n## Notes\n"
	result, err := ReconcileManaged("docs/runtime.md", source, DefaultFormat(), "ddocs", ManagedUpdate{AddTargets: []string{"src/new.go"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "```text\nsrc/existing.go\nsrc/new.go\n```"
	if !strings.Contains(result.Text, want) {
		t.Fatalf("fenced style was not preserved:\n%s", result.Text)
	}
	if strings.Contains(result.Text, "- `src/new.go`") {
		t.Fatalf("fenced codemap received a redundant bullet list:\n%s", result.Text)
	}
}

func TestReconcileManagedUnifiesExistingPartialManagedRegion(t *testing.T) {
	source := "# Runtime\n\n## Code Map\n\n- `src/authored.go`\n\n<!-- ddocs:codemap:start -->\n- `src/generated.go`\n<!-- ddocs:codemap:end -->\n"
	result, err := ReconcileManaged("docs/runtime.md", source, DefaultFormat(), "ddocs", ManagedUpdate{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(result.Text, "ddocs:codemap:start") != 1 || strings.Count(result.Text, "ddocs:codemap:end") != 1 {
		t.Fatalf("markers were not unified:\n%s", result.Text)
	}
	start := strings.Index(result.Text, "ddocs:codemap:start")
	authored := strings.Index(result.Text, "src/authored.go")
	generated := strings.Index(result.Text, "src/generated.go")
	end := strings.Index(result.Text, "ddocs:codemap:end")
	if !(start < authored && authored < generated && generated < end) {
		t.Fatalf("whole section was not adopted:\n%s", result.Text)
	}
}

func TestReconcileManagedRemovesOnlySelectedEntry(t *testing.T) {
	source := "# Runtime\n\n## Code Map\n\n- `src/keep.go`\n- `src/remove.go`\n"
	result, err := ReconcileManaged("docs/runtime.md", source, DefaultFormat(), "ddocs", ManagedUpdate{RemoveTargets: []string{"src/remove.go"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Text, "src/remove.go") || !strings.Contains(result.Text, "src/keep.go") {
		t.Fatalf("unexpected removal:\n%s", result.Text)
	}
	if len(result.Removed) != 1 || result.Removed[0] != "src/remove.go" {
		t.Fatalf("unexpected removed targets: %#v", result.Removed)
	}
}
