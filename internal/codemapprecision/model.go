// Package codemapprecision evaluates manually labeled missing-link suggestions.
package codemapprecision

import "github.com/Lokee86/demon-docs/internal/codemapbench"

const SchemaVersion = 1

type Label string

const (
	ValidMissingLink        Label = "valid_missing_link"
	PlausibleButUnnecessary Label = "plausible_but_unnecessary"
	Incorrect               Label = "incorrect"
)

func (l Label) Valid() bool {
	return l == ValidMissingLink || l == PlausibleButUnnecessary || l == Incorrect
}

type CorpusMetadata struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	ReviewedAt string `json:"reviewed_at"`
}

type SamplingMetadata struct {
	Seed                 string `json:"seed"`
	SourceReport         string `json:"source_report"`
	CandidateCount       int    `json:"candidate_count"`
	RequestedSampleCount int    `json:"requested_sample_count"`
	Method               string `json:"method"`
}

type AuditMetadata struct {
	DocumentSection string `json:"document_section"`
	DocumentRef     string `json:"document_ref"`
	DocumentExcerpt string `json:"document_excerpt"`
	TargetRef       string `json:"target_ref"`
	TargetExcerpt   string `json:"target_excerpt"`
	TargetSHA256    string `json:"target_sha256"`
	TargetKind      string `json:"target_kind"`
}

type LabeledSuggestion struct {
	codemapbench.Suggestion
	Rank                int           `json:"rank"`
	Area                string        `json:"area"`
	Subsystem           string        `json:"subsystem"`
	ScoreBucket         string        `json:"score_bucket"`
	RankBucket          string        `json:"rank_bucket"`
	PrimaryEvidenceKind string        `json:"primary_evidence_kind"`
	EvidenceKinds       []string      `json:"evidence_kinds"`
	Label               Label         `json:"label"`
	Rationale           string        `json:"rationale"`
	Audit               AuditMetadata `json:"audit"`
}

type Benchmark struct {
	SchemaVersion int                 `json:"schema_version"`
	Corpus        CorpusMetadata      `json:"corpus"`
	Sampling      SamplingMetadata    `json:"sampling"`
	Suggestions   []LabeledSuggestion `json:"suggestions"`
}

type Counts struct {
	Valid     int `json:"valid_missing_link"`
	Plausible int `json:"plausible_but_unnecessary"`
	Incorrect int `json:"incorrect"`
	Total     int `json:"total"`
}

type PrecisionMetrics struct {
	Total               int     `json:"total"`
	Valid               int     `json:"valid_missing_link"`
	Accepted            int     `json:"accepted_non_junk"`
	OverallPrecision    float64 `json:"overall_precision"`
	AcceptancePrecision float64 `json:"acceptance_precision"`
}

type DocumentMetrics struct {
	PrecisionAt1 PrecisionMetrics `json:"precision_at_1"`
	PrecisionAt3 PrecisionMetrics `json:"precision_at_3"`
	PrecisionAt5 PrecisionMetrics `json:"precision_at_5"`
}

type Evaluation struct {
	SchemaVersion                  int                         `json:"schema_version"`
	BenchmarkSize                  int                         `json:"benchmark_size"`
	LabelCounts                    Counts                      `json:"label_counts"`
	Overall                        PrecisionMetrics            `json:"overall"`
	PrecisionAt1                   float64                     `json:"precision_at_1"`
	PrecisionAt3                   float64                     `json:"precision_at_3"`
	PrecisionAt5                   float64                     `json:"precision_at_5"`
	HardLinkSampleValidRecall      float64                     `json:"hard_link_sample_valid_recall"`
	HardLinkSuggestionsPerDocument float64                     `json:"hard_link_suggestions_per_document"`
	PerDocument                    map[string]DocumentMetrics  `json:"per_document"`
	ByEvidenceKind                 map[string]PrecisionMetrics `json:"by_evidence_kind"`
	ByScoreBucket                  map[string]PrecisionMetrics `json:"by_score_bucket"`
	ByRankBucket                   map[string]PrecisionMetrics `json:"by_rank_bucket"`
	ByTier                         map[string]PrecisionMetrics `json:"by_tier"`
	SamplingCoverage               map[string]map[string]int   `json:"sampling_coverage"`
}
