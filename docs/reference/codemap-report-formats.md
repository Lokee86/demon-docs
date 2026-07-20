---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-733e-95bc-725e05c90e68
document_type: general
policy_exempt: false
summary: This document defines the current machine-readable and text contracts for codemap datasets, holdout benchmark reports, precision benchmark files, and precision evaluation output.
---
# Codemap Report Formats

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines the current machine-readable and text contracts for codemap datasets, holdout benchmark reports, precision benchmark files, and precision evaluation output.

## Overview

Codemap commands produce several distinct artifacts. They are not interchangeable:

```text
codemap dataset
  authored map extraction and resolution records

benchmark report
  controlled holdout inputs, classifications, precision, and recall

precision benchmark
  sampled candidates plus human labels and audit metadata

precision evaluation
  validated aggregate and subgroup metrics
```

All current JSON schemas use version 1 but have separate owners and meanings.

## General JSON rules

- JSON output is UTF-8.
- Canonical writers use deterministic ordering and two-space indentation.
- Paths use normalized repository-style slash form where the owning schema defines repository paths.
- Unknown trailing JSON after the root value is rejected by precision loaders.
- A schema version identifies field meaning, not merely file shape.
- Additive optional fields may retain a schema only when older consumers have a safe zero-value interpretation.
- Removed, renamed, newly required, or semantically changed fields require a version decision.

## Codemap dataset

Produced by:

```bash
ddocs codemaps export
```

Current schema version: `1`.

The dataset contains repository/document metadata and extracted entries. Important entry fields include:

```text
document path
matched heading
normalized target
target kind
syntax kind
context
description
source span
raw line
target resolution record
```

Document records include byte size, SHA-256, section count, entry count, and diagnostic count.

Resolution records distinguish exact resolution, pattern matches, missing targets, ambiguity, kind mismatch, outside-repository targets, symbols, and unsupported forms according to the dataset model.

Stable repository inputs and options produce stable JSON.

The dataset is rebuildable analysis input. It is not review history and does not authorize writes.

## Holdout benchmark JSON

Produced by:

```bash
ddocs codemaps benchmark --format json
```

Current schema version: `1`.

The root envelope contains `schema_version` plus the benchmark report.

Current report fields include:

```text
seed
known_links
visible_links
hidden_links
recovered_links
recovered_suggestions
missed_links
unmatched_suggestions
already_linked_suggestions
duplicate_suggestions
invalid_suggestions
raw_suggestion_count
unique_suggestion_count
precision
recall
```

A suggestion contains:

```text
document
target
score
evidence
tier
```

Current tier values are:

```text
hard_link
context
```

An empty tier is accepted as a legacy schema-1 value and normalized to `context` by current evaluation consumers.

Invalid suggestions include their original index, suggestion payload, and reason.

## Holdout benchmark text

Text output is a human review surface. It includes run inputs, metrics, classifications, scores, evidence, and tier where relevant.

Text formatting is deterministic but is not intended as a substitute for schema-versioned JSON in automated consumers. Scripts should consume JSON.

## Precision benchmark JSON

Created by:

```bash
ddocs codemaps precision sample
```

and completed through human review.

Current schema version: `1`.

Root fields:

```text
schema_version
corpus
sampling
suggestions
```

### Corpus metadata

```text
repository
revision
reviewed_at
```

### Sampling metadata

```text
seed
source_report
candidate_count
requested_sample_count
method
```

### Labeled suggestion

A labeled suggestion embeds the benchmark suggestion fields and adds:

```text
rank
area
subsystem
score_bucket
rank_bucket
primary_evidence_kind
evidence_kinds
label
rationale
audit
```

Allowed labels:

```text
valid_missing_link
plausible_but_unnecessary
incorrect
```

An unlabeled sample template may contain empty labels for review. Evaluation requires every label, rationale, and audit reference to be complete.

### Audit metadata

```text
document_section
document_ref
document_excerpt
target_ref
target_excerpt
target_sha256
target_kind
```

Audit metadata binds the judgment to repository content and supports later stale-review detection by humans and tooling.

## Precision evaluation JSON

Produced by:

```bash
ddocs codemaps precision evaluate --format json
```

Current schema version: `1`.

The evaluation contains:

```text
benchmark_size
label_counts
overall
precision_at_1
precision_at_3
precision_at_5
hard_link_sample_valid_recall
hard_link_suggestions_per_document
per_document
by_evidence_kind
by_score_bucket
by_rank_bucket
by_tier
sampling_coverage
```

Precision metric objects include total, valid, accepted non-junk, strict overall precision, and acceptance precision.

Evaluation validates the benchmark against the supplied source report before emitting results.

## Precision text output

Text evaluation is intended for human summaries and review. JSON remains the automation and archival format.

## Canonical ordering

Canonical benchmark output sorts links and classifications by normalized document/target identity and uses stable suggestion ordering. Tier may participate in canonical tie-breaking where the current writer requires it.

Precision samples preserve deterministic candidate ranking and deterministic stratified selection for identical source report and seed.

Consumers must not rely on Go map iteration or incidental file input order.

## Compatibility

Schema 1 currently permits additive optional suggestion metadata such as tier because empty tier has a defined legacy interpretation.

A schema bump is required when:

- an existing field is removed or renamed;
- a field changes meaning;
- an optional field becomes required without a safe default;
- metric formulas change while retaining the same field names;
- label semantics change; or
- classification membership changes incompatibly.

Methodology changes may also require a new retained benchmark lineage even when the JSON shape remains compatible.

## Diagnostics and failure behavior

Commands fail rather than emit partial authoritative JSON when:

- inputs cannot be loaded;
- schema version is unsupported;
- JSON has trailing data or invalid shape;
- benchmark selectors conflict;
- required precision labels or audit fields are incomplete;
- a sampled candidate no longer matches the source report; or
- output cannot be written.

Threshold failure after a completed holdout run is distinct from invalid input or execution failure, as documented in the CLI and diagnostics reference.

## Examples

Generate a deterministic dataset:

```bash
ddocs codemaps export --output codemap-dataset.json
```

Generate a holdout report:

```bash
ddocs codemaps benchmark --seed demon-docs-codemap-benchmark-v1 --format json --output holdout.json
```

Generate current suggestions and a review sample:

```bash
ddocs codemaps precision source --output source-report.json
ddocs codemaps precision sample --suggestions source-report.json --seed review-v1 --output labels.json
```

Evaluate completed labels:

```bash
ddocs codemaps precision evaluate --benchmark labels.json --suggestions source-report.json --format json --output evaluation.json
```

## Related docs

- [CLI Reference](cli.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Codemap Extraction and Dataset](../architecture/codemap-extraction-and-dataset.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

Schema compatibility and methodological comparability are related but distinct. Two reports can share a readable schema while measuring different methods that should not be compared directly.
