package codemapcorpus

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, contents := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func initializeHistory(t *testing.T, root string, initialFiles []string) *git.Repository {
	t.Helper()
	repository, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	worktree, err := repository.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range initialFiles {
		if _, err := worktree.Add(file); err != nil {
			t.Fatal(err)
		}
	}
	commit(t, worktree, "initial", time.Unix(1, 0))
	return repository
}

func commit(t *testing.T, worktree *git.Worktree, message string, when time.Time) {
	t.Helper()
	_, err := worktree.Commit(message, &git.CommitOptions{Author: &object.Signature{
		Name: "Test", Email: "test@example.com", When: when,
	}})
	if err != nil {
		t.Fatal(err)
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
