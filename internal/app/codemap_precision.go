package app

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemapprecision"
)

const precisionSampleSeed = "demon-docs-codemap-precision-sample-v1"

func codemapPrecisionHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs codemap precision [-h] {source,sample,evaluate} ...\n\nGenerate a current suggestion report, create a deterministic unlabeled sample, or evaluate a fully labeled benchmark.\n\nsubcommands:\n  source              generate current suggestions with authored links visible\n  sample              create an unlabeled precision benchmark template\n  evaluate            evaluate a fully labeled precision benchmark\n\nThe legacy flag-only form is equivalent to evaluate.")
}

func runCodemapPrecision(ctx context.Context, args []string, out, errOut io.Writer) int {
	if err := ctx.Err(); err != nil {
		return fail(errOut, err)
	}
	if helpRequested(args) {
		codemapPrecisionHelp(out)
		return 0
	}
	if len(args) > 0 && args[0] == "source" {
		return runCodemapPrecisionSource(ctx, args[1:], out, errOut)
	}
	if len(args) > 0 && args[0] == "sample" {
		return runCodemapPrecisionSample(args[1:], out, errOut)
	}
	if len(args) > 0 && args[0] == "evaluate" {
		args = args[1:]
	}
	return runCodemapPrecisionEvaluate(args, out, errOut)
}

func runCodemapPrecisionSample(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: ddocs codemap precision sample [-h] --suggestions PATH [--count N] [--seed TEXT] [--repository TEXT] [--revision TEXT] [--output PATH]\n\nCreate a deterministic unlabeled precision benchmark template.")
		return 0
	}
	fs := flag.NewFlagSet("ddocs codemap precision sample", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var suggestionsPath, output optionalString
	count := 150
	seed := precisionSampleSeed
	repository, revision := "", ""
	fs.Var(&suggestionsPath, "suggestions", "source codemap benchmark report JSON")
	fs.IntVar(&count, "count", count, "number of suggestions to sample")
	fs.StringVar(&seed, "seed", seed, "deterministic sampling seed")
	fs.StringVar(&repository, "repository", repository, "corpus repository name")
	fs.StringVar(&revision, "revision", revision, "corpus revision")
	fs.Var(&output, "output", "sample output file")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs codemap precision sample: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 0 || !suggestionsPath.set {
		fmt.Fprintln(errOut, "ddocs codemap precision sample: error: --suggestions is required")
		return 2
	}
	if count <= 0 {
		fmt.Fprintln(errOut, "ddocs codemap precision sample: error: --count must be positive")
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	suggestionsFile, err := optionalInputPath(cwd, suggestionsPath)
	if err != nil {
		return fail(errOut, err)
	}
	file, err := os.Open(suggestionsFile)
	if err != nil {
		return fail(errOut, err)
	}
	report, loadErr := codemapprecision.LoadSuggestionReport(file)
	closeErr := file.Close()
	if loadErr != nil {
		return fail(errOut, loadErr)
	}
	if closeErr != nil {
		return fail(errOut, closeErr)
	}
	sourceReport := suggestionsFile
	if relative, relErr := filepath.Rel(cwd, suggestionsFile); relErr == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		sourceReport = filepath.ToSlash(relative)
	}
	benchmark, err := codemapprecision.BuildBenchmark(report, codemapprecision.SampleConfig{
		Seed: seed, RequestedCount: count, SourceReport: sourceReport,
		Repository: repository, Revision: revision,
	})
	if err != nil {
		return fail(errOut, err)
	}
	var payload bytes.Buffer
	if err := codemapprecision.WriteBenchmark(&payload, benchmark); err != nil {
		return fail(errOut, err)
	}
	if err := writePrecisionPayload(out, output, payload.Bytes()); err != nil {
		return fail(errOut, err)
	}
	return 0
}

func runCodemapPrecisionEvaluate(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: ddocs codemap precision evaluate [-h] --benchmark PATH --suggestions PATH [--format {text,json}] [--output PATH]\n\nEvaluate a fully labeled codemap precision benchmark against a deterministic suggestion report.")
		return 0
	}
	fs := flag.NewFlagSet("ddocs codemap precision evaluate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var benchmarkPath, suggestionsPath, output optionalString
	format := "text"
	fs.Var(&benchmarkPath, "benchmark", "labeled precision benchmark JSON")
	fs.Var(&suggestionsPath, "suggestions", "source codemap benchmark report JSON")
	fs.StringVar(&format, "format", format, "text or json")
	fs.Var(&output, "output", "report output file")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs codemap precision evaluate: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 0 || !benchmarkPath.set || !suggestionsPath.set {
		fmt.Fprintln(errOut, "ddocs codemap precision evaluate: error: --benchmark and --suggestions are required")
		return 2
	}
	if format != "text" && format != "json" {
		fmt.Fprintf(errOut, "ddocs codemap precision evaluate: error: invalid --format %q; expected text or json\n", format)
		return 2
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	benchmarkFile, err := optionalInputPath(cwd, benchmarkPath)
	if err != nil {
		return fail(errOut, err)
	}
	suggestionsFile, err := optionalInputPath(cwd, suggestionsPath)
	if err != nil {
		return fail(errOut, err)
	}
	benchmarkReader, err := os.Open(benchmarkFile)
	if err != nil {
		return fail(errOut, err)
	}
	benchmark, err := codemapprecision.LoadBenchmark(benchmarkReader)
	benchmarkReader.Close()
	if err != nil {
		return fail(errOut, err)
	}
	suggestionsReader, err := os.Open(suggestionsFile)
	if err != nil {
		return fail(errOut, err)
	}
	report, err := codemapprecision.LoadSuggestionReport(suggestionsReader)
	suggestionsReader.Close()
	if err != nil {
		return fail(errOut, err)
	}
	evaluation, err := codemapprecision.Evaluate(benchmark, report)
	if err != nil {
		return fail(errOut, err)
	}
	var payload []byte
	if format == "json" {
		payload, err = json.MarshalIndent(evaluation, "", "  ")
	} else {
		payload = []byte(codemapprecision.FormatEvaluation(evaluation))
	}
	if err != nil {
		return fail(errOut, err)
	}
	if err := writePrecisionPayload(out, output, payload); err != nil {
		return fail(errOut, err)
	}
	return 0
}

func writePrecisionPayload(out io.Writer, output optionalString, payload []byte) error {
	if !output.set {
		_, err := out.Write(payload)
		return err
	}
	path, err := filepath.Abs(output.value)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote codemap precision report to %s\n", path)
	return nil
}
