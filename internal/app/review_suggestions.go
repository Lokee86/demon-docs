package app

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/review"
)

func suggestionsHelp(out io.Writer) {
	fmt.Fprintln(out, "usage: ddocs suggestions [-h] [FILE]\n       ddocs suggestions {declined,log,show,select,decline,reconsider} ...\n\nInspect and decide ambiguous link repairs and codemap missing-link suggestions. With no subcommand, current suggestions are listed and may be filtered by one repository-relative source FILE. Selecting a candidate converts it into a normal hash-guarded repair.\n\ncommands:\n  declined [FILE]                 show effective declined suggestions\n  log [FILE]                      show suggestion decision history\n  show SUGGESTION                 show one current or historical suggestion\n  select SUGGESTION [CANDIDATE]   apply the selected repair\n  decline SUGGESTION [CANDIDATE] [--reason TEXT]\n                                  decline one candidate or the whole issue\n  reconsider SUGGESTION           clear effective declines for the issue\n\noptions:\n  -h, --help                      show this help message and exit\n\nRun `ddocs suggestions <command> --help` for candidate selection, persistence, and mutation details.")
}

func runSuggestions(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		suggestionsHelp(out)
		return 0
	}
	if len(args) > 1 && isSuggestionCommand(args[0]) && helpRequested(args[1:]) {
		suggestionCommandHelp(out, args[0])
		return 0
	}
	if len(args) == 0 || !isSuggestionCommand(args[0]) {
		return runSuggestionsList(ctx, args, out, errOut)
	}
	switch args[0] {
	case "declined":
		return runSuggestionsDeclined(ctx, args[1:], out, errOut)
	case "log":
		return runSuggestionsLog(args[1:], out, errOut)
	case "show":
		return runSuggestionsShow(ctx, args[1:], out, errOut)
	case "select":
		return runSuggestionsSelect(ctx, args[1:], out, errOut)
	case "decline":
		return runSuggestionsDecline(ctx, args[1:], out, errOut)
	case "reconsider":
		return runSuggestionsReconsider(ctx, args[1:], out, errOut)
	default:
		return 2
	}
}

func isSuggestionCommand(value string) bool {
	switch value {
	case "declined", "log", "show", "select", "decline", "reconsider":
		return true
	default:
		return false
	}
}

func runSuggestionsList(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) > 1 {
		writeUnrecognized(errOut, args[1:])
		return 2
	}
	runtime, code := loadReviewRuntime(ctx, errOut)
	if code != 0 {
		return code
	}
	filter, err := optionalReviewFilter(runtime.scope.RepositoryRoot, args)
	if err != nil {
		return fail(errOut, err)
	}
	count := 0
	for _, suggestion := range runtime.suggestions {
		if !fileMatches(suggestion.SourcePath, filter) {
			continue
		}
		writeSuggestion(out, suggestion)
		count++
	}
	if count == 0 {
		fmt.Fprintln(out, "no current suggestions")
	}
	return 0
}

func runSuggestionsDeclined(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) > 1 {
		writeUnrecognized(errOut, args[1:])
		return 2
	}
	runtime, code := loadReviewRuntime(ctx, errOut)
	if code != 0 {
		return code
	}
	filter, err := optionalReviewFilter(runtime.scope.RepositoryRoot, args)
	if err != nil {
		return fail(errOut, err)
	}
	policy, err := review.LoadPolicy(runtime.scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	decisions := policy.DeclinedSuggestions()
	count := 0
	for _, decision := range decisions {
		if decision.Suggestion == nil || !fileMatches(decision.Suggestion.SourcePath, filter) {
			continue
		}
		fmt.Fprintf(out, "%s  %s  %s\n", decision.Suggestion.ID, decision.Action, decision.Suggestion.SourcePath)
		if decision.CandidateTarget != "" {
			fmt.Fprintf(out, "  candidate: %s\n", decision.CandidateTarget)
		}
		if decision.Reason != "" {
			fmt.Fprintf(out, "  reason: %s\n", decision.Reason)
		}
		fmt.Fprintf(out, "  decided: %s\n", eventTime(decision.DecidedAt))
		count++
	}
	if count == 0 {
		fmt.Fprintln(out, "no declined suggestions")
	}
	return 0
}

func runSuggestionsLog(args []string, out, errOut io.Writer) int {
	if len(args) > 1 {
		writeUnrecognized(errOut, args[1:])
		return 2
	}
	resolved, configPath, code := load(commonFlags{}, errOut)
	if code != 0 {
		return code
	}
	scope, err := resolveScope(optionalString{}, resolved.Root, configPath)
	if err != nil {
		return fail(errOut, err)
	}
	filter, err := optionalReviewFilter(scope.RepositoryRoot, args)
	if err != nil {
		return fail(errOut, err)
	}
	store, err := review.Open(scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	history, err := store.History(0)
	if err != nil {
		return fail(errOut, err)
	}
	count := 0
	for _, event := range history {
		decision := event.Decision
		if decision == nil || decision.Suggestion == nil || !fileMatches(decision.Suggestion.SourcePath, filter) {
			continue
		}
		fmt.Fprintf(out, "%s  %s  %s  %s\n", decision.ID, eventTime(decision.DecidedAt), decision.Action, decision.Suggestion.ID)
		if decision.CandidateTarget != "" {
			fmt.Fprintf(out, "  candidate: %s\n", decision.CandidateTarget)
		}
		if decision.Reason != "" {
			fmt.Fprintf(out, "  reason: %s\n", decision.Reason)
		}
		count++
	}
	if count == 0 {
		fmt.Fprintln(out, "no suggestion decision history")
	}
	return 0
}

func runSuggestionsShow(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: ddocs suggestions show SUGGESTION")
		return 2
	}
	runtime, code := loadReviewRuntime(ctx, errOut)
	if code != 0 {
		return code
	}
	if suggestion, ok := findSuggestion(runtime.suggestions, args[0]); ok {
		writeSuggestion(out, suggestion)
		return 0
	}
	suggestion, err := historicalSuggestion(runtime.scope.RepositoryRoot, args[0])
	if err != nil {
		return fail(errOut, err)
	}
	writeSuggestion(out, suggestion)
	return 0
}

func optionalReviewFilter(root string, args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	return reviewPath(root, args[0])
}

func historicalSuggestion(root, id string) (review.Suggestion, error) {
	store, err := review.Open(root)
	if err != nil {
		return review.Suggestion{}, err
	}
	history, err := store.History(0)
	if err != nil {
		return review.Suggestion{}, err
	}
	for _, event := range history {
		if event.Decision != nil && event.Decision.Suggestion != nil && event.Decision.Suggestion.ID == id {
			return *event.Decision.Suggestion, nil
		}
	}
	return review.Suggestion{}, fmt.Errorf("suggestion not found: %s", id)
}

func writeSuggestion(out io.Writer, suggestion review.Suggestion) {
	fmt.Fprintf(out, "%s  %s  %s  %s\n", suggestion.ID, strings.ToUpper(string(suggestion.Status)), suggestion.Kind, suggestion.SourcePath)
	if suggestion.BrokenTarget != "" {
		fmt.Fprintf(out, "  broken target: %s\n", suggestion.BrokenTarget)
	}
	candidates := append([]review.Candidate(nil), suggestion.Candidates...)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Index < candidates[j].Index })
	for _, candidate := range candidates {
		state := ""
		if candidate.Declined {
			state = " [declined]"
		} else if candidate.Stale {
			state = " [stale decline]"
		}
		fmt.Fprintf(out, "  %d. %s%s\n", candidate.Index, candidate.Target, state)
		if candidate.Tier != "" || candidate.Score != 0 {
			fmt.Fprintf(out, "     tier: %s  score: %.2f\n", candidate.Tier, candidate.Score)
		}
	}
	if suggestion.Reason != "" {
		fmt.Fprintf(out, "  reason: %s\n", suggestion.Reason)
	}
}
