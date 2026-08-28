package codemaprun

import (
	"context"

	"github.com/Lokee86/demon-docs/internal/codemap"
	"github.com/Lokee86/demon-docs/internal/codemapsemantic"
)

func planSemanticBaseline(
	ctx context.Context,
	provider codemap.SemanticStalenessProvider,
	baseline codemapsemantic.Baseline,
	exists bool,
	documentPath string,
	before, after []byte,
	entries []codemap.DatasetEntry,
) ([]codemap.SemanticChange, *codemapsemantic.Baseline, error) {
	if provider == nil || provider.SnapshotID() == "" {
		return nil, nil, nil
	}
	beforeDigest := codemapsemantic.Digest(before)
	var changes []codemap.SemanticChange
	var err error
	if exists && baseline.DocumentSHA256 == beforeDigest && baseline.SnapshotID != provider.SnapshotID() {
		changes, err = provider.AnalyzeStaleness(ctx, baseline.SnapshotID, entries)
		if err != nil {
			return nil, nil, err
		}
	}
	if exists && baseline.DocumentSHA256 == beforeDigest && len(changes) > 0 {
		return changes, nil, nil
	}
	update := &codemapsemantic.Baseline{
		Document: documentPath, DocumentSHA256: codemapsemantic.Digest(after), SnapshotID: provider.SnapshotID(),
	}
	return changes, update, nil
}
