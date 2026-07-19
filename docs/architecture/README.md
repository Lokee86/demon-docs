# Architecture

Parent index: [Demon Docs Documentation](../README.md)

Implemented ownership boundaries, state models, reconciliation pipelines, and internal system behavior. Planned architecture remains under `planning/` until the owning code exists.

## Direct Files

<!-- doc-ledger:files:start -->

- [application-orchestration.md](application-orchestration.md) - CLI application coordination, subsystem selection, planning, application, and command boundaries.
- [codemap-corpus-adapters.md](codemap-corpus-adapters.md) - Repository file, dependency, symbol, related-document, and bounded-history facts supplied to codemap analysis.
- [codemap-evidence-and-ranking.md](codemap-evidence-and-ranking.md) - Candidate evidence, fingerprints, admission, scoring, fan-out discounting, output bounds, and suggestion tiers.
- [codemap-extraction-and-dataset.md](codemap-extraction-and-dataset.md) - Authored code-map syntax, target normalization and resolution, deterministic datasets, holdout stripping, and selected insertion.
- [codemap-pipeline.md](codemap-pipeline.md) - End-to-end codemap ownership from authored extraction through review-selected insertion.
- [ignore-and-traversal.md](ignore-and-traversal.md) - Repository-root and nested `.docignore` domains, permanent exclusions, traversal pruning, and consumer boundaries.
- [managed-markdown-transformation.md](managed-markdown-transformation.md) - Structural managed-section editing, entry transitions, parent links, newline preservation, containment, and apply behavior.
- [markdown-link-reconciliation.md](markdown-link-reconciliation.md) - Repository-local link inventory, identity evidence, deterministic repair, and source-preserving writes.
- [reconciliation-pipeline.md](reconciliation-pipeline.md) - Documentation-tree scan, managed index planning, parent links, and shared reconciliation flow.
- [repository-scope-and-worktrees.md](repository-scope-and-worktrees.md) - Initialized discovery, scope containment, standalone operation, and linked-worktree state isolation.
- [repository-state-and-transactions.md](repository-state-and-transactions.md) - Private object repository, identity/history state, transaction boundaries, and rebuildability.
- [reverse-indexes.md](reverse-indexes.md) - Authored codemap projection into configured code-folder reverse indexes.
- [review-ledger.md](review-ledger.md) - Append-only suggestion decisions, applied changes, undo data, repair blocks, policy projection, and Git-backed publication.
<!-- doc-ledger:files:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Notes

Architecture pages document current code. Future provider contracts, repository graphs, and agent context systems remain in planning.
