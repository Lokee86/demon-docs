package codemaprun

import (
	"context"
	"testing"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapsemantic"
)

type fakeSemanticStaleness struct {
	snapshot string
	changes  []codemap.SemanticChange
	calls    int
}

func (provider *fakeSemanticStaleness) SnapshotID() string { return provider.snapshot }
func (provider *fakeSemanticStaleness) AnalyzeStaleness(context.Context, string, []codemap.DatasetEntry) ([]codemap.SemanticChange, error) {
	provider.calls++
	return append([]codemap.SemanticChange(nil), provider.changes...), nil
}

func TestSemanticBaselineDoesNotAdvanceUntouchedStaleDocument(t *testing.T) {
	before := []byte("same")
	provider := &fakeSemanticStaleness{snapshot: "new", changes: []codemap.SemanticChange{{Target: "src/a.go", Kind: codemap.SemanticChangeMetadata}}}
	baseline := codemapsemantic.Baseline{Document: "docs/a.md", DocumentSHA256: codemapsemantic.Digest(before), SnapshotID: "old"}
	changes, update, err := planSemanticBaseline(context.Background(), provider, baseline, true, "docs/a.md", before, before, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || update != nil || provider.calls != 1 {
		t.Fatalf("unexpected stale baseline plan: changes=%#v update=%#v calls=%d", changes, update, provider.calls)
	}
}

func TestSemanticBaselineAdvancesAfterDocumentEdit(t *testing.T) {
	provider := &fakeSemanticStaleness{snapshot: "new", changes: []codemap.SemanticChange{{Target: "src/a.go", Kind: codemap.SemanticChangeMetadata}}}
	baseline := codemapsemantic.Baseline{Document: "docs/a.md", DocumentSHA256: codemapsemantic.Digest([]byte("old doc")), SnapshotID: "old"}
	changes, update, err := planSemanticBaseline(context.Background(), provider, baseline, true, "docs/a.md", []byte("edited"), []byte("edited"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 || update == nil || update.SnapshotID != "new" || provider.calls != 0 {
		t.Fatalf("unexpected edited baseline plan: changes=%#v update=%#v calls=%d", changes, update, provider.calls)
	}
}

func TestSemanticBaselineAdvancesWhenMappedSemanticsDidNotChange(t *testing.T) {
	before := []byte("same")
	provider := &fakeSemanticStaleness{snapshot: "new"}
	baseline := codemapsemantic.Baseline{Document: "docs/a.md", DocumentSHA256: codemapsemantic.Digest(before), SnapshotID: "old"}
	changes, update, err := planSemanticBaseline(context.Background(), provider, baseline, true, "docs/a.md", before, before, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 || update == nil || provider.calls != 1 {
		t.Fatalf("unexpected clean semantic plan: changes=%#v update=%#v calls=%d", changes, update, provider.calls)
	}
}
