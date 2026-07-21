---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-702e-b2f8-81010f9c18f6
document_type: general
policy_exempt: false
summary: This document defines the current host-neutral feeder workflow for MCP servers, coding agents, editors, and other external processes that keep the repository demon active while work is in progress.
---
# Host Adapter Feeder Integration

Parent index: [Operations](./INDEX.md)

## Purpose

This document defines the current host-neutral feeder workflow for MCP servers, coding agents, editors, and other external processes that keep the repository demon active while work is in progress.

## Overview

External hosts integrate through three public CLI commands:

```text
acquire -> heartbeat -> release
```

The host owns process lifecycle and token handling. Demon Docs owns repository discovery, feeder files, single-owner coordination, watcher startup or recovery, expiry, shutdown grace, and repository-local logs.

This interface keeps host-specific code outside the static reconciliation core. It does not deliver agent context, inspect prompts, or make the repository demon part of `check` and `fix` correctness.

## Public commands

### Acquire

```bash
demon acquire --client NAME [PATH]
# equivalent:
ddocs demon acquire --client NAME [PATH]
```

`PATH` may point anywhere inside the intended initialized repository. On the first mutating demon entry into a linked Git worktree, acquire may bootstrap independent local `.ddocs/` configuration and object storage.

Successful output is:

```text
token=TOKEN claimed=true|false
```

`claimed=true` means this caller caused a fresh detached owner to start. `claimed=false` means a fresh owner already served the repository.

The adapter must retain `TOKEN` privately for later heartbeats and release.

### Heartbeat

```bash
demon heartbeat --token TOKEN [PATH]
```

A heartbeat refreshes the feeder record, clears a pending shutdown request, and recovers the repository demon when its owner lease is missing or stale.

The default runtime checks feeder state every 5 seconds and expires a feeder after 20 seconds without a refresh. An adapter should heartbeat well before expiry; approximately every 5 seconds matches the current runtime cadence.

Successful heartbeat output is intentionally empty.

### Release

```bash
demon release --token TOKEN [PATH]
```

Release removes only the named feeder. Other shell or agent feeders remain active. When no fresh feeders remain, the owner exits after the default 20-second shutdown grace.

Removing an already absent feeder is safe. The adapter should still treat a release command failure as an operational diagnostic.

## Required adapter lifecycle

A host adapter should implement:

```text
resolve the repository path
-> acquire with a stable client name
-> parse and retain the returned token
-> start periodic heartbeats
-> run the host job
-> stop the heartbeat loop
-> release the token on every completion path
```

Release is required after:

- success;
- ordinary failure;
- cancellation;
- user interruption;
- timeout;
- child-process spawn failure;
- adapter initialization failure after acquire; and
- partial startup where a heartbeat loop never began.

Use one token per independently owned host session. Do not copy tokens between repositories or linked worktrees.

## Client naming

`--client NAME` is diagnostic metadata. Use a stable, bounded name such as:

```text
mcp
codex
hermes
editor-extension
ci-helper
```

Do not encode secrets, prompts, repository contents, or unbounded job text into the client name.

## Repository and worktree selection

Pass an explicit repository-contained `PATH` when the host process does not reliably run with the repository as its current directory.

Each linked worktree receives independent feeder and owner state. A token acquired in one worktree must not be heartbeated or released against another.

Read-only status does not bootstrap a linked worktree:

```bash
demon --status PATH
```

Acquire is allowed to bootstrap because it is a mutating lifecycle operation.

## Failure behavior

### Demon disabled

Acquire and heartbeat fail when `[demon].run = false`. The adapter should stop feeder activity and continue only if its core task does not require watcher automation.

The user can enable it explicitly:

```bash
demon run --true PATH
```

### Repository not found

The command fails when the path is outside an initialized repository and is not a linked worktree whose primary worktree has an initialized Demon Docs repository.

### Invalid or expired token

Heartbeat fails for an unknown or invalid token. Reacquire a new feeder rather than inventing or reusing token text.

### Owner start failure

Acquire removes the newly created feeder when it cannot establish or start an owner. The adapter should surface the command error and may retry only under its normal bounded retry policy.

### Host crashes before release

The feeder expires after the heartbeat deadline. The owner eventually exits after no fresh feeders remain for the shutdown grace period. Explicit release is still preferred because it shortens stale activity and improves status accuracy.

## Disabling and recovery

To disable repository-demon automation, remove all feeders, and request shutdown:

```bash
demon run --false PATH
```

Inspect state and logs with:

```bash
demon --status PATH
demon --logs PATH
```

Adapters should use only the public `acquire`, `heartbeat`, and `release` commands. Hidden `__enter`, `__leave`, `__feed`, `__serve`, and `__shutdown` commands are implementation surfaces for generated shell integration and detached ownership.

## Verification

A basic adapter smoke test should verify:

```text
acquire returns one token
status reports one active agent
heartbeats keep the feeder fresh
release removes that agent feeder
another feeder is not removed
owner exits after the final feeder and grace period
```

Focused implementation coverage:

```bash
go test ./internal/demon ./internal/app -count=1
```

## Related docs

- [Repository Demon](repository-demon.md)
- [Watcher and Automation](watcher-and-automation.md)
- [Using Linked Git Worktrees](../guides/linked-worktrees.md)
- [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Recovery and Troubleshooting](recovery-and-troubleshooting.md)

## Notes

The feeder protocol is lifecycle plumbing only. Deterministic task-context delivery belongs to the planned Grimoire Context sibling tool and must not be inferred from a running Demon Docs agent feeder.
