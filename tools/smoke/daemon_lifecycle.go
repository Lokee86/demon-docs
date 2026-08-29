package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (h *harness) daemonLifecycleScenario() error {
	const cycles = 5
	h.step("verify repeated detached daemon lifecycle")
	if err := h.waitDemonStoppedAt(h.repo); err != nil {
		return h.daemonFailure(err)
	}
	for cycle := 1; cycle <= cycles; cycle++ {
		if _, err := h.command(h.repo, h.ddocs, "demon", "run", "--true"); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d start: %w", cycle, err))
		}
		if err := h.waitDemonStateAt(h.repo, "running"); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d running state: %w", cycle, err))
		}
		acquired, err := h.command(h.repo, h.ddocs, "demon", "acquire", "--client", fmt.Sprintf("smoke-cycle-%d", cycle))
		if err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d acquire: %w", cycle, err))
		}
		token := fieldValue(acquired, "token")
		if token == "" {
			return h.daemonFailure(fmt.Errorf("cycle %d acquire returned no token: %q", cycle, acquired))
		}
		if _, err := h.command(h.repo, h.ddocs, "demon", "heartbeat", "--token", token); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d heartbeat: %w", cycle, err))
		}
		if err := h.waitDemonStatusContainsAt(h.repo, "active agents: 1"); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d agent attach: %w", cycle, err))
		}
		if cycle == 1 {
			entered, enterErr := h.command(h.repo, h.ddocs, "demon", "__enter", h.repo, "shell")
			if enterErr != nil {
				return h.daemonFailure(fmt.Errorf("shell hook enter: %w", enterErr))
			}
			shellToken := fieldValue(entered, "token")
			if shellToken == "" {
				return h.daemonFailure(fmt.Errorf("shell hook enter returned no token: %q", entered))
			}
			if _, leaveErr := h.command(h.repo, h.ddocs, "demon", "__leave", h.repo, shellToken); leaveErr != nil {
				return h.daemonFailure(fmt.Errorf("shell hook leave: %w", leaveErr))
			}
			if err := h.waitDemonStatusContainsAt(h.repo, "active shells: 0"); err != nil {
				return h.daemonFailure(fmt.Errorf("shell hook leave cleanup: %w", err))
			}
		}
		if _, err := h.command(h.repo, h.ddocs, "demon", "release", "--token", token); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d release: %w", cycle, err))
		}
		if err := h.waitDemonStatusContainsAt(h.repo, "active agents: 0"); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d agent detach: %w", cycle, err))
		}
		if _, err := h.command(h.repo, h.ddocs, "demon", "run", "--false"); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d stop: %w", cycle, err))
		}
		if err := h.waitDemonStoppedAt(h.repo); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d stopped state: %w", cycle, err))
		}
		if err := h.waitCleanDemonRuntimeAt(h.repo); err != nil {
			return h.daemonFailure(fmt.Errorf("cycle %d runtime cleanup: %w", cycle, err))
		}
	}
	return nil
}

func (h *harness) waitDemonStateAt(root, state string) error {
	return waitFor("demon state "+state, 15*time.Second, func() bool {
		status, err := h.command(root, h.ddocs, "demon", "--status", root)
		return err == nil && strings.Contains(status, "demon: "+state)
	})
}

func (h *harness) waitDemonStoppedAt(root string) error {
	return h.waitDemonStateAt(root, "stopped")
}

func (h *harness) waitDemonStatusContainsAt(root, expected string) error {
	return waitFor("demon status "+expected, 15*time.Second, func() bool {
		status, err := h.command(root, h.ddocs, "demon", "--status", root)
		return err == nil && strings.Contains(status, expected)
	})
}

func (h *harness) waitCleanDemonRuntimeAt(root string) error {
	if err := waitFor("demon runtime cleanup", 5*time.Second, func() bool {
		return requireCleanDemonRuntime(root) == nil
	}); err != nil {
		if cleanupErr := requireCleanDemonRuntime(root); cleanupErr != nil {
			return cleanupErr
		}
		return err
	}
	return nil
}

func requireCleanDemonRuntime(root string) error {
	runtimeRoot := filepath.Join(root, ".ddocs", "runtime")
	for _, name := range []string{"owner.json", "owner-heartbeat", "ready.json", "shutdown-request", "owner.lock"} {
		if _, err := os.Stat(filepath.Join(runtimeRoot, name)); !os.IsNotExist(err) {
			return fmt.Errorf("stale demon runtime artifact remains: %s", name)
		}
	}
	feeders := filepath.Join(runtimeRoot, "feeders")
	entries, err := os.ReadDir(feeders)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			return fmt.Errorf("stale demon feeder remains: %s", entry.Name())
		}
	}
	return nil
}

func fieldValue(line, key string) string {
	prefix := key + "="
	for _, field := range strings.Fields(line) {
		if strings.HasPrefix(field, prefix) {
			return strings.TrimPrefix(field, prefix)
		}
	}
	return ""
}
