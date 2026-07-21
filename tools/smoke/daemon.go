package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (h *harness) daemonScenario() error {
	h.step("verify detached daemon filesystem maintenance")
	defer func() { _, _ = h.command(h.repo, h.ddocs, "demon", "run", "--false") }()
	for _, path := range []string{"docs/daemon-source.md", "docs/daemon-target.md"} {
		if _, err := h.command(h.repo, h.ddocs, "new", "general", path); err != nil {
			return err
		}
	}
	source := filepath.Join(h.repo, "docs", "daemon-source.md")
	if err := appendFile(source, "\n[Daemon target](daemon-target.md)\n"); err != nil {
		return err
	}
	guide := filepath.Join(h.repo, "docs", "guide.md")
	if err := appendFile(guide, "\nSee [Daemon source](daemon-source.md).\n"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "fix", "--indexes", "--links", "--reverse"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "demon", "run", "--true"); err != nil {
		return err
	}
	oldTarget := filepath.Join(h.repo, "docs", "daemon-target.md")
	newTarget := filepath.Join(h.repo, "docs", "daemon-target-renamed.md")
	if err := os.Rename(oldTarget, newTarget); err != nil {
		return err
	}
	if err := waitFor("daemon link rewrite", 15*time.Second, func() bool {
		return fileContains(source, "daemon-target-renamed.md")
	}); err != nil {
		return h.daemonFailure(err)
	}
	created := filepath.Join(h.repo, "docs", "daemon-created.md")
	if err := writeFile(created, "# Daemon Created\n"); err != nil {
		return err
	}
	docsIndex := filepath.Join(h.repo, "docs", "INDEX.md")
	if err := waitFor("daemon index insertion", 15*time.Second, func() bool {
		return fileContains(docsIndex, "daemon-created.md")
	}); err != nil {
		return h.daemonFailure(err)
	}
	for _, unexpected := range []string{"document_id:", "## Purpose"} {
		if err := requireNotContains(created, unexpected); err != nil {
			return err
		}
	}
	if err := os.Remove(created); err != nil {
		return err
	}
	if err := waitFor("daemon index removal", 15*time.Second, func() bool {
		return !fileContains(docsIndex, "daemon-created.md")
	}); err != nil {
		return h.daemonFailure(err)
	}
	if err := writeFile(filepath.Join(h.repo, "src", "worker2.go"), "package src\n\nfunc WorkerTwo() {}\n"); err != nil {
		return err
	}
	codeMap := filepath.Join(h.repo, "docs", "code-map.md")
	if err := appendFile(codeMap, "- `src/worker2.go`\n"); err != nil {
		return err
	}
	if err := waitFor("daemon reverse-index refresh", 15*time.Second, func() bool {
		return fileContains(filepath.Join(h.repo, "src", "INDEX.md"), "worker2.go")
	}); err != nil {
		return h.daemonFailure(err)
	}
	_, err := h.command(h.repo, h.ddocs, "demon", "run", "--false")
	return err
}

func (h *harness) daemonFailure(cause error) error {
	logs, _ := h.command(h.repo, h.ddocs, "demon", "--logs")
	return fmt.Errorf("%w\ndaemon logs:\n%s", cause, logs)
}
