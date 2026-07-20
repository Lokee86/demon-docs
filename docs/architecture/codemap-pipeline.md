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

The codemap system is a one-directional deterministic analysis pipeline:

```text
Authored Markdown code maps
-> extraction and versioned dataset
-> normalized repository corpus
-> evidence candidates and fingerprints
-> admission, score, order, and tier
-> current suggestions or controlled holdout
-> precision sampling and evaluation
-> explicit human review and selected insertion
```

The focused canonical owners are:

- [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)

## Permanent safety rule

Demon Docs only suggests potentially missing semantic links.

It never:

- recommends removing an existing authored codemap link;
- labels an existing link irrelevant;
- treats a score or tier as automatic write authorization; or
- rewrites a code map without an explicit selected candidate.

Declined suggestions are persisted through the review ledger and remain suppressed until their evidence materially changes.

## Code root

```text
internal/codemap/
internal/codemapcorpus/
internal/evidence/
internal/codemapbench/
internal/codemapprecision/
internal/app/codemap_*.go
internal/app/review_codemap.go
```

## Responsibilities

The complete pipeline owns:

- deterministic extraction of authored code-map relationships;
- explicit target resolution records;
- normalized repository facts from supported adapters;
- explainable evidence candidates;
- bounded deterministic ranking and tiering;
- controlled answer-isolated holdouts;
- auditable human-labeled precision evaluation;
- schema-versioned report output; and
- conversion of one user-selected candidate into the normal review/change machinery.

## Does not own

It does not own:

- a complete semantic code graph;
- arbitrary language understanding;
- automatic document authorship;
- existing-link removal decisions;
- persisted review policy or undo storage;
- repository-local Markdown link repair; or
- agent-context assembly planned for future work.

## End-to-end flow

### 1. Extraction

`internal/codemap` finds configured map sections, records authored targets and source metadata, and builds a deterministic dataset with explicit resolution outcomes.

See [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md).

### 2. Corpus

`internal/codemapcorpus` combines current repository files, visible authored targets, supported local dependencies, symbols, related documents, and bounded Git history into normalized facts.

See [Codemap Corpus and Adapters](codemap-corpus-adapters.md).

### 3. Evidence and ranking

`internal/evidence` constructs candidates and evidence fingerprints after excluding the document and existing visible targets. `internal/codemapbench` admits, scores, bounds, orders, and tiers them.

See [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md).

### 4. Current suggestions

`SuggestCurrent` runs with all exact authored links visible and returns only proposed additions. Review commands project these suggestions through persisted decline and staleness policy.

### 5. Controlled holdout

Benchmark mode hides a deterministic subset of trusted exact links and removes answer leakage from map text, visible targets, and related-document inputs before generation.

See [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md).

### 6. Precision evaluation

Precision mode builds a deterministic stratified sample of current unmatched suggestions. Human reviewers label and audit each candidate before validated metric aggregation.

See [Codemap Precision Governance](../research/codemap-precision-governance.md).

### 7. Selected insertion

A user-selected codemap candidate is inserted through `internal/codemap.InsertTarget`, converted into a hash-guarded review change, and published through the normal authored-source and review transaction boundaries.

The analysis pipeline never applies a candidate merely because it ranked highly.

## Command surfaces

```text
ddocs codemap export
  build/export authored dataset

ddocs codemap benchmark
  controlled exact-link holdout

ddocs codemap precision source
  generate current unmatched suggestion report

ddocs codemap precision sample
  create deterministic unlabeled review sample

ddocs codemap precision evaluate
  validate completed labels and calculate metrics

ddocs suggestions select
  explicitly apply one reviewed codemap candidate
```

Exact flags, schemas, and exit behavior are owned by the CLI and report-format references.

## State and data ownership

- datasets, corpora, candidates, suggestions, benchmark reports, and evaluations are rebuildable analysis artifacts;
- source reports and labels may be retained as research evidence;
- decline, reconsideration, selection, applied change, block, and undo state belongs to `internal/review` under `refs/ddocs/review`;
- authored codemap source remains human-owned outside the exact selected insertion span.

## Invariants and safety boundaries

- Paths and output ordering are deterministic.
- Ambiguous extraction or resolution is not guessed into truth.
- Existing targets are excluded from missing-link candidates.
- Holdout answers are absent from generator inputs.
- Evidence remains inspectable in reports.
- Output per document is bounded.
- `hard_link` and `context` are descriptive tiers.
- Research metrics are tied to their corpus, revision, method, seed, and labels.
- Only explicit selection can create an authored codemap entry.

## Failure behavior

Each stage fails with its own context rather than silently dropping required inputs. Unsupported facts normally become explicit resolution states or absent evidence; unreadable required files, invalid schemas, inconsistent labels, or output failures abort the command.

A benchmark threshold failure represents a completed measurement below a requested gate, not an execution failure.

## Code map

- `internal/codemap/` — extraction, datasets, stripping, and selected insertion.
- `internal/codemapcorpus/` — repository facts and polyglot adapters.
- `internal/evidence/` — candidate evidence and fingerprints.
- `internal/codemapbench/` — current suggestions, ranking, holdouts, classification, and reports.
- `internal/codemapprecision/` — samples, labels, validation, and evaluation.
- `internal/app/codemap_*.go` — CLI assembly and output.
- `internal/app/review_common.go` and `review_codemap.go` — review projection and selected write integration.

## Tests

The focused package suites are:

```bash
go test ./internal/codemap ./internal/codemapcorpus ./internal/evidence ./internal/codemapbench ./internal/codemapprecision ./internal/app -count=1
```

The [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md) maps the critical extraction, answer-isolation, ranking, report, and precision contracts to their focused tests.

## Related docs

- [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)
- [Review Ledger](review-ledger.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

This overview owns the cross-stage boundary. Detailed constants, supported adapters, methodology, and format contracts belong to the focused documents linked above.
