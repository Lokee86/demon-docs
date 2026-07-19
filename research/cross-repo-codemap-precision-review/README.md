# Cross-repository codemap precision review

This directory measures real suggestion precision across the non-index benchmark repositories. It is intentionally separate from positive-link holdout recovery and from the monolithic-index stress corpus.

Review baseline: `a5f095f787bd6bbddd625e022398442b77519927` with algorithm commit `6ea39964a77919c3a6228b475904e9c530a16a4d`.

Incidental-target tuning baseline: `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`.

Files:

- `sample-manifest.json` — frozen deterministic sample with score, tier, and evidence metadata.
- `labels.json` — reviewer queue with score and tier omitted to reduce review bias.
- `RUBRIC.md` — fixed labeling definitions.
- `evaluation.json` — generated only after every sampled suggestion is labeled.
- `tuning-pass-3-review-comparison.json` — fixed-sample survival check after tuning.
- `tuning-pass-3-summary.json` — cross-repository, Bifrost, Space Rocks, and holdout validation summary.
- `tools/build_sample.py` — deterministic stratified sampler.
- `tools/evaluate_labels.py` — precision calculator.

The sampler includes only `primary` and `diagnostic` benchmark modes. It caps each repository at 25 hard suggestions and 25 context suggestions, selecting across score buckets and evidence-family combinations. The `stress` mode is excluded.

The repository-level split was frozen before labeling: agent-orchestrator, beads-rust, Genesis, and render-claude-context formed the tuning review set; Bifrost was reserved as the validation repository because it contains all sampled unmatched hard-tier suggestions. The incidental-target rules were selected against the tuning repositories, frozen, and only then checked against Bifrost.

Run from the repository root:

```text
python research/cross-repo-codemap-precision-review/tools/build_sample.py
python research/cross-repo-codemap-precision-review/tools/evaluate_labels.py

# Deliberately discard an existing queue and regenerate blank labels:
python research/cross-repo-codemap-precision-review/tools/build_sample.py --reset-labels
```

## Incidental-target tuning result

The user explicitly kept the tuning pass in this worktree after the review dataset was committed. The final rules suppress unsupported dependency lockfiles and weak basename-only matches to deeply nested content or workflow infrastructure. They preserve explicit-path, structural, and semantic evidence.

Against the frozen 121-item review sample, all 83 valid and 34 plausible suggestions remain and all four incorrect suggestions are removed. Bifrost remains unchanged at 19 valid and nine plausible suggestions. Cross-repository holdout recovery remains 11/18, and the separate gbrain index stress result remains 3/10.

A proposed broad test-counterpart demotion was rejected during Space Rocks validation because it reduced labeled-valid hard-link recovery from 51 to 33 without improving hard-link precision. Test counterparts therefore retain their prior tier behavior.
