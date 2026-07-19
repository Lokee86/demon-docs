# Cross-repository benchmark results

Frozen algorithm baseline: `6ea39964a77919c3a6228b475904e9c530a16a4d`.

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
