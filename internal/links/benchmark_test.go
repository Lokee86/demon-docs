package links

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkInitialIndexing(b *testing.B) {
	b.ReportAllocs()
	var reconcileTotals ReconcileTimings
	var applyTotals ApplyTimings

	for iteration := 0; iteration < b.N; iteration++ {
		b.StopTimer()
		root := filepath.Join(b.TempDir(), fmt.Sprintf("repo-%d", iteration))
		createBenchmarkRepository(b, root, 250)
		b.StartTimer()

		plan, reconcileTimings, err := reconcileWithTimings(root)
		if err != nil {
			b.Fatal(err)
		}
		_, applyTimings, err := applyAndSaveWithTimings(&plan)
		if err != nil {
			b.Fatal(err)
		}
		reconcileTotals.add(reconcileTimings)
		applyTotals.add(applyTimings)
	}

	reportReconcileTimingMetrics(b, reconcileTotals)
	reportApplyTimingMetrics(b, applyTotals)
}

func BenchmarkSingleFileIncrementalUpdate(b *testing.B) {
	root := b.TempDir()
	createBenchmarkRepository(b, root, 500)
	baseline, err := Reconcile(root)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		b.Fatal(err)
	}
	changedPath := filepath.Join(root, "docs", "page-0250.md")
	b.ReportAllocs()
	b.ResetTimer()

	var reconcileTotals ReconcileTimings
	var applyTotals ApplyTimings
	for iteration := 0; iteration < b.N; iteration++ {
		target := "asset-a.bin"
		if iteration%2 == 1 {
			target = "asset-b.bin"
		}
		content := fmt.Sprintf("# Changed\n\n[target](../%s)\n", target)
		if err := os.WriteFile(changedPath, []byte(content), 0o644); err != nil {
			b.Fatal(err)
		}
		plan, reconcileTimings, err := reconcileWithTimings(root)
		if err != nil {
			b.Fatal(err)
		}
		_, applyTimings, err := applyAndSaveWithTimings(&plan)
		if err != nil {
			b.Fatal(err)
		}
		reconcileTotals.add(reconcileTimings)
		applyTotals.add(applyTimings)
	}

	reportReconcileTimingMetrics(b, reconcileTotals)
	reportApplyTimingMetrics(b, applyTotals)
}

func BenchmarkScopedSingleFileIncrementalUpdate(b *testing.B) {
	root := b.TempDir()
	createBenchmarkRepository(b, root, 500)
	baseline, err := Reconcile(root)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		b.Fatal(err)
	}
	changedPath := filepath.Join(root, "docs", "page-0250.md")
	b.ReportAllocs()
	b.ResetTimer()

	for iteration := 0; iteration < b.N; iteration++ {
		target := "asset-a.bin"
		if iteration%2 == 1 {
			target = "asset-b.bin"
		}
		content := fmt.Sprintf("# Changed\n\n[target](../%s)\n", target)
		if err := os.WriteFile(changedPath, []byte(content), 0o644); err != nil {
			b.Fatal(err)
		}
		plan, err := ReconcileChangedPaths(root, []string{changedPath})
		if err != nil {
			b.Fatal(err)
		}
		if _, err := ApplyAndSave(&plan); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObservedRenameImmediateRepair(b *testing.B) {
	root := b.TempDir()
	source := filepath.Join(root, "docs", "source.md")
	targetA := filepath.Join(root, "docs", "target-a.md")
	targetB := filepath.Join(root, "docs", "target-b.md")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("[Target](target-a.md)\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(targetA, []byte("# Target\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	baseline, err := Reconcile(root)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := ApplyAndSave(&baseline); err != nil {
		b.Fatal(err)
	}

	current, next := targetA, targetB
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if iteration > 0 {
			b.StopTimer()
			plan, err := Reconcile(root)
			if err != nil {
				b.Fatal(err)
			}
			if _, err := ApplyAndSave(&plan); err != nil {
				b.Fatal(err)
			}
			b.StartTimer()
		}
		if err := os.Rename(current, next); err != nil {
			b.Fatal(err)
		}
		handled, changed, err := RepairObservedRename(root, current, next)
		if err != nil {
			b.Fatal(err)
		}
		if !handled || changed == 0 {
			b.Fatalf("observed rename was not repaired: handled=%v changed=%d", handled, changed)
		}
		current, next = next, current
	}
}

func BenchmarkHighFanoutTargetMove(b *testing.B) {
	b.ReportAllocs()
	var reconcileTotals ReconcileTimings
	var applyTotals ApplyTimings
	appliedTotal := 0

	for iteration := 0; iteration < b.N; iteration++ {
		b.StopTimer()
		root := filepath.Join(b.TempDir(), fmt.Sprintf("repo-%d", iteration))
		createBenchmarkRepository(b, root, 250)
		baseline, err := Reconcile(root)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := ApplyAndSave(&baseline); err != nil {
			b.Fatal(err)
		}
		if err := os.Rename(filepath.Join(root, "asset-a.bin"), filepath.Join(root, "asset-moved.bin")); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		plan, reconcileTimings, err := reconcileWithTimings(root)
		if err != nil {
			b.Fatal(err)
		}
		applied, applyTimings, err := applyAndSaveWithTimings(&plan)
		if err != nil {
			b.Fatal(err)
		}
		if applied == 0 {
			b.Fatal("target move produced no rewrites")
		}
		appliedTotal += applied
		reconcileTotals.add(reconcileTimings)
		applyTotals.add(applyTimings)
	}

	reportReconcileTimingMetrics(b, reconcileTotals)
	reportApplyTimingMetrics(b, applyTotals)
	b.ReportMetric(float64(appliedTotal)/float64(b.N), "rewrites/op")
}

func createBenchmarkRepository(tb testing.TB, root string, count int) {
	tb.Helper()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		tb.Fatal(err)
	}
	for _, name := range []string{"asset-a.bin", "asset-b.bin"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
	for index := 0; index < count; index++ {
		path := filepath.Join(root, "docs", fmt.Sprintf("page-%04d.md", index))
		content := fmt.Sprintf("# Page %d\n\n[target](../asset-a.bin)\n", index)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
}

func reportReconcileTimingMetrics(b *testing.B, totals ReconcileTimings) {
	b.ReportMetric(float64(totals.StateLoad)/float64(b.N), "reconcile-state-load-ns/op")
	b.ReportMetric(float64(totals.InventoryBuild)/float64(b.N), "reconcile-inventory-build-ns/op")
	b.ReportMetric(float64(totals.Planning)/float64(b.N), "reconcile-planning-ns/op")
	b.ReportMetric(float64(totals.Total)/float64(b.N), "reconcile-total-ns/op")
}

func reportApplyTimingMetrics(b *testing.B, totals ApplyTimings) {
	b.ReportMetric(float64(totals.FilesystemRewrite)/float64(b.N), "apply-filesystem-rewrite-ns/op")
	b.ReportMetric(float64(totals.GeneratedSourceRefresh)/float64(b.N), "apply-generated-source-refresh-ns/op")
	b.ReportMetric(float64(totals.DdocsPublication)/float64(b.N), "apply-ddocs-publication-ns/op")
	b.ReportMetric(float64(totals.Total)/float64(b.N), "apply-total-ns/op")
}
