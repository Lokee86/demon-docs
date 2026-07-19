package links

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/ddrepo"
)

const (
	filesStateName = "files.json"
	linksStateName = "links.json"
	metaRecordName = "meta/state"
)

type stateMetadata struct {
	SchemaVersion int `json:"schema_version"`
}

type sourceRecord struct {
	SchemaVersion int          `json:"schema_version"`
	SourceFileID  string       `json:"source_file_id"`
	SourcePath    string       `json:"source_path"`
	Links         []LinkRecord `json:"links"`
}

type incomingRecord struct {
	SchemaVersion int      `json:"schema_version"`
	TargetFileID  string   `json:"target_file_id"`
	SourceFileID  string   `json:"source_file_id"`
	LinkIDs       []string `json:"link_ids"`
}

type pathRecord struct {
	SchemaVersion int    `json:"schema_version"`
	Scope         string `json:"scope"`
	Path          string `json:"path"`
	FileID        string `json:"file_id"`
}

func statePaths(repositoryRoot string) (string, string) {
	stateRoot := filepath.Join(repositoryRoot, ".ddocs")
	return filepath.Join(stateRoot, filesStateName), filepath.Join(stateRoot, linksStateName)
}

func loadState(repositoryRoot string) (FilesManifest, LinksManifest, bool, error) {
	files := FilesManifest{SchemaVersion: schemaVersion}
	links := LinksManifest{SchemaVersion: schemaVersion}
	repository, err := ddrepo.Open(filepath.Join(repositoryRoot, ".ddocs"))
	if err != nil {
		legacyFiles, legacyLinks, initialized, legacyErr := loadLegacyState(repositoryRoot)
		if legacyErr != nil {
			return files, links, false, legacyErr
		}
		return legacyFiles, legacyLinks, initialized, nil
	}
	tx, err := repository.Begin()
	if err != nil {
		return files, links, false, fmt.Errorf("begin ddocs state read: %w", err)
	}
	metadataData, err := tx.Read(metaRecordName)
	if errors.Is(err, ddrepo.ErrRecordAbsent) {
		legacyFiles, legacyLinks, initialized, legacyErr := loadLegacyState(repositoryRoot)
		if legacyErr != nil {
			return files, links, false, legacyErr
		}
		return legacyFiles, legacyLinks, initialized, nil
	}
	if err != nil {
		return files, links, false, fmt.Errorf("read ddocs state metadata: %w", err)
	}
	var metadata stateMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return files, links, false, fmt.Errorf("decode ddocs state metadata: %w", err)
	}
	if metadata.SchemaVersion != schemaVersion {
		return files, links, false, fmt.Errorf("unsupported link state schema %d", metadata.SchemaVersion)
	}
	fileNames, err := tx.Names("file/")
	if err != nil {
		return files, links, false, fmt.Errorf("list file identity records: %w", err)
	}
	for _, name := range fileNames {
		data, err := tx.Read(name)
		if err != nil {
			return files, links, false, fmt.Errorf("read %s: %w", name, err)
		}
		var record FileRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return files, links, false, fmt.Errorf("decode %s: %w", name, err)
		}
		files.Files = append(files.Files, record)
	}
	sourceNames, err := tx.Names("source/")
	if err != nil {
		return files, links, false, fmt.Errorf("list Markdown source records: %w", err)
	}
	for _, name := range sourceNames {
		data, err := tx.Read(name)
		if err != nil {
			return files, links, false, fmt.Errorf("read %s: %w", name, err)
		}
		var source sourceRecord
		if err := json.Unmarshal(data, &source); err != nil {
			return files, links, false, fmt.Errorf("decode %s: %w", name, err)
		}
		if source.SchemaVersion != schemaVersion {
			return files, links, false, fmt.Errorf("unsupported source record schema in %s", name)
		}
		links.Links = append(links.Links, source.Links...)
	}
	sortManifests(&files, &links)
	return files, links, true, nil
}

func Save(plan Plan) error {
	repository, err := openOrInitializeStateRepository(plan.RepositoryRoot)
	if err != nil {
		return err
	}
	sortManifests(&plan.Files, &plan.Links)
	desired := make(map[string][]byte)
	if err := addJSONRecord(desired, metaRecordName, stateMetadata{SchemaVersion: schemaVersion}); err != nil {
		return err
	}
	for _, record := range plan.Files.Files {
		if record.ID == "" {
			return fmt.Errorf("file identity has no ID: %s", record.Path)
		}
		if err := addJSONRecord(desired, "file/"+record.ID, record); err != nil {
			return err
		}
		path := pathRecord{SchemaVersion: schemaVersion, Scope: record.Scope, Path: record.Path, FileID: record.ID}
		if err := addJSONRecord(desired, "path/"+pathRecordKey(record.Scope, record.Path), path); err != nil {
			return err
		}
	}
	bySource := make(map[string][]LinkRecord)
	incoming := make(map[string]map[string][]string)
	for _, record := range plan.Links.Links {
		bySource[record.SourceFileID] = append(bySource[record.SourceFileID], record)
		if record.TargetFileID != "" && record.ID != "" {
			if incoming[record.TargetFileID] == nil {
				incoming[record.TargetFileID] = make(map[string][]string)
			}
			incoming[record.TargetFileID][record.SourceFileID] = append(incoming[record.TargetFileID][record.SourceFileID], record.ID)
		}
	}
	for sourceID, records := range bySource {
		sort.Slice(records, func(i, j int) bool { return records[i].Ordinal < records[j].Ordinal })
		source := sourceRecord{SchemaVersion: schemaVersion, SourceFileID: sourceID, SourcePath: records[0].SourcePath, Links: records}
		if err := addJSONRecord(desired, "source/"+sourceID, source); err != nil {
			return err
		}
	}
	for targetID, sources := range incoming {
		for sourceID, linkIDs := range sources {
			sort.Strings(linkIDs)
			record := incomingRecord{SchemaVersion: schemaVersion, TargetFileID: targetID, SourceFileID: sourceID, LinkIDs: linkIDs}
			if err := addJSONRecord(desired, "incoming/"+targetID+"/"+sourceID, record); err != nil {
				return err
			}
		}
	}
	for _, suppression := range plan.Suppressions {
		if suppression.SourceFileID == "" {
			return fmt.Errorf("suppression has no source file ID: %s", suppression.Path)
		}
		if err := addJSONRecord(desired, "write/"+suppression.SourceFileID, suppression); err != nil {
			return err
		}
	}
	err = repository.Transaction(func(tx *ddrepo.Transaction) error {
		for _, prefix := range []string{"file/", "path/", "source/", "incoming/", "write/", "meta/"} {
			names, err := tx.Names(prefix)
			if err != nil {
				return err
			}
			for _, name := range names {
				if _, keep := desired[name]; !keep {
					if err := tx.Delete(name); err != nil {
						return err
					}
				}
			}
		}
		names := make([]string, 0, len(desired))
		for name := range desired {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if err := tx.Write(name, desired[name]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("publish ddocs state: %w", err)
	}
	filesPath, linksPath := statePaths(plan.RepositoryRoot)
	if err := os.Remove(filesPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy file state: %w", err)
	}
	if err := os.Remove(linksPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy link state: %w", err)
	}
	return nil
}

func LoadPendingSuppressions(repositoryRoot string) ([]Suppression, error) {
	repository, err := ddrepo.Open(filepath.Join(repositoryRoot, ".ddocs"))
	if err != nil {
		return nil, nil
	}
	tx, err := repository.Begin()
	if err != nil {
		return nil, err
	}
	names, err := tx.Names("write/")
	if err != nil {
		return nil, err
	}
	result := make([]Suppression, 0, len(names))
	for _, name := range names {
		data, err := tx.Read(name)
		if err != nil {
			return nil, err
		}
		var record Suppression
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, fmt.Errorf("decode %s: %w", name, err)
		}
		result = append(result, record)
	}
	return result, nil
}

func DeletePendingSuppression(repositoryRoot, sourceFileID string) error {
	repository, err := ddrepo.Open(filepath.Join(repositoryRoot, ".ddocs"))
	if err != nil {
		return nil
	}
	return repository.Transaction(func(tx *ddrepo.Transaction) error {
		return tx.Delete("write/" + sourceFileID)
	})
}

func openOrInitializeStateRepository(repositoryRoot string) (*ddrepo.Repository, error) {
	path := filepath.Join(repositoryRoot, ".ddocs")
	repository, err := ddrepo.Open(path)
	if err == nil {
		return repository, nil
	}
	if _, statErr := os.Stat(filepath.Join(path, "objects")); statErr == nil {
		return nil, err
	}
	repository, err = ddrepo.Init(path)
	if err != nil {
		return nil, fmt.Errorf("initialize ddocs state repository: %w", err)
	}
	return repository, nil
}

func loadLegacyState(repositoryRoot string) (FilesManifest, LinksManifest, bool, error) {
	filesPath, linksPath := statePaths(repositoryRoot)
	files := FilesManifest{SchemaVersion: schemaVersion}
	links := LinksManifest{SchemaVersion: schemaVersion}
	filesData, filesErr := os.ReadFile(filesPath)
	linksData, linksErr := os.ReadFile(linksPath)
	if errors.Is(filesErr, os.ErrNotExist) && errors.Is(linksErr, os.ErrNotExist) {
		return files, links, false, nil
	}
	if filesErr != nil {
		return files, links, false, fmt.Errorf("read legacy link file state: %w", filesErr)
	}
	if linksErr != nil {
		return files, links, false, fmt.Errorf("read legacy link graph state: %w", linksErr)
	}
	type legacyFilesManifest struct {
		SchemaVersion int          `json:"schema_version"`
		Files         []FileRecord `json:"files"`
	}
	type legacyLinksManifest struct {
		SchemaVersion int          `json:"schema_version"`
		Links         []LinkRecord `json:"links"`
	}
	var legacyFiles legacyFilesManifest
	var legacyLinks legacyLinksManifest
	if err := json.Unmarshal(filesData, &legacyFiles); err != nil {
		return files, links, false, fmt.Errorf("decode %s: %w", filesPath, err)
	}
	if err := json.Unmarshal(linksData, &legacyLinks); err != nil {
		return files, links, false, fmt.Errorf("decode %s: %w", linksPath, err)
	}
	files.Files = legacyFiles.Files
	links.Links = legacyLinks.Links
	return files, links, true, nil
}

func addJSONRecord(records map[string][]byte, name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	records[name] = data
	return nil
}

func pathRecordKey(scope, path string) string {
	digest := sha256.Sum256([]byte(scope + "\x00" + path))
	return hex.EncodeToString(digest[:])
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
