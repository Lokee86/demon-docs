package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordingBenchmarkRunner struct {
	options codemapBenchmarkOptions
	result  codemapBenchmarkResult
	err     error
}

func (r *recordingBenchmarkRunner) Run(_ context.Context, options codemapBenchmarkOptions) (codemapBenchmarkResult, error) {
	r.options = options
	return r.result, r.err
}

func useBenchmarkRunner(t *testing.T, runner codemapBenchmarkRunner) {
	t.Helper()
	previous := codemapBenchmarkCommandRunner
	codemapBenchmarkCommandRunner = runner
	t.Cleanup(func() { codemapBenchmarkCommandRunner = previous })
}

func TestCodemapBenchmarkHelpDocumentsContract(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"codemap", "benchmark", "--help"}, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	for _, expected := range []string{"--repo PATH", "--trusted-links PATH", "--holdout-count N", "--min-recall FLOAT", "Exit codes:"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("help missing %q:\n%s", expected, stdout.String())
		}
	}
}

func TestCodemapBenchmarkPassesNormalizedOptionsToRunner(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		dataset := filepath.Join(root, "dataset.json")
		trusted := filepath.Join(root, "trusted.json")
		writeTestFile(t, dataset, "{}")
		writeTestFile(t, trusted, "[]")
		runner := &recordingBenchmarkRunner{result: codemapBenchmarkResult{Payload: []byte("report\n"), Precision: 0.8, Recall: 0.6}}
		useBenchmarkRunner(t, runner)

		var stdout, stderr bytes.Buffer
		args := []string{"codemap", "benchmark", "--dataset", "dataset.json", "--trusted-links", "trusted.json", "--seed", "fixed", "--holdout-fraction", "0.25", "--format", "json"}
		if code := Run(context.Background(), args, &stdout, &stderr); code != 0 {
			t.Fatalf("code=%d stderr=%q", code, stderr.String())
		}
		if stdout.String() != "report\n" {
			t.Fatalf("stdout=%q", stdout.String())
		}
		if runner.options.RepositoryRoot != root || runner.options.DatasetPath != dataset || runner.options.TrustedPath != trusted {
			t.Fatalf("unexpected paths: %#v", runner.options)
		}
		if runner.options.Seed != "fixed" || runner.options.HoldoutFraction != 0.25 || runner.options.Format != "json" {
			t.Fatalf("unexpected options: %#v", runner.options)
		}
	})
}

func TestCodemapBenchmarkWritesReportAndReturnsOneWhenThresholdFails(t *testing.T) {
	withWorkingDirectory(t, t.TempDir(), func(root string) {
		runner := &recordingBenchmarkRunner{result: codemapBenchmarkResult{Payload: []byte("benchmark"), Precision: 0.7, Recall: 0.4}}
		useBenchmarkRunner(t, runner)
		output := filepath.Join(root, "reports", "benchmark.txt")

		var stdout, stderr bytes.Buffer
		args := []string{"codemap", "benchmark", "--output", output, "--min-precision", "0.6", "--min-recall", "0.5"}
		if code := Run(context.Background(), args, &stdout, &stderr); code != 1 {
			t.Fatalf("code=%d stderr=%q", code, stderr.String())
		}
		contents, err := os.ReadFile(output)
		if err != nil || string(contents) != "benchmark" {
			t.Fatalf("contents=%q err=%v", contents, err)
		}
		if !strings.Contains(stderr.String(), "thresholds failed") || !strings.Contains(stdout.String(), "wrote codemap benchmark report") {
			t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})
}

func TestCodemapBenchmarkRejectsConflictingHoldoutFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"codemap", "benchmark", "--holdout-count", "5", "--holdout-fraction", "0.2"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "not both") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

func TestCodemapBenchmarkRejectsNegativeThreshold(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"codemap", "benchmark", "--min-recall", "-0.5"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "between zero and one") {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
}

func TestCodemapHelpListsBenchmark(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"codemap", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "{export,benchmark}") || !strings.Contains(stdout.String(), "missing-link benchmark") {
		t.Fatalf("unexpected help:\n%s", stdout.String())
	}
}
