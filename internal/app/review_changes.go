package app

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/links"
	"github.com/Lokee86/demon-docs/internal/review"
	"github.com/pmezard/go-difflib/difflib"
)

func changesHelp(out io.Writer) {
	fmt.Fprintln(out, "usage: ddocs changes [-h] [FILE]\n       ddocs changes {related,show,log,undo,undo-run,block,unblock} ...\n\nInspect, undo, and block Demon Docs repairs. Undo is available by reconciliation run, file change, or individual repair while the recorded after-state still matches.\n\ncommands:\n  related FILE                         show repairs caused by or targeting FILE\n  show CHANGE                          show one applied change and its diff\n  log [FILE]                           show applied-change and control history\n  undo CHANGE [--repair REPAIR] [--block] [--reason TEXT]\n  undo-run RUN [--block] [--reason TEXT]\n  block CHANGE [--repair REPAIR] [--reason TEXT]\n  unblock CHANGE [--repair REPAIR]")
}

func runChanges(args []string, out, errOut io.Writer) int {
	if helpRequested(args) {
		changesHelp(out)
		return 0
	}
	if len(args) == 0 || !isChangesCommand(args[0]) {
		return runChangesList(args, false, out, errOut)
	}
	switch args[0] {
	case "related":
		return runChangesList(args[1:], true, out, errOut)
	case "show":
		return runChangesShow(args[1:], out, errOut)
	case "log":
		return runChangesLog(args[1:], out, errOut)
	case "undo":
		return runChangesUndo(args[1:], out, errOut)
	case "undo-run":
		return runChangesUndoRun(args[1:], out, errOut)
	case "block":
		return runChangesBlock(args[1:], false, out, errOut)
	case "unblock":
		return runChangesBlock(args[1:], true, out, errOut)
	default:
		return 2
	}
}

func isChangesCommand(value string) bool {
	switch value {
	case "related", "show", "log", "undo", "undo-run", "block", "unblock":
		return true
	default:
		return false
	}
}

type changeState struct {
	undone  bool
	blocked bool
	stale   bool
}

func runChangesList(args []string, related bool, out, errOut io.Writer) int {
	if len(args) > 1 || related && len(args) != 1 {
		if related {
			fmt.Fprintln(errOut, "usage: ddocs changes related FILE")
		} else {
			writeUnrecognized(errOut, args[1:])
		}
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
	plan, err := links.Reconcile(scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	filter, err := resolveTrackedReviewFilter(scope.RepositoryRoot, plan, args)
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
	policy, err := review.LoadPolicy(scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	states := changeStates(history, policy)
	count := 0
	for _, event := range history {
		change := event.Change
		if change == nil {
			continue
		}
		if related {
			if !changeRelated(*change, filter) {
				continue
			}
		} else if !sourceMatches(*change, filter) {
			continue
		}
		writeChangeLine(out, *change, states[change.ID])
		count++
	}
	if count == 0 {
		fmt.Fprintln(out, "no applied changes")
	}
	return 0
}

func runChangesShow(args []string, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: ddocs changes show CHANGE")
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
	event, err := store.Find(args[0])
	if err != nil || event.Change == nil {
		return fail(errOut, fmt.Errorf("change not found: %s", args[0]))
	}
	change := *event.Change
	fmt.Fprintf(out, "%s  %s  %s\n", change.ID, change.Kind, change.SourcePath)
	fmt.Fprintf(out, "run: %s\nselection: %s\napplied: %s\nbefore: %s\nafter: %s\n", change.RunID, change.Selection, eventTime(change.AppliedAt), change.BeforeSHA256, change.AfterSHA256)
	if change.OriginSuggestionID != "" {
		fmt.Fprintf(out, "origin suggestion: %s\n", change.OriginSuggestionID)
	}
	if change.UndoOf != "" {
		fmt.Fprintf(out, "undo of: %s\n", change.UndoOf)
	}
	for _, repair := range change.Transformations {
		fmt.Fprintf(out, "repair %s: %q -> %q\n", repair.ID, repair.OldText, repair.NewText)
	}
	if event.Before != nil && event.After != nil {
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(event.Before)),
			B:        difflib.SplitLines(string(event.After)),
			FromFile: change.SourcePath + " before",
			ToFile:   change.SourcePath + " after",
			Context:  3,
		})
		if err == nil && diff != "" {
			fmt.Fprintln(out, diff)
		}
	}
	return 0
}

func runChangesLog(args []string, out, errOut io.Writer) int {
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
	plan, err := links.Reconcile(scope.RepositoryRoot)
	if err != nil {
		return fail(errOut, err)
	}
	filter, err := resolveTrackedReviewFilter(scope.RepositoryRoot, plan, args)
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
	changesByID := make(map[string]review.Change)
	for _, event := range history {
		if event.Change != nil {
			changesByID[event.Change.ID] = *event.Change
		}
	}
	count := 0
	for _, event := range history {
		if event.Change != nil && sourceMatches(*event.Change, filter) {
			fmt.Fprintf(out, "%s  %s  change  %s  %s\n", event.CommitHash[:12], eventTime(event.Time), event.Change.ID, event.Change.SourcePath)
			count++
			continue
		}
		if event.Decision != nil && event.Decision.ChangeID != "" {
			change, ok := changesByID[event.Decision.ChangeID]
			if !ok || !sourceMatches(change, filter) {
				continue
			}
			fmt.Fprintf(out, "%s  %s  %s  %s\n", event.CommitHash[:12], eventTime(event.Time), event.Decision.Action, event.Decision.ChangeID)
			count++
		}
	}
	if count == 0 {
		fmt.Fprintln(out, "no change history")
	}
	return 0
}

func changeStates(history []review.StoredEvent, policy review.Policy) map[string]changeState {
	states := make(map[string]changeState)
	for _, event := range history {
		if event.Change != nil && event.Change.UndoOf != "" {
			state := states[event.Change.UndoOf]
			state.undone = true
			states[event.Change.UndoOf] = state
		}
	}
	for _, event := range history {
		if event.Change == nil {
			continue
		}
		state := states[event.Change.ID]
		for _, repair := range event.Change.Transformations {
			match, _ := policy.Repair(repair.RelationKey, repair.Fingerprint)
			state.blocked = state.blocked || match == review.MatchActive
			state.stale = state.stale || match == review.MatchStale
		}
		states[event.Change.ID] = state
	}
	return states
}

func changeRelated(change review.Change, filter trackedReviewFilter) bool {
	if filter.Path == "" {
		return true
	}
	for _, related := range change.Related {
		if filter.FileID != "" && related.FileID == filter.FileID || filepath.ToSlash(filepath.Clean(related.Path)) == filter.Path {
			return true
		}
	}
	return false
}

func writeChangeLine(out io.Writer, change review.Change, state changeState) {
	status := "applied"
	if change.UndoOf != "" {
		status = "undo"
	} else if state.undone {
		status = "undone"
	}
	var controls []string
	if state.blocked {
		controls = append(controls, "blocked")
	}
	if state.stale {
		controls = append(controls, "stale block")
	}
	if len(controls) > 0 {
		status += ", " + strings.Join(controls, ", ")
	}
	fmt.Fprintf(out, "%s  %s  %s  %s  run=%s\n", change.ID, strings.ToUpper(status), change.Kind, change.SourcePath, change.RunID)
	if len(change.Related) > 0 {
		paths := make([]string, 0, len(change.Related))
		for _, item := range change.Related {
			paths = append(paths, item.Path)
		}
		sort.Strings(paths)
		fmt.Fprintf(out, "  related: %s\n", strings.Join(paths, ", "))
	}
}
