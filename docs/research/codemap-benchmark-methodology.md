# Codemap Benchmark Methodology

Parent index: [Research](./README.md)

## Purpose

This document defines the controlled-holdout methodology used to measure whether the codemap suggestion pipeline can recover authored document-to-code relationships without reading the hidden answers.

## Overview

A controlled holdout hides a deterministic subset of trusted authored links and asks the current generator to recover them from remaining repository evidence.

```text
trusted exact links
-> deterministic hidden/visible split
-> remove hidden map text and related-target leakage
-> generate suggestions from visible evidence
-> classify output
-> calculate precision and recall for this run
```

The method measures recovery under a particular repository, revision, trusted-link set, seed, and generator. It is not a universal product guarantee.

## Research status

Implemented and used as a regression and tuning instrument. The methodology remains subject to explicit versioned changes when leakage controls, answer selection, classification, or metrics change.

## Question

Given a set of reviewed document-to-code links, how often does the current deterministic analysis surface a hidden authored relationship, and how much unrelated output does it produce while doing so?

The benchmark does not answer whether every authored link is necessary or whether every unmatched suggestion is wrong.

## Inputs

A run requires:

- repository and revision;
- codemap dataset or reviewed trusted-link JSON;
- deterministic seed;
- exact holdout count or fraction, or the documented default;
- current corpus adapters and generator; and
- output format and optional thresholds.

Only exact resolved links are eligible when using the dataset. Pattern families, unresolved symbols, ambiguous targets, stale paths, and unsupported targets are excluded from exact-link truth.

## Trusted links

A reviewed trusted-link file may replace dataset-derived links. Trusted links are normalized, deduplicated, and sorted.

The trusted set should record its repository and review context outside or alongside the file. A trusted set copied across materially different revisions must be revalidated.

Authored links are treated as benchmark answers for recovery, not as proof that the relationship is still ideal.

## Deterministic holdout selection

The default seed is:

```text
demon-docs-codemap-benchmark-v1
```

Selection hashes the seed with each normalized document-target key. Therefore:

- input order does not change the split;
- identical seed and known links produce the same holdout;
- changing the seed intentionally creates another deterministic sample; and
- reports can be reproduced from retained inputs.

The default hides 20 percent of known links, rounded up. Exact count and fraction selectors are mutually exclusive.

## Answer isolation

The generator must not receive hidden links through any direct or indirect input.

Current controls are:

1. hidden links are absent from `VisibleLinks`;
2. `StripAuthoredSections` removes the document's authored map before evidence collection;
3. existing-target seeds contain only visible links;
4. related-document target facts are sanitized to visible links; and
5. the generator interface receives document names and visible links, not the hidden answer list.

Tests verify both direct request isolation and end-to-end benchmark-engine stripping.

A new corpus fact source must be reviewed for leakage before it is admitted to benchmark generation.

## Method

1. Build or load the codemap dataset.
2. Select exact trusted links.
3. Normalize and split known links into visible and hidden sets.
4. Build the repository corpus.
5. Replace each benchmark document's evidence text with map-stripped text.
6. Sanitize related-document authored targets against the visible set.
7. Generate ranked suggestions.
8. Normalize and classify every suggestion.
9. Produce canonical text or schema-versioned JSON.

## Classification

Generator output is classified as:

- recovered hidden link;
- unmatched suggestion;
- already-linked suggestion;
- duplicate suggestion; or
- invalid suggestion.

Every hidden link not recovered becomes a missed link.

Already-linked output is not useful recovery and is included in the current precision denominator. Duplicate and invalid output are reported separately so generator defects remain visible.

## Metrics

Current benchmark precision is:

```text
recovered hidden links
-----------------------------------------------
recovered + unmatched + already-linked output
```

Recall is:

```text
recovered hidden links
----------------------
all hidden links
```

These metrics evaluate exact-link recovery under the selected holdout. They do not measure plausible-but-unnecessary suggestions, context-tier usefulness, or missing links absent from the trusted set.

## Threshold behavior

The CLI may enforce minimum precision or recall. A completed run that fails a requested threshold returns the command's threshold-failure exit status rather than an execution-error status.

Thresholds are useful in pinned regression jobs only when repository, answers, seed, and method remain stable.

## Corpus guidance

Use multiple repositories when evaluating general portability. Demon Docs' own code maps are useful for extraction, deterministic ordering, and holdout mechanics, but are not independent evidence because the same project shaped both docs and algorithm.

Large or repetitive repositories can expose fan-out and adapter behavior. Small curated sets can expose exact failure modes. Neither alone is representative.

## Results interpretation

A recovered link shows that the current evidence surfaced a relationship already present in the trusted map.

An unmatched suggestion may be:

- a valid missing link not represented in the answer set;
- a plausible but unnecessary relationship; or
- an incorrect suggestion.

Therefore holdout precision is intentionally conservative and must be complemented by curated precision review.

Changes in recall can result from extraction, corpus facts, evidence, ranking, caps, or tier policy. Diagnose stage-level differences rather than changing weights from one aggregate number.

## Limitations

- Authored links are an imperfect proxy for ground truth.
- Hidden links may remain inferable from prose legitimately outside the map.
- Git history and supported adapters vary by repository.
- A small holdout has high variance.
- Exact-link holdouts do not evaluate directory patterns or symbolic relationships.
- Ranking caps can miss valid low-ranked links even when evidence exists.
- Holdout precision penalizes new valid suggestions absent from the trusted set.

## Reproducibility requirements

Retain:

```text
repository and revision
dataset or trusted-link input
seed and holdout selector
command options
generator implementation revision
canonical report
```

When methodology changes, record the old and new method and do not compare scores as though they were produced under one contract.

## Retained artifacts

Current artifacts live under research directories such as:

```text
research/codemap-review/
research/codemap-precision/
research/cross-repo-codemap-benchmark/
```

Repository-specific artifact folders may include trusted links, source reports, holdout reports, labels, evaluation summaries, and helper scripts.

## Verification

Focused mechanics tests:

```bash
go test ./internal/codemap ./internal/codemapbench ./internal/app -count=1
```

Important contracts include deterministic split, order independence, conflicting-selector refusal, answer isolation, generator-error propagation, current-vs-holdout separation, and canonical output.

## Related docs

- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Extraction and Dataset](../architecture/codemap-extraction-and-dataset.md)
- [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md)
- [Codemap Precision Governance](codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)

## Notes

A holdout is a controlled recovery experiment, not a claim that hidden authored links are the only correct documentation relationships.
