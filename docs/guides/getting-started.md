---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-73e8-b9e0-df6ed92de6e5
document_type: general
policy_exempt: false
summary: This guide installs Demon Docs, initializes an existing repository, establishes deterministic index and link state, and reaches a clean ddocs check result.
---
# Getting Started

Parent index: [Guides](./INDEX.md)

## Purpose

This guide installs Demon Docs, initializes an existing repository, establishes deterministic index and link state, and reaches a clean `ddocs check` result.

## Overview

Demon Docs has a static reconciliation core. `fix` applies deterministic repository-contained updates, while `check` verifies the same plan without writing. Watch and daemon automation are optional conveniences layered over those commands.

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

## Initialize a repository

From the repository root:

```bash
ddocs init --root docs/
```

The documentation root must already exist. Initialization writes repository-local configuration under `.ddocs/` and records the repository and documentation boundaries.

Inspect the selected paths:

```bash
ddocs status
ddocs config paths
ddocs config show
```

## Review ignore rules

Create or update `.docignore` at the repository root when generated, private, vendor, or scratch paths should be excluded.

Demon Docs always prunes `.git/`, `.ddocs/`, `.obsidian/`, and `logseq/`. Additional repository-specific exclusions belong in `.docignore`, not in global assumptions.

See [Configuration Reference](../reference/configuration.md) for syntax and precedence.

## Establish the initial state

Run:

```bash
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

## Select one subsystem

Use selectors when adopting one subsystem at a time:

```bash
ddocs check --docs
ddocs check --links
ddocs check --reverse

ddocs fix --docs
ddocs fix --links
ddocs fix --reverse
```

Without selectors, configured documentation indexes, frontmatter, document-body format, and link tracking run. Link repair follows `[links].enabled`, and reverse indexes also run when roots are configured or supplied.

## Expected result

A successful adoption leaves:

- repository-local configuration selected consistently;
- documentation folder indexes in deterministic managed blocks;
- local Markdown link state initialized;
- no unresolved or ambiguous links requiring user decisions;
- a clean second `fix`; and
- a successful `ddocs check`.

## Failure and recovery

### The documentation root does not exist

Create or select the intended root before running `init`. Demon Docs does not invent the product's documentation taxonomy.

### The first link pass reports issues but does not repair moves

This is expected when no prior identity baseline exists. Resolve current broken links manually, run `fix` to establish the baseline, and use later passes for deterministic move repair.

### A link has multiple plausible targets

Demon Docs leaves the source unchanged. Choose the intended target manually, then rerun `fix` and `check`.

### Generated changes are broader than expected

Stop and inspect configuration selection, `docs_root`, include/exclude patterns, `.docignore`, and subsystem selectors before accepting the diff.

### Runtime state appears stale

Stop foreground watchers or the repository demon, then use the recovery guidance in [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md). The private `.ddocs/` state is rebuildable, but deleting it discards link history and should be a deliberate last resort.

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

Run the first adoption on a branch or clean working tree. Deterministic behavior makes review repeatable, but it does not replace reviewing repository changes.
