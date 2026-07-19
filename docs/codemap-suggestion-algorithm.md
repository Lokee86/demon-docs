# Codemap Suggestion Algorithm

## Status

The codemap suggestion algorithm is ready for early implementation testing and repository dogfooding.

It is not an automatic documentation writer. Its output is a deterministic review surface divided into high-confidence `hard_link` candidates and broader `context` relationships. A person or later product workflow must still accept or decline each proposed permanent link.

The current algorithm baseline is commit `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`.

The complete development and tuning history is recorded in [Codemap Algorithm Development Log](codemap-algorithm-development-log.md).

## Product Contract

The system exists to identify potentially missing semantic links between a document and repository targets.

It must:

- suggest only links that do not already exist in the document's codemap;
- retain deterministic evidence and ordering for every candidate;
- distinguish direct-link recommendations from weaker context relationships;
- preserve useful context rather than forcing every relationship into a permanent link;
- remain repository-agnostic and independent of one Markdown codemap layout;
- never recommend that an existing link is irrelevant or should be removed;
- support persistent decline decisions keyed by the evidence fingerprint; and
- allow materially changed evidence to produce a new review opportunity.

The current implementation produces candidates and benchmark reports. Persistent decline storage and the end-user review workflow remain implementation work.

## Pipeline

The algorithm runs in seven stages.

### 1. Normalize repository facts

The corpus layer supplies normalized inputs:

- document path and text;
- repository files and directories;
- accepted codemap targets;
- dependency edges;
- declared symbols;
- bounded Git co-change facts; and
- related documents with their accepted targets.

Existing codemap targets seed structural evidence and are excluded from the missing-link output.

### 2. Collect deterministic evidence

The evidence collector creates one candidate per repository target and attaches one or more evidence records.

| Evidence kind | Meaning | Base weight |
|---|---|---:|
| `declared_symbol_mention` | The document names a symbol declared by the target. | 7 |
| `exact_path_mention` | The document contains the repository-relative target path. | 6 |
| `test_counterpart` | Source and test naming/layout identify a counterpart. | 6 |
| `unique_basename_mention` | A uniquely resolvable file or directory basename appears in the document. | 4 |
| `dependency_neighbor` | The target is a direct observed dependency neighbor of an accepted target. | 4 |
| `related_document_target` | A related document already accepts the target. | 4 |
| `sibling_target` | The target is a direct sibling of an accepted target. | 2 |
| `git_target_cochange` | The target changed with an accepted target. | 1.5 |
| `git_document_cochange` | The target changed directly with the current document. | 1 |

Each evidence record retains its kind, source, detail, count, and deterministic fingerprint.

### 3. Admit evidence-bearing candidates

A candidate enters ranking when it has either:

- at least two different evidence kinds; or
- one independently admissible kind: exact path, unique basename, declared symbol, test counterpart, dependency neighbor, or related-document target.

Weak structural or Git-only evidence cannot enter the output by itself.

### 4. Reject demonstrated incidental targets

Before ranking, narrow negative-evidence rules remove known classes of accidental matches.

A dependency lockfile is rejected only when it has no evidence beyond exact-path or unique-basename mention. Supported lockfiles remain context candidates. This distinction preserves a corroborated `go.sum` relationship while rejecting an unsupported `Cargo.lock` match.

A deeply nested asset, example, fixture, sample, or test-data target is rejected when it arises only from weak unique-basename evidence. The rule requires at least two path levels below the content marker, preventing broad top-level directories from being discarded.

A child of `.github/workflows/` is rejected under the same weak unique-basename-only condition.

For nested content and workflow targets, an explicit path or independent structural or semantic support prevents the weak-basename filter from firing. Lockfiles require support beyond exact-path or unique-basename mention.

These rules are intentionally path-specific. A broader test-counterpart penalty was tested and rejected because it removed many valid hard links without improving measured precision.

### 5. Rank candidates

Each evidence contribution is weighted, repetition-adjusted, and fanout-discounted.

Repeated evidence uses a logarithmic occurrence factor so ten repetitions do not count as ten independent facts. Exact-path and unique-basename mentions do not receive this repetition score boost because textual repetition is tracked separately for hard-link qualification.

Evidence shared across many targets receives a logarithmic fanout discount. This prevents one broad commit, directory, or symbol source from dominating the result.

Candidates are sorted by descending score, then repository-relative target path for deterministic ties.

### 6. Bound the review surface

The default retained list is the top 30 candidates per document.

Up to two repeated exact-path candidates may be reserved outside the top 30. This prevents a repeated explicit dependency from disappearing solely because a large repository produces many higher-scoring structural neighbors.

The final list remains deterministically ordered by score and target.

### 7. Assign suggestion tiers

Every retained candidate defaults to `context`. At most five candidates per document may become `hard_link`.

A candidate qualifies for `hard_link` through one of these paths:

1. **Repeated explicit path:** the exact path appears at least twice and dependency-neighbor or declared-symbol evidence independently corroborates it.
2. **Declared symbol:** a declared symbol from the target is named by the document.
3. **Test counterpart:** counterpart evidence is independently supported by dependency, related-document, or sibling evidence. Test targets may qualify directly; non-test implementation targets additionally require score 20 or greater.
4. **Dependency neighbor:** dependency evidence qualifies at score 18 or greater.
5. **Related document plus direct history:** related-document evidence is corroborated by direct Git co-change between the target and current document.

Single exact-path mentions remain `context`: the target is already visible in prose, but that alone does not prove it belongs in the permanent codemap.

Repeated paths without semantic corroboration also remain `context`.

## Output Semantics

### `hard_link`

A bounded set of higher-confidence candidates suitable for direct missing-link review.

It does not mean:

- automatically correct;
- safe to write without review;
- required for documentation completeness; or
- evidence that another existing link should be removed.

### `context`

A weaker, indirect, optional, or already-explicit relationship that may still improve agent context assembly and repository navigation.

Context candidates are not failures. The tuning program deliberately preserves relevant context while tightening only the direct-link review surface and demonstrated noise classes.

## Safety Boundaries

The algorithm does not:

- create or remove codemap links;
- infer that an existing link is irrelevant;
- delete documentation coverage;
- treat benchmark labels as universal truth;
- promote every related test, sibling, dependency, or Git neighbor;
- traverse the monolithic-index stress corpus as though it were an ordinary codemap convention; or
- reconsider an unchanged declined suggestion once persistent decline storage is implemented.

Identical inputs produce byte-stable candidate ordering, evidence ordering, counts, tiers, and fingerprints.

## Current Measured Baseline

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

## Readiness

The current baseline is appropriate for:

- implementing the suggestion review workflow;
- dogfooding against Demon Docs and Space Rocks;
- collecting accepted and declined outcomes;
- testing persistence keyed by evidence fingerprint;
- observing how real users interpret `hard_link` versus `context`; and
- expanding the benchmark with additional repositories and scoped feature documents.

It is not appropriate for unattended permanent-link insertion.

## Implementation Map

- `internal/evidence/`: deterministic evidence collection and fingerprints.
- `internal/codemapcorpus/`: repository inventory, dependency, symbol, related-document, and Git fact adapters.
- `internal/codemapbench/suggestions.go`: admission, scoring, bounded selection, and tier assignment.
- `internal/codemapbench/suggestion_negative_evidence.go`: narrow incidental-target rejection.
- `internal/codemapbench/`: holdout orchestration and canonical reports.
- `internal/codemapprecision/`: labeled precision evaluation.
- `internal/app/codemap_benchmark*.go`: benchmark CLI integration.
- `internal/app/codemap_precision*.go`: source generation, sampling, and evaluation CLI integration.
- `research/codemap-review/`: trusted Space Rocks links and initial review findings.
- `research/codemap-precision/`: Space Rocks authored-links precision benchmark.
- `research/cross-repo-codemap-benchmark/`: pinned multi-repository holdout corpus.
- `research/cross-repo-codemap-precision-review/`: frozen manual cross-repository precision sample and final tuning summary.

## Remaining Limits

- The cross-repository precision review has one reviewer.
- Most cross-repository documents are repository-wide architecture or agent-guidance documents rather than scoped feature docs.
- Only three unmatched hard-tier suggestions were available outside Space Rocks.
- Recall remains 11/18 on the ordinary cross-repository holdout.
- The current thresholds are empirical defaults, not proven universal constants.
- Decline persistence and the user-facing review loop are specified but not yet implemented.
- Continued tuning against the same fixed errors would risk overfitting; new data should precede another algorithm pass.
