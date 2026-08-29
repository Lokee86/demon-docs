package watch

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/Lokee86/demon-docs/internal/links"
)

const permissionRetryLimit = 5

type reconciliationRetryPolicy struct {
	permissionFailures int
}

func (p *reconciliationRetryPolicy) classify(err error) (bool, string) {
	if links.IsTransientFilesystemRace(err) {
		return true, "stale reconciliation plan"
	}
	if errors.Is(err, fs.ErrPermission) {
		p.permissionFailures++
		if p.permissionFailures <= permissionRetryLimit {
			return true, fmt.Sprintf("transient filesystem access error (%d/%d)", p.permissionFailures, permissionRetryLimit)
		}
	}
	return false, ""
}

func (p *reconciliationRetryPolicy) succeeded() {
	p.permissionFailures = 0
}
