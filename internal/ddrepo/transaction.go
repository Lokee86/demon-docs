package ddrepo

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Transaction struct {
	repo   *Repository
	base   *plumbing.Reference
	shards map[string]plumbing.Hash
	loaded map[string]map[string][]byte
	dirty  map[string]bool
	closed bool
}

type Tx = Transaction

func (r *Repository) Begin() (*Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ref, err := r.currentReference()
	if err != nil && !errors.Is(err, ErrMissingState) {
		return nil, err
	}
	tx := &Transaction{
		repo:   r,
		shards: make(map[string]plumbing.Hash),
		loaded: make(map[string]map[string][]byte),
		dirty:  make(map[string]bool),
	}
	if ref == nil {
		return tx, nil
	}
	root, err := object.GetTree(r.store, ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("read ddocs root: %w", err)
	}
	for _, entry := range root.Entries {
		if entry.Mode != filemode.Regular || !isShardName(entry.Name) {
			return nil, fmt.Errorf("invalid ddocs shard entry %q", entry.Name)
		}
		tx.shards[entry.Name] = entry.Hash
	}
	tx.base = ref
	return tx, nil
}

func (tx *Transaction) Read(name string) ([]byte, error) {
	if err := tx.checkOpen(); err != nil {
		return nil, err
	}
	if err := validateRecordName(name); err != nil {
		return nil, err
	}
	records, err := tx.loadShard(shardName(name))
	if err != nil {
		return nil, err
	}
	value, ok := records[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrRecordAbsent, name)
	}
	return append([]byte(nil), value...), nil
}

func (tx *Transaction) Get(name string) ([]byte, error) { return tx.Read(name) }

func (tx *Transaction) Write(name string, value []byte) error {
	if err := tx.checkOpen(); err != nil {
		return err
	}
	if err := validateRecordName(name); err != nil {
		return err
	}
	shard := shardName(name)
	records, err := tx.loadShard(shard)
	if err != nil {
		return err
	}
	if current, ok := records[name]; ok && bytes.Equal(current, value) {
		return nil
	}
	records[name] = append([]byte(nil), value...)
	tx.dirty[shard] = true
	return nil
}

func (tx *Transaction) Put(name string, value []byte) error { return tx.Write(name, value) }

func (tx *Transaction) Delete(name string) error {
	if err := tx.checkOpen(); err != nil {
		return err
	}
	if err := validateRecordName(name); err != nil {
		return err
	}
	shard := shardName(name)
	records, err := tx.loadShard(shard)
	if err != nil {
		return err
	}
	if _, ok := records[name]; !ok {
		return nil
	}
	delete(records, name)
	tx.dirty[shard] = true
	return nil
}

func (tx *Transaction) Names(prefix string) ([]string, error) {
	if err := tx.checkOpen(); err != nil {
		return nil, err
	}
	allShards := make(map[string]bool, len(tx.shards)+len(tx.loaded))
	for shard := range tx.shards {
		allShards[shard] = true
	}
	for shard := range tx.loaded {
		allShards[shard] = true
	}
	for shard := range allShards {
		if _, err := tx.loadShard(shard); err != nil {
			return nil, err
		}
	}
	var names []string
	for _, records := range tx.loaded {
		for name := range records {
			if strings.HasPrefix(name, prefix) {
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

func (tx *Transaction) Commit() error {
	if err := tx.checkOpen(); err != nil {
		return err
	}
	defer func() { tx.closed = true }()

	return WithRepositoryWriteLock(tx.repo.path, func() error {
		tx.repo.mu.Lock()
		defer tx.repo.mu.Unlock()
		current, err := tx.repo.currentReference()
		if err != nil && !errors.Is(err, ErrMissingState) {
			return err
		}
		if !sameReference(current, tx.base) {
			return ErrConflict
		}
		entries := make(map[string]plumbing.Hash, len(tx.shards))
		for shard, hash := range tx.shards {
			entries[shard] = hash
		}
		dirty := make([]string, 0, len(tx.dirty))
		for shard := range tx.dirty {
			dirty = append(dirty, shard)
		}
		sort.Strings(dirty)
		for _, shard := range dirty {
			records := tx.loaded[shard]
			if len(records) == 0 {
				delete(entries, shard)
				continue
			}
			hash, err := writeShard(tx.repo.store, records)
			if err != nil {
				return err
			}
			entries[shard] = hash
		}
		rootHash, err := writeRoot(tx.repo.store, entries)
		if err != nil {
			return err
		}
		if current != nil && current.Hash() == rootHash {
			return nil
		}
		if err := tx.repo.store.CheckAndSetReference(plumbing.NewHashReference(stateReference, rootHash), current); err != nil {
			return err
		}
		tx.repo.compactAfterWrite()
		return nil
	})
}

func (tx *Transaction) loadShard(name string) (map[string][]byte, error) {
	if records, ok := tx.loaded[name]; ok {
		return records, nil
	}
	hash, ok := tx.shards[name]
	if !ok {
		records := make(map[string][]byte)
		tx.loaded[name] = records
		return records, nil
	}
	records, err := readShard(tx.repo.store, name, hash)
	if err != nil {
		return nil, err
	}
	tx.loaded[name] = records
	return records, nil
}

func (tx *Transaction) checkOpen() error {
	if tx == nil || tx.closed {
		return ErrClosed
	}
	return nil
}

func sameReference(left, right *plumbing.Reference) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Name() == right.Name() && left.Hash() == right.Hash()
}

func isShardName(name string) bool {
	if len(name) != 1 {
		return false
	}
	for _, character := range name {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
