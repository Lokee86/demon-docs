package codemapbench

import (
	"context"

	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
)

// DefaultSeed keeps the default holdout stable across runs.
const DefaultSeed = "demon-docs-codemap-benchmark-v1"

// Link is one authored relationship from a document to a code target.
type Link = codemaprecommend.Link
type SuggestionTier = codemaprecommend.SuggestionTier

const (
	SuggestionTierHardLink = codemaprecommend.SuggestionTierHardLink
	SuggestionTierContext  = codemaprecommend.SuggestionTierContext
)

type Suggestion = codemaprecommend.Suggestion

// Request is the information exposed to a suggestion generator. Hidden links
// are deliberately absent so the benchmark cannot leak its answers.
type Request struct {
	Documents    []string `json:"documents"`
	VisibleLinks []Link   `json:"visible_links"`
}

// Generator proposes missing links from repository evidence.
type Generator interface {
	Generate(context.Context, Request) ([]Suggestion, error)
}

// GeneratorFunc adapts a function to Generator.
type GeneratorFunc func(context.Context, Request) ([]Suggestion, error)

func (f GeneratorFunc) Generate(ctx context.Context, request Request) ([]Suggestion, error) {
	return f(ctx, request)
}

// Config controls deterministic holdout selection.
type Config struct {
	Seed            string  `json:"seed"`
	HoldoutCount    int     `json:"holdout_count,omitempty"`
	HoldoutFraction float64 `json:"holdout_fraction,omitempty"`
}

// InvalidSuggestion records output that cannot identify a document and target.
type InvalidSuggestion struct {
	Index      int        `json:"index"`
	Suggestion Suggestion `json:"suggestion"`
	Reason     string     `json:"reason"`
}

// Report contains deterministic benchmark inputs, classifications, and scores.
type Report struct {
	Seed                  string              `json:"seed"`
	KnownLinks            []Link              `json:"known_links"`
	VisibleLinks          []Link              `json:"visible_links"`
	HiddenLinks           []Link              `json:"hidden_links"`
	RecoveredLinks        []Link              `json:"recovered_links"`
	RecoveredSuggestions  []Suggestion        `json:"recovered_suggestions"`
	MissedLinks           []Link              `json:"missed_links"`
	UnmatchedSuggestions  []Suggestion        `json:"unmatched_suggestions"`
	AlreadyLinked         []Suggestion        `json:"already_linked_suggestions"`
	DuplicateSuggestions  []Suggestion        `json:"duplicate_suggestions"`
	InvalidSuggestions    []InvalidSuggestion `json:"invalid_suggestions"`
	RawSuggestionCount    int                 `json:"raw_suggestion_count"`
	UniqueSuggestionCount int                 `json:"unique_suggestion_count"`
	Precision             float64             `json:"precision"`
	Recall                float64             `json:"recall"`
}
