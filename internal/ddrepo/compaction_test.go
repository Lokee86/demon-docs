package ddrepo

import (
	"errors"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestDefaultCompactionIsDisabled(t *testing.T) {
	thresholds := DefaultCompactionThresholds()
	if thresholds.LooseFileCount != 0 || thresholds.LooseBytes != 0 {
		t.Fatalf("default compaction must remain disabled without a cross-process repository lock: %+v", thresholds)
	}
}

func TestCompactionWaitsUntilCountIsAboveThreshold(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ddocs")
	repository, err := InitWithOptions(path, Options{Compaction: CompactionThresholds{LooseFileCount: 4}})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write("file/one", []byte("one"))
	}); err != nil {
		t.Fatal(err)
	}
	count, _, err := looseObjectStats(path)
	if err != nil {
		t.Fatal(err)
	}
	if count <= 0 {
		t.Fatal("below-threshold write unexpectedly compacted all loose objects")
	}

	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write("file/two", []byte("two"))
	}); err != nil {
		t.Fatal(err)
	}
	count, _, err = looseObjectStats(path)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("above-threshold write left %d loose objects", count)
	}
}

func TestCompactionPrunesUnreachableLooseObjects(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ddocs")
	repository, err := InitWithOptions(path, Options{Compaction: CompactionThresholds{}})
	if err != nil {
		t.Fatal(err)
	}
	encoded := repository.store.NewEncodedObject()
	encoded.SetType(plumbing.BlobObject)
	encoded.SetSize(int64(len("unreachable")))
	writer, err := encoded.Writer()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("unreachable")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	orphan, err := repository.store.SetEncodedObject(encoded)
	if err != nil {
		t.Fatal(err)
	}
	repository.compaction = CompactionThresholds{LooseFileCount: 1}
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write("durable/value", []byte("kept"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := repository.store.HasEncodedObject(orphan); err == nil {
		t.Fatal("unreachable loose object survived compaction")
	}
	tx, err := repository.Begin()
	if err != nil {
		t.Fatal(err)
	}
	value, err := tx.Read("durable/value")
	if err != nil || string(value) != "kept" {
		t.Fatalf("reachable state after compaction = %q, %v", value, err)
	}
}

func TestCompactionFailureDoesNotFailPublishedWriteAndCanRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ddocs")
	repository, err := InitWithOptions(path, Options{Compaction: CompactionThresholds{LooseFileCount: 1}})
	if err != nil {
		t.Fatal(err)
	}
	original := repackRepositoryObjects
	called := false
	repackRepositoryObjects = func(*git.Repository) error {
		called = true
		return errors.New("injected maintenance failure")
	}
	defer func() { repackRepositoryObjects = original }()
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write("durable/value", []byte("published"))
	}); err != nil {
		t.Fatalf("logical write failed with maintenance: %v", err)
	}
	if !called {
		t.Fatal("compaction failure seam was not exercised")
	}
	tx, err := repository.Begin()
	if err != nil {
		t.Fatal(err)
	}
	value, err := tx.Read("durable/value")
	if err != nil || string(value) != "published" {
		t.Fatalf("published value after maintenance failure = %q, %v", value, err)
	}

	repackRepositoryObjects = original
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write("durable/second", []byte("retry"))
	}); err != nil {
		t.Fatalf("write after maintenance recovery failed: %v", err)
	}
	count, _, err := looseObjectStats(path)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("successful retry left %d loose objects", count)
	}
}

func TestCompactionTriggersOnBytesAndPreservesCurrentState(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".ddocs")
	repository, err := InitWithOptions(path, Options{Compaction: CompactionThresholds{LooseBytes: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Transaction(func(tx *Transaction) error {
		return tx.Write("durable/value", []byte("payload"))
	}); err != nil {
		t.Fatal(err)
	}
	count, _, err := looseObjectStats(path)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("byte-threshold compaction left %d loose objects", count)
	}

	reopened, err := OpenWithOptions(path, Options{Compaction: CompactionThresholds{}})
	if err != nil {
		t.Fatal(err)
	}
	tx, err := reopened.Begin()
	if err != nil {
		t.Fatal(err)
	}
	value, err := tx.Read("durable/value")
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != "payload" {
		t.Fatalf("compacted value = %q", value)
	}
}
