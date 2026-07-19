package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/review"
)

type changeOptions struct {
	id       string
	repairID string
	block    bool
	reason   string
}

func parseChangeOptions(args []string, allowRepair, allowBlock bool) (changeOptions, error) {
	var options changeOptions
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--repair":
			if !allowRepair || index+1 >= len(args) {
				return options, fmt.Errorf("invalid --repair option")
			}
			options.repairID = args[index+1]
			index++
		case "--block":
			if !allowBlock {
				return options, fmt.Errorf("invalid --block option")
			}
			options.block = true
		case "--reason":
			if index+1 >= len(args) {
				return options, fmt.Errorf("--reason requires a value")
			}
			options.reason = strings.TrimSpace(args[index+1])
			index++
		default:
			if options.id != "" {
				return options, fmt.Errorf("unexpected argument: %s", args[index])
			}
			options.id = args[index]
		}
	}
	if options.id == "" {
		return options, fmt.Errorf("change or run ID is required")
	}
	return options, nil
}

func runChangesUndo(args []string, out, errOut io.Writer) int {
	options, err := parseChangeOptions(args, true, true)
	if err != nil {
		fmt.Fprintln(errOut, "usage: ddocs changes undo CHANGE [--repair REPAIR] [--block] [--reason TEXT]")
		return 2
	}
	runtime, code := loadReviewRuntime(context.Background(), errOut)
	if code != 0 {
		return code
	}
	store, err := review.Open(runtime.scope.RepositoryRoot)
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
	if err := review.UndoEligible(history, event.Change.ID, runtime.config.Review.UndoDepth, runtime.config.Review.UndoMaxAgeDays, time.Now()); err != nil {
		return fail(errOut, err)
	}
	path := currentFilePath(runtime.scope.RepositoryRoot, runtime.linkPlan, *event.Change)
	current, err := readCurrent(path)
	if err != nil {
		return fail(errOut, err)
	}
	if review.Digest(current) != event.Change.AfterSHA256 {
		return fail(errOut, fmt.Errorf("cannot undo %s: %s changed after the recorded repair", event.Change.ID, filepath.ToSlash(path)))
	}
	updated, err := review.BuildUndoData(*event.Change, event.Before, event.After, options.repairID)
	if err != nil {
		return fail(errOut, err)
	}
	rewrite, err := links.NewGeneratedRewriteBytes(event.Change.SourceFileID, path, current, updated, nil)
	if err != nil {
		return fail(errOut, err)
	}
	if _, err := links.ApplyGenerated([]links.GeneratedRewrite{rewrite}); err != nil {
		return fail(errOut, err)
	}
	_ = links.DeletePendingSuppression(runtime.scope.RepositoryRoot, event.Change.SourceFileID)
	undo, err := recordUndo(store, *event.Change, current, updated, options.repairID)
	if err != nil {
		return fail(errOut, err)
	}
	if options.block {
		sourcePath, err := reviewPath(runtime.scope.RepositoryRoot, path)
		if err != nil {
			return fail(errOut, err)
		}
		if err := appendRepairControls(runtime.scope.RepositoryRoot, *event.Change, options.repairID, false, options.reason, sourcePath); err != nil {
			return fail(errOut, err)
		}
	}
	fmt.Fprintf(out, "undid %s with %s\n", event.Change.ID, undo.ID)
	return 0
}

func runChangesUndoRun(args []string, out, errOut io.Writer) int {
	options, err := parseChangeOptions(args, false, true)
	if err != nil {
		fmt.Fprintln(errOut, "usage: ddocs changes undo-run RUN [--block] [--reason TEXT]")
		return 2
	}
	runtime, code := loadReviewRuntime(context.Background(), errOut)
	if code != 0 {
		return code
	}
	store, err := review.Open(runtime.scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	history, err := store.History(0)
	if err != nil {
		return fail(errOut, err)
	}
	var events []review.StoredEvent
	for _, event := range history {
		if event.Change != nil && event.Change.RunID == options.id && event.Change.UndoOf == "" {
			events = append(events, event)
		}
	}
	if len(events) == 0 {
		return fail(errOut, fmt.Errorf("run not found: %s", options.id))
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Change.SourcePath < events[j].Change.SourcePath })
	seenPaths := make(map[string]bool)
	var rewrites []links.GeneratedRewrite
	type pendingUndo struct {
		event      review.StoredEvent
		current    []byte
		updated    []byte
		sourcePath string
	}
	var pending []pendingUndo
	for _, event := range events {
		if err := review.UndoEligible(history, event.Change.ID, runtime.config.Review.UndoDepth, runtime.config.Review.UndoMaxAgeDays, time.Now()); err != nil {
			return fail(errOut, err)
		}
		path := currentFilePath(runtime.scope.RepositoryRoot, runtime.linkPlan, *event.Change)
		key := strings.ToLower(filepath.Clean(path))
		if seenPaths[key] {
			return fail(errOut, fmt.Errorf("run %s contains multiple changes for %s", options.id, path))
		}
		seenPaths[key] = true
		current, err := readCurrent(path)
		if err != nil {
			return fail(errOut, err)
		}
		if review.Digest(current) != event.Change.AfterSHA256 {
			return fail(errOut, fmt.Errorf("cannot undo run %s: %s changed after repair", options.id, filepath.ToSlash(path)))
		}
		updated, err := review.BuildUndoData(*event.Change, event.Before, event.After, "")
		if err != nil {
			return fail(errOut, err)
		}
		rewrite, err := links.NewGeneratedRewriteBytes(event.Change.SourceFileID, path, current, updated, nil)
		if err != nil {
			return fail(errOut, err)
		}
		rewrites = append(rewrites, rewrite)
		sourcePath, err := reviewPath(runtime.scope.RepositoryRoot, path)
		if err != nil {
			return fail(errOut, err)
		}
		pending = append(pending, pendingUndo{event: event, current: current, updated: updated, sourcePath: sourcePath})
	}
	if _, err := links.ApplyGenerated(rewrites); err != nil {
		return fail(errOut, err)
	}
	for _, item := range pending {
		_ = links.DeletePendingSuppression(runtime.scope.RepositoryRoot, item.event.Change.SourceFileID)
		if _, err := recordUndo(store, *item.event.Change, item.current, item.updated, ""); err != nil {
			return fail(errOut, err)
		}
		if options.block {
			if err := appendRepairControls(runtime.scope.RepositoryRoot, *item.event.Change, "", false, options.reason, item.sourcePath); err != nil {
				return fail(errOut, err)
			}
		}
	}
	fmt.Fprintf(out, "undid run %s (%d change(s))\n", options.id, len(pending))
	return 0
}
