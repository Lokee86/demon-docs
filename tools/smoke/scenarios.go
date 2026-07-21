package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func (h *harness) runSmoke(skipDaemon bool) error {
	if err := h.binarySurface(); err != nil {
		return err
	}
	if err := h.initializeFixture(); err != nil {
		return err
	}
	if err := h.documentationScenario(); err != nil {
		return err
	}
	if err := h.linkAndMoveScenario(); err != nil {
		return err
	}
	if err := h.reverseIndexScenario(); err != nil {
		return err
	}
	if !skipDaemon {
		if err := h.daemonScenario(); err != nil {
			return err
		}
	}
	_, err := h.command(h.repo, h.ddocs, "check", "--indexes", "--links", "--reverse")
	return err
}

func (h *harness) binarySurface() error {
	h.step("verify CLI surfaces")
	commands := [][]string{
		{h.ddocs, "--help"}, {h.ddocs, "--version"}, {h.ddocs, "config", "paths"},
		{h.demon, "--help"}, {h.demon, "--version"}, {h.demon, "config", "paths"},
	}
	for _, command := range commands {
		if _, err := h.command(h.root, command[0], command[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func (h *harness) initializeFixture() error {
	h.step("initialize isolated repository")
	if err := os.MkdirAll(filepath.Join(h.repo, "docs"), 0o755); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "init", "--root", "docs"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "demon", "run", "--false"); err != nil {
		return err
	}
	config := filepath.Join(h.repo, ".ddocs", "config.toml")
	changes := [][2]string{
		{`default_author = ""`, `default_author = "Smoke Harness"`},
		{"[frontmatter.fields.summary]\ntype = \"string\"\nrequired = true", "[frontmatter.fields.summary]\ntype = \"string\"\nrequired = true\ndefault = \"Smoke harness document.\""},
		{"[reverse_index]\nroots = []", "[reverse_index]\nroots = [\"src\"]"},
		{"debounce_seconds = 0.75", "debounce_seconds = 0.1"},
	}
	for _, change := range changes {
		if err := replaceFile(config, change[0], change[1]); err != nil {
			return err
		}
	}
	return writeFile(filepath.Join(h.repo, "src", "worker.go"), "package src\n\nfunc Worker() {}\n")
}

func (h *harness) documentationScenario() error {
	h.step("verify documentation policy convergence")
	guide := "# Guide\n\nSee [Target](target.md).\n"
	target := "# Target\n\nSee [Guide](guide.md).\n"
	if err := writeFile(filepath.Join(h.repo, "docs", "guide.md"), guide); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(h.repo, "docs", "target.md"), target); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(h.repo, "docs", ".obsidian", "workspace.md"), "# Private editor state\n"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "fix", "--docs", "--links"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "check", "--docs", "--links"); err != nil {
		return err
	}
	guidePath := filepath.Join(h.repo, "docs", "guide.md")
	for _, expected := range []string{"document_id:", "author: Smoke Harness", "## Purpose"} {
		if err := requireContains(guidePath, expected); err != nil {
			return err
		}
	}
	if err := requireMissing(filepath.Join(h.repo, "docs", ".obsidian", "INDEX.md")); err != nil {
		return err
	}
	if err := requireNotContains(filepath.Join(h.repo, "docs", "INDEX.md"), ".obsidian"); err != nil {
		return err
	}
	before, err := snapshot(filepath.Join(h.repo, "docs"))
	if err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "fix", "--docs", "--links"); err != nil {
		return err
	}
	after, err := snapshot(filepath.Join(h.repo, "docs"))
	if err != nil {
		return err
	}
	if !equalSnapshots(before, after) {
		return fmt.Errorf("second documentation fix changed repository documents")
	}
	return nil
}

func (h *harness) linkAndMoveScenario() error {
	h.step("verify observed rename repair and explicit moves")
	oldTarget := filepath.Join(h.repo, "docs", "target.md")
	renamedTarget := filepath.Join(h.repo, "docs", "target-renamed.md")
	if err := os.Rename(oldTarget, renamedTarget); err != nil {
		return err
	}
	if err := h.commandFails(h.repo, h.ddocs, "check", "--links"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "fix", "--links"); err != nil {
		return err
	}
	guide := filepath.Join(h.repo, "docs", "guide.md")
	if err := requireContains(guide, "target-renamed.md"); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(h.repo, "docs", "archive"), 0o755); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "mv", "docs/target-renamed.md", "docs/archive/target-final.md"); err != nil {
		return err
	}
	if err := requireContains(guide, "archive/target-final.md"); err != nil {
		return err
	}
	_, err := h.command(h.repo, h.ddocs, "check", "--links")
	return err
}

func (h *harness) reverseIndexScenario() error {
	h.step("verify reverse code-folder indexes")
	if _, err := h.command(h.repo, h.ddocs, "new", "general", "docs/code-map.md"); err != nil {
		return err
	}
	if err := appendFile(filepath.Join(h.repo, "docs", "code-map.md"), "\n## Code map\n\n- `src/worker.go`\n"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "format", "ignore", "--heading", "Code map", "docs/code-map.md"); err != nil {
		return err
	}
	if err := appendFile(filepath.Join(h.repo, "docs", "guide.md"), "\nSee [Code map](code-map.md).\n"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, h.ddocs, "fix", "--links", "--reverse"); err != nil {
		return err
	}
	index := filepath.Join(h.repo, "src", "INDEX.md")
	for _, expected := range []string{"worker.go", "code-map.md"} {
		if err := requireContains(index, expected); err != nil {
			return err
		}
	}
	return nil
}
