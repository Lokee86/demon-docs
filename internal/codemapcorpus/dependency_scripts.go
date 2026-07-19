package codemapcorpus

import (
	"path"
	"regexp"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

var (
	godotReferencePattern = regexp.MustCompile(`(?m)(?:preload|load)\(\s*["']([^"']+)["']|extends\s+["']([^"']+)["']`)
	javascriptFromPattern = regexp.MustCompile(`(?m)(?:from\s+|import\s*\(|require\s*\()\s*["']([^"']+)["']`)
	javascriptSidePattern = regexp.MustCompile(`(?m)^\s*import\s*["']([^"']+)["']`)
	rubyRequirePattern    = regexp.MustCompile(`(?m)require_relative\s*(?:\(\s*)?["']([^"']+)["']`)
	pythonFromPattern     = regexp.MustCompile(`(?m)^\s*from\s+(\.+)([A-Za-z0-9_.]+)\s+import\s+`)
)

func (index dependencyIndex) gdscriptEdges(source, contents string) []evidence.DependencyEdge {
	result := make([]evidence.DependencyEdge, 0)
	for _, match := range godotReferencePattern.FindAllStringSubmatch(contents, -1) {
		reference := match[1]
		if reference == "" {
			reference = match[2]
		}
		var targets []string
		if strings.HasPrefix(reference, "res://") {
			if root := index.godotRoot(source); root != "" {
				targets = index.existing([]string{path.Join(root, strings.TrimPrefix(reference, "res://"))})
			}
		} else {
			targets = index.relativeTargets(source, reference, []string{".gd", ".tscn", ".tres"}, false)
		}
		for _, target := range targets {
			result = append(result, edge(source, target, "gdscript_resource"))
		}
	}
	return result
}

func (index dependencyIndex) godotRoot(source string) string {
	for _, root := range index.godotRoots {
		if source == root || strings.HasPrefix(source, root+"/") {
			return root
		}
	}
	return ""
}

func (index dependencyIndex) javascriptEdges(source, contents string) []evidence.DependencyEdge {
	references := make([]string, 0)
	for _, pattern := range []*regexp.Regexp{javascriptFromPattern, javascriptSidePattern} {
		for _, match := range pattern.FindAllStringSubmatch(contents, -1) {
			references = append(references, match[1])
		}
	}
	result := make([]evidence.DependencyEdge, 0)
	for _, reference := range references {
		for _, target := range index.relativeTargets(source, reference,
			[]string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".json"}, true) {
			result = append(result, edge(source, target, "javascript_import"))
		}
	}
	return result
}

func (index dependencyIndex) rubyEdges(source, contents string) []evidence.DependencyEdge {
	result := make([]evidence.DependencyEdge, 0)
	for _, match := range rubyRequirePattern.FindAllStringSubmatch(contents, -1) {
		reference := match[1]
		if !strings.HasPrefix(reference, ".") {
			reference = "./" + reference
		}
		for _, target := range index.relativeTargets(source, reference, []string{".rb"}, false) {
			result = append(result, edge(source, target, "ruby_require_relative"))
		}
	}
	return result
}

func (index dependencyIndex) pythonEdges(source, contents string) []evidence.DependencyEdge {
	result := make([]evidence.DependencyEdge, 0)
	for _, match := range pythonFromPattern.FindAllStringSubmatch(contents, -1) {
		base := path.Dir(source)
		for level := 1; level < len(match[1]); level++ {
			base = path.Dir(base)
		}
		module := strings.ReplaceAll(match[2], ".", "/")
		candidate := normalizePath(path.Join(base, module))
		for _, target := range index.existing([]string{candidate + ".py", path.Join(candidate, "__init__.py")}) {
			result = append(result, edge(source, target, "python_relative_import"))
		}
	}
	return result
}
