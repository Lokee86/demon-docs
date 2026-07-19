package links

import (
	"fmt"
	"sort"
	"time"

	"github.com/Lokee86/demon-docs/internal/review"
)

func recordGeneratedChanges(plan *Plan) error {
	if len(plan.Rewrites) == 0 {
		return nil
	}
	store, err := review.Open(plan.RepositoryRoot)
	if err != nil {
		return err
	}
	runID := review.NewID("run")
	linksByID := make(map[string]LinkRecord, len(plan.Links.Links))
	for _, record := range plan.Links.Links {
		linksByID[record.ID] = record
	}
	for _, rewrite := range plan.Rewrites {
		kind := rewrite.Kind
		if kind == "" {
			kind = review.SuggestionLinkRepair
		}
		selection := rewrite.Selection
		if selection == "" {
			selection = review.SelectionAutomatic
		}
		change := review.Change{
			ID:                 review.NewID("ch"),
			RunID:              runID,
			Kind:               kind,
			Selection:          selection,
			OriginSuggestionID: rewrite.OriginSuggestionID,
			SourceFileID:       rewrite.SourceFileID,
			SourcePath:         storePath(plan.RepositoryRoot, rewrite.Path),
			BeforeSHA256:       rewrite.ExpectedOldSHA256,
			AfterSHA256:        rewrite.ExpectedNewSHA256,
			AppliedAt:          time.Now().UTC(),
		}
		related := make(map[string]review.RelatedFile)
		for _, item := range rewrite.Transformations {
			record := linksByID[item.LinkID]
			targetFileID := record.TargetFileID
			targetPath := record.ResolvedPath
			if item.TargetFileID != "" {
				targetFileID = item.TargetFileID
			}
			if item.TargetPath != "" {
				targetPath = item.TargetPath
			}
			relationToken := item.LinkID
			if record.ID != "" {
				relationToken = reviewRelationToken(record)
			}
			relation, fingerprint := review.RepairIdentity(rewrite.SourceFileID, relationToken, item.OldDestination, item.NewDestination, targetFileID)
			transformation := review.Transformation{
				ID:            review.TransformationID(relation, fingerprint),
				LinkID:        item.LinkID,
				Start:         item.Start,
				End:           item.End,
				OldText:       item.OldDestination,
				NewText:       item.NewDestination,
				RelationKey:   relation,
				RelationToken: relationToken,
				Fingerprint:   fingerprint,
				TargetFileID:  targetFileID,
				TargetPath:    targetPath,
			}
			change.Transformations = append(change.Transformations, transformation)
			if targetPath != "" {
				related[targetFileID+"\x00"+targetPath] = review.RelatedFile{FileID: targetFileID, Path: targetPath}
			}
		}
		for _, item := range related {
			change.Related = append(change.Related, item)
		}
		sort.Slice(change.Related, func(i, j int) bool { return change.Related[i].Path < change.Related[j].Path })
		event := review.Event{Type: review.EventChange, Time: change.AppliedAt, Change: &change}
		if _, err := store.Append(event, rewrite.OldData(), rewrite.NewData()); err != nil {
			return fmt.Errorf("record applied change %s: %w", change.ID, err)
		}
		plan.AppliedChanges = append(plan.AppliedChanges, change)
	}
	return nil
}
