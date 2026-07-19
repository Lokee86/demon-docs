package codemapcorpus

import (
	"errors"
	"io"

	"github.com/Lokee86/demon-docs/internal/evidence"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func collectCommits(root string, files []string, options Options) ([]evidence.Commit, error) {
	fileSet := make(map[string]struct{}, len(files))
	for _, file := range files {
		fileSet[file] = struct{}{}
	}
	if commits, ok, err := collectCommitsGitCLI(root, fileSet, options); ok || err != nil {
		return commits, err
	}
	return collectCommitsGoGit(root, fileSet, options)
}

func collectCommitsGoGit(
	root string,
	fileSet map[string]struct{},
	options Options,
) ([]evidence.Commit, error) {
	repository, err := git.PlainOpen(root)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	head, err := repository.Head()
	if err != nil {
		return nil, err
	}
	iterator, err := repository.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	result := make([]evidence.Commit, 0, options.MaxCommits)
	for examined := 0; examined < options.MaxCommits; examined++ {
		commit, err := iterator.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if commit.NumParents() != 1 {
			continue
		}
		parent, err := commit.Parent(0)
		if err != nil {
			return nil, err
		}
		parentTree, err := parent.Tree()
		if err != nil {
			return nil, err
		}
		commitTree, err := commit.Tree()
		if err != nil {
			return nil, err
		}
		changes, err := object.DiffTree(parentTree, commitTree)
		if err != nil {
			return nil, err
		}
		if len(changes) > options.MaxPathsPerCommit {
			continue
		}
		paths := make(map[string]struct{}, len(changes))
		for _, change := range changes {
			for _, name := range []string{change.From.Name, change.To.Name} {
				candidate := normalizePath(name)
				if _, exists := fileSet[candidate]; exists {
					paths[candidate] = struct{}{}
				}
			}
		}
		if len(paths) < 2 {
			continue
		}
		result = append(result, evidence.Commit{ID: commit.Hash.String(), Paths: sortedSet(paths)})
	}
	return result, nil
}
