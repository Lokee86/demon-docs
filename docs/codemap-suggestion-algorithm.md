---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7549-855a-80c60bf09e9c
document_type: general
policy_exempt: false
summary: This document defines the current deterministic codemap missing-link algorithm, production generation semantics, confidence tiers, measured baseline, and safety boundaries.
---
# Codemap Missing-Link Algorithm

Parent index: [Demon Docs Documentation](./README.md)

## Purpose

This document defines the current deterministic codemap missing-link algorithm, production generation semantics, confidence tiers, measured baseline, and safety boundaries.

## Overview

The algorithm identifies repository targets that are absent from a document's current codemap but supported by deterministic document, repository, dependency, symbol, related-document, and Git evidence.

The same production ranking package is used by:

```text
explicit codemap fix/check/inspect
the shared suggestions review surface
controlled hidden-link benchmarks
precision-source and sampling workflows
```

Production codemap execution automatically adds every selected non-declined recommendation, including both `hard_link` and `context` tiers. The tier remains confidence and policy metadata; it is not a separate approval gate.

Existing links are excluded from missing-link candidates and retained by default. Optional removal of undiscovered or context-tier existing links belongs to the separate codemap execution policy, not to the ranker itself.

The complete development and tuning history is recorded in [Codemap Algorithm Development Log](codemap-algorithm-development-log.md).

## Current status

The algorithm is production-owned by `internal/codemaprecommend` and is used by the explicit foreground codemap workflow. `internal/codemapbench` consumes that package rather than maintaining an independent production implementation.

It is suitable for repository dogfooding and explicit managed codemap execution under the documented conservative defaults. Its measurements remain corpus-specific rather than universal quality guarantees.

## Product contract

The system exists to identify potentially missing semantic links between a document and repository targets.

It must:

- return only targets not already visible in the current codemap;
- retain deterministic evidence and ordering for every candidate;
- distinguish stronger `hard_link` relationships from broader `context` relationships;
- preserve useful context rather than forcing every relationship into the stronger tier;
- remain repository-agnostic and independent of one Markdown codemap layout;
- never classify an existing link as irrelevant within missing-link generation;
- support persistent decline decisions keyed by evidence fingerprint;
- allow materially changed evidence to produce a new review opportunity;
- feed production execution and evaluation from one ranking implementation; and
- produce stable output for identical normalized inputs.

Production execution may automatically add the returned candidates after shared decline-policy filtering. Existing-link pruning is separately configured and disabled by default.

## Pipeline

The algorithm runs in seven stages.

### 1. Normalize repository facts

The corpus layer supplies normalized inputs:

- document path and visible document text;
- repository files and directories;
- current codemap targets;
- dependency edges;
- declared symbols;
- bounded Git co-change facts; and
- related documents with their current targets.

Existing codemap targets seed structural evidence where appropriate and are excluded from the missing-link output.

For production execution, the codemap section itself is stripped from document text before mention evidence is collected. A target therefore cannot become evidence for itself merely because it is already listed in the map.

### 2. Collect deterministic evidence

The evidence collector creates one candidate per repository target and attaches one or more evidence records.

| Evidence kind | Meaning | Base weight |
|---|---|---:|
| `declared_symbol_mention` | The document names a symbol declared by the target. | 7 |
| `exact_path_mention` | The document contains the repository-relative target path. | 6 |
| `test_counterpart` | Source and test naming/layout identify a counterpart. | 6 |
| `unique_basename_mention` | A uniquely resolvable file or directory basename appears in the document. | 4 |
| `dependency_neighbor` | The target is a direct observed dependency neighbor of a current target. | 4 |
| `related_document_target` | A related document already contains the target. | 4 |
| `sibling_target` | The target is a direct sibling of a current target. | 2 |
| `git_target_cochange` | The target changed with a current target. | 1.5 |
| `git_document_cochange` | The target changed directly with the current document. | 1 |

Each evidence record retains its kind, source, detail, count, and deterministic fingerprint contribution.

### 3. Admit evidence-bearing candidates

A candidate enters ranking when it has either:

- at least two different evidence kinds; or
- one independently admissible kind: exact path, unique basename, declared symbol, test counterpart, dependency neighbor, or related-document target.

Weak structural or Git-only evidence cannot enter the output by itself.

Admission means the candidate is eligible for ranking. It is not a universal semantic claim.

### 4. Reject demonstrated incidental targets

Before ranking, narrow negative-evidence rules remove known classes of accidental matches.

A dependency lockfile is rejected only when it has no evidence beyond exact-path or unique-basename mention. Supported lockfiles remain context candidates.

A deeply nested asset, example, fixture, sample, or test-data target is rejected when it arises only from weak unique-basename evidence. The rule requires at least two path levels below the content marker, preventing broad top-level directories from being discarded.

A child of `.github/workflows/` is rejected under the same weak unique-basename-only condition.

Explicit path evidence or independent structural or semantic support preserves nested content and workflow targets. Lockfiles require support beyond exact-path or unique-basename mention.

These rules are intentionally narrow. Broader class penalties were rejected when they removed reviewed useful links without improving measured quality.

### 5. Rank candidates

Each evidence contribution is weighted, repetition-adjusted, and fan-out-discounted.

Repeated evidence normally uses:

```text
1 + log2(occurrence count)
```

Exact-path and unique-basename mentions use a fixed occurrence factor. Repeated exact-path count remains available for stronger-tier qualification without multiplying the base score indefinitely.

Evidence shared across many targets receives a logarithmic fan-out discount. One broad commit, directory, or shared source therefore contributes less to each candidate than a specific fact.

Candidates are sorted by descending score, then repository-relative target path for deterministic ties.

### 6. Bound the output

The default retained list is the top 30 candidates per document.

Up to two repeated exact-path candidates may be reserved outside the top 30. This prevents a repeated explicit dependency from disappearing solely because a large repository produces many higher-scoring structural neighbors.

The final union remains deterministically ordered by score and target.

### 7. Assign confidence tiers

Every retained candidate defaults to `context`. At most five candidates per document may become `hard_link`.

A candidate qualifies for `hard_link` through one of these paths:

1. **Repeated explicit path:** the exact path appears at least twice and dependency-neighbor or declared-symbol evidence independently corroborates it.
2. **Declared symbol:** a declared symbol from the target is named by the document.
3. **Test counterpart:** counterpart evidence is independently supported by dependency, related-document, or sibling evidence. Test targets may qualify directly; non-test implementation targets additionally require score 20 or greater.
4. **Dependency neighbor:** dependency evidence qualifies at score 18 or greater.
5. **Related document plus direct history:** related-document evidence is corroborated by direct Git co-change between the target and current document.

Single exact-path mentions and repeated paths without independent semantic corroboration remain `context`.

## Output semantics

### `hard_link`

A bounded higher-confidence relationship under the current evidence model.

In production execution it is eligible for automatic addition after decline-policy filtering.

It does not mean:

- universally correct;
- required for documentation completeness;
- evidence that another existing link should be removed; or
- exempt from a persisted decline.

### `context`

A weaker, indirect, optional, or already-explicit relationship that survived admission and negative-evidence filtering.

In production execution it is also eligible for automatic addition after decline-policy filtering. The distinction remains visible in `inspect`, review policy, benchmark reports, and optional existing-link pruning. When `remove_low_score_links` is enabled, a hidden existing target recovered only as `context` is eligible for removal.

Context candidates are not failed hard links. They are the broader retained relationship set.

## Production execution semantics

The explicit codemap execution path:

1. computes current recommendations with all existing targets visible;
2. converts each recommendation to the shared review suggestion model;
3. suppresses unchanged declined candidates;
4. passes all remaining `hard_link` and `context` targets to unified section reconciliation;
5. deduplicates them against the existing codemap; and
6. publishes only exact changed files through the shared transaction layer.

The ranker does not write files directly. `internal/codemaprun` and `internal/codemap` own generation planning and managed-section mutation.

## Existing-link policy

Missing-link generation does not propose existing targets, removals, or irrelevance judgments.

Production execution separately supports two opt-in policies:

```toml
[codemap]
remove_undiscovered_links = false
remove_low_score_links = false
```

For a current resolved entry, the planner hides that entry and reruns the same algorithm:

- no recovery can select it for `remove_undiscovered_links`;
- context-only recovery can select it for `remove_low_score_links`; and
- stronger recovery retains it.

Both settings are disabled by default because failure to reconstruct human intent is not proof of irrelevance.

## Decline semantics

Declines are stored through the shared review lifecycle using document, target, relationship kind, and evidence fingerprint.

- An unchanged declined recommendation remains suppressed.
- Materially changed evidence may produce a different fingerprint and a new decision opportunity.
- A decline suppresses a proposed addition; it does not remove an existing link.
- Manual deletion does not automatically create a decline.

## Safety boundaries

The algorithm and production workflow preserve these boundaries:

- existing targets are excluded from missing-link output;
- existing valid links remain by default;
- pruning requires explicit configuration;
- benchmark labels are not universal truth;
- broad weak evidence is bounded and discounted;
- output per document is bounded;
- identical normalized inputs produce stable ordering, evidence, scores, tiers, and fingerprints;
- section mutation occurs only through explicit foreground codemap commands;
- normal `fix`, `check`, `watch`, and repository-demon paths do not invoke generation; and
- a concurrent source edit is protected by content-addressed preflight.

## Current measured baseline

### Space Rocks authored-links precision

The manually reviewed Space Rocks sample contains 150 suggestions across 25 documents.

| Metric | Current result |
|---|---:|
| Hard-link suggestions | 68 |
| Hard-link strict precision | 75.00% (51/68) |
| Hard-link relevance | 98.53% (67/68) |
| Labeled-valid hard-link recovery | 72.86% (51/70) |
| Context suggestions | 82 |
| Full source pool | 4,493 |
| Full-pool hard links | 621 |
| Full-pool context links | 3,872 |
| Canonical hidden-link holdout | 10/10 recovered |

### Cross-repository recovery

The ordinary calculation corpus contains five repositories. The monolithic per-file index is reported separately as a stress case.

| Metric | Current result |
|---|---:|
| Hidden links | 18 |
| Recovered links | 11 |
| Hard recoveries | 4 |
| Context recoveries | 7 |
| Recall | 61.11% |
| Separate index-stress recovery | 3/10, all context |

### Cross-repository manual precision review

The frozen review sample contains 121 unmatched suggestions.

Before the final incidental-target pass:

| Label | Count |
|---|---:|
| Valid missing link | 83 |
| Plausible but unnecessary | 34 |
| Incorrect | 4 |
| Strict precision | 68.60% |
| Relevance | 96.69% |

After the pass:

- all 83 valid suggestions remain;
- all 34 plausible context suggestions remain;
- all four incorrect suggestions are suppressed;
- retained strict precision is 70.94%; and
- retained relevance is 100% for this fixed sample.

This does not establish universal 100% relevance. It establishes that the four demonstrated errors were removed without losing a reviewed useful candidate in the frozen sample.

## Readiness and limits

The current baseline supports:

- explicit foreground codemap generation;
- review-ledger suppression of unwanted additions;
- dogfooding against Demon Docs and Space Rocks;
- controlled hidden-link recovery evaluation;
- deterministic precision sampling; and
- collection of new accepted and declined outcomes.

Current limits remain:

- quality is corpus-dependent;
- the cross-repository precision review has one reviewer;
- most cross-repository documents are broad architecture or agent-guidance documents rather than scoped feature docs;
- only three unmatched hard-tier suggestions were available outside Space Rocks;
- ordinary cross-repository holdout recovery remains 11/18;
- thresholds are empirical defaults rather than universal constants;
- both tiers are currently auto-added by production execution after decline filtering; and
- production missing-section creation is constrained by selected effective document schemas and remains separate from ranking quality.

Continued tuning against the same fixed errors would risk overfitting. New data should precede another algorithm pass.

## Implementation map

- `internal/evidence/` — deterministic evidence collection and fingerprints.
- `internal/codemapcorpus/` — repository inventory, dependency, symbol, related-document, and Git fact adapters.
- `internal/codemaprecommend/suggestions.go` — admission, scoring, bounded selection, and tier assignment.
- `internal/codemaprecommend/suggestion_negative_evidence.go` — narrow incidental-target rejection.
- `internal/codemaprun/` — production recommendation planning, decline filtering, pruning evaluation, and rewrite plans.
- `internal/codemap/managed*.go` — unified section adoption and syntax-preserving rendering.
- `internal/codemapbench/` — holdout orchestration and canonical reports using the production ranker.
- `internal/codemapprecision/` — labeled precision evaluation.
- `internal/app/codemap_execute*.go` — explicit fix/check/inspect integration.
- `internal/app/codemap_benchmark*.go` — benchmark CLI integration.
- `internal/app/codemap_precision*.go` — source generation, sampling, and evaluation CLI integration.
- `internal/review/` — persisted decline and reconsideration lifecycle.
- `research/codemap-review/` — trusted Space Rocks links and initial review findings.
- `research/codemap-precision/` — Space Rocks authored-links precision benchmark.
- `research/cross-repo-codemap-benchmark/` — pinned multi-repository holdout corpus.
- `research/cross-repo-codemap-precision-review/` — frozen manual cross-repository precision sample and tuning summary.

## Related docs

- [Managing Codemaps](guides/managing-codemaps.md)
- [Codemap Managed Execution](architecture/codemap-managed-execution.md)
- [Codemap Pipeline](architecture/codemap-pipeline.md)
- [Codemap Evidence and Ranking](architecture/codemap-evidence-and-ranking.md)
- [Codemap Algorithm Development Log](codemap-algorithm-development-log.md)
- [Codemap Benchmark Methodology](research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](research/codemap-precision-governance.md)
- [Current Product Limitations](limits/current-limitations.md)

## Notes

The term `suggestion` remains in report and review APIs for compatibility. In the production codemap command, selected non-declined recommendations are generation inputs rather than a mandatory per-run approval queue.
