package watch

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"time"
)

func TestReconciliationCompletionReportsScopeDurationAndSlowRuns(t *testing.T) {
	var out bytes.Buffer
	writeReconciliationCompletion(&out, 3, 2500*time.Millisecond, "scoped", 17)
	text := out.String()
	for _, expected := range []string{
		"updated 3 file(s)",
		"duration=2.5s",
		"scope=scoped",
		"paths=17",
		"slow reconciliation",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("output missing %q: %s", expected, text)
		}
	}
}

func TestReconciliationFailureReportsSubsystemAndPreservesCause(t *testing.T) {
	observation := reconciliationObservation{
		started: time.Now().Add(-1500 * time.Millisecond),
		scope:   "full",
		paths:   4,
	}
	err := observation.failure("links", fs.ErrPermission)
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("wrapped error lost cause: %v", err)
	}
	for _, expected := range []string{"links reconciliation failed", "scope=full", "paths=4"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("error missing %q: %v", expected, err)
		}
	}
}
