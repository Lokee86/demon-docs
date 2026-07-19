package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/links"
)

func TestRollbackAfterReviewFailureRestoresUndoSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.md")
	before := []byte("before undo\n")
	after := []byte("after undo\n")
	if err := os.WriteFile(path, before, 0o644); err != nil {
		t.Fatal(err)
	}
	rewrite, err := links.NewGeneratedRewriteBytes("file-1", path, before, after, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := links.ApplyGenerated([]links.GeneratedRewrite{rewrite}); err != nil {
		t.Fatal(err)
	}

	recordErr := errors.New("review append failed")
	err = rollbackAfterReviewFailure([]links.GeneratedRewrite{rewrite}, recordErr)
	if err == nil || !strings.Contains(err.Error(), recordErr.Error()) {
		t.Fatalf("error = %v, want review append failure", err)
	}
	current, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(current) != string(before) {
		t.Fatalf("undo source was not restored: got %q want %q", current, before)
	}
}
