package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *harness) daemonLinkedWorktreeScenario() error {
	h.step("verify linked-worktree demon isolation")
	if err := h.prepareGitWorktreeFixture(); err != nil {
		return err
	}
	linked := filepath.Join(h.workspace, "linked-worktree")
	if _, err := h.command(h.repo, "git", "worktree", "add", "-b", "smoke-linked", linked); err != nil {
		return err
	}
	if err := requireMissing(filepath.Join(linked, ".ddocs")); err != nil {
		return fmt.Errorf("linked worktree was not initially isolated: %w", err)
	}
	status, err := h.command(linked, h.ddocs, "demon", "--status", linked)
	if err != nil {
		return err
	}
	if !strings.Contains(status, "demon: stopped") {
		return fmt.Errorf("unexpected pre-bootstrap linked-worktree status: %s", status)
	}
	if err := requireMissing(filepath.Join(linked, ".ddocs")); err != nil {
		return fmt.Errorf("read-only status bootstrapped linked worktree: %w", err)
	}
	if _, err := h.command(linked, h.ddocs, "demon", "run", "--true", linked); err != nil {
		return err
	}
	if err := h.waitDemonStateAt(linked, "running"); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(h.repo, ".ddocs", "runtime", "owner.json")); !os.IsNotExist(err) {
		return fmt.Errorf("linked-worktree demon reused primary owner state")
	}
	created := filepath.Join(linked, "docs", "linked-daemon-created.md")
	if err := writeFile(created, "# Linked Daemon Created\n"); err != nil {
		return err
	}
	if err := waitFor("linked-worktree daemon index insertion", 15*time.Second, func() bool {
		return fileContains(filepath.Join(linked, "docs", "INDEX.md"), "linked-daemon-created.md")
	}); err != nil {
		return err
	}
	if _, err := h.command(linked, h.ddocs, "demon", "run", "--false", linked); err != nil {
		return err
	}
	if err := h.waitDemonStoppedAt(linked); err != nil {
		return err
	}
	if err := h.waitCleanDemonRuntimeAt(linked); err != nil {
		return err
	}
	return requireCleanDemonRuntime(h.repo)
}

func (h *harness) prepareGitWorktreeFixture() error {
	if _, err := h.command(h.repo, "git", "init"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, "git", "config", "user.name", "Archivist Smoke"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, "git", "config", "user.email", "smoke@example.invalid"); err != nil {
		return err
	}
	if _, err := h.command(h.repo, "git", "add", "docs", "src"); err != nil {
		return err
	}
	_, err := h.command(h.repo, "git", "commit", "-m", "smoke linked-worktree baseline")
	return err
}
