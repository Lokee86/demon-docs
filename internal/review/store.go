package review

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const reviewReference = plumbing.ReferenceName("refs/ddocs/review")

type Store struct {
	repository *git.Repository
}

func Open(repositoryRoot string) (*Store, error) {
	repository, err := git.PlainOpen(filepath.Join(repositoryRoot, ".ddocs"))
	if err != nil {
		return nil, fmt.Errorf("open ddocs review store: %w", err)
	}
	return &Store{repository: repository}, nil
}

func (s *Store) History(limit int) ([]StoredEvent, error) {
	ref, err := s.repository.Storer.Reference(reviewReference)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read review reference: %w", err)
	}
	var result []StoredEvent
	hash := ref.Hash()
	for !hash.IsZero() && (limit <= 0 || len(result) < limit) {
		commit, err := object.GetCommit(s.repository.Storer, hash)
		if err != nil {
			return nil, fmt.Errorf("read review commit %s: %w", hash, err)
		}
		stored, err := readStoredEvent(commit)
		if err != nil {
			return nil, err
		}
		result = append(result, stored)
		if len(commit.ParentHashes) == 0 {
			break
		}
		hash = commit.ParentHashes[0]
	}
	return result, nil
}

func (s *Store) Find(id string) (StoredEvent, error) {
	history, err := s.History(0)
	if err != nil {
		return StoredEvent{}, err
	}
	for _, event := range history {
		if event.ID == id || event.Change != nil && event.Change.ID == id || event.Decision != nil && event.Decision.ID == id {
			return event, nil
		}
	}
	return StoredEvent{}, fmt.Errorf("review record not found: %s", id)
}

func readStoredEvent(commit *object.Commit) (StoredEvent, error) {
	tree, err := commit.Tree()
	if err != nil {
		return StoredEvent{}, fmt.Errorf("read review tree %s: %w", commit.Hash, err)
	}
	payload, err := readTreeBlob(tree, "event.json")
	if err != nil {
		return StoredEvent{}, err
	}
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return StoredEvent{}, fmt.Errorf("decode review event %s: %w", commit.Hash, err)
	}
	if event.SchemaVersion != SchemaVersion {
		return StoredEvent{}, fmt.Errorf("unsupported review event schema %d", event.SchemaVersion)
	}
	before, _ := readTreeBlob(tree, "before")
	after, _ := readTreeBlob(tree, "after")
	return StoredEvent{Event: event, CommitHash: commit.Hash.String(), Before: before, After: after}, nil
}

func readTreeBlob(tree *object.Tree, name string) ([]byte, error) {
	file, err := tree.File(name)
	if err != nil {
		return nil, err
	}
	reader, err := file.Blob.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func writeBlob(repository *git.Repository, data []byte) (plumbing.Hash, error) {
	encoded := repository.Storer.NewEncodedObject()
	encoded.SetType(plumbing.BlobObject)
	encoded.SetSize(int64(len(data)))
	writer, err := encoded.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return plumbing.ZeroHash, err
	}
	if err := writer.Close(); err != nil {
		return plumbing.ZeroHash, err
	}
	return repository.Storer.SetEncodedObject(encoded)
}

func writeTree(repository *git.Repository, entries []object.TreeEntry) (plumbing.Hash, error) {
	encoded := repository.Storer.NewEncodedObject()
	if err := (&object.Tree{Entries: entries}).Encode(encoded); err != nil {
		return plumbing.ZeroHash, err
	}
	return repository.Storer.SetEncodedObject(encoded)
}

func eventMessage(event Event) string {
	if event.Change != nil {
		return "change " + event.Change.ID
	}
	if event.Decision != nil {
		return string(event.Decision.Action) + " " + event.Decision.ID
	}
	return "review event " + event.ID
}

func clone(data []byte) []byte {
	if data == nil {
		return nil
	}
	return append([]byte(nil), data...)
}
