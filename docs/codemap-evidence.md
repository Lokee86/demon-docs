# Codemap Missing-Link Evidence

Demon Docs can extract authored code maps, build a deterministic repository corpus, rank possible missing targets, and evaluate the result against controlled holdouts or curated labels. This subsystem is implemented on `main`; the end-user accept/decline workflow remains planned.

The system only suggests potentially missing relationships. It never recommends that an existing codemap link is irrelevant or should be removed.

## Commands

### Export authored codemaps

```bash
ddocs codemap export --output .ddocs/codemap.json
```

`codemap export` scans Markdown documents below the selected docs root and emits a deterministic JSON dataset containing:

- document paths and content hashes;
- configured codemap headings;
- normalized folder, file, glob, placeholder, and unresolved targets;
- descriptions and source locations;
- target-resolution diagnostics; and
- stable ordering suitable for fixtures and later analysis.

Useful overrides include repeated `--heading`, `--target-root`, `--target-base repository|document`, and the normal config-selection flags. JSON is written to stdout when `--output` is omitted.

### Run a holdout benchmark

```bash
ddocs codemap benchmark --repo /path/to/repository --format json
```

The benchmark removes a deterministic subset of known authored links and asks the suggestion engine to recover them. It supports a fixed seed, exact holdout count or fraction, a prebuilt dataset, a reviewed trusted-link set, JSON or text reports, and optional minimum precision or recall thresholds.

Holdout recovery measures whether the algorithm can rediscover authored links. It does not prove that every authored link is semantically necessary or that every unheld suggestion is correct.

### Generate and evaluate precision samples

```bash
ddocs codemap precision --help
```

The precision workflow generates ranked candidates, produces deterministic review samples, and evaluates curated labels. Review labels distinguish:

- `valid_missing_link`;
- `plausible_but_unnecessary`; and
- `incorrect`.

This separates strict permanent-link precision from the broader question of whether a candidate is useful non-junk context.

## Current Architecture

The implementation is split into explicit boundaries:

1. `internal/codemap` extracts authored targets without deciding what is missing.
2. `internal/codemapcorpus` gathers repository facts such as files, dependencies, symbols, tests, related documents, and Git history.
3. `internal/evidence` converts normalized facts into inspectable evidence records and fingerprints.
4. `internal/codemapbench` orchestrates holdouts, ranking, tier assignment, and deterministic reports.
5. `internal/codemapprecision` evaluates curated labels and aggregates metrics.

The evidence collector does not parse arbitrary prose into semantic graph truth. It accepts normalized facts supplied by the codemap parser, repository corpus adapters, Git reader, and future code-graph providers.

## Evidence Signals

The current deterministic signals are:

- exact repository-relative path mentions;
- unique basename mentions;
- declared-symbol mentions;
- direct siblings of accepted targets;
- source/test counterparts;
- direct observed dependency neighbours;
- document/code and accepted-target/code Git co-change counts; and
- accepted targets shared by related documents.

Each candidate retains its evidence kind, source, detail, count, and deterministic evidence fingerprint. Scoring, tiering, human acceptance, and authored link mutation are separate layers.

## Suggestion Tiers

Ranked candidates are separated without discarding the broader relationship set:

- `hard_link` is the bounded, high-confidence surface intended for permanent-link review. It is limited to the top five candidates per document and currently requires strong evidence such as a declared-symbol mention, a source/test counterpart, or sufficiently strong dependency-neighbour evidence.
- `context` contains weaker or indirect relationships that may still be useful for bounded task context but are not strong enough to recommend as permanent documentation links.

A useful context candidate is not automatically a valid codemap link. Future code-graph evidence and agent-context ranking must preserve this distinction.

## Recorded Precision Baseline

The pinned Space Rocks evaluation contains 150 manually labeled suggestions across 25 documents. The committed report records:

- overall strict precision: **46.7%**;
- precision at rank 1: **60.0%**;
- precision at rank 3: **54.7%**;
- precision at rank 5: **52.8%**;
- `hard_link` strict precision: **64.2%**;
- `hard_link` non-junk acceptance: **95.1%**; and
- `hard_link` sample recall of valid links: **74.3%**.

These metrics apply only to the pinned sample and its curation rules. They are useful for regression and tuning, not a claim that the algorithm has the same quality on unrelated repositories.

The evidence breakdown in that sample shows the strongest strict precision from declared-symbol mentions, source/test counterparts, and dependency neighbours. Exact path mentions have high non-junk acceptance but low strict permanent-link precision, which is why path presence alone should not qualify a candidate as a hard link.

## Development and Evaluation Corpora

Space Rocks remains the curated labeled benchmark because its codemaps have been manually reviewed and the committed sample has stable labels.

Demon Docs' own documentation now includes implementation code maps. Those maps are useful for:

- testing polyglot-independent extraction on a second repository shape;
- exposing parser and corpus assumptions that were accidentally Space Rocks-specific;
- adding deterministic holdout cases during development; and
- evaluating whether recent documentation changes produce stable exports.

They must not be treated as an independent precision benchmark when the same development process authored both the docs and the algorithm adjustments. Self-authored links are development evidence, not unbiased ground truth.

Broader cross-repository work should preserve repository identity, pinned revisions, labels, and provenance rather than pooling unlabeled suggestions into a misleading aggregate metric.

## Decision Persistence

The intended review lifecycle is:

1. generate a candidate and evidence fingerprint;
2. show the document, target, tier, score, and evidence;
3. allow a human to accept or decline it;
4. persist the decision by document, target, and evidence fingerprint;
5. suppress the same declined fingerprint on later runs; and
6. allow reconsideration only when the underlying evidence materially changes and therefore produces a new fingerprint.

This lifecycle is not yet exposed as a complete public command workflow.

## Safety Contract

- Existing codemap targets are never returned as missing-link candidates.
- There is no removal, irrelevance, or automatic cleanup recommendation.
- Evidence creates a reviewable suggestion, never documentation coverage or an authored graph edge.
- A candidate tier is not an automatic write decision.
- Declines remain suppressible through stable fingerprints.
- Identical inputs produce byte-stable candidate ordering, evidence ordering, counts, fingerprints, and reports.
- Unresolved, ambiguous, placeholder, or ignored targets remain explicit diagnostics rather than guessed matches.

## Research Artifacts

- `research/codemap-audit/` records the initial dataset-quality audit.
- `research/codemap-inventory/` contains codemap syntax fixtures and extraction findings.
- `research/codemap-review/` contains reviewed trusted-link artifacts.
- `research/codemap-precision/` contains the pinned labeled sample, curation inputs, evaluation, helper tools, and the Demon Docs self-corpus development comparison.
- `research/codemap-evidence-validation/` documents signal validation work.

## Code map

- `internal/codemap/` — authored codemap parsing, target normalization, diagnostics, and dataset export.
- `internal/codemapcorpus/` — repository paths, dependency facts, symbols, tests, related documents, and Git history adapters.
- `internal/evidence/` — deterministic evidence collection and fingerprints.
- `internal/codemapbench/` — holdout orchestration, ranking, tier assignment, and report export.
- `internal/codemapprecision/` — curated-label evaluation and metric aggregation.
- `internal/app/codemap_benchmark.go` — benchmark CLI contract.
- `internal/app/codemap_precision.go` — precision CLI contract.
- `research/codemap-precision/` — pinned precision benchmark artifacts.
