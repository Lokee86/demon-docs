---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-7006-9c64-6b8a2ecdc3b2
document_type: general
policy_exempt: false
summary: This document describes the implemented application boundary that resolves configuration, selects subsystems, coordinates check, fix, and watch, and exposes stateless move, suggestion, change-history, codemap, and repository-demon...
---
# Application Orchestration

Parent index: [Architecture](./README.md)

## Purpose

This document describes the implemented application boundary that resolves configuration, selects subsystems, coordinates `check`, `fix`, and `watch`, and exposes stateless move, suggestion, change-history, codemap, and repository-demon command groups.

## Overview

Both `ddocs` and `demon` enter the same internal application package. The `demon` wrapper normalizes bare and help invocations into the repository-demon command family while preserving shared version handling. `internal/app` owns scoped command parsing and coordination across repository discovery, configuration, documentation reconciliation, local links, orphan health checks, stateless moves, review commands, reverse indexes, codemap analysis, watch mode, and demon lifecycle.

The application layer coordinates ownership. It does not reimplement scanner, link, reverse-index, repository-state, or daemon mechanics. The exact `check`, `fix`, and `watch` selection, planning, application, partial-completion, diagnostic, and exit lifecycle is owned by [Reconciliation Command Lifecycle](reconciliation-command-lifecycle.md).

## Code root

```text
internal/app/
cmd/ddocs/
cmd/demon/
```

## Responsibilities

The application boundary owns:

- top-level and nested-subcommand dispatch;
- scoped help and shared version behavior;
- repository and configuration selection;
- translation of CLI flags into subsystem options;
- `--docs`, `--links`, and `--reverse` selection semantics;
- ordering selected reconciliation systems;
- command output and aggregate success/failure decisions;
- orphan-document health-check integration for link-enabled checks;
- explicit stateless move command integration;
- suggestion selection/decline and applied-change inspection/undo command integration;
- foreground watch startup;
- explicit codemap fix, check, inspect, export, benchmark, and precision command integration;
- repository-demon public and hidden commands; and
- compatibility aliases such as `--indexes` and the `demon` executable.

## Does not own

The application layer does not own:

- documentation-tree scanning;
- managed Markdown parsing;
- link target parsing and identity evidence;
- private object repository encoding;
- reverse-index inventory or rendering;
- filesystem watcher scheduling;
- demon lease and runtime-state mechanics;
- codemap evidence scoring internals; or
- research corpus construction.

Those responsibilities remain in focused internal packages.

## Command flow

A normal command follows this shape:

```text
executable main
-> internal/app entry
-> parse command and flags
-> discover repository/configuration
-> resolve subsystem selection
-> invoke selected planners or command service
-> apply writes only for mutating commands
-> render diagnostics and summary
-> return process result
```

`check` and `fix` share the same underlying subsystem models. The difference is whether a safe plan is applied to authored repository files.

`watch` performs one normal reconciliation before entering event-driven scheduling. The repository demon eventually owns a watcher process, but its lifecycle commands still enter through the application boundary.

Codemap generation is a separate foreground command family. Canonical `ddocs codemaps fix|check|inspect`, with singular `codemap` retained as a compatibility alias, resolves a contained file-or-directory scope, builds one production codemap plan, and either inspects, compares, or applies it. They do not enter normal reconciliation, watch scheduling, or repository-demon execution. This exclusion is structural rather than a runtime feature flag.

The application currently supplies configured headings, marker prefix, and removal policy to the codemap planner. It does not yet supply the repository file-type schema provider required to create a missing schema-defined section, so current public execution adopts existing sections only.

## State ownership

The application layer holds command-scoped options and aggregate results. Durable repository identity and history belong to `internal/ddrepo` and link/reverse-index state packages. Runtime demon ownership belongs to `internal/demon`.

Command parsing must not become an alternative source of repository truth. Resolved configuration is passed into the owning subsystems.

## Invariants and safety boundaries

- Supplying any subsystem selector runs only selected systems.
- Without selectors, documentation indexes and links run; reverse indexes join when configured.
- Read-only commands do not apply authored-file reconciliation writes.
- Ambiguous subsystem diagnostics are not converted into guessed fixes by the application layer.
- Both executable names must expose compatible behavior unless a command is intentionally alias-specific.
- Bare `demon` and `demon --help` resolve to repository-demon help; `demon --version` remains the shared product version.
- Nested help must describe the requested subcommand rather than falling back to its parent summary.
- Help text and documentation must change with public command behavior.
- Production codemap execution remains explicit and cannot be enabled indirectly through normal reconciliation or daemon configuration.
- A missing codemap section remains unchanged until a file-type schema provider supplies an explicit placement.

## Code map

Primary files and packages:

- `cmd/ddocs/main.go` - canonical executable entry.
- `cmd/demon/main.go` - repository-demon alias entry and argument normalization.
- `cmd/demon/main_test.go` - alias help and version-routing coverage.
- `internal/app/app.go` - main command parsing and reconciliation orchestration.
- `internal/app/help_test.go` - top-level help and public command coverage.
- `internal/app/help_nested_test.go` - scoped nested help coverage.
- `internal/app/cli_contract_test.go` - CLI contract coverage.
- `internal/app/demon.go` - repository-demon and shell-hook command integration.
- `internal/app/move.go` - stateless refactoring command integration.
- `internal/app/orphans.go` - orphan-document health computation.
- `internal/app/review_*.go` - suggestion, change, undo, and repair-control command integration.
- `internal/app/codemap_execute.go` - explicit codemap fix/check/inspect parsing, configuration, and dispatch.
- `internal/app/codemap_execute_scope.go` - codemap root containment, Markdown file validation, traversal, ignore, symlink, and worktree exclusions.
- `internal/app/codemap_execute_output.go` - codemap summary and evidence-oriented inspection output.
- `internal/app/codemap_execute_test.go` - help, alias, required-root, dry-run, apply, and convergence coverage.
- `internal/app/codemap_benchmark.go` - benchmark command contract.
- `internal/app/codemap_precision.go` - precision command contract.
- `internal/repository/` - repository and worktree discovery used by the application.
- `internal/config/` - configuration resolution and defaults.

Important non-ownership boundaries:

- `internal/reconcile/` owns documentation index planning and application.
- `internal/links/` owns link inventory, evidence, and rewriting.
- `internal/reverseindex/` owns reverse-index planning and rendering.
- `internal/watch/` owns watcher scheduling.
- `internal/demon/` owns runtime leases and lifecycle.
- `internal/review/` owns review history, decision replay, undo construction, and repair controls.
- `internal/codemap/`, `internal/codemaprecommend/`, `internal/codemaprun/`, and `internal/evidence/` own codemap analysis and foreground execution.

## Tests

Relevant coverage includes:

- `internal/app/app_test.go`
- `internal/app/cli_contract_test.go`
- `internal/app/help_test.go`
- `internal/app/help_nested_test.go`
- `cmd/demon/main_test.go`
- `internal/app/demon_test.go`
- `internal/app/feature_flags_test.go`
- `internal/app/codemap_execute_test.go`
- `internal/app/codemap_export_test.go`
- `internal/app/codemap_benchmark_test.go`
- `internal/app/codemap_precision_test.go`
- `internal/app/move_test.go`
- `internal/app/orphans_test.go`
- `internal/app/orphans_integration_test.go`
- `internal/app/review_cli_test.go`

Run:

```bash
go test ./internal/app -count=1
```

## Related docs

- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Reconciliation Command Lifecycle](reconciliation-command-lifecycle.md)
- [Reconciliation Pipeline](reconciliation-pipeline.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Review Ledger](review-ledger.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Document Health Checks](../guides/document-health-checks.md)
- [Reverse Indexes](reverse-indexes.md)
- [Codemap Managed Execution](codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Repository Demon](../operations/repository-demon.md)

## Notes

The application package is allowed to coordinate several subsystems, but new mechanics should remain in their owning packages rather than accumulating in command handlers.
