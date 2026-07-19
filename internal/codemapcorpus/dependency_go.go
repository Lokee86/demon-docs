package codemapcorpus

import (
	"bufio"
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

type goModule struct {
	Path string
	Root string
}

func loadGoModules(root string, files []string) ([]goModule, error) {
	modules := make([]goModule, 0)
	for _, file := range files {
		if path.Base(file) != "go.mod" {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(bytes.NewReader(contents))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 2 && fields[0] == "module" {
				modules = append(modules, goModule{Path: strings.Trim(fields[1], `"`), Root: path.Dir(file)})
				break
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	sort.Slice(modules, func(i, j int) bool { return len(modules[i].Path) > len(modules[j].Path) })
	return modules, nil
}

func (index dependencyIndex) goEdges(source string, contents []byte) []evidence.DependencyEdge {
	parsed, err := parser.ParseFile(token.NewFileSet(), source, contents, parser.ImportsOnly)
	if err != nil {
		return nil
	}
	result := make([]evidence.DependencyEdge, 0)
	for _, imported := range parsed.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			continue
		}
		for _, module := range index.goModules {
			if importPath != module.Path && !strings.HasPrefix(importPath, module.Path+"/") {
				continue
			}
			relativePackage := strings.TrimPrefix(importPath, module.Path)
			relativePackage = strings.TrimPrefix(relativePackage, "/")
			packageDirectory := normalizePath(path.Join(module.Root, relativePackage))
			for _, target := range index.filesByDir[packageDirectory] {
				if path.Ext(target) == ".go" && !strings.HasSuffix(target, "_test.go") {
					result = append(result, edge(source, target, "go_import"))
				}
			}
			break
		}
	}
	return result
}
