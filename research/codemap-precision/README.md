# Space Rocks codemap precision benchmark

This benchmark measures the quality of ordinary missing-link suggestions from the current Demon Docs pipeline. It complements, and does not modify, the positive-only recall benchmark in [`research/codemap-review/space-rocks-trusted-links.json`](../codemap-review/space-rocks-trusted-links.json).

## Corpus and sampling

- Corpus: Space Rocks revision `3387c94d10fdb94008f27b404098f3e0c32d911c` (10 reviewed documents, 51 trusted links).
- Source run: all 51 trusted links were hidden with seed `space-rocks-precision-v1`; the run produced 267 unique suggestions, including 24 recovered trusted links and 243 ordinary unmatched suggestions.
- Sample: 150 unique document-target pairs in [`space-rocks-precision-benchmark.json`](space-rocks-precision-benchmark.json).
- Sampling: every document’s top five was retained where available; the remainder was selected deterministically with SHA-256 least-represented fill across document area, score bucket, primary evidence kind, and rank bucket.
- Labels were assigned after inspecting the referenced document section and target file or directory. Each row records document and target references, excerpts, target kind, a target fingerprint, evidence, and a concise rationale.

The three labels are mutually exclusive: `valid_missing_link` is correct for strict precision, `plausible_but_unnecessary` is accepted only for acceptance precision, and `incorrect` is junk.

## Reproduction

Generate the source suggestion report without changing Space Rocks:

```text
go run ./cmd/ddocs codemap benchmark --repo D:\!bin\space-rocks --trusted-links research\codemap-review\space-rocks-trusted-links.json --holdout-count 51 --seed space-rocks-precision-v1 --format json --output tmp\codemap-precision\full-hidden.json
```

Evaluate the labeled sample:

```text
go run ./cmd/ddocs codemap precision --benchmark research\codemap-precision\space-rocks-precision-benchmark.json --suggestions tmp\codemap-precision\full-hidden.json
```

The evaluator verifies that every sampled pair, score/evidence payload, and rank still exists in the source report before calculating metrics.

## Results

Evaluation date: 2026-07-19.

| metric | result |
| --- | ---: |
| sample size | 150 |
| valid missing link | 98 |
| plausible but unnecessary | 26 |
| incorrect | 26 |
| overall precision | 65.33% |
| acceptance precision | 82.67% |
| precision@1 | 80.00% |
| precision@3 | 73.33% |
| precision@5 | 74.00% |

### Precision@k by document

| document | @1 | @3 | @5 |
| --- | ---: | ---: | ---: |
| data-sync-and-ssot-pipeline | 0.00% | 0.00% | 0.00% |
| observability-contract | 100.00% | 66.67% | 80.00% |
| devtools-window | 100.00% | 100.00% | 100.00% |
| spawn-tools | 100.00% | 100.00% | 100.00% |
| player-data-http-api | 100.00% | 100.00% | 80.00% |
| realtime-websocket-protocol | 100.00% | 100.00% | 100.00% |
| websocket-connection-lifecycle | 100.00% | 100.00% | 100.00% |
| diagnostic-aggregator runtime-and-report-flow | 0.00% | 0.00% | 0.00% |
| player-respawn | 100.00% | 66.67% | 80.00% |
| profile-stats-flow | 100.00% | 100.00% | 100.00% |

### Breakdowns

| primary evidence kind | n | valid | accepted | precision | acceptance |
| --- | ---: | ---: | ---: | ---: | ---: |
| declared_symbol_mention | 58 | 55 | 58 | 94.83% | 100.00% |
| exact_path_mention | 52 | 42 | 49 | 80.77% | 94.23% |
| unique_basename_mention | 40 | 1 | 17 | 2.50% | 42.50% |

| score bucket | n | valid | accepted | precision | acceptance |
| --- | ---: | ---: | ---: | ---: | ---: |
| `<1` | 13 | 1 | 2 | 7.69% | 15.38% |
| `1-<2` | 69 | 35 | 55 | 50.72% | 79.71% |
| `2-<8` | 39 | 34 | 38 | 87.18% | 97.44% |
| `8+` | 29 | 28 | 29 | 96.55% | 100.00% |

| rank bucket | n | valid | accepted | precision | acceptance |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1-5 | 50 | 37 | 46 | 74.00% | 92.00% |
| 6-10 | 33 | 22 | 30 | 66.67% | 90.91% |
| 11-20 | 34 | 20 | 26 | 58.82% | 76.47% |
| 21+ | 33 | 19 | 22 | 57.58% | 66.67% |

### Sampling coverage

The sample covers all 10 documents, all four score buckets, all four rank buckets, all three observed primary evidence kinds, and four document areas: data (26), devtools (41), protocol (33), and services (50). Subsystem counts are recorded in the evaluator’s JSON output and the row-level artifact.

## Limitations

This is a curated sample, not an unbiased estimate over every repository path. It is drawn from a full-hidden run of the 51-link trusted corpus, so it evaluates the current evidence pipeline under that benchmark setup and does not measure recall or unseen-document behavior. Directory suggestions are retained as ordinary candidates but are judged conservatively. Labels are auditable and source-pinned, but this artifact has one review pass rather than independent inter-rater agreement. The existing recall benchmark remains the appropriate measure for recovery of known links.
