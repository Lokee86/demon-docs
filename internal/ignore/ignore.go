package ignore

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const FileName = ".docignore"

var permanentDirectories = []string{
	".git",
	".ddocs",
	".obsidian",
	"logseq",
}

type Policy struct {
	root    string
	matcher gitignore.Matcher
}

func Load(root string) (Policy, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Policy{}, err
	}
	policy := Policy{root: filepath.Clean(abs), matcher: gitignore.NewMatcher(nil)}
	file, err := os.Open(filepath.Join(policy.root, FileName))
	if errors.Is(err, os.ErrNotExist) {
		return policy, nil
	}
	if err != nil {
		return Policy{}, fmt.Errorf("open %s: %w", FileName, err)
	}
	defer file.Close()

	patterns := make([]gitignore.Pattern, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		patterns = append(patterns, gitignore.ParsePattern(scanner.Text(), nil))
	}
	if err := scanner.Err(); err != nil {
		return Policy{}, fmt.Errorf("read %s: %w", FileName, err)
	}
	policy.matcher = gitignore.NewMatcher(patterns)
	return policy, nil
}

func (p Policy) Ignored(path string, isDir bool) (bool, error) {
	parts, err := p.relativeParts(path)
	if err != nil {
		return false, err
	}
	if permanentlyIgnored(parts, isDir) {
		return true, nil
	}
	return p.matcher.Match(parts, isDir), nil
}

func (p Policy) Root() string {
	return p.root
}

func (p Policy) IsControlFile(path string) bool {
	parts, err := p.relativeParts(path)
	return err == nil && len(parts) == 1 && parts[0] == FileName
}

func (p Policy) relativeParts(path string) ([]string, error) {
	rel, err := filepath.Rel(p.root, path)
	if err != nil {
		return nil, err
	}
	rel = filepath.Clean(rel)
	if rel == "." {
		return nil, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path %s is outside ignore root %s", path, p.root)
	}
	return strings.Split(filepath.ToSlash(rel), "/"), nil
}

func permanentlyIgnored(parts []string, isDir bool) bool {
	limit := len(parts)
	if !isDir && limit > 0 {
		limit--
	}
	for _, part := range parts[:limit] {
		for _, permanent := range permanentDirectories {
			if part == permanent || runtime.GOOS == "windows" && strings.EqualFold(part, permanent) {
				return true
			}
		}
	}
	return false
}
