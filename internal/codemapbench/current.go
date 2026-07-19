package codemapbench

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

// SuggestCurrent generates candidate missing links while treating every
// authored codemap relationship as visible. Unlike Run, it performs no
// holdout and is intended for precision sampling of genuinely new links.
func (r Runner) SuggestCurrent(ctx context.Context) (Report, error) {
	if r.Corpus == nil {
		return Report{}, errors.New("benchmark corpus is required")
	}
	links, err := r.Corpus.Links(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("load authored links: %w", err)
	}
	known, err := normalizeKnownLinks(links)
	if err != nil {
		return Report{}, err
	}
	if len(known) == 0 {
		return Report{}, errors.New("current suggestion run requires at least one authored link")
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

	raw, err := r.generate(ctx, Request{
		Documents:    documentPaths(known),
		VisibleLinks: append([]Link(nil), known...),
	}, collector, builder)
	if err != nil {
		return Report{}, fmt.Errorf("generate current suggestions: %w", err)
	}

	report := Report{
		Seed:               "current-authored-links",
		KnownLinks:         known,
		VisibleLinks:       append([]Link(nil), known...),
		RawSuggestionCount: len(raw),
	}
	classifySuggestions(&report, raw)
	calculateScores(&report)
	return report, nil
}
