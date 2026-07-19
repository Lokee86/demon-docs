package links

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReconciliationAndApplicationTimings(t *testing.T) {
	root := t.TempDir()
	createBenchmarkRepository(t, root, 96)

	baseline, baselineReconcileTimings, err := reconcileWithTimings(root)
	if err != nil {
		t.Fatal(err)
	}
	assertReconcileTimingInvariants(t, baselineReconcileTimings)

	if len(baseline.Rewrites) != 0 {
		t.Fatalf("baseline produced %d rewrites; want zero", len(baseline.Rewrites))
	}

	baselineApplied, baselineApplyTimings, err := applyAndSaveWithTimings(&baseline)
	if err != nil {
		t.Fatal(err)
	}
	assertApplyTimingInvariants(t, baselineApplyTimings)
	if baselineApplied != 0 {
		t.Fatalf("baseline applied %d rewrites; want zero", baselineApplied)
	}

	oldTarget := filepath.Join(root, "asset-a.bin")
	newTarget := filepath.Join(root, "asset-renamed.bin")
	if err := os.Rename(oldTarget, newTarget); err != nil {
		t.Fatal(err)
	}

	plan, reconcileTimings, err := reconcileWithTimings(root)
	if err != nil {
		t.Fatal(err)
	}
	assertReconcileTimingInvariants(t, reconcileTimings)
	if len(plan.Rewrites) == 0 {
		t.Fatal("rename produced no rewrites")
	}
	if len(plan.Updates) == 0 {
		t.Fatal("rename produced no updates")
	}

	applied, applyTimings, err := applyAndSaveWithTimings(&plan)
	if err != nil {
		t.Fatal(err)
	}
	assertApplyTimingInvariants(t, applyTimings)
	if applied != len(plan.Rewrites) {
		t.Fatalf("applied %d rewrites; plan contains %d", applied, len(plan.Rewrites))
	}

	followup, followupReconcileTimings, err := reconcileWithTimings(root)
	if err != nil {
		t.Fatal(err)
	}
	assertReconcileTimingInvariants(t, followupReconcileTimings)
	if len(followup.Rewrites) != 0 || len(followup.Updates) != 0 {
		t.Fatalf("idempotent follow-up produced %d rewrites and %d updates", len(followup.Rewrites), len(followup.Updates))
	}

	followupApplied, followupApplyTimings, err := applyAndSaveWithTimings(&followup)
	if err != nil {
		t.Fatal(err)
	}
	assertApplyTimingInvariants(t, followupApplyTimings)
	if followupApplied != 0 {
		t.Fatalf("idempotent follow-up applied %d rewrites; want zero", followupApplied)
	}

	t.Logf("baseline reconcile: %s", baselineReconcileTimings)
	t.Logf("baseline apply:     %s", baselineApplyTimings)
	t.Logf("move reconcile:     %s; rewrites=%d", reconcileTimings, len(plan.Rewrites))
	t.Logf("move apply:         %s; rewrites=%d", applyTimings, applied)
	t.Logf("follow-up reconcile: %s", followupReconcileTimings)
	t.Logf("follow-up apply:     %s", followupApplyTimings)
}

func TestTimingStringOrdering(t *testing.T) {
	assertStringOrdering(t, (ReconcileTimings{}).String(), []string{
		"state-load=",
		"inventory-build=",
		"planning=",
		"total=",
	})
	assertStringOrdering(t, (ApplyTimings{}).String(), []string{
		"filesystem-rewrite=",
		"generated-source-refresh=",
		"ddocs-publication=",
		"total=",
	})
}

func assertReconcileTimingInvariants(t *testing.T, timings ReconcileTimings) {
	t.Helper()
	for index, phase := range []time.Duration{timings.StateLoad, timings.InventoryBuild, timings.Planning} {
		if phase < 0 {
			t.Fatalf("reconcile phase %d has negative duration %s", index, phase)
		}
	}
	if timings.Total < timings.StateLoad+timings.InventoryBuild+timings.Planning {
		t.Fatalf("reconcile total %s is less than phase sum", timings.Total)
	}
	assertStringOrdering(t, timings.String(), []string{"state-load=", "inventory-build=", "planning=", "total="})
}

func assertApplyTimingInvariants(t *testing.T, timings ApplyTimings) {
	t.Helper()
	for index, phase := range []time.Duration{timings.FilesystemRewrite, timings.GeneratedSourceRefresh, timings.DdocsPublication} {
		if phase < 0 {
			t.Fatalf("apply phase %d has negative duration %s", index, phase)
		}
	}
	if timings.Total < timings.FilesystemRewrite+timings.GeneratedSourceRefresh+timings.DdocsPublication {
		t.Fatalf("apply total %s is less than phase sum", timings.Total)
	}
	assertStringOrdering(t, timings.String(), []string{"filesystem-rewrite=", "generated-source-refresh=", "ddocs-publication=", "total="})
}

func assertStringOrdering(t *testing.T, value string, fields []string) {
	t.Helper()
	previous := -1
	for _, field := range fields {
		current := strings.Index(value, field)
		if current < 0 {
			t.Fatalf("timing summary %q does not contain %q", value, field)
		}
		if current <= previous {
			t.Fatalf("timing summary fields are out of order: %q", value)
		}
		previous = current
	}
}
