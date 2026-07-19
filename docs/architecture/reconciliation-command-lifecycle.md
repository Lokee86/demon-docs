# Reconciliation Command Lifecycle

Parent index: [Architecture](./README.md)

## Purpose

This document defines how `ddocs check`, `ddocs fix`, and `ddocs watch` resolve configuration and scope, select reconciliation systems, order planners and writes, aggregate diagnostics, and return command results.

## Overview

The reconciliation commands coordinate three independently owned systems:

```text
documentation indexes
repository-local links and orphan health
reverse indexes
```

`internal/app` selects and orders those systems. It does not merge them into one storage transaction. Each subsystem retains its own plan, apply boundary, diagnostics, and persistent state.

This distinction is central to failure behavior: a mutating `fix` command can complete an earlier subsystem before a later subsystem fails. The command is deterministic and bounded, but it is not an all-or-nothing transaction across every authored file and private state store.

## Code root

```text
internal/app/
cmd/ddocs/
cmd/demon/
```

## Responsibilities

The application boundary owns:

- top-level dispatch into `check`, `fix`, and `watch`;
- command-scoped flag parsing and help behavior;
- configuration selection and CLI overrides;
- repository and docs-root scope resolution;
- feature-selector defaults;
- reverse-root and codemap-heading option resolution;
- precondition checks before subsystem planning;
- planner and applier ordering;
- diagnostic and update-path rendering;
- aggregate exit-code decisions; and
- handoff into foreground watch scheduling.

## Does not own

The application boundary does not own:

- documentation-tree scanning or managed-block transformation;
- link parsing, target resolution, generated rewrites, or link-state publication;
- orphan relationship semantics;
- reverse-index inventory or rendering;
- watcher debounce and concurrency mechanics;
- private object-repository transactions; or
- review-ledger publication.

It composes those owners without weakening their individual safety checks.

## Entry and dispatch

Both executables eventually enter `internal/app`.

```text
cmd/ddocs main
-> app.Run
-> top-level command dispatch
-> runTree for check, fix, or watch
```

The `demon` executable normalizes repository-demon invocations, but normal reconciliation behavior remains owned by the same application package.

Argument and help failures return before configuration or repository discovery.

## Configuration and scope lifecycle

Every tree command follows the same resolution sequence.

```text
parse command flags
-> select one configuration file or built-in defaults
-> apply command-line overrides
-> resolve repository and docs-root scope
-> resolve selected features
-> validate docs-root and reverse-root preconditions
```

### Configuration selection

The selected configuration is loaded as one complete source. Local and global configuration files are not merged. CLI flags then override fields in the loaded `config.Config` for the current invocation.

Tree-command overrides can replace:

- docs root;
- index filename;
- draft folder and description prefix;
- include and exclude patterns;
- marker prefix;
- parent-link label and enablement;
- reverse roots; and
- codemap headings.

### Repository scope

`repository.ResolveScope` receives:

- current working directory;
- selected config path;
- configured docs root; and
- optional `--root` override.

The resulting scope owns the repository root, docs root, repository config, and ignore path used by all selected systems.

### Preconditions

If documentation indexes or reverse indexes are selected, the docs root must exist before planning begins.

Link-only operation can run without an existing docs root because repository-local links are scoped to the repository and may target Markdown outside the configured documentation tree. Orphan projection is skipped when the docs root does not exist.

Reverse selection additionally resolves and validates reverse roots and output format before command execution.

## Feature selection

The selectors are:

```text
-d, --docs, -i, --indexes
-l, --links
-r, --reverse
```

When any selector is supplied, only explicitly selected systems run.

When no selector is supplied:

```text
documentation indexes = enabled
links = enabled
reverse indexes = enabled only when reverse roots or codemap-heading overrides are present
```

The compatibility aliases `-i` and `--indexes` select the same documentation-index system as `-d` and `--docs`.

Feature selection is command-scoped. It does not mutate repository configuration.

## Shared planning model

The three subsystem plans are independent values:

- `model.ReconcileResult` for documentation indexes;
- `links.Plan` for links; and
- `reverseindex.Plan` for reverse indexes.

Each plan carries its own updates and diagnostics. The application layer does not translate one subsystem's internal state into another subsystem's plan.

The exact planning order differs between read-only and mutating commands because link reconciliation during `fix` must see documentation-index writes already applied.

## `check` lifecycle

`check` is read-only with respect to authored files and private reconciliation state.

### Planning order

```text
1. build documentation-index plan when selected
2. build reverse-index plan when selected
3. build link plan when selected
4. derive orphan documents from current link plan when links are selected and docs root exists
5. aggregate failure state
```

Reverse planning occurs before link planning in the current application code. The plans are independent, so this ordering does not grant either subsystem ownership of the other.

### Failure aggregation

`check` fails when any selected current condition is true:

```text
documentation-index updates are pending
link plan reports initialization, unresolved links, updates, or rewrites
reverse-index plan reports failure
orphan documents exist
```

`links.Plan.Failed` includes:

- uninitialized link state;
- unresolved link records;
- pending link-related file updates; and
- pending generated rewrites.

### Output order

When failed, output is rendered in this order:

```text
ddocs check failed
pending documentation-index paths
pending link-update paths
pending reverse-index paths
documentation-index messages
link messages
orphan messages
reverse-index diagnostics
```

A passing check prints only `ddocs check passed`.

### Exit behavior

- `0` — every selected system is already reconciled and no orphan health failure exists.
- `1` — reconciliation or health work is pending, but planning completed normally.
- `2` — argument, configuration, scope, planner, or I/O error prevented a valid check result.

The read-only command may load private state, but it does not publish a baseline or save a new link-state projection.

## `fix` lifecycle

`fix` plans and applies selected systems in a deliberate sequence.

### Pre-apply planning

Before the first authored write:

```text
1. build documentation-index plan when selected
2. build reverse-index plan when selected
```

The link plan is not built yet.

### Mutation order

```text
1. apply documentation-index updates
2. reconcile links against the resulting repository contents
3. apply generated link rewrites and publish link/review state
4. apply the previously built reverse-index plan
```

This order gives link reconciliation the current post-index Markdown content. It also keeps reverse-index application last.

### Changed-file count

The reported `updated N file(s)` count is the sum returned by each selected applier:

```text
index files changed
+ Markdown sources rewritten by link repair
+ reverse-index files changed
```

Private `.ddocs` objects, review events, and state records are not counted as updated files in this summary.

### Diagnostics

After all selected applies succeed, `fix` prints:

```text
ddocs fix updated N file(s)
documentation-index messages
link messages
reverse-index diagnostics
```

If link reconciliation completed with unresolved records, it additionally prints:

```text
ddocs fix unresolved N link(s)
```

and returns exit code `1` even though safe deterministic writes may already have been applied successfully.

### Partial-completion boundaries

There is no command-wide rollback across the three systems.

#### Index apply succeeds, link planning or apply fails

Documentation-index writes remain. The command returns an error. A later `check` or `fix` sees the new index contents.

#### Index and link apply succeed, reverse apply fails

Documentation-index writes, authored link rewrites, review events, and published link state remain. Reverse-index files may remain unchanged or partially governed by their own apply guarantees. The command returns an error.

#### Link apply fails inside its own publication lifecycle

The link subsystem applies its documented source-hash, rollback, review-history, refresh, and state-publication rules. Those internal guarantees do not extend backward to documentation-index writes or forward to reverse-index writes.

#### Unresolved links remain

Unresolved link conditions are not application errors. Deterministic repairs and other selected subsystem writes complete, diagnostics are printed, and the command returns `1`.

### Exit behavior

- `0` — selected writes and publications succeeded and no unresolved links remain.
- `1` — selected writes succeeded, but unresolved link conditions remain.
- `2` — argument, configuration, scope, planning, application, or I/O error interrupted the command.

## `watch` handoff lifecycle

After shared parsing, configuration, scope, feature selection, and precondition validation, `watch` hands control to `runSelectedWatch`.

```text
resolve optional debounce override
-> run one selected reconciliation immediately
-> if --once, return after that result
-> otherwise construct selected watchers and enter foreground scheduling
```

The watcher reuses the same subsystem planners and appliers. Scheduling, event relevance, dynamic watch scope, suppression consumption, and serialized follow-up runs belong to `internal/watch` and reverse-index watch coordination, not to this command lifecycle.

A foreground watch error returns exit code `2`. Normal context cancellation and scheduler shutdown follow the watcher ownership contract.

## Orphan-health integration

Orphan detection is a projection over the current link plan and managed documentation scope. It runs only for read-only `check` when links are selected and the docs root exists.

`fix` does not fail or write based on orphan status. Orphan documents are a health finding, not an automatic link-generation mechanism.

The application layer renders each orphan as:

```text
message: Orphan document: <path>
```

and includes the finding in the aggregate `check` result.

## Diagnostic ownership

Each subsystem produces its own messages or diagnostics. The application layer owns only ordering, prefixes, summaries, and exit aggregation.

- Documentation reconciliation messages are emitted from `model.ReconcileResult`.
- Link messages and unresolved counts are emitted from `links.Plan`.
- Reverse diagnostics retain their own structured status and rendering.
- Orphan paths are calculated by `internal/app/orphans.go` and rendered as health messages.
- Fatal errors are written to standard error as `ddocs error: ...`.

A diagnostic that represents unresolved or pending work normally returns `1`. A failure to construct a trustworthy result returns `2`.

## State and data ownership

The application layer stores no durable reconciliation graph.

| Data | Owner |
| --- | --- |
| Managed index plans and file updates | `internal/reconcile` and `internal/model` |
| Link identities, statuses, rewrites, and private state | `internal/links` |
| Review events and controls | `internal/review` |
| Reverse-index inventory and updates | `internal/reverseindex` |
| Orphan projection | `internal/app/orphans.go` using link and docs-scope inputs |
| Watch scheduling state | `internal/watch` |
| Command-scoped flags, resolved options, and aggregate result | `internal/app` |

## Invariants and safety boundaries

- `check` never calls subsystem authored-file appliers.
- Any supplied selector disables unselected default systems.
- Link reconciliation during `fix` observes successful documentation-index writes from the same command.
- Reverse-index application remains after link application.
- No subsystem failure is hidden by aggregate success output.
- Unresolved link conditions are distinguished from fatal command errors.
- Orphan health never creates guessed links.
- The application layer does not bypass subsystem containment or expected-content checks.
- Private state objects are not included in the authored-file update count.
- Command output order is deterministic.
- Cross-subsystem rollback is not promised.

## Failure behavior

### Parse failure

Invalid flags, missing arguments, or unexpected positional arguments return `2` before configuration selection.

### Configuration failure

A selected config that cannot be loaded returns `2`. No subsystem planner runs.

### Scope failure

Repository discovery, root containment, or docs-root resolution failure returns `2`. No subsystem planner runs.

### Planner failure

The first planner error aborts the command. For `check`, no authored writes have occurred. For `fix`, a later planner such as link reconciliation can fail after documentation-index writes have completed.

### Apply failure

The command stops at the failing subsystem and does not attempt later appliers. Earlier successful subsystem mutations remain.

### Output failure

Most summary writes use the provided writer without converting writer errors into command failures. Filesystem and subsystem errors remain authoritative. CLI embedding should provide reliable writers.

## Extension rules

When adding a new reconciled subsystem:

1. define an owning plan and apply boundary outside `internal/app`;
2. define selector/default behavior explicitly;
3. decide whether planning must occur before or after earlier writes;
4. document whether failures can leave prior subsystem writes durable;
5. add deterministic diagnostic ordering;
6. add exact exit-code behavior;
7. update `check`, `fix`, and `watch` consistently where applicable;
8. update scoped help and CLI reference; and
9. add focused orchestration tests.

Do not imply a command-wide transaction unless implementation provides rollback across every participating subsystem and state store.

## Code map

- `cmd/ddocs/main.go` — canonical executable entry.
- `cmd/demon/main.go` — alias normalization before shared application entry.
- `internal/app/app.go` — tree-command parsing, configuration/scope resolution, selection, plan/apply ordering, output, and exit results.
- `internal/app/reverse_index.go` — reverse-root resolution and reverse watch coordination.
- `internal/app/orphans.go` — read-only orphan projection.
- `internal/app/feature_flags_test.go` — selector and default-feature behavior.
- `internal/app/reverse_index_test.go` — reverse selection across check, fix, and watch.
- `internal/app/orphans_integration_test.go` — link-enabled health integration.
- `internal/reconcile/` — documentation-index planning and apply owner.
- `internal/links/` — link plan, generated rewrite, review event, and link-state owner.
- `internal/reverseindex/` — reverse-index plan and apply owner.
- `internal/watch/` — scheduler and foreground watch owner.

## Tests

Focused command-lifecycle coverage includes:

- `internal/app/app_test.go` — basic fix/check behavior, configuration, and overrides.
- `internal/app/feature_flags_test.go` — selected and default subsystem behavior.
- `internal/app/cli_contract_test.go` — command-line override contracts.
- `internal/app/reverse_index_test.go` — reverse planning and application.
- `internal/app/orphans_integration_test.go` — orphan failure aggregation.
- `internal/app/review_cli_test.go` — unresolved and blocked link exit behavior.
- `internal/app/help_test.go` and `help_nested_test.go` — public command/help synchronization.

Subsystem guarantees remain covered by their package tests.

```bash
go test ./internal/app ./internal/reconcile ./internal/links ./internal/reverseindex ./internal/watch -count=1
```

## Related docs

- [Application Orchestration](application-orchestration.md)
- [Reconciliation Pipeline](reconciliation-pipeline.md)
- [Link Reconciliation State Machine](link-reconciliation-state-machine.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [Reverse Indexes](reverse-indexes.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Document Health Checks](../guides/document-health-checks.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [CLI Reference](../reference/cli.md)

## Notes

The lifecycle described here is the current implementation contract. Future cross-subsystem transaction work would require a new ownership boundary rather than documentation language alone.
