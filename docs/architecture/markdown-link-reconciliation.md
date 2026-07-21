---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-785d-a2b6-d4a132ddf955
document_type: general
policy_exempt: false
summary: This document describes the implemented repository-local link graph, supported source forms, persistent identity evidence, deterministic move repair, and source-preserving write boundary.
---
# Markdown Link Reconciliation

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes the implemented repository-local link graph, supported source forms, persistent identity evidence, deterministic move repair, and source-preserving write boundary.

## Overview

Demon Docs maintains a repository-scoped graph of local Markdown links. This is a focused link graph for validation and path repair; it is not the later repository, code, symbol, or agent-context graph.

This page owns the subsystem overview and supported behavior. The detailed record transitions are owned by [Link Reconciliation State Machine](link-reconciliation-state-machine.md), while filesystem writes and multi-store publication are owned by [Generated Rewrite Publication](generated-rewrite-publication.md).

## Code root

```text
internal/links/
```

## Responsibilities

This boundary owns link parsing, local target resolution, file identity and path history, incoming-link groups, deterministic repair evidence, source-hash validation, atomic replacement, and link diagnostics.

## Does not own

It does not own authored link labels or prose, external target contents, semantic documentation relationships, codemap inference, heading-anchor validation, or selection among ambiguous destinations.

## Invariants and safety boundaries

- Only repository Markdown sources are rewritten.
- Only the resolved destination path changes.
- One deterministic destination is required.
- Expected source content must still match before replacement.
- Worker concurrency must not change planned output or diagnostic order.

## Scope

Markdown source files and repository-local targets are scanned throughout the Demon Docs repository root, subject to `.docignore` and the permanent traversal exclusions. Nested `.worktrees/` and `.workingtrees/` directories are pruned so attached checkout copies do not enter the repository link graph. A link to an ignored repository path is left outside the link graph. Explicit targets outside the repository are not governed by the repository's `.docignore`.

Local targets may be:

- Markdown files;
- images, PDFs, archives, source files, and other non-Markdown files;
- directories;
- relative paths that resolve outside the repository;
- absolute filesystem paths; or
- `file://` URLs.

Web URLs and other non-filesystem schemes are not part of the local link graph.

Demon Docs only rewrites Markdown source files inside the repository. A target outside the repository can be checked and used as reconciliation evidence, but the external target itself is never modified.

## Supported Markdown Forms

The link scanner handles:

- inline links such as `[Guide](guide.md)`;
- images such as `![Diagram](assets/diagram.png)`;
- angle-wrapped destinations such as `[File](<files/a b.pdf>)`;
- reference definitions such as `[guide]: docs/guide.md`;
- path-based wiki links such as `[[guide]]`, `[[docs/guide|Guide]]`, and `![[assets/diagram.png]]`; and
- local HTML targets in common `href`, `src`, and `poster` attributes.

Extensionless wiki targets resolve as Markdown files. A unique matching Markdown basename elsewhere in the repository is accepted for Obsidian-style wiki links; ambiguous matches are reported and left unchanged. Wiki aliases and embed markers are preserved during repair.

Explicit and collapsed reference uses such as `[Guide][guide]` and `[guide][]` are checked against reference definitions. Missing labels are reported as unresolved links. Shortcut references such as `[guide]` remain untreated because they are indistinguishable from ordinary bracketed prose without a definition.

HTML target coverage includes `a[href]`, `link[href]`, `img[src]`, `script[src]`, `source[src]`, `video[src]`, `video[poster]`, `audio[src]`, and `iframe[src]`.

Link-like text inside fenced code blocks and inline code spans is ignored. Heading fragments and query strings are preserved when a path is rewritten. Heading-anchor existence is not yet validated.

## Persistent State

`.ddocs/` is a private Demon Docs repository, independent of the project's `.git/`. It uses go-git object, tree, reference, and filesystem-storage plumbing internally, but exposes no staging, branch, merge, commit-history, or manual repository workflow.

State is stored as deterministic records for file identities, current paths, Markdown sources and outgoing links, incoming-link groups, fingerprints, and pending generated writes. Record names are distributed across 16 content-addressed shards. A state reference atomically publishes the new root tree after all affected shard objects exist.

A single-file change rewrites only its affected shard or shards; unchanged objects and root entries are reused. The old `.ddocs/files.json` and `.ddocs/links.json` manifests are read only for migration and are removed after the first successful repository-backed publication.

The state is implementation-owned and schema-versioned. Source files are not modified to embed Demon Docs file IDs. When exactly one present file carries a `document_id` that also appears on absent duplicate private records, reconciliation collapses those stale aliases into the live file identity, remaps stored source and target references, and merges historical paths before ordinary candidate discovery.

## First Scan

The first link-enabled `fix` or `watch` pass establishes the baseline state and does not repair links. Existing broken links are reported.

`check -l` remains read-only. When no link-state baseline exists, it reports the missing baseline and exits non-zero. This does not mean `ddocs init` is required; a mutating link-enabled `fix` or `watch` pass establishes the baseline in standalone or initialized mode.

After the baseline exists, later passes can repair links using recorded identity and current filesystem evidence.

## Reconciliation Evidence

Demon Docs prefers deterministic evidence in this order:

1. the previous target file ID still resolves to a present file, including a canonical live identity recovered from an unambiguous `document_id` alias;
2. the target remains at the recorded current path, including a case-only correction;
3. one merged historical path record identifies the new target location;
4. an exact, unique content fingerprint identifies a moved file;
5. a unique filename candidate exists inside the repository; or
6. a bounded search near a missing external target finds a unique candidate.

A unique candidate can be rewritten automatically and recorded as an applied change. Multiple candidates are recorded as `link_repair` suggestions, and the source link remains unchanged until the user selects a candidate.

Relative links remain relative. Absolute filesystem links remain absolute. Link labels, titles, query strings, fragments, angle wrapping, and the source file's newline style are preserved; only the filesystem path is replaced.

## External Edits and Generated Rewrites

User-authored Markdown changes and Demon Docs-generated repairs follow separate paths.

Repository traversal remains serial and deterministic. Files whose path, size, and modification time still match reuse stored fingerprints and `document_id` values. Changed and new regular files are read through a bounded 16-worker pool; Markdown content is read once for both fingerprinting and document-identity extraction, and results merge in traversal order.

For external edits, Demon Docs first identifies every source that cannot reuse stored link records. Those changed sources are read and parsed through a bounded 16-worker pool. Each worker writes only to its assigned source-result slot; results then merge in deterministic source-path order before target resolution, file-identity mutation, diagnostics, review-policy decisions, and repair planning. A content change currently causes a complete source parse; line- or chunk-level incremental parsing is not implemented.

For a known target move, Demon Docs queries stored incoming links by target identity and identifies unchanged affected sources. Each source independently reads its document, calculates exact destination replacements from existing link records, consults the read-only review policy, and constructs a detached generated-rewrite plan through the bounded 16-worker pool. Results remain indexed by deterministic source-path order and merge serially before graph records, diagnostics, updates, and rewrites enter the shared plan. Each generated rewrite records the source file ID, expected old and new content hashes, affected link IDs, and old and new destinations. Successful generated repairs also append an applied-change event to the review ledger.

If stored occurrence offsets no longer match current source text despite unchanged file metadata, Demon Docs abandons that internal fast path and reparses the current source before rebuilding the repair. It does not fail the entire reconciliation or write using stale offsets.

Before writing, every source must still match its expected old hash. Writes use a same-directory temporary file and atomic replacement. The known graph mutation is then published directly. Reparsing the rewritten source is limited to verifying the expected links and refreshing byte offsets, line numbers, and fingerprints.

If a source changed concurrently, the generated rewrite aborts without overwriting the user's content. The next reconciliation processes that source through the external-edit path.

After index, frontmatter, document-format, or reverse-index writes, application orchestration calls scoped link tracking only for Markdown source paths that actually changed. Unselected source records, incoming groups, path history, and pending suppressions are retained. A clean non-link fix skips link tracking entirely and does not initialize absent link state. Explicit link selection still performs complete reconciliation.

Unchanged files reuse stored fingerprints when path, size, and modification time agree. Current benchmarks cover initial indexing, single-file incremental updates, high-fanout target moves, repeated whole-corpus filename renames, scoped post-write refresh, and warmed validation reuse so storage, scanning, planning, and write regressions remain visible.

Known-move source rewrites are prepared through bounded workers and merged deterministically before publication. Generated writes then use a separate bounded worker phase. Each source still receives its own expected-hash check, same-directory temporary file, and atomic replacement. Worker completion order does not change planned output, stored identities, or diagnostic ordering.

Recorded Windows measurements show the 16-worker implementation applying a synthetic 250-source high-fanout move in 322–358 ms, compared with 885–954 ms before bounded parallel writes. A copied Space Rocks stress test repaired 3,717 links across 340 Markdown sources in a median 1.93–1.98 seconds per mass-rename pass. See [Markdown Link Performance](../research/link-performance.md) for methodology, phase timings, throughput, and retained artifacts.

## Commands and Feature Selection

With no selector flags, index and link reconciliation both run:

```bash
ddocs check
ddocs fix
ddocs watch
```

Run only one subsystem with either the short or long selector. `-i` / `--indexes` selects indexes only; `-d` / `--docs` selects indexes, frontmatter, and document-body format together:

```bash
ddocs check -i
ddocs check --indexes
ddocs check -d
ddocs check --docs
ddocs check --frontmatter
ddocs check --format
ddocs check -l
ddocs check --links
ddocs check -r
ddocs check --reverse

ddocs fix -i
ddocs fix -d
ddocs fix --frontmatter
ddocs fix --format
ddocs fix -l
ddocs fix -r

ddocs watch -i
ddocs watch -d
ddocs watch --frontmatter
ddocs watch --format
ddocs watch -l
ddocs watch -r
```

Supplying selectors runs only those systems. Without selectors, configured documentation indexes, frontmatter, document-body format, and link tracking run; link repair follows `[links].enabled`, and reverse indexes also run when reverse roots are configured or supplied.

`check` reports pending rewrites, broken links, ambiguous links, undefined reference labels, and missing baseline state without modifying files. `fix` applies repository-contained source rewrites and saves the resulting state. `watch` uses the same reconciliation path automatically after relevant filesystem events and prints each reconciliation diagnostic rather than only a message count.

When links are enabled, watch mode observes the repository root because moves of non-Markdown targets can require Markdown updates. It also watches the nearest existing parent directories of explicitly linked external targets, so an external rename or removal can trigger the same bounded reconciliation attempt. Documentation-only watch mode remains scoped to the configured docs root. Reverse-only watch mode remains scoped to configured or supplied reverse roots.

## Code map

- `internal/links/` — parsing, target resolution, identity state, diagnostics, generated rewrites, scoped tracking, document-identity alias recovery, and bounded workers.
- `internal/links/internal_move_rewrites.go` — deterministic job selection, bounded per-source known-move rewrite preparation, and ordered result merge.
- `internal/links/wiki_links.go` — path-based wiki links, aliases, embeds, and extensionless Markdown resolution.
- `internal/links/html_links.go` — supported local HTML `href`, `src`, and `poster` targets.
- `internal/links/reference_labels.go` — explicit and collapsed reference-label validation.
- `internal/links/review_suggestions.go` — ambiguous-target suggestion construction.
- `internal/links/review_record.go` — applied repair event recording.
- `internal/links/move.go` and `move_apply.go` — explicit stateless move planning and application.
- `internal/app/app.go` — `check`, `fix`, and foreground `watch` CLI integration.
- `internal/watch/watch.go` — filesystem event scheduling and reconciliation diagnostics.
- `internal/reconcile/reconcile.go` — shared index reconciliation boundary.

## Tests

Focused coverage lives throughout `internal/links/`, including parser, wiki, HTML, reference-label, rewrite, reconciliation, timing, concurrency, and external-fixture tests.

```bash
go test ./internal/links -count=1
```

## Related docs

- [Architecture](INDEX.md)
- [Link Reconciliation State Machine](link-reconciliation-state-machine.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [Review Lifecycles](review-lifecycles.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Repository State and Transactions](repository-state-and-transactions.md)
- [Review Ledger](review-ledger.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Markdown Link Performance](../research/link-performance.md)

## Notes

Heading fragments are preserved during path repair, but heading-anchor existence is not yet part of the implemented validation contract.
