package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type options struct {
	ddocs      string
	demon      string
	keep       bool
	skipDaemon bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "SMOKE FAILED:", err)
		os.Exit(1)
	}
}

func run() error {
	var opts options
	flag.StringVar(&opts.ddocs, "ddocs", "", "path to a prebuilt ddocs binary")
	flag.StringVar(&opts.demon, "demon", "", "path to a prebuilt demon binary")
	flag.BoolVar(&opts.keep, "keep", false, "keep the disposable workspace after success")
	flag.BoolVar(&opts.skipDaemon, "skip-daemon", false, "skip detached daemon behavior checks")
	flag.Parse()

	if (opts.ddocs == "") != (opts.demon == "") {
		return fmt.Errorf("--ddocs and --demon must be supplied together")
	}
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	workspace, err := os.MkdirTemp("", "demon-docs-smoke-")
	if err != nil {
		return err
	}
	h := newHarness(root, workspace)
	passed := false
	defer func() {
		if passed && !opts.keep {
			_ = os.RemoveAll(workspace)
		} else {
			fmt.Fprintln(os.Stderr, "smoke workspace:", workspace)
		}
	}()

	if opts.ddocs == "" {
		if err := h.buildBinaries(); err != nil {
			return err
		}
	} else if err := h.useBinaries(opts.ddocs, opts.demon); err != nil {
		return err
	}
	if err := h.runSmoke(opts.skipDaemon); err != nil {
		return err
	}
	passed = true
	fmt.Println("SMOKE PASSED")
	if opts.keep {
		fmt.Println("workspace:", workspace)
	}
	return nil
}

func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above current directory")
		}
		dir = parent
	}
}
