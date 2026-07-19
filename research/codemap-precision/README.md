# Space Rocks codemap precision benchmark

This benchmark measures whether Demon Docs' current missing-link suggestions are useful, not whether they rediscover links deliberately hidden from the system.

## Why this benchmark exists

The original holdout benchmark is a recall benchmark. It hides known-good links and measures whether Demon Docs can recover them. Unknown suggestions are counted as false, so its apparent precision is not a trustworthy user-facing precision measurement.

This benchmark labels ordinary current-state suggestions as:

- `valid_missing_link`: a clear omitted semantic edge that belongs in the document's code map.
- `plausible_but_unnecessary`: genuinely related, but indirect, redundant, broad, contextual, or optional.
- `incorrect`: incidental or misleading evidence; the target does not materially support the documented subject.

Strict precision counts only `valid_missing_link`. Acceptance precision counts both valid and plausible suggestions as non-junk.

## Source corpus

- Repository: Space Rocks
- Revision: `3387c94d10fdb94008f27b404098f3e0c32d911c`
- Existing authored codemap links remain visible during suggestion generation.
- In-repository `.worktrees/` documents are excluded.
- Current-state source pool: 4,493 unlinked suggestions across 149 mapped documents.

Keeping every authored link visible is essential: a target already present in a document's codemap cannot be labeled as a missing link.

## Sampling

The committed benchmark contains 150 deterministic suggestions selected from the current-state source pool.

- 25 documents
- six suggestions per document
- complete ranks 1–5 for every sampled document
- one additional lower-ranked suggestion per document
- coverage across data, devtools, protocol, and service documentation
- balancing across subsystem, score bucket, evidence kind, and rank

Seed: `space-rocks-current-precision-v1`

Sampling method: `balanced-documents-top-5-plus-lower-rank-fill-sha256`

## Review method

Each suggestion is reviewed against both the full document and the actual target in Space Rocks. Each benchmark row records:

- the label and an evidence-specific rationale
- document section and excerpt/paraphrase
- target symbol or location and excerpt/paraphrase
- target kind
- target content hash

Curation shards are validated before merge. Validation rejects changed suggestion fields, duplicate or missing document-target pairs, invalid labels, empty audit fields, and stale target hashes.

Codex use for this work was run in `danger-full-access` with hackathon session logging enabled under `.codex-hackathon/sessions/`. Some review shards were completed by Hermes subagents after the Codex connector repeatedly exceeded its response window. The final merged artifact is validated by local tooling regardless of reviewer.

## Results

Results are generated from the committed labeled benchmark with:

```text
ddocs codemap precision evaluate \
  --benchmark research/codemap-precision/space-rocks-precision-benchmark.json \
  --suggestions research/codemap-precision/current-source-report.json
```

The reproducible source report is intentionally not committed because it is large and can be regenerated with:

```text
ddocs codemap precision source \
  --repo D:\!bin\space-rocks \
  --exclude-prefix .worktrees \
  --output research/codemap-precision/current-source-report.json
```

### Measured results

| Metric | Result |
|---|---:|
| Strict precision | 46.67% |
| Acceptance precision | 88.00% |
| Precision@1 | 60.00% |
| Precision@3 | 54.67% |
| Precision@5 | 52.80% |

Label distribution:

- 70 `valid_missing_link`
- 62 `plausible_but_unnecessary`
- 18 `incorrect`

The previous 3.45% figure was therefore not user-facing precision. It was a positive-only holdout score that treated every unknown suggestion as false. The labeled benchmark shows that most suggestions are genuinely related, but only about half are strong enough to be clear missing codemap entries.

### Evidence findings

| Primary evidence | Strict precision | Acceptance precision |
|---|---:|---:|
| Declared symbol mention | 80.00% | 100.00% |
| Test counterpart | 64.52% | 90.32% |
| Dependency neighbor | 58.54% | 90.24% |
| Related-document target | 40.74% | 70.37% |
| Sibling of existing target | 25.00% | 83.33% |
| Exact path mention | 16.67% | 95.83% |
| Unique basename mention | 0.00% | 100.00% |

Exact path and basename mentions are often real contextual relationships, but can appear in non-ownership or boundary statements. They should not be treated as automatic missing links without semantic qualification. Declared symbols, focused test counterparts, and strong dependency relationships are substantially better strict-link evidence.

Rank quality drops sharply after the first five suggestions. Ranks 1–5 measured 52.80% strict precision; ranks 11–20 measured 0%, and ranks 21+ measured 13.33%. The current product surface should therefore remain heavily capped rather than exposing the long tail.

## Limitations

- Space Rocks is one repository with unusually detailed LLM-assisted documentation.
- Labels are reviewer judgments, even with source excerpts and hashes.
- This benchmark measures the current algorithm and repository revision; candidate membership can change after evidence or ranking changes.
- Precision and recall remain separate benchmarks. Improvements should be checked against both.
