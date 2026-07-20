---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-752b-b343-5d1994112f29
document_type: general
policy_exempt: false
summary: This document records retained codemap evidence findings, measured baselines, corpus provenance, and the limits on interpreting production missing-link generation quality.
---
# Codemap Missing-Link Evidence

Parent index: [Research](./README.md)

## Purpose

This document records retained codemap evidence findings, measured baselines, corpus provenance, and the limits on interpreting production missing-link generation quality.

## Overview

Demon Docs has implemented a deterministic missing-link analysis pipeline and an explicit foreground managed-section writer. This page owns research evidence and interpretation, not the exact production command or implementation contract.

The missing-link ranker returns targets absent from the current codemap. Production execution automatically adds selected non-declined candidates from both confidence tiers. Existing links remain by default; optional confidence-based pruning belongs to a separate execution policy and is disabled by default.

Current implementation owners are:

- [Codemap Missing-Link Algorithm](../codemap-suggestion-algorithm.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Codemap Benchmark Methodology](codemap-benchmark-methodology.md)
- [Codemap Precision Governance](codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)

## Research status

Implemented research tooling with a production-owned scoring model. Recorded metrics describe pinned labeled samples and must not be interpreted as universal repository performance.

The ranker moved from `internal/codemapbench` into `internal/codemaprecommend`. Benchmarks now import the same implementation used by production execution.

## Research question

Can deterministic repository facts identify missing documentation-to-code links with enough useful relevance to support explicit managed codemap generation, while preserving conservative existing-link retention and auditable decline policy?

This question has separate components:

```text
recovery
= can the model rediscover known authored links when they are hidden?

strict precision
= how often is a surfaced candidate a valid missing permanent link?

relevance
= how often is a surfaced candidate useful non-junk context even when not necessary as a permanent link?

operational safety
= can generation remain deterministic, bounded, inspectable, suppressible, idempotent, and non-daemonized?
```

## Method

### Export current codemaps

```bash
ddocs codemaps export --output .ddocs/codemap.json
```

The export records:

- document paths and content hashes;
- configured codemap headings;
- normalized target forms;
- descriptions and source locations;
- target-resolution diagnostics; and
- stable ordering suitable for fixtures and analysis.

### Run controlled holdouts

```bash
ddocs codemaps benchmark --repo /path/to/repository --format json
```

The benchmark hides a deterministic subset of trusted exact links and removes answer leakage from visible map text, current targets, and related-document facts before asking the production ranker to recover them.

Holdout recovery measures rediscovery. It does not prove that every authored link is necessary or every unheld recommendation is correct.

### Generate and evaluate precision samples

```bash
ddocs codemaps precision --help
```

Review labels distinguish:

- `valid_missing_link`;
- `plausible_but_unnecessary`; and
- `incorrect`.

This separates strict permanent-link precision from broader retained relevance.

## Evidence inputs

The deterministic evidence population includes:

- exact repository-relative path mentions;
- unique basename mentions;
- declared-symbol mentions;
- direct siblings of current targets;
- source/test counterparts;
- direct observed dependency neighbours;
- document/code and current-target/code Git co-change counts; and
- current targets shared by related documents.

Each candidate retains evidence kind, source, detail, count, and a deterministic fingerprint.

Production execution strips the current codemap section from the document text before mention evidence is collected. This prevents a current map entry from becoming evidence for itself.

## Confidence tiers

The production ranker retains a bounded relationship set:

- `hard_link` is the stronger-confidence tier, capped at five candidates per document;
- `context` is the broader weaker or indirect set that still passed admission and negative-evidence filtering.

Both tiers are eligible for automatic addition by explicit codemap execution after persisted decline filtering.

The tier remains useful for:

- human inspection;
- subgroup precision measurement;
- review-policy explanation;
- compatibility reports; and
- optional `remove_low_score_links` evaluation for existing targets.

A context candidate is not a failed hard link. It may be plausible and useful while still unnecessary as a permanent relationship.

## Current measured baseline

### Space Rocks labeled sample

The manually reviewed Space Rocks sample contains 150 recommendations across 25 documents.

| Metric | Current result |
|---|---:|
| Hard-link recommendations | 68 |
| Hard-link strict precision | 75.00% (51/68) |
| Hard-link relevance | 98.53% (67/68) |
| Labeled-valid hard-link recovery | 72.86% (51/70) |
| Context recommendations | 82 |
| Full source pool | 4,493 |
| Full-pool hard links | 621 |
| Full-pool context links | 3,872 |
| Canonical hidden-link holdout | 10/10 recovered |

### Cross-repository recovery

The ordinary calculation corpus contains five repositories. A monolithic per-file index is retained separately as a stress case.

| Metric | Current result |
|---|---:|
| Hidden links | 18 |
| Recovered links | 11 |
| Hard recoveries | 4 |
| Context recoveries | 7 |
| Recall | 61.11% |
| Separate index-stress recovery | 3/10, all context |

### Frozen cross-repository precision review

The fixed sample contains 121 unmatched recommendations.

Before the final incidental-target pass:

| Label | Count |
|---|---:|
| Valid missing link | 83 |
| Plausible but unnecessary | 34 |
| Incorrect | 4 |
| Strict precision | 68.60% |
| Relevance | 96.69% |

After the pass:

- all 83 valid recommendations remain;
- all 34 plausible context recommendations remain;
- all four demonstrated incorrect recommendations are suppressed;
- retained strict precision is 70.94%; and
- retained relevance is 100% for this fixed sample.

The 100% result is not a universal guarantee. It states only that the four reviewed errors in that frozen sample were removed without losing another reviewed useful candidate.

## Interpretation

The measurements support these conclusions:

- declared symbols, supported counterparts, and dependency evidence remain the strongest strict-link signals;
- exact path mentions often indicate relevant context but do not independently prove permanent codemap membership;
- broader context retention improves relevance coverage but increases plausible-unnecessary output;
- narrow negative-evidence rules can remove demonstrated noise without broad file-class suppression;
- the current model is suitable for explicit dogfooding with inspection, dry-run, and persisted declines; and
- repositories with different languages, naming conventions, or document shapes require independent evaluation.

The measurements do not support:

- universal precision claims;
- daemon-triggered unattended generation;
- default removal of links the algorithm cannot reconstruct;
- treating self-authored Demon Docs links as an unbiased benchmark; or
- assuming context-tier additions are always necessary permanent links.

## Decision persistence

The shared review lifecycle records decisions by document, target, relationship kind, and evidence fingerprint.

Production behavior is:

1. generate a current recommendation and fingerprint;
2. replay persisted decline policy;
3. suppress unchanged declined relationships;
4. automatically pass remaining recommendations to managed-section reconciliation;
5. allow materially changed evidence to produce a new current fingerprint; and
6. expose decline and reconsideration through `ddocs suggestions`.

A decline suppresses a future addition. It does not remove an already-present codemap entry. Manual deletion does not create a decline.

## Operational evidence

Production execution adds safety requirements beyond ranking quality:

- one explicit file or directory scope;
- configured heading recognition;
- complete-section managed ownership;
- deterministic syntax-preserving rendering;
- default existing-link retention;
- read-only inspect, check, and dry-run modes;
- source hash preflight;
- atomic replacement and guarded rollback;
- idempotent second execution; and
- no call path from normal reconciliation, watch, or the repository demon.

These contracts are documented in [Codemap Managed Execution](../architecture/codemap-managed-execution.md) and protected in the [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md).

## Limitations

- Quality remains repository- and sample-dependent.
- The cross-repository precision review has one reviewer.
- Most non-Space-Rocks evaluation documents are broad repository guidance rather than narrowly scoped feature documents.
- Few unmatched hard-tier recommendations were available outside Space Rocks.
- Ordinary cross-repository holdout recovery remains 11/18.
- Thresholds are empirical defaults rather than universal constants.
- Production currently auto-adds both selected tiers after decline filtering.
- Production execution now creates missing codemap sections only through selected effective document schemas; schema placement is separate from ranking quality.
- Continued tuning on the same frozen errors risks overfitting.

## Retained artifacts

- `research/codemap-audit/` — initial dataset-quality audit.
- `research/codemap-inventory/` — syntax fixtures and extraction findings.
- `research/codemap-review/` — reviewed trusted-link artifacts.
- `research/codemap-precision/` — Space Rocks labeled sample and evaluation.
- `research/codemap-evidence-validation/` — signal validation work.
- `research/cross-repo-codemap-benchmark/` — pinned multi-repository holdout corpus.
- `research/cross-repo-codemap-precision-review/` — frozen cross-repository manual sample and tuning summary.

## Code map

- `internal/codemap/` — extraction, target normalization, datasets, authored-section stripping, and managed reconciliation.
- `internal/codemapcorpus/` — repository paths, dependencies, symbols, related documents, and Git history adapters.
- `internal/evidence/` — deterministic evidence collection and fingerprints.
- `internal/codemaprecommend/` — production admission, ranking, negative evidence, bounds, and tiers.
- `internal/codemaprun/` — production decline filtering, pruning evaluation, and rewrite planning.
- `internal/codemapbench/` — holdout orchestration and reports using the production ranker.
- `internal/codemapprecision/` — curated-label sampling and metric aggregation.
- `internal/review/` — persisted decision replay.
- `internal/app/codemap_execute*.go` — production fix/check/inspect integration.

## Related docs

- [Research](README.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Codemap Missing-Link Algorithm](../codemap-suggestion-algorithm.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Codemap Benchmark Methodology](codemap-benchmark-methodology.md)
- [Codemap Precision Governance](codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)
- [Current Product Limitations](../limits/current-limitations.md)

## Notes

Research results govern confidence and future tuning. They do not override the production safety contract: explicit foreground execution, persisted declines, deterministic output, and conservative existing-link retention by default.
