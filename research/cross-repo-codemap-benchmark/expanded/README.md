# Step 10 expanded codemap validation

Implementation baseline: `94d3b463962e149b3b44f6f8274b45ef839f3b03`.

This directory freezes the August 28, 2026 closeout validation for the completed Arcana-backed codemap algorithm. It is intentionally separate from the original cross-repository benchmark so historical results remain unchanged.

## Corpus

| Repository | Revision | Resolved authored links | Role |
| --- | --- | ---: | --- |
| Lemonade | `93c888224db3a746986cd748a47d5542c073d3b3` | 22 | primary-sized |
| Seepient | `78b5ea08db72b157d15702ab38bd0a021cf63aea` | 25 | primary-sized |
| cclint | `47cea00a1ef78209232fddcb36e587b9b611a60c` | 5 | diagnostic |
| text-to-speech | `edfafe3e2ce88b636a7ac7baf33bffc4815fb151` | 8 | diagnostic |
| emisso-sii | `cbad5bfc1038b98950f84e0b0d5f0ad2895c5b23` | 3 | diagnostic |

All mappings come from human-authored `AGENTS.md` Key Files sections and resolve against the pinned source revisions.

## Evaluation design

Lemonade and Seepient use ten deterministic holdout seeds each. Each seed hides the same number of authored links in both modes:

```text
fallback
= repository evidence without a usable Arcana snapshot

Arcana
= identical benchmark with current pinned Arcana state available
```

Lemonade hides four links per seed. Seepient hides five. The smaller repositories use one fallback-only diagnostic split because their authored maps are too small for useful multi-seed conclusions.

## Results

| Repository | Fallback | Arcana |
| --- | ---: | ---: |
| Lemonade | 12/40 (30%) | 14/40 (35%) |
| Seepient | 11/50 (22%) | 25/50 (50%) |
| cclint | 1/1 | not run |
| text-to-speech | 1/2 | not run |
| emisso-sii | 0/1 | not run |

Every recovered relationship remained `context` tier. Arcana therefore increased discovery without increasing automatic `hard_link` insertion in this expanded sample.

The Archivist self-map was also used as the convergence fixture. Its first dry-run proposed `added=0 removed=0` and only needed managed-section ownership adoption. After applying that ownership-only change, the second dry-run reported `0 file(s)` to update. This pins the expected mature-map behavior: no semantic expansion and full idempotence after adoption.

## Interpretation limits

This is a recovery benchmark, not a precision benchmark. A hidden authored link is known-positive ground truth, but an unmatched recommendation is not automatically wrong. Do not derive strict precision from these reports.

The retained Space Rocks and frozen cross-repository manual-review corpora remain the evidence for precision and relevance. This expanded run answers the narrower question: does the completed algorithm rediscover useful human-authored relationships on additional repositories, and does Arcana materially improve that recovery?

## Retained artifacts

- `candidates.json` — repository definitions and extraction conventions.
- `discovery.json` — pinned discovery metadata and unresolved examples.
- `corpus/` — normalized authored document-to-code pairs.
- `datasets/` — Archivist benchmark datasets.
- `evaluation.json` — exact multi-seed aggregates, Arcana snapshot IDs, and diagnostic results.

Ignored `checkouts/` and temporary per-seed reports are not retained. The pinned revisions and datasets are sufficient to recreate the benchmark inputs.
