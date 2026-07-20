---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-720d-b9ba-bfa926a5d8e5
document_type: general
policy_exempt: false
summary: This is the top-level documentation index and rulebook entry point for Demon Docs.
---
# Demon Docs Documentation

This is the top-level documentation index and rulebook entry point for Demon Docs.

The repository root [README](../README.md) is the product introduction and quick start. This documentation tree owns detailed user workflows, exact reference material, implemented architecture, operational behavior, research evidence, future planning, and contributor guidance.

## Documentation rules

- Documentation is organized by type and ownership, not only by feature name.
- Current implemented behavior must not live only in planning or research documents.
- Future or unresolved work belongs under `planning/`.
- Research results describe evidence, not automatically supported product behavior.
- Every normal documentation folder contains a `README.md` index.
- Every normal document includes purpose, overview, related docs, and notes sections.
- Implementation-facing architecture and development documents include code maps when useful.
- The detailed rules are in [Documentation Policy](documentation-policy.md).
- The required workflow is in [Documentation Procedure](documentation-procedure.md).

## Direct Files

<!-- doc-ledger:files:start -->

- [codemap-algorithm-development-log.md](codemap-algorithm-development-log.md) - Chronological codemap algorithm, benchmark, tuning, rejected-experiment, commit, and artifact record.
- [codemap-evidence.md](codemap-evidence.md) - Current codemap evidence signals, suggestion tiers, negative-evidence rules, and safety contract.
- [codemap-suggestion-algorithm.md](codemap-suggestion-algorithm.md) - Current codemap admission, scoring, ranking, tiering, measured baseline, and readiness contract.
- [documentation-policy.md](documentation-policy.md) - Documentation ownership, taxonomy, required document shapes, indexing rules, and maintenance policy.
- [documentation-procedure.md](documentation-procedure.md) - Standard process for creating, moving, updating, graduating, and removing documentation.
<!-- doc-ledger:files:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->

- [agent](agent/INDEX.md) - Agent documentation.
- [Architecture](architecture/INDEX.md) - Implemented ownership boundaries, state models, reconciliation pipelines, and internal system behavior.
- [Development](development/INDEX.md) - Contributor workflow, testing, repository layout, fixtures, and release verification.
- [Guides](guides/INDEX.md) - Task-oriented workflows for installing, adopting, refactoring, reviewing, and operating Demon Docs.
- [Limits](limits/INDEX.md) - Current user-visible limitations, incomplete surfaces, workarounds, ownership, and removal conditions.
- [Operations](operations/INDEX.md) - Watcher, repository demon, recovery, troubleshooting, and runtime behavior.
- [Planning](planning/INDEX.md) - Future, unresolved, proposed, or back-burnered work.
- [Reference](reference/INDEX.md) - Exact CLI, configuration, state, syntax, diagnostic, and file-format reference.
- [Research](research/INDEX.md) - Benchmarks, corpora, evaluation methodology, and recorded experimental evidence.
<!-- doc-ledger:folders:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Notes

The repository follows the product default and uses `INDEX.md` for generated folder indexes. The committed `demon-docs.toml` also enables file-level parent links and indexes the development fixture script.