package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapbench"
	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
)

func runCodemapPrecisionSource(ctx context.Context, args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		fmt.Fprintln(out, "usage: ddocs codemap precision source [-h] [--repo PATH] [--dataset PATH] [--exclude-prefix PATH] [--output PATH]\n\nGenerate current missing-link suggestions while treating every authored codemap link as visible. The JSON report is suitable input for `ddocs codemap precision sample`.\n\noptions:\n  --repo PATH            repository to analyze\n  --dataset PATH         use a prebuilt codemap dataset\n  --exclude-prefix PATH  exclude documents under this repository-relative prefix; repeatable\n  --output PATH          source report output file")
		return 0
	}
	fs := flag.NewFlagSet("ddocs codemap precision source", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var repo, dataset, output optionalString
	var excludes stringsFlag
	fs.Var(&repo, "repo", "repository to analyze")
	fs.Var(&dataset, "dataset", "prebuilt codemap dataset")
	fs.Var(&excludes, "exclude-prefix", "repository-relative document prefix to exclude")
	fs.Var(&output, "output", "source report output file")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(errOut, "ddocs codemap precision source: error: %v\n", err)
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(errOut, "ddocs codemap precision source: error: unrecognized arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
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
	datasetValue, format, err := loadBenchmarkDataset(codemapBenchmarkOptions{
		RepositoryRoot: repositoryRoot,
		DatasetPath:    datasetPath,
	})
	if err != nil {
		return fail(errOut, err)
	}
	datasetValue = filterCodemapDataset(datasetValue, excludes.values)
	corpus, err := codemapcorpus.Build(repositoryRoot, datasetValue, codemapcorpus.Options{})
	if err != nil {
		return fail(errOut, fmt.Errorf("build precision corpus: %w", err))
	}
	links := codemapbench.ResolvedLinksFromDataset(datasetValue)
	runner := codemapbench.NewRunner(benchmarkCorpus{links: links, corpus: corpus, format: format}, codemapbench.Config{})
	report, err := runner.SuggestCurrent(ctx)
	if err != nil {
		return fail(errOut, err)
	}
	payload, err := codemapbench.MarshalJSONReport(report)
	if err != nil {
		return fail(errOut, err)
	}
	if err := writePrecisionPayload(out, output, payload); err != nil {
		return fail(errOut, err)
	}
	return 0
}

func filterCodemapDataset(dataset codemap.Dataset, prefixes []string) codemap.Dataset {
	if len(prefixes) == 0 {
		return dataset
	}
	normalized := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		prefix = strings.Trim(strings.ReplaceAll(strings.TrimSpace(prefix), `\`, "/"), "/")
		if prefix != "" {
			normalized = append(normalized, prefix)
		}
	}
	if len(normalized) == 0 {
		return dataset
	}
	excluded := func(path string) bool {
		path = strings.TrimPrefix(strings.ReplaceAll(path, `\`, "/"), "./")
		for _, prefix := range normalized {
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}
		}
		return false
	}
	filtered := codemap.Dataset{SchemaVersion: dataset.SchemaVersion}
	for _, document := range dataset.Documents {
		if !excluded(document.Path) {
			filtered.Documents = append(filtered.Documents, document)
		}
	}
	for _, entry := range dataset.Entries {
		if !excluded(entry.Entry.DocumentPath) {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}
	for _, diagnostic := range dataset.Diagnostics {
		if !excluded(diagnostic.DocumentPath) {
			filtered.Diagnostics = append(filtered.Diagnostics, diagnostic)
		}
	}
	return filtered
}
