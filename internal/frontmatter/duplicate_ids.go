package frontmatter

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/Lokee86/demon-docs/internal/config"
)

func selectDuplicateOwners(sources []plannedSource, immutable immutableIndex) map[string]string {
	pathsByID := make(map[string][]string)
	for _, source := range sources {
		if source.parseErr != nil {
			continue
		}
		if id := sourceDocumentID(source); id != "" {
			pathsByID[id] = append(pathsByID[id], source.relative)
		}
	}

	owners := make(map[string]string)
	for id, paths := range pathsByID {
		if len(paths) < 2 {
			continue
		}
		sort.Strings(paths)
		owner := paths[0]
		if record, ok := immutable.byID[id]; ok && containsPath(paths, record.Path) {
			owner = filepath.ToSlash(record.Path)
		} else {
			for _, path := range paths {
				record, ok := immutable.byPath[path]
				if ok && documentID(record.Values) == id {
					owner = path
					break
				}
			}
		}
		owners[id] = owner
	}
	return owners
}

func collectDocumentIDs(sources []plannedSource) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, source := range sources {
		if source.parseErr != nil {
			continue
		}
		if id := sourceDocumentID(source); id != "" {
			ids[id] = struct{}{}
		}
	}
	return ids
}

func generateUniqueDocumentID(schema config.Frontmatter, now time.Time, used map[string]struct{}) (string, bool, error) {
	definition, ok := schema.Fields["document_id"]
	if !ok || !definition.Generated || normalizedType(definition.Type) != "uuid" {
		return "", false, nil
	}
	for attempt := 0; attempt < 16; attempt++ {
		value, available, err := replacementValue(definition, schema, nil, false, now)
		if err != nil {
			return "", false, err
		}
		if !available {
			return "", false, nil
		}
		id := documentID(map[string]any{"document_id": value})
		if id == "" {
			return "", false, fmt.Errorf("generated document_id is not a UUID")
		}
		if _, exists := used[id]; !exists {
			return id, true, nil
		}
	}
	return "", false, fmt.Errorf("could not generate a unique document_id after 16 attempts")
}

func containsPath(paths []string, candidate string) bool {
	candidate = filepath.ToSlash(candidate)
	for _, path := range paths {
		if filepath.ToSlash(path) == candidate {
			return true
		}
	}
	return false
}
