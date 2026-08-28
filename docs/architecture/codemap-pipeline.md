---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-70fa-9a38-ce6c368c9a43
document_type: general
policy_exempt: false
summary: This document provides the implemented end-to-end ownership map for authored codemap extraction, repository facts, missing-link evidence, ranking, controlled holdouts, precision evaluation, and review selection.
---
# Codemap Pipeline

Parent index: [Architecture](./INDEX.md)

## Purpose

This document provides the implemented end-to-end ownership map for authored codemap extraction, repository facts, missing-link evidence, ranking, controlled holdouts, precision evaluation, and review selection.

## Overview

The codemap system is a deterministic generation and research pipeline:

```text
Existing or schema-required codemap section
-> extraction and versioned dataset
-> pinned Arcana file/symbol resolution when current
-> normalized repository corpus
-> visible-target-bounded Arcana relationship evidence when current
-> evidence candidates and fingerprints
-> deterministic candidate role
-> production admission and score
-> score-banded role/directory coverage selection
-> conservative hard-link allocation
-> semantic-baseline comparison against prior Arcana snapshot
-> shared decline-policy filtering
-> unified managed-section reconciliation
-> atomic foreground write
```

The same production ranking package also feeds controlled holdouts, precision sampling, and evaluation.

The focused canonical owners are:

- [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md)
- [Codemap Managed Execution](codemap-managed-execution.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)

## Permanent safety rule

Demon Docs owns the complete configured codemap section as one unified managed artifact. Existing and newly generated links are not split into separate authored and generated lists.

By default it never removes an existing valid semantic link merely because the algorithm does not rediscover it or ranks it below the hard-link tier. Projects may explicitly enable undiscovered-link or low-score removal. Declined proposed additions are persisted through the shared review ledger and remain suppressed until their evidence materially changes.

## Code root

```text
internal/codemap/
internal/codemaparcana/
internal/codemapcorpus/
internal/evidence/
internal/codemaprecommend/
internal/codemaprun/
internal/codemapbench/
internal/codemapprecision/
internal/app/codemap_*.go
internal/app/review_codemap.go
```

## Responsibilities

The complete pipeline owns:

- deterministic extraction of existing codemap relationships;
- schema-gated placement of missing codemap sections;
- complete managed-section adoption and rendering;
- explicit target resolution records;
- normalized repository facts from supported adapters;
- explainable evidence candidates;
- bounded deterministic ranking and tiering;
- shared decline-policy filtering;
- optional conservative link removal;
- atomic foreground rewrites;
- controlled answer-isolated holdouts; and
- auditable human-labeled precision evaluation.

## Does not own

It does not own:

- a complete semantic code graph;
- arbitrary language understanding;
- prose authorship or rewriting;
- persisted review policy or undo storage;
- repository-local Markdown link repair; or
- daemon/watch scheduling.

Codemap execution is deliberately absent from ordinary `fix`, `check`, `watch`, and repository-demon reconciliation.

## End-to-end flow

### 1. Extraction

`internal/codemap` finds configured map sections, records authored targets and source metadata, and builds a deterministic dataset with explicit resolution outcomes. Current production/export/benchmark paths also attempt to attach `internal/codemaparcana` as a `TargetResolver`: exact files and symbol targets may then be resolved against one Arcana snapshot pinned to matching Lexicon state. Unavailable or stale Arcana state degrades to explicit unverified/unsupported outcomes rather than guessed semantic truth.

See [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md).

### 2. Corpus

`internal/codemapcorpus` combines current repository files, authored target provenance and concrete coverage, code-intelligence provider facts, related documents, and bounded Git history into normalized facts. File, directory, pattern, and symbol targets retain their authored abstraction; only explicit resolved files become outward shallow-structure/dependency/history seeds. Dependency and declared-symbol facts cross a narrow repository-wide `CodeIntelligenceProvider` seam; the existing shallow language adapters are the default fallback provider.

At per-document input time, an independent `RelationshipProvider` receives only currently visible exact file targets plus verified symbol targets. When Arcana is current, this adds one-hop allowlisted semantic relationships without exposing hidden benchmark holdouts or handing graph traversal ownership to the ranker. These relationships remain a distinct context-only evidence kind for mutation policy, but Step 6 now uses their direction and relation type as deterministic role-classification input.

See [Codemap Corpus and Adapters](codemap-corpus-adapters.md).

### 3. Evidence and ranking

`internal/evidence` constructs candidates and evidence fingerprints after excluding the document and existing visible targets. `internal/codemaprecommend` owns production admission, deterministic candidate-role classification, scoring, coverage-aware selection, bounding, ordering, negative-evidence filtering, and tiering. Roles distinguish primary implementation, supporting implementation, verification/test, interface/boundary, and context-only relationships. Role does not change numeric score; it is used within comparable score bands to preserve semantic/directory coverage and to prevent one role from monopolizing the hard-link surface. `internal/codemapbench` consumes that package rather than owning a second algorithm.

See [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md).

### 4. Semantic staleness

When current Arcana state is available, Demon Docs compares each document against the Arcana snapshot recorded by its last accepted semantic baseline in `.ddocs`. Arcana's deterministic snapshot `diff` identifies mapped nodes whose definition metadata, ownership-like identity, or relationships changed, and detects mapped targets that disappeared or moved. Directory and glob targets are not treated as semantic nodes.

Semantic staleness is analysis only. It does not remove authored links, alter recommendation score, or authorize pruning. `inspect` reports the individual mapped targets and change kinds; `check` fails when a document has semantic-staleness findings even if its managed codemap text would not change.

A successful `codemaps fix` initializes a missing baseline and advances a baseline when mapped semantics are unchanged. If semantic changes exist and the document itself has not changed since the prior baseline, `fix` deliberately leaves the old baseline in place, so rerunning the command cannot silently clear the warning. Editing the document and then running `fix` accepts the current snapshot as its new baseline.

### 5. Foreground generation

`internal/codemaprun` computes current recommendations with all existing links visible, projects them through persisted decline and staleness policy, and reconciles the complete codemap section. Existing sections are processed regardless of schema. The application supplies the document-policy schema provider, so a required missing codemap section is created at its schema-defined position; schemas without one leave the document unchanged.

### 6. Controlled holdout

Benchmark mode hides a deterministic subset of trusted exact links and removes answer leakage from map text, visible targets, and related-document inputs before generation.

See [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md).

### 7. Precision evaluation

Precision mode builds a deterministic stratified sample of current unmatched suggestions. Human reviewers label and audit each candidate before validated metric aggregation.

See [Codemap Precision Governance](../research/codemap-precision-governance.md).

### 8. Unified reconciliation

The codemap section is adopted under codemap-specific managed markers. Existing syntax is preserved where possible: fenced Space Rocks-style path lists remain fenced, and bullet maps retain their bullet prefix. Qualified non-declined `hard_link` recommendations are added automatically; `context` recommendations remain visible to inspection and review without being written. Existing links are retained unless an explicit removal policy applies. Writes use the shared content-addressed transactional file layer. The detailed scope, adoption, rendering, pruning, transaction, and failure lifecycle is owned by [Codemap Managed Execution](codemap-managed-execution.md).

## Command surfaces

```text
ddocs codemaps fix [--root FILE_OR_DIRECTORY] [--dry-run]
  adopt and update unified codemap sections

ddocs codemap ...
  singular compatibility alias for the canonical plural command family

ddocs codemaps check --root FILE_OR_DIRECTORY
  report stale selected codemaps without writing

ddocs codemaps inspect --root FILE_OR_DIRECTORY
  explain recommendations, evidence, declines, and removals

ddocs codemaps export
  build/export authored dataset

ddocs codemaps benchmark
  controlled exact-link holdout

ddocs codemaps precision source
  generate current unmatched suggestion report

ddocs codemaps precision sample
  create deterministic unlabeled review sample

ddocs codemaps precision evaluate
  validate completed labels and calculate metrics

ddocs suggestions decline|reconsider
  manage shared persisted recommendation policy
```

Exact flags, schemas, and exit behavior are owned by the CLI and report-format references.

## State and data ownership

- datasets, corpora, candidates, recommendations, benchmark reports, and evaluations are rebuildable analysis artifacts;
- per-document semantic validation baselines are stored in `.ddocs` and record document digest plus the accepted Arcana snapshot;
- source reports and labels may be retained as research evidence;
- decline and reconsideration state belongs to `internal/review` under `refs/ddocs/review`;
- Demon Docs owns the complete recognized codemap section while preserving existing valid links by default.

## Invariants and safety boundaries

- Paths and output ordering are deterministic.
- Ambiguous extraction or resolution is not guessed into truth.
- Arcana semantic resolution is trusted only when Arcana/Lexicon snapshot alignment and the required source freshness checks succeed.
- Arcana relationship evidence is one-hop, allowlisted, per-document, and seeded only from currently visible exact files or verified symbols; truncated seed neighborhoods are discarded.
- Missing or stale Arcana state degrades explicitly; an opened current protocol session failing mid-query is an error.
- Existing authored coverage is excluded from missing-link candidates, including descendants covered by directory targets and concrete matches covered by patterns.
- Pattern matches and directory descendants do not become independent outward evidence seeds.
- Holdout answers are absent from generator inputs.
- Evidence and deterministic candidate roles remain inspectable in reports.
- Candidate role never changes numeric score; selection uses it only within score bands and hard-link coverage limits.
- Lower score bands cannot displace higher-band candidates in the normal bounded surface.
- No role may consume more than two hard-link slots and no target directory more than three.
- Output per document is bounded.
- `hard_link` and `context` remain deterministic recommendation tiers; only `hard_link` is an automatic generation tier.
- Existing links are retained unless configured removal policy applies.
- Declined unchanged additions remain suppressed through evidence fingerprints.
- Research metrics are tied to their corpus, revision, method, seed, and labels.
- The daemon and watcher never invoke codemap execution.

## Failure behavior

Each stage fails with its own context rather than silently dropping required inputs. Unsupported facts normally become explicit resolution states or absent evidence; unreadable required files, multiple matching codemap sections, malformed ownership markers, invalid schema placements, concurrent source changes, inconsistent labels, or output failures abort the relevant command.

A benchmark threshold failure represents a completed measurement below a requested gate, not an execution failure.

## Code map

<!-- doc-ledger:codemap:start -->

- `internal/codemap/` — extraction, datasets, semantic target-resolution contract, managed-section adoption, schema placement seam, and syntax-preserving rendering.
- `internal/codemaparcana/` — current-snapshot discovery, Arcana JSONL transport, freshness checks, file/symbol resolution, snapshot diff, and mapped semantic-staleness analysis.
- `internal/codemapsemantic/` — durable per-document semantic validation baselines in `.ddocs`.
- `internal/codemapcorpus/` — repository facts and polyglot adapters.
- `internal/evidence/` — candidate evidence and fingerprints.
- `internal/codemaprecommend/` — production role classification, ranking, filtering, ordering, and tiers.
- `internal/codemaprun/` — foreground planning, decline filtering, removal policy, and transactional rewrites.
- `internal/codemapbench/` — holdouts, classification, and reports using the production ranker.
- `internal/codemapprecision/` — samples, labels, validation, and evaluation.
- `internal/app/codemap_*.go` — explicit CLI assembly and output.
- `internal/review/` — shared decline and reconsideration policy.
<!-- doc-ledger:codemap:end -->

## Tests

The focused package suites are:

```bash
go test ./internal/codemap ./internal/codemapcorpus ./internal/evidence ./internal/codemaprecommend ./internal/codemaprun ./internal/codemapbench ./internal/codemapprecision ./internal/app -count=1
```

The [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md) maps the critical extraction, answer-isolation, ranking, report, and precision contracts to their focused tests.

## Related docs

- [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md)
- [Codemap Managed Execution](codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)
- [Review Ledger](review-ledger.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

This overview owns the cross-stage boundary. Detailed constants, supported adapters, methodology, and format contracts belong to the focused documents linked above.
