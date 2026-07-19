package codemapcorpus

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Lokee86/demon-docs/internal/evidence"
	ignorepolicy "github.com/Lokee86/demon-docs/internal/ignore"
)

const commitMarker = "@@DDOCS_COMMIT@@"

func gitCLIRepositoryFiles(root string, policy ignorepolicy.Policy) ([]string, bool, error) {
	output, ok, err := runGit(root, "ls-files", "-z")
	if !ok || err != nil {
		return nil, ok, err
	}
	files := map[string]struct{}{}
	for _, raw := range bytes.Split(output, []byte{0}) {
		file := normalizePath(string(raw))
		if file == "" {
			continue
		}
		ignored, err := policy.Ignored(joinRepositoryPath(root, file), false)
		if err != nil {
			return nil, true, err
		}
		if !ignored {
			files[file] = struct{}{}
		}
	}
	return sortedSet(files), true, nil
}

func collectCommitsGitCLI(
	root string,
	fileSet map[string]struct{},
	options Options,
) ([]evidence.Commit, bool, error) {
	output, ok, err := runGit(root,
		"-c", "core.quotepath=false",
		"log", "--no-merges", "--min-parents=1",
		"-n", strconv.Itoa(options.MaxCommits),
		"--pretty=format:"+commitMarker+"%H", "--name-only",
	)
	if !ok || err != nil {
		return nil, ok, err
	}
	result := make([]evidence.Commit, 0, options.MaxCommits)
	var id string
	paths := map[string]struct{}{}
	pathCount := 0
	flush := func() {
		if id != "" && pathCount <= options.MaxPathsPerCommit && len(paths) >= 2 {
			result = append(result, evidence.Commit{ID: id, Paths: sortedSet(paths)})
		}
		id = ""
		paths = map[string]struct{}{}
		pathCount = 0
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if strings.HasPrefix(line, commitMarker) {
			flush()
			id = strings.TrimPrefix(line, commitMarker)
			continue
		}
		if id == "" || line == "" {
			continue
		}
		pathCount++
		candidate := normalizePath(line)
		if _, exists := fileSet[candidate]; exists {
			paths[candidate] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, true, err
	}
	flush()
	return result, true, nil
}

func runGit(root string, arguments ...string) ([]byte, bool, error) {
	args := append([]string{"-C", root}, arguments...)
	command := exec.Command("git", args...)
	output, err := command.Output()
	if err == nil {
		return output, true, nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return nil, false, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return nil, false, nil
	}
	return nil, true, err
}
