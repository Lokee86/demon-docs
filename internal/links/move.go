package links

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/Lokee86/demon-docs/internal/repository"
	"github.com/Lokee86/demon-docs/internal/textio"
)

// MoveUpdate describes one Markdown source that must be rewritten after a move.
type MoveUpdate struct {
	Path  string
	Links int
}

// MovePlan is a complete stateless filesystem move and link-rewrite plan.
type MovePlan struct {
	RepositoryRoot    string
	Source            string
	Destination       string
	SourceIsDirectory bool
	Updates           []MoveUpdate
	RewrittenLinks    int

	rewrites []plannedMoveRewrite
}

type plannedMoveRewrite struct {
	originPath string
	mode       fs.FileMode
	rewrite    GeneratedRewrite
}

// PlanMove scans repository Markdown, resolves links before the move, and
// plans only the destination-path changes required to preserve those links.
func PlanMove(repositoryRoot, source, destination string) (MovePlan, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return MovePlan{}, err
	}
	root = filepath.Clean(root)
	source, err = absoluteFrom(repositoryRoot, source)
	if err != nil {
		return MovePlan{}, err
	}
	destination, err = absoluteFrom(repositoryRoot, destination)
	if err != nil {
		return MovePlan{}, err
	}
	if !repository.Contains(root, source) {
		return MovePlan{}, fmt.Errorf("move source must be inside repository root: %s", source)
	}
	if !repository.Contains(root, destination) {
		return MovePlan{}, fmt.Errorf("move destination must be inside repository root: %s", destination)
	}

	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return MovePlan{}, fmt.Errorf("stat move source %s: %w", source, err)
	}
	if sourceInfo.Mode()&os.ModeSymlink != 0 {
		return MovePlan{}, fmt.Errorf("symbolic-link move sources are not supported: %s", source)
	}
	if filepath.Clean(source) == root {
		return MovePlan{}, errors.New("repository root cannot be moved")
	}

	destination, err = finalMoveDestination(source, destination)
	if err != nil {
		return MovePlan{}, err
	}
	if filepath.Clean(source) == filepath.Clean(destination) {
		return MovePlan{}, errors.New("move source and destination are the same path")
	}
	if sourceInfo.IsDir() && repository.Contains(source, destination) {
		return MovePlan{}, fmt.Errorf("move destination cannot be inside source directory: %s", destination)
	}
	parentInfo, err := os.Stat(filepath.Dir(destination))
	if err != nil {
		return MovePlan{}, fmt.Errorf("move destination parent does not exist: %s", filepath.Dir(destination))
	}
	if !parentInfo.IsDir() {
		return MovePlan{}, fmt.Errorf("move destination parent is not a directory: %s", filepath.Dir(destination))
	}
	if err := validateMoveContainment(root, source, destination); err != nil {
		return MovePlan{}, err
	}

	inventory, err := buildInventory(root, FilesManifest{})
	if err != nil {
		return MovePlan{}, err
	}
	for _, check := range []struct {
		label string
		path  string
	}{{"source", source}, {"destination", destination}} {
		ignored, err := inventory.ignored(check.path)
		if err != nil {
			return MovePlan{}, fmt.Errorf("evaluate move %s ignore policy %s: %w", check.label, check.path, err)
		}
		if ignored {
			return MovePlan{}, fmt.Errorf("move %s is excluded by repository ignore policy: %s", check.label, check.path)
		}
	}
	plan := MovePlan{
		RepositoryRoot:    root,
		Source:            source,
		Destination:       destination,
		SourceIsDirectory: sourceInfo.IsDir(),
	}

	for _, markdownSource := range markdownSources(inventory) {
		document, err := textio.Read(markdownSource.path)
		if err != nil {
			return MovePlan{}, fmt.Errorf("read Markdown move source %s: %w", markdownSource.path, err)
		}
		finalSourcePath := remapMovedPath(markdownSource.path, source, destination)
		var transformations []LinkTransformation
		for ordinal, found := range parseMarkdownLinks(document.Text) {
			resolved, style, local := resolveLocalTarget(found.RawPath, markdownSource.path, found.Angle)
			if !local {
				continue
			}
			ignored, err := inventory.ignored(resolved)
			if err != nil {
				return MovePlan{}, fmt.Errorf("evaluate move target ignore policy %s: %w", resolved, err)
			}
			if ignored {
				continue
			}
			actualTarget, err := resolveMoveTarget(inventory, resolved, found.Syntax, source)
			if err != nil {
				return MovePlan{}, fmt.Errorf("resolve affected link in %s:%d:%d: %w", markdownSource.record.Path, found.Line, found.Column, err)
			}
			if actualTarget == "" {
				continue
			}
			if found.RawPath == "" && filepath.Clean(actualTarget) == filepath.Clean(markdownSource.path) {
				continue
			}
			finalTargetPath := remapMovedPath(actualTarget, source, destination)
			if pathKey(finalSourcePath) == pathKey(markdownSource.path) && pathKey(finalTargetPath) == pathKey(actualTarget) && filepath.Clean(finalSourcePath) == filepath.Clean(markdownSource.path) && filepath.Clean(finalTargetPath) == filepath.Clean(actualTarget) {
				continue
			}
			newPath := renderMoveTarget(inventory, found.Syntax, found.RawPath, style, finalSourcePath, finalTargetPath, source, destination)
			if newPath == found.RawPath {
				continue
			}
			transformations = append(transformations, LinkTransformation{
				LinkID:         fmt.Sprintf("move:%s:%d", filepath.ToSlash(markdownSource.record.Path), ordinal),
				Start:          found.Start,
				End:            found.End,
				OldDestination: found.RawPath,
				NewDestination: newPath,
			})
		}
		if len(transformations) == 0 {
			continue
		}
		rewrite, err := NewGeneratedRewrite("move:"+filepath.ToSlash(markdownSource.record.Path), finalSourcePath, document, transformations)
		if err != nil {
			return MovePlan{}, err
		}
		info, err := os.Stat(markdownSource.path)
		if err != nil {
			return MovePlan{}, fmt.Errorf("stat Markdown move source %s: %w", markdownSource.path, err)
		}
		plan.rewrites = append(plan.rewrites, plannedMoveRewrite{
			originPath: markdownSource.path,
			mode:       info.Mode(),
			rewrite:    rewrite,
		})
		plan.Updates = append(plan.Updates, MoveUpdate{Path: finalSourcePath, Links: len(transformations)})
		plan.RewrittenLinks += len(transformations)
	}
	sort.Slice(plan.Updates, func(i, j int) bool { return pathKey(plan.Updates[i].Path) < pathKey(plan.Updates[j].Path) })
	sort.Slice(plan.rewrites, func(i, j int) bool {
		return pathKey(plan.rewrites[i].rewrite.Path) < pathKey(plan.rewrites[j].rewrite.Path)
	})
	return plan, nil
}
