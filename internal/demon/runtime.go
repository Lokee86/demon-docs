package demon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	OwnerFile          = "owner.json"
	OwnerHeartbeat     = "owner-heartbeat"
	ReadyFile          = "ready.json"
	ShutdownRequest    = "shutdown-request"
	FeedersDir         = "feeders"
	LogsDir            = "logs"
	DemonLog           = "demon.log"
	ownerLockDir       = "owner.lock"
	ownerLockTokenFile = "token"

	ownerLockRetry      = 2 * time.Millisecond
	ownerLockTimeout    = 15 * time.Second
	ownerLockStaleAfter = 10 * time.Second
)

type Timing struct {
	FeederHeartbeat time.Duration
	FeederExpiry    time.Duration
	ShutdownGrace   time.Duration
	OwnerLease      time.Duration
}

func DefaultTiming() Timing {
	return Timing{5 * time.Second, 20 * time.Second, 20 * time.Second, 20 * time.Second}
}

type Paths struct {
	Root, Config, Runtime, Owner, Heartbeat, Ready, Shutdown, Feeders, Logs, Log string
}

func NewPaths(root string) Paths {
	runtimeRoot := filepath.Join(root, ".ddocs", "runtime")
	return Paths{Root: root, Config: filepath.Join(root, ".ddocs", "config.toml"), Runtime: runtimeRoot,
		Owner: filepath.Join(runtimeRoot, OwnerFile), Heartbeat: filepath.Join(runtimeRoot, OwnerHeartbeat),
		Ready: filepath.Join(runtimeRoot, ReadyFile), Shutdown: filepath.Join(runtimeRoot, ShutdownRequest),
		Feeders: filepath.Join(runtimeRoot, FeedersDir), Logs: filepath.Join(runtimeRoot, LogsDir),
		Log: filepath.Join(runtimeRoot, LogsDir, DemonLog)}
}

type Owner struct {
	Token     string    `json:"token"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	Heartbeat time.Time `json:"heartbeat"`
}

type readyState struct {
	Token string `json:"token"`
}

type Feeder struct {
	Token     string    `json:"token"`
	Kind      string    `json:"kind"`
	Client    string    `json:"client,omitempty"`
	PID       int       `json:"pid,omitempty"`
	ParentPID int       `json:"parent_pid,omitempty"`
	Heartbeat time.Time `json:"heartbeat"`
}

type Runtime struct {
	Paths  Paths
	Timing Timing
	mu     sync.Mutex
}

func New(root string) *Runtime {
	root, _ = filepath.Abs(root)
	return &Runtime{Paths: NewPaths(root), Timing: DefaultTiming()}
}

func (r *Runtime) Ensure() error {
	return os.MkdirAll(r.Paths.Feeders, 0o755)
}

func (r *Runtime) ensureRuntime() error {
	return os.MkdirAll(r.Paths.Runtime, 0o755)
}

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func atomicJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func exclusiveJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	remove := true
	defer func() {
		_ = f.Close()
		if remove {
			_ = os.Remove(path)
		}
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	remove = false
	return nil
}

func (r *Runtime) Claim(pid int) (Owner, bool, error) {
	unlock, err := r.acquireOwnerLock()
	if err != nil {
		return Owner{}, false, err
	}
	defer unlock()

	token, err := newToken()
	if err != nil {
		return Owner{}, false, err
	}
	current, readErr := r.ReadOwner()
	if readErr == nil {
		if time.Since(current.Heartbeat) <= r.Timing.OwnerLease {
			return current, false, nil
		}
		if err := r.quarantineOwner(token); err != nil {
			return Owner{}, false, err
		}
	} else if !os.IsNotExist(readErr) {
		info, statErr := os.Stat(r.Paths.Owner)
		if statErr != nil {
			if !os.IsNotExist(statErr) {
				return Owner{}, false, statErr
			}
		} else {
			if time.Since(info.ModTime()) <= r.Timing.OwnerLease {
				return Owner{}, false, fmt.Errorf("unable to read fresh demon owner: %w", readErr)
			}
			if err := r.quarantineOwner(token); err != nil {
				return Owner{}, false, err
			}
		}
	}

	now := time.Now().UTC()
	owner := Owner{Token: token, PID: pid, StartedAt: now, Heartbeat: now}
	claimPath := filepath.Join(r.Paths.Runtime, ".owner-claim-"+token)
	if err := atomicJSON(claimPath, owner); err != nil {
		return Owner{}, false, err
	}
	defer os.Remove(claimPath)

	// Hard-link publication is atomic and fails instead of replacing an
	// owner published by an older client that does not honor owner.lock.
	err = os.Link(claimPath, r.Paths.Owner)
	if err != nil && !errors.Is(err, os.ErrExist) {
		if _, statErr := os.Stat(r.Paths.Owner); os.IsNotExist(statErr) {
			err = exclusiveJSON(r.Paths.Owner, owner)
		}
	}
	if err != nil {
		if current, readErr := r.ReadOwner(); readErr == nil && time.Since(current.Heartbeat) <= r.Timing.OwnerLease {
			return current, false, nil
		}
		return Owner{}, false, err
	}
	_ = os.Remove(r.Paths.Ready)
	_ = os.WriteFile(r.Paths.Heartbeat, []byte(now.Format(time.RFC3339Nano)+"\n"), 0o644)
	return owner, true, nil
}

func (r *Runtime) quarantineOwner(token string) error {
	stale := r.Paths.Owner + ".stale-" + token
	if err := os.Rename(r.Paths.Owner, stale); err != nil {
		return err
	}
	_ = os.Remove(stale)
	return nil
}

func (r *Runtime) acquireOwnerLock() (func(), error) {
	if err := r.ensureRuntime(); err != nil {
		return nil, err
	}
	lockToken, err := newToken()
	if err != nil {
		return nil, err
	}
	lockPath := filepath.Join(r.Paths.Runtime, ownerLockDir)
	tokenPath := filepath.Join(lockPath, ownerLockTokenFile)
	deadline := time.Now().Add(ownerLockTimeout)
	var lastErr error
	for {
		err := os.Mkdir(lockPath, 0o700)
		if err == nil {
			if err := os.WriteFile(tokenPath, []byte(lockToken+"\n"), 0o600); err != nil {
				_ = os.RemoveAll(lockPath)
				return nil, err
			}
			return func() {
				data, readErr := os.ReadFile(tokenPath)
				if readErr == nil && string(data) == lockToken+"\n" {
					_ = os.Remove(tokenPath)
					_ = os.Remove(lockPath)
				}
			}, nil
		}
		lastErr = err
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > ownerLockStaleAfter {
			_ = os.RemoveAll(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for demon owner lock: %w", lastErr)
		}
		// Windows may briefly report access denied while another claimant
		// removes the directory. Treat every post-Ensure mkdir failure as
		// contention until the bounded deadline expires.
		time.Sleep(ownerLockRetry)
	}
}

func (r *Runtime) ReadOwner() (Owner, error) {
	data, err := os.ReadFile(r.Paths.Owner)
	if err != nil {
		return Owner{}, err
	}
	var owner Owner
	if err := json.Unmarshal(data, &owner); err != nil {
		return Owner{}, err
	}
	return owner, nil
}

func (r *Runtime) OwnerFresh() (Owner, bool) {
	owner, err := r.ReadOwner()
	return owner, err == nil && owner.Token != "" && time.Since(owner.Heartbeat) <= r.Timing.OwnerLease
}

func (r *Runtime) MarkReady(owner Owner) error {
	current, err := r.ReadOwner()
	if err != nil {
		return err
	}
	if current.Token != owner.Token {
		return fmt.Errorf("demon ownership token mismatch")
	}
	return atomicJSON(r.Paths.Ready, readyState{Token: owner.Token})
}

func (r *Runtime) Ready(owner Owner) bool {
	data, err := os.ReadFile(r.Paths.Ready)
	if err != nil {
		return false
	}
	var state readyState
	return json.Unmarshal(data, &state) == nil && state.Token != "" && state.Token == owner.Token
}

func (r *Runtime) WaitReady(ctx context.Context, owner Owner, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if r.Ready(owner) {
			return nil
		}
		current, err := r.ReadOwner()
		if os.IsNotExist(err) && !r.ownerHeartbeatFresh() {
			return fmt.Errorf("demon stopped before its watcher became ready")
		}
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return fmt.Errorf("timed out waiting for demon watcher readiness after %s: last owner read: %w", timeout, err)
			case <-ticker.C:
				continue
			}
		}
		if current.Token != owner.Token {
			return fmt.Errorf("demon ownership changed before its watcher became ready")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return fmt.Errorf("timed out waiting for demon watcher readiness after %s", timeout)
		case <-ticker.C:
		}
	}
}

func (r *Runtime) ownerHeartbeatFresh() bool {
	info, err := os.Stat(r.Paths.Heartbeat)
	return err == nil && time.Since(info.ModTime()) <= r.Timing.OwnerLease
}

func (r *Runtime) Heartbeat(owner Owner) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.acquireOwnerLock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := r.ReadOwner()
	if err != nil {
		return err
	}
	if current.Token != owner.Token {
		return fmt.Errorf("demon ownership token mismatch")
	}
	now := time.Now().UTC()
	current.Heartbeat = now
	if err := atomicJSON(r.Paths.Owner, current); err != nil {
		return err
	}
	return os.WriteFile(r.Paths.Heartbeat, []byte(now.Format(time.RFC3339Nano)+"\n"), 0o644)
}

func (r *Runtime) SetPID(token string, pid int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.acquireOwnerLock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := r.ReadOwner()
	if err != nil {
		return err
	}
	if current.Token != token {
		return fmt.Errorf("demon ownership token mismatch")
	}
	current.PID = pid
	return atomicJSON(r.Paths.Owner, current)
}

func (r *Runtime) Release(owner Owner) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := r.acquireOwnerLock()
	if err != nil {
		return err
	}
	defer unlock()
	current, err := r.ReadOwner()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if current.Token != owner.Token {
		return fmt.Errorf("demon ownership token mismatch")
	}
	if err := os.Remove(r.Paths.Owner); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = os.Remove(r.Paths.Heartbeat)
	_ = os.Remove(r.Paths.Ready)
	return nil
}

func (r *Runtime) AddFeeder(kind string, pid, parentPID int) (Feeder, error) {
	return r.addFeeder(kind, "", pid, parentPID)
}

func (r *Runtime) FindFeeder(kind string, parentPID int) (Feeder, bool) {
	feeders, err := r.SnapshotFeeders()
	if err != nil {
		return Feeder{}, false
	}
	now := time.Now()
	for _, feeder := range feeders {
		if feeder.Kind == kind && feeder.ParentPID == parentPID && now.Sub(feeder.Heartbeat) <= r.Timing.FeederExpiry {
			return feeder, true
		}
	}
	return Feeder{}, false
}

func (r *Runtime) feederPath(token string) string {
	return filepath.Join(r.Paths.Feeders, token+".json")
}

func (r *Runtime) FeedHeartbeat(feeder Feeder) error {
	_, err := r.HeartbeatFeeder(feeder.Token)
	return err
}

func (r *Runtime) RemoveFeeder(token string) error {
	if !validFeederToken(token) {
		return fmt.Errorf("invalid feeder token")
	}
	err := os.Remove(r.feederPath(token))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (r *Runtime) RemoveAllFeeders() error {
	feeders, err := r.SnapshotFeeders()
	if err != nil {
		return err
	}
	for _, feeder := range feeders {
		if err := r.RemoveFeeder(feeder.Token); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runtime) ListFeeders() ([]Feeder, error) {
	return r.listFeeders(true)
}

// SnapshotFeeders reads feeder state without cleaning expired records. It is
// used by status inspection, which must remain read-only.
func (r *Runtime) SnapshotFeeders() ([]Feeder, error) {
	return r.listFeeders(false)
}

func (r *Runtime) listFeeders(cleanup bool) ([]Feeder, error) {
	entries, err := os.ReadDir(r.Paths.Feeders)
	if os.IsNotExist(err) {
		return []Feeder{}, nil
	}
	if err != nil {
		return nil, err
	}
	feeders := make([]Feeder, 0, len(entries))
	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(r.Paths.Feeders, entry.Name())
		data, readErr := os.ReadFile(path)
		var feeder Feeder
		if readErr != nil || json.Unmarshal(data, &feeder) != nil || feeder.Token == "" {
			continue
		}
		if cleanup && now.Sub(feeder.Heartbeat) > r.Timing.FeederExpiry {
			_ = os.Remove(path)
			continue
		}
		if !cleanup && now.Sub(feeder.Heartbeat) > r.Timing.FeederExpiry {
			continue
		}
		feeders = append(feeders, feeder)
	}
	return feeders, nil
}

func (r *Runtime) RequestShutdown() error {
	if err := r.Ensure(); err != nil {
		return err
	}
	return os.WriteFile(r.Paths.Shutdown, []byte(time.Now().UTC().Format(time.RFC3339Nano)+"\n"), 0o644)
}

func (r *Runtime) ShutdownRequested() bool { _, err := os.Stat(r.Paths.Shutdown); return err == nil }

func (r *Runtime) ClearShutdown() { _ = os.Remove(r.Paths.Shutdown) }

func CountKinds(feeders []Feeder) (shells, agents int) {
	for _, feeder := range feeders {
		if feeder.Kind == "agent" {
			agents++
		} else if feeder.Kind == "shell" {
			shells++
		}
	}
	return shells, agents
}

type WatchFunc func(context.Context, io.Writer) error

func (r *Runtime) Serve(ctx context.Context, owner Owner, enabled func() (bool, error), watch WatchFunc, log io.Writer) error {
	if log == nil {
		log = io.Discard
	}
	if err := r.Ensure(); err != nil {
		return err
	}
	defer r.Release(owner)
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	watchErr := make(chan error, 1)
	go func() { watchErr <- watch(child, log) }()
	ticker := time.NewTicker(r.Timing.FeederHeartbeat)
	defer ticker.Stop()
	lastFeeder := time.Now()
	for {
		select {
		case err := <-watchErr:
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.Heartbeat(owner); err != nil {
				return err
			}
			feeders, err := r.ListFeeders()
			if err != nil {
				return err
			}
			if len(feeders) > 0 {
				lastFeeder = time.Now()
			} else if time.Since(lastFeeder) >= r.Timing.ShutdownGrace {
				return nil
			}
			if r.ShutdownRequested() {
				r.ClearShutdown()
				return nil
			}
			on, err := enabled()
			if err != nil {
				return err
			}
			if !on {
				return nil
			}
		}
	}
}
