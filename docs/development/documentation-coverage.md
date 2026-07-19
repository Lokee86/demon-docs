# Documentation Coverage Map

Parent index: [Development](./README.md)

## Purpose

This document maps every current production package, public command family, and independent stateful flow to canonical documentation owners so implemented behavior does not exist only in code, tests, the root README, research notes, or the roadmap.

## Overview

Coverage is ownership-based rather than one-file-per-package. A focused utility package may be documented inside the architecture page for the subsystem it serves. A major independent boundary requires its own current architecture or operations owner.

Coverage also tracks independent stateful flows. One package pointer does not cover several distinct mutation, publication, rollback, lifecycle, or concurrency boundaries merely because they share a directory.

The map covers current production code under `cmd/` and `internal/` plus the stateful flows those packages implement. Planning packages do not exist in code and therefore are not counted as current implementation coverage.

## Coverage rules

A production boundary is covered when:

- one current guide, reference, architecture, operations, or development document owns its behavior;
- the document identifies important non-ownership boundaries;
- public commands and mutation scope are represented in reference documentation;
- implementation-facing pages provide useful code maps and tests;
- independent state transitions, mutation sequences, publication boundaries, rollback behavior, and recovery paths have canonical owners;
- known incomplete surfaces appear in limits or planning; and
- research pages are not the sole authority for shipped behavior.

The root README and roadmap are entry and status documents. They do not satisfy implementation coverage by themselves.

## Executable and application coverage

| Code boundary | Responsibility | Canonical current docs |
| --- | --- | --- |
| `cmd/ddocs/` | Canonical executable entry | [Application Orchestration](../architecture/application-orchestration.md), [CLI Reference](../reference/cli.md) |
| `cmd/demon/` | Repository-demon alias entry and argument normalization | [Application Orchestration](../architecture/application-orchestration.md), [Repository Demon](../operations/repository-demon.md), [CLI Reference](../reference/cli.md) |
| `internal/app/` | Command parsing, selection, orchestration, output, aggregate result | [Application Orchestration](../architecture/application-orchestration.md), [Reconciliation Command Lifecycle](../architecture/reconciliation-command-lifecycle.md), [CLI Reference](../reference/cli.md), command-specific guides |
| `internal/app/move.go` | Stateless move command integration | [Stateless Document Refactoring](../guides/document-refactoring.md), [Stateless Move Transaction](../architecture/stateless-move-transaction.md), [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md) |
| `internal/app/orphans.go` | Orphan health projection | [Document Health Checks](../guides/document-health-checks.md), [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md) |
| `internal/app/review_*.go` | Suggestion, change, undo, and block commands | [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md), [Review Ledger](../architecture/review-ledger.md), [Review Lifecycles](../architecture/review-lifecycles.md) |
| `internal/app/codemap_*.go` | Codemap export, benchmark, and precision commands | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md) |
| `internal/app/demon*.go` | Demon commands, shell hooks, feeder integration, detached startup | [Repository Demon](../operations/repository-demon.md), [Repository Demon Lease Protocol](../architecture/repository-demon-lease-protocol.md), [Host Adapter Feeder Integration](../operations/host-adapters.md) |
| `internal/app/reverse_index.go` | Reverse option resolution and mixed watch coordination | [Reverse Index Architecture](../architecture/reverse-indexes.md), [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md), [Adopting Reverse Indexes](../guides/reverse-indexes.md) |

## Repository and configuration coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/config/` | Defaults, TOML decoding, compatibility keys, selection, demon toggle mutation | [Configuration Reference](../reference/configuration.md), [Compatibility and Migrations](../reference/compatibility-and-migrations.md) |
| `internal/repository/` | Repository discovery, scope, containment, linked-worktree bootstrap | [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md), [Using Linked Git Worktrees](../guides/linked-worktrees.md) |
| `internal/ignore/` | `.docignore` policies, nested domains, permanent exclusions | [Ignore and Traversal](../architecture/ignore-and-traversal.md), [Configuration Reference](../reference/configuration.md) |
| `internal/ddrepo/` | Private Git-object repository, codecs, transactions | [Private Object Repository](../architecture/private-object-repository.md), [Repository State and Transactions](../architecture/repository-state-and-transactions.md), [Managed Files and State](../reference/managed-files-and-state.md) |

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
| `internal/links/` | Syntax parsing, inventory, identities, evidence, repair, moves, application | [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md), [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md), [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md), [Supported Link Syntax](../reference/supported-link-syntax.md), [Stateless Move Transaction](../architecture/stateless-move-transaction.md), [Stateless Document Refactoring](../guides/document-refactoring.md) |
| `internal/review/` | Append-only review events, fingerprints, policy replay, undo | [Review Ledger](../architecture/review-ledger.md), [Review Lifecycles](../architecture/review-lifecycles.md), [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md), [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md) |

## Stateful flow coverage

| Stateful flow | Current owner | Covered contracts |
| --- | --- | --- |
| Link occurrence and target lifecycle | [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md) | Identity reuse, parser invalidation, exact resolution, candidate discovery, statuses, generated-rewrite planning, refresh, and convergence. |
| Generated source and private-state publication | [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md) | Batch preflight, atomic file replacement, rollback, review publication, source refresh, state publication, suppression durability, and partial-completion recovery. |
| Suggestion, decision, change, block, and undo lifecycle | [Review Lifecycles](../architecture/review-lifecycles.md) | Stable fingerprints, decline scope, staleness, selection, change/run grouping, undo eligibility, repair blocks, replay, and compare-and-swap history publication. |
| `check`, `fix`, and `watch` command lifecycle | [Reconciliation Command Lifecycle](../architecture/reconciliation-command-lifecycle.md) | Configuration and scope, feature selection, planner/apply ordering, diagnostics, exit codes, orphan integration, and cross-subsystem partial completion. |
| Private object repository publication | [Private Object Repository](../architecture/private-object-repository.md) | Sharded record storage, deterministic codecs, transactions, conflict detection, and state-reference publication. |
| Stateless move transaction | [Stateless Move Transaction](../architecture/stateless-move-transaction.md) | Move planning, path remapping, preflight, rewrite ordering, filesystem mutation, and best-effort rollback. |
| Watch scheduling and serialization | [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md) | Debounce state, pending follow-up runs, mixed-watcher serialization, cancellation, and error propagation. |
| Repository demon ownership and feeder demand | [Repository Demon Lease Protocol](../architecture/repository-demon-lease-protocol.md) | Owner claims, feeder leases, stale recovery, detached startup, shutdown requests, and token-safe cleanup. |

New stateful behavior must be added here when its ownership cannot be explained completely by an existing row. A package row is not evidence that all flows inside that package are documented.

## Reverse-index coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/reverseindex/` | Root scope, traversal, codemap projection, rendering, apply, watch | [Reverse Index Architecture](../architecture/reverse-indexes.md), [Adopting Reverse Indexes](../guides/reverse-indexes.md) |

## Watcher and daemon coverage

| Package | Responsibility | Canonical current docs |
| --- | --- | --- |
| `internal/watch/` | Event scope, filters, debounce, serialized scheduling | [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md), [Dynamic Watch Scope](../operations/dynamic-watch-scope.md), [Watcher and Automation](../operations/watcher-and-automation.md) |
| `internal/demon/` | Owner lease, feeders, runtime files, logs, shutdown grace | [Repository Demon Lease Protocol](../architecture/repository-demon-lease-protocol.md), [Repository Demon](../operations/repository-demon.md), [Host Adapter Feeder Integration](../operations/host-adapters.md) |

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
| Install, version, and scoped help discovery | [Getting Started](../guides/getting-started.md) | [CLI Reference](../reference/cli.md), [Testing and Fixtures](testing-and-fixtures.md) |
| Initialize and first baseline | [Getting Started](../guides/getting-started.md) | [CLI Reference](../reference/cli.md), [Configuration Reference](../reference/configuration.md) |
| `status`, `config paths`, `config show` | [Getting Started](../guides/getting-started.md) | [Configuration Reference](../reference/configuration.md) |
| `check`, `fix`, selectors | [Getting Started](../guides/getting-started.md), [CI and Automation](../guides/ci-and-automation.md) | [Reconciliation Command Lifecycle](../architecture/reconciliation-command-lifecycle.md), [CLI Reference](../reference/cli.md), [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md) |
| Managed folder indexes and parent links | [Getting Started](../guides/getting-started.md) | [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md), [Managed Files and State](../reference/managed-files-and-state.md) |
| Local links | [Getting Started](../guides/getting-started.md) | [Supported Link Syntax](../reference/supported-link-syntax.md), [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) |
| Orphan health | [Document Health Checks](../guides/document-health-checks.md) | [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md) |
| `mv` | [Stateless Document Refactoring](../guides/document-refactoring.md) | [CLI Reference](../reference/cli.md), [Stateless Move Transaction](../architecture/stateless-move-transaction.md), [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) |
| Reverse indexes | [Adopting Reverse Indexes](../guides/reverse-indexes.md) | [Reverse Index Architecture](../architecture/reverse-indexes.md) |
| Suggestions and changes | [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md) | [Review Ledger](../architecture/review-ledger.md), [Review Lifecycles](../architecture/review-lifecycles.md), [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md) |
| Codemap export/benchmark/precision | [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md) | [Codemap Pipeline](../architecture/codemap-pipeline.md), [Codemap Evidence](../research/codemap-evidence.md), [CLI Reference](../reference/cli.md) |
| Foreground `watch` | [CI and Automation](../guides/ci-and-automation.md) | [Watcher and Automation](../operations/watcher-and-automation.md), [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md), [Dynamic Watch Scope](../operations/dynamic-watch-scope.md) |
| Repository demon | [CI and Automation](../guides/ci-and-automation.md) | [Repository Demon](../operations/repository-demon.md), [Repository Demon Lease Protocol](../architecture/repository-demon-lease-protocol.md) |
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
