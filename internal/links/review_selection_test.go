package links

import (
	"testing"

	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/review"
)

func TestPrepareSelectionPlanRemovesAutomaticWritesAndRestoresRecords(t *testing.T) {
	plan := Plan{
		Updates:        []model.FileUpdate{{Path: "source.md"}},
		Suppressions:   []Suppression{{Path: "source.md"}},
		AppliedChanges: []review.Change{{ID: "ch-test"}},
		Unresolved:     1,
		Rewrites: []GeneratedRewrite{{
			Transformations: []LinkTransformation{
				{LinkID: "moved", OldDestination: "old/guide.md", NewDestination: "new/guide.md"},
				{LinkID: "case", OldDestination: "README.md", NewDestination: "readme.md"},
			},
		}},
		Links: LinksManifest{Links: []LinkRecord{
			{ID: "moved", RawPath: "new/guide.md", Suffix: "#topic", Target: "new/guide.md#topic", ResolvedPath: "new/guide.md", Status: "moved"},
			{ID: "case", RawPath: "readme.md", Target: "readme.md", ResolvedPath: "readme.md", Status: "case_mismatch"},
			{ID: "ambiguous", RawPath: "guide.md", Target: "guide.md", Status: "ambiguous"},
		}},
	}

	PrepareSelectionPlan(&plan)

	if plan.Updates != nil || plan.Rewrites != nil || plan.Suppressions != nil || plan.AppliedChanges != nil {
		t.Fatalf("planned writes were not cleared: %#v", plan)
	}
	if plan.Unresolved != 3 {
		t.Fatalf("unresolved=%d, want 3", plan.Unresolved)
	}
	moved := plan.Links.Links[0]
	if moved.RawPath != "old/guide.md" || moved.Target != "old/guide.md#topic" || moved.Status != "moved" {
		t.Fatalf("moved record was not restored: %#v", moved)
	}
	if len(moved.Candidates) != 1 || moved.Candidates[0] != "new/guide.md" {
		t.Fatalf("moved candidate was not retained: %#v", moved.Candidates)
	}
	caseMismatch := plan.Links.Links[1]
	if caseMismatch.RawPath != "README.md" || caseMismatch.Target != "README.md" || caseMismatch.Status != "case_mismatch" {
		t.Fatalf("case mismatch record was not restored: %#v", caseMismatch)
	}
	ambiguous := plan.Links.Links[2]
	if ambiguous.RawPath != "guide.md" || ambiguous.Status != "ambiguous" {
		t.Fatalf("unrelated record changed: %#v", ambiguous)
	}
}
