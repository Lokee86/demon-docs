---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-79ca-b3ff-0b46a1bbedff
document_type: general
policy_exempt: false
summary: This document describes foreground watch behavior, event scope, debounce and serialization, output, and the boundary between explicit terminal ownership and the repository demon.
---
# Watcher and Automation

Parent index: [Operations](./INDEX.md)

## Purpose

This document describes foreground watch behavior, event scope, debounce and serialization, output, and the boundary between explicit terminal ownership and the repository demon.

## Overview

Demon Docs exposes the same watcher through two operational surfaces:

- `ddocs watch` runs explicitly in the foreground;
- the [Repository Demon](./repository-demon.md) manages a detached watcher while shells or agents are actively feeding it.

Both surfaces call the same deterministic reconciliation core. Neither is required for `ddocs check` or `ddocs fix`.

Foreground `ddocs watch` does not require repository initialization. Without `.ddocs/config.toml`, it uses the standalone docs-root scope and may create link state beneath that root. The detached repository demon is different: it requires an initialized repository because its configuration, ownership lease, feeders, shutdown state, and logs are repository-local.

## Operating model

Foreground watch performs one reconciliation immediately, observes relevant filesystem locations, debounces noisy event bursts, and runs one reconciliation at a time. Changes arriving during a pass schedule one follow-up pass.

The watcher is optional automation. A later `ddocs check` must be able to verify the same repository state without the watcher running.

Codemap generation is not a watcher feature. Neither foreground watch nor the repository demon invokes `codemap fix`, `codemap check`, recommendation planning, managed codemap adoption, or confidence pruning. A watcher may observe Markdown files written by an explicit codemap command and reconcile its own selected systems, but it does not regenerate the codemap or enqueue a codemap-specific follow-up.

## Watch Commands

```bash
ddocs watch --root docs
ddocs watch --root docs --once
ddocs watch -i
ddocs watch --frontmatter
ddocs watch --format
ddocs watch -l
```

`ddocs watch --help` shows the watch-specific flags and examples.

`--once` runs one reconciliation pass and exits. Regular watch mode runs one reconciliation immediately and then observes relevant filesystem changes until the foreground process is stopped.

Documentation-index, frontmatter-only, and document-format-only modes watch the docs root. Link-enabled mode watches the repository root so changes and moves involving non-Markdown targets can trigger link reconciliation.

## What the Watcher Does

The watcher reruns the same selected operations used by `fix` when relevant repository content changes.

- It starts with an immediate reconciliation pass.
- It watches the docs root for indexes, frontmatter, or document format, or the repository root when links are enabled.
- It reacts to relevant file events and directory create, delete, and move events.
- It debounces event bursts.
- It runs one reconciliation at a time.
- If changes arrive during a run, it schedules one follow-up pass.
- It applies `.docignore`, configured ignored directories, ignored suffixes, and index include/exclude rules.
- It observes Markdown source changes and changes to non-ignored local link targets.
- It observes configured shared and document-specific schema directories when document format is selected, so schema edits trigger a new plan.
- Explicit external targets add watches on their nearest existing parent directories.
- It adds watches for newly created nested directories and removes deleted or renamed watched directories.
- Observer errors are surfaced rather than silently terminating observation.
- Each reconciliation diagnostic is printed as its own watcher message instead of being collapsed into an opaque count.

Generated Markdown rewrites record their expected content hash and affected link IDs before watcher feedback is processed. A matching event is consumed as the expected self-write. A mismatched hash invalidates that suppression and the file is processed normally, preserving concurrent user edits.

## Foreground Watch versus Repository Demon

Use foreground `watch` when you want standalone operation, a process attached to the current terminal, directly visible output, or manually controlled lifetime.

Use the repository demon for self-managed local operation only after initializing the repository. Do not wrap `ddocs watch` in an additional PID-file, `setsid`, scheduled-task, or shell-startup daemonization script when the repository demon is enabled. A second lifecycle wrapper can create competing watchers and misleading status.

The repository demon owns detached process startup, single-owner coordination, feeder heartbeats, shutdown grace, stale-owner recovery, and bounded repository-local logs. See [Repository Demon](./repository-demon.md).

## Output

Foreground watcher output includes timestamped status lines and the current process ID:

```text
2026-06-18T23:59:59 ddocs watch watching docs pid=12345
2026-06-18T23:59:59 ddocs watch updated 3 file(s)
```

Detached watcher output is written to the bounded log set under `.ddocs/runtime/logs/` and is available through:

```bash
ddocs demon --logs
```

## Practical Usage

- Use `ddocs watch` for an explicit foreground session.
- Use the repository demon for automatic shell or agent-driven lifecycle.
- Use `ddocs check` before commit or in CI.
- Use `ddocs fix` for deterministic recovery or a deliberate one-shot repair.
- Use explicit `ddocs codemaps inspect|fix|check` when managed codemap generation is desired.
- Expect `fix` to report zero updated files when a watcher already reconciled the tree.
- Do not assume watcher health implies codemap-generation convergence.

## Test Coverage

Watcher unit and temporary-filesystem integration tests cover source and destination rename events, nested directory creation, watched-directory deletion, configured filtering, operation selection, events queued during reconciliation, explicit debounce overrides, observer errors, clean cancellation, and self-write convergence.

Repository-demon tests separately cover ownership exclusion and stale recovery, feeder expiry and counting, read-only status snapshots, shell-feeder reuse, bounded logs, shutdown grace, linked-worktree discovery, persistent enablement, and generated shell-hook contracts.

## Code map

- `internal/watch/watch.go` — observer setup, watched scopes, event filtering, and reconciliation execution.
- `internal/watch/scheduler.go` — debounce, single-run ownership, and queued follow-up scheduling.
- `internal/watch/features.go` — selected reconciliation feature contract.
- `internal/demon/runtime.go` — detached owner and feeder lifecycle around the same watcher.
- `internal/demon/log.go` — bounded detached watcher logs.
- `internal/app/demon.go` — daemon CLI and generated shell hooks.
- `internal/repository/worktree.go` — linked-worktree runtime isolation.

## Failure and recovery

Stop foreground watch before diagnosing unexpected changes. If a manual `fix` reports no changes after an edit, inspect watcher output because the watcher may already have reconciled the tree.

Do not add a second detached wrapper when the repository demon owns the watcher.

## Verification

```bash
ddocs watch --once
ddocs check
```

Focused tests:

```bash
go test ./internal/watch ./internal/app -count=1
```

## Related docs

- [Operations](INDEX.md)
- [CI and Automation](../guides/ci-and-automation.md)
- [Repository Demon](repository-demon.md)
- [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md)
- [Dynamic Watch Scope](dynamic-watch-scope.md)
- [Recovery and Troubleshooting](recovery-and-troubleshooting.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)

## Notes

Link-enabled watch may observe repository and bounded external parent paths because non-Markdown target moves can require Markdown source repair.
