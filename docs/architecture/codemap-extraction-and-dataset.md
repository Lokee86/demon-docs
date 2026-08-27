---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-7b3e-b0d4-2dfff823b601
document_type: general
policy_exempt: false
summary: This document describes how Demon Docs recognizes authored Markdown code maps, normalizes their targets, resolves those targets against repository scope, and exports the deterministic schema-1 dataset consumed by later analysis stages.
---
# Codemap Extraction and Dataset

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes how Demon Docs recognizes authored Markdown code maps, normalizes their targets, resolves those targets against repository scope, and exports the deterministic schema-1 dataset consumed by later analysis stages.

## Overview

Extraction converts authored documentation into explicit records without deciding whether the code map is complete or semantically correct.

```text
configured Markdown section
-> authored entry extraction
-> lexical target classification
-> repository/document-relative resolution
-> optional Arcana file/symbol verification against one pinned snapshot
-> explicit resolution state
-> stable dataset document and entry records
```

The dataset is the boundary between authored code-map syntax and repository analysis. Corpus, evidence, ranking, holdouts, and precision evaluation consume it but do not reinterpret the original Markdown syntax.

## Code root

```text
internal/codemap/model.go
internal/codemap/extractor.go
internal/codemap/dataset.go
internal/codemap/dataset_build.go
internal/codemap/dataset_workers.go
internal/codemap/target_content_cache.go
internal/codemap/semantic_resolver.go
internal/codemap/semantic_target_resolution.go
internal/codemaparcana/
internal/codemap/strip.go
internal/codemap/insert.go
internal/app/app.go
```

## Responsibilities

This boundary owns:

- locating configured code-map sections in Markdown;
- stopping extraction at the next peer or higher-level heading;
- accepting the implemented bullet, fenced, and legacy target forms;
- preserving authored context, description, source span, syntax kind, and raw line;
- normalizing repository-style path separators and directory suffixes;
- classifying targets as file, directory, glob, symbol, or unknown;
- resolving targets against the selected base and optional target roots;
- resolving file-backed and standalone symbol targets through a deterministic `TargetResolver` when one is safely available;
- retaining unresolved, ambiguous, pattern, kind-mismatch, outside-repository, unverified-symbol, and unsupported outcomes explicitly;
- hashing source documents and resolved files;
- stable dataset ordering and schema-1 JSON export;
- stripping authored map sections from benchmark document text; and
- inserting one already-selected target into an existing configured map section.

## Does not own

It does not own:

- repository dependency, history, symbol, or related-document facts;
- missing-link evidence;
- suggestion score, admission, or tier;
- review decisions;
- automatic target selection;
- semantic validation that an authored relationship is necessary; or
- general Markdown link reconciliation.

## Extraction flow

`Extract` parses Markdown using the configured section headings. Defaults are:

```text
Code map
Codemap
Code or source map
Code and test map
```

Heading comparison is normalized for matching. A repository may replace the heading set through configuration or command options.

Within a matching section, the extractor recognizes the current authored forms:

- Markdown bullets whose first code span is the target;
- fenced lines with arrow, equals, leading-path, or indented descriptions;
- indented legacy targets with inline or following descriptions; and
- nested headings or colon-terminated labels used as entry context.

Extraction stops at the next heading at the same or higher level. Fenced content outside a configured code-map section is ignored. Prose-only bullets and TODO-only sections do not become entries.

## Source record

Each `codemap.Entry` retains:

```text
document path
matched map heading
normalized target
target kind
syntax kind
nearest context
description
one-based byte source span
raw authored line
```

Source columns are UTF-8 byte positions rather than grapheme indexes. Consumers that display source positions must preserve that interpretation.

The syntax kind is descriptive. Demon Docs does not rewrite all authored code maps into one syntax.

## Target normalization

Normalization:

- trims surrounding whitespace;
- converts backslashes to slashes;
- cleans redundant path segments;
- retains a trailing slash for directory targets; and
- preserves symbols, globs, and templates as explicit target kinds or unsupported resolution states rather than guessing them into files.

The lexical target kind is separate from resolution. A target that looks like a file may still resolve as missing, ambiguous, or outside repository scope.

## Resolution model

A `Format` selects:

- repository-relative or document-relative primary resolution; and
- zero or more repository-relative target roots.

Target roots support component-relative code maps without inferring a component from the document's location.

Resolution records keep important outcomes distinct, including:

```text
resolved exact target
resolved Arcana symbol
resolved pattern
missing exact or semantic target
missing pattern
ambiguous target roots or semantic target
kind mismatch
outside repository
symbol not verified because semantic resolution is unavailable/stale
unsupported target
```

Demon Docs does not choose among ambiguous roots or coerce a directory into a file target. Pattern families are not later treated as one exact benchmark answer.

## Arcana file and symbol resolution

`BuildDatasetContext` accepts a narrow `TargetResolver`. The current production, export, benchmark, and precision paths attempt to open the current Arcana snapshot through `internal/codemaparcana`; `BuildDataset` remains the compatibility path with no semantic resolver.

The Arcana integration uses the versioned `arcana.query.v1` JSONL process protocol and only the `resolve_file` and `resolve_symbol` operations in this phase. Demon Docs never reads Arcana's packed graph storage directly.

Before a resolver is trusted, `.arcana/CURRENT` must equal `.lexicon/CURRENT`, the Arcana snapshot must be bound to that same Lexicon snapshot, and the content-addressed Lexicon snapshot manifest must verify against its published ID. Path-qualified symbol/file queries are used only when the current file SHA-256 still matches the Lexicon manifest content ID. Standalone/global symbols are more conservative: they require the Lexicon preparation Git head to equal the current clean repository head. If those checks cannot establish freshness, semantic resolution is not attempted.

Resolution policy is explicit:

```text
path#symbol / path::symbol
  Arcana current + one exact node -> resolved
  Arcana current + zero nodes     -> missing
  Arcana current + many nodes     -> ambiguous
  Arcana unavailable/stale        -> symbol_unverified

symbol:Name / standalone symbol
  Arcana current + one exact node -> resolved backing path
  Arcana current + zero nodes     -> missing
  Arcana current + many nodes     -> ambiguous
  Arcana unavailable/stale        -> unsupported
```

Resolved semantic records retain Arcana's durable external node identity, node kind, repository path, qualified name, and exact source span when present. Plain file existence remains filesystem truth; Arcana file resolution adds semantic identity but does not turn an existing unsupported-language file into a missing file.

An Arcana executable is discovered next to the running Demon Docs binary or on `PATH`; `DDOCS_ARCANA_COMMAND` is available as a development/host override. Missing or structurally stale Arcana state degrades to the explicit fallback states above. Once a current Arcana protocol session has been opened, query/protocol failures abort the operation rather than silently mixing partial semantic truth with fallback results.

Relationship expansion is not part of dataset construction. Step 5 consumes the same verified Arcana session through the separate corpus `RelationshipProvider`, so target resolution remains an authored-target concern while graph-neighborhood evidence remains per-document and benchmark-safe.

## Dataset construction

`BuildDatasetContext` first walks Markdown documents below the selected docs root under the repository ignore policy. `BuildDataset` is the background-context, no-semantic-resolver compatibility wrapper. Traversal, ignore evaluation, and document discovery remain serial and deterministic. Discovered document jobs are then sorted by repository-relative path.

A bounded 16-worker pool independently reads, hashes, extracts, and resolves each document. Workers store complete detached results in their assigned job slots. The main goroutine merges documents, entries, diagnostics, and errors in deterministic document-path order before the final canonical sort.

Resolved target files use one concurrent per-build content-hash cache. The first resolver for a target reads and hashes it; other document workers wait for and reuse that result. Different targets can still be read concurrently. The cache is rebuildable, lasts only for one dataset build, and does not become persistent repository state.

It skips:

- non-Markdown files;
- ignored paths;
- directories as document inputs; and
- symlinked files.

Each document record includes path, byte size, SHA-256, section count, entry count, and diagnostic count. Each extracted entry includes its normalized target plus a `TargetRecord` containing resolution state, resolved path or pattern matches, existence, file size, and file hash when applicable. When Arcana safely resolves a file or symbol, the record also carries optional semantic node identity/span metadata; ambiguous semantic candidates remain explicit.

Documents and entries are emitted in deterministic path and source order. Worker completion order cannot affect exported ordering or which error is returned first.

### Dataset construction performance

A retained Windows benchmark builds 500 Markdown documents containing 4,000 total codemap entries that repeatedly reference 64 target files. The comparison used `GOMAXPROCS=16`, five one-iteration runs, and commit `e6917a5` as the serial baseline. Excluding the first sample to reduce host warm-up noise, the mean dataset-build time improved from 1.750 seconds to 198.1 milliseconds, an 8.83x speedup.

The improvement combines parallel document preparation with hashing each repeatedly referenced target only once per build. It does not parallelize directory traversal or alter target-resolution outcomes. The retained benchmark is `BenchmarkBuildDataset` in `internal/codemap/dataset_benchmark_test.go`.

```bash
go test ./internal/codemap -run '^$' -bench '^BenchmarkBuildDataset$' -benchmem -count=5 -benchtime=1x
```

## Dataset serialization

The exported dataset uses schema version 1 and canonical indented JSON. Stable input produces byte-stable output.

Additive optional metadata may be introduced only when old consumers have a safe zero-value interpretation. Renamed fields, removed fields, newly required fields, or changed meanings require a schema decision documented in [Codemap Report Formats](../reference/codemap-report-formats.md).

## Benchmark stripping

`StripAuthoredSections` removes configured map sections while retaining other document text and line structure. It uses the same section-heading rules as extraction and ignores heading-like text inside fences.

The benchmark engine must use stripped text before collecting mention evidence. This prevents a hidden authored target from being recovered by reading the answer directly from the document's map.

Stripping is an analysis transformation; it does not modify the repository file.

## Selected insertion

`InsertTarget` is a narrow write primitive used only after a user selects a codemap suggestion.

It:

- finds the first configured code-map section;
- rejects a target already authored in that section;
- appends one canonical bullet;
- returns inserted byte offsets and text for review/change recording; and
- leaves candidate selection and policy to the review/app layers.

It does not create a missing code-map section and does not apply a ranked candidate automatically.

## State and data ownership

- `codemap.Result` owns extracted entries and extraction diagnostics for one document.
- `codemap.Dataset` owns versioned repository-wide authored code-map records.
- Source spans refer to the exact source version used to build the dataset.
- Dataset hashes describe input and resolved target content; they are not persistent file identities.
- The dataset may be rebuilt from current repository facts and is not historical review state.

## Invariants and safety boundaries

- Only configured sections are extracted.
- Extraction does not normalize or rewrite authored source.
- Ambiguous and unsupported targets remain explicit.
- Exact and pattern resolution remain different classes.
- Symlinked document inputs are skipped.
- Stable input produces deterministic records and JSON.
- Hidden benchmark targets must be removed from map text before evidence collection.
- Selected insertion rejects duplicates and requires an explicit selected target.

## Failure behavior

Dataset construction fails on unreadable required documents, invalid root scope, target-resolution I/O failures, or output errors. Unsupported authored shapes are skipped or diagnosed rather than guessed into an entry.

A missing or ambiguous target does not normally abort dataset export; it becomes a resolution record so downstream tools can decide whether that record is admissible for their purpose.

Selected insertion fails when no configured section exists, the target is already authored, or the source cannot be transformed safely.

## Code map

- `internal/codemap/model.go` — entry, target, syntax, span, format, and diagnostic types.
- `internal/codemap/extractor.go` — configured-section extraction and normalization.
- `internal/codemap/dataset.go` — target resolution, schema records, canonical sorting, and export.
- `internal/codemap/dataset_build.go` — deterministic discovery, bounded per-document preparation, ordered merge, and dataset assembly.
- `internal/codemap/dataset_workers.go` — fixed 16-worker execution boundary with indexed errors.
- `internal/codemap/target_content_cache.go` — concurrent single-read target hashing shared by one dataset build.
- `internal/codemap/semantic_resolver.go` and `semantic_target_resolution.go` — generic semantic target contract and mapping into dataset resolution states.
- `internal/codemaparcana/` — pinned Arcana/Lexicon state validation, JSONL query transport, and Arcana target resolver.
- `internal/codemap/strip.go` — holdout text sanitization.
- `internal/codemap/insert.go` — explicit selected-target insertion.
- `internal/app/app.go` — `codemap export` scope and output orchestration.

## Tests

Focused coverage includes:

- `extractor_test.go` — supported syntax, configured headings, boundaries, prose rejection, and symbol forms;
- `inventory_fixture_test.go` — extraction against retained authored fixtures;
- `dataset_test.go` — repository/document bases, target roots, ambiguity, templates, hashes, and stable JSON;
- `semantic_target_resolution_test.go` — resolved/missing/ambiguous/unsupported symbol outcomes and optional file identity;
- `internal/codemaparcana/*_test.go` — snapshot alignment, file-content freshness, global-resolution gating, qualified ambiguity, and JSONL framing;
- `dataset_workers_test.go` — bounded concurrency and deterministic indexed errors;
- `target_content_cache_test.go` — concurrent repeated-target reads collapse to one content read;
- `dataset_benchmark_test.go` — retained large deterministic dataset-build benchmark;
- `strip_test.go` — map removal, configured aliases, and fenced-heading exclusion;
- `insert_test.go` — selected insertion and duplicate rejection; and
- `internal/app/codemap_export_test.go` — command options and deterministic output.

```bash
go test ./internal/codemap ./internal/app -run 'Codemap|Extract|Dataset|Strip|Insert' -count=1
```

## Related docs

- [Codemap Pipeline](codemap-pipeline.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

The dataset records what authors wrote and how those targets resolve. It does not assert that the authored map is complete, minimal, or semantically correct.
