package links

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ApplyMove verifies the complete plan, moves the source, and applies all
// generated Markdown rewrites. A failed rewrite triggers best-effort rollback.
func ApplyMove(plan MovePlan) error {
	if err := preflightMove(plan); err != nil {
		return err
	}
	if err := renameMovePath(plan.Source, plan.Destination); err != nil {
		return fmt.Errorf("move %s to %s: %w", plan.Source, plan.Destination, err)
	}
	rewrites := make([]GeneratedRewrite, len(plan.rewrites))
	for index := range plan.rewrites {
		rewrites[index] = plan.rewrites[index].rewrite
	}
	if _, err := ApplyGenerated(rewrites); err != nil {
		rollbackErr := rollbackMove(plan)
		if rollbackErr != nil {
			return fmt.Errorf("apply move rewrites: %w; rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("apply move rewrites: %w; move rolled back", err)
	}
	return nil
}

func preflightMove(plan MovePlan) error {
	info, err := os.Lstat(plan.Source)
	if err != nil {
		return fmt.Errorf("stat move source %s: %w", plan.Source, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || info.IsDir() != plan.SourceIsDirectory {
		return fmt.Errorf("move source changed after planning: %s", plan.Source)
	}
	if pathKey(plan.Source) != pathKey(plan.Destination) {
		if _, err := os.Stat(plan.Destination); err == nil {
			return fmt.Errorf("move destination appeared after planning: %s", plan.Destination)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat move destination %s: %w", plan.Destination, err)
		}
	}
	for _, item := range plan.rewrites {
		current, err := os.ReadFile(item.originPath)
		if err != nil {
			return fmt.Errorf("read planned move rewrite %s: %w", item.originPath, err)
		}
		if actual := sha256Digest(current); actual != item.rewrite.ExpectedOldSHA256 {
			return fmt.Errorf("planned move rewrite source changed before apply %s: expected %s, got %s", item.originPath, item.rewrite.ExpectedOldSHA256, actual)
		}
	}
	return nil
}

func renameMovePath(source, destination string) error {
	if pathKey(source) == pathKey(destination) && filepath.Clean(source) != filepath.Clean(destination) {
		temporary, err := os.MkdirTemp(filepath.Dir(source), ".ddocs-case-move-*")
		if err != nil {
			return err
		}
		if err := os.Remove(temporary); err != nil {
			return err
		}
		if err := os.Rename(source, temporary); err != nil {
			return err
		}
		if err := os.Rename(temporary, destination); err != nil {
			_ = os.Rename(temporary, source)
			return err
		}
		return nil
	}
	return os.Rename(source, destination)
}

func rollbackMove(plan MovePlan) error {
	var failures []string
	for _, item := range plan.rewrites {
		if _, err := os.Stat(item.rewrite.Path); err != nil {
			if !os.IsNotExist(err) {
				failures = append(failures, fmt.Sprintf("stat %s: %v", item.rewrite.Path, err))
			}
			continue
		}
		if err := replaceGenerated(item.rewrite.Path, item.rewrite.OldData(), item.mode); err != nil {
			failures = append(failures, fmt.Sprintf("restore %s: %v", item.rewrite.Path, err))
		}
	}
	if err := renameMovePath(plan.Destination, plan.Source); err != nil {
		failures = append(failures, fmt.Sprintf("restore move: %v", err))
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}
