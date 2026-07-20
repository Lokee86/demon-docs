package links

import (
	"fmt"

	"github.com/Lokee86/demon-docs/internal/filetxn"
)

// RollbackGenerated restores a successfully applied batch. Every file must
// still contain the generated after-state; otherwise the rollback refuses to
// overwrite the newer content.
func RollbackGenerated(rewrites []GeneratedRewrite) error {
	transactions := make([]filetxn.Rewrite, len(rewrites))
	for index, rewrite := range rewrites {
		if !rewrite.transaction.Prepared() {
			return fmt.Errorf("rollback generated rewrite %s was not prepared", rewrite.Path)
		}
		transactions[index] = rewrite.transaction
	}
	if err := filetxn.Rollback(transactions); err != nil {
		return fmt.Errorf("rollback generated rewrites: %w", err)
	}
	return nil
}
