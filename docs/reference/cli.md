---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-713c-b3c5-e35edf5c86f6
document_type: general
policy_exempt: false
summary: This document summarizes the public Demon Docs command surface, subsystem selectors, mutation behavior, and command ownership.
---
# CLI Reference

Parent index: [Reference](./INDEX.md)

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

Help is scoped to the requested command. For example, `ddocs suggestions select --help` describes candidate selection rather than repeating the parent suggestions summary, and `ddocs codemaps precision sample --help` lists the required report input and sampling flags.

Commands either inspect state, plan without writing, apply deterministic repository changes, run foreground automation, or manage the repository demon lifecycle.

## Standalone and initialized modes

`ddocs init` is not a prerequisite for ordinary reconciliation.

Without an initialized repository, `check`, `fix`, and foreground `watch` use the selected config or built-in defaults. The resolved docs root becomes both the managed docs root and the standalone scope boundary. Link-enabled mutating passes may create private state under `<docs-root>/.ddocs/`; they do not create `.ddocs/config.toml` or make `ddocs status` and the detached demon available.

`ddocs mv` is independently stateless. It uses the detected initialized repository root when available, otherwise the current directory or explicit `--root`, and never creates or updates `.ddocs/`.

Initialize when a stable repository-wide boundary, repository-local configuration, feature toggles, starter schemas, linked-worktree demon bootstrap, reverse projections outside the docs root, or detached demon ownership is required.

## Global commands

### `ddocs init --root PATH`

Optionally establishes an initialized repository boundary and writes repository-local configuration. The documentation root must already exist inside the repository. Core `check`, `fix`, `watch`, and `mv` behavior does not require this command.

Mutation scope: repository-local Demon Docs configuration and state initialization.

### `ddocs status`

Displays the initialized repository root, documentation root, config path, and repository `.docignore` path. It fails when no initialized repository is discoverable.

Mutation scope: none.

### `ddocs mv [--root PATH] [--dry-run] SOURCE DESTINATION`

Moves one repository-contained file or directory and rewrites affected incoming links and relative links inside moved Markdown sources. It does not require, create, or update `.ddocs/` state. There is no separate `ddocs rename` command; use `ddocs mv SOURCE DESTINATION` for both moves and renames.

Mutation scope: the requested filesystem source and affected repository Markdown files inside the selected boundary. `--dry-run` is read-only.

### `ddocs new [--force] DOCUMENT_TYPE PATH`

Creates a Markdown document from `.ddocs/schemas/<document-type>.toml`. The schema name becomes `document_type`; configured frontmatter and Markdown sections are created together. Existing files require interactive confirmation or `--force`.

Mutation scope: the requested new document only.

### `ddocs format ignore|merge|delete ...`

Resolves one document-body format conflict explicitly. `ignore` creates or updates the document-specific TOML schema, `merge` combines duplicate sibling sections, and `delete` removes one selected occurrence.

Mutation scope: the selected document and, for `ignore`, `.ddocs/document-schemas/<document-id>.toml`.

### `ddocs schema init [--force]`

Writes the provided Space Rocks-derived starter TOML schemas. Existing schema files are preserved unless `--force` is supplied.

Mutation scope: the configured shared schema directory.

### `ddocs check`

Computes reconciliation without writing selected authored surfaces or frontmatter/document-format state. It reports pending updates and unresolved conditions and returns non-zero when the selected systems are not clean. When link maintenance is disabled but link tracking is selected, it may persist the tracking baseline; enabled link checks do not publish a replacement baseline. When links are selected, it also reports managed Markdown documents with no meaningful inbound link.

Mutation scope: no authored-file writes. Internal read/cache behavior remains implementation-owned.

### `ddocs fix`

Computes and applies safe deterministic updates for selected systems, then persists the state needed for later reconciliation.

Link repair runs first. After frontmatter, document format, reverse indexes, and documentation indexes apply, link state is refreshed only for source paths that changed. A clean frontmatter-only, format-only, or index-only fix does not run repository-wide link tracking. Explicit `--links` retains the full link reconciliation path, including its review history, rollback, and watcher-suppression behavior.

Mutation scope: managed documentation indexes, configured Markdown frontmatter and body structure beneath the docs root, recognized repository Markdown link paths, configured reverse-index outputs, and private `.ddocs/` state.

### `ddocs watch`

Runs one reconciliation immediately, watches relevant filesystem paths, debounces events, and serializes subsequent reconciliation passes. Foreground watch works in standalone and initialized modes.

Mutation scope: the same selected authored surfaces as `fix`, plus watcher runtime activity.

## Subsystem selectors

```text
-d, --docs     documentation indexes, configured frontmatter, and document-body format
    --frontmatter
               configured frontmatter only
    --format   document-body format only
-l, --links    repository-local Markdown link inventory and reconciliation
-r, --reverse  code-folder reverse indexes
-i, --indexes  reconcile documentation indexes only
```

When any selector is supplied, only selected systems run. Without selectors, documentation indexes, configured frontmatter, document-body format, and links run; reverse indexes also run when reverse roots are configured or supplied.

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

These commands inspect current ambiguous link repairs, join them with persisted decisions, and convert a selected candidate into the normal hash-guarded repair path. They do not generate codemap recommendations; those are owned by the explicit `ddocs codemaps` command family. Declines persist by stable relationship and evidence fingerprint.

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

Repository-local `.ddocs/config.toml` is the initialized-repository configuration. It is optional when built-in defaults or a standalone local/global config are sufficient.

## Codemap commands

### Production execution

```bash
ddocs codemaps fix [--root FILE_OR_DIRECTORY] [--dry-run]
ddocs codemaps check --root FILE_OR_DIRECTORY
ddocs codemaps inspect --root FILE_OR_DIRECTORY
```

`ddocs codemap ...` remains accepted as a singular compatibility alias. Help and current examples render the canonical plural `ddocs codemaps ...` form.

Common execution flags are:

```text
--root PATH        existing Markdown file or directory beneath the configured docs root
--config PATH      explicit configuration file
--no-local-config  skip local configuration discovery
--no-global-config skip global configuration discovery
--heading TEXT     replace configured codemap headings; repeatable
--dry-run          fix only; report the plan without writing
```

A file root must already exist and have the `.md` extension. A directory root selects regular `.md` files recursively, honors `.docignore`, skips symbolic-link entries, and prunes `.worktrees/` and `.workingtrees/`. Every explicit root must remain beneath the configured documentation root.

`codemap fix` may omit `--root`; it then targets the configured documentation root. It adopts the complete matching section as one managed unit, filters deterministic recommendations through shared decline policy, preserves existing valid links by default, applies configured removals, and publishes changed files through the shared content-addressed transaction layer.

`--dry-run` builds the same plan and prints:

```text
ddocs codemaps fix would update N file(s)
PATH: added=N removed=N adopted=true|false created=true|false
```

Dry-run performs no document or review-state writes.

A successful mutating run prints the corresponding `updated N file(s)` summary. A clean plan reports zero updated files.

`codemap check` requires `--root`. It builds the production plan without writing. Exit behavior is:

```text
0  no selected document would change
1  one or more selected documents would change
2  command-line usage failure
non-zero configuration, scope, read, extraction, or planning failure
```

When stale, it prints `ddocs codemaps check failed` followed by changed document paths. When clean, it prints `ddocs codemaps check passed`.

`codemap inspect` requires `--root` and writes nothing. For every selected document it reports:

```text
section: missing | existing | schema-created
changed: true | false
add TARGET score=SCORE tier=TIER
declined TARGET score=SCORE tier=TIER
evidence lines
remove TARGET
```

Configured heading matching is case-insensitive and ignores heading-like lines inside fenced code blocks. Multiple matching sections are an error. Malformed or duplicated codemap ownership markers are an error.

Existing configured sections are processed regardless of document schema. For a missing section, the application selects the effective schema from `document_type` metadata, configured path rules, and any document-specific exception. A required codemap section is created at its schema-defined position; schemas without one report `missing` and leave the document unchanged.

The complete section body is adopted between codemap-specific markers derived from `[markers].prefix`. Existing and newly generated links are not split into provenance groups. Fenced path lists remain fenced; bullet maps retain their first recognized bullet prefix. All selected non-declined `hard_link` and `context` recommendations are eligible for addition.

Confidence-based removal is disabled by default and controlled by `[codemap].remove_undiscovered_links` and `[codemap].remove_low_score_links`. Declines suppress proposed additions only; they do not remove existing links.

Codemap execution is intentionally excluded from generic `ddocs fix`, `ddocs check`, `ddocs watch`, and every repository-demon path.

See [Managing Codemaps](../guides/managing-codemaps.md) for the task workflow and [Codemap Managed Execution](../architecture/codemap-managed-execution.md) for the complete ownership, planning, rendering, transaction, and failure lifecycle.

### Dataset and evaluation commands

```bash
ddocs codemaps export --output PATH
ddocs codemaps benchmark ...
ddocs codemaps precision source ...
ddocs codemaps precision sample ...
ddocs codemaps precision evaluate ...
```

`export` writes a deterministic authored-codemap dataset. `benchmark` runs controlled holdouts. `precision source` generates current recommendations without hiding authored links, `precision sample` creates a deterministic unlabeled review set, and `precision evaluate` compares a labeled benchmark with its deterministic recommendation report.

The evaluation commands and production CLI consume the same `internal/codemaprecommend` ranking implementation.

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
index file:      INDEX.md
draft folder:    stubs
parent label:    Parent index
marker prefix:   doc-ledger
```

Configuration can override these conventions.

## Diagnostics and failure behavior

`check` returns non-zero for pending deterministic updates and unresolved selected-system conditions, including frontmatter violations, broken or ambiguous links, a missing link-state baseline, and orphan managed Markdown documents when links are selected.

`fix` does not guess among multiple plausible targets. Ambiguous sources remain unchanged and are exposed through `ddocs suggestions`.

`mv` refuses paths outside its selected boundary, affected ambiguous wiki targets, source-content changes after planning, symbolic-link sources, and existing non-directory destinations.

Undo refuses to overwrite a file whose current content no longer matches the recorded after hash.

Use [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md) for the behavioral contract and the command's own scoped `--help` output for exact flag syntax, required identifiers, default values, mutation guards, and output behavior.

## Examples

```bash
# Reconcile the default docs root without repository initialization.
ddocs fix --root docs --docs
ddocs fix --root docs --links
ddocs check --root docs --docs --links

# Optionally initialize a repository-wide configuration and state boundary.
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

# Reconcile documentation indexes, frontmatter, and body format.
ddocs fix --docs

# Run the policy operations independently.
ddocs check --frontmatter
ddocs fix --format

# Create and explicitly resolve a document.
ddocs new service docs/services/new-service.md
ddocs format ignore --heading "Appendix" docs/guide.md

# Run one watcher-path pass and exit.
ddocs watch --root docs --once

# Inspect resolved configuration.
ddocs config paths
ddocs config show
```

## Related docs

- [Getting Started](../guides/getting-started.md)
- [Using Document Schemas](../guides/document-schemas.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Document Health Checks](../guides/document-health-checks.md)
- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Adopting Reverse Indexes](../guides/reverse-indexes.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)
- [Supported Link Syntax](supported-link-syntax.md)
- [Configuration Reference](configuration.md)
- [Document Schemas And Format Enforcement](document-schemas.md)
- [Compatibility and Migrations](compatibility-and-migrations.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Managed Files and State](managed-files-and-state.md)
- [Application Orchestration](../architecture/application-orchestration.md)
- [Repository Demon](../operations/repository-demon.md)

## Notes

This page intentionally summarizes command ownership rather than reproducing every generated help line. Command help must be updated with implementation changes and remains the exact invocation reference.
