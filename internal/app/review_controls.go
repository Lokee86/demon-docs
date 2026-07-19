package app

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/review"
)

func runChangesBlock(args []string, unblock bool, out, errOut io.Writer) int {
	options, err := parseChangeOptions(args, true, false)
	if err != nil {
		verb := "block"
		if unblock {
			verb = "unblock"
		}
		fmt.Fprintf(errOut, "usage: ddocs changes %s CHANGE [--repair REPAIR] [--reason TEXT]\n", verb)
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
	store, err := review.Open(scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	history, err := store.History(0)
	if err != nil {
		return fail(errOut, err)
	}
	event, err := changeEvent(history, options.id)
	if err != nil {
		return fail(errOut, err)
	}
	plan, err := links.Reconcile(scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	path := currentFilePath(scope.RepositoryRoot, plan, *event.Change)
	sourcePath, err := reviewPath(scope.RepositoryRoot, path)
	if err != nil {
		return fail(errOut, err)
	}
	if err := appendRepairControls(scope.RepositoryRoot, *event.Change, options.repairID, unblock, options.reason, sourcePath); err != nil {
		return fail(errOut, err)
	}
	verb := "blocked"
	if unblock {
		verb = "unblocked"
	}
	fmt.Fprintf(out, "%s %s\n", verb, event.Change.ID)
	return 0
}

func buildUndoRequest(original review.Change, before, after []byte, repairID string) (review.Change, review.AppendRequest) {
	now := time.Now().UTC()
	undo := review.Change{
		ID:           review.NewID("ch"),
		RunID:        review.NewID("run"),
		Kind:         original.Kind,
		Selection:    review.SelectionUndo,
		SourceFileID: original.SourceFileID,
		SourcePath:   original.SourcePath,
		BeforeSHA256: review.Digest(before),
		AfterSHA256:  review.Digest(after),
		Related:      append([]review.RelatedFile(nil), original.Related...),
		UndoOf:       original.ID,
		UndoRepairID: repairID,
		AppliedAt:    now,
	}
	request := review.AppendRequest{
		Event:  review.Event{Type: review.EventChange, Time: now, Change: &undo},
		Before: before,
		After:  after,
	}
	return undo, request
}

func appendRepairControls(root string, change review.Change, repairID string, unblock bool, reason, currentSourcePath string) error {
	action := review.DecisionBlockRepair
	if unblock {
		action = review.DecisionUnblockRepair
	}
	matched := false
	for _, repair := range change.Transformations {
		if repairID != "" && repair.ID != repairID {
			continue
		}
		matched = true
		relations := map[string]string{repair.RelationKey: repair.Fingerprint}
		if repair.RelationToken != "" {
			for _, sourcePath := range []string{change.SourcePath, currentSourcePath} {
				if sourcePath == "" {
					continue
				}
				relation, fingerprint := review.RepairIdentity(review.PathIdentity(sourcePath), repair.RelationToken, repair.OldText, repair.NewText, repair.TargetFileID)
				relations[relation] = fingerprint
			}
		}
		ordered := make([]string, 0, len(relations))
		for relation := range relations {
			ordered = append(ordered, relation)
		}
		sort.Strings(ordered)
		for _, relation := range ordered {
			_, err := appendDecision(root, review.Decision{
				Action:      action,
				RelationKey: relation,
				Fingerprint: relations[relation],
				ChangeID:    change.ID,
				Reason:      reason,
			})
			if err != nil {
				return err
			}
		}
	}
	if !matched {
		return fmt.Errorf("repair not found in change %s: %s", change.ID, repairID)
	}
	return nil
}

func changeEvent(history []review.StoredEvent, id string) (review.StoredEvent, error) {
	for _, event := range history {
		if event.Change != nil && event.Change.ID == id {
			return event, nil
		}
	}
	return review.StoredEvent{}, fmt.Errorf("change not found: %s", id)
}
