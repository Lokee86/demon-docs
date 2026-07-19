package links

import (
	"fmt"
	"time"
)

// ReconcileTimings records the major phases of a reconciliation pass.
type ReconcileTimings struct {
	StateLoad      time.Duration
	InventoryBuild time.Duration
	Planning       time.Duration
	Total          time.Duration
}

func (t ReconcileTimings) String() string {
	return fmt.Sprintf(
		"state-load=%s inventory-build=%s planning=%s total=%s",
		t.StateLoad,
		t.InventoryBuild,
		t.Planning,
		t.Total,
	)
}

func (t *ReconcileTimings) add(other ReconcileTimings) {
	t.StateLoad += other.StateLoad
	t.InventoryBuild += other.InventoryBuild
	t.Planning += other.Planning
	t.Total += other.Total
}

// ApplyTimings records the major phases of applying and persisting a plan.
type ApplyTimings struct {
	FilesystemRewrite      time.Duration
	GeneratedSourceRefresh time.Duration
	DdocsPublication       time.Duration
	Total                  time.Duration
}

func (t ApplyTimings) String() string {
	return fmt.Sprintf(
		"filesystem-rewrite=%s generated-source-refresh=%s ddocs-publication=%s total=%s",
		t.FilesystemRewrite,
		t.GeneratedSourceRefresh,
		t.DdocsPublication,
		t.Total,
	)
}

func (t *ApplyTimings) add(other ApplyTimings) {
	t.FilesystemRewrite += other.FilesystemRewrite
	t.GeneratedSourceRefresh += other.GeneratedSourceRefresh
	t.DdocsPublication += other.DdocsPublication
	t.Total += other.Total
}
