---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-70fa-9a38-ce6c368c9a43
document_type: general
policy_exempt: false
summary: This document provides the implemented end-to-end ownership map for authored codemap extraction, repository facts, missing-link evidence, ranking, controlled holdouts, precision evaluation, and review selection.
---
# Codemap Pipeline

Parent index: [Architecture](./README.md)

## Purpose

This document provides the implemented end-to-end ownership map for authored codemap extraction, repository facts, missing-link evidence, ranking, controlled holdouts, precision evaluation, and review selection.

## Overview

The codemap system is a deterministic generation and research pipeline:

```text
Existing or schema-required codemap section
-> extraction and versioned dataset
-> normalized repository corpus
-> evidence candidates and fingerprints
-> production admission, score, order, and tier
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

`internal/codemap` finds configured map sections, records authored targets and source metadata, and builds a deterministic dataset with explicit resolution outcomes.

See [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md).

### 2. Corpus

`internal/codemapcorpus` combines current repository files, visible authored targets, supported local dependencies, symbols, related documents, and bounded Git history into normalized facts.

See [Codemap Corpus and Adapters](codemap-corpus-adapters.md).

### 3. Evidence and ranking

`internal/evidence` constructs candidates and evidence fingerprints after excluding the document and existing visible targets. `internal/codemaprecommend` owns production admission, scoring, bounding, ordering, negative-evidence filtering, and tiering. `internal/codemapbench` consumes that package rather than owning a second algorithm.

See [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md).

### 4. Foreground generation

`internal/codemaprun` computes current recommendations with all existing links visible, projects them through persisted decline and staleness policy, and reconciles the complete codemap section. Existing sections are processed regardless of schema. The internal file-type schema placement seam can create a required missing section, but the public application does not yet supply a repository schema provider, so current CLI execution skips documents without a configured section.

### 5. Controlled holdout

Benchmark mode hides a deterministic subset of trusted exact links and removes answer leakage from map text, visible targets, and related-document inputs before generation.

See [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md).

### 6. Precision evaluation

Precision mode builds a deterministic stratified sample of current unmatched suggestions. Human reviewers label and audit each candidate before validated metric aggregation.

See [Codemap Precision Governance](../research/codemap-precision-governance.md).

### 7. Unified reconciliation

The codemap section is adopted under codemap-specific managed markers. Existing syntax is preserved where possible: fenced Space Rocks-style path lists remain fenced, and bullet maps retain their bullet prefix. Qualified missing links from both tiers are added automatically after shared decline filtering. Existing links are retained unless an explicit removal policy applies. Writes use the shared content-addressed transactional file layer. The detailed scope, adoption, rendering, pruning, transaction, and failure lifecycle is owned by [Codemap Managed Execution](codemap-managed-execution.md).

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
- source reports and labels may be retained as research evidence;
- decline and reconsideration state belongs to `internal/review` under `refs/ddocs/review`;
- Demon Docs owns the complete recognized codemap section while preserving existing valid links by default.

## Invariants and safety boundaries

- Paths and output ordering are deterministic.
- Ambiguous extraction or resolution is not guessed into truth.
- Existing targets are excluded from missing-link candidates.
- Holdout answers are absent from generator inputs.
- Evidence remains inspectable in reports.
- Output per document is bounded.
- `hard_link` and `context` remain deterministic generation tiers.
- Existing links are retained unless configured removal policy applies.
- Declined unchanged additions remain suppressed through evidence fingerprints.
- Research metrics are tied to their corpus, revision, method, seed, and labels.
- The daemon and watcher never invoke codemap execution.

## Failure behavior

Each stage fails with its own context rather than silently dropping required inputs. Unsupported facts normally become explicit resolution states or absent evidence; unreadable required files, multiple matching codemap sections, malformed ownership markers, invalid schema placements, concurrent source changes, inconsistent labels, or output failures abort the relevant command.

A benchmark threshold failure represents a completed measurement below a requested gate, not an execution failure.

## Code map

- `internal/codemap/` — extraction, datasets, managed-section adoption, schema placement seam, and syntax-preserving rendering.
- `internal/codemapcorpus/` — repository facts and polyglot adapters.
- `internal/evidence/` — candidate evidence and fingerprints.
- `internal/codemaprecommend/` — production ranking, filtering, ordering, and tiers.
- `internal/codemaprun/` — foreground planning, decline filtering, removal policy, and transactional rewrites.
- `internal/codemapbench/` — holdouts, classification, and reports using the production ranker.
- `internal/codemapprecision/` — samples, labels, validation, and evaluation.
- `internal/app/codemap_*.go` — explicit CLI assembly and output.
- `internal/review/` — shared decline and reconsideration policy.

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
