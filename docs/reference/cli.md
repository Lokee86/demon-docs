# CLI Reference

Parent index: [Reference](./README.md)

## Purpose

This document summarizes the public Demon Docs command surface, subsystem selectors, mutation behavior, and command ownership.

## Overview

`ddocs` is the canonical executable. `demon` is an installed alias for the repository-demon command family. Top-level, command, and nested-subcommand help remain the source of truth for every accepted flag:

```bash
ddocs --help
ddocs <command> --help
ddocs <command> <subcommand> --help
demon --help
demon <command> --help
```

Help is scoped to the requested command. For example, `ddocs suggestions select --help` describes candidate selection rather than repeating the parent suggestions summary, and `ddocs codemap precision sample --help` lists the required report input and sampling flags.

Commands either inspect state, plan without writing, apply deterministic repository changes, run foreground automation, or manage the repository demon lifecycle.

## Global commands

### `ddocs init --root PATH`

Initializes the current repository boundary and writes repository-local configuration. The documentation root must already exist inside the repository.

Mutation scope: repository-local Demon Docs configuration and state initialization.

### `ddocs status`

Displays selected repository root, documentation root, config path, and repository `.docignore` path.

Mutation scope: none.

### `ddocs mv [--root PATH] [--dry-run] SOURCE DESTINATION`

Moves one repository-contained file or directory and rewrites affected incoming links and relative links inside moved Markdown sources. It does not require, create, or update `.ddocs/` state.

Mutation scope: the requested filesystem source and affected repository Markdown files inside the selected boundary. `--dry-run` is read-only.

### `ddocs check`

Computes reconciliation without writing authored repository files. It reports pending updates and unresolved conditions and returns non-zero when the selected systems are not clean. When links are selected, it also reports managed Markdown documents with no meaningful inbound link.

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

## Suggestion commands

```bash
ddocs suggestions [FILE]
ddocs suggestions declined [FILE]
ddocs suggestions log [FILE]
ddocs suggestions show SUGGESTION
ddocs suggestions select SUGGESTION [CANDIDATE]
ddocs suggestions decline SUGGESTION [CANDIDATE] --reason "..."
ddocs suggestions reconsider SUGGESTION
```

These commands inspect current ambiguous link repairs and codemap missing-link candidates, join them with persisted decisions, and convert a selected candidate into the normal hash-guarded repair path. Declines persist by stable relationship and evidence fingerprint.

## Applied-change commands

```bash
ddocs changes [FILE]
ddocs changes related FILE
ddocs changes show CHANGE
ddocs changes log [FILE]
ddocs changes undo CHANGE [--repair REPAIR] [--block] [--reason "..."]
ddocs changes undo-run RUN [--block] [--reason "..."]
ddocs changes block CHANGE [--repair REPAIR] [--reason "..."]
ddocs changes unblock CHANGE [--repair REPAIR]
```

These commands inspect the private applied-change ledger, perform bounded hash-guarded undo, and control exact repair fingerprints. They do not perform arbitrary historical selective reverts through later user edits.

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
ddocs codemap precision source ...
ddocs codemap precision sample ...
ddocs codemap precision evaluate ...
```

`export` writes a deterministic authored-codemap dataset. `benchmark` runs controlled holdouts. `precision source` generates current suggestions without hiding authored links, `precision sample` creates a deterministic unlabeled review set, and `precision evaluate` compares a labeled benchmark with its deterministic suggestion report. The legacy flag-only precision form remains equivalent to `evaluate`.

These commands do not silently modify authored codemap sections.

## Repository demon commands

The same lifecycle is available through `demon ...` and `ddocs demon ...`. Running `demon` with no arguments or `demon --help` opens the repository-demon help page; `demon --version` reports the shared Demon Docs version.

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

`check` returns non-zero for pending deterministic updates and unresolved selected-system conditions, including broken or ambiguous links, uninitialized link state, and orphan managed Markdown documents when links are selected.

`fix` does not guess among multiple plausible targets. Ambiguous sources remain unchanged and are exposed through `ddocs suggestions`.

`mv` refuses paths outside its selected boundary, affected ambiguous wiki targets, source-content changes after planning, symbolic-link sources, and existing non-directory destinations.

Undo refuses to overwrite a file whose current content no longer matches the recorded after hash.

Use [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md) for the behavioral contract and the command's own scoped `--help` output for exact flag syntax, required identifiers, default values, mutation guards, and output behavior.

## Examples

```bash
# Initialize and reconcile the default docs root.
ddocs init --root docs/
ddocs fix
ddocs check

# Preview and apply an explicit link-aware move.
ddocs mv --dry-run docs/old.md docs/new.md
ddocs mv docs/old.md docs/new.md

# Verify links and orphan-document health.
ddocs check --links

# Review unresolved suggestions and recorded changes.
ddocs suggestions
ddocs changes

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
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Document Health Checks](../guides/document-health-checks.md)
- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Adopting Reverse Indexes](../guides/reverse-indexes.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)
- [Supported Link Syntax](supported-link-syntax.md)
- [Configuration Reference](configuration.md)
- [Compatibility and Migrations](compatibility-and-migrations.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Managed Files and State](managed-files-and-state.md)
- [Application Orchestration](../architecture/application-orchestration.md)
- [Repository Demon](../operations/repository-demon.md)

## Notes

This page intentionally summarizes command ownership rather than reproducing every generated help line. Command help must be updated with implementation changes and remains the exact invocation reference.
