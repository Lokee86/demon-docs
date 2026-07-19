# Cross-repository codemap precision findings

Baseline algorithm: `6ea39964a77919c3a6228b475904e9c530a16a4d`

The review covers 121 unmatched suggestions from five non-index repositories. Monolithic per-file indexes remain excluded. Labels use the frozen `valid`, `plausible`, and `incorrect` rubric.

## Aggregate results

| Scope | Reviewed | Valid | Plausible | Incorrect | Strict precision | Relevance |
|---|---:|---:|---:|---:|---:|---:|
| Overall | 121 | 83 | 34 | 4 | 68.60% | 96.69% |
| Tuning repositories | 93 | 64 | 25 | 4 | 68.82% | 95.70% |
| Bifrost validation repository | 28 | 19 | 9 | 0 | 67.86% | 100.00% |
| Hard tier | 3 | 3 | 0 | 0 | 100.00% | 100.00% |
| Context tier | 118 | 80 | 34 | 4 | 67.80% | 96.61% |

The hard-tier result is encouraging but not statistically broad: all three unmatched hard suggestions come from Bifrost. It confirms those concrete suggestions, not a universal 100% hard precision claim.

## Repository results

| Repository | Reviewed | Valid | Plausible | Incorrect | Strict precision | Relevance |
|---|---:|---:|---:|---:|---:|---:|
| agent-orchestrator | 25 | 16 | 9 | 0 | 64.00% | 100.00% |
| beads-rust | 25 | 20 | 4 | 1 | 80.00% | 96.00% |
| bifrost | 28 | 19 | 9 | 0 | 67.86% | 100.00% |
| genesis | 25 | 18 | 4 | 3 | 72.00% | 88.00% |
| render-claude-context | 18 | 10 | 8 | 0 | 55.56% | 100.00% |

Repository-wide agent guides produce many genuinely relevant candidates, but broad directories, supporting tests, and utility files are often context rather than direct codemap additions. Genesis supplied all three incidental asset or workflow errors, while the only beads-rust error was `Cargo.lock`.

## Evidence findings

| Evidence family | Reviewed | Strict precision | Relevance |
|---|---:|---:|---:|
| declared symbol mention | 8 | 75.00% | 100.00% |
| dependency neighbor | 60 | 66.67% | 100.00% |
| exact path mention | 31 | 61.29% | 96.77% |
| sibling of existing target | 43 | 74.42% | 100.00% |
| test counterpart | 9 | 33.33% | 100.00% |
| unique basename mention | 27 | 85.19% | 88.89% |

The families overlap, so these rows are diagnostic slices rather than additive populations.

The strongest next-pass conclusions are:

- Test counterparts are usually relevant but frequently optional; they should remain context without stronger evidence.
- Dependency neighbors are highly relevant but only moderately precise as direct codemap additions.
- Exact path mentions are usually relevant, supporting the current requirement for repetition plus independent corroboration before hard promotion.
- Unique basename matches can identify strong subsystem directories, but need negative boundaries for incidental assets, workflow infrastructure, and dependency lockfiles.
- Sibling evidence generalized well as a relevance signal, but still needs another ownership signal before automatic hard promotion.

## Next tuning boundary

The next algorithm pass should use only the four tuning repositories. Bifrost labels are reserved for repository-level validation and must not be consulted while selecting rules or thresholds.

The narrowest justified tuning target is negative evidence for incidental targets:

- dependency lockfiles;
- deeply nested asset or sample-content directories;
- CI or workflow infrastructure when the document describes runtime architecture;
- broad test counterparts without a named behavior or symbol relationship.

The pass should not lower global thresholds or convert all relevant context into hard links. Its target is fewer clearly incorrect candidates while preserving the current 95.70% tuning-set relevance and the frozen positive-link recovery benchmark.

## Incidental-target tuning pass

Final algorithm: `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`

The pass adds narrow negative evidence rather than changing global thresholds:

- dependency lockfiles are suppressed only when path or basename evidence is unsupported by another structural or semantic signal;
- deeply nested asset, example, fixture, sample, or test-data targets are suppressed when they arise only from weak unique-basename evidence;
- children of `.github/workflows/` are suppressed under the same weak-evidence condition; and
- explicit path mentions, dependency relationships, declared symbols, related-document targets, sibling evidence, and Git co-change continue to preserve relevant context.

The fixed review sample changes as follows:

| Scope | Retained valid | Retained plausible | Suppressed incorrect | Retained strict precision | Retained relevance |
|---|---:|---:|---:|---:|---:|
| Overall | 83 | 34 | 4 | 70.94% (83/117) | 100.00% |
| Tuning repositories | 64 | 25 | 4 | 71.91% (64/89) | 100.00% |
| Bifrost validation | 19 | 9 | 0 | 67.86% (19/28) | 100.00% |

No reviewed valid or plausible suggestion changes tier or disappears. Four replacement context candidates surfaced after the removals: `rust-toolchain.toml` and the three Genesis solver implementations. Manual inspection found all four to be direct owners explicitly named by their documents.

The frozen cross-repository holdout remains 11/18 recovered with four hard and seven context recoveries. Bifrost remains 2/3 recovered, both hard tier. The separate gbrain index stress result remains 3/10, all context.

Space Rocks validation remains unchanged at 75.00% hard-link strict precision, 98.53% hard-link relevance, 51/70 labeled-valid hard-link recovery, 621 hard candidates, 3,872 context candidates, and 10/10 canonical holdout recovery.

A broader rule that demoted generic test counterparts was evaluated and rejected. It left hard precision at 75.00% but reduced labeled-valid hard-link recovery from 51/70 to 33/70. The final pass therefore does not alter test-counterpart tiering.

## Limitations

This is a single-reviewer semantic judgment pass. The corpus is dominated by repository-wide architecture and agent-guidance documents, and only three unmatched hard-tier suggestions were available. A later adjudication pass or additional repositories with scoped feature documents would strengthen the precision estimate.
