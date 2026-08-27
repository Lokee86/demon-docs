package codemaparcana

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestDiscoverSnapshotRequiresArcanaLexiconAlignment(t *testing.T) {
	root := t.TempDir()
	id := writeSnapshotFixture(t, root, map[string]string{"src/main.go": "package main\n"})
	state, availability := discoverSnapshot(root)
	if !availability.Available || availability.SnapshotID != id || state.id != id {
		t.Fatalf("unexpected availability: %#v state=%#v", availability, state)
	}
	if state.fileContentIDs["src/main.go"] == "" {
		t.Fatalf("missing Lexicon file content ID: %#v", state.fileContentIDs)
	}

	writeTestStateFile(t, filepath.Join(root, ".arcana", "CURRENT"), "sha256:"+repeatHex('f', 64)+"\n")
	_, availability = discoverSnapshot(root)
	if availability.Available || availability.Reason == "" {
		t.Fatalf("mismatched snapshots were accepted: %#v", availability)
	}
}

func TestPreparedGitStateCurrentRequiresPreparedHeadAndCleanWorktree(t *testing.T) {
	root := t.TempDir()
	writeTestStateFile(t, filepath.Join(root, "src", "main.go"), "package main\n")
	writeTestStateFile(t, filepath.Join(root, ".gitignore"), ".lexicon/\n.arcana/\n")
	repository, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	worktree, err := repository.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"src/main.go", ".gitignore"} {
		if _, err := worktree.Add(path); err != nil {
			t.Fatal(err)
		}
	}
	hash, err := worktree.Commit("initial", &git.CommitOptions{Author: &object.Signature{Name: "test", Email: "test@example.com", When: time.Unix(1, 0)}})
	if err != nil {
		t.Fatal(err)
	}
	lexiconRoot := filepath.Join(root, ".lexicon")
	marker, _ := json.Marshal(freshnessMarker{GitHead: hash.String()})
	writeTestStateFile(t, filepath.Join(lexiconRoot, ".repostate.json"), string(marker))
	if !preparedGitStateCurrent(root, lexiconRoot) {
		t.Fatal("clean prepared Git state was not accepted")
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n// dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if preparedGitStateCurrent(root, lexiconRoot) {
		t.Fatal("dirty worktree was accepted for global symbol resolution")
	}
}

func TestDiscoverSnapshotRejectsTamperedLexiconManifest(t *testing.T) {
	root := t.TempDir()
	id := writeSnapshotFixture(t, root, map[string]string{"src/main.go": "package main\n"})
	digest := id[len("sha256:"):]
	manifestPath := filepath.Join(root, ".lexicon", "snapshots", digest+".json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, ' ')
	data = append(data, []byte("tampered")...)
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, availability := discoverSnapshot(root)
	if availability.Available {
		t.Fatal("tampered Lexicon manifest was accepted")
	}
}

func writeSnapshotFixture(t *testing.T, root string, files map[string]string) string {
	t.Helper()
	entries := make([]lexiconFile, 0, len(files))
	for path, contents := range files {
		writeTestStateFile(t, filepath.Join(root, filepath.FromSlash(path)), contents)
		sum := sha256.Sum256([]byte(contents))
		entries = append(entries, lexiconFile{Path: path, ContentID: "sha256:" + hex.EncodeToString(sum[:])})
	}
	manifest := lexiconManifest{Version: 1, Languages: []lexiconLanguage{{Files: entries}}}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("lexicon:snapshot:v1\x00"))
	_, _ = hash.Write(data)
	id := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	digest := id[len("sha256:"):]
	writeTestStateFile(t, filepath.Join(root, ".lexicon", "CURRENT"), id+"\n")
	writeTestStateFile(t, filepath.Join(root, ".lexicon", "snapshots", digest+".json"), string(data)+"\n")
	writeTestStateFile(t, filepath.Join(root, ".arcana", "CURRENT"), id+"\n")
	snapshot := filepath.Join(root, ".arcana", "snapshots", digest)
	writeTestStateFile(t, filepath.Join(snapshot, "repository.manifest"), "version=1\n")
	writeTestStateFile(t, filepath.Join(snapshot, "lexicon.snapshot"), id+"\n")
	return id
}

func writeTestStateFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func repeatHex(char byte, count int) string {
	result := make([]byte, count)
	for index := range result {
		result[index] = char
	}
	return string(result)
}
