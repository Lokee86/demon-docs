# CLI Reference

Parent index: [Reference](./README.md)

## Purpose

This document summarizes the public Demon Docs command surface, subsystem selectors, mutation behavior, and command ownership.

## Overview

`ddocs` is the canonical executable. `demon` is an installed alias backed by the same internal application implementation. Top-level and subcommand help remain the source of truth for every accepted flag:

```bash
ddocs --help
ddocs <command> --help
```

Commands either inspect state, plan without writing, apply deterministic repository changes, run foreground automation, or manage the repository demon lifecycle.

## Global commands

### `ddocs init --root PATH`

Initializes the current repository boundary and writes repository-local configuration. The documentation root must already exist inside the repository.

Mutation scope: repository-local Demon Docs configuration and state initialization.

### `ddocs status`

Displays selected repository root, documentation root, config path, and repository `.docignore` path.

Mutation scope: none.

### `ddocs check`

Computes reconciliation without writing authored repository files. It reports pending updates and unresolved conditions and returns non-zero when the selected systems are not clean.

Mutation scope: no authored-file writes. Internal read/cache behavior remains implementation-owned.

### `ddocs fix`

Computes and applies safe deterministic updates for selected systems, then persists the state needed for later reconciliation.

Mutation scope: managed documentation indexes, recognized repository Markdown link paths, configured reverse-index outputs, and private `.ddocs/` state.

### `ddocs watch`

Runs one reconciliation immediately, watches relevant filesystem paths, debounces events, and serializes subsequent reconciliation passes.

Mutation scope: the same selected authored surfaces as `fix`, plus watcher runtime activity.

## Subsystem selectors

```text
-d, --docs     documentation folder indexes and parent navigation
-l, --links    repository-local Markdown link inventory and reconciliation
-r, --reverse  code-folder reverse indexes
-i, --indexes  compatibility alias for --docs
```

When any selector is supplied, only selected systems run. Without selectors, documentation indexes and links run; reverse indexes also run when reverse roots are configured or supplied.

Selectors apply to `check`, `fix`, and `watch` where supported.

## Configuration commands

```bash
ddocs config paths
ddocs config show
ddocs config init --local
ddocs config init --global
```

`paths` reports configuration locations. `show` displays the resolved configuration. `init` writes a local or global standalone configuration template.

Repository-local `.ddocs/config.toml` remains the preferred initialized-repository configuration.

## Codemap commands

```bash
ddocs codemap export --output PATH
ddocs codemap benchmark ...
ddocs codemap precision ...
```

`export` writes a deterministic authored-codemap dataset. `benchmark` runs controlled holdouts. `precision` generates and evaluates ranked candidates against curated labels.

These commands do not silently modify authored codemap sections.

## Repository demon commands

The same lifecycle is available through `demon ...` and `ddocs demon ...`.

Primary operations include:

```bash
demon run
demon --status
demon --logs
demon acquire --client NAME
demon heartbeat --token TOKEN
demon release --token TOKEN
ddocs demon __shell-hook bash
ddocs demon __shell-hook powershell
```

The feeder commands are intended for shell and agent host adapters. A host acquires a token, refreshes it before expiry, and releases it on every completion path.

## Version and help

```bash
ddocs -v
ddocs --version
ddocs --help
demon --version
demon --help
```

## Defaults

```text
docs root:       docs
index file:      README.md
draft folder:    stubs
parent label:    Parent index
marker prefix:   doc-ledger
```

Configuration can override these conventions.

## Diagnostics and failure behavior

`check` returns non-zero for pending deterministic updates and unresolved selected-system conditions, including broken or ambiguous links and uninitialized link state where applicable.

`fix` does not guess among multiple plausible targets. Ambiguous sources remain unchanged and are reported.

Use [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md) for the behavioral contract and the command's own `--help` output for exact flag syntax.

## Examples

```bash
# Initialize and reconcile the default docs root.
ddocs init --root docs/
ddocs fix
ddocs check

# Verify only local links.
ddocs check --links

# Reconcile only documentation indexes.
ddocs fix --docs

# Run one watcher-path pass and exit.
ddocs watch --root docs --once

# Inspect resolved configuration.
ddocs config paths
ddocs config show
```

## Related docs

- [Getting Started](../guides/getting-started.md)
- [Configuration Reference](configuration.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Managed Files and State](managed-files-and-state.md)
- [Application Orchestration](../architecture/application-orchestration.md)
- [Repository Demon](../operations/repository-demon.md)

## Notes

This page intentionally summarizes command ownership rather than reproducing every generated help line. Command help must be updated with implementation changes and remains the exact invocation reference.
