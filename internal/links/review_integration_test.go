package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/demon-docs/internal/review"
)

func TestSelectedAmbiguousSuggestionBecomesRecordedRepair(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[manual](old/manual.pdf)\n")
	writeTestFile(t, filepath.Join(root, "one", "manual.pdf"), "one")
	writeTestFile(t, filepath.Join(root, "two", "manual.pdf"), "two")
	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	suggestions, err := ReviewSuggestions(plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 || len(suggestions[0].Candidates) != 2 {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
	if err := ApplySelectedSuggestion(&plan, suggestions[0], suggestions[0].Candidates[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&plan); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "one/manual.pdf") {
		t.Fatalf("selected repair was not applied: %q", data)
	}
	if len(plan.AppliedChanges) != 1 || plan.AppliedChanges[0].Selection != review.SelectionUser || plan.AppliedChanges[0].OriginSuggestionID != suggestions[0].ID {
		t.Fatalf("selected repair was not recorded: %#v", plan.AppliedChanges)
	}
	store, err := review.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	history, err := store.History(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Change == nil || history[0].Change.ID != plan.AppliedChanges[0].ID {
		t.Fatalf("unexpected review history: %#v", history)
	}
}

func TestSelectedExternalCandidatePreservesAbsolutePath(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	oldTarget := filepath.Join(external, "old", "manual.pdf")
	candidateTarget := filepath.Join(external, "new", "manual.pdf")
	oldTargetText := filepath.ToSlash(oldTarget)
	candidateTargetText := filepath.ToSlash(candidateTarget)
	sourceText := "[manual](<" + oldTargetText + ">)\n"
	sourcePath := filepath.Join(root, "README.md")
	writeTestFile(t, sourcePath, sourceText)
	writeTestFile(t, candidateTarget, "manual")
	start := strings.Index(sourceText, oldTargetText)

	plan := Plan{
		RepositoryRoot: root,
		Unresolved:     1,
		Files: FilesManifest{Files: []FileRecord{
			{ID: "source", Path: "README.md", Scope: "repository", Kind: "file", Present: true},
			{ID: "target", Path: candidateTargetText, Scope: "external", Kind: "file", Present: true},
		}},
		Links: LinksManifest{Links: []LinkRecord{{
			ID:           "link",
			SourceFileID: "source",
			SourcePath:   "README.md",
			Start:        start,
			End:          start + len(oldTargetText),
			Line:         1,
			Column:       start + 1,
			Syntax:       "inline",
			RawPath:      oldTargetText,
			Angle:        true,
			Target:       oldTargetText,
			Status:       "ambiguous",
			Candidates:   []string{candidateTargetText},
		}}},
	}
	suggestion := review.LinkSuggestion("source", "README.md", "link", oldTargetText, []string{candidateTargetText})
	candidate := suggestion.Candidates[0]

	if err := ApplySelectedSuggestion(&plan, suggestion, candidate); err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 {
		t.Fatalf("updates = %d, want 1", len(plan.Updates))
	}
	if !strings.Contains(plan.Updates[0].NewText, "<"+candidateTargetText+">") {
		t.Fatalf("external candidate path was not preserved: %q", plan.Updates[0].NewText)
	}
}

func TestBlockedDeterministicRepairIsNotReapplied(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "[manual](old/manual.pdf)\n")
	writeTestFile(t, filepath.Join(root, "references", "manual.pdf"), "manual")
	first, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(first); err != nil {
		t.Fatal(err)
	}
	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) != 1 || len(plan.Rewrites[0].Transformations) != 1 {
		t.Fatalf("expected deterministic repair: %#v", plan)
	}
	transformation := plan.Rewrites[0].Transformations[0]
	record := plan.Links.Links[0]
	relation, fingerprint := review.RepairIdentity(record.SourceFileID, reviewRelationToken(record), transformation.OldDestination, transformation.NewDestination, record.TargetFileID)
	store, err := review.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	decision := review.Decision{ID: "dc-block", Action: review.DecisionBlockRepair, RelationKey: relation, Fingerprint: fingerprint, Reason: "intentional old path", DecidedAt: now}
	if _, err := store.Append(review.Event{Type: review.EventDecision, Time: now, Decision: &decision}, nil, nil); err != nil {
		t.Fatal(err)
	}
	blocked, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocked.Rewrites) != 0 || blocked.Unresolved != 1 || blocked.Links.Links[0].Status != "blocked" {
		t.Fatalf("blocked repair was still planned: %#v", blocked)
	}
}
