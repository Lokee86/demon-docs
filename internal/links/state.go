package links

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	filesStateName = "files.json"
	linksStateName = "links.json"
)

func statePaths(repositoryRoot string) (string, string) {
	stateRoot := filepath.Join(repositoryRoot, ".ddocs")
	return filepath.Join(stateRoot, filesStateName), filepath.Join(stateRoot, linksStateName)
}

func loadState(repositoryRoot string) (FilesManifest, LinksManifest, bool, error) {
	filesPath, linksPath := statePaths(repositoryRoot)
	files := FilesManifest{SchemaVersion: schemaVersion}
	links := LinksManifest{SchemaVersion: schemaVersion}
	filesData, filesErr := os.ReadFile(filesPath)
	linksData, linksErr := os.ReadFile(linksPath)
	if errors.Is(filesErr, os.ErrNotExist) && errors.Is(linksErr, os.ErrNotExist) {
		return files, links, false, nil
	}
	if filesErr != nil {
		return files, links, false, fmt.Errorf("read link file state: %w", filesErr)
	}
	if linksErr != nil {
		return files, links, false, fmt.Errorf("read link graph state: %w", linksErr)
	}
	if err := json.Unmarshal(filesData, &files); err != nil {
		return files, links, false, fmt.Errorf("decode %s: %w", filesPath, err)
	}
	if err := json.Unmarshal(linksData, &links); err != nil {
		return files, links, false, fmt.Errorf("decode %s: %w", linksPath, err)
	}
	if files.SchemaVersion != schemaVersion || links.SchemaVersion != schemaVersion {
		return files, links, false, fmt.Errorf("unsupported link state schema")
	}
	return files, links, true, nil
}

func Save(plan Plan) error {
	filesPath, linksPath := statePaths(plan.RepositoryRoot)
	if err := os.MkdirAll(filepath.Dir(filesPath), 0o755); err != nil {
		return fmt.Errorf("create link state directory: %w", err)
	}
	sortManifests(&plan.Files, &plan.Links)
	if err := writeJSON(filesPath, plan.Files); err != nil {
		return err
	}
	if err := writeJSON(linksPath, plan.Links); err != nil {
		return err
	}
	return nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func sortManifests(files *FilesManifest, links *LinksManifest) {
	sort.Slice(files.Files, func(i, j int) bool {
		if files.Files[i].Scope != files.Files[j].Scope {
			return files.Files[i].Scope < files.Files[j].Scope
		}
		return files.Files[i].Path < files.Files[j].Path
	})
	for i := range files.Files {
		sort.Strings(files.Files[i].PathHistory)
	}
	sort.Slice(links.Links, func(i, j int) bool {
		left, right := links.Links[i], links.Links[j]
		if left.SourcePath != right.SourcePath {
			return left.SourcePath < right.SourcePath
		}
		return left.Ordinal < right.Ordinal
	})
}

func newFileID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	raw[6] = raw[6]&0x0f | 0x40
	raw[8] = raw[8]&0x3f | 0x80
	encoded := hex.EncodeToString(raw[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32], nil
}
