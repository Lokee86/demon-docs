# Codemap Precision Governance

Parent index: [Research](./README.md)

## Purpose

This document defines how codemap suggestions are sampled, manually audited, labeled, evaluated, retained, and used to govern ranking changes.

## Overview

Controlled holdouts measure recovery of known links. Curated precision review asks a different question: are current unmatched suggestions actually useful missing links?

```text
current source report
-> deterministic candidate decoration and ranking
-> stratified sample
-> human document/target audit
-> complete labels and rationales
-> source-report consistency validation
-> aggregate and subgroup metrics
```

The process is designed to resist tuning labels to the current algorithm.

## Research status

Implemented with schema-1 benchmark and evaluation models. Pinned labeled corpora are used as regression evidence. Cross-repository expansion remains ongoing research rather than a universal product guarantee.

## Question

Among current unmatched suggestions, what proportion are:

- valid missing links;
- plausible but unnecessary relationships; or
- incorrect suggestions?

How does quality vary by rank, evidence kind, score bucket, tier, document, subsystem, and repository?

## Source report

Precision sampling begins from a current suggestion report produced without holdouts. All exact authored links are visible, so unmatched candidates represent proposed additions rather than intentionally hidden answers.

Candidates are ranked within each document by score and target. Recovered trusted links are excluded from the precision candidate pool.

The source report is part of the audit record. Evaluation later verifies that sampled score, evidence, tier, and rank still match it.

## Deterministic sampling

Sampling requires a seed and records:

```text
source report
candidate count
requested sample count
method
```

The current sampler balances available candidates across dimensions such as:

- repository area;
- subsystem;
- score bucket;
- rank bucket;
- primary evidence kind; and
- representative top-ranked candidates.

It preserves complete top-five sets where available and fills remaining capacity deterministically. Identical inputs and seed produce the same sample.

Sampling aims for diagnostic coverage, not simple random population estimates. Metric interpretation must retain the sampling method.

## Labels

Allowed labels are:

### `valid_missing_link`

The target is materially within the document's ownership or explanatory scope and adding the link would improve the authored code map.

### `plausible_but_unnecessary`

The relationship is real or understandable, but the document does not need the target in its permanent code map. It may be contextual, transitive, incidental, or too broad.

### `incorrect`

The target is not meaningfully within the document's scope, is based on misleading evidence, or would make the map worse.

The strict precision metric counts only `valid_missing_link`. Non-junk acceptance counts both valid and plausible labels.

## Audit requirements

Every labeled suggestion requires:

- rationale;
- document section/reference;
- document excerpt;
- target reference;
- target excerpt;
- target SHA-256; and
- target kind.

The reviewer must inspect both the document and target. Filename similarity or evidence strings alone are insufficient.

The audit should answer whether the permanent authored relationship is useful, not merely whether the files interact somewhere.

## Label stability

Do not change a label because a ranking change moves the candidate or because a desired metric declined.

Relabel only when:

- the original audit was factually wrong;
- the repository content materially changed;
- the document's ownership changed; or
- the label definition changed through an explicit methodology revision.

Retain the reason and date for substantive relabeling.

## Evaluation validation

Before computing metrics, evaluation verifies:

- schema version;
- complete valid labels;
- non-empty rationales and audit references;
- no trailing or malformed JSON;
- known tier values;
- candidate presence in the source report; and
- exact score, evidence, tier, and rank consistency.

This prevents evaluating stale labels against a different report while presenting the result as one run.

## Metrics

Current evaluation includes:

- label counts;
- strict overall precision;
- non-junk acceptance;
- precision at ranks 1, 3, and 5;
- per-document top-k metrics;
- metrics by evidence kind;
- score-bucket metrics;
- rank-bucket metrics;
- tier metrics;
- sampling coverage;
- hard-link sample valid recall; and
- hard-link suggestions per document.

No single number is sufficient. A change that improves global precision while collapsing recall, one repository, or an evidence subgroup requires investigation.

## Governance for ranking changes

Before merging a weight, admission, cap, fan-out, or tier change:

1. identify the observed false-positive or false-negative pattern;
2. add focused synthetic tests;
3. regenerate source reports from the same pinned revisions;
4. evaluate existing labels against unchanged audited candidate identities where possible;
5. identify candidates added, removed, or reordered;
6. inspect strict precision, non-junk acceptance, top-k, tier, and subgroup changes;
7. run controlled holdouts separately for known-link recall;
8. test at least one independent repository when claiming portability; and
9. retain a written interpretation and known tradeoffs.

A tuning pass should be rejected when it gains one metric by violating the permanent safety policy or producing obvious new high-ranked junk.

## Cross-repository evidence

Repository-specific documentation conventions affect apparent precision. Report results separately and combined.

A combined score must not hide that one repository improved while another degraded. Differences in language adapters, history depth, code-map style, and document granularity are material context.

## Results interpretation

Strict precision estimates how often surfaced candidates are judged worthy of a permanent authored link under the sample and label rules.

Non-junk acceptance indicates how often the analysis finds a real relationship even when it is not necessary in the permanent map. That can inform future bounded context use, but it must not be used to inflate permanent-link claims.

`hard_link` quality is the key permanent-link surface. `context` quality is informative but has a different product interpretation.

## Limitations

- Human labels contain judgment.
- Stratified samples are not simple random samples.
- Existing labels cover only candidates present in their source report.
- Repository revisions can invalidate excerpts and hashes.
- Small evidence subgroups have unstable rates.
- Auditors familiar with the algorithm may introduce expectation bias.
- Authored documentation style affects what counts as necessary.

## Retained artifacts

Retain together:

```text
source suggestion report
unlabeled sample template
completed labels and audit metadata
evaluation output
repository revision
seed and sampling method
interpretation or tuning summary
```

Pinned artifacts currently live primarily under `research/codemap-precision/` and cross-repository benchmark folders.

## Verification

```bash
go test ./internal/codemapprecision ./internal/codemapbench ./internal/app -count=1
```

Tests cover candidate decoration, deterministic stratification, complete-label validation, schema and trailing-data refusal, source-report consistency, top-k metrics, tier metrics, and subgroup breakdowns.

## Related docs

- [Codemap Benchmark Methodology](codemap-benchmark-methodology.md)
- [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)
- [Codemap Missing-Link Evidence](codemap-evidence.md)

## Notes

Labels govern evaluation, not product behavior. A valid label still requires explicit user selection before Demon Docs writes an authored codemap link.
