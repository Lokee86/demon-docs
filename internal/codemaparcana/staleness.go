package codemaparcana

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/codemap"
)

const maxSemanticDiffNodes = 10000

type semanticDiffResult struct {
	Truncated bool `json:"truncated"`
	Nodes     struct {
		Added               []protocolNode `json:"added"`
		Removed             []protocolNode `json:"removed"`
		MetadataChanged     []protocolNode `json:"metadata_changed"`
		RelationshipChanged []protocolNode `json:"relationship_changed"`
	} `json:"nodes"`
}

func (resolver *Resolver) SnapshotID() string {
	if resolver == nil {
		return ""
	}
	return resolver.state.id
}

func (resolver *Resolver) AnalyzeStaleness(ctx context.Context, previousSnapshotID string, entries []codemap.DatasetEntry) ([]codemap.SemanticChange, error) {
	if resolver == nil || previousSnapshotID == "" || previousSnapshotID == resolver.state.id {
		return nil, nil
	}
	previousDirectory, err := historicalSnapshotDirectory(resolver.repositoryRoot, previousSnapshotID)
	if err != nil {
		return nil, fmt.Errorf("open previous Arcana snapshot %s: %w", previousSnapshotID, err)
	}
	diff, err := resolver.semanticDiff(ctx, previousSnapshotID, previousDirectory)
	if err != nil {
		return nil, err
	}
	previousClient, err := resolver.historicalClient(ctx, previousDirectory)
	if err != nil {
		return nil, fmt.Errorf("open previous Arcana query session: %w", err)
	}

	metadata := changedNodeSet(diff.Nodes.MetadataChanged)
	relationships := changedNodeSet(diff.Nodes.RelationshipChanged)
	var changes []codemap.SemanticChange
	for _, entry := range entries {
		if entry.Entry.Kind == codemap.TargetDirectory || entry.Entry.Kind == codemap.TargetGlob {
			continue
		}
		previous, err := resolveHistoricalEntry(ctx, previousClient, entry)
		if err != nil {
			return nil, fmt.Errorf("resolve prior semantic target %s: %w", entry.Entry.Target, err)
		}
		if previous == nil {
			continue
		}
		previousNode := protocolSemanticNode(*previous)
		current := entry.Resolution.SemanticNode
		if current == nil {
			counterpart := currentCounterpart(*previous, diff)
			if counterpart != nil && resolver.pathCurrent(normalizeRepositoryPath(counterpart.Path)) {
				currentNode := protocolSemanticNode(*counterpart)
				changes = append(changes, semanticChange(entry.Entry.Target, changedIdentityKind(previousNode, currentNode), &previousNode, &currentNode))
				appendChangedNodeSignals(&changes, entry.Entry.Target, currentNode, nil, relationships)
			} else if entry.Resolution.Status == codemap.ResolutionAmbiguous {
				changes = append(changes, semanticChange(entry.Entry.Target, codemap.SemanticChangeResolutionAmbiguous, &previousNode, nil))
			} else {
				changes = append(changes, semanticChange(entry.Entry.Target, codemap.SemanticChangeDisappeared, &previousNode, nil))
			}
			continue
		}

		identityChanged := semanticIdentityChanged(previousNode, *current)
		if identityChanged {
			changes = append(changes, semanticChange(entry.Entry.Target, changedIdentityKind(previousNode, *current), &previousNode, current))
			appendChangedNodeSignals(&changes, entry.Entry.Target, *current, nil, relationships)
		} else {
			appendChangedNodeSignals(&changes, entry.Entry.Target, *current, metadata, relationships)
		}
	}
	return dedupeSemanticChanges(changes), nil
}

func (resolver *Resolver) semanticDiff(ctx context.Context, previousSnapshotID, previousDirectory string) (semanticDiffResult, error) {
	resolver.stalenessMu.Lock()
	defer resolver.stalenessMu.Unlock()
	if diff, ok := resolver.stalenessDiffs[previousSnapshotID]; ok {
		return diff, nil
	}
	var diff semanticDiffResult
	if err := resolver.client.query(ctx, map[string]any{
		"op": "diff", "other_snapshot": previousDirectory, "limit": maxSemanticDiffNodes,
	}, &diff); err != nil {
		return semanticDiffResult{}, fmt.Errorf("diff Arcana snapshots: %w", err)
	}
	if diff.Truncated {
		return semanticDiffResult{}, fmt.Errorf("Arcana semantic diff exceeds %d nodes", maxSemanticDiffNodes)
	}
	if resolver.stalenessDiffs == nil {
		resolver.stalenessDiffs = map[string]semanticDiffResult{}
	}
	resolver.stalenessDiffs[previousSnapshotID] = diff
	return diff, nil
}

func (resolver *Resolver) historicalClient(ctx context.Context, directory string) (queryClient, error) {
	resolver.historyMu.Lock()
	defer resolver.historyMu.Unlock()
	if client, ok := resolver.historicalClients[directory]; ok {
		return client, nil
	}
	if resolver.openSnapshot == nil {
		return nil, fmt.Errorf("historical Arcana query opener is unavailable")
	}
	client, err := resolver.openSnapshot(ctx, directory)
	if err != nil {
		return nil, err
	}
	if resolver.historicalClients == nil {
		resolver.historicalClients = map[string]queryClient{}
	}
	resolver.historicalClients[directory] = client
	return client, nil
}

func historicalSnapshotDirectory(repositoryRoot, snapshotID string) (string, error) {
	if !validSnapshotID(snapshotID) {
		return "", fmt.Errorf("invalid snapshot ID")
	}
	directory := filepath.Join(repositoryRoot, ".arcana", "snapshots", strings.TrimPrefix(snapshotID, "sha256:"))
	for _, name := range []string{"repository.manifest", "lexicon.snapshot"} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil || info.IsDir() {
			return "", fmt.Errorf("snapshot is unavailable or incomplete")
		}
	}
	return directory, nil
}
