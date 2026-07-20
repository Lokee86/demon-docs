---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-78ad-bdd8-091628f2a7fb
document_type: general
policy_exempt: false
summary: This document provides a safe diagnostic and recovery sequence for configuration mistakes, unexpected writes, ambiguous links, stale watchers, repository-demon ownership problems, and damaged private state.
---
# Recovery and Troubleshooting

Parent index: [Operations](./INDEX.md)

## Purpose

This document provides a safe diagnostic and recovery sequence for configuration mistakes, unexpected writes, ambiguous links, stale watchers, repository-demon ownership problems, and damaged private state.

## Overview

Recovery should preserve evidence before resetting state. Most problems can be isolated by stopping automation, inspecting repository/configuration selection, running narrow read-only checks, and reviewing logs or diffs. Deleting `.ddocs/` is a last resort because it discards historical identity evidence.

## First response

Use this order:

```text
1. Stop foreground watchers and repository-demon feeders.
2. Preserve the current Git diff and relevant logs.
3. Inspect selected repository and configuration paths.
4. Run the narrowest applicable check.
5. Correct authored files or configuration.
6. Run fix only after the plan is understood.
7. Verify with a second fix and check.
```

Commands:

```bash
demon --status
demon --logs
ddocs status
ddocs config paths
ddocs config show
ddocs check --docs
ddocs check --links
ddocs check --reverse
```

## Unexpected index changes

Check:

- selected `docs_root`;
- `index_file` and section headings;
- include/exclude patterns;
- draft-folder name;
- parent-link toggles;
- `.docignore`; and
- whether authored text was placed inside managed markers.

Demon Docs reconciles managed blocks back to the configured filesystem model. Put hand-authored guidance outside markers.

## Unexpected link rewrites

Review the exact source diff. Demon Docs should change only the target path portion of a recognized link.

Check persistent identity/history and whether the destination was uniquely determined. If the label, title, alias, fragment, query, prose, or unrelated content changed, preserve the failing input and treat it as a bug rather than accepting the write.

## Broken or ambiguous links

A broken link with one deterministic historical or fingerprint target may be repairable. Multiple plausible targets are intentionally not selected.

Resolve ambiguity manually:

```text
choose the intended destination
edit the authored source
run ddocs fix --links
run ddocs check --links
```

Do not rename candidates merely to force the algorithm to guess differently unless the repository naming itself is wrong.

## First-pass link limitations

Without prior identity state, Demon Docs cannot know where a target used to live. The first link-enabled mutating pass establishes a baseline.

Repair current broken links manually before relying on later move reconciliation.

## Watcher appears to race manual work

Stop the foreground watcher or repository demon. Check logs for a reconciliation that completed between the manual edit and a later command.

A manual `fix` reporting zero changed files after an edit may mean the watcher already applied the deterministic result.

Do not run multiple detached wrappers around the same repository.

## Repository demon problems

Use:

```bash
demon --status
demon --logs
```

Check feeder activity, owner freshness, shutdown state, linked-worktree selection, and runtime files under `.ddocs/runtime/`.

Release feeder tokens on all success, failure, cancellation, timeout, and spawn-failure paths. A host adapter that leaks heartbeats can keep the watcher alive longer than intended.

Static commands can be used after the active owner stops or stale ownership is recovered.

## Configuration mismatch across environments

Compare:

```bash
ddocs config paths
ddocs config show
```

Likely differences include:

- explicit `--config` use;
- upward discovery of `.ddocs/config.toml`;
- current-directory legacy configs;
- global config location;
- platform path case or separators;
- `.docignore` contents; and
- generated or untracked files present in only one environment.

## Damaged private state

Before resetting:

```text
stop automation
copy or preserve .ddocs/ and logs
record command output
record the Git diff
try the narrowest read-only command
```

If state cannot be decoded or recovered and no targeted repair exists, move or delete `.ddocs/`, reinitialize, and establish a new baseline.

Consequences include losing prior move identity/history and resetting repository-demon runtime state.

## Verification after recovery

Run:

```bash
ddocs fix
ddocs fix
ddocs check
```

Then run the project's normal test or release gate when recovery involved configuration, code, fixtures, or broad document moves.

## Related docs

- [Getting Started](../guides/getting-started.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Repository Demon](repository-demon.md)
- [Watcher and Automation](watcher-and-automation.md)

## Notes

Preserve a failing fixture before broad cleanup. Deterministic reconciliation bugs are easiest to fix when the exact source bytes, state, configuration, and command are retained.
