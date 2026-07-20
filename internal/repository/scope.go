package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Scope struct {
	RepositoryRoot string
	DocsRoot       string
	ConfigPath     string
	IgnorePath     string
	Initialized    bool
}

type ScopeOptions struct {
	WorkingDirectory string
	ConfigPath       string
	ConfiguredRoot   string
	RootOverride     string
	HasRootOverride  bool
}

func ResolveScope(options ScopeOptions) (Scope, error) {
	cwd, err := filepath.Abs(options.WorkingDirectory)
	if err != nil {
		return Scope{}, err
	}

	configPath := options.ConfigPath
	base := cwd
	repositoryRoot := ""
	initialized := false
	if configPath != "" {
		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(cwd, configPath)
		}
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return Scope{}, err
		}
		base = filepath.Dir(configPath)
		repositoryRoot = base
		if root, ok := RootForConfig(configPath); ok {
			repositoryRoot = root
			base = root
			initialized = true
		}
	}

	rootValue := options.ConfiguredRoot
	if options.HasRootOverride {
		rootValue = options.RootOverride
		if initialized {
			base = repositoryRoot
		} else {
			base = cwd
		}
	}
	if strings.TrimSpace(rootValue) == "" {
		return Scope{}, fmt.Errorf("docs root cannot be empty")
	}
	docsRoot := rootValue
	if !filepath.IsAbs(docsRoot) {
		docsRoot = filepath.Join(base, docsRoot)
	}
	docsRoot, err = filepath.Abs(docsRoot)
	if err != nil {
		return Scope{}, err
	}
	if options.HasRootOverride && !initialized {
		repositoryRoot = docsRoot
	}

	if repositoryRoot != "" {
		if !Contains(repositoryRoot, docsRoot) {
			return Scope{}, fmt.Errorf("docs root must be inside repository root: %s", rootValue)
		}
		if err := validateRealContainment(repositoryRoot, docsRoot); err != nil {
			return Scope{}, err
		}
	} else {
		repositoryRoot = docsRoot
	}

	cleanConfigPath := ""
	if configPath != "" {
		cleanConfigPath = filepath.Clean(configPath)
	}
	return Scope{
		RepositoryRoot: filepath.Clean(repositoryRoot),
		DocsRoot:       filepath.Clean(docsRoot),
		ConfigPath:     cleanConfigPath,
		IgnorePath:     filepath.Join(filepath.Clean(repositoryRoot), ".docignore"),
		Initialized:    initialized,
	}, nil
}

func Contains(root, path string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func validateRealContainment(repositoryRoot, docsRoot string) error {
	realRepositoryRoot, err := filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return nil
	}
	realDocsRoot, err := filepath.EvalSymlinks(docsRoot)
	if err != nil {
		return nil
	}
	if !Contains(realRepositoryRoot, realDocsRoot) {
		return fmt.Errorf("docs root resolves outside repository root: %s", docsRoot)
	}
	return nil
}

func DocsRootExists(scope Scope) bool {
	info, err := os.Stat(scope.DocsRoot)
	return err == nil && info.IsDir()
}
