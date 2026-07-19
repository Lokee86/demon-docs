# Application Orchestration

Parent index: [Architecture](./README.md)

## Purpose

This document describes the implemented application boundary that resolves configuration, selects subsystems, coordinates `check`, `fix`, and `watch`, and exposes codemap and repository-demon command groups.

## Overview

Both `ddocs` and `demon` enter the same internal application package. The executable wrappers supply process arguments and version metadata; `internal/app` owns command parsing and coordination across repository discovery, configuration, documentation reconciliation, local links, reverse indexes, codemap analysis, watch mode, and demon lifecycle.

The application layer coordinates ownership. It does not reimplement scanner, link, reverse-index, repository-state, or daemon mechanics.

## Code root

```text
internal/app/
cmd/ddocs/
cmd/demon/
```

## Responsibilities

The application boundary owns:

- top-level and subcommand dispatch;
- help and version behavior;
- repository and configuration selection;
- translation of CLI flags into subsystem options;
- `--docs`, `--links`, and `--reverse` selection semantics;
- ordering selected reconciliation systems;
- command output and aggregate success/failure decisions;
- foreground watch startup;
- codemap export, benchmark, and precision command integration;
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

## State ownership

The application layer holds command-scoped options and aggregate results. Durable repository identity and history belong to `internal/ddrepo` and link/reverse-index state packages. Runtime demon ownership belongs to `internal/demon`.

Command parsing must not become an alternative source of repository truth. Resolved configuration is passed into the owning subsystems.

## Invariants and safety boundaries

- Supplying any subsystem selector runs only selected systems.
- Without selectors, documentation indexes and links run; reverse indexes join when configured.
- Read-only commands do not apply authored-file reconciliation writes.
- Ambiguous subsystem diagnostics are not converted into guessed fixes by the application layer.
- Both executable names must expose compatible behavior unless a command is intentionally alias-specific.
- Help text and documentation must change with public command behavior.

## Code map

Primary files and packages:

- `cmd/ddocs/main.go` - canonical executable entry.
- `cmd/demon/main.go` - alias executable entry.
- `internal/app/app.go` - main command parsing and reconciliation orchestration.
- `internal/app/help_test.go` - help and public command coverage.
- `internal/app/cli_contract_test.go` - CLI contract coverage.
- `internal/app/demon.go` - repository-demon and shell-hook command integration.
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
- `internal/codemap*` and `internal/evidence/` own codemap analysis.

## Tests

Relevant coverage includes:

- `internal/app/app_test.go`
- `internal/app/cli_contract_test.go`
- `internal/app/help_test.go`
- `internal/app/demon_test.go`
- `internal/app/feature_flags_test.go`
- `internal/app/codemap_export_test.go`
- `internal/app/codemap_benchmark_test.go`
- `internal/app/codemap_precision_test.go`

Run:

```bash
go test ./internal/app -count=1
```

## Related docs

- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Reconciliation Pipeline](reconciliation-pipeline.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Reverse Indexes](reverse-indexes.md)
- [Repository Demon](../operations/repository-demon.md)

## Notes

The application package is allowed to coordinate several subsystems, but new mechanics should remain in their owning packages rather than accumulating in command handlers.
