package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/review"
)

func TestBuildInternalMoveRewritesReturnsErrorsInSourcePathOrder(t *testing.T) {
	root := t.TempDir()
	previousBySource := map[string][]LinkRecord{
		"source-b": {internalMoveTestRecord("source-b")},
		"source-a": {internalMoveTestRecord("source-a")},
	}
	previousByID := map[string]*FileRecord{
		"source-a": internalMoveTestFile("source-a", "a.md", "source-a"),
		"source-b": internalMoveTestFile("source-b", "b.md", "source-b"),
		"target":   internalMoveTestFile("target", "old/target.bin", "target"),
	}
	currentByID := map[string]*FileRecord{
		"source-a": internalMoveTestFile("source-a", "a.md", "source-a"),
		"source-b": internalMoveTestFile("source-b", "b.md", "source-b"),
		"target":   internalMoveTestFile("target", "new/target.bin", "target"),
	}

	_, err := buildInternalMoveRewrites(root, previousBySource, previousByID, currentByID, review.Policy{})
	if err == nil {
		t.Fatal("expected missing-source error")
	}
	firstPath := filepath.Join(root, "a.md")
	if !strings.Contains(err.Error(), firstPath) {
		t.Fatalf("error order was not deterministic: %v", err)
	}
}

func TestInternalMoveRewritesPlanEveryHighFanoutSourceInPathOrder(t *testing.T) {
	root := t.TempDir()
	createBenchmarkRepository(t, root, 32)

	baseline, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(root, "asset-a.bin"), filepath.Join(root, "asset-moved.bin")); err != nil {
		t.Fatal(err)
	}

	plan, err := Reconcile(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Rewrites) != 32 {
		t.Fatalf("rewrites = %d, want 32", len(plan.Rewrites))
	}
	for index := 1; index < len(plan.Rewrites); index++ {
		if plan.Rewrites[index-1].Path > plan.Rewrites[index].Path {
			t.Fatalf("rewrite order is not deterministic: %q before %q", plan.Rewrites[index-1].Path, plan.Rewrites[index].Path)
		}
	}
}

func internalMoveTestRecord(sourceID string) LinkRecord {
	return LinkRecord{
		ID:            sourceID + "-link",
		SourceFileID:  sourceID,
		TargetFileID:  "target",
		Start:         0,
		End:           len("old/target.bin"),
		Syntax:        "markdown",
		RawPath:       "old/target.bin",
		Target:        "old/target.bin",
		Status:        "valid",
		ParserVersion: linkParserVersion,
	}
}

func internalMoveTestFile(id, path, fingerprint string) *FileRecord {
	return &FileRecord{
		ID:          id,
		Scope:       "repository",
		Path:        path,
		Kind:        "file",
		Present:     true,
		Fingerprint: fingerprint,
	}
}
