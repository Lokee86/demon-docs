package ignore

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Hierarchy applies repository-root and nested .docignore files using
// gitignore domains rooted at the directory containing each file.
type Hierarchy struct {
	root     string
	patterns []gitignore.Pattern
	matcher  gitignore.Matcher
	loaded   map[string]struct{}
}

func LoadHierarchy(root string) (*Hierarchy, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	h := &Hierarchy{
		root:    filepath.Clean(abs),
		matcher: gitignore.NewMatcher(nil),
		loaded:  map[string]struct{}{},
	}
	if err := h.LoadDirectory(h.root); err != nil {
		return nil, err
	}
	return h, nil
}

// LoadDirectory adds the .docignore in dir, if present. Repeated loads are
// ignored so callers can safely share one hierarchy across multiple roots.
func (h *Hierarchy) LoadDirectory(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir = filepath.Clean(dir)
	if _, err := h.relativeParts(dir); err != nil {
		return err
	}
	if _, ok := h.loaded[dir]; ok {
		return nil
	}
	h.loaded[dir] = struct{}{}

	file, err := os.Open(filepath.Join(dir, FileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", filepath.Join(dir, FileName), err)
	}
	defer file.Close()

	domain, err := h.relativeParts(dir)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		h.patterns = append(h.patterns, gitignore.ParsePattern(scanner.Text(), domain))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", filepath.Join(dir, FileName), err)
	}
	h.matcher = gitignore.NewMatcher(h.patterns)
	return nil
}

// LoadAncestors loads .docignore files from the repository root through dir.
func (h *Hierarchy) LoadAncestors(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	dir = filepath.Clean(dir)
	parts, err := h.relativeParts(dir)
	if err != nil {
		return err
	}
	current := h.root
	if err := h.LoadDirectory(current); err != nil {
		return err
	}
	for _, part := range parts {
		current = filepath.Join(current, part)
		if err := h.LoadDirectory(current); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hierarchy) Ignored(path string, isDir bool) (bool, error) {
	parts, err := h.relativeParts(path)
	if err != nil {
		return false, err
	}
	if permanentlyIgnored(parts, isDir) {
		return true, nil
	}
	return h.matcher.Match(parts, isDir), nil
}

func (h *Hierarchy) Root() string { return h.root }

func (h *Hierarchy) IsControlFile(path string) bool {
	return filepath.Base(path) == FileName
}

func (h *Hierarchy) relativeParts(path string) ([]string, error) {
	rel, err := filepath.Rel(h.root, filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if rel == "." {
		return nil, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path %s is outside ignore root %s", path, h.root)
	}
	return strings.Split(filepath.ToSlash(rel), "/"), nil
}
