package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

type AppendRequest struct {
	Event  Event
	Before []byte
	After  []byte
}

type preparedAppend struct {
	event   Event
	payload []byte
	before  []byte
	after   []byte
}

func (s *Store) Append(event Event, before, after []byte) (StoredEvent, error) {
	stored, err := s.AppendBatch([]AppendRequest{{Event: event, Before: before, After: after}})
	if err != nil {
		return StoredEvent{}, err
	}
	return stored[0], nil
}

// AppendBatch writes a chain of review commits and publishes the chain with one
// compare-and-swap reference update. Callers observe either every event in the
// batch or none of them.
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

	for attempt := 0; attempt < 3; attempt++ {
		stored, conflict, err := s.appendBatchOnce(prepared)
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
	payload, err := json.Marshal(event)
	if err != nil {
		return preparedAppend{}, fmt.Errorf("encode review event: %w", err)
	}
	return preparedAppend{
		event:   event,
		payload: payload,
		before:  clone(request.Before),
		after:   clone(request.After),
	}, nil
}

func (s *Store) appendBatchOnce(prepared []preparedAppend) ([]StoredEvent, bool, error) {
	current, err := s.repository.Storer.Reference(reviewReference)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		current = nil
	} else if err != nil {
		return nil, false, fmt.Errorf("read review reference: %w", err)
	}

	parent := plumbing.ZeroHash
	if current != nil {
		parent = current.Hash()
	}
	stored := make([]StoredEvent, len(prepared))
	for index, item := range prepared {
		entry, hash, err := s.writeReviewCommit(item, parent)
		if err != nil {
			return nil, false, err
		}
		stored[index] = entry
		parent = hash
	}

	updated := plumbing.NewHashReference(reviewReference, parent)
	if err := s.repository.Storer.CheckAndSetReference(updated, current); err != nil {
		if errors.Is(err, storage.ErrReferenceHasChanged) {
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("advance review reference: %w", err)
	}
	return stored, false, nil
}

func (s *Store) writeReviewCommit(item preparedAppend, parent plumbing.Hash) (StoredEvent, plumbing.Hash, error) {
	entries := []object.TreeEntry{{Name: "event.json", Mode: filemode.Regular, Hash: plumbing.ZeroHash}}
	var err error
	entries[0].Hash, err = writeBlob(s.repository, item.payload)
	if err != nil {
		return StoredEvent{}, plumbing.ZeroHash, err
	}
	if item.before != nil {
		hash, err := writeBlob(s.repository, item.before)
		if err != nil {
			return StoredEvent{}, plumbing.ZeroHash, err
		}
		entries = append(entries, object.TreeEntry{Name: "before", Mode: filemode.Regular, Hash: hash})
	}
	if item.after != nil {
		hash, err := writeBlob(s.repository, item.after)
		if err != nil {
			return StoredEvent{}, plumbing.ZeroHash, err
		}
		entries = append(entries, object.TreeEntry{Name: "after", Mode: filemode.Regular, Hash: hash})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	treeHash, err := writeTree(s.repository, entries)
	if err != nil {
		return StoredEvent{}, plumbing.ZeroHash, err
	}
	commit := object.Commit{
		Author:    object.Signature{Name: "Demon Docs", Email: "ddocs@local", When: item.event.Time},
		Committer: object.Signature{Name: "Demon Docs", Email: "ddocs@local", When: item.event.Time},
		Message:   eventMessage(item.event),
		TreeHash:  treeHash,
	}
	if !parent.IsZero() {
		commit.ParentHashes = []plumbing.Hash{parent}
	}
	encoded := s.repository.Storer.NewEncodedObject()
	if err := commit.Encode(encoded); err != nil {
		return StoredEvent{}, plumbing.ZeroHash, fmt.Errorf("encode review commit: %w", err)
	}
	commitHash, err := s.repository.Storer.SetEncodedObject(encoded)
	if err != nil {
		return StoredEvent{}, plumbing.ZeroHash, fmt.Errorf("write review commit: %w", err)
	}
	return StoredEvent{
		Event:      item.event,
		CommitHash: commitHash.String(),
		Before:     clone(item.before),
		After:      clone(item.after),
	}, commitHash, nil
}
