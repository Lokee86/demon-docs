package links

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/model"
	"github.com/Lokee86/demon-docs/internal/textio"
)

type replacement struct {
	start, end int
	value      string
}

func Reconcile(repositoryRoot string) (Plan, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return Plan{}, err
	}
	previousFiles, previousLinks, initialized, err := loadState(root)
	if err != nil {
		return Plan{}, err
	}
	inventory, err := buildInventory(root, previousFiles)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{
		RepositoryRoot:      filepath.Clean(root),
		Initialized:         initialized,
		NeedsInitialization: !initialized,
		Files:               inventory.manifest,
		Links:               LinksManifest{SchemaVersion: schemaVersion},
	}
	if !initialized {
		plan.Messages = append(plan.Messages, "Link state is not initialized; this pass records a baseline and does not repair links.")
	}
	previousBySource := previousLinkIndex(previousLinks)
	sources := markdownSources(inventory)
	for _, source := range sources {
		document, err := textio.Read(source.path)
		if err != nil {
			return Plan{}, fmt.Errorf("read Markdown source %s: %w", source.path, err)
		}
		occurrences := parseMarkdownLinks(document.Text)
		var replacements []replacement
		ordinal := 0
		for _, found := range occurrences {
			resolved, style, local := resolveLocalTarget(found.RawPath, source.path, found.Angle)
			if !local {
				continue
			}
			ignored, err := inventory.ignored(resolved)
			if err != nil {
				return Plan{}, fmt.Errorf("evaluate link target ignore policy %s: %w", resolved, err)
			}
			if ignored {
				continue
			}
			originalTarget := found.RawPath + found.Suffix
			previous := findPreviousLink(previousBySource[source.record.ID], ordinal, originalTarget)
			record := LinkRecord{
				SourceFileID: source.record.ID,
				SourcePath:   source.record.Path,
				Ordinal:      ordinal,
				Line:         found.Line,
				Column:       found.Column,
				Syntax:       found.Syntax,
				Target:       originalTarget,
			}
			ordinal++
			targetRecord, actualPath := inventory.exact(resolved)
			if targetRecord == nil {
				if _, statErr := os.Stat(resolved); statErr == nil {
					targetRecord, actualPath, err = inventory.ensureTarget(resolved, "")
					if err != nil {
						return Plan{}, fmt.Errorf("record link target %s: %w", resolved, err)
					}
				}
			}
			if targetRecord != nil {
				record.TargetFileID = targetRecord.ID
				record.ResolvedPath = storePath(root, actualPath)
				record.Status = "valid"
				if filepath.Clean(actualPath) != filepath.Clean(resolved) {
					record.Status = "case_mismatch"
					if initialized {
						newPath := renderTargetPath(style, found.RawPath, source.path, actualPath)
						replacements = append(replacements, replacement{found.Start, found.End, newPath})
						record.Target = newPath + found.Suffix
						plan.Messages = append(plan.Messages, fmt.Sprintf("Updated link case in %s:%d: %s -> %s", source.record.Path, found.Line, found.RawPath, newPath))
					}
				}
				plan.Links.Links = append(plan.Links.Links, record)
				continue
			}

			preferredID := ""
			if previous != nil {
				preferredID = previous.TargetFileID
			}
			var candidates []string
			if initialized && preferredID != "" {
				if _, moved := inventory.byID(preferredID); moved != "" {
					candidates = []string{moved}
				}
			}
			if initialized && len(candidates) == 0 {
				candidates = candidatePaths(inventory, resolved, preferredID)
			}
			record.Candidates = displayPaths(root, candidates)
			switch len(candidates) {
			case 0:
				record.Status = "broken"
				plan.Unresolved++
				plan.Messages = append(plan.Messages, fmt.Sprintf("Broken link in %s:%d:%d: %s", source.record.Path, found.Line, found.Column, originalTarget))
			case 1:
				candidate := candidates[0]
				targetRecord, actualPath, err = inventory.ensureTarget(candidate, preferredID)
				if err != nil {
					return Plan{}, fmt.Errorf("record moved link target %s: %w", candidate, err)
				}
				record.TargetFileID = targetRecord.ID
				record.ResolvedPath = storePath(root, actualPath)
				record.Status = "moved"
				newPath := renderTargetPath(style, found.RawPath, source.path, actualPath)
				replacements = append(replacements, replacement{found.Start, found.End, newPath})
				record.Target = newPath + found.Suffix
				plan.Messages = append(plan.Messages, fmt.Sprintf("Repair link in %s:%d: %s -> %s", source.record.Path, found.Line, found.RawPath, newPath))
			default:
				record.Status = "ambiguous"
				plan.Unresolved++
				plan.Messages = append(plan.Messages, fmt.Sprintf("Ambiguous link in %s:%d:%d: %s; candidates: %s", source.record.Path, found.Line, found.Column, originalTarget, strings.Join(record.Candidates, ", ")))
			}
			plan.Links.Links = append(plan.Links.Links, record)
		}
		if initialized && len(replacements) > 0 {
			updated := applyReplacements(document.Text, replacements)
			if updated != document.Text {
				old := document.Text
				plan.Updates = append(plan.Updates, model.FileUpdate{Path: source.path, OldText: &old, NewText: updated})
			}
		}
	}
	plan.Files = inventory.manifest
	sortManifests(&plan.Files, &plan.Links)
	return plan, nil
}

type markdownSource struct {
	path   string
	record *FileRecord
}

func markdownSources(inventory *inventory) []markdownSource {
	var result []markdownSource
	for index := range inventory.manifest.Files {
		record := &inventory.manifest.Files[index]
		if record.Scope != "repository" || !record.Present || record.Kind != "file" || !isMarkdown(record.Path) {
			continue
		}
		result = append(result, markdownSource{path: recordAbsolute(inventory.root, *record), record: record})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].record.Path < result[j].record.Path })
	return result
}

func isMarkdown(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown", ".mkd", ".mdx":
		return true
	default:
		return false
	}
}

func previousLinkIndex(manifest LinksManifest) map[string][]LinkRecord {
	result := map[string][]LinkRecord{}
	for _, record := range manifest.Links {
		result[record.SourceFileID] = append(result[record.SourceFileID], record)
	}
	return result
}

func findPreviousLink(records []LinkRecord, ordinal int, target string) *LinkRecord {
	for index := range records {
		if records[index].Ordinal == ordinal && records[index].Target == target {
			return &records[index]
		}
	}
	var match *LinkRecord
	for index := range records {
		if records[index].Target == target {
			if match != nil {
				return nil
			}
			match = &records[index]
		}
	}
	return match
}

func candidatePaths(inventory *inventory, missingPath, preferredID string) []string {
	base := filepath.Base(missingPath)
	kind := "file"
	fingerprint := ""
	if preferred := inventory.recordByID(preferredID); preferred != nil {
		kind = preferred.Kind
		fingerprint = preferred.Fingerprint
	}
	candidates := inventory.candidates(base, kind)
	if fingerprint != "" {
		var exact []string
		for _, candidate := range candidates {
			if record, _ := inventory.exact(candidate); record != nil && record.Fingerprint == fingerprint {
				exact = append(exact, candidate)
			}
		}
		if len(exact) > 0 {
			candidates = exact
		}
	}
	if !strings.EqualFold(filepath.Dir(missingPath), inventory.root) {
		candidates = append(candidates, discoverExternalCandidates(missingPath, base, kind, fingerprint)...)
	}
	return uniquePaths(candidates)
}

func displayPaths(root string, paths []string) []string {
	result := make([]string, len(paths))
	for index, path := range paths {
		result[index] = storePath(root, path)
	}
	sort.Strings(result)
	return result
}

func applyReplacements(source string, replacements []replacement) string {
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].start > replacements[j].start })
	result := source
	for _, replacement := range replacements {
		result = result[:replacement.start] + replacement.value + result[replacement.end:]
	}
	return result
}
