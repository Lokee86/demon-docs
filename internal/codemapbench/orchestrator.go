package codemapbench

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

// DocumentRequest asks a corpus for the non-codemap evidence needed to analyze
// one document. Hidden links are never included.
type DocumentRequest struct {
	Document       string   `json:"document"`
	VisibleLinks   []Link   `json:"visible_links"`
	VisibleTargets []string `json:"visible_targets"`
}

// Corpus supplies benchmark truth and repository evidence. DocumentInput must
// omit authored codemap content; the runner owns existing-target relationships.
type Corpus interface {
	Links(context.Context) ([]Link, error)
	DocumentInput(context.Context, DocumentRequest) (evidence.Input, error)
}

// CandidateCollector extracts deterministic evidence candidates.
type CandidateCollector interface {
	Collect(context.Context, evidence.Input) ([]evidence.Candidate, error)
}

// CandidateCollectorFunc adapts a function to CandidateCollector.
type CandidateCollectorFunc func(context.Context, evidence.Input) ([]evidence.Candidate, error)

func (f CandidateCollectorFunc) Collect(ctx context.Context, input evidence.Input) ([]evidence.Candidate, error) {
	return f(ctx, input)
}

// SuggestionBuilder converts evidence candidates into ranked suggestions.
type SuggestionBuilder interface {
	Build(context.Context, string, []evidence.Candidate) ([]Suggestion, error)
}

// SuggestionBuilderFunc adapts a function to SuggestionBuilder.
type SuggestionBuilderFunc func(context.Context, string, []evidence.Candidate) ([]Suggestion, error)

func (f SuggestionBuilderFunc) Build(ctx context.Context, document string, candidates []evidence.Candidate) ([]Suggestion, error) {
	return f(ctx, document, candidates)
}

// Runner orchestrates corpus loading, deterministic holdout, evidence
// collection, suggestion generation, and benchmark scoring.
type Runner struct {
	Corpus    Corpus
	Collector CandidateCollector
	Builder   SuggestionBuilder
	Config    Config
}

// NewRunner returns a runner using the built-in deterministic evidence
// collector and evidence-to-suggestion adapter.
func NewRunner(corpus Corpus, config Config) Runner {
	return Runner{Corpus: corpus, Config: config}
}

// Run executes the complete benchmark without exposing hidden links to the
// corpus evidence requests or downstream collectors.
func (r Runner) Run(ctx context.Context) (Report, error) {
	if r.Corpus == nil {
		return Report{}, errors.New("benchmark corpus is required")
	}
	links, err := r.Corpus.Links(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("load benchmark links: %w", err)
	}

	collector := r.Collector
	if collector == nil {
		collector = CandidateCollectorFunc(func(ctx context.Context, input evidence.Input) ([]evidence.Candidate, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return evidence.Collect(input), nil
		})
	}
	builder := r.Builder
	if builder == nil {
		builder = SuggestionBuilderFunc(func(ctx context.Context, document string, candidates []evidence.Candidate) ([]Suggestion, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return SuggestionsFromEvidence(document, candidates), nil
		})
	}

	generator := GeneratorFunc(func(ctx context.Context, request Request) ([]Suggestion, error) {
		return r.generate(ctx, request, collector, builder)
	})
	return Run(ctx, links, generator, r.Config)
}

func (r Runner) generate(
	ctx context.Context,
	request Request,
	collector CandidateCollector,
	builder SuggestionBuilder,
) ([]Suggestion, error) {
	visibleByDocument := linksByDocument(request.VisibleLinks)
	all := make([]Suggestion, 0)

	for _, document := range request.Documents {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		documentRequest := DocumentRequest{
			Document:       document,
			VisibleLinks:   append([]Link(nil), request.VisibleLinks...),
			VisibleTargets: append([]string(nil), visibleByDocument[document]...),
		}
		input, err := r.Corpus.DocumentInput(ctx, documentRequest)
		if err != nil {
			return nil, fmt.Errorf("build evidence input for %s: %w", document, err)
		}
		input = sanitizeInput(input, document, visibleByDocument)

		candidates, err := collector.Collect(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("collect evidence for %s: %w", document, err)
		}
		suggestions, err := builder.Build(ctx, document, candidates)
		if err != nil {
			return nil, fmt.Errorf("build suggestions for %s: %w", document, err)
		}
		all = append(all, suggestions...)
	}
	return all, nil
}

func linksByDocument(links []Link) map[string][]string {
	result := make(map[string][]string)
	for _, link := range links {
		result[link.Document] = append(result[link.Document], link.Target)
	}
	for document := range result {
		sort.Strings(result[document])
	}
	return result
}

func sanitizeInput(input evidence.Input, document string, visible map[string][]string) evidence.Input {
	input.DocumentPath = document
	input.ExistingTargets = append([]string(nil), visible[document]...)

	relatedByPath := make(map[string]evidence.RelatedDocument, len(input.RelatedDocuments))
	for _, item := range input.RelatedDocuments {
		targets, ok := visible[item.Path]
		if !ok || item.Path == document {
			continue
		}
		relatedByPath[item.Path] = evidence.RelatedDocument{
			Path:    item.Path,
			Targets: append([]string(nil), targets...),
		}
	}
	related := make([]evidence.RelatedDocument, 0, len(relatedByPath))
	for _, item := range relatedByPath {
		related = append(related, item)
	}
	sort.Slice(related, func(i, j int) bool { return related[i].Path < related[j].Path })
	input.RelatedDocuments = related
	return input
}
