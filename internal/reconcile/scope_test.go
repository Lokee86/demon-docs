package reconcile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lokee86/demon-docs/internal/model"
)

func TestApplyWithinRejectsOutsideWriteBeforeMutation(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "inside.md")
	outside := filepath.Join(t.TempDir(), "outside.md")
	result := model.ReconcileResult{Updates: []model.FileUpdate{
		{Path: inside, NewText: "inside"},
		{Path: outside, NewText: "outside"},
	}}

	if _, err := ApplyWithin(result, root); err == nil {
		t.Fatal("outside write was accepted")
	}
	if _, err := os.Stat(inside); !os.IsNotExist(err) {
		t.Fatal("inside update was written before boundary validation completed")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatal("outside update was written")
	}
}
