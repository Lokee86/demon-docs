package review

import (
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

func TestStoreAppendsGitBackedHistoryWithUndoBlobs(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	store, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	change := Change{ID: "ch-test", RunID: "run-test", SourceFileID: "file-1", SourcePath: "README.md", BeforeSHA256: Digest([]byte("old")), AfterSHA256: Digest([]byte("new")), AppliedAt: now}
	stored, err := store.Append(Event{ID: "ev-test", Type: EventChange, Time: now, Change: &change}, []byte("old"), []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	if stored.CommitHash == "" {
		t.Fatal("missing review commit hash")
	}
	history, err := store.History(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Change == nil || history[0].Change.ID != "ch-test" {
		t.Fatalf("unexpected history: %#v", history)
	}
	if string(history[0].Before) != "old" || string(history[0].After) != "new" {
		t.Fatalf("undo blobs were not retained: before=%q after=%q", history[0].Before, history[0].After)
	}
}

func TestPolicyKeepsDeclineUntilFingerprintChanges(t *testing.T) {
	root := t.TempDir()
	if _, err := ddrepo.Init(root); err != nil {
		t.Fatal(err)
	}
	suggestion := LinkSuggestion("file-1", "README.md", "link-1", "old.md", []string{"one.md", "two.md"})
	if _, err := appendTestDecision(root, Decision{
		Action:       DecisionDeclineIssue,
		RelationKey:  suggestion.RelationKey,
		Fingerprint:  suggestion.Fingerprint,
		SuggestionID: suggestion.ID,
		Suggestion:   &suggestion,
	}); err != nil {
		t.Fatal(err)
	}
	policy, err := LoadPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := policy.ApplySuggestion(suggestion).Status; got != StatusDeclined {
		t.Fatalf("status = %q, want declined", got)
	}
	changed := LinkSuggestion("file-1", "README.md", "link-1", "old.md", []string{"one.md", "three.md"})
	if got := policy.ApplySuggestion(changed).Status; got != StatusStale {
		t.Fatalf("changed evidence status = %q, want stale", got)
	}
	if _, err := appendTestDecision(root, Decision{Action: DecisionReconsider, RelationKey: suggestion.RelationKey, SuggestionID: suggestion.ID, Suggestion: &suggestion}); err != nil {
		t.Fatal(err)
	}
	policy, err = LoadPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := policy.ApplySuggestion(suggestion).Status; got != StatusPending {
		t.Fatalf("reconsidered status = %q, want pending", got)
	}
}

func appendTestDecision(root string, decision Decision) (StoredEvent, error) {
	store, err := Open(root)
	if err != nil {
		return StoredEvent{}, err
	}
	decision.ID = NewID("dc")
	decision.DecidedAt = time.Now().UTC()
	return store.Append(Event{Type: EventDecision, Time: decision.DecidedAt, Decision: &decision}, nil, nil)
}
