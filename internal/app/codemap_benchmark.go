package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemapbench"
)

type codemapBenchmarkOptions struct {
	RepositoryRoot  string
	DatasetPath     string
	TrustedPath     string
	Seed            string
	HoldoutCount    int
	HoldoutFraction float64
	Format          string
}

type codemapBenchmarkResult struct {
	Payload   []byte
	Precision float64
	Recall    float64
}

type codemapBenchmarkRunner interface {
	Run(context.Context, codemapBenchmarkOptions) (codemapBenchmarkResult, error)
}

var codemapBenchmarkCommandRunner codemapBenchmarkRunner = benchmarkEngine{}

type optionalFloat struct {
	set   bool
	value float64
}

func (f *optionalFloat) String() string { return fmt.Sprint(f.value) }
func (f *optionalFloat) Set(value string) error {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float value %q", value)
	}
	f.set = true
	f.value = parsed
	return nil
}

func codemapBenchmarkHelp(w io.Writer) {
	fmt.Fprintln(w, "usage: ddocs codemap benchmark [-h] [--repo PATH] [--dataset PATH]\n                                [--trusted-links PATH] [--seed TEXT]\n                                [--holdout-count N | --holdout-fraction FLOAT]\n                                [--format {text,json}] [--output PATH]\n                                [--min-precision FLOAT] [--min-recall FLOAT]\n\nRun the deterministic missing-link benchmark. By default, the current directory is the target repository and 20% of known links are held out.\n\noptions:\n  -h, --help            show this help message and exit\n  --repo PATH           repository to benchmark (default current directory)\n  --dataset PATH        use a prebuilt codemap dataset\n  --trusted-links PATH  restrict ground truth to a reviewed link set\n  --seed TEXT           deterministic holdout seed\n  --holdout-count N     hide exactly N known links\n  --holdout-fraction FLOAT\n                        hide this fraction of known links (default 0.2)\n  --format {text,json}  report format (default text)\n  --output PATH         write the report to a file instead of stdout\n  --min-precision FLOAT return exit code 1 below this precision\n  --min-recall FLOAT    return exit code 1 below this recall\n\nExit codes:\n  0  benchmark completed and thresholds passed\n  1  benchmark completed but a requested threshold failed\n  2  invalid arguments or benchmark execution failed")
}

func runCodemapBenchmark(ctx context.Context, args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		codemapBenchmarkHelp(out)
		return 0
	}
	fs := flag.NewFlagSet("ddocs codemap benchmark", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var repo, dataset, trusted, output optionalString
	var fraction optionalFloat
	count, format := 0, "text"
	seed := codemapbench.DefaultSeed
	minPrecision, minRecall := -1.0, -1.0
	fs.Var(&repo, "repo", "repository to benchmark")
	fs.Var(&dataset, "dataset", "prebuilt codemap dataset")
	fs.Var(&trusted, "trusted-links", "reviewed ground-truth links")
	fs.StringVar(&seed, "seed", seed, "deterministic holdout seed")
	fs.IntVar(&count, "holdout-count", 0, "exact holdout size")
	fs.Var(&fraction, "holdout-fraction", "holdout fraction")
	fs.StringVar(&format, "format", format, "text or json")
	fs.Var(&output, "output", "report output file")
	fs.Float64Var(&minPrecision, "min-precision", -1, "minimum precision")
	fs.Float64Var(&minRecall, "min-recall", -1, "minimum recall")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs codemap benchmark: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(errOut, "ddocs codemap benchmark: error: unrecognized arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}
	if code := validateBenchmarkFlags(count, fraction, format, minPrecision, minRecall, errOut); code != 0 {
		return code
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(errOut, err)
	}
	repositoryRoot := cwd
	if repo.set {
		repositoryRoot = repo.value
	}
	repositoryRoot, err = filepath.Abs(repositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	if err := requireDirectory(repositoryRoot); err != nil {
		return fail(errOut, err)
	}
	datasetPath, err := optionalInputPath(cwd, dataset)
	if err != nil {
		return fail(errOut, err)
	}
	trustedPath, err := optionalInputPath(cwd, trusted)
	if err != nil {
		return fail(errOut, err)
	}
	result, err := codemapBenchmarkCommandRunner.Run(ctx, codemapBenchmarkOptions{
		RepositoryRoot: repositoryRoot, DatasetPath: datasetPath, TrustedPath: trustedPath,
		Seed: seed, HoldoutCount: count, HoldoutFraction: fraction.value, Format: format,
	})
	if err != nil {
		return fail(errOut, err)
	}
	if err := writeBenchmarkPayload(out, output, result.Payload); err != nil {
		return fail(errOut, err)
	}
	if minPrecision >= 0 && result.Precision < minPrecision || minRecall >= 0 && result.Recall < minRecall {
		fmt.Fprintf(errOut, "ddocs codemap benchmark: thresholds failed (precision %.4f, recall %.4f)\n", result.Precision, result.Recall)
		return 1
	}
	return 0
}

func validateBenchmarkFlags(count int, fraction optionalFloat, format string, minPrecision, minRecall float64, errOut io.Writer) int {
	if count < 0 {
		fmt.Fprintln(errOut, "ddocs codemap benchmark: error: --holdout-count cannot be negative")
		return 2
	}
	if count > 0 && fraction.set {
		fmt.Fprintln(errOut, "ddocs codemap benchmark: error: set --holdout-count or --holdout-fraction, not both")
		return 2
	}
	if fraction.set && (fraction.value <= 0 || fraction.value > 1) {
		fmt.Fprintln(errOut, "ddocs codemap benchmark: error: --holdout-fraction must be greater than zero and at most one")
		return 2
	}
	if format != "text" && format != "json" {
		fmt.Fprintf(errOut, "ddocs codemap benchmark: error: invalid --format %q; expected text or json\n", format)
		return 2
	}
	if !validBenchmarkThreshold(minPrecision) || !validBenchmarkThreshold(minRecall) {
		fmt.Fprintln(errOut, "ddocs codemap benchmark: error: precision and recall thresholds must be between zero and one")
		return 2
	}
	return 0
}

func validBenchmarkThreshold(value float64) bool {
	return value == -1 || value >= 0 && value <= 1
}

func requireDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("repository root: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("repository root is not a directory: %s", path)
	}
	return nil
}

func optionalInputPath(cwd string, value optionalString) (string, error) {
	if !value.set || value.value == "" {
		return "", nil
	}
	path := value.value
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("expected file, found directory: %s", path)
	}
	return path, nil
}

func writeBenchmarkPayload(out io.Writer, output optionalString, payload []byte) error {
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
	fmt.Fprintf(out, "wrote codemap benchmark report to %s\n", path)
	return nil
}
