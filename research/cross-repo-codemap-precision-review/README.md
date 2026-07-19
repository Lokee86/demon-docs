# Cross-repository codemap precision review

This directory measures real suggestion precision across the non-index benchmark repositories. It is intentionally separate from positive-link holdout recovery and from the monolithic-index stress corpus.

Baseline: `a5f095f787bd6bbddd625e022398442b77519927` with algorithm commit `6ea39964a77919c3a6228b475904e9c530a16a4d`.

Files:

- `sample-manifest.json` — frozen deterministic sample with score, tier, and evidence metadata.
- `labels.json` — reviewer queue with score and tier omitted to reduce review bias.
- `RUBRIC.md` — fixed labeling definitions.
- `evaluation.json` — generated only after every sampled suggestion is labeled.
- `tools/build_sample.py` — deterministic stratified sampler.
- `tools/evaluate_labels.py` — precision calculator.

The sampler includes only `primary` and `diagnostic` benchmark modes. It caps each repository at 25 hard suggestions and 25 context suggestions, selecting across score buckets and evidence-family combinations. The `stress` mode is excluded.

The repository-level split is frozen before labeling: agent-orchestrator, beads-rust, Genesis, and render-claude-context form the tuning review set; Bifrost is reserved as the validation repository because it contains all sampled unmatched hard-tier suggestions. The next algorithm pass must not inspect Bifrost labels until its changes and thresholds are frozen.

Run from the repository root:

```text
python research/cross-repo-codemap-precision-review/tools/build_sample.py
python research/cross-repo-codemap-precision-review/tools/evaluate_labels.py

# Deliberately discard an existing queue and regenerate blank labels:
python research/cross-repo-codemap-precision-review/tools/build_sample.py --reset-labels
```

Do not change the algorithm on this branch. Complete and commit the review dataset first; the next tuning branch should start from the reviewed result.
