package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapbench"
	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/config"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/review"
)

type reviewRuntime struct {
	config      config.Config
	scope       repository.Scope
	linkPlan    links.Plan
	suggestions []review.Suggestion
}

func loadReviewRuntime(ctx context.Context, errOut io.Writer) (reviewRuntime, int) {
	resolved, configPath, code := load(commonFlags{}, errOut)
	if code != 0 {
		return reviewRuntime{}, code
	}
	scope, err := resolveScope(optionalString{}, resolved.Root, configPath)
	if err != nil {
		return reviewRuntime{}, fail(errOut, err)
	}
	plan, err := links.Reconcile(scope.RepositoryRoot)
	if err != nil {
		return reviewRuntime{}, fail(errOut, err)
	}
	linkSuggestions, err := links.ReviewSuggestions(plan)
	if err != nil {
		return reviewRuntime{}, fail(errOut, err)
	}
	codemapSuggestions, err := currentCodemapSuggestions(ctx, scope, resolved)
	if err != nil {
		return reviewRuntime{}, fail(errOut, err)
	}
	all := append(linkSuggestions, codemapSuggestions...)
	sort.Slice(all, func(i, j int) bool {
		if all[i].SourcePath != all[j].SourcePath {
			return all[i].SourcePath < all[j].SourcePath
		}
		if all[i].Kind != all[j].Kind {
			return all[i].Kind < all[j].Kind
		}
		return all[i].ID < all[j].ID
	})
	return reviewRuntime{config: resolved, scope: scope, linkPlan: plan, suggestions: all}, 0
}

func currentCodemapSuggestions(ctx context.Context, scope repository.Scope, resolved config.Config) ([]review.Suggestion, error) {
	if !repository.DocsRootExists(scope) {
		return nil, nil
	}
	format := codemap.DefaultFormat()
	format.SectionHeadings = append([]string(nil), resolved.Codemap.Headings...)
	dataset, err := codemap.BuildDataset(scope.RepositoryRoot, scope.DocsRoot, format)
	if err != nil {
		return nil, err
	}
	if len(dataset.Entries) == 0 {
		return nil, nil
	}
	corpus, err := codemapcorpus.Build(scope.RepositoryRoot, dataset, codemapcorpus.Options{})
	if err != nil {
		return nil, fmt.Errorf("build codemap suggestion corpus: %w", err)
	}
	runner := codemapbench.NewRunner(benchmarkCorpus{
		links:  codemapbench.ResolvedLinksFromDataset(dataset),
		corpus: corpus,
		format: format,
	}, codemapbench.Config{})
	report, err := runner.SuggestCurrent(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := review.LoadPolicy(scope.RepositoryRoot)
	if err != nil {
		return nil, err
	}
	result := make([]review.Suggestion, 0, len(report.UnmatchedSuggestions))
	for _, item := range report.UnmatchedSuggestions {
		suggestion := review.CodemapSuggestion(item.Document, item.Target, item.Score, string(item.Tier), item.Evidence)
		result = append(result, policy.ApplySuggestion(suggestion))
	}
	return result, nil
}

func reviewPath(repositoryRoot, input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", nil
	}
	path := input
	if !filepath.IsAbs(path) {
		path = filepath.Join(repositoryRoot, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(repositoryRoot, absolute)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path is outside the repository: %s", input)
	}
	return filepath.ToSlash(filepath.Clean(relative)), nil
}

func findSuggestion(suggestions []review.Suggestion, id string) (review.Suggestion, bool) {
	for _, suggestion := range suggestions {
		if suggestion.ID == id {
			return suggestion, true
		}
	}
	return review.Suggestion{}, false
}

func chooseCandidate(suggestion review.Suggestion, selector string) (review.Candidate, error) {
	if selector == "" && len(suggestion.Candidates) == 1 {
		return suggestion.Candidates[0], nil
	}
	if index, err := strconv.Atoi(selector); err == nil {
		for _, candidate := range suggestion.Candidates {
			if candidate.Index == index {
				return candidate, nil
			}
		}
	}
	for _, candidate := range suggestion.Candidates {
		if candidate.Target == selector {
			return candidate, nil
		}
	}
	return review.Candidate{}, fmt.Errorf("candidate not found for %s: %s", suggestion.ID, selector)
}

func appendDecision(repositoryRoot string, decision review.Decision) (review.StoredEvent, error) {
	store, err := review.Open(repositoryRoot)
	if err != nil {
		return review.StoredEvent{}, err
	}
	if decision.ID == "" {
		decision.ID = review.NewID("dc")
	}
	if decision.DecidedAt.IsZero() {
		decision.DecidedAt = time.Now().UTC()
	}
	return store.Append(review.Event{Type: review.EventDecision, Time: decision.DecidedAt, Decision: &decision}, nil, nil)
}

func splitReason(args []string) ([]string, string, error) {
	var positional []string
	reason := ""
	for index := 0; index < len(args); index++ {
		if args[index] != "--reason" {
			positional = append(positional, args[index])
			continue
		}
		if index+1 >= len(args) {
			return nil, "", fmt.Errorf("--reason requires a value")
		}
		reason = args[index+1]
		index++
	}
	return positional, strings.TrimSpace(reason), nil
}

func currentFilePath(root string, plan links.Plan, change review.Change) string {
	for _, file := range plan.Files.Files {
		if file.ID == change.SourceFileID && file.Present {
			return filepath.Join(root, filepath.FromSlash(file.Path))
		}
	}
	return filepath.Join(root, filepath.FromSlash(change.SourcePath))
}

type trackedReviewFilter struct {
	Path   string
	FileID string
}

func resolveTrackedReviewFilter(root string, plan links.Plan, args []string) (trackedReviewFilter, error) {
	path, err := optionalReviewFilter(root, args)
	if err != nil || path == "" {
		return trackedReviewFilter{Path: path}, err
	}
	for _, file := range plan.Files.Files {
		if filepath.ToSlash(filepath.Clean(file.Path)) == path {
			return trackedReviewFilter{Path: path, FileID: file.ID}, nil
		}
		for _, historical := range file.PathHistory {
			if filepath.ToSlash(filepath.Clean(historical)) == path {
				return trackedReviewFilter{Path: path, FileID: file.ID}, nil
			}
		}
	}
	return trackedReviewFilter{Path: path}, nil
}

func fileMatches(path string, filter string) bool {
	return filter == "" || filepath.ToSlash(filepath.Clean(path)) == filter
}

func sourceMatches(change review.Change, filter trackedReviewFilter) bool {
	return filter.Path == "" || filter.FileID != "" && change.SourceFileID == filter.FileID || fileMatches(change.SourcePath, filter.Path)
}

func eventTime(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}
	return value.Local().Format("2006-01-02 15:04:05")
}

func readCurrent(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return data, nil
}
