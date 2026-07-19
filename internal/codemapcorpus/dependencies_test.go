package codemapcorpus

import (
	"testing"

	"github.com/Lokee86/demon-docs/internal/evidence"
)

func TestBuildCollectsSupportedLocalDependencies(t *testing.T) {
	root := t.TempDir()
	files := fixtureFiles()
	writeFiles(t, root, files)
	tracked := make([]string, 0, len(files))
	for file := range files {
		if file != ".ddocs/private.go" && file != "vendor/ignored.go" {
			tracked = append(tracked, file)
		}
	}
	initializeHistory(t, root, tracked)

	corpus, err := Build(root, fixtureDataset(), Options{MaxCommits: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []evidence.DependencyEdge{
		{Source: "server/a/a.go", Target: "server/b/b.go", Relation: "go_import"},
		{Source: "client/main.gd", Target: "client/shared/tool.gd", Relation: "gdscript_resource"},
		{Source: "web/a.ts", Target: "web/b.ts", Relation: "javascript_import"},
		{Source: "api/a.rb", Target: "api/b.rb", Relation: "ruby_require_relative"},
		{Source: "tools/pkg/a.py", Target: "tools/pkg/b.py", Relation: "python_relative_import"},
	} {
		if !containsEdge(corpus.DependencyEdges, expected) {
			t.Errorf("missing dependency edge %#v in %#v", expected, corpus.DependencyEdges)
		}
	}
}

func containsEdge(edges []evidence.DependencyEdge, expected evidence.DependencyEdge) bool {
	for _, edge := range edges {
		if edge == expected {
			return true
		}
	}
	return false
}
