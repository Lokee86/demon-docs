package watch

import (
	"fmt"
	"io"
	"time"
)

const slowReconciliationThreshold = 2 * time.Second

type reconciliationObservation struct {
	started time.Time
	scope   string
	paths   int
}

func observeReconciliation(changedPaths []string, full bool) reconciliationObservation {
	scope := "scoped"
	if full || len(changedPaths) == 0 {
		scope = "full"
	}
	return reconciliationObservation{started: time.Now(), scope: scope, paths: len(changedPaths)}
}

func (o reconciliationObservation) failure(subsystem string, err error) error {
	return fmt.Errorf(
		"%s reconciliation failed duration=%s scope=%s paths=%d: %w",
		subsystem,
		formatWatchDuration(time.Since(o.started)),
		o.scope,
		o.paths,
		err,
	)
}

func (o reconciliationObservation) complete(out io.Writer, changed int) {
	writeReconciliationCompletion(out, changed, time.Since(o.started), o.scope, o.paths)
}

func writeReconciliationCompletion(out io.Writer, changed int, elapsed time.Duration, scope string, paths int) {
	duration := formatWatchDuration(elapsed)
	fmt.Fprintf(out, "%s ddocs watch updated %d file(s) duration=%s scope=%s paths=%d\n", timestamp(), changed, duration, scope, paths)
	if elapsed >= slowReconciliationThreshold {
		fmt.Fprintf(out, "%s ddocs watch slow reconciliation duration=%s scope=%s paths=%d\n", timestamp(), duration, scope, paths)
	}
}

func formatWatchDuration(duration time.Duration) time.Duration {
	if duration < time.Millisecond {
		return duration.Round(time.Microsecond)
	}
	return duration.Round(time.Millisecond)
}
