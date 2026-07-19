package ddrepo

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestRepositoryTransactionPersistsAndReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ddocs")
	repository, err := Init(path)
	if err != nil {
		t.Fatal(err)
	}
	initialRoot, err := repository.CurrentRoot()
	if err != nil {
		t.Fatal(err)
	}
	if initialRoot.IsZero() {
		t.Fatal("initialized repository has a zero root")
	}
	if err := repository.Transaction(func(tx *Transaction) error {
		if err := tx.Write("file/alpha", []byte("value")); err != nil {
			return err
		}
		return tx.Write("source/beta", []byte("links"))
	}); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := reopened.Begin()
	if err != nil {
		t.Fatal(err)
	}
	value, err := tx.Read("file/alpha")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(value, []byte("value")) {
		t.Fatalf("value = %q", value)
	}
	names, err := tx.Names("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "file/alpha" || names[1] != "source/beta" {
		t.Fatalf("names = %v", names)
	}
}

func TestRepositoryRejectsStaleTransaction(t *testing.T) {
	repository, err := Init(filepath.Join(t.TempDir(), ".ddocs"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := repository.Begin()
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Write("file/first", []byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := second.Write("file/second", []byte("two")); err != nil {
		t.Fatal(err)
	}
	if err := second.Commit(); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale commit error = %v, want ErrConflict", err)
	}
}

func TestSingleRecordUpdateOnlyChangesItsShard(t *testing.T) {
	repository, err := Init(filepath.Join(t.TempDir(), ".ddocs"))
	if err != nil {
		t.Fatal(err)
	}
	firstName, secondName := namesInDifferentShards()
	if err := repository.Transaction(func(tx *Transaction) error {
		if err := tx.Write(firstName, []byte("first")); err != nil {
			return err
		}
		return tx.Write(secondName, []byte("second"))
	}); err != nil {
		t.Fatal(err)
	}
	before := rootEntries(t, repository)
	beforeRoot, err := repository.CurrentRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write(firstName, []byte("changed"))
	}); err != nil {
		t.Fatal(err)
	}
	after := rootEntries(t, repository)
	afterRoot, err := repository.CurrentRoot()
	if err != nil {
		t.Fatal(err)
	}
	if beforeRoot == afterRoot {
		t.Fatal("changed record did not advance the root")
	}
	firstShard := shardName(firstName)
	secondShard := shardName(secondName)
	if before[firstShard] == after[firstShard] {
		t.Fatal("changed record shard hash did not change")
	}
	if before[secondShard] != after[secondShard] {
		t.Fatal("unaffected shard hash changed")
	}
	stableRoot := afterRoot
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write(firstName, []byte("changed"))
	}); err != nil {
		t.Fatal(err)
	}
	if root, err := repository.CurrentRoot(); err != nil || root != stableRoot {
		t.Fatalf("byte-identical update advanced root: root=%s err=%v", root, err)
	}
}

func namesInDifferentShards() (string, string) {
	first := "file/record-0"
	for index := 1; ; index++ {
		candidate := "file/record-" + string(rune('a'+index%26))
		if shardName(candidate) != shardName(first) {
			return first, candidate
		}
	}
}

func rootEntries(t *testing.T, repository *Repository) map[string]plumbing.Hash {
	t.Helper()
	rootHash, err := repository.CurrentRoot()
	if err != nil {
		t.Fatal(err)
	}
	tree, err := object.GetTree(repository.store, rootHash)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]plumbing.Hash, len(tree.Entries))
	for _, entry := range tree.Entries {
		result[entry.Name] = entry.Hash
	}
	return result
}
