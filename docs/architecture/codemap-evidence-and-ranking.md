---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-787b-ae80-fc5555714de5
document_type: general
policy_exempt: false
summary: This document describes how normalized repository facts become deterministic potentially-missing-link candidates, roles, scores, ordered suggestions, and hardlink or context tiers.
---
# Codemap Evidence and Ranking

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes how normalized repository facts become deterministic potentially-missing-link candidates, semantic roles, scores, ordered suggestions, and `hard_link` or `context` tiers.

## Overview

Evidence and ranking are separate ownership seams:

```text
Corpus facts + visible authored targets
-> evidence candidates and fingerprints
-> candidate admission
-> deterministic role classification
-> weighted scoring and fan-out discount
-> bounded deterministic ordering
-> suggestion tier
```

Evidence explains why a target may be relevant. Ranking decides which candidates are surfaced first. Neither establishes universal semantic truth. The explicit production codemap command automatically adds only selected non-declined `hard_link` candidates; `context` remains an inspect/review surface. Mutation ownership remains in `internal/codemaprun` and `internal/codemap`, not in the ranker.

## Code root

```text
internal/evidence/
internal/codemaprecommend/
internal/codemaprun/
internal/codemapbench/adapters.go
internal/codemapbench/current.go
```

## Responsibilities

This boundary owns:

- excluding the current document and visible authored targets;
- collecting implemented mention, structural, dependency, semantic-relationship, history, related-document, and symbol evidence;
- canonical evidence ordering and candidate fingerprints;
- candidate admission rules;
- deterministic candidate-role classification;
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
- managed-section rendering or transactional document mutation;
- shared decline persistence and replay; or
- removal or irrelevance judgments for existing links.

## Evidence input

`evidence.Input` contains:

```text
document path and visible text
repository files
visible existing targets
authored target provenance and resolved coverage
dependency edges
bounded semantic relationship edges
bounded commit facts
related documents and their visible targets
symbol declarations
```

All paths are normalized before candidate creation. The current document and every visible existing target are excluded from the candidate set. Authored directories additionally cover their descendants, and resolved pattern matches remain covered without becoming independent outward expansion seeds.

Outward structural, shallow-dependency, and target-history expansion is seeded only by explicitly authored resolved files. Arcana semantic relationships use a separate per-document seed set: visible exact files plus symbols that Step 4 resolved to an exact Arcana node. Directory and pattern entries never seed either expansion path. A basename-only pattern also constrains inferred evidence for non-matching siblings in its literal parent directory; direct path, basename, or symbol evidence from the current document may still surface an explicit exception.

## Current evidence kinds

### Exact path mention

The document contains the repository-relative target path at a token boundary.

This is explicit prose evidence, but repeated occurrences use a fixed occurrence factor so a copied path does not compound score indefinitely.

### Unique basename mention

The document contains a file basename that maps to exactly one repository path. Non-unique basenames are not admitted as this evidence.

### Declared symbol mention

The document mentions a declared symbol that resolves to one repository path. Ambiguous symbols are excluded by corpus construction or evidence collection.

### Sibling of existing target

The candidate shares a structural directory relationship with an explicitly authored file target. Resolved members of authored directories, patterns, or symbols do not become sibling seeds merely because resolution found a backing path.

### Test counterpart

The candidate is the source/test or implementation/spec counterpart recognized by current naming and directory rules.

### Dependency neighbor

A dependency edge connects the candidate and an explicitly authored file expansion seed in either direction. The relation and edge source contribute to the evidence atom.

### Semantic relationship

A current Arcana relationship connects the candidate and one currently visible exact authored file or verified symbol seed. The relation is drawn from the bounded allowlist owned by the corpus relationship-provider seam. This evidence remains distinct from shallow dependency evidence so Arcana call/inheritance/test structure cannot silently inherit dependency promotion policy.

A semantic relationship is currently context-only evidence for tier policy: it may be admitted as a single evidence kind and ranked for inspection, but it does not satisfy any `hard_link` eligibility path. Its score is also excluded from the numeric thresholds used by dependency and non-test counterpart hard-link promotion, so Arcana evidence cannot indirectly push an otherwise-context candidate across a mutation threshold. Its direction and relation type now contribute to deterministic candidate-role classification; stronger role-aware promotion remains deferred.

### Git co-change with document

The candidate appears in a bounded commit with the document.

### Git co-change with existing target

The candidate appears in a bounded commit with an explicitly authored file expansion seed.

### Related-document target

A locally linked related document already authors the candidate as a direct resolved file target. Broader directory, pattern, and symbol abstractions are not flattened into inherited file-level evidence.

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
semantic relationship
related-document target
```

Sibling or history evidence alone is not sufficient.

Admission is a surfacing policy, not a claim that the candidate is valid.

## Candidate role classification

Every admitted candidate receives exactly one deterministic role before scoring is published:

```text
primary_implementation
supporting_implementation
verification_test
interface_boundary
context_only
```

Role answers **what kind of relationship the candidate appears to have to the document**. Tier answers **whether current mutation policy considers the candidate strong enough for permanent insertion**. The two are deliberately independent in this phase.

Current precedence is:

1. `verification_test` for recognized test/spec paths or candidates connected to a visible seed by an incoming Arcana `tests` relationship;
2. `interface_boundary` when a visible seed points outward through `implements`, `extends`, `overrides`, `uses-trait`, or `includes`, making the candidate the contract/base/trait side of that relationship;
3. `primary_implementation` for direct current-document evidence: declared-symbol mention, exact path mention, or unique basename mention;
4. `supporting_implementation` for non-direct structural or semantic support such as dependency neighbors, semantic relationships, test counterparts, sibling targets, or related-document targets; and
5. `context_only` when the retained evidence does not establish one of the stronger roles, such as history-only corroboration.

The opposite side of `implements`/`extends`/`overrides` remains supporting implementation rather than interface/boundary. An outgoing Arcana `tests` edge means the candidate is the thing under test, not verification; an incoming `tests` edge means the candidate performs verification.

This classification is intentionally conservative and evidence-derived. It does not use an LLM, repository naming guesses for interfaces, or unbounded graph traversal. Step 7 may use roles for coverage-aware selection, but Step 6 does **not** alter score, ordering, hard-link thresholds, or automatic mutation eligibility.

Roles are emitted in inspect output and benchmark/precision reports. Legacy reports may omit role; current consumers treat an empty legacy role as `context_only` for role-level evaluation.

## Score policy

Current base weights are:

```text
declared symbol mention             7
exact path mention                  6
test counterpart                    6
unique basename mention             4
dependency neighbor                 4
semantic relationship               3
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

Dependency, semantic-relationship, and declared-symbol evidence retain detail because different relations or symbols are meaningfully distinct. Other kinds use the normalized source/kind identity needed by current scoring.

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

At most five ordered suggestions receive `hard_link`. Eligibility requires one of the implemented paths:

- a declared-symbol mention;
- an exact path mentioned at least twice and independently corroborated by declared-symbol or dependency evidence;
- a test counterpart independently supported by dependency, related-document, or sibling evidence, with non-test implementation counterparts additionally requiring score 20 or greater;
- dependency-neighbor evidence with total score at least 18; or
- for non-test targets, related-document evidence reinforced by direct Git co-change with the current document.

The tier distinguishes stronger relationships from the broader context set. Only `hard_link` is eligible for automatic addition by explicit codemap execution after decline-policy filtering. `context` remains visible for inspection, review, and downstream analysis. The tier does not make a candidate universally true and does not imply that an existing link should be removed.

An empty legacy tier in schema-1 reports is interpreted as `context` by current consumers.

## Current-suggestion flow

`SuggestCurrent` and production execution treat all current codemap links supplied by the corpus as visible. They ask for only additional candidates and return the same ranked suggestion model used by explicit codemap execution, review commands, precision-source commands, and controlled evaluation.

Production execution strips the codemap section from document text before collecting mention evidence and filters the result through shared decline policy. Only remaining `hard_link` targets are passed to unified section reconciliation; `context` recommendations remain visible without mutating the codemap. Current recommendations are not holdout answers, and measured validity still requires repository-specific review or labeled evaluation.

## State and data ownership

- `internal/evidence` owns candidate evidence and fingerprints.
- `internal/codemaprecommend` owns production admission, role classification, score, ordering, limits, negative-evidence filtering, and tier.
- `internal/codemaprun` owns production recommendation planning, decline replay, and optional pruning evaluation.
- `internal/codemapbench` owns holdouts and reports while importing the production ranker.
- Ranked recommendations are rebuildable analysis output.
- Persisted decline, selection, and block state belongs to `internal/review`.

## Invariants and safety boundaries

- The current document is never its own candidate.
- Visible authored targets are excluded, including descendants covered by authored directories and concrete matches covered by authored patterns.
- Only explicit resolved file targets seed outward structural, shallow-dependency, and target-history expansion; Arcana relationships additionally permit exact verified symbol seeds without flattening them into file-neighborhood seeds.
- Basename-only authored patterns constrain inferred non-matching siblings in their literal parent directory unless direct current-document evidence overrides the boundary.
- Existing links are never emitted as removal or irrelevance suggestions.
- Ambiguous basenames and symbols do not produce unique evidence.
- Weak history or sibling evidence alone is not admitted.
- Semantic relationships may surface context candidates but cannot independently promote a candidate to `hard_link`.
- Broad evidence fan-out is discounted.
- Ordering is deterministic for identical inputs.
- Per-document output is bounded.
- Role is deterministic relationship metadata and does not currently affect score or mutation policy.
- Tier is confidence and mutation-policy metadata.
- Explicit production generation adds only non-declined `hard_link` recommendations; `context` remains non-mutating.
- A planned write still passes through managed-section reconciliation and source-hash guards.

## Failure behavior

Invalid or outside-repository paths are ignored during normalization. Evidence collectors omit facts they cannot establish without guessing.

A ranking change can pass unit tests while reducing real precision. Such changes require the governance workflow in [Extending Codemap Analysis](../development/extending-codemap-analysis.md) and [Codemap Precision Governance](../research/codemap-precision-governance.md).

## Code map

- `internal/evidence/model.go` — evidence kinds, normalized inputs, candidates, and fingerprints.
- `collect.go` and `target_selection.go` — candidate aggregation, exclusions, and distinct file-versus-semantic expansion seed selection.
- `mentions.go`, `structure.go`, `symbols.go`, and `history.go` — current signal collectors.
- `internal/codemaprecommend/roles.go` — deterministic candidate-role classification.
- `internal/codemaprecommend/suggestions.go` — weights, admission, fan-out discount, bounds, and tiers.
- `internal/codemaprecommend/suggestion_negative_evidence.go` — narrow incidental-target filtering.
- `internal/codemaprun/build.go` — production recommendation, decline, and pruning planning.
- `internal/codemapbench/adapters.go` — conversion between dataset/evidence/report models.
- `internal/codemapbench/current.go` — current recommendation compatibility flow.

## Tests

Focused tests cover every evidence signal, token boundaries, repeated mentions, unique basenames, symbol ambiguity, existing-target exclusion, fingerprints, weights, caps, weak-signal rejection, fan-out discount, hard-link limits, and deterministic ordering.

```bash
go test ./internal/evidence ./internal/codemaprecommend ./internal/codemaprun ./internal/codemapbench -count=1
```

Pinned validation cases live under `internal/evidence/testdata/` and research evaluation artifacts.

## Related docs

- [Codemap Missing-Link Evidence](../codemap-evidence.md)
- [Codemap Managed Execution](codemap-managed-execution.md)
- [Codemap Pipeline](codemap-pipeline.md)
- [Codemap Corpus and Adapters](codemap-corpus-adapters.md)
- [Codemap Benchmark Methodology](../research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](../research/codemap-precision-governance.md)
- [Review Ledger](review-ledger.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

Weights and thresholds are current implemented policy, not timeless constants. Changes require both focused mechanics tests and reviewed empirical evidence.
