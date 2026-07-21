package links

import (
	"net/url"
	"path/filepath"
	"strings"
)

func exactTargetForSyntax(inventory *inventory, resolved, syntax string) (*FileRecord, string) {
	record, actual := inventory.exact(resolved)
	if record != nil || syntax != "wiki" || filepath.Ext(resolved) != "" {
		return record, actual
	}
	if record, actual = inventory.exact(wikiResolvedPath(resolved)); record != nil {
		return record, actual
	}
	candidates := inventory.candidates(filepath.Base(wikiResolvedPath(resolved)), "file")
	if len(candidates) == 1 {
		return inventory.exact(candidates[0])
	}
	return nil, ""
}

func exactObsidianTarget(inventory *inventory, rawPath, syntax string) (*FileRecord, string) {
	if rawPath == "" || (syntax != "inline" && syntax != "reference" && syntax != "wiki") {
		return nil, ""
	}
	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		decoded = rawPath
	}
	if hasScheme(decoded) || filepath.IsAbs(filepath.FromSlash(decoded)) || isDrivePath(decoded) || strings.HasPrefix(decoded, `\\`) {
		return nil, ""
	}

	candidate := filepath.FromSlash(decoded)
	if syntax == "wiki" && filepath.Ext(candidate) == "" {
		candidate += ".md"
	}
	if strings.ContainsAny(decoded, `/\`) {
		if strings.HasPrefix(decoded, "./") || strings.HasPrefix(decoded, ".\\") || strings.HasPrefix(decoded, "../") || strings.HasPrefix(decoded, "..\\") {
			return nil, ""
		}
		return inventory.exact(filepath.Join(inventory.root, candidate))
	}
	candidates := inventory.candidates(filepath.Base(candidate), "file")
	if len(candidates) == 1 {
		return inventory.exact(candidates[0])
	}
	return nil, ""
}

func isObsidianBareMarkdownPath(rawPath, syntax string) bool {
	return rawPath != "" && (syntax == "inline" || syntax == "reference") &&
		!strings.ContainsAny(rawPath, `/\`) && rawPath != "." && rawPath != ".."
}

func targetCaseMismatch(syntax, resolved, actual string) bool {
	expected := expectedTargetPath(syntax, resolved)
	if filepath.Clean(actual) == filepath.Clean(expected) {
		return false
	}
	if strings.EqualFold(filepath.Clean(actual), filepath.Clean(expected)) {
		return true
	}
	return syntax != "wiki" || filepath.Ext(resolved) != ""
}

func expectedTargetPath(syntax, resolved string) string {
	if syntax == "wiki" && filepath.Ext(resolved) == "" {
		return wikiResolvedPath(resolved)
	}
	return resolved
}

func candidatePathsForSyntax(inventory *inventory, resolved, preferredID, syntax string) []string {
	if syntax == "wiki" && filepath.Ext(resolved) == "" {
		return candidatePaths(inventory, wikiResolvedPath(resolved), preferredID)
	}
	return candidatePaths(inventory, resolved, preferredID)
}

func renderTargetForSyntax(syntax, originalPath string, style targetStyle, sourcePath, targetPath string) string {
	if syntax == "wiki" && originalPath == "" && filepath.Clean(sourcePath) == filepath.Clean(targetPath) {
		return ""
	}
	if syntax == "wiki" && isBareWikiPath(originalPath) {
		return strings.TrimSuffix(filepath.Base(targetPath), filepath.Ext(targetPath))
	}
	rendered := renderTargetPath(style, originalPath, sourcePath, targetPath)
	if syntax == "wiki" && filepath.Ext(originalPath) == "" && strings.EqualFold(filepath.Ext(rendered), ".md") {
		rendered = strings.TrimSuffix(rendered, filepath.Ext(rendered))
	}
	return rendered
}

func isBareWikiPath(path string) bool {
	return path != "" && filepath.Ext(path) == "" && !strings.ContainsAny(path, `/\`)
}

func wikiResolvedPath(path string) string {
	if filepath.Ext(path) == "" {
		return path + ".md"
	}
	return path
}
