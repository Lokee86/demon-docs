package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

const (
	reviewBatchSchemaVersion = 1
	reviewBatchBlobName      = "batch.json"
)

type AppendRequest struct {
	Event  Event
	Before []byte
	After  []byte
}

type preparedAppend struct {
	event  Event
	before []byte
	after  []byte
}

type persistedReviewBatch struct {
	SchemaVersion int                    `json:"schema_version"`
	Entries       []persistedReviewEntry `json:"entries"`
}

type persistedReviewEntry struct {
	Event  Event          `json:"event"`
	Before persistedBytes `json:"before"`
	After  persistedBytes `json:"after"`
}

type persistedBytes struct {
	Present bool   `json:"present"`
	Data    []byte `json:"data,omitempty"`
}

func (s *Store) Append(event Event, before, after []byte) (StoredEvent, error) {
	stored, err := s.AppendBatch([]AppendRequest{{Event: event, Before: before, After: after}})
	if err != nil {
		return StoredEvent{}, err
	}
	return stored[0], nil
}

// AppendBatch stores one atomic review commit containing every requested event.
// History expands the commit back into individual StoredEvent values so callers
// retain per-change history and undo behavior without one Git commit per file.
func (s *Store) AppendBatch(requests []AppendRequest) ([]StoredEvent, error) {
	if len(requests) == 0 {
		return nil, nil
	}
	prepared := make([]preparedAppend, len(requests))
	for index, request := range requests {
		item, err := prepareAppend(request)
		if err != nil {
			return nil, fmt.Errorf("prepare review event %d: %w", index, err)
		}
		prepared[index] = item
	}
	payload, err := encodeReviewBatch(prepared)
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < 3; attempt++ {
		stored, conflict, err := s.appendBatchOnce(prepared, payload)
		if err != nil {
			return nil, err
		}
		if !conflict {
			return stored, nil
		}
	}
	return nil, errors.New("review history changed during append")
}

func prepareAppend(request AppendRequest) (preparedAppend, error) {
	event := request.Event
	if event.SchemaVersion == 0 {
		event.SchemaVersion = SchemaVersion
	}
	if event.SchemaVersion != SchemaVersion {
		return preparedAppend{}, fmt.Errorf("unsupported review event schema %d", event.SchemaVersion)
	}
	if event.ID == "" {
		event.ID = NewID("ev")
	}
	if event.Time.IsZero() {
		event.Time = time.Now().UTC()
	}
	return preparedAppend{
		event:  event,
		before: clone(request.Before),
		after:  clone(request.After),
	}, nil
}

func encodeReviewBatch(prepared []preparedAppend) ([]byte, error) {
	batch := persistedReviewBatch{
		SchemaVersion: reviewBatchSchemaVersion,
		Entries:       make([]persistedReviewEntry, len(prepared)),
	}
	for index, item := range prepared {
		batch.Entries[index] = persistedReviewEntry{
			Event:  item.event,
			Before: persistBytes(item.before),
			After:  persistBytes(item.after),
		}
	}
	payload, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("encode review batch: %w", err)
	}
	return payload, nil
}

func persistBytes(data []byte) persistedBytes {
	return persistedBytes{Present: data != nil, Data: clone(data)}
}

func (stored persistedBytes) bytes() []byte {
	if !stored.Present {
		return nil
	}
	if stored.Data == nil {
		return []byte{}
	}
	return clone(stored.Data)
}

func (s *Store) appendBatchOnce(prepared []preparedAppend, payload []byte) ([]StoredEvent, bool, error) {
	var stored []StoredEvent
	var conflict bool
	var commitHash plumbing.Hash
	err := ddrepo.WithRepositoryWriteLock(s.path, func() error {
		current, err := s.repository.Storer.Reference(reviewReference)
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			current = nil
		} else if err != nil {
			return fmt.Errorf("read review reference: %w", err)
		}

		parent := plumbing.ZeroHash
		if current != nil {
			parent = current.Hash()
		}
		stored, commitHash, err = s.writeReviewBatchCommit(prepared, payload, parent)
		if err != nil {
			return err
		}

		updated := plumbing.NewHashReference(reviewReference, commitHash)
		if err := s.repository.Storer.CheckAndSetReference(updated, current); err != nil {
			if errors.Is(err, storage.ErrReferenceHasChanged) {
				conflict = true
				return nil
			}
			return fmt.Errorf("advance review reference: %w", err)
		}
		// The review reference is now durable. Compaction is best effort and
		// cannot turn a completed logical append into an error.
		_, _ = ddrepo.CompactIfNeeded(s.repository, s.path, s.compaction)
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return stored, conflict, nil
}

func (s *Store) writeReviewBatchCommit(prepared []preparedAppend, payload []byte, parent plumbing.Hash) ([]StoredEvent, plumbing.Hash, error) {
	batchHash, err := writeBlob(s.repository, payload)
	if err != nil {
		return nil, plumbing.ZeroHash, err
	}
	treeHash, err := writeTree(s.repository, []object.TreeEntry{{Name: reviewBatchBlobName, Mode: filemode.Regular, Hash: batchHash}})
	if err != nil {
		return nil, plumbing.ZeroHash, err
	}
	when := prepared[len(prepared)-1].event.Time
	commit := object.Commit{
		Author:    object.Signature{Name: "Demon Docs", Email: "ddocs@local", When: when},
		Committer: object.Signature{Name: "Demon Docs", Email: "ddocs@local", When: when},
		Message:   reviewBatchMessage(prepared),
		TreeHash:  treeHash,
	}
	if !parent.IsZero() {
		commit.ParentHashes = []plumbing.Hash{parent}
	}
	encoded := s.repository.Storer.NewEncodedObject()
	if err := commit.Encode(encoded); err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("encode review commit: %w", err)
	}
	commitHash, err := s.repository.Storer.SetEncodedObject(encoded)
	if err != nil {
		return nil, plumbing.ZeroHash, fmt.Errorf("write review commit: %w", err)
	}
	stored := make([]StoredEvent, len(prepared))
	for index, item := range prepared {
		stored[index] = StoredEvent{
			Event:      item.event,
			CommitHash: commitHash.String(),
			Before:     clone(item.before),
			After:      clone(item.after),
		}
	}
	return stored, commitHash, nil
}

func reviewBatchMessage(prepared []preparedAppend) string {
	if len(prepared) == 1 {
		return eventMessage(prepared[0].event)
	}
	runID := ""
	for index, item := range prepared {
		if item.event.Change == nil || item.event.Change.RunID == "" {
			runID = ""
			break
		}
		if index == 0 {
			runID = item.event.Change.RunID
			continue
		}
		if item.event.Change.RunID != runID {
			runID = ""
			break
		}
	}
	if runID != "" {
		return fmt.Sprintf("run %s (%d changes)", runID, len(prepared))
	}
	return fmt.Sprintf("review batch (%d events)", len(prepared))
}
