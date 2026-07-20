package frontmatter

import (
	"fmt"
	"path/filepath"

	"github.com/Lokee86/demon-docs/internal/textio"
)

type plannedSource struct {
	path     string
	relative string
	document textio.Document
	parsed   Document
	parseErr error
}

func loadSources(repoRoot string, files []string, allowedFormats []string) ([]plannedSource, map[string]bool, error) {
	sources := make([]plannedSource, 0, len(files))
	pathsByID := make(map[string][]string)
	for _, path := range files {
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return nil, nil, err
		}
		document, err := textio.Read(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read frontmatter source %s: %w", path, err)
		}
		parsed, parseErr := Parse(document.Text, allowedFormats)
		source := plannedSource{
			path:     path,
			relative: filepath.ToSlash(relative),
			document: document,
			parsed:   parsed,
			parseErr: parseErr,
		}
		sources = append(sources, source)
		if parseErr == nil {
			if id := documentID(parsed.Values); id != "" {
				pathsByID[id] = append(pathsByID[id], path)
			}
		}
	}

	duplicates := make(map[string]bool)
	for _, paths := range pathsByID {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			duplicates[path] = true
		}
	}
	return sources, duplicates, nil
}
