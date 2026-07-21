package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/review"
)

type trackingReviewBatchAppender struct {
	calls int
}

func (t *trackingReviewBatchAppender) AppendBatch([]review.AppendRequest) ([]review.StoredEvent, error) {
	t.calls++
	return nil, nil
}

func TestApplyAndSaveRestoresSourcesWhenGeneratedRewriteVerificationFails(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	oldData := []byte("[target](old.md)\n")
	if err := os.WriteFile(path, oldData, 0o644); err != nil {
		t.Fatal(err)
	}

	transformation := LinkTransformation{
		LinkID:         "link-1",
		Start:          9,
		End:            15,
		OldDestination: "old.md",
		NewDestination: "new.md",
	}
	rewrite, err := NewGeneratedRewriteBytes(
		"file-1",
		path,
		oldData,
		[]byte("[target](new.md)\n"),
		[]LinkTransformation{transformation},
	)
	if err != nil {
		t.Fatal(err)
	}
	plan := Plan{
		RepositoryRoot: root,
		Rewrites:       []GeneratedRewrite{rewrite},
		Links: LinksManifest{Links: []LinkRecord{{
			ID:           "link-1",
			SourceFileID: "file-1",
			Ordinal:      0,
			Target:       "missing.md",
		}}},
	}

	originalOpen := openReviewBatchStore
	appender := &trackingReviewBatchAppender{}
	openReviewBatchStore = func(string) (reviewBatchAppender, error) { return appender, nil }
	t.Cleanup(func() { openReviewBatchStore = originalOpen })

	updates, err := ApplyAndSave(&plan)
	if err == nil || !strings.Contains(err.Error(), "generated rewrite verification could not find link") {
		t.Fatalf("error = %v, want generated rewrite verification failure", err)
	}
	if updates != 0 {
		t.Fatalf("updates = %d, want 0", updates)
	}
	current, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(current) != string(oldData) {
		t.Fatalf("source was not restored: got %q want %q", current, oldData)
	}
	if appender.calls != 0 {
		t.Fatalf("review history append calls = %d, want 0", appender.calls)
	}
	if len(plan.AppliedChanges) != 0 || len(plan.Suppressions) != 0 {
		t.Fatalf("failed transaction mutated plan state: changes=%#v suppressions=%#v", plan.AppliedChanges, plan.Suppressions)
	}
}
