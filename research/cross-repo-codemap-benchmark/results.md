# Cross-repository benchmark results

Frozen algorithm baseline: `aa6eb48c686b0423e104530418b4e9fd32e3aa78`.

| Repository | Mode | Language(s) | Known | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| agent-orchestrator | diagnostic | Go, TypeScript, TSX | 9 | 3 | 1 | 0 | 1 | 33.33% | 3.33% |
| beads-rust | diagnostic | Rust | 14 | 3 | 1 | 0 | 1 | 33.33% | 3.12% |
| bifrost | diagnostic | Go, TypeScript | 14 | 3 | 2 | 1 | 1 | 66.67% | 6.25% |
| gbrain | stress | TypeScript | 357 | 10 | 3 | 0 | 3 | 30.00% | 10.00% |
| genesis | diagnostic | Python | 5 | 1 | 1 | 0 | 1 | 100.00% | 3.33% |
| render-claude-context | primary | TypeScript | 38 | 8 | 6 | 0 | 6 | 75.00% | 25.00% |

The precision column is positive-only holdout precision: unmatched suggestions are counted as false because this corpus has not yet been manually labeled for genuinely new links. It must not be compared with the manually reviewed Space Rocks precision benchmark.

The gbrain result is a stress test: one document owns hundreds of targets, and redacting the authored index removes nearly all topical prose. Primary, diagnostic, and stress results must remain separate.

This run measures recovery only. Cross-repository precision still requires manual labeling of sampled unmatched suggestions.
