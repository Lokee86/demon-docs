package filetxn

import (
	"errors"
	"fmt"
	"os"
)

func Rollback(rewrites []Rewrite) error {
	rollback := make([]Rewrite, len(rewrites))
	for index, rewrite := range rewrites {
		rollback[index] = New(rewrite.path, rewrite.newData, rewrite.oldData)
	}
	if _, err := Apply(rollback); err != nil {
		return fmt.Errorf("rollback rewrites: %w", err)
	}
	return nil
}

func applyFailure(pending []pendingRewrite, attempted []int, applyErr error) error {
	if rollbackErr := rollbackPending(pending, attempted); rollbackErr != nil {
		return errors.Join(applyErr, fmt.Errorf("rollback rewrite batch: %w", rollbackErr))
	}
	return applyErr
}

func rollbackPending(pending []pendingRewrite, attempted []int) error {
	var rollbackErrors []error
	for position := len(attempted) - 1; position >= 0; position-- {
		item := pending[attempted[position]]
		current, err := os.ReadFile(item.path)
		if err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("read rewrite rollback source %s: %w", item.path, err))
			continue
		}
		digest := Digest(current)
		switch digest {
		case item.rewrite.expectedOldSHA256:
			continue
		case item.rewrite.expectedNewSHA256:
			if err := Replace(item.path, item.rewrite.oldData, item.mode); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restore rewrite %s: %w", item.path, err))
				continue
			}
			restored, err := os.ReadFile(item.path)
			if err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("verify restored rewrite %s: %w", item.path, err))
				continue
			}
			if actual := Digest(restored); actual != item.rewrite.expectedOldSHA256 {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restored rewrite hash mismatch %s: expected %s, got %s", item.path, item.rewrite.expectedOldSHA256, actual))
			}
		default:
			rollbackErrors = append(rollbackErrors, fmt.Errorf("refuse to rollback changed rewrite %s: expected %s or %s, got %s", item.path, item.rewrite.expectedOldSHA256, item.rewrite.expectedNewSHA256, digest))
		}
	}
	return errors.Join(rollbackErrors...)
}
