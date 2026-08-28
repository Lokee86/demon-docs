# Cross-repository benchmark results

Frozen algorithm baseline: `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`.

Monolithic per-file indexes are excluded from the calculation corpus. They remain visible as a separate stress result.

## Calculation corpus

| Scope | Repositories | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Primary + diagnostic | 5 | 18 | 11 | 4 | 7 | 61.11% | 7.43% |
| Primary | 1 | 8 | 6 | 2 | 4 | 75.00% | 25.00% |
| Diagnostic | 4 | 10 | 5 | 2 | 3 | 50.00% | 4.03% |

## Index stress corpus

| Scope | Repositories | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Monolithic per-file index | 1 | 10 | 3 | 0 | 3 | 30.00% | 10.00% |

## Repository details

| Repository | Mode | Language(s) | Known | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| agent-orchestrator | diagnostic | Go, TypeScript, TSX | 9 | 3 | 1 | 0 | 1 | 33.33% | 3.33% |
| beads-rust | diagnostic | Rust | 14 | 3 | 1 | 0 | 1 | 33.33% | 3.12% |
| bifrost | diagnostic | Go, TypeScript | 14 | 3 | 2 | 2 | 0 | 66.67% | 6.25% |
| gbrain | stress | TypeScript | 357 | 10 | 3 | 0 | 3 | 30.00% | 10.00% |
| genesis | diagnostic | Python | 5 | 1 | 1 | 0 | 1 | 100.00% | 3.33% |
| render-claude-context | primary | TypeScript | 38 | 8 | 6 | 2 | 4 | 75.00% | 25.00% |

The precision columns are positive-only holdout precision: unmatched suggestions are counted as false because this corpus has not yet been manually labeled for genuinely new links. They must not be compared with the manually reviewed Space Rocks precision benchmark.

The gbrain result is a stress test: one document owns hundreds of targets, and redacting the authored index removes nearly all topical prose. It is not used to tune or summarize ordinary document-to-code behavior.

This run measures recovery only. Cross-repository precision still requires manual labeling of sampled unmatched suggestions.

## Step 10 expanded validation

The completed Arcana-backed algorithm was separately validated at implementation baseline `94d3b463962e149b3b44f6f8274b45ef839f3b03`. The original result above remains frozen and is not recomputed or pooled with this later run.

| Repository | Resolved authored links | Holdout design | Fallback recovery | Arcana recovery |
| --- | ---: | --- | ---: | ---: |
| Lemonade | 22 | 10 seeds x 4 hidden | 12/40 (30%) | 14/40 (35%) |
| Seepient | 25 | 10 seeds x 5 hidden | 11/50 (22%) | 25/50 (50%) |
| cclint | 5 | 1 hidden | 1/1 | not run |
| text-to-speech | 8 | 2 hidden | 1/2 | not run |
| emisso-sii | 3 | 1 hidden | 0/1 | not run |

All recovered links in the expanded run remained `context` tier. Arcana therefore increased rediscovery without expanding the set of links eligible for automatic insertion under the current `hard_link` mutation policy.

The Seepient result is the strongest independent structural signal: identical deterministic splits improved from 11/50 to 25/50 recovered links with Arcana. Lemonade improved more modestly from 12/40 to 14/40. These are recall measurements only; they do not replace the manually labeled precision corpora.

Pinned candidate definitions, normalized corpora, datasets, discovery metadata, snapshot identities, and aggregate results are retained under `expanded/`.

The self-dogfood convergence fixture also completed cleanly: `docs/architecture/codemap-pipeline.md` required one ownership-only adoption with `added=0 removed=0`, then immediately converged to `0 file(s)` on the second dry-run.
