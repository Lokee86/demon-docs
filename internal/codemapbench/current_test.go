package codemapbench

import (
	"context"
	"reflect"
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

type currentTestCorpus struct {
	links    []Link
	requests []DocumentRequest
}

func (c *currentTestCorpus) Links(context.Context) ([]Link, error) {
	return append([]Link(nil), c.links...), nil
}

func (c *currentTestCorpus) DocumentInput(_ context.Context, request DocumentRequest) (evidence.Input, error) {
	c.requests = append(c.requests, request)
	return evidence.Input{DocumentPath: request.Document}, nil
}

func TestSuggestCurrentUsesAllAuthoredLinksAsVisible(t *testing.T) {
	corpus := &currentTestCorpus{links: []Link{
		{Document: "docs/a.md", Target: "existing/a.go"},
		{Document: "docs/b.md", Target: "existing/b.go"},
	}}
	runner := Runner{
		Corpus: corpus,
		Collector: CandidateCollectorFunc(func(_ context.Context, input evidence.Input) ([]evidence.Candidate, error) {
			return []evidence.Candidate{{Path: "new/" + input.DocumentPath + ".go"}}, nil
		}),
		Builder: SuggestionBuilderFunc(func(_ context.Context, document string, candidates []evidence.Candidate) ([]Suggestion, error) {
			return []Suggestion{
				{Link: Link{Document: document, Target: candidates[0].Path}},
				{Link: Link{Document: document, Target: map[string]string{"docs/a.md": "existing/a.go", "docs/b.md": "existing/b.go"}[document]}},
			}, nil
		}),
	}

	report, err := runner.SuggestCurrent(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.requests) != 2 {
		t.Fatalf("got %d document requests", len(corpus.requests))
	}
	for _, request := range corpus.requests {
		var want []string
		if request.Document == "docs/a.md" {
			want = []string{"existing/a.go"}
		} else {
			want = []string{"existing/b.go"}
		}
		if !reflect.DeepEqual(request.VisibleTargets, want) {
			t.Fatalf("%s visible targets = %#v, want %#v", request.Document, request.VisibleTargets, want)
		}
	}
	if len(report.HiddenLinks) != 0 || len(report.RecoveredLinks) != 0 || report.Recall != 0 {
		t.Fatalf("current run unexpectedly contains holdout scoring: %#v", report)
	}
	if len(report.UnmatchedSuggestions) != 2 {
		t.Fatalf("unmatched suggestions = %#v", report.UnmatchedSuggestions)
	}
	if len(report.AlreadyLinked) != 2 {
		t.Fatalf("already linked suggestions = %#v", report.AlreadyLinked)
	}
	if !reflect.DeepEqual(report.VisibleLinks, report.KnownLinks) {
		t.Fatalf("visible links differ from authored links")
	}
}
