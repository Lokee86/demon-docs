---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-74a7-b456-2b3b0aa24a6e
document_type: general
policy_exempt: false
summary: This document is the durable chronological record of the codemap missing-link algorithm, its benchmark program, tuning decisions, rejected experiments, and measured baselines.
---
# Codemap Algorithm Development Log

Parent index: [Demon Docs Documentation](./README.md)

## Purpose

This document is the durable chronological record of the codemap missing-link algorithm, its benchmark program, tuning decisions, rejected experiments, and measured baselines.

The maintained description of current behavior lives in [Codemap Suggestion Algorithm](codemap-suggestion-algorithm.md). Research artifacts remain under `research/` and are linked below.

The work was developed on July 18–19, 2026. The final algorithm baseline covered by this log is `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`.

## Fixed Design Decisions

The development program began with several constraints that remained unchanged:

- Suggest potentially missing semantic links only.
- Never suggest that an existing link is irrelevant or should be removed.
- Preserve declined suggestions and suppress the same evidence fingerprint in future; reconsider only after material evidence change.
- Separate deterministic evidence collection from scoring, tiering, human judgment, and graph mutation.
- Treat high-confidence direct-link review and useful agent context as different outputs.
- Keep monolithic per-file indexes separate from ordinary codemap calculations.
- Prefer narrow evidence rules over repository-specific special cases or global threshold chasing.

## Development Phases

### Phase 1: Format inventory and trusted evidence foundation

The first phase established that the system could learn from Space Rocks without hard-coding the Space Rocks Markdown layout.

The work inventoried codemap conventions, normalized authored document-to-code pairs, created a trusted Space Rocks review set, and added the deterministic evidence collector.

Evidence included explicit paths, unique basenames, accepted-target siblings, source/test counterparts, dependency neighbors, Git co-change, and targets inherited from related documents.

Key outcome: evidence became a reviewable fact layer rather than an automatic coverage decision.

### Phase 2: Benchmark harness and repository adapters

The second phase built a repeatable hidden-link benchmark:

1. Extract known authored links.
2. Hide a deterministic subset.
3. Run the evidence and ranking pipeline without those links.
4. Measure which hidden links reappear and at what tier.
5. Export canonical text and JSON reports.

The benchmark CLI, repository corpus adapter, report exporters, orchestration, and evidence validation landed as independent streams and were then reconciled.

Key outcome: tuning could be measured against frozen inputs rather than anecdotal inspection.

### Phase 3: Ranking and stronger semantic evidence

The first ranked algorithm assigned evidence weights, applied logarithmic repetition and fanout controls, and bounded output per document.

Declared-symbol evidence was then added as the strongest direct semantic signal. This allowed a document that names an implementation symbol to identify its owner without relying only on paths or repository shape.

Key outcome: the algorithm moved from an unordered evidence inventory to a deterministic suggestion surface.

### Phase 4: Space Rocks authored-links precision benchmark

The hidden-link benchmark measured recovery but could not measure user-facing precision because every suggestion outside the hidden positive set looked false.

A separate authored-links benchmark therefore left existing links visible, generated genuinely new suggestions, and manually labeled a deterministic 150-item sample:

- 70 `valid_missing_link`;
- 62 `plausible_but_unnecessary`;
- 18 `incorrect`.

Initial aggregate results:

| Metric | Initial result |
|---|---:|
| Strict precision | 46.67% (70/150) |
| Relevance/acceptance | 88.00% (132/150) |
| Precision@1 | 60.00% |
| Precision@3 | 54.67% |
| Precision@5 | 52.80% |

The result showed that one binary suggestion list conflated direct missing links with useful but optional context.

### Phase 5: Hard-link and context tiers

The first tuning pass retained the whole relationship set but divided it into:

- `hard_link`: bounded direct-link review; and
- `context`: weaker, indirect, optional, or already-explicit relationships.

Initial hard-link qualification used declared symbols, test counterparts, or sufficiently strong dependency evidence, with at most five hard links per document.

Results:

| Metric | Pass 1 |
|---|---:|
| Hard-link suggestions | 81 |
| Hard strict precision | 64.20% (52/81) |
| Hard relevance | 95.06% (77/81) |
| Labeled-valid hard recovery | 74.29% (52/70) |
| Full-pool hard links | 602 |
| Hidden-link holdout | 10/10 |

Key decision: useful context should not be discarded merely because it is not appropriate for permanent insertion.

### Phase 6: Corroborated structural qualification

The second tuning pass tightened several paths:

- exact-path mentions stayed context instead of automatically becoming hard links;
- source/test counterparts required independent dependency, related-document, or sibling support;
- related-document targets could qualify when direct Git document co-change corroborated them; and
- the five-item cap counted qualifying candidates rather than allowing weak context items to consume hard-link positions.

Results:

| Metric | Pass 2 |
|---|---:|
| Hard-link suggestions | 70 |
| Hard strict precision | 72.86% (51/70) |
| Hard relevance | 97.14% (68/70) |
| Labeled-valid hard recovery | 72.86% (51/70) |
| Full-pool hard links | 631 |
| Hidden-link holdout | 10/10 |

Strict precision increased substantially while losing one labeled-valid hard candidate.

### Phase 7: Directional counterpart confidence

The third tuning pass distinguished test verification targets from production implementation targets:

- dependency-only hard qualification required score 18 rather than 16;
- supported test counterparts could still qualify directly; and
- non-test implementation counterparts required score 20 in addition to independent support.

Results:

| Metric | Pass 3 |
|---|---:|
| Hard-link suggestions | 68 |
| Hard strict precision | 75.00% (51/68) |
| Hard relevance | 98.53% (67/68) |
| Labeled-valid hard recovery | 72.86% (51/70) |
| Full-pool hard links | 621 |
| Hidden-link holdout | 10/10 |

This became the stable Space Rocks precision baseline.

### Phase 8: Cross-repository benchmark

Space Rocks was not sufficient evidence for a repository-agnostic claim. A separate corpus was built from pinned open-source repositories with explicit document-to-code mappings.

The calculation corpus covered five ordinary repositories and multiple languages. A sixth repository, gbrain, used one monolithic per-file index with hundreds of targets. It was classified as a stress case because hiding the index removed nearly all topical evidence.

Initial ordinary-corpus recovery:

- 18 hidden links;
- 11 recovered;
- one hard recovery;
- ten context recoveries.

The first cross-repository tuning rule promoted an exact path only when:

- it appeared at least twice; and
- dependency or declared-symbol evidence independently corroborated it.

After tuning:

| Metric | Before | After |
|---|---:|---:|
| Total recovered | 11/18 | 11/18 |
| Hard recovered | 1 | 4 |
| Context recovered | 10 | 7 |
| Primary recovery | 6/8 | 6/8 |
| Primary hard recovery | 0 | 2 |
| Stress recovery | 3/10 context | 3/10 context |

This was a confidence-tier improvement, not a recall increase.

### Phase 9: Cross-repository manual precision review

Positive-link recovery still could not determine whether unmatched suggestions were useful. A deterministic sample of 121 unmatched suggestions was manually labeled across the five ordinary repositories.

The split was frozen before tuning:

- tuning: agent-orchestrator, beads-rust, Genesis, render-claude-context;
- validation: Bifrost.

Initial results:

| Scope | Reviewed | Valid | Plausible | Incorrect | Strict | Relevance |
|---|---:|---:|---:|---:|---:|---:|
| Overall | 121 | 83 | 34 | 4 | 68.60% | 96.69% |
| Tuning | 93 | 64 | 25 | 4 | 68.82% | 95.70% |
| Bifrost | 28 | 19 | 9 | 0 | 67.86% | 100.00% |
| Hard tier | 3 | 3 | 0 | 0 | 100.00% | 100.00% |
| Context tier | 118 | 80 | 34 | 4 | 67.80% | 96.61% |

The three hard suggestions all came from Bifrost, so the 100% result was explicitly not treated as a broad hard-tier precision estimate.

The four errors were:

- unsupported `Cargo.lock`;
- two deeply nested asset/source directories matched by generic basenames; and
- workflow scripts matched by the generic basename `scripts`.

### Phase 10: Narrow incidental-target rejection

The final pass added negative evidence for only the demonstrated failure classes.

Rules:

- suppress dependency lockfiles without evidence beyond exact path or unique basename;
- suppress deeply nested asset/example/fixture/sample/test-data targets produced only by unique-basename matching; and
- suppress `.github/workflows/` children produced only by unique-basename matching.

Nested-content and workflow targets with explicit paths remain. Lockfiles and basename matches remain when independently corroborated.

Final fixed-sample result:

| Scope | Valid retained | Plausible retained | Incorrect suppressed | Retained strict | Retained relevance |
|---|---:|---:|---:|---:|---:|
| Overall | 83 | 34 | 4 | 70.94% | 100.00% |
| Tuning | 64 | 25 | 4 | 71.91% | 100.00% |
| Bifrost | 19 | 9 | 0 | 67.86% | 100.00% |

No reviewed valid or plausible suggestion changed tier or disappeared.

Four replacement context candidates surfaced after the removals: `rust-toolchain.toml` and three Genesis solver implementations. Manual inspection found all four to be direct owners explicitly named by their documents.

Cross-repository recovery remained 11/18 with four hard recoveries. Bifrost remained 2/3, both hard. The separate index stress result remained 3/10, all context.

Space Rocks remained unchanged at:

- 75.00% hard strict precision;
- 98.53% hard relevance;
- 51/70 labeled-valid hard recovery;
- 621 hard and 3,872 context candidates in the full source pool; and
- 10/10 canonical hidden-link recovery.

## Rejected or Revised Experiments

### Pooling the monolithic index with ordinary repositories

**Observed:** gbrain uses one index document with hundreds of file mappings. Hiding its authored index removes the document's topical evidence and tests a different retrieval problem.

**Decision:** exclude `stress` mode from ordinary aggregate calculations and report it separately.

### Broad dependency-lockfile suppression

**Observed:** the first negative rule also removed a Space Rocks `go.sum` candidate that had sibling and Git co-change support and had been reviewed as relevant context.

**Decision:** suppress lockfiles only when exact-path or basename evidence is unsupported. Corroborated lockfiles remain context.

### Generic test-counterpart demotion

**Observed:** broadly demoting test counterparts left hard precision at 75.00% but reduced labeled-valid hard recovery from 51/70 to 33/70.

**Decision:** reject the experiment. Test counterparts retain their prior behavior and require independent support for hard qualification.

### Global score reduction or broad context deletion

**Observed:** the manual review showed 34 plausible context candidates and only four outright errors. Broad deletion would improve strict precision by discarding useful relationships rather than improving semantic discrimination.

**Decision:** retain the context tier and tune only demonstrated negative-evidence classes.

### Further tuning against the same four errors

**Risk:** repeated tuning on the same fixed sample would overfit the existing corpus.

**Decision:** stop algorithm tuning and move to early product implementation testing. New repositories, scoped feature documents, and real accept/decline outcomes should precede another pass.

## Commit Ledger

The following commits form the algorithm and benchmark development chain.

| Commit | Change |
|---|---|
| `683a15d9` | Added deterministic missing-link evidence collection. |
| `b071e162` | Added the trusted Space Rocks codemap review set. |
| `529bed1a` | Inventoried Space Rocks codemap formats. |
| `b02686f8` | Added the hidden-link benchmark harness. |
| `09d587db` | Added repository dataset extraction and export. |
| `3fc77e5c` | Merged codemap format inventory. |
| `ba9a96f2` | Merged evidence collector stream. |
| `f246d023` | Merged benchmark harness stream. |
| `750085a0` | Merged trusted review-set stream. |
| `548cbaf0` | Reconciled extraction and benchmark streams. |
| `2e43bbc0` | Added benchmark report exports. |
| `447db213` | Added benchmark orchestration. |
| `5ba7dda7` | Added benchmark CLI contract. |
| `18a7f26d` | Added evidence-signal validation. |
| `0115aa11` | Added repository corpus adapter. |
| `278b0ec1` | Merged benchmark runner. |
| `b07278e8` | Merged benchmark reports. |
| `d5561011` | Merged evidence validation. |
| `a3c99fa2` | Merged benchmark CLI. |
| `e47c7ff9` | Integrated benchmark command. |
| `904ce82b` | Added deterministic ranked suggestions. |
| `802db193` | Added declared-symbol evidence. |
| `2420deaf` | Added the first curated precision benchmark. |
| `486aabd8` | Added the labeled precision benchmark implementation. |
| `d477759f` | Finalized authored-links precision artifacts. |
| `c7d09234` | Kept the large precision source report temporary. |
| `0ed2cd62` | Added `hard_link` and `context` tiers. |
| `6acbbbc5` | Tightened structural hard-link qualification. |
| `aa6eb48c` | Tightened directional counterpart confidence. |
| `95b3ed43` | Added the cross-repository benchmark corpus. |
| `6ea39964` | Promoted corroborated repeated references. |
| `a5f095f7` | Separated index stress and recorded wider tuning. |
| `2dac7740` | Added cross-repository manual precision review. |
| `3c98fedb` | Added incidental-target rejection. |
| `215cef7b` | Narrowed lockfile handling to preserve supported context. |
| `b7dfc598` | Reverted the harmful broad test-counterpart penalty. |
| `73657346` | Recorded final incidental-target tuning results. |

The commit subjects are not the complete specification. The current behavior is defined by the code and [Codemap Suggestion Algorithm](codemap-suggestion-algorithm.md).

## Artifact Registry

### Initial review and evidence validation

- `research/codemap-inventory/`: codemap format fixtures and normalized inventory.
- `research/codemap-review/`: trusted Space Rocks links and review findings.
- `research/codemap-evidence-validation/`: evidence validation record.

### Space Rocks precision

- `research/codemap-precision/space-rocks-precision-sample-150.json`: deterministic sample.
- `research/codemap-precision/space-rocks-precision-benchmark.json`: labels, rationales, references, and hashes.
- `research/codemap-precision/evaluation.json`: current precision evaluation.
- `research/codemap-precision/README.md`: pass-by-pass Space Rocks results and reproduction.

### Cross-repository recovery

- `research/cross-repo-codemap-benchmark/candidates.json`: discovery shortlist and extraction modes.
- `research/cross-repo-codemap-benchmark/corpus/`: normalized explicit mappings.
- `research/cross-repo-codemap-benchmark/datasets/`: benchmark inputs.
- `research/cross-repo-codemap-benchmark/reports/`: per-repository outputs.
- `research/cross-repo-codemap-benchmark/evaluation.json`: aggregate evaluation.
- `research/cross-repo-codemap-benchmark/results.md`: readable results.
- `research/cross-repo-codemap-benchmark/README.md`: workflow, modes, and pass history.

### Cross-repository precision

- `research/cross-repo-codemap-precision-review/sample-manifest.json`: frozen stratified sample with evidence metadata.
- `research/cross-repo-codemap-precision-review/labels.json`: blind review queue and completed labels.
- `research/cross-repo-codemap-precision-review/RUBRIC.md`: fixed label definitions.
- `research/cross-repo-codemap-precision-review/evaluation.json`: pre-tuning manual precision.
- `research/cross-repo-codemap-precision-review/tuning-pass-3-review-comparison.json`: fixed-sample survival check.
- `research/cross-repo-codemap-precision-review/tuning-pass-3-summary.json`: final validation summary.
- `research/cross-repo-codemap-precision-review/FINDINGS.md`: analysis and tuning conclusions.

## Verification Gate

The final algorithm and research baseline passed:

```text
go test ./... -count=1
go vet ./...
go build ./cmd/ddocs ./cmd/demon
```

The benchmark scripts also regenerated the pinned cross-repository reports, validated Bifrost only after rules were frozen, reproduced Space Rocks precision, and recovered the canonical Space Rocks holdout 10/10.

## Current Readiness Decision

The algorithm is accepted for early implementation testing.

The next work should implement the review workflow, persistent declines, and real repository dogfooding. The system must continue to present suggestions as reviewable evidence rather than automatic truth.

Another tuning pass should begin only after collecting materially new evidence from additional repositories or actual user accept/decline outcomes.
