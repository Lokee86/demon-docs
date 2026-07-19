package app

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/Lokee86/demon-docs/internal/textio"
)

func applyCodemapSelection(runtime *reviewRuntime, suggestion review.Suggestion, candidate review.Candidate) error {
	sourcePath := filepath.Join(runtime.scope.RepositoryRoot, filepath.FromSlash(suggestion.SourcePath))
	document, err := textio.Read(sourcePath)
	if err != nil {
		return err
	}
	updated, start, end, inserted, err := codemap.InsertTarget(document.Text, runtime.config.Codemap.Headings, candidate.Target)
	if err != nil {
		return err
	}
	sourceFileID := ""
	for _, file := range runtime.linkPlan.Files.Files {
		if file.Present && filepath.ToSlash(filepath.Clean(file.Path)) == filepath.ToSlash(filepath.Clean(suggestion.SourcePath)) {
			sourceFileID = file.ID
			break
		}
	}
	if sourceFileID == "" {
		return fmt.Errorf("tracked source file not found for %s", suggestion.SourcePath)
	}
	transformation := links.LinkTransformation{
		LinkID:         suggestion.ID,
		Start:          start,
		End:            end,
		OldDestination: "",
		NewDestination: inserted,
		TargetPath:     candidate.Target,
	}
	rewrite, err := links.NewGeneratedRewriteBytes(
		sourceFileID,
		sourcePath,
		document.Encode(document.Text),
		document.Encode(updated),
		[]links.LinkTransformation{transformation},
	)
	if err != nil {
		return err
	}
	rewrite.Kind = review.SuggestionCodemap
	rewrite.Selection = review.SelectionUser
	rewrite.OriginSuggestionID = suggestion.ID
	runtime.linkPlan.Rewrites = append(runtime.linkPlan.Rewrites, rewrite)
	return nil
}
