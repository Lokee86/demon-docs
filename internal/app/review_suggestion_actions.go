package app

import (
	"context"
	"fmt"
	"io"

	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/review"
)

func runSuggestionsSelect(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) < 1 || len(args) > 2 {
		fmt.Fprintln(errOut, "usage: ddocs suggestions select SUGGESTION [CANDIDATE]")
		return 2
	}
	runtime, code := loadReviewRuntime(ctx, errOut)
	if code != 0 {
		return code
	}
	suggestion, ok := findSuggestion(runtime.suggestions, args[0])
	if !ok {
		return fail(errOut, fmt.Errorf("current suggestion not found: %s", args[0]))
	}
	if suggestion.Status == review.StatusDeclined || suggestion.Status == review.StatusBlocked {
		return fail(errOut, fmt.Errorf("suggestion %s must be reconsidered or unblocked before selection", suggestion.ID))
	}
	selector := ""
	if len(args) == 2 {
		selector = args[1]
	}
	var applied []review.Change
	if len(runtime.linkPlan.Rewrites) > 0 {
		if _, err := links.ApplyAndSave(&runtime.linkPlan); err != nil {
			return fail(errOut, err)
		}
		applied = append(applied, runtime.linkPlan.AppliedChanges...)
		runtime, code = loadReviewRuntime(ctx, errOut)
		if code != 0 {
			return code
		}
		suggestion, ok = findSuggestion(runtime.suggestions, args[0])
		if !ok {
			return fail(errOut, fmt.Errorf("suggestion changed after deterministic repairs: %s", args[0]))
		}
	}
	candidate, err := chooseCandidate(suggestion, selector)
	if err != nil {
		return fail(errOut, err)
	}
	if candidate.Declined {
		return fail(errOut, fmt.Errorf("candidate %d is declined; reconsider the suggestion first", candidate.Index))
	}
	if suggestion.Kind == review.SuggestionLinkRepair {
		err = links.ApplySelectedSuggestion(&runtime.linkPlan, suggestion, candidate)
	} else {
		err = applyCodemapSelection(&runtime, suggestion, candidate)
	}
	if err != nil {
		return fail(errOut, err)
	}
	if _, err := links.ApplyAndSave(&runtime.linkPlan); err != nil {
		return fail(errOut, err)
	}
	applied = append(applied, runtime.linkPlan.AppliedChanges...)
	for _, change := range applied {
		fmt.Fprintf(out, "applied %s  %s  %s\n", change.ID, change.Kind, change.SourcePath)
	}
	return 0
}

func runSuggestionsDecline(ctx context.Context, args []string, out, errOut io.Writer) int {
	positional, reason, err := splitReason(args)
	if err != nil || len(positional) < 1 || len(positional) > 2 {
		fmt.Fprintln(errOut, "usage: ddocs suggestions decline SUGGESTION [CANDIDATE] [--reason TEXT]")
		return 2
	}
	runtime, code := loadReviewRuntime(ctx, errOut)
	if code != 0 {
		return code
	}
	suggestion, ok := findSuggestion(runtime.suggestions, positional[0])
	if !ok {
		return fail(errOut, fmt.Errorf("current suggestion not found: %s", positional[0]))
	}
	decision := review.Decision{
		Action:       review.DecisionDeclineIssue,
		RelationKey:  suggestion.RelationKey,
		Fingerprint:  suggestion.Fingerprint,
		SuggestionID: suggestion.ID,
		Reason:       reason,
		Suggestion:   &suggestion,
	}
	if len(positional) == 2 {
		candidate, err := chooseCandidate(suggestion, positional[1])
		if err != nil {
			return fail(errOut, err)
		}
		decision.Action = review.DecisionDeclineCandidate
		decision.CandidateTarget = candidate.Target
		decision.CandidateFingerprint = candidate.Fingerprint
	}
	stored, err := appendDecision(runtime.scope.RepositoryRoot, decision)
	if err != nil {
		return fail(errOut, err)
	}
	fmt.Fprintf(out, "%s %s\n", stored.Decision.Action, stored.Decision.ID)
	return 0
}

func runSuggestionsReconsider(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: ddocs suggestions reconsider SUGGESTION")
		return 2
	}
	runtime, code := loadReviewRuntime(ctx, errOut)
	if code != 0 {
		return code
	}
	suggestion, ok := findSuggestion(runtime.suggestions, args[0])
	if !ok {
		var err error
		suggestion, err = historicalSuggestion(runtime.scope.RepositoryRoot, args[0])
		if err != nil {
			return fail(errOut, err)
		}
	}
	stored, err := appendDecision(runtime.scope.RepositoryRoot, review.Decision{
		Action:       review.DecisionReconsider,
		RelationKey:  suggestion.RelationKey,
		SuggestionID: suggestion.ID,
		Suggestion:   &suggestion,
	})
	if err != nil {
		return fail(errOut, err)
	}
	fmt.Fprintf(out, "reconsidered %s with %s\n", suggestion.ID, stored.Decision.ID)
	return 0
}
