package links

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/review"
)

type failingReviewBatchAppender struct {
	err      error
	requests []review.AppendRequest
}

func (f *failingReviewBatchAppender) AppendBatch(requests []review.AppendRequest) ([]review.StoredEvent, error) {
	f.requests = append([]review.AppendRequest(nil), requests...)
	return nil, f.err
}

func TestApplyAndSaveRestoresSourcesWhenReviewBatchFails(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "README.md")
	oldData := []byte("old contents\n")
	newData := []byte("new contents\n")
	if err := os.WriteFile(path, oldData, 0o644); err != nil {
		t.Fatal(err)
	}
	rewrite, err := NewGeneratedRewriteBytes("file-1", path, oldData, newData, nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := Plan{RepositoryRoot: root, Rewrites: []GeneratedRewrite{rewrite}}

	originalOpen := openReviewBatchStore
	failure := errors.New("review store unavailable")
	appender := &failingReviewBatchAppender{err: failure}
	openReviewBatchStore = func(string) (reviewBatchAppender, error) { return appender, nil }
	t.Cleanup(func() { openReviewBatchStore = originalOpen })

	updates, err := ApplyAndSave(&plan)
	if err == nil || !strings.Contains(err.Error(), failure.Error()) {
		t.Fatalf("error = %v, want review store failure", err)
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
	if len(appender.requests) != 1 {
		t.Fatalf("review requests = %d, want 1", len(appender.requests))
	}
	if len(plan.AppliedChanges) != 0 || len(plan.Suppressions) != 0 {
		t.Fatalf("failed transaction mutated plan state: changes=%#v suppressions=%#v", plan.AppliedChanges, plan.Suppressions)
	}
}
