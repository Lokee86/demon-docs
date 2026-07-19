package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapbench"
	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/evidence"
)

type benchmarkEngine struct{}

func (benchmarkEngine) Run(ctx context.Context, options codemapBenchmarkOptions) (codemapBenchmarkResult, error) {
	dataset, format, err := loadBenchmarkDataset(options)
	if err != nil {
		return codemapBenchmarkResult{}, err
	}
	corpus, err := codemapcorpus.Build(options.RepositoryRoot, dataset, codemapcorpus.Options{})
	if err != nil {
		return codemapBenchmarkResult{}, fmt.Errorf("build benchmark corpus: %w", err)
	}
	links, err := loadBenchmarkLinks(options.TrustedPath, dataset)
	if err != nil {
		return codemapBenchmarkResult{}, err
	}

	runner := codemapbench.NewRunner(benchmarkCorpus{
		links:  links,
		corpus: corpus,
		format: format,
	}, codemapbench.Config{
		Seed:            options.Seed,
		HoldoutCount:    options.HoldoutCount,
		HoldoutFraction: options.HoldoutFraction,
	})
	report, err := runner.Run(ctx)
	if err != nil {
		return codemapBenchmarkResult{}, err
	}
	payload, err := encodeBenchmarkReport(report, options.Format)
	if err != nil {
		return codemapBenchmarkResult{}, err
	}
	return codemapBenchmarkResult{Payload: payload, Precision: report.Precision, Recall: report.Recall}, nil
}

type benchmarkCorpus struct {
	links  []codemapbench.Link
	corpus codemapcorpus.Corpus
	format codemap.Format
}

func (c benchmarkCorpus) Links(ctx context.Context) ([]codemapbench.Link, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]codemapbench.Link(nil), c.links...), nil
}

func (c benchmarkCorpus) DocumentInput(ctx context.Context, request codemapbench.DocumentRequest) (evidence.Input, error) {
	if err := ctx.Err(); err != nil {
		return evidence.Input{}, err
	}
	input, err := c.corpus.Input(request.Document, request.VisibleTargets)
	if err != nil {
		return evidence.Input{}, err
	}
	input.DocumentText = codemap.StripAuthoredSections(input.DocumentText, c.format)
	return input, nil
}

func loadBenchmarkDataset(options codemapBenchmarkOptions) (codemap.Dataset, codemap.Format, error) {
	if options.DatasetPath == "" {
		format := codemap.DefaultFormat()
		dataset, err := codemap.BuildDataset(options.RepositoryRoot, options.RepositoryRoot, format)
		return dataset, format, err
	}
	file, err := os.Open(options.DatasetPath)
	if err != nil {
		return codemap.Dataset{}, codemap.Format{}, fmt.Errorf("open codemap dataset: %w", err)
	}
	defer file.Close()
	var dataset codemap.Dataset
	if err := json.NewDecoder(file).Decode(&dataset); err != nil {
		return codemap.Dataset{}, codemap.Format{}, fmt.Errorf("decode codemap dataset: %w", err)
	}
	if dataset.SchemaVersion != codemap.DatasetSchemaVersion {
		return codemap.Dataset{}, codemap.Format{}, fmt.Errorf("unsupported codemap dataset schema %d", dataset.SchemaVersion)
	}
	return dataset, datasetFormat(dataset), nil
}

func datasetFormat(dataset codemap.Dataset) codemap.Format {
	format := codemap.DefaultFormat()
	set := map[string]struct{}{}
	for _, item := range dataset.Entries {
		if item.Entry.Heading != "" {
			set[item.Entry.Heading] = struct{}{}
		}
	}
	if len(set) == 0 {
		return format
	}
	format.SectionHeadings = format.SectionHeadings[:0]
	for heading := range set {
		format.SectionHeadings = append(format.SectionHeadings, heading)
	}
	sort.Strings(format.SectionHeadings)
	return format
}

func loadBenchmarkLinks(trustedPath string, dataset codemap.Dataset) ([]codemapbench.Link, error) {
	if trustedPath == "" {
		return codemapbench.ResolvedLinksFromDataset(dataset), nil
	}
	file, err := os.Open(trustedPath)
	if err != nil {
		return nil, fmt.Errorf("open trusted links: %w", err)
	}
	defer file.Close()
	links, err := codemapbench.DecodeTrustedReviewLinks(file)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func encodeBenchmarkReport(report codemapbench.Report, format string) ([]byte, error) {
	if format == "json" {
		return codemapbench.MarshalJSONReport(report)
	}
	return []byte(codemapbench.FormatTextReport(report)), nil
}
