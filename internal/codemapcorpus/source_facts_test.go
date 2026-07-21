package codemapcorpus

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCollectSourceFactsReadsEachSupportedSourceOnce(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":               "module example.com/demo\n\ngo 1.26\n",
		"server/a/a.go":        "package a\nimport _ \"example.com/demo/server/b\"\ntype AlphaRuntime struct{}\n",
		"server/b/b.go":        "package b\ntype BetaRuntime struct{}\n",
		"client/project.godot": "[application]\n",
		"client/main.gd":       "class_name MainRuntime\nfunc start_runtime():\n\tpass\n",
		"web/main.ts":          "export const value = 1\n",
		"README.md":            "# ignored source kind\n",
	}
	writeFiles(t, root, files)
	ordered := make([]string, 0, len(files))
	for file := range files {
		ordered = append(ordered, file)
	}

	reads := map[string]int{}
	var lock sync.Mutex
	dependencies, symbols, err := collectSourceFactsWithReader(root, ordered, func(fullPath string) ([]byte, error) {
		relative, err := filepath.Rel(root, fullPath)
		if err != nil {
			return nil, err
		}
		relative = filepath.ToSlash(relative)
		lock.Lock()
		reads[relative]++
		lock.Unlock()
		return os.ReadFile(fullPath)
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"server/a/a.go", "server/b/b.go", "client/main.gd", "web/main.ts"} {
		if reads[file] != 1 {
			t.Fatalf("source reads for %s = %d, want 1", file, reads[file])
		}
	}
	if reads["README.md"] != 0 || reads["go.mod"] != 0 || reads["client/project.godot"] != 0 {
		t.Fatalf("unsupported files passed to source reader: %#v", reads)
	}
	if len(dependencies) != 1 || dependencies[0].Source != "server/a/a.go" || dependencies[0].Target != "server/b/b.go" {
		t.Fatalf("unexpected dependencies: %#v", dependencies)
	}
	if len(symbols) < 3 {
		t.Fatalf("expected Go and GDScript symbols, got %#v", symbols)
	}
}

func TestCollectSourceFactsReturnsReadErrorsInPathOrder(t *testing.T) {
	root := t.TempDir()
	first := errors.New("first failure")
	second := errors.New("second failure")
	_, _, err := collectSourceFactsWithReader(root, []string{"b.go", "a.go"}, func(fullPath string) ([]byte, error) {
		switch filepath.Base(fullPath) {
		case "a.go":
			time.Sleep(2 * time.Millisecond)
			return nil, first
		default:
			return nil, second
		}
	})
	if !errors.Is(err, first) || !strings.Contains(err.Error(), "a.go") {
		t.Fatalf("error = %v, want first path-ordered failure", err)
	}
}
