package codemaprun

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapcorpus"
	"github.com/Lokee86/demon-docs/internal/codemaprecommend"
	"github.com/Lokee86/demon-docs/internal/evidence"
	"github.com/Lokee86/demon-docs/internal/filetxn"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func Build(ctx context.Context, options Options) (Plan, error) {
	format := codemap.DefaultFormat()
	format.SectionHeadings = append([]string(nil), options.Headings...)
	dataset, err := codemap.BuildDataset(options.RepositoryRoot, options.DocsRoot, format)
	if err != nil {
		return Plan{}, err
	}
	corpus, err := codemapcorpus.Build(options.RepositoryRoot, dataset, codemapcorpus.Options{})
	if err != nil {
		return Plan{}, fmt.Errorf("build codemap corpus: %w", err)
	}
	policy, err := review.LoadPolicy(options.RepositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	entriesByDocument := datasetEntriesByDocument(dataset)
	files := append([]string(nil), options.TargetFiles...)
	sort.Strings(files)
	plan := Plan{Documents: make([]DocumentPlan, 0, len(files))}
	for _, filePath := range files {
		if err := ctx.Err(); err != nil {
			return Plan{}, err
		}
		documentPath, err := filepath.Rel(options.RepositoryRoot, filePath)
		if err != nil {
			return Plan{}, err
		}
		documentPath = filepath.ToSlash(filepath.Clean(documentPath))
		document, err := textio.Read(filePath)
		if err != nil {
			return Plan{}, err
		}
		docPlan, err := buildDocument(ctx, documentPath, document, format, corpus, entriesByDocument[documentPath], policy, options)
		if err != nil {
			return Plan{}, fmt.Errorf("plan codemap %s: %w", documentPath, err)
		}
		plan.Documents = append(plan.Documents, docPlan)
		if docPlan.Changed {
			plan.Rewrites = append(plan.Rewrites, filetxn.New(filePath, docPlan.Before, docPlan.After))
		}
	}
	return plan, nil
}

func buildDocument(
	ctx context.Context,
	documentPath string,
	document textio.Document,
	format codemap.Format,
	corpus codemapcorpus.Corpus,
	entries []codemap.DatasetEntry,
	policy review.Policy,
	options Options,
) (DocumentPlan, error) {
	result := DocumentPlan{
		Path:     documentPath,
		Existing: authoredTargets(entries),
		Before:   document.Encode(document.Text),
		After:    document.Encode(document.Text),
	}
	hasSection, err := codemap.HasSection(document.Text, format)
	if err != nil {
		return DocumentPlan{}, err
	}
	if !hasSection && options.Schema == nil {
		return result, nil
	}

	existingTargets := corpus.KnownTargets(documentPath)
	recommendations, err := recommend(ctx, corpus, format, documentPath, existingTargets)
	if err != nil {
		return DocumentPlan{}, err
	}
	addTargets := make([]string, 0, len(recommendations))
	for _, item := range recommendations {
		suggestion := review.CodemapSuggestion(item.Document, item.Target, item.Score, string(item.Tier), item.Evidence)
		applied := policy.ApplySuggestion(suggestion)
		declined := applied.Status == review.StatusDeclined || len(applied.Candidates) == 1 && applied.Candidates[0].Declined
		result.Recommendations = append(result.Recommendations, Recommendation{Suggestion: item, Declined: declined})
		if declined {
			result.Suppressed = append(result.Suppressed, item.Target)
			continue
		}
		addTargets = append(addTargets, item.Target)
	}

	removeTargets := []string{}
	if options.RemoveUndiscoveredLinks || options.RemoveLowScoreLinks {
		removeTargets, err = plannedRemovals(ctx, corpus, format, documentPath, existingTargets, entries, options)
		if err != nil {
			return DocumentPlan{}, err
		}
	}
	managed, err := codemap.ReconcileManaged(documentPath, document.Text, format, options.MarkerPrefix, codemap.ManagedUpdate{
		AddTargets: addTargets, RemoveTargets: removeTargets,
	}, options.Schema)
	if err != nil {
		return DocumentPlan{}, err
	}
	result.SectionFound = managed.SectionFound
	result.SectionCreated = managed.SectionCreated
	result.Added = managed.Added
	result.Removed = managed.Removed
	result.After = document.Encode(managed.Text)
	result.Changed = string(result.Before) != string(result.After)
	sort.Strings(result.Suppressed)
	return result, nil
}

func recommend(ctx context.Context, corpus codemapcorpus.Corpus, format codemap.Format, documentPath string, existingTargets []string) ([]codemaprecommend.Suggestion, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	input, err := corpus.Input(documentPath, existingTargets)
	if err != nil {
		return nil, err
	}
	input.DocumentText = codemap.StripAuthoredSections(input.DocumentText, format)
	return codemaprecommend.SuggestionsFromEvidence(documentPath, evidence.Collect(input)), nil
}

func plannedRemovals(
	ctx context.Context,
	corpus codemapcorpus.Corpus,
	format codemap.Format,
	documentPath string,
	existingTargets []string,
	entries []codemap.DatasetEntry,
	options Options,
) ([]string, error) {
	removals := map[string]struct{}{}
	for _, entry := range entries {
		if entry.Resolution.Status != codemap.ResolutionResolved || entry.Resolution.ResolvedPath == "" {
			continue
		}
		visible := withoutTarget(existingTargets, entry.Resolution.ResolvedPath)
		items, err := recommend(ctx, corpus, format, documentPath, visible)
		if err != nil {
			return nil, err
		}
		found := false
		lowScore := false
		for _, item := range items {
			if cleanTarget(item.Target) == cleanTarget(entry.Resolution.ResolvedPath) {
				found = true
				lowScore = item.Tier == codemaprecommend.SuggestionTierContext
				break
			}
		}
		if !found && options.RemoveUndiscoveredLinks || found && lowScore && options.RemoveLowScoreLinks {
			removals[entry.Entry.Target] = struct{}{}
		}
	}
	result := make([]string, 0, len(removals))
	for target := range removals {
		result = append(result, target)
	}
	sort.Strings(result)
	return result, nil
}

func datasetEntriesByDocument(dataset codemap.Dataset) map[string][]codemap.DatasetEntry {
	result := map[string][]codemap.DatasetEntry{}
	for _, entry := range dataset.Entries {
		result[entry.Entry.DocumentPath] = append(result[entry.Entry.DocumentPath], entry)
	}
	return result
}

func authoredTargets(entries []codemap.DatasetEntry) []string {
	set := map[string]struct{}{}
	for _, entry := range entries {
		set[entry.Entry.Target] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for target := range set {
		result = append(result, target)
	}
	sort.Strings(result)
	return result
}

func withoutTarget(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if cleanTarget(value) != cleanTarget(target) {
			result = append(result, value)
		}
	}
	return result
}

func cleanTarget(value string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(value)), "./")
}
