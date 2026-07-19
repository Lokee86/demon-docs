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

- [documentation-policy.md](documentation-policy.md) - Documentation ownership, taxonomy, required document shapes, indexing rules, and maintenance policy.
- [documentation-procedure.md](documentation-procedure.md) - Standard process for creating, moving, updating, graduating, and removing documentation.
<!-- doc-ledger:files:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->

- [Architecture](architecture/README.md) - Implemented ownership boundaries, state models, reconciliation pipelines, and internal system behavior.
- [Development](development/README.md) - Contributor workflow, testing, repository layout, fixtures, and release verification.
- [Guides](guides/README.md) - Task-oriented workflows for installing, adopting, refactoring, reviewing, and operating Demon Docs.
- [Operations](operations/README.md) - Watcher, repository demon, recovery, troubleshooting, and runtime behavior.
- [Planning](planning/README.md) - Future, unresolved, proposed, or back-burnered work.
- [Reference](reference/README.md) - Exact CLI, configuration, state, syntax, diagnostic, and file-format reference.
- [Research](research/README.md) - Benchmarks, corpora, evaluation methodology, and recorded experimental evidence.
<!-- doc-ledger:folders:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Notes

Demon Docs uses `README.md` rather than Space Rocks' `!INDEX.md` convention because `README.md` is the product default. The committed `demon-docs.toml` additionally enables file-level parent links and indexes the development fixture script.
