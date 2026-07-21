package codemapcorpus

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

type sourceFactJob struct {
	path         string
	dependencies bool
	symbols      bool
}

type sourceFactResult struct {
	dependencies []evidence.DependencyEdge
	symbols      []evidence.SymbolDeclaration
}

func collectSourceFacts(root string, files []string) ([]evidence.DependencyEdge, []evidence.SymbolDeclaration, error) {
	return collectSourceFactsWithReader(root, files, os.ReadFile)
}

func collectSourceFactsWithReader(
	root string,
	files []string,
	readFile func(string) ([]byte, error),
) ([]evidence.DependencyEdge, []evidence.SymbolDeclaration, error) {
	index, err := newDependencyIndex(root, files)
	if err != nil {
		return nil, nil, err
	}
	jobs := sourceFactJobs(files)
	results := make([]sourceFactResult, len(jobs))
	errors := runCorpusWorkers(len(jobs), func(indexPosition int) error {
		job := jobs[indexPosition]
		contents, err := readFile(filepath.Join(root, filepath.FromSlash(job.path)))
		if err != nil {
			return fmt.Errorf("read source %s: %w", job.path, err)
		}
		result := sourceFactResult{}
		if job.dependencies {
			result.dependencies = index.edgesFor(job.path, contents)
		}
		if job.symbols {
			result.symbols = declarationsForSource(job.path, contents)
		}
		results[indexPosition] = result
		return nil
	})
	for _, err := range errors {
		if err != nil {
			return nil, nil, err
		}
	}
	dependencies, symbols := mergeSourceFacts(results)
	return dependencies, symbols, nil
}

func sourceFactJobs(files []string) []sourceFactJob {
	jobs := make([]sourceFactJob, 0)
	for _, file := range files {
		extension := strings.ToLower(path.Ext(file))
		dependencies := supportedDependencySource(file)
		symbols := extension == ".go" || extension == ".gd"
		if dependencies || symbols {
			jobs = append(jobs, sourceFactJob{path: file, dependencies: dependencies, symbols: symbols})
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].path < jobs[j].path })
	return jobs
}

func declarationsForSource(file string, contents []byte) []evidence.SymbolDeclaration {
	var symbols []string
	switch strings.ToLower(path.Ext(file)) {
	case ".go":
		symbols = goDeclaredSymbols(file, contents)
	case ".gd":
		symbols = gdscriptDeclaredSymbols(contents)
	}
	result := make([]evidence.SymbolDeclaration, 0, len(symbols))
	for _, symbol := range symbols {
		result = append(result, evidence.SymbolDeclaration{Path: file, Symbol: symbol})
	}
	return result
}

func mergeSourceFacts(results []sourceFactResult) ([]evidence.DependencyEdge, []evidence.SymbolDeclaration) {
	edges := map[string]evidence.DependencyEdge{}
	declarations := map[string]evidence.SymbolDeclaration{}
	for _, result := range results {
		for _, item := range result.dependencies {
			if item.Source == item.Target || item.Target == "" {
				continue
			}
			edges[item.Source+"\x00"+item.Target+"\x00"+item.Relation] = item
		}
		for _, item := range result.symbols {
			declarations[item.Path+"\x00"+item.Symbol] = item
		}
	}

	dependencyResult := make([]evidence.DependencyEdge, 0, len(edges))
	for _, item := range edges {
		dependencyResult = append(dependencyResult, item)
	}
	sort.Slice(dependencyResult, func(i, j int) bool {
		left := dependencyResult[i].Source + "\x00" + dependencyResult[i].Target + "\x00" + dependencyResult[i].Relation
		right := dependencyResult[j].Source + "\x00" + dependencyResult[j].Target + "\x00" + dependencyResult[j].Relation
		return left < right
	})

	symbolResult := make([]evidence.SymbolDeclaration, 0, len(declarations))
	for _, item := range declarations {
		symbolResult = append(symbolResult, item)
	}
	sort.Slice(symbolResult, func(i, j int) bool {
		if symbolResult[i].Path != symbolResult[j].Path {
			return symbolResult[i].Path < symbolResult[j].Path
		}
		return symbolResult[i].Symbol < symbolResult[j].Symbol
	})
	return dependencyResult, symbolResult
}
