---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-786b-b8fa-227d06a1db44
document_type: general
policy_exempt: false
summary: This document describes the implemented single-owner repository demon, feeder lease protocol, shell integration, linked-worktree behavior, runtime state, shutdown, recovery, and logs.
---
# Repository Demon

Parent index: [Operations](./INDEX.md)

## Purpose

This document describes the implemented single-owner repository demon, feeder lease protocol, shell integration, linked-worktree behavior, runtime state, shutdown, recovery, and logs.

## Overview

The repository demon is an optional lifecycle owner around the normal watcher. It keeps one fresh repository-local watcher active while shell or agent feeders remain registered. Unlike foreground `ddocs watch`, it requires an initialized repository with `.ddocs/config.toml`. It does not replace static `check` or `fix` correctness.

## Operating model

```text
host enters repository work
-> acquire feeder token
-> demon owner starts or is reused
-> host refreshes heartbeat
-> watcher reconciles filesystem changes
-> host releases token
-> demon exits after grace when no feeders remain
```

The repository demon is the self-managing background lifecycle around the existing Demon Docs watcher. It runs the same deterministic reconciliation operations as `ddocs watch`; it does not introduce a second indexing, link-repair, or repository-truth system.

The static commands remain authoritative:

- `ddocs check` verifies standalone or initialized scope state without requiring the demon;
- `ddocs fix` rebuilds or repairs standalone or initialized scope state without requiring the demon; and
- `ddocs watch` runs the watcher explicitly in the foreground without requiring repository initialization.

The repository demon exists only to keep that watcher available while shells or agents are actively working in a repository.

It never invokes production codemap generation. Codemap fix, check, inspect, managed-section adoption, recommendation planning, and pruning remain explicit foreground commands. The demon may observe a Markdown write produced by one of those commands and refresh normal watcher-owned state, but it does not regenerate the codemap or schedule another codemap pass.

## One Demon per Repository

An initialized repository owns its demon state below its local `.ddocs/` directory. One fresh owner lease may exist for that repository at a time, regardless of how many shells or agents are using it.

A feeder entering the repository follows this lifecycle:

1. register a repository-local feeder and receive an opaque token;
2. claim ownership only when no fresh owner exists;
3. start the detached watcher when the ownership claim succeeds;
4. refresh the feeder heartbeat while the shell or agent remains active;
5. remove only that feeder when its shell, job, or session ends; and
6. allow the demon to stop after the grace period when no active feeders remain.

Ownership publication and state replacement are atomic. A stale owner lease can be recovered, but a second live caller must not replace a fresh owner. Status reports the detached demon process ID rather than the process that requested startup.

## Feeders

Feeders describe active demand for the repository demon. They do not own repository truth and they do not perform reconciliation themselves.

Two feeder kinds are supported:

- `shell`: an interactive Bash or PowerShell session currently inside the repository;
- `agent`: an MCP job, native agent integration, or other automated session working in the repository.

Each feeder has an opaque token and its own heartbeat record. Leaving one shell or finishing one agent job removes only that feeder. It does not shut down a demon still needed by another feeder.

The generic agent feeder boundary is deliberately host-neutral. Demon Docs does not need to know whether an agent feeder came from Codex, Hermes, an MCP server, Claude Code, or another plugin. Each adapter supplies a client name, registers before work begins, refreshes its opaque token while work continues, and releases the token on every terminal path, including success, failure, cancellation, timeout, and spawn failure.

The generic feeder protocol exists in Demon Docs core. Thin MCP, Codex, Hermes, and other host adapters remain integration work; the daemon does not invoke those hosts itself.

Agent registration is operational only. It keeps the watcher alive while an adapter is active; it does not make the demon an MCP server, context service, or host integration. A missed release is bounded by feeder expiry, and a later heartbeat can recover a missing or stale demon owner.

## Public Commands

```bash
demon run
demon acquire --client mcp
demon heartbeat --token TOKEN
demon release --token TOKEN
demon --status
demon --logs
```

The same operations are available through the general CLI by prefixing them with `ddocs`:

```bash
ddocs demon run
ddocs demon acquire --client codex
ddocs demon heartbeat --token TOKEN
ddocs demon release --token TOKEN
```

`ddocs demon run` ensures the demon is enabled, registers the current shell as a feeder, and starts the detached watcher when necessary. It does not return until the initial reconciliation has completed and every configured filesystem watcher is registered, so an immediately following filesystem change cannot fall into a startup gap. Readiness remains bounded by a two-minute deadline, which accommodates cold initial reconciliation on large repositories while still reporting a startup that cannot converge.

`ddocs demon run --false` persists `[demon].run = false`, removes current feeders, and requests shutdown.

`ddocs demon run --true` persists `[demon].run = true`, clears an earlier shutdown request, and allows a feeder to start the demon again.

`demon acquire --client NAME` registers an externally managed `agent` feeder and prints `token=... claimed=...`. MCP, Codex, Hermes, and other adapters retain that token for the lifetime of their repository session.

`demon heartbeat --token TOKEN` refreshes that feeder. If the repository still has active demand but the demon owner is missing or stale, the heartbeat claims ownership and starts a replacement watcher.

`demon release --token TOKEN` removes only that feeder. Other shells and agents remain unaffected, and the demon stops after the configured grace period once no fresh feeders remain.

`ddocs demon --status` is read-only. It reports:

- repository root;
- configured enablement;
- starting, running, stale, or stopped ownership state;
- detached demon PID;
- active shell and agent counts;
- last owner heartbeat; and
- watched docs root.

`ddocs demon --logs` prints retained repository-local logs from oldest to newest.

The hidden `__enter`, `__leave`, `__feed`, and `__serve` commands remain internal shell and detached-process plumbing. External agent adapters use the public `acquire`, `heartbeat`, and `release` commands instead.

## Shell Integration

Bash startup files can install the repository transition hook with:

```bash
eval "$(ddocs demon __shell-hook bash)"
```

PowerShell profiles can install it with:

```powershell
Invoke-Expression (& ddocs demon __shell-hook powershell)
```

The PowerShell command emits one physical bootstrap line, even though the decoded hook is multiline. This keeps Windows PowerShell 5.1 from converting native-command output into an `Object[]` that `Invoke-Expression` cannot execute. Repository and active-shell values are parsed by removing their named prefixes rather than fixed character offsets, preserving Windows drive letters and single-digit counts.

The hook tracks its repository root and feeder token. Entering a Demon Docs repository registers one shell feeder. Moving to another repository or leaving the repository removes the old feeder rather than issuing a repository-wide shutdown request.

The hook announces when it actually claims and starts a demon, then reports the current active-shell count. The ownership result comes from the enter operation itself rather than a separate status guess.

## Linked Worktrees

A linked Git worktree receives independent Demon Docs runtime and object state under that worktree's own `.ddocs/` directory.

Read-only discovery can identify a linked worktree from nested directories without creating runtime state. The first mutating demon entry bootstraps the worktree by copying the primary worktree's Demon Docs configuration and initializing fresh local `.ddocs/` object storage. The primary and linked worktrees therefore do not share a running demon or mutable Demon Docs state.

Git awareness is limited to this worktree adapter. Ordinary Demon Docs repository discovery remains based on `.ddocs/config.toml`.

## Runtime State

Runtime files live below `.ddocs/runtime/`:

```text
.ddocs/runtime/
  owner.json
  owner-heartbeat
  ready.json
  shutdown-request
  feeders/
  logs/
    demon.log
    demon.log.1
    demon.log.2
    demon.log.3
    demon.log.4
```

The owner record stores the ownership token, detached PID, startup time, and last heartbeat. `ready.json` is token-scoped startup state written only after the initial reconciliation completes and all configured filesystem watches are registered. Feeder files store their token, kind, optional external client name, process information, and last heartbeat.

Runtime state is operational and disposable. It is excluded from document traversal and is separate from the schema-versioned `.ddocs/` object repository used for link and repository state.

Logs are bounded to five files, with each file limited to approximately 1 MiB. Rotation preserves recent operational history without allowing unbounded repository-local log growth.

## Shutdown and Recovery

The demon stops when any of these conditions apply:

- `[demon].run` becomes false;
- an explicit shutdown request is present;
- no fresh feeders remain for the configured grace period;
- the watcher exits; or
- the owning process loses its valid ownership token.

Expired feeder records do not count as active. Normal status inspection does not create runtime directories, delete stale feeder files, bootstrap a linked worktree, or otherwise mutate the repository.

A live feeder can recover a stale or missing owner by claiming the lease and starting a replacement watcher. Re-enabling the demon clears an earlier shutdown request so the new owner does not immediately exit.

During watcher startup, transient source movement or content-hash races discard the stale plan and retry after a bounded quiet delay instead of terminating the owner. Once observation is active, an operating-system event-buffer overflow is treated as lost event detail: the watcher schedules a complete reconciliation and continues using the recursive root observer rather than stopping the demon.

## Configuration

Initialized repositories default to:

```toml
[demon]
run = true
```

This setting permits self-managed operation; it does not make the demon a correctness dependency. Disabling it leaves `check`, `fix`, and foreground `watch` available.

See [Configuration](../reference/configuration.md) for the complete configuration model and [Watcher and Automation](./watcher-and-automation.md) for foreground watcher behavior.

## Code map

- `internal/demon/runtime.go` — owner claims, heartbeats, feeders, shutdown requests, stale recovery, and runtime paths.
- `internal/demon/log.go` — bounded repository-local log rotation.
- `internal/app/demon.go` — public and hidden daemon commands plus Bash and PowerShell hook generation.
- `internal/repository/worktree.go` — linked-worktree discovery and first-mutating-entry bootstrap.
- `internal/watch/` — the foreground reconciliation watcher reused by the daemon owner.
- `internal/app/demon_test.go` — command, feeder, worktree, lifecycle, and status coverage.
- `internal/demon/runtime_test.go` — owner and feeder state coverage, including single-owner behavior.

## Verification

Use:

```bash
demon --status
demon --logs
ddocs check
```

Focused tests cover ownership, feeder leases, runtime state, stale recovery, worktrees, CLI integration, and logs:

```bash
go test ./internal/demon ./internal/app -count=1
```

## Related docs

- [Operations](INDEX.md)
- [CI and Automation](../guides/ci-and-automation.md)
- [CLI Reference](../reference/cli.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Repository Demon Lease Protocol](../architecture/repository-demon-lease-protocol.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Watcher and Automation](watcher-and-automation.md)
- [Dynamic Watch Scope](dynamic-watch-scope.md)
- [Recovery and Troubleshooting](recovery-and-troubleshooting.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)

## Notes

Host adapters must release feeder tokens on success, failure, cancellation, timeout, and spawn failure. The demon does not deliver agent context.
