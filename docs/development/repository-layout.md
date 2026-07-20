---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7874-9b0a-5ab2f1373ce8
document_type: general
policy_exempt: false
summary: This document maps the Demon Docs repository into command entry points, implementation packages, tests, research artifacts, and generated or runtime boundaries.
---
# Repository Layout

Parent index: [Development](./INDEX.md)

## Purpose

This document maps the Demon Docs repository into command entry points, implementation packages, tests, research artifacts, and generated or runtime boundaries.

## Overview

Demon Docs is a Go CLI application with focused internal packages. The repository keeps production code under `cmd/` and `internal/`, black-box and fixture regression tests under `tests/`, durable documentation under `docs/`, and evaluation artifacts under `research/`.

Package boundaries should preserve direct ownership. New mechanics belong in the narrowest durable package rather than accumulating in `internal/app` or being hidden behind unnecessary wrappers.

## Top-level layout

```text
cmd/          executable entry points
internal/     production implementation packages
tests/        repository-level regression tests
docs/         canonical documentation
research/     retained benchmark corpora, scripts, labels, and reports
.github/      CI workflows
Makefile      local development and release gates
bin/          local build outputs; not committed
```

## Command entry points

```text
cmd/ddocs/   canonical CLI executable
cmd/demon/   alias executable using the same application implementation
```

Executable packages should remain thin. Command mechanics belong in `internal/app`; subsystem mechanics belong in their owning packages.

## Core implementation packages

```text
internal/app/             CLI coordination and command groups
internal/config/          configuration defaults, selection, and behavior
internal/repository/      repository scope, discovery, and worktrees
internal/frontmatter/     frontmatter parsing, repair, and immutable-value state
internal/documentpolicy/  document schemas, body-format enforcement, and schema history
internal/scan/            documentation-tree inventory
internal/markdown/        managed Markdown parsing and source-preserving edits
internal/reconcile/       forward documentation index planning/application
internal/filetxn/         shared content-addressed file replacement and rollback
internal/links/           local link inventory, evidence, state, and rewrites
internal/ddrepo/          private object repository and transactions
internal/reverseindex/    code-folder reverse-index projection
internal/watch/           filesystem watcher filtering and scheduling
internal/demon/           repository-local owner, feeders, runtime, and logs
internal/review/          suggestion decisions, applied history, undo, and blocks
internal/codemap/         authored codemap extraction and datasets
internal/evidence/        deterministic missing-link evidence collection
internal/codemapbench/    holdout, ranking, tiering, and reports
internal/codemapcorpus/   repository fact collection for codemap analysis
internal/codemapprecision/ curated precision evaluation
internal/codemaprecommend/ production evidence admission, ranking, and tiers
internal/codemaprun/       explicit managed codemap planning and publication
```

Shared utility packages such as `internal/pathutil`, `internal/textio`, `internal/model`, and `internal/ignore` should remain small and concrete.

## Test layout

Most package tests live beside implementation. Repository-level fixture and CLI regression tests live under:

```text
tests/
```

Focused behavioral tests protect source preservation, deterministic ordering, configuration precedence, worktree boundaries, daemon ownership, link syntax, stateless moves, orphan health rules, review-history safety, concurrency, codemap evidence, and benchmark reporting.

## Research layout

`research/` retains evidence that should not be mixed with production state:

```text
context-benchmarking/
codemap-inventory/
codemap-review/
codemap-evidence-validation/
codemap-precision/
codemap-audit/
cross-repo-codemap-benchmark/
mass-rename-results/
mass-rename-timing/
link-performance/
```

Research scripts and outputs must identify whether they are reproducible inputs, generated reports, curated labels, or historical evidence.

## Documentation layout

```text
docs/agent/         agent workflow, orientation, testing, and editing guardrails
docs/guides/        task-oriented user workflows
docs/reference/     exact public contracts
docs/architecture/  implemented ownership and behavior
docs/operations/    runtime operation and recovery
docs/research/      narrative research evidence and interpretation
docs/planning/      future and unresolved work
docs/development/   contributor workflow and repository maintenance
```

The repository root README remains the product entry point, not the complete manual.

## Generated and runtime boundaries

Do not commit normal local outputs:

```text
bin/
ddocs
demon
ddocs.exe
demon.exe
.cache/
dummy-docs/
```

`.ddocs/` is repository-local private state and runtime ownership. Its exact commit policy depends on the repository integration, but runtime files and logs are not authored documentation.

Nested `.worktrees/` directories must be excluded from repository-wide scans, tests, formatters, and documentation traversal to avoid recursively processing attached worktrees.

## Failure modes

Repository-wide commands can become misleading when they recurse into `.worktrees/`, generated fixtures, private `.ddocs/` state, or benchmark corpora not intended for normal tests. Use package/test commands and configured exclusions that preserve the intended scope.

Do not treat research reports as production fixtures unless a test explicitly pins them.

## Code map

- `cmd/ddocs/main.go`
- `cmd/demon/main.go`
- `internal/app/`
- `internal/config/`
- `internal/repository/`
- `internal/scan/`
- `internal/markdown/`
- `internal/reconcile/`
- `internal/links/`
- `internal/ddrepo/`
- `internal/reverseindex/`
- `internal/watch/`
- `internal/demon/`
- `internal/review/`
- `internal/codemap/`
- `internal/evidence/`
- `internal/codemapbench/`
- `internal/codemapcorpus/`
- `internal/codemapprecision/`
- `tests/`
- `research/`

## Related docs

- [Documentation Policy](../documentation-policy.md)
- [Application Orchestration](../architecture/application-orchestration.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Review Ledger](../architecture/review-ledger.md)
- [Testing and Fixtures](testing-and-fixtures.md)
- [Roadmap](../planning/roadmap.md)

## Notes

The package list documents current ownership, not a prohibition on adding seams. New packages are appropriate when they establish a concrete durable owner rather than a vague abstraction layer.
