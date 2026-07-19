package links

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestExternalFixtureLinkTiming(t *testing.T) {
	source := os.Getenv("DDOCS_LINK_TIMING_FIXTURE")
	if source == "" {
		t.Skip("DDOCS_LINK_TIMING_FIXTURE is not set")
	}

	source, err := filepath.Abs(source)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(source)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("fixture path %q is not a regular directory", source)
	}

	root := filepath.Join(t.TempDir(), "fixture-copy")
	copyExternalFixture(t, source, root)

	baseline, baselineReconcileTimings, err := reconcileWithTimings(root)
	if err != nil {
		t.Fatal(err)
	}
	assertReconcileTimingInvariants(t, baselineReconcileTimings)
	baselineApplied, baselineApplyTimings, err := applyAndSaveWithTimings(&baseline)
	if err != nil {
		t.Fatal(err)
	}
	assertApplyTimingInvariants(t, baselineApplyTimings)

	ready, readyTimings, err := reconcileWithTimings(root)
	if err != nil {
		t.Fatal(err)
	}
	assertReconcileTimingInvariants(t, readyTimings)
	if len(ready.Rewrites) != 0 {
		t.Fatalf("fixture baseline did not converge; %d rewrites remain", len(ready.Rewrites))
	}

	targetRecord, inbound, ok := mostLinkedFixtureFile(ready)
	if !ok {
		t.Skip("fixture contains no uniquely identifiable repository file with inbound links")
	}
	target := recordAbsolute(root, targetRecord)
	renamed := availableTimingRenamePath(t, target)
	if err := os.Rename(target, renamed); err != nil {
		t.Fatal(err)
	}

	plan, reconcileTimings, err := reconcileWithTimings(root)
	if err != nil {
		t.Fatal(err)
	}
	assertReconcileTimingInvariants(t, reconcileTimings)
	if len(plan.Rewrites) == 0 {
		t.Fatalf("renaming %q with %d inbound links produced no rewrites", target, inbound)
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

	t.Logf("baseline reconcile:  %s", baselineReconcileTimings)
	t.Logf("baseline apply:      %s; rewrites=%d", baselineApplyTimings, baselineApplied)
	t.Logf("stable reconcile:    %s", readyTimings)
	t.Logf("rename target:       %s; inbound-links=%d", targetRecord.Path, inbound)
	t.Logf("rename reconcile:    %s; rewrites=%d updates=%d", reconcileTimings, len(plan.Rewrites), len(plan.Updates))
	t.Logf("rename apply:        %s; rewrites=%d", applyTimings, applied)
	t.Logf("follow-up reconcile: %s", followupReconcileTimings)
	t.Logf("follow-up apply:     %s", followupApplyTimings)
}

func mostLinkedFixtureFile(plan Plan) (FileRecord, int, bool) {
	fingerprintCounts := make(map[string]int)
	for _, record := range plan.Files.Files {
		if record.Present && record.Scope == "repository" && record.Kind == "file" && record.Fingerprint != "" {
			fingerprintCounts[record.Fingerprint]++
		}
	}

	candidates := make([]FileRecord, 0)
	candidateIDs := make(map[string]bool)
	for _, record := range plan.Files.Files {
		if record.Present && record.Scope == "repository" && record.Kind == "file" &&
			record.Fingerprint != "" && fingerprintCounts[record.Fingerprint] == 1 {
			candidates = append(candidates, record)
			candidateIDs[record.ID] = true
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Path != candidates[j].Path {
			return candidates[i].Path < candidates[j].Path
		}
		return candidates[i].ID < candidates[j].ID
	})

	inbound := make(map[string]int)
	for _, link := range plan.Links.Links {
		if candidateIDs[link.TargetFileID] && (link.Status == "valid" || link.Status == "moved") {
			inbound[link.TargetFileID]++
		}
	}

	var selected FileRecord
	selectedCount := 0
	for _, candidate := range candidates {
		count := inbound[candidate.ID]
		if count > selectedCount {
			selected = candidate
			selectedCount = count
		}
	}
	if selectedCount == 0 {
		return FileRecord{}, 0, false
	}
	return selected, selectedCount, true
}

func availableTimingRenamePath(t *testing.T, path string) string {
	t.Helper()
	extension := filepath.Ext(path)
	base := strings.TrimSuffix(path, extension)
	for suffix := 1; ; suffix++ {
		candidate := base + ".timing-renamed"
		if suffix > 1 {
			candidate += "-" + strconv.Itoa(suffix)
		}
		candidate += extension
		_, err := os.Lstat(candidate)
		if os.IsNotExist(err) {
			return candidate
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}
