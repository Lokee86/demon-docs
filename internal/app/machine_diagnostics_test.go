package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type machineDiagnosticV1 struct {
	Code        string   `json:"code"`
	Severity    string   `json:"severity"`
	Subsystem   string   `json:"subsystem"`
	Message     string   `json:"message"`
	Path        string   `json:"path"`
	Field       string   `json:"field"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Target      string   `json:"target"`
	Replacement string   `json:"replacement"`
	Candidates  []string `json:"candidates"`
	Section     string   `json:"section"`
}

type machineReportV1 struct {
	SchemaVersion int                   `json:"schema_version"`
	Command       string                `json:"command"`
	Status        string                `json:"status"`
	ExitCode      int                   `json:"exit_code"`
	Diagnostics   []machineDiagnosticV1 `json:"diagnostics"`
}

func TestCheckLinksJSONDiagnosticContract(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docs, "INDEX.md"), "[Target](target.md)\n")
	target := filepath.Join(docs, "target.md")
	writeTestFile(t, target, "# Target\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--links"}, &out, &errOut); code != 0 {
			t.Fatalf("baseline code=%d out=%q err=%q", code, out.String(), errOut.String())
		}

		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--links", "--output-format", "json"}, &out, &errOut); code != 0 {
			t.Fatalf("clean json code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		var clean machineReportV1
		if err := json.Unmarshal(out.Bytes(), &clean); err != nil {
			t.Fatalf("decode clean report: %v\n%s", err, out.String())
		}
		if clean.SchemaVersion != 1 || clean.Command != "check" || clean.Status != "passed" || clean.ExitCode != 0 || clean.Diagnostics == nil || len(clean.Diagnostics) != 0 {
			t.Fatalf("unexpected clean report: %#v", clean)
		}

		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--links", "--output-format", "json"}, &out, &errOut); code != 1 {
			t.Fatalf("broken json code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if errOut.Len() != 0 {
			t.Fatalf("machine diagnostics wrote stderr: %q", errOut.String())
		}
		var report machineReportV1
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatalf("decode failed report: %v\n%s", err, out.String())
		}
		if report.SchemaVersion != 1 || report.Command != "check" || report.Status != "failed" || report.ExitCode != 1 {
			t.Fatalf("unexpected report envelope: %#v", report)
		}
		if len(report.Diagnostics) != 1 {
			t.Fatalf("diagnostics=%#v", report.Diagnostics)
		}
		diagnostic := report.Diagnostics[0]
		if diagnostic.Code != "links.broken" || diagnostic.Severity != "error" || diagnostic.Subsystem != "links" || diagnostic.Message != "Local link target does not exist" {
			t.Fatalf("unexpected diagnostic identity: %#v", diagnostic)
		}
		if diagnostic.Path != "docs/INDEX.md" || diagnostic.Line != 1 || diagnostic.Column != 10 || diagnostic.Target != "target.md" {
			t.Fatalf("unexpected diagnostic location/evidence: %#v", diagnostic)
		}
	})
}

func TestCheckLinksJSONIncludesOrphanHealthDiagnostic(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docs, "INDEX.md"), "# Docs\n")
	writeTestFile(t, filepath.Join(docs, "orphan.md"), "# Orphan\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--links"}, &out, &errOut); code != 0 {
			t.Fatalf("baseline code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--links", "--output-format", "json"}, &out, &errOut); code != 1 {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		var report machineReportV1
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "links.orphan_document" || report.Diagnostics[0].Severity != "warning" || report.Diagnostics[0].Path != "docs/orphan.md" {
			t.Fatalf("unexpected orphan report: %#v", report)
		}
	})
}

func TestCheckIndexesJSONDiagnosticContract(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docs, "page.md"), "# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--indexes", "--output-format", "json"}, &out, &errOut); code != 1 {
			t.Fatalf("missing index code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		var missing machineReportV1
		if err := json.Unmarshal(out.Bytes(), &missing); err != nil {
			t.Fatal(err)
		}
		if len(missing.Diagnostics) != 1 || missing.Diagnostics[0].Code != "indexes.missing" || missing.Diagnostics[0].Severity != "warning" || missing.Diagnostics[0].Subsystem != "indexes" || missing.Diagnostics[0].Path != "docs/INDEX.md" {
			t.Fatalf("unexpected missing-index report: %#v", missing)
		}

		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--indexes"}, &out, &errOut); code != 0 {
			t.Fatalf("index baseline code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		writeTestFile(t, filepath.Join(docs, "second.md"), "# Second\n")
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--indexes", "--output-format", "json"}, &out, &errOut); code != 1 {
			t.Fatalf("outdated index code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		var outdated machineReportV1
		if err := json.Unmarshal(out.Bytes(), &outdated); err != nil {
			t.Fatal(err)
		}
		if len(outdated.Diagnostics) != 1 || outdated.Diagnostics[0].Code != "indexes.out_of_date" || outdated.Diagnostics[0].Path != "docs/INDEX.md" {
			t.Fatalf("unexpected outdated-index report: %#v", outdated)
		}
	})
}

func TestCheckJSONComposesMigratedSubsystems(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	configText := strings.Replace(frontmatterTestConfig(true, "yaml"), "[links]\nenabled = false", "[links]\nenabled = true", 1)
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), configText)
	writeTestFile(t, filepath.Join(docs, "source.md"), "[Target](target.md)\n")
	target := filepath.Join(docs, "target.md")
	writeTestFile(t, target, "# Target\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--indexes"}, &out, &errOut); code != 0 {
			t.Fatalf("index baseline code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"fix", "--links"}, &out, &errOut); code != 0 {
			t.Fatalf("link baseline code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--links", "--indexes", "--frontmatter", "--output-format", "json"}, &out, &errOut); code != 1 {
			t.Fatalf("mixed code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		var report machineReportV1
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, diagnostic := range report.Diagnostics {
			seen[diagnostic.Code] = true
		}
		if !seen["indexes.out_of_date"] || !seen["frontmatter.missing_field"] || !seen["links.broken"] {
			t.Fatalf("mixed report missing migrated subsystem diagnostics: %#v", report.Diagnostics)
		}
		if len(report.Diagnostics) < 3 || report.Diagnostics[0].Subsystem != "indexes" {
			t.Fatalf("mixed report did not start with index diagnostics: %#v", report.Diagnostics)
		}
		frontmatterSeen := false
		linkSeen := false
		for _, diagnostic := range report.Diagnostics {
			if diagnostic.Subsystem == "frontmatter" {
				frontmatterSeen = true
			}
			if diagnostic.Subsystem == "links" {
				if !frontmatterSeen {
					t.Fatalf("link diagnostics preceded frontmatter diagnostics: %#v", report.Diagnostics)
				}
				linkSeen = true
			}
		}
		if !frontmatterSeen || !linkSeen {
			t.Fatalf("mixed report missing ordered subsystems: %#v", report.Diagnostics)
		}
	})
}

func TestCheckFrontmatterJSONDiagnosticContract(t *testing.T) {
	repo := t.TempDir()
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), frontmatterTestConfig(false, "yaml"))
	writeTestFile(t, filepath.Join(repo, "docs", "page.md"), "---\nauthor: 12\ncreated: \"2026-07-20\"\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\n---\n# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"check", "--frontmatter", "--output-format", "json"}, &out, &errOut); code != 1 {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		if errOut.Len() != 0 {
			t.Fatalf("machine diagnostics wrote stderr: %q", errOut.String())
		}
		var report machineReportV1
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if report.SchemaVersion != 1 || report.Status != "failed" || report.ExitCode != 1 || len(report.Diagnostics) != 1 {
			t.Fatalf("unexpected report: %#v", report)
		}
		diagnostic := report.Diagnostics[0]
		if diagnostic.Code != "frontmatter.invalid_value" || diagnostic.Severity != "error" || diagnostic.Subsystem != "frontmatter" || diagnostic.Path != "docs/page.md" || diagnostic.Field != "author" {
			t.Fatalf("unexpected frontmatter diagnostic: %#v", diagnostic)
		}
	})
}

func TestCheckFrontmatterJSONWarningDoesNotFail(t *testing.T) {
	repo := t.TempDir()
	configText := strings.Replace(frontmatterTestConfig(false, "yaml"), `unknown_fields = "remove"`, `unknown_fields = "warn"`, 1)
	writeTestFile(t, filepath.Join(repo, ".ddocs", "config.toml"), configText)
	writeTestFile(t, filepath.Join(repo, "docs", "page.md"), "---\nauthor: Human\ncreated: \"2026-07-20\"\ndocument_id: 11111111-2222-4333-8444-555555555555\ndocument_type: guide\nsummary: Existing\nunknown: kept\n---\n# Page\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"check", "--frontmatter", "--output-format", "json"}, &out, &errOut); code != 0 {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		var report machineReportV1
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if report.Status != "passed" || report.ExitCode != 0 || len(report.Diagnostics) != 1 {
			t.Fatalf("unexpected warning report: %#v", report)
		}
		diagnostic := report.Diagnostics[0]
		if diagnostic.Code != "frontmatter.unknown_field" || diagnostic.Severity != "warning" || diagnostic.Field != "unknown" {
			t.Fatalf("unexpected warning diagnostic: %#v", diagnostic)
		}
	})
}

func TestJSONDiagnosticsRejectUnmigratedReconciliationSelection(t *testing.T) {
	repo := t.TempDir()
	docs := filepath.Join(repo, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docs, "INDEX.md"), "# Docs\n")

	withWorkingDirectory(t, repo, func(string) {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), []string{"init", "--root", "docs"}, &out, &errOut); code != 0 {
			t.Fatalf("init code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
		out.Reset()
		errOut.Reset()
		if code := Run(context.Background(), []string{"check", "--links", "--format", "--output-format", "json"}, &out, &errOut); code != 2 {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}
