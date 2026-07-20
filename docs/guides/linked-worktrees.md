---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7dc6-8055-c44e60f2f0f5
document_type: general
policy_exempt: false
summary: This guide explains how Demon Docs state is isolated in linked Git worktrees and how to bootstrap repository-demon operation without copying runtime or historical state from the primary worktree.
---
# Using Linked Git Worktrees

Parent index: [Guides](./README.md)

## Purpose

This guide explains how Demon Docs state is isolated in linked Git worktrees and how to bootstrap repository-demon operation without copying runtime or historical state from the primary worktree.

## Overview

A linked Git worktree shares Git history with its primary worktree but has a separate filesystem snapshot. Demon Docs therefore gives each worktree its own `.ddocs/` object repository and runtime directory.

The current Git-aware bootstrap exists at the repository-demon boundary. Read-only linked-worktree detection can locate the primary worktree configuration without writing. A demon command that requires local mutable state can copy that configuration into the linked worktree and initialize fresh local object storage.

Normal repository discovery remains filesystem-based and independent of Git once the linked worktree has its own `.ddocs/config.toml`.

## Prerequisites

- The primary worktree is initialized with `.ddocs/config.toml`.
- The linked worktree was created by Git and contains the normal `.git` worktree pointer file.
- The configured docs root exists in the linked worktree snapshot.
- The primary configuration uses repository-relative paths that are valid in both worktrees.

## Bootstrap through the repository demon

From anywhere inside the linked worktree, run a demon command that acquires mutable repository-local lifecycle state:

```bash
demon run
```

Agent adapters may instead acquire a feeder:

```bash
demon acquire --client mcp
```

During linked-worktree detection, Demon Docs follows the `.git` pointer and `commondir` metadata to find the primary worktree. When the primary worktree is initialized and the linked worktree is not, bootstrap:

```text
copies primary .ddocs/config.toml
creates linked-worktree .ddocs/ object storage
creates no copied owner, feeder, heartbeat, log, or link-history state
```

After bootstrap, the linked worktree is discovered as its own Demon Docs repository.

## Verify local isolation

```bash
ddocs status
ddocs config paths
demon --status
```

The reported repository root, config path, `.docignore`, object storage, and runtime state should all belong to the linked worktree.

The configuration contents initially match the primary worktree, but later edits are ordinary local file edits. Demon Docs does not continuously synchronize configuration between worktrees.

## Run reconciliation

After local bootstrap:

```bash
ddocs fix
ddocs check
```

Each worktree reconciles its own current files. Stable identities, path history, review events, undo eligibility, watchers, and runtime ownership are not shared automatically between worktrees.

This isolation prevents a path or generated write observed in one working directory from being treated as the current filesystem state of another.

## Worktree cleanup

Before deleting a linked worktree:

```bash
demon --status
demon run --false
```

Release any agent feeder tokens and stop foreground watchers. Git may then remove the worktree normally. The linked worktree's `.ddocs/` state disappears with that working directory; the primary worktree state is unaffected.

Do not copy `.ddocs/runtime/` between worktrees. Do not reuse feeder or owner tokens across worktrees.

## Expected result

- Each worktree has an independent `.ddocs/` directory.
- Only configuration is copied during initial linked-worktree bootstrap.
- Object history and runtime ownership begin fresh in the linked worktree.
- Commands resolve paths against the current worktree.
- The primary worktree remains unchanged by linked-worktree bootstrap.

## Failure and recovery

### No linked worktree is detected

Confirm the worktree has a `.git` pointer file and that its referenced Git worktree directory contains a valid `commondir` file.

### The primary worktree is not initialized

Initialize the intended worktree explicitly with `ddocs init --root PATH`. Automatic linked bootstrap requires the primary worktree's `.ddocs/config.toml`.

### The linked marker exists but is not a directory

Remove or rename the conflicting `.ddocs` filesystem entry after reviewing it. Bootstrap refuses to replace a non-directory marker.

### The copied configuration selects an invalid docs root

Update the linked worktree's `.ddocs/config.toml` or ensure the expected directory exists in that branch. Configuration is copied, not adapted to branch-specific layout changes.

### State from another worktree appears to be in use

Stop active automation, inspect `ddocs status`, and verify that commands are running inside the intended worktree. Preserve diagnostics before deleting any `.ddocs/` state.

## Related docs

- [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Repository Demon](../operations/repository-demon.md)
- [Host Adapter Feeder Integration](../operations/host-adapters.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Git history is shared; Demon Docs working state is not. Branches may contain different docs, targets, and generated output, so sharing one live `.ddocs/` state across worktrees would be unsafe.
