# Cross-repository codemap benchmark

This benchmark is intentionally separate from `research/codemap-precision/`.

The Space Rocks precision corpus measures tuning behavior on one deeply documented repository. This corpus asks a different question: whether the same frozen suggestion rules generalize across unrelated repositories and documentation conventions.

## Corpus classes

Candidates are kept in separate classes rather than pooled blindly:

- `native_*`: documentation stored with the source repository and maintained as project guidance.
- `generated_*`: documentation generated from a repository or stored in a paired wiki. These mappings are useful ground truth but must be reported independently from human-maintained mappings.

A repository qualifies only when selected documents contain explicit paths that resolve in the pinned code checkout. General architecture prose without resolvable document-to-code pairs is rejected.

## Reproducible workflow

`candidates.json` records the reconnaissance shortlist and extraction mode for each convention.

```text
python research/cross-repo-codemap-benchmark/tools/prepare_corpus.py
python research/cross-repo-codemap-benchmark/tools/build_benchmark_inputs.py
python research/cross-repo-codemap-benchmark/tools/run_benchmarks.py
```

The first command shallow-clones repositories into the ignored `checkouts/` directory, pins revisions, validates explicit paths, writes the compact `discovery.json`, and writes normalized resolved pairs under `corpus/`. The complete discovery trace stays under the ignored `runs/` directory.

The second command converts eligible normalized corpora into Demon Docs dataset schema without changing the source repositories. `benchmark-plan.json` pins the algorithm baseline, repository revisions, corpus sizes, and holdout counts.

The third command runs the frozen algorithm named by `benchmark-plan.json`. Per-repository reports are stored under `reports/`; `evaluation.json` and `results.md` summarize the run. `--checkout-root` can reuse an existing ignored checkout corpus from another worktree.

## Benchmark modes

- `primary`: large enough and structurally suitable for the first held-out evaluation.
- `diagnostic`: useful language or repository diversity, but too small for standalone conclusions.
- `stress`: a valid but pathological convention that tests a specific structural limit.
- `extraction_only`: explicit links can be normalized, but authored-section redaction is not yet safe enough for a leakage-free holdout run.
- `discovery_only`: the current checkout no longer contains the expected mapping convention or did not yield resolvable links.

The gbrain per-file index is a stress case rather than a normal primary corpus: one document owns hundreds of explicit targets, and removing its authored index also removes nearly all topical evidence. Its result must not be pooled with ordinary multi-document codemaps.

## First frozen run

The first run covers six repositories and TypeScript, Go, Rust, Python, TSX, and mixed Go/TypeScript layouts.

- `render-claude-context` recovered 6 of 8 hidden links.
- The four small diagnostic corpora recovered 1/3, 1/3, 2/3, and 1/1 hidden links.
- The gbrain stress corpus recovered 3 of 10 hidden links.
- Only one recovered link qualified as `hard_link`; the remaining recovered links were context-tier suggestions.

These are recall checks, not cross-repository precision measurements. The reported positive-only precision treats every unmatched suggestion as false because the corpus only labels existing authored links. It cannot identify good new links. A real wider precision claim still requires manual review samples from multiple repositories.

## Repeated-reference tuning pass

The first cross-repository tuning pass excludes the monolithic `gbrain` index from all ordinary calculations and keeps it as a separate stress result. The calculation corpus therefore contains the primary and diagnostic modes only.

The tuning promotes an already-mentioned target to `hard_link` only when the document references its exact path at least twice and dependency or declared-symbol evidence independently corroborates the relationship. A single path mention and filename-only counterpart evidence remain `context`.

Against the frozen holdouts:

- the calculation corpus remains 11/18 recovered, so overall discovery recall is unchanged;
- hard-tier recoveries increase from 1 to 4;
- the primary corpus remains 6/8 recovered, with hard-tier recovery increasing from 0 to 2;
- diagnostic hard-tier recovery increases from 1 to 2; and
- the separate index stress result remains 3/10, all in context.

This is a tier-confidence improvement rather than a recall increase. It demonstrates that a narrow rule learned from multiple repositories can move known-good relationships into the direct-review surface without lowering held-out recovery.

## Incidental-target negative-evidence pass

The next pass used the manually reviewed four-repository tuning split and reserved Bifrost for validation. It suppresses only unsupported dependency lockfiles and weak basename-only matches to deeply nested content or workflow infrastructure. Explicit, structural, or semantic corroboration preserves the candidate.

Results at algorithm commit `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`:

- all 83 valid and 34 plausible suggestions in the frozen review sample remain;
- all four reviewed incorrect suggestions are removed;
- Bifrost remains unchanged and recovers 2/3 hidden links, both in the hard tier;
- the ordinary calculation corpus remains 11/18 recovered with four hard recoveries; and
- the separate gbrain index stress result remains 3/10, all context.

A generic test-counterpart demotion was rejected after Space Rocks validation showed a large loss of valid hard-link coverage with no precision gain. The final pass contains no broad test-target penalty.

## Acceptance requirements

A primary candidate should normally provide:

- multiple selected documents or independently meaningful document sections;
- at least twenty distinct resolvable document-to-code pairs;
- a pinned code revision and, when separate, a pinned docs revision;
- enough repository diversity to avoid evaluating only one language or project family; and
- a clear provenance class so generated mappings are never presented as human judgment.

Small but structurally distinct repositories remain as diagnostic fixtures rather than being pooled into the primary result.
