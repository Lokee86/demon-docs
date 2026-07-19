# Extending Codemap Analysis

Parent index: [Development](./README.md)

## Purpose

This document defines the safe workflow for adding codemap evidence, changing ranking policy, adding language adapters, and evolving benchmark or precision report contracts.

## Overview

Codemap analysis is deterministic but empirical. A change can be internally correct while reducing suggestion quality on real repositories. Implementation tests and pinned evaluation evidence are therefore both required.

The permanent policy remains:

```text
suggest potentially missing links
never suggest that an existing authored link is irrelevant or should be removed
persist declines through the review system
```

## Adding an evidence kind

A new `evidence.Kind` must identify a reproducible repository fact, not an opaque model judgment.

Define:

- normalized source facts required by the collector;
- candidate path normalization;
- source and detail fields used in evidence atoms;
- count semantics;
- fingerprint inputs;
- false-positive controls;
- whether one instance is strong enough for admission; and
- whether broad fan-out should discount it.

Implement the fact collection in the owning corpus adapter when it requires repository analysis. Keep `internal/evidence` focused on candidate construction from normalized inputs.

Add focused positive, negative, ambiguity, exclusion, order-independence, and fingerprint tests.

## Changing ranking or admission

Ranking policy lives in `internal/codemapbench`, not in evidence collection.

A change to any of these is a product-quality change:

- evidence base weight;
- repeated-occurrence factor;
- fan-out discount;
- single-signal admission;
- per-document limit;
- repeated-mention reserve;
- hard-link cap;
- tier eligibility; or
- tie-breaking.

### Required procedure

1. State the failure mode in the current output.
2. Add or update a focused synthetic test that isolates it.
3. Regenerate current source reports for pinned corpora.
4. Evaluate against the pinned labeled benchmark without relabeling changed candidates merely to improve metrics.
5. Run controlled holdouts as a separate recall signal.
6. Compare global and subgroup effects, especially by evidence kind, tier, rank, and repository.
7. Record the result and limitations in research documentation.
8. Update [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md) when implemented constants or admission rules change.

Do not tune only against Demon Docs' own authored codemaps. They are useful for deterministic and portability checks but are not independent quality evidence.

## Adding a dependency adapter

Dependency adapters belong in `internal/codemapcorpus` and emit local `DependencyEdge` facts.

Define:

- recognized source extensions;
- exact import/reference grammar;
- repository or component root rules;
- extension and index-file resolution order;
- treatment of aliases, packages, vendored modules, and external imports;
- whether one reference may resolve to multiple files; and
- explicit unsupported forms.

The adapter must:

- emit repository-relative normalized paths;
- reject external and ambiguous targets rather than guessing;
- deduplicate edges;
- exclude self-edges; and
- produce stable sorted output.

Add fixtures for valid local imports, unsupported/external imports, ambiguous resolution, extension fallbacks, and deterministic ordering. Update [Codemap Corpus and Adapters](../architecture/codemap-corpus-adapters.md).

## Adding symbol extraction

Symbol extraction is a corpus seam distinct from dependency extraction.

A new language extractor must define which declaration kinds are stable enough to expose, how qualified names are formed, and how ambiguous duplicate declarations are handled. Generic or common local names should not become unique symbol evidence merely because parsing found them.

Add declaration, filtering, qualification, and ambiguity tests before enabling the symbols as evidence.

## Changing report schemas

Machine-readable reports are public development contracts.

For additive optional fields:

- retain the current schema only when existing field meanings and required fields do not change;
- give old readers a safe zero-value interpretation; and
- update canonical serialization tests.

For removal, rename, required-field change, or changed meaning:

- increment the schema version;
- update loaders to reject or migrate old data explicitly;
- update [Codemap Report Formats](../reference/codemap-report-formats.md); and
- retain old fixture tests when compatibility is supported.

Canonical output must remain deterministically ordered and indented as documented.

## Benchmark governance

A benchmark change must preserve answer isolation:

- hidden authored links are removed from map text;
- hidden links are absent from visible-target seeds;
- related-document facts expose only visible links;
- generators never receive the hidden answer list; and
- trusted-link inputs are pinned and reviewed.

Changing holdout selection, classification, or metric formulas requires a methodology update and invalidates direct comparison with reports produced under the previous method unless the difference is explicitly normalized.

## Precision governance

Do not edit labels in response to a ranking change unless the old audit is factually wrong. A candidate's label is based on the document and target relationship, not whether the current algorithm ranks it conveniently.

New samples must retain:

- repository and revision;
- source report;
- deterministic seed;
- sampling method;
- document and target audit references;
- rationale; and
- complete labels before evaluation.

See [Codemap Precision Governance](../research/codemap-precision-governance.md).

## Commands

Focused implementation tests:

```bash
go test ./internal/codemap ./internal/codemapcorpus ./internal/evidence ./internal/codemapbench ./internal/codemapprecision -count=1
```

Then run the relevant pinned source, holdout, sample, and evaluation commands described in the codemap guide and research pages, followed by:

```bash
go test ./... -count=1
go vet ./...
```

## Failure modes

- Evidence collector reads hidden benchmark answers through a side channel.
- A broad commit or directory signal dominates because fan-out is not discounted.
- A new adapter treats package imports as local files without proof.
- Ranking changes improve one headline number while degrading hard-link precision or another repository.
- Report ordering follows map iteration.
- Schema meaning changes without a version bump.
- Labels are rewritten to fit the current algorithm.
- Existing authored links enter the candidate set.

## Code map

- `internal/codemap/` — extraction, datasets, map stripping, and selected insertion.
- `internal/codemapcorpus/` — repository facts, dependency adapters, symbols, history, and related documents.
- `internal/evidence/` — deterministic evidence candidates and fingerprints.
- `internal/codemapbench/` — admission, weights, ranking, tiers, holdouts, and reports.
- `internal/codemapprecision/` — samples, labels, validation, and evaluation.
- `internal/app/codemap_*.go` — command orchestration.
- `research/codemap-precision/` — pinned evaluation artifacts.

## Related docs

- [Safe Extension Procedures](safe-extension-procedures.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Corpus and Adapters](../architecture/codemap-corpus-adapters.md)
- [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Codemap Report Formats](../reference/codemap-report-formats.md)

## Notes

Synthetic unit tests protect deterministic mechanics. They do not substitute for reviewed cross-repository quality evidence when the ranking policy changes.
