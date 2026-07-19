# Codemap Evidence and Ranking

Parent index: [Architecture](./README.md)

## Purpose

This document describes how normalized repository facts become deterministic potentially-missing-link candidates, scores, ordered suggestions, and `hard_link` or `context` tiers.

## Overview

Evidence and ranking are separate ownership seams:

```text
Corpus facts + visible authored targets
-> evidence candidates and fingerprints
-> candidate admission
-> weighted scoring and fan-out discount
-> bounded deterministic ordering
-> suggestion tier
```

Evidence explains why a target may be relevant. Ranking decides which candidates are surfaced first. Neither establishes semantic truth or authorizes automatic insertion.

## Code root

```text
internal/evidence/
internal/codemapbench/suggestions.go
internal/codemapbench/adapters.go
internal/codemapbench/current.go
```

## Responsibilities

This boundary owns:

- excluding the current document and visible authored targets;
- collecting implemented mention, structural, dependency, history, related-document, and symbol evidence;
- canonical evidence ordering and candidate fingerprints;
- candidate admission rules;
- evidence base weights;
- repeated-occurrence handling;
- evidence-atom fan-out discounting;
- per-document suggestion bounds and repeated-mention reserve;
- deterministic score and target ordering; and
- assignment of `hard_link` and `context` tiers.

## Does not own

It does not own:

- extraction of authored maps;
- repository fact adapters;
- holdout answer selection;
- human validity labels;
- review decline persistence;
- automatic document mutation; or
- removal or irrelevance judgments for existing links.

## Evidence input

`evidence.Input` contains:

```text
document path and visible text
repository files
visible existing targets
dependency edges
bounded commit facts
related documents and their visible targets
symbol declarations
```

All paths are normalized before candidate creation. The current document and every visible existing target are excluded from the candidate set.

## Current evidence kinds

### Exact path mention

The document contains the repository-relative target path at a token boundary.

This is explicit prose evidence, but repeated occurrences use a fixed occurrence factor so a copied path does not compound score indefinitely.

### Unique basename mention

The document contains a file basename that maps to exactly one repository path. Non-unique basenames are not admitted as this evidence.

### Declared symbol mention

The document mentions a declared symbol that resolves to one repository path. Ambiguous symbols are excluded by corpus construction or evidence collection.

### Sibling of existing target

The candidate shares a structural directory relationship with an existing authored target.

### Test counterpart

The candidate is the source/test or implementation/spec counterpart recognized by current naming and directory rules.

### Dependency neighbor

A dependency edge connects the candidate and an existing target in either direction. The relation and edge source contribute to the evidence atom.

### Git co-change with document

The candidate appears in a bounded commit with the document.

### Git co-change with existing target

The candidate appears in a bounded commit with a visible authored target.

### Related-document target

A locally linked related document already authors the candidate target.

## Evidence aggregation

A candidate contains one or more evidence records:

```text
kind
source
detail
count
```

Equivalent evidence occurrences are aggregated. Evidence and candidates are sorted deterministically. The candidate fingerprint is a SHA-256 derived from the target and canonical evidence fields.

The fingerprint supports review staleness and reproducibility. Changing the evidence set or canonical evidence meaning may intentionally change the fingerprint.

## Candidate admission

A candidate enters ranking when it has:

- at least two distinct evidence kinds; or
- one currently admitted stronger kind.

Current single-kind admission includes:

```text
exact path mention
unique basename mention
declared symbol mention
test counterpart
dependency neighbor
related-document target
```

Sibling or history evidence alone is not sufficient.

Admission is a surfacing policy, not a claim that the candidate is valid.

## Score policy

Current base weights are:

```text
declared symbol mention             7
exact path mention                  6
test counterpart                    6
unique basename mention             4
dependency neighbor                 4
related-document target             4
sibling target                      2
Git target co-change              1.5
Git document co-change              1
```

For most kinds, repeated evidence uses:

```text
1 + log2(occurrence count)
```

Exact path and unique basename mentions use a fixed occurrence factor.

Each evidence atom is divided by the logarithm of its candidate fan-out. A broad commit, directory, or shared source therefore contributes less to each candidate than a specific fact.

## Evidence atom identity

Fan-out is measured by an atom composed from evidence kind, source, and selected detail.

Dependency and declared-symbol evidence retain detail because different dependencies or symbols are meaningfully distinct. Other kinds use the normalized source/kind identity needed by current scoring.

Changing atom identity changes ranking behavior and requires benchmark review.

## Selection bounds

Suggestions are first sorted by descending score and target-path tie-breaker.

Current bounds:

```text
normal suggestions per document: 30
additional repeated exact-path reserve: 2
minimum repeated explicit mentions: 2
```

The reserve may include high-count exact-path candidates outside the normal top 30. The final union is resorted by score and target.

## Tier assignment

All selected suggestions default to `context`.

Only the first five ordered suggestions are eligible for `hard_link`. Eligibility requires:

- declared-symbol evidence;
- test-counterpart evidence; or
- dependency-neighbor evidence with total score at least 16.

The tier distinguishes the current permanent-link review surface from weaker context relationships. It does not alter authored files and does not make a candidate true.

An empty legacy tier in schema-1 reports is interpreted as `context` by current consumers.

## Current-suggestion flow

`SuggestCurrent` treats all exact authored links supplied by the corpus as visible. It asks for only additional candidates and returns the same ranked suggestion model used by review and precision-source commands.

Current suggestions are not holdout answers. Their validity requires review or labeled evaluation.

## State and data ownership

- `internal/evidence` owns candidate evidence and fingerprints.
- `internal/codemapbench` owns admission, score, ordering, limits, and tier.
- Ranked suggestions are rebuildable analysis output.
- Persisted decline, selection, and block state belongs to `internal/review`.

## Invariants and safety boundaries

- The current document is never its own candidate.
- Visible authored targets are excluded.
- Existing links are never emitted as removal or irrelevance suggestions.
- Ambiguous basenames and symbols do not produce unique evidence.
- Weak history or sibling evidence alone is not admitted.
- Broad evidence fan-out is discounted.
- Ordering is deterministic for identical inputs.
- Per-document output is bounded.
- Tier is metadata, not mutation authorization.
- A selected write still passes through review and source hash guards.

## Failure behavior

Invalid or outside-repository paths are ignored during normalization. Evidence collectors omit facts they cannot establish without guessing.

A ranking change can pass unit tests while reducing real precision. Such changes require the governance workflow in [Extending Codemap Analysis](../development/extending-codemap-analysis.md) and [Codemap Precision Governance](../research/codemap-precision-governance.md).

## Code map

- `internal/evidence/model.go` — evidence kinds, normalized inputs, candidates, and fingerprints.
- `collect.go` — candidate aggregation and exclusions.
- `mentions.go`, `structure.go`, `symbols.go`, and `history.go` — current signal collectors.
- `internal/codemapbench/suggestions.go` — weights, admission, fan-out discount, bounds, and tiers.
- `adapters.go` — conversion between dataset/evidence/report models.
- `current.go` — non-holdout current suggestion flow.

## Tests

Focused tests cover every evidence signal, token boundaries, repeated mentions, unique basenames, symbol ambiguity, existing-target exclusion, fingerprints, weights, caps, weak-signal rejection, fan-out discount, hard-link limits, and deterministic ordering.

```bash
go test ./internal/evidence ./internal/codemapbench -count=1
```

Pinned validation cases live under `internal/evidence/testdata/` and research evaluation artifacts.

## Related docs

- [Codemap Pipeline](codemap-pipeline.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Review Ledger](review-ledger.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

Weights and thresholds are current implemented policy, not timeless constants. Changes require both focused mechanics tests and reviewed empirical evidence.
