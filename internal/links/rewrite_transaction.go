package links

import (
	"errors"
	"fmt"
	"os"
)

// RollbackGenerated restores a successfully applied batch. Every file must
// still contain the generated after-state; otherwise the rollback refuses to
// overwrite the newer content.
func RollbackGenerated(rewrites []GeneratedRewrite) error {
	rollback := make([]GeneratedRewrite, len(rewrites))
	for index, rewrite := range rewrites {
		rollback[index] = newGeneratedRewriteBytes(rewrite.SourceFileID, rewrite.Path, rewrite.newData, rewrite.oldData, nil)
	}
	if _, err := ApplyGenerated(rollback); err != nil {
		return fmt.Errorf("rollback generated rewrites: %w", err)
	}
	return nil
}

func generatedApplyFailure(pending []pendingRewrite, attempted []int, applyErr error) error {
	if rollbackErr := rollbackPendingGenerated(pending, attempted); rollbackErr != nil {
		return errors.Join(applyErr, fmt.Errorf("rollback generated rewrite batch: %w", rollbackErr))
	}
	return applyErr
}

func rollbackPendingGenerated(pending []pendingRewrite, attempted []int) error {
	var rollbackErrors []error
	for position := len(attempted) - 1; position >= 0; position-- {
		item := pending[attempted[position]]
		current, err := os.ReadFile(item.path)
		if err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Errorf("read generated rewrite rollback source %s: %w", item.path, err))
			continue
		}
		digest := sha256Digest(current)
		switch digest {
		case item.rewrite.ExpectedOldSHA256:
			continue
		case item.rewrite.ExpectedNewSHA256:
			if err := replaceGenerated(item.path, item.rewrite.oldData, item.mode); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restore generated rewrite %s: %w", item.path, err))
				continue
			}
			restored, err := os.ReadFile(item.path)
			if err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("verify restored generated rewrite %s: %w", item.path, err))
				continue
			}
			if actual := sha256Digest(restored); actual != item.rewrite.ExpectedOldSHA256 {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restored generated rewrite hash mismatch %s: expected %s, got %s", item.path, item.rewrite.ExpectedOldSHA256, actual))
			}
		default:
			rollbackErrors = append(rollbackErrors, fmt.Errorf("refuse to rollback changed generated rewrite %s: expected %s or %s, got %s", item.path, item.rewrite.ExpectedOldSHA256, item.rewrite.ExpectedNewSHA256, digest))
		}
	}
	return errors.Join(rollbackErrors...)
}
