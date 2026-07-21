---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-72fb-a795-935b163af75a
document_type: general
policy_exempt: false
summary: 'This document describes the implemented documentation-tree scan and forward-index reconciliation flow: how filesystem inventory becomes a deterministic plan of index, description, and parent-link updates.'
---
# Reconciliation Model

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes the implemented documentation-tree scan and forward-index reconciliation flow: how filesystem inventory becomes a deterministic plan of index, description, and parent-link updates.

## Overview

Demon Docs keeps folder indexes in a predictable shape by scanning the managed tree, matching current files and folders to existing generated entries, and planning the smallest repository-contained update set.

The byte-level transformation and newline-preservation boundary is owned separately by [Managed Markdown Transformation](managed-markdown-transformation.md). Repository-local link reconciliation is owned by [Markdown Link Reconciliation](markdown-link-reconciliation.md).

## Code root

```text
internal/scan/
internal/markdown/
internal/reconcile/
```

## Responsibilities

The forward reconciliation boundary owns documentation-tree inventory, index creation, managed-section planning, parent-link planning, description preservation, deterministic ordering, and application of documentation-index updates.

## Does not own

It does not own semantic topic placement, authored prose outside managed blocks, repository-local link identity, reverse-index projection, daemon lifecycle, or codemap inference.

## Flow

```text
resolve repository and configuration
-> scan documentation tree
-> read existing indexes and parent-editable documents through bounded workers
-> parse existing managed sections from the retained source snapshot
-> prepare independent folder and parent-link plans through bounded workers
-> merge updates and matched-entry claims serially in folder order
-> report or apply the plan serially
-> verify deterministic state on the next pass
```

## Parallel preparation

The scanner still produces one deterministic folder order. After scanning, a bounded 16-worker pool reads existing index files and parent-editable documents into detached indexed results. Each source is retained once and reused for root-title discovery, managed-entry parsing, parent-link comparison, and expected-old-content guards.

Once cross-folder preservation facts and unmatched-name counts are computed serially, each folder independently prepares its index update, parent-link updates, and matched-entry claims. Workers read immutable maps and publish only their indexed detached result. Updates and matched claims then merge serially in the original folder order, so completion timing cannot change output ordering, stale-entry reporting, or first-error selection. File application remains serial and retains existing stale-source protections.

The Markdown parent-link regular-expression cache uses concurrent-safe publication because multiple folder preparations may request the same configured label simultaneously.

## Scan Model

The scanner starts from the configured managed root and builds a tree of folders.

- The managed root is the folder Demon Docs owns, such as `docs/` by default.
- Normal folders are folders that can have their own index file.
- Draft folders, also called stub folders in the implementation, are the configured draft folder name such as `stubs/` by default.
- Direct files are indexed files that live directly inside a normal folder.
- Stub files are indexed files that live directly inside the draft folder for a normal folder.
- Direct folders are child folders of a normal folder, excluding the draft folder itself.
- Draft folders do not get their own index file.

The scan model is descriptive only. It records what exists on disk and where Demon Docs should look for managed content.

## Folder Index Behavior

Demon Docs treats configured folder index files as structured documents with managed sections. The exact structural recognition, migration, bounded replacement, parent-line editing, and byte-preservation rules are documented in [Managed Markdown Transformation](managed-markdown-transformation.md).

- Managed blocks are wrapped in HTML comment markers.
- The managed sections are Direct Files, Stub Files, and Direct Folders.
- Human-authored content outside the managed markers is preserved.
- Existing managed entries are parsed from those marker blocks before reconciliation rewrites them.
- The default index filename is `INDEX.md`. Repositories can retain earlier or project-specific conventions with an explicit setting such as `index_file = "README.md"`, `index_file = "!README.md"`, or `index_file = "!INDEX.md"`.
- Folder index files get `Parent index` links by default.
- Indexed files do not get `Parent index` links unless `indexed_files = true` is set.

If a README already has the expected managed sections, Demon Docs updates only the content inside those managed blocks.

Goldmark determines which headings and HTML comments are Markdown structure. Heading- and marker-like text inside fenced code blocks is code content and is never treated as a managed section. Parent-link-shaped lines inside fenced code are likewise examples rather than editable parent links. This is an intentional compatibility correction: fenced examples are not treated as real headings or managed sections.

## Missing README Creation

During reconciliation, Demon Docs creates missing index files where they belong.

- Normal folders get an index file if one is missing.
- The root folder gets an index file if one is missing.
- Draft folders do not get an index file.

The generated folder-index template includes the managed sections so reconciliation can fill them in on the first pass.

## Parent Index Behavior

Demon Docs maintains parent index lines according to the configured parent-link toggles.

- The root index file has no parent index line.
- Child folder index files point to the parent folder using `../<index file>`.
- Normal docs point to their folder index using `./<index file>` when `indexed_files = true`.
- Stub docs point to the owning parent folder index using `../<index file>` when `indexed_files = true`.
- `folder_indexes = false` disables parent links in child folder indexes.
- `indexed_files = false` disables parent links in indexed files.

The parent index line is only written for file types that are configured as editable for parent links.

Parent-link insertion, replacement, and removal preserve whether the source document ended with a newline. This is an intentional compatibility guarantee for source preservation.

## Entry Preservation

Reconciliation prefers to preserve stable, existing index content when the target still belongs in the same place.

- Stable entries keep their existing descriptions.
- Graduating a stub file into a normal file removes a leading `Stub:` prefix when present.
- Moving a canonical file into the draft folder adds a `Stub:` prefix when needed.
- Unambiguous cross-folder file and folder moves preserve descriptions.
- Stale entries are removed from managed blocks and reported as reconciliation messages.

This preservation is intentionally narrow. Demon Docs matches by the current filesystem model and existing managed entries; it does not try to guess every historical rename pattern.

## Preparation performance

A retained Windows benchmark constructs a stable documentation tree with 128 folders and four Markdown documents per folder, then measures a complete read-only `Tree` plan. The comparison used five one-iteration runs, `GOMAXPROCS=16`, and commit `23b0a3f` as the serial baseline.

Excluding the first sample to reduce host filesystem and antivirus warm-up noise, mean planning time improved from 531.6 milliseconds to 279.7 milliseconds: a 1.90x speedup and 47.4% latency reduction. The benchmark is `BenchmarkTreePreparation` in `internal/reconcile/preparation_benchmark_test.go`.

```bash
go test ./internal/reconcile -run '^$' -bench '^BenchmarkTreePreparation$' -benchmem -count=5 -benchtime=1x
```

The optimization targets latency rather than allocation volume. Detached concurrent source and folder results increase transient memory, which remains visible in `-benchmem` output.

## Markdown Link Behavior

Link reconciliation scans Markdown sources throughout the repository root rather than only the configured docs root. It records local inline links, images, reference definitions, explicit and collapsed reference uses, path-based wiki links and embeds, supported local HTML targets, stable file IDs, fingerprints, path history, and reverse-link records in the private `.ddocs/` object repository.

The first link-enabled fix or watch pass records a baseline and reports issues without repairing links. Later passes preserve direct valid targets and can repair a moved target when its recorded ID, exact fingerprint, case-only path, or unique filename candidate identifies one result. Multiple candidates remain unchanged and are reported for user resolution. Undefined explicit or collapsed reference labels are reported without interpreting shortcut bracket syntax as a link.

Relative and absolute filesystem links are both checked. Targets may be non-Markdown files or may resolve outside the repository. Only Markdown source files inside the repository are rewritten, and only the resolved destination path changes; labels, titles, wiki aliases, queries, fragments, angle wrapping, and surrounding prose remain intact.

Generated rewrites are planned deterministically and then applied through a bounded worker pool. Each source still uses an expected-content hash and same-directory atomic replacement, so concurrency changes throughput rather than ownership or output semantics.

## Explicit Stateless Moves

`ddocs mv` exposes the same link parser, target resolver, and syntax-preserving renderer as an explicit filesystem refactoring command. Unlike normal link reconciliation, it does not depend on prior identity state: it resolves links against the pre-move filesystem, maps the source and every descendant to the requested destination, and recalculates affected paths before changing files.

The move planner scans within an explicit repository boundary, supports files and directories, recalculates relative links inside moved Markdown sources, and rewrites incoming links to moved targets. It refuses an affected ambiguous wiki target rather than choosing a candidate. `--dry-run` returns the complete plan without writing. Apply verifies source hashes, moves the filesystem entry, performs atomic Markdown replacements, and attempts to restore both content and location if a rewrite fails.

See [Stateless Document Refactoring](../guides/document-refactoring.md).

## Document health projection

During `check`, a link-enabled pass projects the current link graph back onto managed Markdown documents under the docs root. Normal documents with no meaningful inbound link are reported as orphans. Folder indexes, draft documents, self-links, and inbound links originating from indexes or drafts are excluded from the reachability decision.

This projection is diagnostic only. It does not add links or decide which document should own the missing relationship. See [Document Health Checks](../guides/document-health-checks.md).

Documentation policy, documentation indexes, Markdown links, and code-folder reverse indexes are selected independently with `-d` / `--docs`, `-i` / `--indexes`, `-l` / `--links`, and `-r` / `--reverse`. `--docs` selects indexes, frontmatter, and document-body format; `--indexes` selects indexes only. When any selector is supplied, only selected systems run.

## Safety Boundaries

Demon Docs is a reconciliation tool, not a semantic documentation author.

- It does not decide which folder should own a topic.
- It rewrites only the resolved filesystem path portion of recognized local links.
- It does not edit target files, binary files, link labels, aliases, or authored prose.
- It does not automatically choose among multiple plausible targets.
- Codemap evidence produces review candidates rather than authored links or removal recommendations.
- `check` reports pending reconciliation, but it does not inspect git status.

Those boundaries keep the tool predictable and keep hand-authored prose under human control.

## Code map

- `internal/scan/scan.go` — recursive documentation-tree inventory.
- `internal/markdown/markdown.go` — managed-section parsing, concurrent-safe parent-pattern reuse, and source-preserving Markdown edits.
- `internal/reconcile/reconcile.go` — forward-index orchestration and serial application.
- `internal/reconcile/source_loading.go` — bounded index and parent-editable document loading with deterministic indexed merge.
- `internal/reconcile/preparation.go` — independent folder preparation and deterministic serial plan merge.
- `internal/links/` — repository-local link inventory, resolution, state, diagnostics, rewrites, and stateless move planning.
- `internal/app/move.go` — explicit `ddocs mv` command orchestration.
- `internal/app/orphans.go` — link-graph projection used by the orphan health check.
- `internal/model/model.go` — shared reconciliation structures.
- `internal/app/app.go` — `check`, `fix`, and `watch` orchestration.

## Tests

Focused coverage includes scan scope, managed Markdown behavior, source preservation, move transitions, configuration, deterministic indexed merge and errors, bounded preparation concurrency, parent-cache race safety, line endings, and reconciliation scope.

```bash
go test ./internal/scan ./internal/markdown ./internal/reconcile -count=1
```

## Related docs

- [Architecture](INDEX.md)
- [Getting Started](../guides/getting-started.md)
- [Configuration Reference](../reference/configuration.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Managed Markdown Transformation](managed-markdown-transformation.md)
- [Application Orchestration](application-orchestration.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Document Health Checks](../guides/document-health-checks.md)

## Notes

The documentation scanner describes current filesystem structure; it does not decide whether the repository selected the right documentation taxonomy.
