---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-71c0-9170-576569cd158c
document_type: general
policy_exempt: false
summary: This document explains which filesystem locations Demon Docs observes, how watch scope changes while a process is running, how ignore-policy and external-target updates affect observation, and how to recover from watch-scope failures.
---
# Dynamic Watch Scope

Parent index: [Operations](./INDEX.md)

## Purpose

This document explains which filesystem locations Demon Docs observes, how watch scope changes while a process is running, how ignore-policy and external-target updates affect observation, and how to recover from watch-scope failures.

## Overview

Demon Docs computes watch scope from the selected features rather than watching every reachable filesystem path.

Documentation-index, frontmatter, and document-format watch modes are bounded to the documentation root. Document-format selection also observes the configured shared and document-specific schema directories. Link-enabled watch observes the repository root because changes to non-Markdown local targets can require Markdown link repair. Reverse-index watch observes configured code roots and the ancestor directories needed to discover new scope folders. Explicit external link targets may add bounded watches outside the repository at their nearest existing parent directories.

Watch scope is dynamic. New repository directories can be added while the watcher runs, `.docignore` changes can reveal previously excluded directories, deleted or renamed watched directories are removed from internal tracking, and link reconciliation can discover new external target parents.

Dynamic observation changes which events can schedule future reconciliation. It does not itself mutate indexes, links, reverse indexes, private state, or authored content.

## Scope by selected feature

### Documentation policy only

When indexes, frontmatter, or document format are selected without links, the base watcher observes the documentation root. When document format is selected, it additionally watches the configured schema directories and their nearest existing ancestors.

Relevant content is limited by:

- repository containment;
- `.docignore` policy;
- permanent traversal exclusions;
- configured watch ignored directories;
- configured ignored filename suffixes;
- temporary editor-file filtering; and
- indexable document rules.

Frontmatter and document-format selection use their own non-ignored Markdown policy; they are not limited by index include/exclude membership.

Directory removal and rename events remain relevant because they can change a parent index even after the directory no longer exists.

### Links enabled

When link reconciliation is selected, the base watcher observes the repository root.

This wider scope is necessary because a Markdown source may refer to:

- another Markdown document;
- an image;
- a code file;
- a generated artifact that is not ignored;
- a directory-local asset; or
- another supported repository-contained target.

A create, write, rename, or removal event for a non-ignored repository target can therefore require link-state refresh or Markdown repair even when the target is not itself a document.

Link-enabled event relevance still applies `.docignore` and configured watch filters. The watcher does not treat every event below the repository root as actionable.

### Reverse indexes only

Reverse-index watch observes configured code roots and discovered non-ignored directories below them.

It also watches ancestor directories between the repository root and each configured root. Ancestor watches allow a missing or later-created root path to become observable without restarting the process.

Reverse-index relevance is bounded to:

- paths inside one of the resolved reverse roots; and
- `.docignore` control-file changes that can alter traversal.

### Mixed base and reverse watch

When reverse indexes are selected together with forward indexes or links, Demon Docs runs the base watcher and reverse-index watcher concurrently.

Their observer scopes remain separate. Their mutation-capable reconciliation callbacks share one run lock so they do not apply repository changes concurrently.

See [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md) for run ownership and error joining.

## Initial watch construction

The base watcher performs its initial reconciliation before creating an observer.

After the initial run succeeds, it:

1. loads the repository ignore policy;
2. creates the filesystem observer;
3. optionally watches the repository root separately when the selected watch root is narrower;
4. recursively adds non-ignored directories below the selected watch root;
5. adds current external target parent directories; and
6. enters the event loop.

The separate repository-root watch lets a docs-root watcher observe repository-owned control files such as `.docignore` even when the docs root is a subdirectory.

If any required directory cannot be added, startup fails. Demon Docs does not silently continue with incomplete observation.

## Recursive directory admission

The base watcher uses a recursive directory walk when building or refreshing scope.

For every directory:

- repository and nested `.docignore` policy is evaluated;
- configured ignored directory names are applied;
- already watched directories are skipped; and
- accepted directories are added individually to the observer.

Ignored directories are pruned from traversal rather than walked and filtered file by file.

The `watchedDirs` map records directories successfully admitted to the observer. It prevents duplicate additions and provides directory identity for removal and rename events that may arrive after the path no longer exists.

## Newly created directories

On a create event inside the current watch root, the base watcher checks whether the new path is a directory.

When it is a directory and remains visible under current ignore and configured-watch policy, the watcher recursively adds that directory and any existing visible descendants.

The same event is then evaluated for reconciliation relevance. Creating a directory can therefore both expand future observation and schedule a reconciliation for the current structural change.

If recursive addition fails, the watcher returns the error. It does not keep running while pretending the new subtree is covered.

## Deleted and renamed directories

Filesystem observers can report removal or rename events after a watched directory is no longer stat-able.

The base watcher records whether the event path was previously in `watchedDirs` before removing it from internal tracking. That remembered directory status is passed to relevance filtering so forward-index structural changes can still schedule reconciliation.

The internal map entry is removed for both remove and rename operations.

The watcher does not attempt to infer the destination of a directory rename from one event. Source and destination events are processed independently, and reconciliation scans current repository state.

## `.docignore` reload

Any repository-owned `.docignore` control-file event is always relevant to watch management.

The base watcher responds by:

```text
reload the complete ignore policy
-> recursively walk the current watch root again
-> add directories that are now visible and not already watched
-> mark reconciliation pending
```

This allows a directory excluded by a previous rule to become observed after the rule is removed.

The refresh is additive at the operating-system observer layer. Directories newly ignored are filtered by the updated policy and no longer schedule normal content reconciliation, even if the platform watcher still has an existing registration for them.

The implementation does not remove every newly ignored directory from the observer immediately. Correctness depends on current policy filtering, not on the absence of an operating-system watch handle.

If the new ignore policy cannot be loaded or newly visible directories cannot be added, the watcher exits with an error.

## Configured watch filters

In addition to `.docignore`, the base watcher applies configured watch exclusions:

- ignored directory names;
- ignored filename suffixes; and
- filenames beginning with `.#`.

These filters affect recursive directory admission and event relevance.

They are configuration-driven process state. Changing the configuration file does not currently imply automatic live reload of the entire selected configuration. Restart the watcher after changing watch configuration.

## External link target watches

Link reconciliation can track explicit targets outside the repository boundary.

After each successful link reconciliation, Demon Docs derives the current set of external records from the link files manifest. For each external path, it finds the nearest existing directory at or above the target's parent.

The resulting directory list is:

- de-duplicated;
- sorted for deterministic processing; and
- added to the base observer if it is not already present.

Watching the nearest existing parent allows creation, deletion, or movement below that parent to schedule a new link scan even when the exact target path is currently absent.

External events are considered relevant when their path is contained by any currently tracked external watch directory.

## External scope growth and removal

External watch additions occur:

- after the initial link reconciliation, before the event loop begins; and
- after later successful link reconciliation runs discover new external target parents.

External watch tracking is additive except when an event directly removes or renames a tracked external directory, in which case its internal map entry is removed.

If a link is deleted and an external directory is no longer needed, the current implementation does not proactively remove an existing operating-system watch handle solely because the manifest stopped referencing it. Events from that directory can still arrive and be treated as external while the directory remains in the internal map.

A process restart reconstructs external scope from current link state and removes obsolete additive history.

## Reverse-index dynamic discovery

The reverse-index watcher maintains its own set of observed directories.

Its refresh function:

1. discovers current visible folders under configured roots;
2. adds ancestor directories required to detect future root creation;
3. sorts the directory set;
4. adds only directories not already watched; and
5. records successful additions.

Refresh requests are delivered through a buffered channel with capacity one. Multiple requests collapse while one request is pending.

A dedicated refresh worker executes the directory discovery and addition. Refresh results return to the main event loop through a result channel.

A refresh is requested when:

- a `.docignore` event occurs;
- a directory is created inside a reverse root; or
- a reverse reconciliation run completes successfully.

This keeps future observation aligned with newly created folders and changed ignore policy.

Like the base watcher, reverse-index refresh is additive. It does not remove operating-system watches solely because a directory becomes ignored or leaves the current discovered set.

## Event filtering and suppression

Observation is broader than reconciliation admission.

Before a repository-contained link event is considered normally, the base watcher asks the link subsystem whether the event matches a pending generated-write suppression.

- A matching generated-write event is consumed and ignored.
- A mismatching file hash invalidates the suppression and the event continues normally.

This prevents self-generated rewrites from causing an endless watch loop while preserving concurrent user edits.

Ignore policy and configured watch filters are applied after suppression handling for ordinary repository events.

## Observer errors

Demon Docs treats observer errors as operational failures.

Examples include:

- failure to create the watcher;
- failure to add the initial tree;
- failure to add a newly created directory;
- failure to add an external target parent;
- failure to refresh reverse-index discovery; and
- an error received from the filesystem observer error channel.

The current watch process exits rather than continuing with an unverified partial scope.

The repository demon records the watcher failure in its bounded logs and its owner lifecycle ends. A later feeder heartbeat can start a replacement owner after the lease becomes missing or stale.

## Cancellation and normal shutdown

Foreground cancellation and repository-demon shutdown cancel the watcher context.

The base watcher returns cleanly when the context is done and closes its observer.

The reverse-index watcher cancels its refresh worker and closes and drains the observer so buffered platform events do not block shutdown.

A clean watcher exit does not imply that every path event was individually processed. Correctness is re-established by a later `ddocs check`, `ddocs fix`, `ddocs watch --once`, or replacement demon owner.

## Operational diagnostics

### A new directory is not being reconciled

Check:

1. the directory is inside the selected watch root or reverse root;
2. no root or nested `.docignore` rule excludes it;
3. its name is not in configured ignored directories;
4. the watcher did not report a dynamic-add error; and
5. the watcher was running when the directory was created.

Restarting the watcher rebuilds scope from current filesystem state.

### A newly visible directory remains inactive after `.docignore` changes

Inspect watcher output for ignore-policy load or directory-add errors. Then run:

```bash
ddocs watch --once
ddocs check
```

A process restart reconstructs all watches from the current policy.

### External target changes are not detected

Confirm the link is represented as an explicit external target and that its nearest existing parent was accessible when link reconciliation ran. Run a link-enabled one-shot reconciliation to refresh the external manifest before restarting watch.

### Unexpected events continue from an ignored directory

The operating-system watch may still exist after a policy change, but current event filtering should reject the path. If reconciliation is still triggered, inspect nested `.docignore` precedence and configured watch filters, then restart the watcher to discard additive watch history.

### Watch exits after a filesystem error

Treat the process as stopped. Correct any permission, deleted-root, or observer-resource problem, run a static check, and start watch again. Do not assume an exited watcher still covers part of the tree.

## Invariants

Dynamic scope must preserve these invariants:

- Every required initial directory is watched or startup fails.
- Newly visible directories are added or the watcher reports failure.
- Events are filtered against current ignore policy even when an old watch handle remains.
- Directory removal and rename can schedule structural reconciliation after the path disappears.
- External watches are derived from explicit link-state scope rather than arbitrary filesystem reachability.
- Reverse-index roots remain bounded by resolved configuration and repository containment.
- Additive watch history never changes the authoritative reconciliation scope.
- Static `check` and `fix` remain valid recovery paths when observation is incomplete or stopped.

## Verification

Focused verification:

```bash
go test ./internal/watch ./internal/reverseindex ./internal/app -count=1
```

Important contracts include:

- source and destination rename handling;
- nested directory addition;
- `.docignore` reload and newly visible directories;
- dynamic-add error propagation;
- deleted watched-directory reconciliation;
- external target parent derivation;
- reverse-root refresh; and
- clean cancellation and watcher draining.

## Code map

Primary implementation:

- `internal/watch/watch.go` — base observer construction, recursive directory admission, control-file reload, removal tracking, event intake, and shutdown.
- `internal/watch/features.go` — selected-feature relevance, external target parent derivation, and external watch admission.
- `internal/reverseindex/watch.go` — reverse-root discovery, ancestor watches, refresh worker, and additive reverse watch management.
- `internal/ignore/` — repository and nested `.docignore` policy.
- `internal/links/state.go` and `internal/links/suppression.go` — link manifest scope and generated-write suppression.
- `internal/app/reverse_index.go` — mixed watcher lifecycle and shared cancellation.

Focused tests:

- `internal/watch/watcher_contract_test.go`
- `internal/watch/filter_test.go`
- `internal/watch/external_test.go`
- `internal/watch/watch_test.go`
- `internal/reverseindex/watch_test.go`
- `internal/app/reverse_index_test.go`

## Related docs

- [Watcher and Automation](watcher-and-automation.md)
- [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md)
- [Ignore and Traversal](../architecture/ignore-and-traversal.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Reverse Index Architecture](../architecture/reverse-indexes.md)
- [Repository Demon](repository-demon.md)
- [Recovery and Troubleshooting](recovery-and-troubleshooting.md)

## Notes

Watch handles are an operational optimization. Current filesystem scans, ignore policy, and reconciliation plans remain authoritative when additive observer state is broader than current scope.
