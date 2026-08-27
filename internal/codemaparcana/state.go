package codemaparcana

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
)

type snapshotState struct {
	id             string
	directory      string
	fileContentIDs map[string]string
	globalCurrent  bool
}

type Availability struct {
	Available  bool
	SnapshotID string
	Reason     string
}

type lexiconManifest struct {
	Version   int               `json:"version"`
	Languages []lexiconLanguage `json:"languages"`
}

type lexiconLanguage struct {
	Files []lexiconFile `json:"files"`
}

type lexiconFile struct {
	Path      string `json:"path"`
	ContentID string `json:"content_id"`
}

type freshnessMarker struct {
	GitHead string `json:"git_head"`
}

func discoverSnapshot(repositoryRoot string) (snapshotState, Availability) {
	arcanaRoot := filepath.Join(repositoryRoot, ".arcana")
	lexiconRoot := filepath.Join(repositoryRoot, ".lexicon")
	arcanaID, err := readCurrent(filepath.Join(arcanaRoot, "CURRENT"))
	if err != nil {
		return unavailable("Arcana CURRENT is unavailable: " + err.Error())
	}
	lexiconID, err := readCurrent(filepath.Join(lexiconRoot, "CURRENT"))
	if err != nil {
		return unavailable("Lexicon CURRENT is unavailable: " + err.Error())
	}
	if !validSnapshotID(arcanaID) || !validSnapshotID(lexiconID) {
		return unavailable("Arcana or Lexicon CURRENT has an invalid snapshot ID")
	}
	if arcanaID != lexiconID {
		return unavailable("Arcana CURRENT does not match Lexicon CURRENT")
	}
	digest := strings.TrimPrefix(arcanaID, "sha256:")
	directory := filepath.Join(arcanaRoot, "snapshots", digest)
	for _, name := range []string{"repository.manifest", "lexicon.snapshot"} {
		if info, statErr := os.Stat(filepath.Join(directory, name)); statErr != nil || info.IsDir() {
			return unavailable("Arcana snapshot is incomplete")
		}
	}
	boundID, err := readCurrent(filepath.Join(directory, "lexicon.snapshot"))
	if err != nil || boundID != arcanaID {
		return unavailable("Arcana snapshot is not bound to CURRENT Lexicon state")
	}
	manifestPath := filepath.Join(lexiconRoot, "snapshots", digest+".json")
	manifest, err := loadLexiconManifest(manifestPath, lexiconID)
	if err != nil {
		return unavailable("Lexicon snapshot is unavailable or invalid: " + err.Error())
	}
	files, err := manifestFileContentIDs(manifest)
	if err != nil {
		return unavailable("Lexicon snapshot file inventory is invalid: " + err.Error())
	}
	state := snapshotState{
		id:             arcanaID,
		directory:      directory,
		fileContentIDs: files,
		globalCurrent:  preparedGitStateCurrent(repositoryRoot, lexiconRoot),
	}
	return state, Availability{Available: true, SnapshotID: arcanaID}
}

func loadLexiconManifest(path, expectedID string) (lexiconManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return lexiconManifest{}, err
	}
	canonical := bytes.TrimSpace(data)
	hash := sha256.New()
	_, _ = hash.Write([]byte("lexicon:snapshot:v1\x00"))
	_, _ = hash.Write(canonical)
	actual := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	if actual != expectedID {
		return lexiconManifest{}, fmt.Errorf("snapshot content ID is %s; expected %s", actual, expectedID)
	}
	var manifest lexiconManifest
	if err := json.Unmarshal(canonical, &manifest); err != nil {
		return lexiconManifest{}, err
	}
	if manifest.Version != 1 {
		return lexiconManifest{}, fmt.Errorf("unsupported manifest version %d", manifest.Version)
	}
	return manifest, nil
}

func manifestFileContentIDs(manifest lexiconManifest) (map[string]string, error) {
	files := map[string]string{}
	for _, language := range manifest.Languages {
		for _, item := range language.Files {
			path := normalizeRepositoryPath(item.Path)
			if path == "" || !validContentID(item.ContentID) {
				return nil, fmt.Errorf("invalid file record %q", item.Path)
			}
			if existing, ok := files[path]; ok && existing != item.ContentID {
				return nil, fmt.Errorf("conflicting content IDs for %s", path)
			}
			files[path] = item.ContentID
		}
	}
	return files, nil
}

func preparedGitStateCurrent(repositoryRoot, lexiconRoot string) bool {
	var marker freshnessMarker
	data, err := os.ReadFile(filepath.Join(lexiconRoot, ".repostate.json"))
	if err != nil || json.Unmarshal(data, &marker) != nil || marker.GitHead == "" {
		return false
	}
	repository, err := git.PlainOpenWithOptions(repositoryRoot, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return false
	}
	head, err := repository.Head()
	if err != nil || head.Hash().String() != marker.GitHead {
		return false
	}
	worktree, err := repository.Worktree()
	if err != nil {
		return false
	}
	status, err := worktree.Status()
	return err == nil && status.IsClean()
}

func readCurrent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("%s is empty", path)
	}
	return value, nil
}

func unavailable(reason string) (snapshotState, Availability) {
	return snapshotState{}, Availability{Reason: reason}
}

func validSnapshotID(value string) bool {
	digest, ok := strings.CutPrefix(value, "sha256:")
	return ok && len(digest) == 64 && lowercaseHex(digest)
}

func validContentID(value string) bool {
	digest, ok := strings.CutPrefix(value, "sha256:")
	return ok && len(digest) == 64 && lowercaseHex(digest)
}

func lowercaseHex(value string) bool {
	for _, char := range value {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return false
		}
	}
	return true
}

func normalizeRepositoryPath(value string) string {
	value = filepath.ToSlash(strings.TrimSpace(value))
	if value == "" || filepath.IsAbs(filepath.FromSlash(value)) {
		return ""
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	return clean
}
