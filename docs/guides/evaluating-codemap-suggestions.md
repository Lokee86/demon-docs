# Evaluating Codemap Suggestions

Parent index: [Guides](./README.md)

## Purpose

This guide exports authored codemap data, runs deterministic missing-link holdouts, creates a curated precision sample, evaluates labeled suggestions, and applies review thresholds without treating benchmark output as automatic documentation authority.

## Overview

Demon Docs has two related but separate codemap workflows:

```text
research and calibration
= export, benchmark, precision source, sample, evaluate

repository decisions
= ddocs suggestions select, decline, reconsider
```

The research commands measure and inspect the evidence pipeline. They do not silently edit authored codemaps. Repository suggestion commands expose current candidates through the review ledger and convert only an explicit selection into a normal hash-guarded repair.

The system is one-directional: it may suggest a potentially missing code target. It never recommends removing an existing authored codemap link as irrelevant.

## Prerequisites

- The repository contains Markdown documents with configured codemap headings.
- Codemap entries resolve against the intended repository snapshot.
- The repository revision and generated artifacts can be pinned for repeatable evaluation.
- Human reviewers can label sampled suggestions independently of the scoring code.

## Export the authored dataset

Inspect command options:

```bash
ddocs codemap export --help
```

A typical export is:

```bash
ddocs codemap export --output research/codemap-dataset.json
```

The dataset records documents, normalized codemap entries, source locations, target-resolution results, content hashes, and diagnostics. Use a committed or otherwise pinned repository snapshot when comparing reports across changes.

An export describes authored relationships. It does not generate missing-link suggestions by itself.

## Run a controlled holdout benchmark

```bash
ddocs codemap benchmark \
  --repo . \
  --seed review-v1 \
  --holdout-fraction 0.2 \
  --format json \
  --output research/codemap-holdout.json
```

The benchmark hides a deterministic subset of known authored links and measures whether the evidence pipeline ranks the hidden targets again.

Use exactly one holdout-size option:

```text
--holdout-count N
--holdout-fraction FLOAT
```

The default fraction is `0.2`. `--trusted-links PATH` can restrict ground truth to an independently reviewed link set. `--dataset PATH` reuses a previously exported dataset.

Optional threshold gates:

```bash
ddocs codemap benchmark \
  --min-precision 0.60 \
  --min-recall 0.70
```

Exit behavior:

```text
0 = benchmark completed and requested thresholds passed
1 = benchmark completed but a threshold failed
2 = arguments or benchmark execution failed
```

Holdout recovery measures whether authored relationships can be rediscovered. It is useful for regression testing, but it is not independent proof that new suggestions are correct.

## Generate current suggestions for precision review

```bash
ddocs codemap precision source \
  --repo . \
  --output research/current-suggestions.json
```

This command keeps all authored codemap links visible and generates current missing-link candidates. It does not simulate hidden links.

Optional inputs:

```text
--dataset PATH
--exclude-prefix PATH   repeatable
--output PATH
```

Use `--exclude-prefix` for index files, generated material, or other document populations intentionally excluded from the evaluation. Record exclusions with the retained report.

## Create a deterministic sample

```bash
ddocs codemap precision sample \
  --suggestions research/current-suggestions.json \
  --count 150 \
  --seed precision-v1 \
  --repository example/repository \
  --revision COMMIT_SHA \
  --output research/precision-sample.json
```

The default sample count is `150`. The output is an unlabeled benchmark template. Sampling is deterministic for the same source report, count, and seed.

Repository and revision metadata identify the evaluated population. They do not alter ranking.

## Label the sample

Review each sampled candidate against the pinned repository snapshot and record the required labels in the benchmark file.

The evaluator must not use hidden implementation assumptions or expected score changes as the oracle. Labels should answer whether the proposed documentation-to-code relationship is valid for the sampled document and target.

Keep distinctions such as valid link, useful contextual relationship, and junk evidence explicit when the benchmark schema provides them. Do not relabel an existing authored link as a removal candidate; removal quality is outside this product contract.

## Evaluate the labeled sample

```bash
ddocs codemap precision evaluate \
  --benchmark research/precision-sample.json \
  --suggestions research/current-suggestions.json \
  --format json \
  --output research/precision-evaluation.json
```

The legacy flag-only `ddocs codemap precision` form is equivalent to `evaluate`, but new scripts should use the explicit subcommand.

Evaluation compares the labeled sample with the exact deterministic suggestion report. Use the same pinned report that produced the sample unless the purpose is explicitly to compare a changed model against fixed labels.

## Interpret the results

Report at least:

- repository and revision;
- suggestion-report hash or retained path;
- sample count and seed;
- exclusions;
- ranking or tier population;
- precision and recall definitions;
- missing or stale sampled candidates;
- labeling limitations; and
- whether thresholds changed.

A result from one repository or curated sample does not establish universal product quality. Compare like-for-like reports and add new independently labeled corpora before broadening claims.

## Use repository suggestion decisions

After calibration, inspect current repository candidates with:

```bash
ddocs suggestions
ddocs suggestions show SUGGESTION
ddocs suggestions select SUGGESTION CANDIDATE
ddocs suggestions decline SUGGESTION CANDIDATE --reason "..."
```

Selecting a candidate applies the normal repair and records it in `ddocs changes`. Declines persist while the relationship and evidence fingerprint remain unchanged.

Research reports do not bypass this review path.

## Expected result

- Dataset and reports are reproducible from a pinned repository snapshot.
- Holdouts detect ranking regressions without pretending to be independent precision labels.
- Precision samples are deterministic and independently reviewed.
- Metrics state their population and limitations.
- Threshold failures are visible to CI or research scripts.
- No benchmark command edits authored codemap relationships.
- Current accepted or declined candidates remain auditable through the review ledger.

## Failure and recovery

### The export contains unresolved targets

Correct authored paths or record the resolution limitation before benchmarking. Unresolved entries should not be silently converted into known links.

### Holdout results are unexpectedly high

Check for leakage: generated reverse indexes, index documents, duplicated codemaps, retained oracle files, or target paths copied into evaluation-only inputs may make recovery trivial.

### Holdout results are unexpectedly low

Inspect corpus construction, exclusions, target normalization, document quality, and whether the known links are appropriate ground truth. Do not tune solely to one repository without preserving out-of-sample checks.

### The sample cannot be evaluated

Confirm every required item is labeled and the evaluation uses a suggestion report compatible with the sampled identifiers.

### A threshold fails

Retain both reports, inspect rank and evidence changes, and decide whether the change is a regression, a corrected false positive, or a population change. Do not lower thresholds merely to make the command pass.

## Related docs

- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Missing-Link Evidence](../research/codemap-evidence.md)
- [Reviewing Suggestions and Changes](reviewing-suggestions-and-changes.md)
- [CLI Reference](../reference/cli.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)
- [Current Product Limitations](../limits/current-limitations.md)

## Notes

Precision and holdout evaluation answer different questions. Keep their artifacts, terminology, and conclusions separate.
