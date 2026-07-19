package ddrepo

import (
	"fmt"
	"io"
	"sort"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

func writeShard(store storage.Storer, records map[string][]byte) (plumbing.Hash, error) {
	data, err := encodeShard(records)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	return writeBlob(store, data)
}

func readShard(store storage.Storer, name string, hash plumbing.Hash) (map[string][]byte, error) {
	data, err := readBlob(store, hash)
	if err != nil {
		return nil, fmt.Errorf("read shard %s: %w", name, err)
	}
	records, err := decodeShard(data)
	if err != nil {
		return nil, fmt.Errorf("decode shard %s: %w", name, err)
	}
	return records, nil
}

func writeRoot(store storage.Storer, shards map[string]plumbing.Hash) (plumbing.Hash, error) {
	entries := make([]object.TreeEntry, 0, len(shards))
	for name, hash := range shards {
		if !isShardName(name) {
			return plumbing.ZeroHash, fmt.Errorf("invalid ddocs shard entry %q", name)
		}
		entries = append(entries, object.TreeEntry{
			Name: name,
			Mode: filemode.Regular,
			Hash: hash,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	encoded := store.NewEncodedObject()
	if err := (&object.Tree{Entries: entries}).Encode(encoded); err != nil {
		return plumbing.ZeroHash, err
	}
	return store.SetEncodedObject(encoded)
}

func writeBlob(store storage.Storer, data []byte) (plumbing.Hash, error) {
	encoded := store.NewEncodedObject()
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
	return store.SetEncodedObject(encoded)
}

func readBlob(store storage.Storer, hash plumbing.Hash) ([]byte, error) {
	encoded, err := store.EncodedObject(plumbing.BlobObject, hash)
	if err != nil {
		return nil, err
	}
	reader, err := encoded.Reader()
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}
