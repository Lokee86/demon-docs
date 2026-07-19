package demon

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type claimResult struct {
	owner Owner
	won   bool
	err   error
}

func TestClaimRecoversAbandonedOwnerLock(t *testing.T) {
	runtime := testRuntime(t)
	if err := runtime.ensureRuntime(); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(runtime.Paths.Runtime, ownerLockDir)
	if err := os.Mkdir(lockPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lockPath, ownerLockTokenFile), []byte("abandoned\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-ownerLockStaleAfter - time.Second)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatal(err)
	}
	owner, won, err := runtime.Claim(os.Getpid())
	if err != nil || !won {
		t.Fatalf("claim after abandoned lock: won=%t err=%v", won, err)
	}
	if err := runtime.Release(owner); err != nil {
		t.Fatal(err)
	}
}

func TestClaimContentionStress(t *testing.T) {
	const (
		rounds     = 20
		contenders = 24
	)
	for round := 0; round < rounds; round++ {
		root := t.TempDir()
		start := make(chan struct{})
		results := make(chan claimResult, contenders)
		var ready sync.WaitGroup
		ready.Add(contenders)
		for contender := 0; contender < contenders; contender++ {
			runtime := New(root)
			runtime.Timing.OwnerLease = time.Minute
			go func(pid int) {
				ready.Done()
				<-start
				owner, won, err := runtime.Claim(pid)
				results <- claimResult{owner: owner, won: won, err: err}
			}(os.Getpid() + contender + 1)
		}
		ready.Wait()
		close(start)

		won := 0
		claims := make([]claimResult, 0, contenders)
		for contender := 0; contender < contenders; contender++ {
			result := <-results
			if result.err != nil {
				t.Fatalf("round %d claim failed: %v", round, result.err)
			}
			if result.won {
				won++
			}
			claims = append(claims, result)
		}
		if won != 1 {
			t.Fatalf("round %d winners=%d claims=%+v", round, won, claims)
		}
		persisted, err := New(root).ReadOwner()
		if err != nil {
			t.Fatalf("round %d read owner: %v", round, err)
		}
		for _, result := range claims {
			if result.owner.Token != persisted.Token {
				t.Fatalf("round %d claimant observed token %q; persisted %q", round, result.owner.Token, persisted.Token)
			}
		}
		if err := New(root).Release(persisted); err != nil {
			t.Fatalf("round %d release: %v", round, err)
		}
	}
}

func TestStaleOwnerRecoveryContentionStress(t *testing.T) {
	const (
		rounds     = 10
		contenders = 16
	)
	for round := 0; round < rounds; round++ {
		root := t.TempDir()
		seedRuntime := New(root)
		seed, won, err := seedRuntime.Claim(os.Getpid())
		if err != nil || !won {
			t.Fatalf("round %d seed claim: won=%t err=%v", round, won, err)
		}
		seed.Heartbeat = time.Now().Add(-time.Hour)
		if err := atomicJSON(seedRuntime.Paths.Owner, seed); err != nil {
			t.Fatalf("round %d age owner: %v", round, err)
		}

		start := make(chan struct{})
		results := make(chan claimResult, contenders)
		for contender := 0; contender < contenders; contender++ {
			runtime := New(root)
			runtime.Timing.OwnerLease = time.Second
			go func(pid int) {
				<-start
				owner, claimed, claimErr := runtime.Claim(pid)
				results <- claimResult{owner: owner, won: claimed, err: claimErr}
			}(os.Getpid() + contender + 1)
		}
		close(start)

		winners := 0
		claims := make([]claimResult, 0, contenders)
		for contender := 0; contender < contenders; contender++ {
			result := <-results
			if result.err != nil {
				t.Fatalf("round %d stale recovery failed: %v", round, result.err)
			}
			if result.won {
				winners++
			}
			claims = append(claims, result)
		}
		if winners != 1 {
			t.Fatalf("round %d stale recovery winners=%d claims=%+v", round, winners, claims)
		}
		persisted, err := New(root).ReadOwner()
		if err != nil {
			t.Fatalf("round %d read recovered owner: %v", round, err)
		}
		if persisted.Token == seed.Token {
			t.Fatalf("round %d stale owner was not replaced", round)
		}
		for _, result := range claims {
			if result.owner.Token != persisted.Token {
				t.Fatalf("round %d claimant observed token %q; persisted %q", round, result.owner.Token, persisted.Token)
			}
		}
		if err := New(root).Release(persisted); err != nil {
			t.Fatalf("round %d release: %v", round, err)
		}
	}
}
