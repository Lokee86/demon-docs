# Space Rocks authored-links codemap precision benchmark

This benchmark measures the usefulness of Demon Docs' current missing-link suggestions when every authored codemap link remains visible. It is a user-facing precision benchmark for genuinely new suggestions, not a hidden-link recovery test.

## Corpus and source pool

- Corpus: Space Rocks at revision `3387c94d10fdb94008f27b404098f3e0c32d911c`.
- Authored links remain visible during generation; `.worktrees/` documents are excluded.
- The current source pool contains 4,493 unique unmatched suggestions across 149 mapped documents. The large source report is regenerated temporarily for validation and is intentionally not committed.

Keeping authored links visible matters: a target already represented in a document's codemap is not a missing link.

## Sampling

[space-rocks-precision-sample-150.json](space-rocks-precision-sample-150.json) is a deterministic sample from that source pool:

- seed: `space-rocks-current-precision-v1`
- method: `balanced-documents-top-5-plus-lower-rank-fill-sha256`
- 25 documents, six suggestions per document
- complete ranks 1–5 plus one lower-ranked fill per document
- coverage across data, devtools, protocol, and services
- 125 rows in ranks 1–5, five in ranks 6–10, five in ranks 11–20, and 15 in ranks 21+

The benchmark preserves every suggestion's document, target, score, evidence, rank, subsystem, and bucket fields exactly as sampled.

## Labels and audit

Each row in [space-rocks-precision-benchmark.json](space-rocks-precision-benchmark.json) was reviewed against the full document and the actual target:

- `valid_missing_link`: direct owner, implementation, definition, or verification necessary to the documented behavior, where a direct codemap link materially improves navigation or coverage.
- `plausible_but_unnecessary`: meaningfully related but indirect, redundant with stronger links, generated/reference-only, overly broad, or not useful enough for a direct link.
- `incorrect`: evidence noise or no meaningful semantic relationship.

Every row records a rationale, document section/reference/excerpt, target reference/excerpt, target kind, and SHA-256. File hashes are hashes of exact bytes at the pinned checkout. Directory hashes are SHA-256 over newline-joined, sorted `repository-relative-path:sha256` entries for Git-tracked files under the directory, so ignored build products cannot affect a pinned-revision audit.

The curation shards under `curation-current/` are canonical audit copies. `reconciliation.md` records the duplicate-review comparison for inputs 01–03: 0 label disagreements across 54 rows; the more specific `labeled-01` audit text was retained where it was better than `output-01`.

## Results

The merged benchmark has 150 rows:

| Label/metric | Result |
|---|---:|
| `valid_missing_link` | 70 |
| `plausible_but_unnecessary` | 62 |
| `incorrect` | 18 |
| Strict precision | 46.67% (70/150) |
| Acceptance precision | 88.00% (132/150) |
| Precision@1 | 60.00% |
| Precision@3 | 54.67% |
| Precision@5 | 52.80% |

Breakdowns are also recorded in [evaluation.json](evaluation.json). Selected exact breakdowns:

| Primary evidence | Rows | Strict | Acceptance |
|---|---:|---:|---:|
| declared symbol mention | 10 | 80.00% | 100.00% |
| test counterpart | 31 | 64.52% | 90.32% |
| dependency neighbor | 41 | 58.54% | 90.24% |
| related-document target | 27 | 40.74% | 70.37% |
| sibling of existing target | 12 | 25.00% | 83.33% |
| exact path mention | 24 | 16.67% | 95.83% |
| unique basename mention | 3 | 0.00% | 100.00% |
| git co-change with existing target | 2 | 0.00% | 100.00% |

By rank bucket: ranks 1–5 are 52.80% strict / 91.20% acceptance (66/125 and 114/125); ranks 6–10 are 40.00% / 100.00% (2/5 and 5/5); ranks 11–20 are 0.00% / 60.00% (0/5 and 3/5); ranks 21+ are 13.33% / 66.67% (2/15 and 10/15). By score bucket: `<1` is 50.00% / 50.00% (1/2 and 1/2), `1-<2` is 50.00% / 75.00% (2/4 and 3/4), `2-<8` is 13.33% / 80.00% (4/30 and 24/30), and `8+` is 55.26% / 91.23% (63/114 and 104/114).

## First tuning pass: hard links and context

The first tuning pass preserves the complete suggestion pool but assigns each candidate one of two product tiers:

- `hard_link`: a bounded direct-link review surface, limited to the top five candidates per document and requiring a declared-symbol mention, a source/test counterpart, or dependency-neighbor evidence with score at least 16.
- `context`: a weaker or indirect relationship retained for bounded agent-context assembly rather than proposed as a permanent codemap link.

Measured against the same 150 labels:

| Tier/metric | Result |
|---|---:|
| Hard-link suggestions | 81 |
| Hard-link strict precision | 64.20% (52/81) |
| Hard-link relevance precision | 95.06% (77/81) |
| Hard-link recovery of labeled valid links | 74.29% (52/70) |
| Hard-link suggestions per sampled document | 3.24 |
| Context suggestions | 69 |
| Context strict precision | 26.09% (18/69) |
| Context relevance precision | 79.71% (55/69) |

On the complete current Space Rocks source pool, 602 of 4,493 candidates are `hard_link` and 3,891 are `context`, averaging 4.04 hard-link candidates per mapped document. Nine of 149 documents have no hard-link candidate.

The positive-only ten-link holdout still recovers 10/10 links because context candidates are not discarded. Four recovered links are in the hard-link tier and six remain context. Therefore the tuning improves the direct recommendation surface without pretending that the broader relevant context has no value. Hard-link-only recall remains deliberately lower and must continue to be measured alongside precision.

## Reproduction

From the Demon Docs repository, with the pinned Space Rocks checkout available:

```text
$env:GOCACHE = Join-Path (Get-Location) '.go-cache'
go run ./cmd/ddocs codemap precision source `
  --repo D:\!bin\space-rocks `
  --exclude-prefix .worktrees `
  --output research/codemap-precision/current-source-report.json

go run ./cmd/ddocs codemap precision sample `
  --suggestions research/codemap-precision/current-source-report.json `
  --count 150 `
  --seed space-rocks-current-precision-v1 `
  --repository space-rocks `
  --revision 3387c94d10fdb94008f27b404098f3e0c32d911c `
  --output research/codemap-precision/space-rocks-precision-sample-150.json

go run ./cmd/ddocs codemap precision evaluate `
  --benchmark research/codemap-precision/space-rocks-precision-benchmark.json `
  --suggestions research/codemap-precision/current-source-report.json

go run ./cmd/ddocs codemap precision evaluate `
  --benchmark research/codemap-precision/space-rocks-precision-benchmark.json `
  --suggestions research/codemap-precision/current-source-report.json `
  --format json `
  --output research/codemap-precision/evaluation.json
```

Validate and merge canonical curation shards with:

```text
python research/codemap-precision/tools/merge_curation.py `
  --source research/codemap-precision/space-rocks-precision-sample-150.json `
  --repository D:\!bin\space-rocks `
  --output research/codemap-precision/space-rocks-precision-benchmark.json `
  --reviewed-at "" `
  research/codemap-precision/curation-current/labeled-01.json `
  research/codemap-precision/curation-current/labeled-02.json `
  research/codemap-precision/curation-current/labeled-03.json `
  research/codemap-precision/curation-current/labeled-04.json `
  research/codemap-precision/curation-current/labeled-05.json `
  research/codemap-precision/curation-current/labeled-06.json `
  research/codemap-precision/curation-current/labeled-07.json `
  research/codemap-precision/curation-current/labeled-08.json `
  research/codemap-precision/curation-current/labeled-09.json
```

## Relationship to the positive-only recall benchmark

The older holdout benchmark deliberately hides known-good authored links and measures whether the system recovers them. That is useful for positive-only recall, but it does not estimate user-facing precision: suggestions that are not in the hidden positive set are treated as false even when they are useful new links. This benchmark leaves authored links visible, labels a deterministic sample of current unmatched suggestions, and reports strict missing-link precision plus a broader non-junk acceptance precision. The two benchmarks answer different questions and should be reported separately.

## Limitations

- This is one repository and one pinned revision with unusually detailed documentation.
- Labels remain reviewer judgments, despite full-document review, source excerpts, target excerpts, and content hashes.
- Sampling is balanced for coverage rather than a natural production traffic distribution.
- Candidate membership, ranking, and metrics can change when evidence extraction or authored links change.
- This benchmark measures precision of current suggestions; it does not establish recall for links not suggested and does not replace the positive-only recall benchmark.

Codex work used `danger-full-access` with hackathon session logging under `.codex-hackathon/sessions/`; those logs are intentionally preserved.
