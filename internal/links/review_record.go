package links

import (
	"fmt"
	"sort"
	"time"

	"github.com/Lokee86/demon-docs/internal/review"
)

type reviewBatchAppender interface {
	AppendBatch([]review.AppendRequest) ([]review.StoredEvent, error)
}

var openReviewBatchStore = func(root string) (reviewBatchAppender, error) {
	return review.Open(root)
}

type generatedChangeBatch struct {
	store    reviewBatchAppender
	changes  []review.Change
	requests []review.AppendRequest
}

func prepareGeneratedChanges(plan *Plan) (generatedChangeBatch, error) {
	if len(plan.Rewrites) == 0 {
		return generatedChangeBatch{}, nil
	}
	store, err := openReviewBatchStore(plan.RepositoryRoot)
	if err != nil {
		return generatedChangeBatch{}, err
	}

	runID := review.NewID("run")
	linksByID := make(map[string]LinkRecord, len(plan.Links.Links))
	for _, record := range plan.Links.Links {
		linksByID[record.ID] = record
	}
	batch := generatedChangeBatch{
		store:    store,
		changes:  make([]review.Change, 0, len(plan.Rewrites)),
		requests: make([]review.AppendRequest, 0, len(plan.Rewrites)),
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
		batch.changes = append(batch.changes, change)
	}
	for index := range batch.changes {
		change := &batch.changes[index]
		rewrite := plan.Rewrites[index]
		batch.requests = append(batch.requests, review.AppendRequest{
			Event:  review.Event{Type: review.EventChange, Time: change.AppliedAt, Change: change},
			Before: rewrite.OldData(),
			After:  rewrite.NewData(),
		})
	}
	return batch, nil
}

func recordGeneratedChanges(plan *Plan, batch generatedChangeBatch) error {
	if len(batch.requests) == 0 {
		return nil
	}
	if _, err := batch.store.AppendBatch(batch.requests); err != nil {
		return fmt.Errorf("record applied change batch: %w", err)
	}
	plan.AppliedChanges = append(plan.AppliedChanges, batch.changes...)
	return nil
}
