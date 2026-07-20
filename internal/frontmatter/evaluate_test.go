package frontmatter

import (
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestEvaluateRepairsMissingSourcedFieldsButLeavesUnsourcedRequiredField(t *testing.T) {
	outcome := Evaluate("docs/guide.md", Document{Values: map[string]any{}}, schema(), true, nil, time.Date(2026, 7, 20, 1, 0, 0, 0, time.FixedZone("PDT", -7*3600)))
	if !outcome.Changed {
		t.Fatal("expected repair changes")
	}
	if outcome.Values["author"] != "Demon Docs" || outcome.Values["document_type"] != "general" || outcome.Values["created"] != "2026-07-20" {
		t.Fatalf("defaults not repaired: %#v", outcome.Values)
	}
	id, ok := outcome.Values["document_id"].(string)
	if !ok || !uuidPattern.MatchString(id) {
		t.Fatalf("UUID not generated: %#v", outcome.Values["document_id"])
	}
	if _, present := outcome.Values["summary"]; present {
		t.Fatal("unsourced required summary should not be invented")
	}
	if !hasUnresolved(outcome.Diagnostics, "summary") {
		t.Fatalf("missing summary not diagnosed: %+v", outcome.Diagnostics)
	}
}

func TestEvaluateNeverOverwritesExistingValidMutableValues(t *testing.T) {
	values := completeValues()
	values["author"] = "Human"
	values["document_type"] = "guide"
	outcome := Evaluate("docs/guide.md", Document{Values: values}, schema(), true, nil, time.Now())
	if outcome.Changed || outcome.Values["author"] != "Human" || outcome.Values["document_type"] != "guide" {
		t.Fatalf("valid mutable values changed: %+v", outcome)
	}
}

func TestEvaluateReportsInvalidMutableAndRepairsInvalidImmutable(t *testing.T) {
	values := completeValues()
	values["author"] = int64(12)
	values["created"] = "not-a-date"
	outcome := Evaluate("docs/guide.md", Document{Values: values}, schema(), true, nil, time.Date(2026, 7, 19, 23, 0, 0, 0, time.FixedZone("PDT", -7*3600)))
	if outcome.Values["author"] != int64(12) {
		t.Fatalf("invalid mutable author was overwritten: %#v", outcome.Values["author"])
	}
	if outcome.Values["created"] != "2026-07-19" {
		t.Fatalf("invalid immutable date was not repaired in local time: %#v", outcome.Values["created"])
	}
	if !hasUnresolved(outcome.Diagnostics, "author") || hasUnresolved(outcome.Diagnostics, "created") {
		t.Fatalf("unexpected diagnostics: %+v", outcome.Diagnostics)
	}
}

func TestEvaluateRestoresRecordedImmutableValue(t *testing.T) {
	values := completeValues()
	values["document_id"] = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	recorded := map[string]any{"document_id": "11111111-2222-4333-8444-555555555555"}
	outcome := Evaluate("docs/guide.md", Document{Values: values}, schema(), true, recorded, time.Now())
	if outcome.Values["document_id"] != recorded["document_id"] || !outcome.Changed {
		t.Fatalf("immutable baseline not restored: %+v", outcome)
	}
}

func TestUnknownFieldModes(t *testing.T) {
	for _, mode := range []string{"remove", "warn", "ignore"} {
		t.Run(mode, func(t *testing.T) {
			cfg := schema()
			cfg.UnknownFields = mode
			values := completeValues()
			values["unknown"] = "value"
			outcome := Evaluate("docs/guide.md", Document{Values: values}, cfg, true, nil, time.Now())
			switch mode {
			case "remove":
				if _, present := outcome.Values["unknown"]; present || !outcome.Changed || hasUnresolved(outcome.Diagnostics, "unknown") {
					t.Fatalf("remove failed: %+v", outcome)
				}
			case "warn":
				if _, present := outcome.Values["unknown"]; !present || len(outcome.Diagnostics) != 1 || !outcome.Diagnostics[0].Warning {
					t.Fatalf("warn failed: %+v", outcome)
				}
			case "ignore":
				if _, present := outcome.Values["unknown"]; !present || len(outcome.Diagnostics) != 0 {
					t.Fatalf("ignore failed: %+v", outcome)
				}
			}
		})
	}
}

func TestConditionalRuleCanRemainUnresolvedOrUseConfiguredSource(t *testing.T) {
	cfg := schema()
	cfg.Fields["policy_exempt"] = config.FrontmatterField{Type: "boolean", Default: false}
	cfg.Fields["policy_exempt_reason"] = config.FrontmatterField{Type: "string"}
	cfg.Rules = []config.FrontmatterRule{{WhenField: "policy_exempt", Equals: true, Require: "policy_exempt_reason"}}
	values := completeValues()
	values["policy_exempt"] = true
	outcome := Evaluate("docs/guide.md", Document{Values: values}, cfg, true, nil, time.Now())
	if !hasUnresolved(outcome.Diagnostics, "policy_exempt_reason") {
		t.Fatalf("conditional requirement not diagnosed: %+v", outcome.Diagnostics)
	}

	field := cfg.Fields["policy_exempt_reason"]
	field.Default = "approved exception"
	cfg.Fields["policy_exempt_reason"] = field
	outcome = Evaluate("docs/guide.md", Document{Values: values}, cfg, true, nil, time.Now())
	if outcome.Values["policy_exempt_reason"] != "approved exception" || hasUnresolved(outcome.Diagnostics, "policy_exempt_reason") {
		t.Fatalf("conditional source not repaired: %+v", outcome)
	}
}

func TestValidateConfigRejectsInvalidSchemas(t *testing.T) {
	cfg := schema()
	cfg.DefaultFormat = "json"
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("expected invalid format error")
	}
	cfg = schema()
	field := cfg.Fields["document_id"]
	field.Default = "11111111-2222-4333-8444-555555555555"
	cfg.Fields["document_id"] = field
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("expected multiple source error")
	}
	cfg = schema()
	cfg.Fields["bad"] = config.FrontmatterField{Type: "string_list", Default: []any{"okay", int64(2)}}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatal("expected invalid string-list default error")
	}
}
