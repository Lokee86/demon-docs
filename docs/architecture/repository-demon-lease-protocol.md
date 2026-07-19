# Repository Demon Lease Protocol

Parent index: [Architecture](./README.md)

## Purpose

This document defines the implemented repository-demon ownership claim, feeder demand leases, heartbeat and release safety, stale-owner recovery, detached startup, shutdown conditions, and platform process seams.

## Overview

The repository demon is a single-owner lifecycle around the normal Demon Docs watcher.

One fresh owner record identifies the process responsible for the repository-local watcher. Independent feeder records represent shells or external agent sessions that currently need that watcher. The owner remains alive while configuration permits it and fresh feeders continue to exist, then exits after the no-feeder grace period or another shutdown condition.

Owner and feeder state lives under `.ddocs/runtime/`. It is operational and disposable. It is not part of `refs/ddocs/state`, review history, or authored repository truth.

The protocol provides single-owner exclusion and bounded recovery through filesystem primitives. It is not a distributed consensus system and does not guarantee that a recorded PID still identifies a live process. Freshness, ownership tokens, and periodic heartbeats are the authority.

## Primary ownership

The lease protocol owns:

- repository-local runtime paths;
- owner-lock acquisition and stale-lock cleanup;
- fresh-owner detection;
- stale or abandoned owner quarantine;
- atomic publication of a new owner record;
- opaque owner-token generation;
- token-checked owner heartbeat, PID update, and release;
- shell and agent feeder registration;
- feeder token validation and token-checked heartbeat;
- feeder expiry and cleanup;
- shell-feeder reuse by parent process;
- read-only feeder snapshots for status;
- shutdown-request files;
- no-feeder grace timing;
- configuration rechecks while serving;
- watcher cancellation when the owner exits;
- detached watcher and feeder process startup; and
- recovery of a missing or stale owner while demand remains active.

## Explicit non-ownership

The lease protocol does not own:

- index, link, or reverse-index reconciliation behavior;
- authored-file writes or rollback;
- durable link and reverse-index state;
- review decisions or applied-change history;
- repository configuration parsing beyond the demon enablement callback;
- host-specific agent scheduling;
- shell process creation outside generated integration hooks;
- operating-system process supervision after detached startup;
- proving PID identity independently of the owner token; or
- correctness when static `check` and `fix` are not run after an operational failure.

The watcher remains the mutation-capable worker. The demon only owns whether one watcher should be running and when it should stop.

## Runtime layout

The runtime owns these paths:

```text
.ddocs/runtime/
  owner.json
  owner-heartbeat
  owner.lock/
    token
  shutdown-request
  feeders/
    <token>.json
  logs/
    demon.log
    demon.log.1
    demon.log.2
    demon.log.3
    demon.log.4
```

Temporary claim and stale-owner paths may appear briefly:

```text
.owner-claim-<token>
owner.json.stale-<token>
```

Runtime initialization creates the feeder directory and its parents. Status inspection does not call runtime initialization and therefore does not create these paths.

## Timing model

The default timing values are:

```text
feeder heartbeat interval: 5 seconds
feeder expiry:             20 seconds
no-feeder shutdown grace:  20 seconds
owner lease:               20 seconds
```

The owner-lock implementation also uses:

```text
lock retry interval:       2 milliseconds
lock acquisition timeout:  15 seconds
lock stale threshold:      10 seconds
```

Tests may replace runtime timing values with shorter durations. Product behavior should be described in terms of the configured `Timing` values rather than relying on wall-clock constants in higher-level code.

## Owner record

`owner.json` contains:

```text
token       opaque 128-bit random token encoded as 32 lowercase hex characters
pid         detached demon process ID when known
started_at  UTC claim time
heartbeat   UTC owner heartbeat time
```

The claim initially records the claimant process ID. After detached startup succeeds, the caller updates the owner record with the detached process ID using the same owner token.

The PID is informational and useful for status. The token is the mutation authority for heartbeat, PID update, and release.

## Owner lock

Every owner claim, heartbeat, PID update, and release acquires the repository-local owner lock.

The lock is a directory:

```text
.ddocs/runtime/owner.lock/
```

Directory creation is the exclusion primitive. A successful claimant writes its own random token to `owner.lock/token`.

Release removes the lock only when the token file still contains the claimant's token. This prevents a delayed unlock function from deleting a lock that was replaced after stale recovery.

When directory creation fails, the caller treats it as contention and retries until the bounded deadline. This includes transient Windows access-denied behavior while another claimant removes the directory.

A lock directory older than the stale threshold is removed and acquisition is retried. The lock's modification time is the stale-lock signal; there is no active lock heartbeat.

If acquisition does not succeed before the timeout, the operation returns an error rather than proceeding without exclusion.

## Claim algorithm

`Runtime.Claim` performs the following sequence under the owner lock:

```text
create a new owner token
-> read current owner.json
-> return the current owner when it is fresh
-> quarantine a stale owner
-> handle unreadable owner state conservatively
-> build a new owner record
-> write a temporary claim file atomically
-> publish owner.json without replacement
-> mirror the heartbeat timestamp
-> return claimed=true
```

### Fresh owner

An owner is fresh when:

- `owner.json` decodes successfully;
- its token is non-empty; and
- its embedded heartbeat is not older than the owner lease.

When a fresh owner exists, `Claim` returns that owner with `claimed=false`. It does not replace it and does not validate the PID separately.

### Stale readable owner

A readable owner whose embedded heartbeat exceeds the owner lease is quarantined.

Quarantine renames `owner.json` to a token-qualified stale path and then removes the stale file. The rename prevents a new owner from being published while the old owner path still occupies the canonical name.

### Unreadable owner

If `owner.json` cannot be decoded or read for a reason other than absence, the protocol checks the file modification time.

- A recently modified unreadable owner is treated as potentially fresh and claim fails with an error.
- An unreadable owner older than the owner lease is quarantined and replacement may proceed.

This avoids silently replacing a partially written or temporarily inaccessible fresh owner while still permitting bounded recovery from abandoned invalid state.

## Atomic owner publication

A claimant writes the complete owner JSON to a same-directory temporary claim path using the atomic JSON helper.

It then attempts to create `owner.json` as a hard link to that claim file. Hard-link creation is atomic and fails rather than replacing an owner published by another claimant, including an older client that may not honor `owner.lock`.

When hard-link publication fails for a reason other than `already exists`, and a follow-up stat shows that `owner.json` is absent, the implementation falls back to exclusive file creation with `O_CREATE|O_EXCL`.

When publication fails because another fresh owner now exists, the current owner is returned with `claimed=false`.

Other publication errors are surfaced.

The temporary claim path is removed after the attempt. The published hard link remains as `owner.json`.

## Owner heartbeat

`Runtime.Heartbeat`:

1. acquires the process-local runtime mutex;
2. acquires the owner lock;
3. rereads `owner.json`;
4. compares its token with the caller's owner token;
5. updates the embedded UTC heartbeat;
6. atomically replaces `owner.json`; and
7. writes the same timestamp to `owner-heartbeat`.

A token mismatch is an ownership-loss error. The caller must stop serving rather than updating or releasing a newer owner's record.

Owner freshness is calculated from the heartbeat embedded in `owner.json`. The separate `owner-heartbeat` file mirrors activity for operational visibility and does not override an invalid or stale owner record.

## PID update

After detached startup, the spawning process calls `SetPID` with the owner token and child PID.

The method acquires the same process mutex and owner lock, rereads the current owner, rejects a token mismatch, and atomically replaces the owner JSON with the new PID.

PID publication is intentionally separate from the initial claim because the detached process does not exist until after ownership has been reserved.

A PID update failure does not automatically terminate the child. The caller currently ignores the returned error after successful spawn, so freshness and token ownership remain authoritative even when status retains the claimant PID.

## Owner release

`Runtime.Release` acquires the process mutex and owner lock, then rereads the owner record.

- Missing owner state is treated as already released.
- A token mismatch returns an error and leaves the current owner untouched.
- A matching owner removes `owner.json` and best-effort removes `owner-heartbeat`.

`Runtime.Serve` defers owner release. Normal watcher completion, context cancellation, disabled configuration, shutdown request, no-feeder grace expiry, and most serve errors therefore attempt token-safe cleanup.

If ownership was already replaced, deferred release fails token validation and does not remove the replacement owner.

## Feeder record

Each feeder file contains:

```text
token       opaque 32-character lowercase hexadecimal token
kind        shell or agent
client      optional external host name for agent feeders
pid         feeder helper or requesting process ID
parent_pid  shell or host parent process ID when available
heartbeat   UTC demand heartbeat time
```

The filename is the token plus `.json`.

Valid tokens are exactly 32 lowercase hexadecimal characters and cannot contain path separators. Reading a feeder also verifies that the JSON token matches the filename token.

## Feeder registration

Shell feeders use `AddFeeder("shell", ...)`.

Agent feeders use `AddAgentFeeder`, which additionally validates the trimmed client name:

- non-empty;
- no more than 128 bytes; and
- no NUL, carriage return, newline, or tab.

Registration creates the runtime directories when necessary, generates a token, builds the feeder record, and atomically writes its JSON file.

Feeder creation does not itself claim or start an owner. Application orchestration registers demand, calls `Claim`, and starts detached processes as necessary.

## Feeder heartbeat and release

`HeartbeatFeeder` reads one token-validated feeder, updates its UTC heartbeat, and atomically replaces only that feeder file.

`RemoveFeeder` validates the token and removes only the corresponding file. Missing files are treated as already released.

Feeder operations do not require the owner lock because each feeder token owns a distinct path. They also do not modify `owner.json` directly.

Public agent heartbeat orchestration may separately detect missing owner freshness, claim ownership, and start a replacement watcher while refreshing the feeder demand record.

## Feeder expiry

A feeder is fresh when its heartbeat age does not exceed the feeder-expiry duration.

`ListFeeders` is the mutating operational view. It:

- ignores directories and non-JSON files;
- ignores unreadable, malformed, or tokenless records;
- removes expired valid feeder files; and
- returns only fresh feeders.

`SnapshotFeeders` is the read-only status view. It applies the same validity and expiry filtering but does not remove expired files.

This distinction preserves the contract that `demon --status` does not mutate runtime state.

Malformed feeder files are skipped rather than deleted by either path.

## Shell-feeder reuse

Repeated shell entry from the same parent process should represent one demand lease.

`FindFeeder` scans a read-only feeder snapshot and returns a fresh record when both match:

```text
kind == shell
parent_pid == current parent PID
```

`demon run` and the hidden shell-entry command reuse that feeder instead of creating a new token and heartbeat helper.

Agent feeders are not reused by client name. External adapters own the token returned by each acquire call and must heartbeat and release it explicitly.

## Parent-process identity

The requesting parent PID comes from:

1. `DDOCS_PARENT_PID` when it contains a valid integer; otherwise
2. the operating system's parent PID for the command process.

Generated Bash integration sets `DDOCS_PARENT_PID` to the interactive shell PID so short-lived command processes can share one shell feeder.

The detached `__feed` helper checks parent liveness on every heartbeat cycle.

- Unix uses signal 0 through `syscall.Kill`.
- Windows invokes `tasklist` filtered to the PID.

A non-positive parent PID is treated as alive.

On Windows, failure to execute `tasklist` is treated conservatively as alive rather than expiring demand because process inspection failed.

When the parent is considered dead, the feeder helper exits and its deferred cleanup removes the feeder record.

## Detached startup

Detached processes restart the current executable with the `demon` alias and hidden command arguments.

For the owner:

```text
demon __serve <repository-root> <owner-token>
```

For a shell feeder heartbeat helper:

```text
demon __feed <repository-root> <feeder-token>
```

Standard input, output, and error are disconnected.

Platform behavior:

- Unix starts a new session with `Setsid`.
- Windows creates a new process group and uses the detached-process creation flag.

The starter returns the child PID after `Start`; it does not wait for process readiness.

Ownership is reserved before spawning `__serve`. If spawning fails, the caller releases the owner token and removes a newly created feeder where applicable.

If feeder-helper spawning fails, the newly created feeder is removed. Depending on the entry path, a newly claimed owner may also be released.

## Serve lifecycle

The hidden `__serve` entry validates that the supplied owner token matches the current owner record before opening logs or starting work.

It loads configuration, constructs the selected watcher features, and calls `Runtime.Serve`.

`Serve`:

1. ensures runtime directories exist;
2. defers owner release;
3. creates a child context for the watcher;
4. starts the watcher in one goroutine;
5. starts a ticker at the feeder-heartbeat interval; and
6. monitors watcher completion, outer cancellation, and each lease tick.

At each tick it:

```text
heartbeat the owner
-> list and clean expired feeders
-> update the last-seen-feeder time when demand exists
-> exit after shutdown grace with no feeders
-> exit when a shutdown request exists
-> reload demon enablement and exit when disabled
```

When `Serve` returns, deferred context cancellation stops the watcher and deferred release removes the matching owner.

## No-feeder grace

`lastFeeder` begins at serve startup time.

Whenever one or more fresh feeders exist, it is reset to the current time.

When no fresh feeders exist and the elapsed time since `lastFeeder` reaches the shutdown-grace duration, the owner exits.

The grace period absorbs brief shell transitions, heartbeat timing gaps, and feeder replacement without immediately tearing down the watcher.

The owner does not require that a feeder existed at least once after startup. A newly started owner with no valid feeder will exit after one grace period.

## Shutdown request

`RequestShutdown` ensures runtime directories exist and writes a timestamped `shutdown-request` file.

The owner checks only for file existence during each lease tick.

Disabling the demon:

```text
persist [demon].run = false
-> remove all currently readable fresh feeders
-> write shutdown-request
```

Re-enabling clears the shutdown request before allowing startup.

A live owner can remain active until its next tick observes the request. The shutdown file is not a force-kill signal.

## Watcher completion and failure

The watcher runs under the child context and reports one result to the serve loop.

- A nil result ends the owner cleanly.
- `context.Canceled` is treated as clean completion.
- Another watcher error ends the owner and is returned to the hidden serve command, which records it in repository-local logs.

The demon does not restart the watcher inside the same owner after an error.

When fresh feeder demand remains, a feeder helper or later public heartbeat can observe the missing or stale owner, claim a replacement, and start a new detached owner.

Static `ddocs check` and `ddocs fix` remain the correctness recovery surfaces.

## Ownership loss

Every serve tick calls token-checked owner heartbeat.

If `owner.json` is missing, unreadable, or contains another token, heartbeat fails and `Serve` exits. It does not continue running a watcher after losing the repository lease.

Deferred release then either finds the owner already absent or rejects the token mismatch without deleting the replacement owner.

This is the primary protection against an old detached process continuing indefinitely after stale recovery replaced its ownership record.

## Read-only status

`demon --status` discovers repository location without bootstrapping a linked worktree, loads configuration, and reads current runtime state.

It does not:

- create `.ddocs/runtime/`;
- acquire the owner lock;
- clean expired feeder files;
- clear shutdown requests;
- claim ownership;
- start a process; or
- initialize a linked worktree's local state.

It reports `running` only when `OwnerFresh` succeeds. A readable but expired owner is `stale`. Missing or unreadable owner state appears as `stopped` unless a readable record was obtained.

Feeder counts come from `SnapshotFeeders`, so expired records are excluded from counts without being deleted.

## Linked-worktree isolation

Mutating demon entry can bootstrap a linked Git worktree that has not yet initialized local Demon Docs state.

The worktree receives independent:

- configuration copy;
- `.ddocs` object storage;
- runtime owner and feeder records;
- logs; and
- watcher process.

Read-only status can detect the linked worktree location without creating that state.

Owner locks and leases are therefore scoped to the local worktree's `.ddocs/runtime/`, not shared through the primary worktree's Git common directory.

## Contention and recovery scenarios

### Two callers claim simultaneously

The owner lock serializes compliant callers. Atomic hard-link or exclusive-file publication protects against another publisher that does not honor the lock. One caller publishes; the other returns the fresh owner with `claimed=false` or receives a bounded error.

### Claimant dies while holding `owner.lock`

After the lock directory's modification time exceeds the stale threshold, another caller removes it and retries.

The token-checked unlock prevents an older deferred unlock from deleting a replacement lock.

### Claimant dies after publishing owner but before detached startup

The owner heartbeat remains at claim time and eventually exceeds the owner lease. A later caller quarantines the stale owner and publishes a replacement.

### Owner file is temporarily unreadable

A recently modified unreadable file blocks replacement. An old unreadable file is quarantined after the owner-lease threshold.

### Owner process dies

The embedded heartbeat stops advancing. Fresh feeder demand can replace it after the owner lease expires.

### Feeder helper dies without release

Its feeder expires after the feeder-expiry interval. `ListFeeders` removes the stale record during owner service.

### Host adapter misses release

The same expiry bound applies. The adapter should still release on success, failure, cancellation, timeout, and spawn failure to avoid unnecessary grace-period extension.

### Owner loses its token

The next heartbeat fails and the old serve loop exits. It cannot overwrite or release the replacement owner.

## Failure boundaries

The protocol surfaces errors for:

- runtime-directory creation;
- token generation;
- owner-lock timeout;
- fresh unreadable owner state;
- stale-owner quarantine failure;
- owner publication failure;
- token mismatch;
- feeder read or atomic write failure;
- configuration reload failure;
- owner heartbeat failure;
- feeder listing failure;
- watcher failure; and
- detached process startup failure.

Some cleanup operations are best effort:

- removing the mirrored heartbeat file;
- removing temporary claim paths;
- removing quarantined stale owner paths after rename;
- clearing shutdown requests; and
- setting the detached PID after successful spawn.

Best-effort cleanup does not authorize replacing a fresh owner or bypassing token checks.

## Invariants

The lease design must preserve these invariants:

- At most one fresh owner record is authoritative per local `.ddocs/runtime/`.
- A compliant owner mutation occurs under `owner.lock`.
- An unlock removes only the lock instance with the same token.
- A fresh owner is never intentionally replaced by `Claim`.
- An unreadable fresh owner blocks replacement.
- Heartbeat, PID update, and release require the current owner token.
- An old owner cannot release a replacement owner.
- Feeder tokens address only their own feeder paths.
- Status inspection remains read-only.
- Expired feeders do not count as active demand.
- Shell reuse is scoped to fresh shell feeders with the same parent PID.
- Watcher service ends when ownership heartbeat fails.
- No-feeder shutdown is delayed by the configured grace period.
- Detached lifecycle remains independent for each linked worktree.
- Runtime state never becomes authored repository truth.

## Extension rules

### Adding an owner field

Preserve backward-compatible JSON decoding or define stale/unreadable recovery behavior. Fields required for authority must not default silently from an older record.

### Changing freshness rules

Owner freshness currently uses the embedded owner heartbeat. Changing to another source requires updating claim, status, heartbeat, stale recovery, and tests together.

### Adding a feeder kind

Update kind validation, status counting, reuse rules, public help, host-adapter guidance, and expiry tests. Do not make one feeder capable of deleting another feeder's record.

### Adding a host adapter

Use public acquire, heartbeat, and release operations. Retain the opaque token for one repository session, heartbeat less frequently than feeder expiry, and release on every terminal path.

Do not read or write runtime JSON directly from the adapter.

### Changing detached startup

Preserve owner-before-spawn reservation, token validation in `__serve`, disconnected standard streams, failure cleanup, and platform process-group isolation.

### Adding automatic restart

Define backoff, repeated failure visibility, ownership-token transfer, configuration disablement, and watcher-state correctness before restarting inside the same process. Current recovery is driven by later feeder demand.

## Verification

Focused verification:

```bash
go test ./internal/demon ./internal/app -count=1
```

Important contracts include:

- single fresh owner under contention;
- stale owner replacement;
- stale owner-lock recovery;
- token-safe heartbeat and release;
- feeder registration, heartbeat, expiry, and counting;
- shell-feeder reuse;
- read-only status snapshots;
- no-feeder shutdown grace;
- watcher failure propagation;
- detached startup failure cleanup;
- parent-process liveness behavior;
- linked-worktree bootstrap and isolation; and
- bounded log operation.

## Code map

Primary implementation:

- `internal/demon/runtime.go` — runtime paths, owner lock, claims, heartbeats, PID update, release, feeder enumeration, shutdown requests, and serve loop.
- `internal/demon/activity.go` — agent feeder validation, feeder creation, token validation, reading, and heartbeat.
- `internal/demon/log.go` — bounded repository-local log rotation.
- `internal/app/demon.go` — public and hidden lifecycle commands, shell feeder reuse, detached startup orchestration, status, serve, feed, and generated shell hooks.
- `internal/app/demon_activity.go` — public agent acquire, heartbeat, and release orchestration.
- `internal/app/detach_unix.go` — Unix session detachment.
- `internal/app/detach_windows.go` — Windows detached process and process-group flags.
- `internal/app/parent_unix.go` — Unix parent-process probe.
- `internal/app/parent_windows.go` — Windows parent-process probe.
- `internal/repository/worktree.go` — linked-worktree detection and mutating bootstrap.
- `internal/watch/` — watcher lifecycle owned by the active demon.

Focused tests:

- `internal/demon/runtime_test.go`
- `internal/demon/activity_test.go`
- `internal/demon/ownership_stress_test.go`
- `internal/app/demon_test.go`
- `internal/app/demon_activity_test.go`

## Related docs

- [Repository Demon](../operations/repository-demon.md)
- [Host Adapter Feeder Integration](../operations/host-adapters.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Watch Scheduler and Reconciliation Serialization](watch-scheduler.md)
- [Dynamic Watch Scope](../operations/dynamic-watch-scope.md)
- [Repository Scope and Worktrees](repository-scope-and-worktrees.md)
- [Private Object Repository](private-object-repository.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

The separate `owner-heartbeat` file mirrors successful owner activity, but current freshness decisions use the heartbeat embedded in `owner.json`. Operational tooling should not treat the mirror as an independent lease.
