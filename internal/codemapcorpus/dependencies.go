package codemapcorpus

import (
	"path"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

type dependencyIndex struct {
	files      map[string]struct{}
	filesByDir map[string][]string
	goModules  []goModule
	godotRoots []string
}

func newDependencyIndex(root string, files []string) (dependencyIndex, error) {
	index := dependencyIndex{
		files:      make(map[string]struct{}, len(files)),
		filesByDir: map[string][]string{},
	}
	for _, file := range files {
		index.files[file] = struct{}{}
		directory := path.Dir(file)
		index.filesByDir[directory] = append(index.filesByDir[directory], file)
		if path.Base(file) == "project.godot" {
			index.godotRoots = append(index.godotRoots, directory)
		}
	}
	for directory := range index.filesByDir {
		sort.Strings(index.filesByDir[directory])
	}
	sort.Slice(index.godotRoots, func(i, j int) bool { return len(index.godotRoots[i]) > len(index.godotRoots[j]) })
	modules, err := loadGoModules(root, files)
	if err != nil {
		return dependencyIndex{}, err
	}
	index.goModules = modules
	return index, nil
}

func (index dependencyIndex) edgesFor(source string, contents []byte) []evidence.DependencyEdge {
	switch strings.ToLower(path.Ext(source)) {
	case ".go":
		return index.goEdges(source, contents)
	case ".gd":
		return index.gdscriptEdges(source, string(contents))
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return index.javascriptEdges(source, string(contents))
	case ".rb":
		return index.rubyEdges(source, string(contents))
	case ".py":
		return index.pythonEdges(source, string(contents))
	default:
		return nil
	}
}

func supportedDependencySource(file string) bool {
	switch strings.ToLower(path.Ext(file)) {
	case ".go", ".gd", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".rb", ".py":
		return true
	default:
		return false
	}
}

func (index dependencyIndex) relativeTargets(source, reference string, extensions []string, indexes bool) []string {
	if !strings.HasPrefix(reference, ".") {
		return nil
	}
	base := normalizePath(path.Join(path.Dir(source), reference))
	if base == "" {
		return nil
	}
	candidates := []string{base}
	if path.Ext(base) == "" {
		for _, extension := range extensions {
			candidates = append(candidates, base+extension)
		}
		if indexes {
			for _, extension := range extensions {
				candidates = append(candidates, path.Join(base, "index"+extension))
			}
		}
	}
	return index.existing(candidates)
}

func (index dependencyIndex) existing(candidates []string) []string {
	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		candidate = normalizePath(candidate)
		if _, exists := index.files[candidate]; exists {
			seen[candidate] = struct{}{}
		}
	}
	return sortedSet(seen)
}

func edge(source, target, relation string) evidence.DependencyEdge {
	return evidence.DependencyEdge{Source: source, Target: target, Relation: relation}
}
