# Repository Demon

The repository demon is the self-managing background lifecycle around the existing Demon Docs watcher. It runs the same deterministic reconciliation operations as `ddocs watch`; it does not introduce a second indexing, link-repair, or repository-truth system.

The static commands remain authoritative:

- `ddocs check` verifies repository state without requiring the demon;
- `ddocs fix` rebuilds or repairs state without requiring the demon; and
- `ddocs watch` runs the watcher explicitly in the foreground.

The repository demon exists only to keep that watcher available while shells or agents are actively working in a repository.

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

The generic agent feeder boundary is deliberately host-neutral. Demon Docs does not need to know whether an agent feeder came from Codex, Hermes, an MCP server, Claude Code, or another plugin. Each adapter is responsible for registering before work begins and unregistering on every terminal path, including success, failure, cancellation, timeout, and spawn failure.

Agent registration is operational only. It keeps the watcher alive while an adapter is active; it does not make the demon an MCP server, context service, or host integration.

## Public Commands

```bash
ddocs demon run
ddocs demon run --false
ddocs demon run --true
ddocs demon --status
ddocs demon --logs
```

`ddocs demon run` ensures the demon is enabled, registers the current shell as a feeder, and starts the detached watcher when necessary.

`ddocs demon run --false` persists `[demon].run = false`, removes current feeders, and requests shutdown.

`ddocs demon run --true` persists `[demon].run = true`, clears an earlier shutdown request, and allows a feeder to start the demon again.

`ddocs demon --status` is read-only. It reports:

- repository root;
- configured enablement;
- running, stale, or stopped ownership state;
- detached demon PID;
- active shell and agent counts;
- last owner heartbeat; and
- watched docs root.

`ddocs demon --logs` prints retained repository-local logs from oldest to newest.

The hidden `__enter`, `__leave`, `__feed`, and `__serve` commands are adapter and lifecycle plumbing rather than normal interactive commands.

## Shell Integration

Bash startup files can install the repository transition hook with:

```bash
eval "$(ddocs demon __shell-hook bash)"
```

PowerShell profiles can install it with:

```powershell
Invoke-Expression (& ddocs demon __shell-hook powershell)
```

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
  shutdown-request
  feeders/
  logs/
    demon.log
    demon.log.1
    demon.log.2
    demon.log.3
    demon.log.4
```

The owner record stores the ownership token, detached PID, startup time, and last heartbeat. Feeder files store their token, kind, process information, and last heartbeat.

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

## Configuration

Initialized repositories default to:

```toml
[demon]
run = true
```

This setting permits self-managed operation; it does not make the demon a correctness dependency. Disabling it leaves `check`, `fix`, and foreground `watch` available.

See [Configuration](configuration.md) for the complete configuration model and [Watcher and Automation](watcher-and-automation.md) for foreground watcher behavior.
