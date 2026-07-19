# Watcher and Automation

Demon Docs exposes the same watcher through two operational surfaces:

- `ddocs watch` runs explicitly in the foreground;
- the [Repository Demon](repository-demon.md) manages a detached watcher while shells or agents are actively feeding it.

Both surfaces call the same deterministic reconciliation core. Neither is required for `ddocs check` or `ddocs fix`.

## Watch Commands

```bash
ddocs watch --root docs
ddocs watch --root docs --once
ddocs watch -i
ddocs watch -l
```

`ddocs watch --help` shows the watch-specific flags and examples.

`--once` runs one reconciliation pass and exits. Regular watch mode runs one reconciliation immediately and then observes relevant filesystem changes until the foreground process is stopped.

Index-only mode watches the docs root. Link-enabled mode watches the repository root so changes and moves involving non-Markdown targets can trigger link reconciliation.

## What the Watcher Does

The watcher reruns the same selected operations used by `fix` when relevant repository content changes.

- It starts with an immediate reconciliation pass.
- It watches the docs root for `-i`, or the repository root when links are enabled.
- It reacts to relevant file events and directory create, delete, and move events.
- It debounces event bursts.
- It runs one reconciliation at a time.
- If changes arrive during a run, it schedules one follow-up pass.
- It applies `.docignore`, configured ignored directories, ignored suffixes, and index include/exclude rules.
- It observes Markdown source changes and changes to non-ignored local link targets.
- Explicit external targets add watches on their nearest existing parent directories.
- It adds watches for newly created nested directories and removes deleted or renamed watched directories.
- Observer errors are surfaced rather than silently terminating observation.

Generated Markdown rewrites record their expected content hash and affected link IDs before watcher feedback is processed. A matching event is consumed as the expected self-write. A mismatched hash invalidates that suppression and the file is processed normally, preserving concurrent user edits.

## Foreground Watch versus Repository Demon

Use foreground `watch` when you deliberately want the process attached to the current terminal, its output visible directly, or its lifetime controlled manually.

Use the repository demon for normal self-managed local operation. Do not wrap `ddocs watch` in an additional PID-file, `setsid`, scheduled-task, or shell-startup daemonization script when the repository demon is enabled. A second lifecycle wrapper can create competing watchers and misleading status.

The repository demon owns detached process startup, single-owner coordination, feeder heartbeats, shutdown grace, stale-owner recovery, and bounded repository-local logs. See [Repository Demon](repository-demon.md).

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
- Expect `fix` to report zero updated files when a watcher already reconciled the tree.

## Test Coverage

Watcher unit and temporary-filesystem integration tests cover source and destination rename events, nested directory creation, watched-directory deletion, configured filtering, operation selection, events queued during reconciliation, explicit debounce overrides, observer errors, clean cancellation, and self-write convergence.

Repository-demon tests separately cover ownership exclusion and stale recovery, feeder expiry and counting, read-only status snapshots, shell-feeder reuse, bounded logs, shutdown grace, linked-worktree discovery, persistent enablement, and generated shell-hook contracts.

## Related Files

- `internal/watch/watch.go`
- `internal/watch/scheduler.go`
- `internal/demon/runtime.go`
- `internal/demon/log.go`
- `internal/app/demon.go`
- `internal/repository/worktree.go`
