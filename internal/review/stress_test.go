package review

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestReviewLedgerDeepHistoryPreservesFindAndUndoDepth(t *testing.T) {
	const (
		batches  = 64
		perBatch = 8
		total    = batches * perBatch
	)
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	for batch := 0; batch < batches; batch++ {
		requests := make([]AppendRequest, perBatch)
		for offset := 0; offset < perBatch; offset++ {
			index := batch*perBatch + offset
			id := fmt.Sprintf("ch-%03d", index)
			change := Change{
				ID:           id,
				RunID:        fmt.Sprintf("run-%02d", batch),
				SourceFileID: fmt.Sprintf("file-%03d", index),
				SourcePath:   fmt.Sprintf("docs/file-%03d.md", index),
				AppliedAt:    base.Add(time.Duration(index) * time.Second),
			}
			requests[offset] = AppendRequest{
				Event:  Event{ID: "ev-" + id, Type: EventChange, Time: change.AppliedAt, Change: &change},
				Before: []byte("before-" + id),
				After:  []byte("after-" + id),
			}
		}
		if _, err := store.AppendBatch(requests); err != nil {
			t.Fatalf("append batch %d: %v", batch, err)
		}
	}

	history, err := store.History(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != total {
		t.Fatalf("history length=%d want=%d", len(history), total)
	}
	if history[0].Change == nil || history[0].Change.ID != "ch-511" || history[total-1].Change == nil || history[total-1].Change.ID != "ch-000" {
		t.Fatalf("unexpected deep history endpoints: newest=%#v oldest=%#v", history[0].Change, history[total-1].Change)
	}
	oldest, err := store.Find("ch-000")
	if err != nil {
		t.Fatal(err)
	}
	if string(oldest.Before) != "before-ch-000" || string(oldest.After) != "after-ch-000" {
		t.Fatalf("oldest snapshots changed: before=%q after=%q", oldest.Before, oldest.After)
	}
	now := base.Add(total * time.Second)
	if err := UndoEligible(history, "ch-000", total, 0, now); err != nil {
		t.Fatalf("oldest change should be eligible at exact depth: %v", err)
	}
	if err := UndoEligible(history, "ch-000", total-1, 0, now); err == nil || !strings.Contains(err.Error(), "outside the configured undo depth") {
		t.Fatalf("oldest change unexpectedly eligible below depth: %v", err)
	}
}

func TestPolicyReplayStressDeclineStaleAndReconsider(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	targets := make([]string, 64)
	for index := range targets {
		targets[index] = fmt.Sprintf("src/target-%02d.go", index)
	}
	original := LinkSuggestion("file-main", "docs/main.md", "link-main", "missing.go", targets)
	requests := []AppendRequest{{Event: decisionEvent(Decision{
		ID:           "dc-main-old",
		Action:       DecisionDeclineIssue,
		RelationKey:  original.RelationKey,
		Fingerprint:  original.Fingerprint,
		SuggestionID: original.ID,
		Suggestion:   &original,
	})}}
	for _, candidate := range original.Candidates {
		requests = append(requests, AppendRequest{Event: decisionEvent(Decision{
			ID:                   "dc-candidate-" + fmt.Sprint(candidate.Index),
			Action:               DecisionDeclineCandidate,
			RelationKey:          original.RelationKey,
			CandidateTarget:      candidate.Target,
			CandidateFingerprint: candidate.Fingerprint,
			SuggestionID:         original.ID,
		})})
	}
	for index := 0; index < 256; index++ {
		requests = append(requests, AppendRequest{Event: decisionEvent(Decision{
			ID:          fmt.Sprintf("dc-noise-%03d", index),
			Action:      DecisionDeclineIssue,
			RelationKey: fmt.Sprintf("unrelated-%03d", index),
			Fingerprint: fmt.Sprintf("fingerprint-%03d", index),
		})})
	}
	if _, err := store.AppendBatch(requests); err != nil {
		t.Fatal(err)
	}

	policy, err := LoadPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	applied := policy.ApplySuggestion(original)
	if applied.Status != StatusDeclined {
		t.Fatalf("original status=%q want declined", applied.Status)
	}
	for _, candidate := range applied.Candidates {
		if !candidate.Declined || candidate.Stale {
			t.Fatalf("candidate decline lost at index %d: %#v", candidate.Index, candidate)
		}
	}

	changedTargets := append(append([]string(nil), targets...), "src/new-evidence.go")
	changed := LinkSuggestion("file-main", "docs/main.md", "link-main", "missing.go", changedTargets)
	stale := policy.ApplySuggestion(changed)
	if stale.Status != StatusStale {
		t.Fatalf("changed evidence status=%q want stale", stale.Status)
	}
	originalTargets := make(map[string]bool, len(targets))
	for _, target := range targets {
		originalTargets[target] = true
	}
	for _, candidate := range stale.Candidates {
		if originalTargets[candidate.Target] {
			if candidate.Declined || !candidate.Stale {
				t.Fatalf("candidate evidence was not marked stale: %#v", candidate)
			}
			continue
		}
		if candidate.Declined || candidate.Stale {
			t.Fatalf("new candidate inherited decline state: %#v", candidate)
		}
	}
	for _, candidate := range changed.Candidates {
		if candidate.Declined || candidate.Stale {
			t.Fatalf("policy projection mutated reusable suggestion evidence: %#v", candidate)
		}
	}

	if _, err := store.Append(decisionEvent(Decision{ID: "dc-reconsider", Action: DecisionReconsider, RelationKey: original.RelationKey}), nil, nil); err != nil {
		t.Fatal(err)
	}
	policy, err = LoadPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	pending := policy.ApplySuggestion(changed)
	if pending.Status != StatusPending {
		t.Fatalf("reconsidered status=%q want pending", pending.Status)
	}
	for _, candidate := range pending.Candidates {
		if candidate.Declined || candidate.Stale {
			t.Fatalf("reconsider did not clear candidate state: %#v", candidate)
		}
	}

	if _, err := store.Append(decisionEvent(Decision{ID: "dc-main-new", Action: DecisionDeclineIssue, RelationKey: changed.RelationKey, Fingerprint: changed.Fingerprint}), nil, nil); err != nil {
		t.Fatal(err)
	}
	policy, err = LoadPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := policy.ApplySuggestion(changed).Status; got != StatusDeclined {
		t.Fatalf("new evidence decline status=%q want declined", got)
	}
	if got := policy.ApplySuggestion(original).Status; got != StatusStale {
		t.Fatalf("old evidence status=%q want stale", got)
	}
}

func decisionEvent(decision Decision) Event {
	return Event{ID: "ev-" + decision.ID, Type: EventDecision, Decision: &decision}
}
