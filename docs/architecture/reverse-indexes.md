---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7140-8037-b4805305b2cd
document_type: general
policy_exempt: false
summary: This document describes the implemented reverse-index projection from authored documentation codemap targets into managed indexes inside configured code roots.
---
# Code-Folder Reverse Indexes

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes the implemented reverse-index projection from authored documentation codemap targets into managed indexes inside configured code roots.

## Overview

Reverse indexing is the third reconciliation subsystem alongside documentation indexes and local links. It reads explicit targets from configured codemap sections, resolves those targets against the current repository filesystem, inventories selected code roots, and writes deterministic managed blocks into code-folder index files.

The current implementation deliberately operates at file and folder level. General repository relationships and symbol adapters belong to ArcanaGraph; Demon Docs reverse indexes do not infer semantic ownership or use LLM judgment.

## Code root

```text
internal/reverseindex/
internal/app/reverse_index.go
```

## Responsibilities

The reverse-index boundary owns:

- resolving configured or command-line reverse roots;
- recursively traversing those roots within repository scope;
- loading repository-root and nested `.docignore` rules during traversal;
- extracting authored codemap targets through `internal/codemap`;
- resolving existing file and folder targets;
- grouping source documents by resolved code target;
- inventorying eligible code files under selected folders;
- rendering one deterministic managed reverse-index block per selected folder;
- reporting unresolved in-scope codemap targets;
- reporting eligible in-scope code files with no resolved authored file reference during read-only checks; and
- planning, checking, applying, and watching reverse-index updates.

## Does not own

Reverse indexing does not own:

- authored codemap relationships;
- missing-link candidate generation or ranking;
- ordinary Markdown link reconciliation;
- symbol-level code references;
- dependency or call graphs;
- judgments that an existing codemap link is irrelevant;
- repair of ambiguous authored targets; or
- code-root selection beyond explicit configuration or command flags.

## Inputs

A build receives:

```text
repository root
configured documentation root
one or more reverse roots
Demon Docs configuration
configured codemap headings
current repository files
current .docignore hierarchy
```

`codemap.BuildDataset` extracts targets from documents under the documentation root. The reverse-index builder then accepts resolved targets only when they fall inside a selected reverse root and survive traversal exclusions.

The builder fails when no reverse roots are selected, no configured codemap section exists, or matching codemap sections contain no targets. Individual unresolved targets that could belong to the selected scope become sorted diagnostics rather than guessed relationships.

## Root resolution and scope

Configured roots come from:

```toml
[reverse_index]
roots = ["client", "services/game-server"]
```

The compatibility key `folders` is also accepted when `roots` is absent.

Repeated `--reverse-root PATH` values replace configured roots for one invocation. Relative command-line roots resolve from the current working directory; configured roots resolve from the repository root. Every selected root must remain inside the repository, outside the documentation root, outside permanent exclusions, and outside nested Git worktrees.

Duplicate and overlapping roots are normalized so traversal remains deterministic.

## Traversal and inventory

The builder discovers eligible folders below the selected roots. Repository-root and nested `.docignore` files are loaded with domains rooted at the directory containing each ignore file. Permanently excluded directories are pruned before traversal.

For each selected folder, inventory identifies eligible direct files. Existing files named as the configured index file are treated as index surfaces rather than code entries. A folder is selected for output when at least one of these is true:

- it contains eligible direct code files;
- an authored codemap target resolves to the folder; or
- it already contains a managed reverse-index block that may need reconciliation or removal.

## Target facts

Resolved codemap targets become one of two current fact types:

```text
folder target -> documentation references attached to that folder
file target   -> documentation references attached to that exact file
```

A file target does not imply documentation for its containing folder. A folder target is rendered as folder documentation and does not automatically mark every descendant file as documented.

Multiple documents may reference the same target. Every distinct source document is retained and rendered in sorted order.

## Generated format

The generated block uses the configured marker prefix and fixed `reverse-index` section name:

```markdown
<!-- doc-ledger:reverse-index:start -->

## Code Files

Folder documentation:
- [Architecture](../../docs/architecture/example.md)

- [server.go](server.go)
  - [Server Architecture](../../docs/architecture/server.md)

<!-- doc-ledger:reverse-index:end -->
```

When a folder has no index file, Demon Docs creates one with a title, a short generated-purpose sentence, and the managed block. When an index already exists, authored content outside the reverse-index markers is preserved.

An incomplete marker pair is an error. Reverse indexing does not take ownership of ordinary documentation-index marker blocks in the same file.

## Reconciliation flow

```text
resolve roots and codemap headings
-> discover scoped folders and ignore hierarchy
-> build authored codemap dataset
-> resolve in-scope file and folder targets
-> inventory eligible files and existing managed indexes through bounded workers
-> prepare rendered blocks and current-index comparisons through bounded workers
-> merge updates and errors serially in sorted folder order
-> report or apply file updates serially
```

`check --reverse` reports pending updates and reverse-index orphans without writing. `fix --reverse` applies only the planned file updates; orphan status is check-only because the command does not invent authored links. `watch --reverse` runs the same build and apply path after relevant filesystem events.

Folder inventory and per-folder reconciliation use one bounded 16-worker preparation seam. Inventory workers read the current managed-marker state, directory entries, and immutable ignore hierarchy, then return sorted detached file lists. Reconciliation workers render one folder block and compare it with the current index source. Results remain indexed by sorted folder path and merge serially, so worker completion order cannot alter update order, index counts, or first-error selection. No worker writes repository files; `Apply` remains the sole serial publication path.

When reverse indexing is selected together with documentation indexes or links, foreground watch runs the normal watcher and reverse-index watcher concurrently under one command context. This changes scheduling, not correctness or output ownership.

## Preparation performance

A retained Windows benchmark builds a stable reverse-index plan across 128 code folders with four Go files per folder and one authored codemap document per folder. The comparison used five one-iteration runs, `GOMAXPROCS=16`, and commit `23b0a3f` as the serial baseline.

Excluding the first sample to reduce host filesystem and antivirus warm-up noise, mean planning time improved from 326.2 milliseconds to 95.4 milliseconds: a 3.42x speedup and 70.8% latency reduction. The benchmark is `BenchmarkReverseIndexPreparation` in `internal/reverseindex/preparation_benchmark_test.go`.

```bash
go test ./internal/reverseindex -run '^$' -bench '^BenchmarkReverseIndexPreparation$' -benchmem -count=5 -benchtime=1x
```

The optimization prioritizes latency and uses additional transient memory for detached concurrent results. Allocation and byte totals remain observable through `-benchmem`.

## Diagnostics

Fatal build errors include:

- no selected reverse roots;
- no configured codemap headings;
- no matching codemap section under the docs root;
- matching sections with no code targets;
- roots outside permitted repository scope; and
- malformed managed marker pairs.

Non-fatal target diagnostics identify source document, source line, resolution status, and target. Diagnostics are sorted before output.

Eligible files in the selected inventory with no entry in the resolved authored file facts are reported during `check --reverse` as sorted `message: Reverse-index orphan: <repo-relative-path>` messages. The inventory already excludes hard ignores, `.docignore` paths, generated reverse-index files, and files outside selected roots.

## Invariants and safety boundaries

- Authored codemap sections are the only source of documentation-to-code relationships.
- Existing codemap links are never classified as irrelevant or removal candidates.
- Only explicit file and folder targets create current reverse-index facts.
- Unresolved or out-of-scope targets do not create substitute edges.
- Orphan health does not create substitute edges or alter the fix/watch projection path.
- Output is deterministic for the same repository snapshot and configuration.
- Writes remain inside configured index files and the reverse-index marker pair.
- Authored content outside the managed block is preserved.
- Static `check` and `fix` remain authoritative; watch and the repository demon are optional automation.

## Code map

- `internal/reverseindex/reverseindex.go` - plan and collected fact types.
- `internal/reverseindex/paths.go` - configured and command-line root resolution.
- `internal/reverseindex/scope.go` - root validation and scope normalization.
- `internal/reverseindex/traversal.go` - folder discovery and nested ignore loading.
- `internal/reverseindex/inventory.go` - bounded folder inventory, shared worker scheduling, and deterministic indexed merge.
- `internal/reverseindex/targets.go` - resolved target acceptance and grouping.
- `internal/reverseindex/build.go` - complete deterministic build, bounded folder reconciliation preparation, and serial plan merge.
- `internal/reverseindex/render.go` - managed block and document-link rendering.
- `internal/reverseindex/apply.go` - file-update application.
- `internal/reverseindex/watch.go` - reverse-index watch scheduling.
- `internal/app/reverse_index.go` - CLI option resolution and mixed-feature watch coordination.

## Tests

Focused coverage includes root scope, nested traversal, target resolution, bounded folder preparation, deterministic update and error order, rendering, managed-block preservation, command integration, and watch behavior.

```bash
go test ./internal/reverseindex ./internal/app -count=1
```

## Related docs

- [Adopting Reverse Indexes](../guides/reverse-indexes.md)
- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Codemap Pipeline](codemap-pipeline.md)
- [Ignore and Traversal](ignore-and-traversal.md)
- [Codemap Missing-Link Evidence](../research/codemap-evidence.md)
- [Planned Code Intelligence](../planning/code-intelligence/INDEX.md)

## Notes

Symbol references, dependency facts, move-aware authored-target repair, and richer coverage exports remain future work. They must not be inferred from the current file/folder projection.
