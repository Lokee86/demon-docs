package codemapbench

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestRunClassifiesSuggestionsAndDoesNotLeakHiddenLinks(t *testing.T) {
	known := fixtureLinks()
	config := Config{Seed: "classification", HoldoutCount: 2}
	visible, hidden, _, err := splitHoldout(known, config)
	if err != nil {
		t.Fatal(err)
	}

	var request Request
	generator := GeneratorFunc(func(_ context.Context, got Request) ([]Suggestion, error) {
		request = got
		return []Suggestion{
			{Link: hidden[0], Score: 0.9},
			{Link: visible[0], Score: 0.8},
			{Link: Link{Document: "docs/a.md", Target: "src/unmatched.go"}, Score: 0.7},
			{Link: hidden[0], Score: 0.6},
			{Link: Link{Document: "", Target: "src/invalid.go"}},
		}, nil
	})

	report, err := Run(context.Background(), known, generator, config)
	if err != nil {
		t.Fatal(err)
	}

	if containsAny(request.VisibleLinks, hidden) {
		t.Fatalf("generator request leaked hidden links: %#v", request)
	}
	if !reflect.DeepEqual(request.VisibleLinks, visible) {
		t.Fatalf("unexpected visible request links: %#v", request.VisibleLinks)
	}
	if !reflect.DeepEqual(request.Documents, []string{"docs/a.md", "docs/b.md", "docs/c.md"}) {
		t.Fatalf("unexpected document list: %#v", request.Documents)
	}

	if report.RawSuggestionCount != 5 || report.UniqueSuggestionCount != 3 {
		t.Fatalf("unexpected suggestion counts: %#v", report)
	}
	if len(report.RecoveredLinks) != 1 || len(report.RecoveredSuggestions) != 1 || len(report.MissedLinks) != 1 {
		t.Fatalf("unexpected recovery result: %#v", report)
	}
	if report.RecoveredSuggestions[0].Score != 0.9 {
		t.Fatalf("recovered suggestion lost score: %#v", report.RecoveredSuggestions[0])
	}
	if len(report.AlreadyLinked) != 1 || len(report.UnmatchedSuggestions) != 1 {
		t.Fatalf("unexpected classifications: %#v", report)
	}
	if len(report.DuplicateSuggestions) != 1 || len(report.InvalidSuggestions) != 1 {
		t.Fatalf("unexpected rejected output: %#v", report)
	}
	if math.Abs(report.Precision-(1.0/3.0)) > 0.000001 {
		t.Fatalf("got precision %f, want %f", report.Precision, 1.0/3.0)
	}
	if report.Recall != 0.5 {
		t.Fatalf("got recall %f, want 0.5", report.Recall)
	}
}

func TestRunIsStableAcrossKnownLinkOrder(t *testing.T) {
	known := fixtureLinks()
	generator := GeneratorFunc(func(_ context.Context, request Request) ([]Suggestion, error) {
		result := make([]Suggestion, 0, len(request.VisibleLinks))
		for _, link := range request.VisibleLinks {
			result = append(result, Suggestion{Link: link})
		}
		return result, nil
	})
	config := Config{Seed: "stable-report", HoldoutCount: 2}

	first, err := Run(context.Background(), known, generator, config)
	if err != nil {
		t.Fatal(err)
	}
	reversed := append([]Link(nil), known...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	second, err := Run(context.Background(), reversed, generator, config)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("report changed with known-link order:\n%#v\n%#v", first, second)
	}
}

func TestRunWrapsGeneratorError(t *testing.T) {
	want := errors.New("collector failed")
	_, err := Run(context.Background(), fixtureLinks(), GeneratorFunc(
		func(context.Context, Request) ([]Suggestion, error) {
			return nil, want
		},
	), Config{HoldoutCount: 1})
	if !errors.Is(err, want) {
		t.Fatalf("got error %v, want wrapped %v", err, want)
	}
}

func containsAny(haystack, needles []Link) bool {
	set := linkSet(haystack)
	for _, needle := range needles {
		if _, ok := set[linkKey(needle)]; ok {
			return true
		}
	}
	return false
}
