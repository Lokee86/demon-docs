package codemapprecision

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemapbench"
)

type Candidate struct {
	codemapbench.Suggestion
	Rank                int
	Area                string
	Subsystem           string
	ScoreBucket         string
	RankBucket          string
	EvidenceKinds       []string
	PrimaryEvidenceKind string
}

type SampleConfig struct {
	Seed           string
	RequestedCount int
	SourceReport   string
	Repository     string
	Revision       string
}

func LoadBenchmark(reader io.Reader) (Benchmark, error) {
	var benchmark Benchmark
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&benchmark); err != nil {
		return Benchmark{}, fmt.Errorf("decode precision benchmark: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Benchmark{}, fmt.Errorf("decode precision benchmark: %w", err)
	}
	if benchmark.SchemaVersion != SchemaVersion {
		return Benchmark{}, fmt.Errorf("unsupported precision benchmark schema %d", benchmark.SchemaVersion)
	}
	if err := ValidateBenchmark(benchmark); err != nil {
		return Benchmark{}, err
	}
	return benchmark, nil
}

func WriteBenchmark(writer io.Writer, benchmark Benchmark) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(canonicalBenchmark(benchmark))
}

func LoadSuggestionReport(reader io.Reader) (codemapbench.Report, error) {
	var envelope struct {
		SchemaVersion int `json:"schema_version"`
		codemapbench.Report
	}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&envelope); err != nil {
		return codemapbench.Report{}, fmt.Errorf("decode suggestion report: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return codemapbench.Report{}, fmt.Errorf("decode suggestion report: %w", err)
	}
	if envelope.SchemaVersion != codemapbench.ReportSchemaVersion {
		return codemapbench.Report{}, fmt.Errorf("unsupported suggestion report schema %d", envelope.SchemaVersion)
	}
	return envelope.Report, nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON data")
		}
		return err
	}
	return nil
}

// BuildBenchmark turns a source report into an unlabeled curation template.
// Labels, rationales, and audit fields intentionally remain empty.
func BuildBenchmark(report codemapbench.Report, config SampleConfig) (Benchmark, error) {
	candidates := CandidatesFromReport(report)
	suggestions, err := Sample(candidates, config)
	if err != nil {
		return Benchmark{}, err
	}
	return Benchmark{
		SchemaVersion: SchemaVersion,
		Corpus: CorpusMetadata{
			Repository: config.Repository,
			Revision:   config.Revision,
		},
		Sampling: SamplingMetadata{
			Seed:                 config.Seed,
			SourceReport:         config.SourceReport,
			CandidateCount:       len(candidates),
			RequestedSampleCount: config.RequestedCount,
			Method:               "balanced-documents-top-5-plus-lower-rank-fill-sha256",
		},
		Suggestions: suggestions,
	}, nil
}

// CandidatesFromReport extracts ordinary unmatched suggestions and assigns
// their deterministic rank within each document. Recovered trusted links are
// intentionally excluded from a precision sample.
func CandidatesFromReport(report codemapbench.Report) []Candidate {
	byDocument := map[string][]codemapbench.Suggestion{}
	for _, suggestion := range report.UnmatchedSuggestions {
		byDocument[suggestion.Document] = append(byDocument[suggestion.Document], suggestion)
	}
	result := make([]Candidate, 0, len(report.UnmatchedSuggestions))
	for _, suggestions := range byDocument {
		sort.Slice(suggestions, func(i, j int) bool {
			if suggestions[i].Score != suggestions[j].Score {
				return suggestions[i].Score > suggestions[j].Score
			}
			return suggestions[i].Target < suggestions[j].Target
		})
		for index, suggestion := range suggestions {
			result = append(result, decorateCandidate(suggestion, index+1))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Document != result[j].Document {
			return result[i].Document < result[j].Document
		}
		return result[i].Rank < result[j].Rank
	})
	return result
}

func Sample(candidates []Candidate, config SampleConfig) ([]LabeledSuggestion, error) {
	if config.RequestedCount <= 0 {
		return nil, errors.New("precision sample count must be positive")
	}
	seed := strings.TrimSpace(config.Seed)
	if seed == "" {
		return nil, errors.New("precision sample seed is required")
	}
	unique := uniqueCandidates(candidates)
	if config.RequestedCount > len(unique) {
		return nil, fmt.Errorf("precision sample count %d exceeds %d candidates", config.RequestedCount, len(unique))
	}

	byDocument := make(map[string][]Candidate)
	for _, candidate := range unique {
		byDocument[candidate.Document] = append(byDocument[candidate.Document], candidate)
	}
	for document := range byDocument {
		sort.Slice(byDocument[document], func(i, j int) bool {
			if byDocument[document][i].Rank != byDocument[document][j].Rank {
				return byDocument[document][i].Rank < byDocument[document][j].Rank
			}
			return byDocument[document][i].Target < byDocument[document][j].Target
		})
	}

	documentLimit := config.RequestedCount / 6
	if documentLimit < 1 {
		documentLimit = 1
	}
	if documentLimit > len(byDocument) {
		documentLimit = len(byDocument)
	}
	selectedDocuments := selectBalancedDocuments(byDocument, documentLimit, seed+":documents")
	selected := map[string]Candidate{}
	for _, document := range selectedDocuments {
		for _, candidate := range byDocument[document] {
			if candidate.Rank > 5 || len(selected) == config.RequestedCount {
				break
			}
			selected[candidateKey(candidate)] = candidate
		}
	}

	selectedPool := make([]Candidate, 0)
	for _, document := range selectedDocuments {
		selectedPool = append(selectedPool, byDocument[document]...)
	}
	fillBalanced(selected, selectedPool, config.RequestedCount, seed+":selected-documents")
	if len(selected) < config.RequestedCount {
		fillBalanced(selected, unique, config.RequestedCount, seed+":all-documents")
	}

	result := make([]LabeledSuggestion, 0, len(selected))
	for _, candidate := range selected {
		result = append(result, LabeledSuggestion{
			Suggestion: candidate.Suggestion, Rank: candidate.Rank, Area: candidate.Area,
			Subsystem: candidate.Subsystem, ScoreBucket: candidate.ScoreBucket,
			RankBucket: candidate.RankBucket, PrimaryEvidenceKind: candidate.PrimaryEvidenceKind,
			EvidenceKinds: append([]string(nil), candidate.EvidenceKinds...),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Document != result[j].Document {
			return result[i].Document < result[j].Document
		}
		if result[i].Rank != result[j].Rank {
			return result[i].Rank < result[j].Rank
		}
		return result[i].Target < result[j].Target
	})
	return result, nil
}

func selectBalancedDocuments(byDocument map[string][]Candidate, limit int, seed string) []string {
	preferred := make([]string, 0, len(byDocument))
	fallback := make([]string, 0, len(byDocument))
	for document, values := range byDocument {
		if len(values) >= 5 {
			preferred = append(preferred, document)
		} else {
			fallback = append(fallback, document)
		}
	}
	sort.Strings(preferred)
	sort.Strings(fallback)
	pool := append([]string(nil), preferred...)
	if len(pool) < limit {
		pool = append(pool, fallback...)
	}
	selected := make([]string, 0, limit)
	selectedSet := map[string]struct{}{}
	for len(selected) < limit {
		counts := documentCounts(selected, byDocument)
		best := ""
		var bestKey documentSelectionKey
		for _, document := range pool {
			if _, ok := selectedSet[document]; ok {
				continue
			}
			representative := byDocument[document][0]
			key := documentSelectionKey{
				Area:      counts["area:"+representative.Area],
				Subsystem: counts["subsystem:"+representative.Subsystem],
				Score:     counts["score:"+representative.ScoreBucket],
				Evidence:  counts["evidence:"+representative.PrimaryEvidenceKind],
				Hash:      sampleHash(seed, representative),
			}
			if best == "" || key.less(bestKey) {
				best, bestKey = document, key
			}
		}
		if best == "" {
			break
		}
		selected = append(selected, best)
		selectedSet[best] = struct{}{}
	}
	sort.Strings(selected)
	return selected
}

type documentSelectionKey struct {
	Area      int
	Subsystem int
	Score     int
	Evidence  int
	Hash      string
}

func (key documentSelectionKey) less(other documentSelectionKey) bool {
	if key.Area != other.Area {
		return key.Area < other.Area
	}
	if key.Subsystem != other.Subsystem {
		return key.Subsystem < other.Subsystem
	}
	if key.Score != other.Score {
		return key.Score < other.Score
	}
	if key.Evidence != other.Evidence {
		return key.Evidence < other.Evidence
	}
	return key.Hash < other.Hash
}

func documentCounts(documents []string, byDocument map[string][]Candidate) map[string]int {
	counts := map[string]int{}
	for _, document := range documents {
		representative := byDocument[document][0]
		counts["area:"+representative.Area]++
		counts["subsystem:"+representative.Subsystem]++
		counts["score:"+representative.ScoreBucket]++
		counts["evidence:"+representative.PrimaryEvidenceKind]++
	}
	return counts
}

func fillBalanced(selected map[string]Candidate, pool []Candidate, requested int, seed string) {
	for len(selected) < requested {
		counts := sampleCounts(selected)
		bestIndex := -1
		var bestKey sampleSelectionKey
		for index, candidate := range pool {
			if _, ok := selected[candidateKey(candidate)]; ok {
				continue
			}
			key := sampleSelectionKey{
				Document:  counts["document:"+candidate.Document],
				Composite: counts["stratum:"+stratum(candidate)],
				Area:      counts["area:"+candidate.Area],
				Subsystem: counts["subsystem:"+candidate.Subsystem],
				Score:     counts["score:"+candidate.ScoreBucket],
				Evidence:  counts["evidence:"+candidate.PrimaryEvidenceKind],
				Rank:      counts["rank:"+candidate.RankBucket],
				Hash:      sampleHash(seed, candidate),
			}
			if bestIndex < 0 || key.less(bestKey) {
				bestIndex, bestKey = index, key
			}
		}
		if bestIndex < 0 {
			return
		}
		candidate := pool[bestIndex]
		selected[candidateKey(candidate)] = candidate
	}
}

type sampleSelectionKey struct {
	Document  int
	Composite int
	Area      int
	Subsystem int
	Score     int
	Evidence  int
	Rank      int
	Hash      string
}

func (key sampleSelectionKey) less(other sampleSelectionKey) bool {
	if key.Document != other.Document {
		return key.Document < other.Document
	}
	if key.Composite != other.Composite {
		return key.Composite < other.Composite
	}
	if key.Area != other.Area {
		return key.Area < other.Area
	}
	if key.Subsystem != other.Subsystem {
		return key.Subsystem < other.Subsystem
	}
	if key.Score != other.Score {
		return key.Score < other.Score
	}
	if key.Evidence != other.Evidence {
		return key.Evidence < other.Evidence
	}
	if key.Rank != other.Rank {
		return key.Rank < other.Rank
	}
	return key.Hash < other.Hash
}

func sampleCounts(selected map[string]Candidate) map[string]int {
	counts := make(map[string]int, len(selected)*8)
	for _, candidate := range selected {
		counts["document:"+candidate.Document]++
		counts["stratum:"+stratum(candidate)]++
		counts["area:"+candidate.Area]++
		counts["subsystem:"+candidate.Subsystem]++
		counts["score:"+candidate.ScoreBucket]++
		counts["evidence:"+candidate.PrimaryEvidenceKind]++
		counts["rank:"+candidate.RankBucket]++
	}
	return counts
}

func ValidateBenchmark(benchmark Benchmark) error {
	if benchmark.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported precision benchmark schema %d", benchmark.SchemaVersion)
	}
	seen := map[string]struct{}{}
	for index, item := range benchmark.Suggestions {
		if item.Label != "" && !item.Label.Valid() {
			return fmt.Errorf("suggestion %d has invalid label %q", index, item.Label)
		}
		if item.Document == "" || item.Target == "" || item.Rank < 1 || item.Area == "" || item.ScoreBucket == "" || item.RankBucket == "" || item.PrimaryEvidenceKind == "" {
			return fmt.Errorf("suggestion %d has incomplete identity or rank", index)
		}
		key := item.Document + "\x00" + item.Target
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate suggestion %s -> %s", item.Document, item.Target)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// ValidateLabeledBenchmark applies the curation requirements used by
// Evaluate. Templates are structurally valid before labels are filled, but
// they must not be evaluated until every item is fully audited.
func ValidateLabeledBenchmark(benchmark Benchmark) error {
	if err := ValidateBenchmark(benchmark); err != nil {
		return err
	}
	for _, item := range benchmark.Suggestions {
		if !item.Label.Valid() {
			return fmt.Errorf("suggestion %s -> %s has no valid label", item.Document, item.Target)
		}
		if strings.TrimSpace(item.Rationale) == "" || item.Audit.DocumentRef == "" || item.Audit.DocumentExcerpt == "" || item.Audit.TargetRef == "" || item.Audit.TargetExcerpt == "" {
			return fmt.Errorf("suggestion %s -> %s is missing audit rationale or references", item.Document, item.Target)
		}
	}
	return nil
}

func Evaluate(benchmark Benchmark, report codemapbench.Report) (Evaluation, error) {
	if err := ValidateLabeledBenchmark(benchmark); err != nil {
		return Evaluation{}, err
	}
	source := map[string]codemapbench.Suggestion{}
	for _, suggestion := range report.UnmatchedSuggestions {
		source[suggestion.Document+"\x00"+suggestion.Target] = suggestion
	}
	sourceRanks := map[string]int{}
	for _, candidate := range CandidatesFromReport(report) {
		sourceRanks[candidateKey(candidate)] = candidate.Rank
	}
	for _, item := range benchmark.Suggestions {
		key := item.Document + "\x00" + item.Target
		suggestion, ok := source[key]
		if !ok {
			return Evaluation{}, fmt.Errorf("benchmark suggestion is absent from source report: %s -> %s", item.Document, item.Target)
		}
		if suggestion.Score != item.Score || strings.Join(sortedEvidence(suggestion.Evidence), "\x00") != strings.Join(sortedEvidence(item.Evidence), "\x00") {
			return Evaluation{}, fmt.Errorf("benchmark evidence changed in source report: %s -> %s", item.Document, item.Target)
		}
		if sourceRanks[key] != item.Rank {
			return Evaluation{}, fmt.Errorf("benchmark rank changed in source report: %s -> %s", item.Document, item.Target)
		}
	}
	evaluation := Evaluation{
		SchemaVersion: SchemaVersion, BenchmarkSize: len(benchmark.Suggestions),
		PerDocument: map[string]DocumentMetrics{}, ByEvidenceKind: map[string]PrecisionMetrics{},
		ByScoreBucket: map[string]PrecisionMetrics{}, ByRankBucket: map[string]PrecisionMetrics{},
		SamplingCoverage: map[string]map[string]int{},
	}
	evaluation.LabelCounts.Total = len(benchmark.Suggestions)
	for _, item := range benchmark.Suggestions {
		switch item.Label {
		case ValidMissingLink:
			evaluation.LabelCounts.Valid++
		case PlausibleButUnnecessary:
			evaluation.LabelCounts.Plausible++
		case Incorrect:
			evaluation.LabelCounts.Incorrect++
		}
		addMetrics := func(table map[string]PrecisionMetrics, key string) {
			metrics := table[key]
			metrics.Total++
			if item.Label == ValidMissingLink {
				metrics.Valid++
			}
			if item.Label != Incorrect {
				metrics.Accepted++
			}
			table[key] = metrics
		}
		addMetrics(evaluation.ByEvidenceKind, item.PrimaryEvidenceKind)
		addMetrics(evaluation.ByScoreBucket, item.ScoreBucket)
		addMetrics(evaluation.ByRankBucket, item.RankBucket)
		for _, key := range []string{"area", "subsystem"} {
			value := item.Area
			if key == "subsystem" {
				value = item.Subsystem
			}
			if evaluation.SamplingCoverage[key] == nil {
				evaluation.SamplingCoverage[key] = map[string]int{}
			}
			evaluation.SamplingCoverage[key][value]++
		}
		evaluation.SamplingCoverage["score_bucket"] = incrementCoverage(evaluation.SamplingCoverage["score_bucket"], item.ScoreBucket)
		evaluation.SamplingCoverage["rank_bucket"] = incrementCoverage(evaluation.SamplingCoverage["rank_bucket"], item.RankBucket)
		evaluation.SamplingCoverage["evidence_kind"] = incrementCoverage(evaluation.SamplingCoverage["evidence_kind"], item.PrimaryEvidenceKind)
		evaluation.SamplingCoverage["document"] = incrementCoverage(evaluation.SamplingCoverage["document"], item.Document)
		for _, k := range []int{1, 3, 5} {
			if item.Rank <= k {
				metrics := evaluation.PerDocument[item.Document]
				metricsForK := metrics.PrecisionAt1
				if k == 3 {
					metricsForK = metrics.PrecisionAt3
				}
				if k == 5 {
					metricsForK = metrics.PrecisionAt5
				}
				metricsForK = addMetric(metricsForK, item)
				if k == 1 {
					metrics.PrecisionAt1 = metricsForK
				}
				if k == 3 {
					metrics.PrecisionAt3 = metricsForK
				}
				if k == 5 {
					metrics.PrecisionAt5 = metricsForK
				}
				evaluation.PerDocument[item.Document] = metrics
			}
		}
	}
	evaluation.Overall = metricsForItems(benchmark.Suggestions)
	evaluation.PrecisionAt1 = aggregateAtRank(benchmark.Suggestions, 1).OverallPrecision
	evaluation.PrecisionAt3 = aggregateAtRank(benchmark.Suggestions, 3).OverallPrecision
	evaluation.PrecisionAt5 = aggregateAtRank(benchmark.Suggestions, 5).OverallPrecision
	for key, metrics := range evaluation.ByEvidenceKind {
		evaluation.ByEvidenceKind[key] = finalize(metrics)
	}
	for key, metrics := range evaluation.ByScoreBucket {
		evaluation.ByScoreBucket[key] = finalize(metrics)
	}
	for key, metrics := range evaluation.ByRankBucket {
		evaluation.ByRankBucket[key] = finalize(metrics)
	}
	for document, metrics := range evaluation.PerDocument {
		metrics.PrecisionAt1 = finalize(metrics.PrecisionAt1)
		metrics.PrecisionAt3 = finalize(metrics.PrecisionAt3)
		metrics.PrecisionAt5 = finalize(metrics.PrecisionAt5)
		evaluation.PerDocument[document] = metrics
	}
	return evaluation, nil
}

func addMetric(metrics PrecisionMetrics, item LabeledSuggestion) PrecisionMetrics {
	metrics.Total++
	if item.Label == ValidMissingLink {
		metrics.Valid++
	}
	if item.Label != Incorrect {
		metrics.Accepted++
	}
	return metrics
}

func metricsForItems(items []LabeledSuggestion) PrecisionMetrics {
	var metrics PrecisionMetrics
	for _, item := range items {
		metrics = addMetric(metrics, item)
	}
	return finalize(metrics)
}

func aggregateAtRank(items []LabeledSuggestion, rank int) PrecisionMetrics {
	filtered := make([]LabeledSuggestion, 0)
	for _, item := range items {
		if item.Rank <= rank {
			filtered = append(filtered, item)
		}
	}
	return metricsForItems(filtered)
}

func finalize(metrics PrecisionMetrics) PrecisionMetrics {
	if metrics.Total > 0 {
		metrics.OverallPrecision = float64(metrics.Valid) / float64(metrics.Total)
		metrics.AcceptancePrecision = float64(metrics.Accepted) / float64(metrics.Total)
	}
	return metrics
}

func incrementCoverage(values map[string]int, key string) map[string]int {
	if values == nil {
		values = map[string]int{}
	}
	values[key]++
	return values
}

func decorateCandidate(suggestion codemapbench.Suggestion, rank int) Candidate {
	kinds := evidenceKinds(suggestion.Evidence)
	primary := primaryEvidenceKind(kinds)
	area, subsystem := documentArea(suggestion.Document)
	return Candidate{Suggestion: suggestion, Rank: rank, Area: area, Subsystem: subsystem,
		ScoreBucket: scoreBucket(suggestion.Score), RankBucket: rankBucket(rank),
		EvidenceKinds: kinds, PrimaryEvidenceKind: primary}
}

func uniqueCandidates(candidates []Candidate) []Candidate {
	seen := map[string]struct{}{}
	result := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		key := candidateKey(candidate)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if candidate.Rank < 1 || candidate.ScoreBucket == "" {
			candidate = decorateCandidate(candidate.Suggestion, candidate.Rank)
		}
		result = append(result, candidate)
	}
	return result
}

func candidateKey(candidate Candidate) string { return candidate.Document + "\x00" + candidate.Target }

func sampleHash(seed string, candidate Candidate) string {
	digest := sha256.Sum256([]byte(seed + "\x00" + candidateKey(candidate)))
	return hex.EncodeToString(digest[:])
}

func stratum(candidate Candidate) string {
	return candidate.Area + "\x00" + candidate.ScoreBucket + "\x00" + candidate.PrimaryEvidenceKind
}

func evidenceKinds(values []string) []string {
	set := map[string]struct{}{}
	for _, value := range values {
		kind := strings.SplitN(value, ":", 2)[0]
		if kind != "" {
			set[kind] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for kind := range set {
		result = append(result, kind)
	}
	sort.Strings(result)
	return result
}

func sortedEvidence(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func primaryEvidenceKind(kinds []string) string {
	best, bestWeight := "", -1.0
	for _, kind := range kinds {
		weight := map[string]float64{
			"declared_symbol_mention": 7, "exact_path_mention": 6, "test_counterpart": 6,
			"unique_basename_mention": 4, "dependency_neighbor": 4, "related_document_target": 4,
			"sibling_of_existing_target": 2, "git_cochange_with_existing_target": 1.5,
			"git_cochange_with_document": 1,
		}[kind]
		if weight > bestWeight || weight == bestWeight && kind < best {
			best, bestWeight = kind, weight
		}
	}
	return best
}

func scoreBucket(score float64) string {
	switch {
	case score < 1:
		return "<1"
	case score < 2:
		return "1-<2"
	case score < 8:
		return "2-<8"
	default:
		return "8+"
	}
}

func rankBucket(rank int) string {
	switch {
	case rank <= 5:
		return "1-5"
	case rank <= 10:
		return "6-10"
	case rank <= 20:
		return "11-20"
	default:
		return "21+"
	}
}

func documentArea(document string) (string, string) {
	parts := strings.Split(strings.Trim(document, "/"), "/")
	if len(parts) < 2 {
		return "other", "other"
	}
	if parts[0] == "docs" {
		area := parts[1]
		subsystem := ""
		if len(parts) > 2 {
			subsystem = strings.Join(parts[2:len(parts)-1], "/")
		}
		if subsystem == "" {
			subsystem = area
		}
		return area, subsystem
	}
	if parts[0] == "services" && len(parts) >= 3 {
		area := "services/" + parts[1]
		subsystem := strings.Join(parts[2:len(parts)-1], "/")
		if subsystem == "" {
			subsystem = parts[1]
		}
		return area, subsystem
	}
	return parts[0], strings.Join(parts[1:len(parts)-1], "/")
}

func canonicalBenchmark(benchmark Benchmark) Benchmark {
	result := benchmark
	result.Suggestions = append([]LabeledSuggestion(nil), benchmark.Suggestions...)
	sort.Slice(result.Suggestions, func(i, j int) bool {
		left, right := result.Suggestions[i], result.Suggestions[j]
		if left.Document != right.Document {
			return left.Document < right.Document
		}
		if left.Rank != right.Rank {
			return left.Rank < right.Rank
		}
		return left.Target < right.Target
	})
	return result
}

// FormatEvaluation renders the precision report consumed by the research
// report and is intentionally stable for review diffs.
func FormatEvaluation(evaluation Evaluation) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Precision benchmark evaluation\nSample size: %d\n", evaluation.BenchmarkSize)
	fmt.Fprintf(&output, "Labels: valid=%d plausible=%d incorrect=%d\n", evaluation.LabelCounts.Valid, evaluation.LabelCounts.Plausible, evaluation.LabelCounts.Incorrect)
	fmt.Fprintf(&output, "Overall precision: %.2f%%\nAcceptance precision: %.2f%%\n", evaluation.Overall.OverallPrecision*100, evaluation.Overall.AcceptancePrecision*100)
	fmt.Fprintf(&output, "Precision@1: %.2f%%\nPrecision@3: %.2f%%\nPrecision@5: %.2f%%\n", evaluation.PrecisionAt1*100, evaluation.PrecisionAt3*100, evaluation.PrecisionAt5*100)
	output.WriteString("\nPer document:\n")
	for _, document := range sortedDocumentKeys(evaluation.PerDocument) {
		metrics := evaluation.PerDocument[document]
		fmt.Fprintf(&output, "- %s: @1 %.2f%%, @3 %.2f%%, @5 %.2f%%\n", document, metrics.PrecisionAt1.OverallPrecision*100, metrics.PrecisionAt3.OverallPrecision*100, metrics.PrecisionAt5.OverallPrecision*100)
	}
	return output.String()
}

func sortedDocumentKeys(values map[string]DocumentMetrics) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// TargetFingerprint is used by curation tooling to detect target drift.
func TargetFingerprint(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
