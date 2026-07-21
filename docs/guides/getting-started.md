---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-73e8-b9e0-df6ed92de6e5
document_type: general
policy_exempt: false
summary: This guide installs Demon Docs, explains standalone and initialized operation, establishes deterministic index and link state, and reaches a clean ddocs check result.
---
# Getting Started

Parent index: [Guides](./INDEX.md)

## Purpose

This guide installs Demon Docs, explains when repository initialization is optional or required, establishes deterministic index and link state, and reaches a clean `ddocs check` result.

## Overview

Demon Docs has a static reconciliation core. `fix` applies deterministic scope-contained updates, while `check` verifies the same plan without writing. Core reconciliation and foreground `watch` work without `ddocs init`; the detached repository demon is an optional initialized-repository convenience.

## Prerequisites

- Go 1.26 or a compatible supported Go toolchain.
- A Git repository or ordinary project directory containing a documentation root.
- An existing documentation directory such as `docs/`.
- A clean or intentionally reviewed working tree before the first mutating pass.

## Install

From a checkout:

```bash
git clone https://github.com/Lokee86/demon-docs.git
cd demon-docs
go install ./cmd/ddocs
go install ./cmd/demon
```

Verify the commands:

```bash
ddocs --version
ddocs --help
demon --help
```

`ddocs` is the canonical command. `demon` is an alias backed by the same application implementation.

## Choose an operating mode

### Standalone reconciliation

Core index, link, health, move, and foreground-watch operations can run without initialization:

```bash
ddocs fix --root docs --indexes
ddocs fix --root docs --links
ddocs watch --root docs --once
ddocs check --root docs --indexes --links
```

Without a selected config, built-in defaults apply. The resolved docs root is also the standalone scope boundary, so repository-wide targets outside that root are not part of normal link or reverse-index scope. Link-enabled mutating passes create private state under `docs/.ddocs/`, but no `.ddocs/config.toml` is created.

### Initialized repository

Initialize from the repository root when the project needs a stable repository-wide boundary, repository-local configuration, starter schemas, persistent feature toggles, linked-worktree demon bootstrap, reverse projections outside the docs root, or the detached repository demon:

```bash
ddocs init --root docs/
```

The documentation root must already exist. Initialization writes `.ddocs/config.toml`, installs starter schemas, and makes the directory above `.ddocs/` the repository boundary.

Inspect initialized-repository paths with:

```bash
ddocs status
```

Inspect configuration selection in either mode with:

```bash
ddocs config paths
ddocs config show
```

## Review ignore rules

Create or update `.docignore` at the active scope root when generated, private, vendor, or scratch paths should be excluded. That is the docs root in standalone mode and the repository root in initialized mode.

Demon Docs always prunes `.git/`, `.ddocs/`, `.obsidian/`, and `logseq/`. Additional repository-specific exclusions belong in `.docignore`, not in global assumptions.

See [Configuration Reference](../reference/configuration.md) for syntax and precedence.

## Establish the initial state

Run a link-enabled mutating pass in the selected mode:

```bash
# Standalone
ddocs fix --root docs --links

# Initialized repository
ddocs fix
```

The first link-enabled pass establishes private link identity and history state. It does not guess historical moves that occurred before the baseline existed.

Review the resulting diff. Managed index blocks may be created or updated; authored prose outside managed blocks remains untouched.

Run a second pass and then check:

```bash
ddocs fix
ddocs check
```

The second `fix` should normally be idempotent. `check` should exit successfully when no pending deterministic changes or unresolved link conditions remain.

Once private state exists, unchanged clean frontmatter and document-format checks can reuse durable validation-cache records. Narrow non-link fixes refresh link metadata only for Markdown sources they actually changed; a clean index-, frontmatter-, or format-only fix does not run repository-wide link tracking or initialize missing link state.

## Select one subsystem

Use selectors when adopting one subsystem at a time:

```bash
ddocs check --indexes
ddocs check --frontmatter
ddocs check --format
ddocs check --links
ddocs check --reverse

ddocs fix --indexes
ddocs fix --frontmatter
ddocs fix --format
ddocs fix --links
ddocs fix --reverse
```

`--indexes` selects folder indexes only. `--docs` is the umbrella selector for indexes, frontmatter, and document-body format. Without selectors, configured documentation indexes, frontmatter, document-body format, and link tracking run. Link repair follows `[links].enabled`, and reverse indexes also run when roots are configured or supplied.

## Expected result

A successful adoption leaves:

- a consistent standalone scope or initialized repository boundary;
- documentation folder indexes in deterministic managed blocks;
- a private local Markdown link-state baseline;
- no unresolved or ambiguous links requiring user decisions;
- a clean second `fix`; and
- a successful `ddocs check`.

## Failure and recovery

### The documentation root does not exist

Create or select the intended docs root before running reconciliation or `init`. Demon Docs does not invent the product's documentation taxonomy.

### The first link pass reports issues but does not repair moves

This is expected when no prior identity baseline exists. Resolve current broken links manually, run `fix` to establish the baseline, and use later passes for deterministic move repair.

### A link has multiple plausible targets

Demon Docs leaves the source unchanged. Choose the intended target manually, then rerun `fix` and `check`.

### Generated changes are broader than expected

Stop and inspect configuration selection, `docs_root`, include/exclude patterns, `.docignore`, and subsystem selectors before accepting the diff.

### Runtime state appears stale

Stop foreground watchers and, for initialized repositories, the repository demon. Then use the recovery guidance in [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md). Private `.ddocs/` state is rebuildable, but deleting it discards link history and should be a deliberate last resort.

## Related docs

- [Product Walkthrough](product-walkthrough.md)
- [CLI Reference](../reference/cli.md)
- [Using Document Schemas](document-schemas.md)
- [Configuration Reference](../reference/configuration.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [CI and Automation](ci-and-automation.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Run the first adoption on a branch or clean working tree. Start standalone when docs-scoped defaults are sufficient; initialize only when repository-level configuration or daemon ownership is useful. Deterministic behavior makes review repeatable, but it does not replace reviewing repository changes.
