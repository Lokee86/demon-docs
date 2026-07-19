package links

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func ReviewSuggestions(plan Plan) ([]review.Suggestion, error) {
	policy, err := review.LoadPolicy(plan.RepositoryRoot)
	if err != nil {
		return nil, err
	}
	var suggestions []review.Suggestion
	for _, record := range plan.Links.Links {
		if record.Status != "ambiguous" && record.Status != "blocked" && record.Status != "stale_block" {
			continue
		}
		suggestion := review.LinkSuggestion(record.SourceFileID, record.SourcePath, record.ID, record.Target, record.Candidates)
		suggestion.Line = record.Line
		suggestion.Column = record.Column
		if record.Status == "ambiguous" {
			suggestion = policy.ApplySuggestion(suggestion)
		} else {
			suggestion.Status = review.StatusBlocked
			if record.Status == "stale_block" {
				suggestion.Status = review.StatusStale
			}
		}
		suggestions = append(suggestions, suggestion)
	}
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].SourcePath != suggestions[j].SourcePath {
			return suggestions[i].SourcePath < suggestions[j].SourcePath
		}
		if suggestions[i].Line != suggestions[j].Line {
			return suggestions[i].Line < suggestions[j].Line
		}
		return suggestions[i].ID < suggestions[j].ID
	})
	return suggestions, nil
}

func ApplySelectedSuggestion(plan *Plan, suggestion review.Suggestion, candidate review.Candidate) error {
	if suggestion.Kind != review.SuggestionLinkRepair {
		return fmt.Errorf("suggestion %s is not a link repair", suggestion.ID)
	}
	if candidate.Declined {
		return fmt.Errorf("candidate %d for %s is declined", candidate.Index, suggestion.ID)
	}
	linkIndex := -1
	for index := range plan.Links.Links {
		if plan.Links.Links[index].ID == suggestion.LinkID {
			linkIndex = index
			break
		}
	}
	if linkIndex < 0 {
		return fmt.Errorf("link for suggestion %s is no longer present", suggestion.ID)
	}
	record := &plan.Links.Links[linkIndex]
	if record.Status != "ambiguous" && record.Status != "stale_block" && record.Status != "blocked" {
		return fmt.Errorf("suggestion %s is no longer unresolved", suggestion.ID)
	}
	sourcePath := filepath.Join(plan.RepositoryRoot, filepath.FromSlash(record.SourcePath))
	targetPath := filepath.FromSlash(candidate.Target)
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(plan.RepositoryRoot, targetPath)
	}
	document, err := textio.Read(sourcePath)
	if err != nil {
		return err
	}
	_, style, local := resolveLocalTarget(record.RawPath, sourcePath, record.Angle)
	if !local {
		return fmt.Errorf("suggestion %s no longer points to a local target", suggestion.ID)
	}
	newPath := renderTargetForSyntax(record.Syntax, record.RawPath, style, sourcePath, targetPath)
	transformation := LinkTransformation{
		LinkID:         record.ID,
		Start:          record.Start,
		End:            record.End,
		OldDestination: record.RawPath,
		NewDestination: newPath,
	}
	rewrite, err := NewGeneratedRewrite(record.SourceFileID, sourcePath, document, []LinkTransformation{transformation})
	if err != nil {
		return err
	}
	rewrite.Selection = review.SelectionUser
	rewrite.OriginSuggestionID = suggestion.ID
	plan.Rewrites = append(plan.Rewrites, rewrite)
	old := document.Text
	updated := document.Text[:record.Start] + newPath + document.Text[record.End:]
	plan.Updates = append(plan.Updates, model.FileUpdate{Path: sourcePath, OldText: &old, NewText: updated})

	targetID := ""
	for _, file := range plan.Files.Files {
		if filepath.Clean(filepath.FromSlash(file.Path)) == filepath.Clean(filepath.FromSlash(candidate.Target)) {
			targetID = file.ID
			break
		}
	}
	record.TargetFileID = targetID
	record.ResolvedPath = candidate.Target
	record.RawPath = newPath
	record.Target = newPath + record.Suffix
	record.Status = "moved"
	record.Candidates = nil
	if plan.Unresolved > 0 {
		plan.Unresolved--
	}
	plan.Messages = append(plan.Messages, fmt.Sprintf("Selected link repair in %s:%d: %s -> %s", record.SourcePath, record.Line, transformation.OldDestination, newPath))
	return nil
}
