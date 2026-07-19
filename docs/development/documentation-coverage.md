# Documentation Coverage Map

Parent index: [Development](./README.md)

## Purpose

This document maps every current production package and public command family to canonical documentation owners so implemented behavior does not exist only in code, tests, the root README, research notes, or the roadmap.

## Overview

Coverage is ownership-based rather than one-file-per-package. A focused utility package may be documented inside the architecture page for the subsystem it serves. A major independent boundary requires its own current architecture or operations owner.

The map covers current production code under `cmd/` and `internal/`. Planning packages do not exist in code and therefore are not counted as current implementation coverage.

## Coverage rules

A production boundary is covered when:

- one current guide, reference, architecture, operations, or development document owns its behavior;
- the document identifies important non-ownership boundaries;
- public commands and mutation scope are represented in reference documentation;
- implementation-facing pages provide useful code maps and tests;
- known incomplete surfaces appear in limits or planning; and
- research pages are not the sole authority for shipped behavior.

The root README and roadmap are entry and status documents. They do not satisfy implementation coverage by themselves.

## Executable and application coverage

| Code boundary | Responsibility | Canonical current docs |
| --- | --- | --- |
| `cmd/ddocs/` | Canonical executable entry | [Application Orchestration](../architecture/application-orchestration.md), [CLI Reference](../reference/cli.md) |
| `cmd/demon/` | Alias executable entry | [Application Orchestration](../architecture/application-orchestration.md), [Repository Demon](../operations/repository-demon.md) |
| `internal/app/` | Command parsing, selection, orchestration, output, aggregate result | [Application Orchestration](../architecture/application-orchestration.md), [CLI Reference](../reference/cli.md), command-specific guides |
| `internal/app/move.go` | Stateless move command integration | [Stateless Document Refactoring](../guides/document-refactoring.md), [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md) |
| `internal/app/orphans.go` | Orphan health projection | [Document Health Checks](../guides/document-health-checks.md), [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md) |
| `internal/app/review_*.go` | Suggestion, change, undo, and block commands | [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md), [Review Ledger](../architecture/review-ledger.md) |
| `internal/app/codemap_*.go` | Codemap export, benchmark, and precision commands | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |
| `internal/app/demon*.go` | Demon commands, shell hooks, feeder integration, detached startup | [Repository Demon](../operations/repository-demon.md), [Host Adapter Feeder Integration](../operations/host-adapters.md) |
| `internal/app/reverse_index.go` | Reverse option resolution and mixed watch coordination | [Reverse Index Architecture](../architecture/reverse-indexes.md), [Adopting Reverse Indexes](../guides/reverse-indexes.md) |

## Repository and configuration coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/config/` | Defaults, TOML decoding, compatibility keys, selection, demon toggle mutation | [Configuration Reference](../reference/configuration.md), [Compatibility and Migrations](../reference/compatibility-and-migrations.md) |
| `internal/repository/` | Repository discovery, scope, containment, linked-worktree bootstrap | [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md), [Using Linked Git Worktrees](../guides/linked-worktrees.md) |
| `internal/ignore/` | `.docignore` policies, nested domains, permanent exclusions | [Ignore and Traversal](../architecture/ignore-and-traversal.md), [Configuration Reference](../reference/configuration.md) |
| `internal/ddrepo/` | Private Git-object repository, codecs, transactions | [Repository State and Transactions](../architecture/repository-state-and-transactions.md), [Managed Files and State](../reference/managed-files-and-state.md) |

## Documentation reconciliation coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/scan/` | Documentation-tree inventory and scope | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md), [Ignore and Traversal](../architecture/ignore-and-traversal.md) |
| `internal/markdown/` | Managed sections, headings, parent links, source-preserving text changes | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md), [Managed Files and State](../reference/managed-files-and-state.md), [Compatibility and Migrations](../reference/compatibility-and-migrations.md) |
| `internal/reconcile/` | Forward index planning and application | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md), [Getting Started](../guides/getting-started.md) |
| `internal/model/` | Shared file-update and reconciliation structures | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md) |
| `internal/pathutil/` | Relative path rendering used by generated documentation links | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md), [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) |
| `internal/textio/` | Newline-aware text reads used by generated edits | [Managed Files and State](../reference/managed-files-and-state.md), [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md) |

## Link and review coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/links/` | Syntax parsing, inventory, identities, evidence, repair, moves, application | [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md), [Supported Link Syntax](../reference/supported-link-syntax.md), [Stateless Document Refactoring](../guides/document-refactoring.md) |
| `internal/review/` | Append-only review events, fingerprints, policy replay, undo | [Review Ledger](../architecture/review-ledger.md), [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md) |

## Reverse-index coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/reverseindex/` | Root scope, traversal, codemap projection, rendering, apply, watch | [Reverse Index Architecture](../architecture/reverse-indexes.md), [Adopting Reverse Indexes](../guides/reverse-indexes.md) |

## Watcher and daemon coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/watch/` | Event scope, filters, debounce, serialized scheduling | [Watcher and Automation](../operations/watcher-and-automation.md) |
| `internal/demon/` | Owner lease, feeders, runtime files, logs, shutdown grace | [Repository Demon](../operations/repository-demon.md), [Host Adapter Feeder Integration](../operations/host-adapters.md) |

## Codemap and evidence coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/codemap/` | Codemap extraction, normalized entries, datasets, selected insertion | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |
| `internal/codemapcorpus/` | Repository files, symbols, dependencies, history, and related-fact corpus | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |
| `internal/evidence/` | Deterministic structural, mention, history, and symbol evidence | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |
| `internal/codemapbench/` | Holdouts, current suggestions, ranking, tiers, reports | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |
| `internal/codemapprecision/` | Curated-label evaluation and aggregate metrics | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |

## Public command coverage

| Public surface | Task guide | Exact/current owner |
| --- | --- | --- |
| Install, initialize, first baseline | [Getting Started](../guides/getting-started.md) | [CLI Reference](../reference/cli.md), [Configuration Reference](../reference/configuration.md) |
| `status`, `config paths`, `config show` | [Getting Started](../guides/getting-started.md) | [Configuration Reference](../reference/configuration.md) |
| `check`, `fix`, selectors | [Getting Started](../guides/getting-started.md), [CI and Automation](../guides/ci-and-automation.md) | [CLI Reference](../reference/cli.md), [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md) |
| Managed folder indexes and parent links | [Getting Started](../guides/getting-started.md) | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md), [Managed Files and State](../reference/managed-files-and-state.md) |
| Local links | [Getting Started](../guides/getting-started.md) | [Supported Link Syntax](../reference/supported-link-syntax.md), [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) |
| Orphan health | [Document Health Checks](../guides/document-health-checks.md) | [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md) |
| `mv` | [Stateless Document Refactoring](../guides/document-refactoring.md) | [CLI Reference](../reference/cli.md), [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) |
| Reverse indexes | [Adopting Reverse Indexes](../guides/reverse-indexes.md) | [Reverse Index Architecture](../architecture/reverse-indexes.md) |
| Suggestions and changes | [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md) | [Review Ledger](../architecture/review-ledger.md) |
| Codemap export/benchmark/precision | [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md) | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md), [CLI Reference](../reference/cli.md) |
| Foreground `watch` | [CI and Automation](../guides/ci-and-automation.md) | [Watcher and Automation](../operations/watcher-and-automation.md) |
| Repository demon | [CI and Automation](../guides/ci-and-automation.md) | [Repository Demon](../operations/repository-demon.md) |
| External feeder adapters | Host-specific integration | [Host Adapter Feeder Integration](../operations/host-adapters.md) |
| Linked worktrees | [Using Linked Git Worktrees](../guides/linked-worktrees.md) | [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md) |
| Upgrade and migration | [Upgrading Demon Docs](../guides/upgrading.md) | [Compatibility and Migrations](../reference/compatibility-and-migrations.md) |
| Current incomplete surfaces | Not applicable | [Current Product Limitations](../limits/current-limitations.md) |

## Research and planning separation

Current architecture is owned by the package mappings above. These pages remain evidence or future direction only:

```text
docs/research/       benchmark evidence and methodology
docs/planning/       unresolved or future product work
docs/limits/         current incomplete user-visible surfaces
```

A research result must update an architecture, reference, limits, or planning owner when it changes product decisions. A shipped plan must graduate its current facts into the package owners listed here.

## Verification workflow

When production packages or public commands change:

1. list current `cmd/` and `internal/` directories;
2. map the changed owner to this document;
3. update the canonical current document, not only the roadmap;
4. add or update a task guide when normal user workflow changes;
5. update exact reference for flags, configuration, formats, state, or diagnostics;
6. update limits when an incomplete surface changes;
7. run documentation reconciliation and link checks; and
8. run the normal Go test and vet gates.

A missing row is a documentation defect. A row pointing only to research or planning is also a documentation defect for current code.

## Failure modes

Coverage can appear complete while remaining weak when:

- a page lists files but does not explain ownership or flow;
- a broad architecture page claims future behavior as current;
- a CLI command is present only in generated help or the root README;
- a compatibility alias is undocumented;
- an internal package is mapped to a page that never mentions its responsibility;
- research metrics are treated as product guarantees; or
- an implemented feature remains described as planned.

Review the linked content, not only the existence of links in this matrix.

## Code map

- `cmd/` - executable packages covered by this matrix.
- `internal/` - production package boundaries covered by this matrix.
- `docs/guides/` - task-oriented public workflows.
- `docs/reference/` - exact public contracts.
- `docs/architecture/` - current implementation ownership.
- `docs/operations/` - runtime operation and recovery.
- `docs/limits/` - current incomplete surfaces.
- `docs/research/` - evidence and methodology.
- `docs/planning/` - future and unresolved work.

## Related docs

- [Documentation Policy](../documentation-policy.md)
- [Documentation Procedure](../documentation-procedure.md)
- [Repository Layout](repository-layout.md)
- [Testing and Fixtures](testing-and-fixtures.md)
- [Current Product Limitations](../limits/current-limitations.md)
- [Roadmap](../planning/roadmap.md)

## Notes

This map should be updated in the same change that adds, removes, or materially reassigns a production package. It is an audit surface, not a substitute for substantive documentation.
