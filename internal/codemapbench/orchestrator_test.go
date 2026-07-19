package codemapbench

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

type corpusStub struct {
	links      []Link
	linksErr   error
	inputErr   error
	requests   []DocumentRequest
	knownByDoc map[string][]string
}

func (c *corpusStub) Links(context.Context) ([]Link, error) {
	return append([]Link(nil), c.links...), c.linksErr
}

func (c *corpusStub) DocumentInput(_ context.Context, request DocumentRequest) (evidence.Input, error) {
	c.requests = append(c.requests, request)
	if c.inputErr != nil {
		return evidence.Input{}, c.inputErr
	}
	related := make([]evidence.RelatedDocument, 0, len(c.knownByDoc))
	for document, targets := range c.knownByDoc {
		related = append(related, evidence.RelatedDocument{Path: document, Targets: targets})
	}
	return evidence.Input{
		DocumentPath:     "wrong/document.md",
		RepositoryFiles:  allTargets(c.links),
		ExistingTargets:  allTargets(c.links),
		RelatedDocuments: related,
	}, nil
}

func TestRunnerOrchestratesWithoutLeakingHiddenLinks(t *testing.T) {
	known := fixtureLinks()
	config := Config{Seed: "orchestrator", HoldoutCount: 2}
	visible, hidden, _, err := splitHoldout(known, config)
	if err != nil {
		t.Fatal(err)
	}
	corpus := &corpusStub{links: known, knownByDoc: targetMap(known)}

	var collected []evidence.Input
	runner := NewRunner(corpus, config)
	runner.Collector = CandidateCollectorFunc(func(_ context.Context, input evidence.Input) ([]evidence.Candidate, error) {
		collected = append(collected, input)
		result := make([]evidence.Candidate, 0)
		for _, link := range hidden {
			if link.Document == input.DocumentPath {
				result = append(result, evidence.Candidate{
					Path:     link.Target,
					Evidence: []evidence.Evidence{{Kind: evidence.KindExactPathMention, Count: 1}},
				})
			}
		}
		return result, nil
	})

	report, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.RecoveredLinks, hidden) || report.Recall != 1 || report.Precision != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if len(corpus.requests) != 3 || len(collected) != 3 {
		t.Fatalf("expected one request per document, got requests=%d inputs=%d", len(corpus.requests), len(collected))
	}

	visibleMap := targetMap(visible)
	for index, request := range corpus.requests {
		if containsAny(request.VisibleLinks, hidden) {
			t.Fatalf("request %d leaked hidden links: %#v", index, request)
		}
		if !reflect.DeepEqual(request.VisibleTargets, visibleMap[request.Document]) {
			t.Fatalf("request %d visible targets = %#v, want %#v", index, request.VisibleTargets, visibleMap[request.Document])
		}
		input := collected[index]
		if input.DocumentPath != request.Document {
			t.Fatalf("collector document = %q, want %q", input.DocumentPath, request.Document)
		}
		if !reflect.DeepEqual(input.ExistingTargets, visibleMap[request.Document]) {
			t.Fatalf("collector targets = %#v, want %#v", input.ExistingTargets, visibleMap[request.Document])
		}
		for _, related := range input.RelatedDocuments {
			if !reflect.DeepEqual(related.Targets, visibleMap[related.Path]) {
				t.Fatalf("related document leaked targets: %#v", related)
			}
		}
	}
}

func TestRunnerUsesDefaultCollectorAndSuggestionBuilder(t *testing.T) {
	known := []Link{
		{Document: "docs/a.md", Target: "src/a.go"},
		{Document: "docs/a.md", Target: "src/b.go"},
	}
	config := Config{Seed: "defaults", HoldoutCount: 1}
	_, hidden, _, err := splitHoldout(known, config)
	if err != nil {
		t.Fatal(err)
	}
	corpus := &corpusStub{links: known, knownByDoc: targetMap(known)}

	// The corpus deliberately includes the hidden path in repository files and
	// prose, but not in ExistingTargets. The built-in collector must recover it.
	runner := NewRunner(documentTextCorpus{corpusStub: corpus, text: hidden[0].Target}, config)
	report, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.RecoveredLinks, hidden) {
		t.Fatalf("recovered links = %#v, want %#v", report.RecoveredLinks, hidden)
	}
	if len(report.RecoveredSuggestions) != 1 || len(report.RecoveredSuggestions[0].Evidence) == 0 {
		t.Fatalf("expected evidence-bearing suggestion: %#v", report.RecoveredSuggestions)
	}
}

type documentTextCorpus struct {
	*corpusStub
	text string
}

func (c documentTextCorpus) DocumentInput(ctx context.Context, request DocumentRequest) (evidence.Input, error) {
	input, err := c.corpusStub.DocumentInput(ctx, request)
	input.DocumentText = c.text
	return input, err
}

func TestRunnerWrapsStageErrors(t *testing.T) {
	known := fixtureLinks()
	cases := []struct {
		name   string
		runner Runner
		want   string
	}{
		{
			name:   "links",
			runner: NewRunner(&corpusStub{linksErr: errors.New("links failed")}, Config{}),
			want:   "load benchmark links",
		},
		{
			name:   "input",
			runner: NewRunner(&corpusStub{links: known, inputErr: errors.New("input failed")}, Config{HoldoutCount: 1}),
			want:   "build evidence input for",
		},
		{
			name: "collector",
			runner: Runner{
				Corpus: &corpusStub{links: known}, Config: Config{HoldoutCount: 1},
				Collector: CandidateCollectorFunc(func(context.Context, evidence.Input) ([]evidence.Candidate, error) {
					return nil, errors.New("collect failed")
				}),
			},
			want: "collect evidence for",
		},
		{
			name: "suggestions",
			runner: Runner{
				Corpus: &corpusStub{links: known}, Config: Config{HoldoutCount: 1},
				Builder: SuggestionBuilderFunc(func(context.Context, string, []evidence.Candidate) ([]Suggestion, error) {
					return nil, errors.New("suggest failed")
				}),
			},
			want: "build suggestions for",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.runner.Run(context.Background())
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestRunnerRequiresCorpus(t *testing.T) {
	_, err := (Runner{}).Run(context.Background())
	if err == nil || err.Error() != "benchmark corpus is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func targetMap(links []Link) map[string][]string {
	result := linksByDocument(links)
	return result
}

func allTargets(links []Link) []string {
	result := make([]string, 0, len(links))
	for _, link := range links {
		result = append(result, link.Target)
	}
	return result
}
