package demon

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func testRuntime(t *testing.T) *Runtime {
	t.Helper()
	r := New(t.TempDir())
	r.Timing = Timing{FeederHeartbeat: 10 * time.Millisecond, FeederExpiry: time.Second, ShutdownGrace: time.Second, OwnerLease: 10 * time.Second}
	return r
}

func TestClaimAllowsExactlyOneOwner(t *testing.T) {
	r := testRuntime(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	claimed := 0
	owners := make([]Owner, 0, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			owner, won, err := r.Claim(os.Getpid())
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			mu.Lock()
			if won {
				claimed++
			}
			owners = append(owners, owner)
			mu.Unlock()
		}()
	}
	wg.Wait()
	if claimed != 1 {
		t.Fatalf("claimed=%d owners=%v", claimed, owners)
	}
	owner, err := r.ReadOwner()
	if err != nil || owner.Token == "" {
		t.Fatalf("owner=%+v err=%v", owner, err)
	}
	if err := r.Release(owner); err != nil {
		t.Fatal(err)
	}
}

func TestTokenSafeReleaseAndStaleRecovery(t *testing.T) {
	r := testRuntime(t)
	first, won, err := r.Claim(1)
	if err != nil || !won {
		t.Fatalf("first claim: %+v %v %t", first, err, won)
	}
	if err := r.Release(Owner{Token: "other"}); err == nil {
		t.Fatal("mismatched token released ownership")
	}
	first.Heartbeat = time.Now().Add(-time.Hour)
	if err := atomicJSON(r.Paths.Owner, first); err != nil {
		t.Fatal(err)
	}
	second, won, err := r.Claim(2)
	if err != nil || !won || second.Token == first.Token {
		t.Fatalf("stale recovery failed: %+v %v %t", second, err, won)
	}
	if err := r.Release(second); err != nil {
		t.Fatal(err)
	}
}

func TestFeederExpiryAndKindCounts(t *testing.T) {
	r := testRuntime(t)
	shell, err := r.AddFeeder("shell", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	agent, err := r.AddFeeder("agent", 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	feeders, err := r.SnapshotFeeders()
	if err != nil || len(feeders) != 2 {
		t.Fatalf("feeders=%+v err=%v", feeders, err)
	}
	shells, agents := CountKinds(feeders)
	if shells != 1 || agents != 1 {
		t.Fatalf("counts=%d,%d", shells, agents)
	}
	if err := r.RemoveFeeder(shell.Token); err != nil {
		t.Fatal(err)
	}
	_ = agent
	r.Timing.FeederExpiry = 20 * time.Millisecond
	time.Sleep(40 * time.Millisecond)
	feeders, err = r.ListFeeders()
	if err != nil || len(feeders) != 0 {
		t.Fatalf("expired feeders=%+v err=%v", feeders, err)
	}
}

func TestStatusSnapshotDoesNotDeleteExpiredFeeder(t *testing.T) {
	r := testRuntime(t)
	feeder, err := r.AddFeeder("shell", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(r.Paths.Feeders, feeder.Token+".json")
	old := feeder
	old.Heartbeat = time.Now().Add(-time.Hour)
	if err := atomicJSON(path, old); err != nil {
		t.Fatal(err)
	}
	if _, err := r.SnapshotFeeders(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot mutated feeder state: %v", err)
	}
}

func TestSnapshotMissingRuntimeIsReadOnly(t *testing.T) {
	r := testRuntime(t)
	if _, err := r.SnapshotFeeders(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(r.Paths.Runtime); !os.IsNotExist(err) {
		t.Fatalf("snapshot created runtime directory: %v", err)
	}
}

func TestFindFeederPreventsDuplicateShellSessions(t *testing.T) {
	r := testRuntime(t)
	first, err := r.AddFeeder("shell", 1, 99)
	if err != nil {
		t.Fatal(err)
	}
	found, ok := r.FindFeeder("shell", 99)
	if !ok || found.Token != first.Token {
		t.Fatalf("did not find existing shell feeder: %+v %t", found, ok)
	}
	if _, ok := r.FindFeeder("agent", 99); ok {
		t.Fatal("shell feeder was returned for an agent lookup")
	}
}

func TestRotatingLogIsBounded(t *testing.T) {
	r := testRuntime(t)
	log, err := OpenLog(r.Paths)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	for i := 0; i < LogFiles+2; i++ {
		if _, err := log.Write(make([]byte, MaxLogBytes)); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(r.Paths.Logs)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		info, _ := entry.Info()
		if info.Size() > MaxLogBytes {
			t.Fatalf("log %s too large: %d", entry.Name(), info.Size())
		}
	}
	if len(entries) > LogFiles {
		t.Fatalf("retained %d log files", len(entries))
	}
}

func TestServeStopsAfterGraceWithoutFeeders(t *testing.T) {
	r := testRuntime(t)
	r.Timing.ShutdownGrace = 30 * time.Millisecond
	owner, won, err := r.Claim(1)
	if err != nil || !won {
		t.Fatal(err)
	}
	watchStarted := make(chan struct{})
	err = r.Serve(context.Background(), owner, func() (bool, error) { return true, nil }, func(ctx context.Context, _ io.Writer) error {
		close(watchStarted)
		<-ctx.Done()
		return nil
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-watchStarted:
	default:
		t.Fatal("watcher was not started")
	}
}
