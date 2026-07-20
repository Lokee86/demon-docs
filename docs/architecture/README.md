---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7379-acb5-839b25450de6
document_type: general
policy_exempt: false
summary: Implemented ownership boundaries, state models, reconciliation pipelines, and internal system behavior. Planned architecture remains under planning/ until the owning code exists.
---
# Architecture

Parent index: [Demon Docs Documentation](../README.md)

Implemented ownership boundaries, state models, reconciliation pipelines, and internal system behavior. Planned architecture remains under `planning/` until the owning code exists.

## Direct Files

<!-- doc-ledger:files:start -->

- [application-orchestration.md](application-orchestration.md) - CLI application coordination, subsystem selection, planning, application, and command boundaries.
- [codemap-corpus-adapters.md](codemap-corpus-adapters.md) - Repository file, dependency, symbol, related-document, and bounded-history facts supplied to codemap analysis.
- [codemap-evidence-and-ranking.md](codemap-evidence-and-ranking.md) - Candidate evidence, fingerprints, admission, scoring, fan-out discounting, output bounds, and suggestion tiers.
- [codemap-extraction-and-dataset.md](codemap-extraction-and-dataset.md) - Authored code-map syntax, target normalization and resolution, deterministic datasets, holdout stripping, and selected insertion.
- [codemap-pipeline.md](codemap-pipeline.md) - End-to-end codemap ownership from authored extraction through repository corpus facts, ranking, evaluation, and review-selected insertion.
- [generated-rewrite-publication.md](generated-rewrite-publication.md) - Authored-source preflight, atomic replacement, review publication, rollback, metadata refresh, and private-state convergence.
- [ignore-and-traversal.md](ignore-and-traversal.md) - Repository-root and nested `.docignore` domains, permanent exclusions, traversal pruning, and consumer boundaries.
- [link-reconciliation-state-machine.md](link-reconciliation-state-machine.md) - Link identity reuse, target resolution, repair statuses, review controls, generated rewrites, and graph convergence.
- [managed-markdown-transformation.md](managed-markdown-transformation.md) - Structural managed-section editing, entry transitions, parent links, newline preservation, containment, and apply behavior.
- [markdown-link-reconciliation.md](markdown-link-reconciliation.md) - Repository-local link inventory, identity evidence, deterministic repair, and source-preserving writes.
- [private-object-repository.md](private-object-repository.md) - Sharded Git-object record storage, deterministic codecs, transactions, conflict detection, and state-reference publication.
- [reconciliation-command-lifecycle.md](reconciliation-command-lifecycle.md) - `check`, `fix`, and `watch` selection, planning, mutation order, diagnostics, exit codes, and partial-completion boundaries.
- [reconciliation-pipeline.md](reconciliation-pipeline.md) - Documentation-tree scan, managed index planning, parent links, and shared reconciliation flow.
- [repository-demon-lease-protocol.md](repository-demon-lease-protocol.md) - Single-owner claims, feeder demand leases, stale recovery, detached startup, and token-safe shutdown.
- [repository-scope-and-worktrees.md](repository-scope-and-worktrees.md) - Initialized discovery, scope containment, standalone operation, and linked-worktree state isolation.
- [repository-state-and-transactions.md](repository-state-and-transactions.md) - Private state domains, authored-file boundaries, transaction composition, and rebuildability.
- [reverse-indexes.md](reverse-indexes.md) - Authored codemap projection into configured code-folder reverse indexes.
- [review-ledger.md](review-ledger.md) - Append-only suggestion decisions, applied changes, undo data, repair blocks, policy projection, and Git-backed publication.
- [review-lifecycles.md](review-lifecycles.md) - Suggestion decisions, staleness, selection, applied changes, undo, repair blocks, and append-only event replay.
- [stateless-move-transaction.md](stateless-move-transaction.md) - Explicit move planning, path remapping, preflight, rewrite ordering, and best-effort rollback.
- [watch-scheduler.md](watch-scheduler.md) - Debounce state, follow-up runs, mixed-watcher serialization, cancellation, and error propagation.
<!-- doc-ledger:files:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Notes

Architecture pages document current code. Future provider contracts, repository graphs, and agent context systems remain in planning.
