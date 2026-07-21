package links

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lokee86/demon-docs/internal/repository"
)

func validateMoveContainment(root, source, destination string) error {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve repository root %s: %w", root, err)
	}
	realSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		return fmt.Errorf("resolve move source %s: %w", source, err)
	}
	if !repository.Contains(realRoot, realSource) {
		return fmt.Errorf("move source resolves outside repository root: %s", source)
	}
	realDestinationParent, err := filepath.EvalSymlinks(filepath.Dir(destination))
	if err != nil {
		return fmt.Errorf("resolve move destination parent %s: %w", filepath.Dir(destination), err)
	}
	if !repository.Contains(realRoot, realDestinationParent) {
		return fmt.Errorf("move destination resolves outside repository root: %s", destination)
	}
	return nil
}

func renderMoveTarget(inventory *inventory, syntax, originalPath string, style targetStyle, sourcePath, targetPath, movedSource, destination string) string {
	if syntax != "wiki" || !isBareWikiPath(originalPath) || bareWikiUnambiguousAfterMove(inventory, originalPath, sourcePath, targetPath, movedSource, destination) {
		return renderTargetForSyntax(syntax, originalPath, style, sourcePath, targetPath)
	}
	rendered := renderTargetPath(style, originalPath, sourcePath, targetPath)
	if strings.EqualFold(filepath.Ext(rendered), ".md") {
		rendered = strings.TrimSuffix(rendered, filepath.Ext(rendered))
	}
	return rendered
}

func bareWikiUnambiguousAfterMove(inventory *inventory, originalPath, sourcePath, targetPath, movedSource, destination string) bool {
	adjacent := filepath.Join(filepath.Dir(sourcePath), wikiResolvedPath(originalPath))
	if strings.EqualFold(filepath.Clean(adjacent), filepath.Clean(targetPath)) {
		return true
	}
	seen := map[string]bool{}
	for _, record := range inventory.manifest.Files {
		if !record.Present || record.Kind != "file" || !strings.EqualFold(filepath.Ext(record.Path), ".md") {
			continue
		}
		path := remapMovedPath(recordAbsolute(inventory.root, record), movedSource, destination)
		if !strings.EqualFold(filepath.Base(path), wikiResolvedPath(originalPath)) {
			continue
		}
		seen[pathKey(path)] = true
	}
	return len(seen) <= 1
}

func absoluteFrom(base, value string) (string, error) {
	if !filepath.IsAbs(value) {
		value = filepath.Join(base, value)
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

func finalMoveDestination(source, destination string) (string, error) {
	info, err := os.Stat(destination)
	if err == nil {
		if pathKey(source) == pathKey(destination) && filepath.Clean(source) != filepath.Clean(destination) {
			return filepath.Clean(destination), nil
		}
		if !info.IsDir() {
			return "", fmt.Errorf("move destination already exists: %s", destination)
		}
		destination = filepath.Join(destination, filepath.Base(source))
	}
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("stat move destination %s: %w", destination, err)
	}
	if _, err := os.Stat(destination); err == nil && pathKey(source) != pathKey(destination) {
		return "", fmt.Errorf("move destination already exists: %s", destination)
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("stat move destination %s: %w", destination, err)
	}
	return filepath.Clean(destination), nil
}

func resolveMoveTarget(inventory *inventory, resolved, rawPath, syntax, movedSource string) (string, error) {
	if record, actual := exactTargetForSyntax(inventory, resolved, syntax); record != nil {
		return filepath.Clean(actual), nil
	}
	if record, actual := exactObsidianTarget(inventory, rawPath, syntax); record != nil {
		return filepath.Clean(actual), nil
	}
	candidate := expectedTargetPath(syntax, resolved)
	if info, err := os.Stat(candidate); err == nil && (info.Mode().IsRegular() || info.IsDir()) {
		return filepath.Clean(candidate), nil
	}
	if syntax != "wiki" {
		return "", nil
	}
	candidates := candidatePathsForSyntax(inventory, resolved, "", syntax)
	if len(candidates) == 1 {
		return filepath.Clean(candidates[0]), nil
	}
	if len(candidates) > 1 {
		for _, candidate := range candidates {
			if repository.Contains(movedSource, candidate) {
				return "", fmt.Errorf("ambiguous wiki target %q includes the move source; candidates: %s", resolved, strings.Join(displayPaths(inventory.root, candidates), ", "))
			}
		}
	}
	return "", nil
}

func remapMovedPath(path, source, destination string) string {
	if !repository.Contains(source, path) {
		return filepath.Clean(path)
	}
	relative, err := filepath.Rel(source, path)
	if err != nil || relative == "." {
		return filepath.Clean(destination)
	}
	return filepath.Clean(filepath.Join(destination, relative))
}
