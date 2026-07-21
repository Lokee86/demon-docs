package ddrepo

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
)

const stateReference = plumbing.ReferenceName("refs/ddocs/state")

var (
	ErrConflict     = errors.New("ddrepo: transaction conflict")
	ErrMissingState = errors.New("ddrepo: state is missing")
	ErrRecordAbsent = errors.New("ddrepo: record is absent")
	ErrClosed       = errors.New("ddrepo: transaction is closed")

	ErrAbsentRecord      = ErrRecordAbsent
	ErrRecordNotFound    = ErrRecordAbsent
	ErrNotFound          = ErrRecordAbsent
	ErrTransactionClosed = ErrClosed
	ErrClosedTransaction = ErrClosed
)

type Repository struct {
	store      storage.Storer
	git        *git.Repository
	mu         sync.Mutex
	path       string
	compaction CompactionThresholds
}

func Init(path string) (*Repository, error) {
	return InitWithOptions(path, Options{Compaction: DefaultCompactionThresholds()})
}

func InitWithOptions(path string, options Options) (*Repository, error) {
	storagePath, err := ddocsPath(path)
	if err != nil {
		return nil, err
	}
	gitRepository, err := git.PlainInit(storagePath, true)
	if err != nil {
		return nil, err
	}
	repository := &Repository{store: gitRepository.Storer, git: gitRepository, path: storagePath, compaction: options.Compaction}
	root, err := writeRoot(repository.store, nil)
	if err != nil {
		return nil, fmt.Errorf("write empty ddocs root: %w", err)
	}
	if err := repository.store.SetReference(plumbing.NewHashReference(stateReference, root)); err != nil {
		return nil, fmt.Errorf("write ddocs state: %w", err)
	}
	return repository, nil
}

func Open(path string) (*Repository, error) {
	return OpenWithOptions(path, Options{Compaction: DefaultCompactionThresholds()})
}

func OpenWithOptions(path string, options Options) (*Repository, error) {
	storagePath, err := ddocsPath(path)
	if err != nil {
		return nil, err
	}
	gitRepository, err := git.PlainOpen(storagePath)
	if err != nil {
		return nil, err
	}
	return &Repository{store: gitRepository.Storer, git: gitRepository, path: storagePath, compaction: options.Compaction}, nil
}

func (r *Repository) Transaction(fn func(*Transaction) error) error {
	tx, err := r.Begin()
	if err != nil {
		return err
	}
	if fn == nil {
		tx.closed = true
		return errors.New("ddrepo: transaction callback is nil")
	}
	if err := fn(tx); err != nil {
		tx.closed = true
		return err
	}
	if tx.closed {
		return nil
	}
	return tx.Commit()
}

func (r *Repository) CurrentRoot() (plumbing.Hash, error) {
	if r == nil {
		return plumbing.ZeroHash, fmt.Errorf("%w: repository is nil", ErrMissingState)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	ref, err := r.currentReference()
	if err != nil {
		return plumbing.ZeroHash, err
	}
	return ref.Hash(), nil
}

func (r *Repository) currentReference() (*plumbing.Reference, error) {
	if r == nil || r.store == nil {
		return nil, fmt.Errorf("%w: repository is nil", ErrMissingState)
	}
	ref, err := r.store.Reference(stateReference)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, ErrMissingState
	}
	if err != nil {
		return nil, fmt.Errorf("read ddocs state: %w", err)
	}
	if ref == nil || ref.Type() != plumbing.HashReference || ref.Hash().IsZero() {
		return nil, ErrMissingState
	}
	return ref, nil
}

func ddocsPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("ddrepo: path is empty")
	}
	path = filepath.Clean(path)
	if filepath.Base(path) == ".ddocs" {
		return path, nil
	}
	return filepath.Join(path, ".ddocs"), nil
}
