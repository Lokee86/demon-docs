---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7366-a679-359c693253d2
document_type: general
policy_exempt: false
summary: Active Demon Docs priorities for correctness, diagnostics, incremental reconciliation, and operational hardening.
---
# Demon Docs Roadmap

Parent index: [Planning](./INDEX.md)

## Purpose

Record the small set of active Demon Docs priorities. Shipped behavior belongs in architecture, reference, operations, guides, and limitations documents rather than being duplicated here.

## Overview

Demon Docs is a deterministic documentation-maintenance engine. This roadmap is limited to work that strengthens links, indexes, schemas, authored codemaps, reverse indexes, review history, and watcher automation.

Repository graph intelligence belongs to ArcanaGraph. Context discovery and delivery belong to Grimoire Context. Neither is unfinished Demon Docs work.

## Current status

The current main branch provides:

- recursive indexes, parent navigation, orphan health, and local-link repair;
- explicit link-aware moves and observed filesystem-move recovery;
- frontmatter and document-format policy with independent caches;
- authored codemap management and file/folder reverse indexes;
- review decisions, guarded undo, and repair blocks;
- foreground watching and the optional repository demon; and
- a checked-in correctness smoke harness for source and release binaries.

The latest tagged release is `v0.3.5`. Exact behavior is documented outside this roadmap.

## Ownership boundary

Demon Docs owns deterministic maintenance of repository-owned Markdown and explicit managed surfaces. It may exchange versioned facts with sibling Warlock tools without absorbing their implementation responsibilities.

It does not own polyglot repository graphs, symbols, dependencies, impact analysis, agent context delivery, or autonomous prose generation.

## Current Product: Implemented in the Current Branch

Canonical product details live in [Architecture](../architecture/INDEX.md), [Reference](../reference/INDEX.md), [Operations](../operations/INDEX.md), [Guides](../guides/INDEX.md), and [Current Product Limitations](../limits/current-limitations.md). This roadmap should not become a second feature reference.

## Active Work

### 1. Path-scoped link and index reconciliation

Ordinary Markdown edits already scope frontmatter and document-format validation. Link and folder-index maintenance still performs broader work than necessary.

- retain changed-source and changed-target batches;
- update only affected link sources and index folders when evidence is complete;
- preserve deterministic ordering and serial writes; and
- fall back to full reconciliation on overflow, uncertain events, configuration changes, or incomplete state.

Measure end-to-end watcher and move latency rather than debounce duration alone.

### 2. Stable machine-readable diagnostics

Add a versioned native JSON format without destabilizing human-readable output.

- stable diagnostic code, severity, subsystem, and message;
- repository-relative path and source position where available;
- documented exit semantics; and
- explicit schema versioning for CI adapters.

Consider SARIF only after the smaller native contract is stable.

### 3. Link and reverse-index correctness

- validate heading fragments against one documented deterministic anchor model;
- improve reverse-index coverage, unresolved-target, and scope diagnostics; and
- add focused move-aware authored-reference and nested-root coverage.

External network reachability remains a separate opt-in candidate, not part of local path reconciliation.

### 4. Watcher and demon resilience

- stress large moves and watcher-event bursts;
- retain race-focused single-owner lease coverage;
- verify stale-owner recovery and repeated restarts;
- verify Windows, Bash, PowerShell, and linked-worktree lifecycle paths; and
- improve diagnostics for slow or failed reconciliation.

`ddocs check`, `ddocs fix`, and explicit `ddocs mv` remain authoritative recovery surfaces.

### 5. Review and codemap evidence hardening

- validate recommendations on broader labeled repositories;
- compare scoring changes against pinned precision and holdout samples;
- preserve no-pruning defaults and remembered declines;
- stress review-ledger append, undo-depth, and stale-evidence behavior; and
- keep quality claims tied to named datasets.

## Near-Term Hardening

Priority order:

1. Path-scoped link and index reconciliation.
2. Stable machine-readable diagnostics.
3. Heading-fragment and reverse-index diagnostics.
4. Watcher, lease, and large-move stress coverage.
5. Review-ledger and codemap corpus hardening.
6. Broader release-platform and installation verification.

Shared immutable source snapshots should be introduced only where measurements show duplicated reads or parsing. Incremental changed-region Markdown parsing remains deferred until whole-document parsing is a material bottleneck after path scoping.

## Warlock Toolchain Boundaries

- [ArcanaGraph](https://github.com/Lokee86/arcana-graph) owns normalized repository relationships, polyglot facts, symbols, dependencies, and graph queries.
- **Grimoire Context** owns bounded context discovery, selection, packaging, provenance, truncation, and delivery.
- The [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain) owns shared terminology, cross-tool contracts, and integration direction.

The retained [code-intelligence](./code-intelligence/INDEX.md), [agent-context](./agent-context-and-integrations.md), and [context-benchmark](../research/context-injection-benchmarking.md) pages are historical design provenance, not active Demon Docs commitments.

## Optional LLM Assistance

Optional LLM assistance may propose changes from deterministic evidence, but it is not an active priority and cannot become a correctness dependency.

## Principles

- Deterministic inputs produce stable plans, diagnostics, and ordering.
- Generated ownership remains explicit and narrow.
- Writes remain serial, atomic, hash-guarded, and reviewable.
- Ambiguity is reported rather than guessed.
- Declines remain effective until evidence materially changes.
- Watchers automate the static core rather than defining a second behavior model.
- Performance changes require representative measurements and conservative fallbacks.
- Tool boundaries remain explicit.

## Explicit Non-Goals

- General repository graph or symbol intelligence inside Demon Docs.
- Agent context assembly or delivery through the repository demon.
- Required network, LLM, language-server, or external-indexer dependencies.
- Automatic semantic prose rewriting.
- Silent ambiguous repair.
- Universal quality claims from one repository or benchmark corpus.

## Code map

- `internal/reconcile/` and `internal/links/` — indexes, link state, moves, and rewrites.
- `internal/frontmatter/` and `internal/documentpolicy/` — document policy.
- `internal/reverseindex/` — file/folder reverse projections.
- `internal/watch/` and `internal/demon/` — automation and ownership lifecycle.
- `internal/review/`, `internal/codemap*`, and `internal/evidence/` — decisions, codemap execution, and research.
- `tools/smoke/` — black-box correctness verification.

## Implementation sequence

Implement one owning seam at a time. Preserve full-pass fallbacks until scoped evidence is complete, keep writes serial, and move shipped details into canonical product documentation.

## Related docs

- [Planning](INDEX.md)
- [Current Product Limitations](../limits/current-limitations.md)
- [Validation Cache](../architecture/validation-cache.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Watch Scheduler](../architecture/watch-scheduler.md)
- [Reverse Indexes](../architecture/reverse-indexes.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)

## Notes

Resolved work should leave this roadmap rather than accumulating as history. Benchmark results belong under `research/`; exact behavior belongs in canonical product documentation.
