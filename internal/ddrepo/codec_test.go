package ddrepo

import (
	"bytes"
	"reflect"
	"testing"
)

func TestShardCodecIsDeterministicAndRoundTrips(t *testing.T) {
	records := map[string][]byte{
		"file/alpha":  []byte("one"),
		"source/beta": []byte("two"),
		"write/gamma": nil,
	}
	first, err := encodeShard(records)
	if err != nil {
		t.Fatal(err)
	}
	second, err := encodeShard(records)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("shard encoding is not deterministic")
	}
	decoded, err := decodeShard(first)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, records) {
		t.Fatalf("decoded records = %#v, want %#v", decoded, records)
	}
}

func TestShardCodecRejectsInvalidData(t *testing.T) {
	for _, name := range []string{"", "/absolute", `back\\slash`, "empty//segment", "dot/./segment", "parent/../segment"} {
		if _, err := encodeShard(map[string][]byte{name: nil}); err == nil {
			t.Fatalf("invalid name %q was accepted", name)
		}
	}
	valid, err := encodeShard(map[string][]byte{"file/alpha": []byte("value")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeShard(append(valid, 0)); err == nil {
		t.Fatal("trailing shard data was accepted")
	}
	if _, err := decodeShard(valid[:len(valid)-1]); err == nil {
		t.Fatal("truncated shard data was accepted")
	}
}
