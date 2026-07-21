package reconcile

import (
	"fmt"
	"os"

	"github.com/Lokee86/demon-docs/internal/config"
	md "github.com/Lokee86/demon-docs/internal/markdown"
	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/textio"
	"github.com/Lokee86/demon-docs/internal/validationworkers"
)

type indexSourceResult struct {
	text    string
	exists  bool
	entries []*model.IndexEntry
}

func loadIndexSources(folders []*model.FolderInfo, c config.Config) (map[string]string, map[string]bool, map[string][]*model.IndexEntry, error) {
	results := make([]indexSourceResult, len(folders))
	errors := validationworkers.Run(len(folders), func(index int) error {
		folder := folders[index]
		if folder.IndexPath == "" {
			return nil
		}
		doc, err := textio.Read(folder.IndexPath)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read index %s: %w", folder.IndexPath, err)
		}
		results[index] = indexSourceResult{
			text:    doc.Text,
			exists:  true,
			entries: md.ParseEntries(folder.IndexPath, doc.Text, c),
		}
		return nil
	})

	texts := map[string]string{}
	exists := map[string]bool{}
	entries := map[string][]*model.IndexEntry{}
	for index, err := range errors {
		if err != nil {
			return nil, nil, nil, err
		}
		if !results[index].exists {
			continue
		}
		folder := folders[index]
		texts[folder.Path] = results[index].text
		exists[folder.Path] = true
		entries[folder.Path] = results[index].entries
	}
	return texts, exists, entries, nil
}

type editableSourceResult struct {
	path   string
	text   string
	exists bool
}

func loadEditableSources(folders []*model.FolderInfo, c config.Config) (map[string]string, error) {
	paths := editableSourcePaths(folders, c)
	results := make([]editableSourceResult, len(paths))
	errors := validationworkers.Run(len(paths), func(index int) error {
		path := paths[index]
		doc, err := textio.Read(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read indexed file %s: %w", path, err)
		}
		results[index] = editableSourceResult{path: path, text: doc.Text, exists: true}
		return nil
	})

	texts := map[string]string{}
	for index, err := range errors {
		if err != nil {
			return nil, err
		}
		if results[index].exists {
			texts[results[index].path] = results[index].text
		}
	}
	return texts, nil
}

func editableSourcePaths(folders []*model.FolderInfo, c config.Config) []string {
	paths := []string{}
	for _, folder := range folders {
		for _, path := range append(append([]string{}, folder.DirectFiles...), folder.StubFiles...) {
			if config.IsParentEditable(path, c) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}
