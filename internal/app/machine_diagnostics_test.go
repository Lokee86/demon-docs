package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type machineDiagnosticV1 struct {
	Code        string   `json:"code"`
	Severity    string   `json:"severity"`
	Subsystem   string   `json:"subsystem"`
	Message     string   `json:"message"`
	Path        string   `json:"path"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Target      string   `json:"target"`
	Replacement string   `json:"replacement"`
	Candidates  []string `json:"candidates"`
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

func TestJSONDiagnosticsRejectMixedReconciliationSelection(t *testing.T) {
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
		if code := Run(context.Background(), []string{"check", "--links", "--indexes", "--output-format", "json"}, &out, &errOut); code != 2 {
			t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
		}
	})
}
