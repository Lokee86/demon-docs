# Demon Docs Self-Corpus Development Benchmark

This note records a development-only codemap holdout run against Demon Docs' own refreshed documentation. It is not an independent precision benchmark: the project authors control both the documentation links and the algorithm under evaluation.

## Corpus

The refreshed documentation exported:

- 15 scanned Markdown documents;
- 70 code-map entries;
- 65 exact resolved links used by the holdout benchmark;
- 5 resolved glob patterns; and
- 0 extraction diagnostics or unresolved exact targets.

The default deterministic holdout hid 13 of the 65 exact links.

## Main baseline

The main-branch algorithm at `1fa508b` produced:

- 270 raw and unique suggestions;
- 12 of 13 hidden links recovered;
- raw precision: 4.44%;
- recall: 92.31%;
- 17 `hard_link` suggestions and 253 `context` suggestions; and
- recovered links by tier: 1 `hard_link`, 11 `context`.

## Tuning pass 2

The `tuning/codemap-pass-2` algorithm at `aa6eb48` produced:

- 270 raw and unique suggestions;
- 12 of 13 hidden links recovered;
- raw precision: 4.44%;
- recall: 92.31%;
- 28 `hard_link` suggestions and 242 `context` suggestions; and
- recovered links by tier: 0 `hard_link`, 12 `context`.

## Finding

The tuning pass does not improve raw precision or recall on this corpus. It increases the hard-link surface from 17 to 28 candidates while demoting the only recovered hidden link that main classified as a hard link.

That is a negative directional result for this development corpus. Before merging the tuning pass, investigate why its qualification changes promote additional non-held candidates without improving recovery. The likely useful pressure is not broader candidate generation; it is stronger hard-link discrimination and better suppression of indirect sibling, related-document, and co-change noise.

## Reproduction

From an initialized Demon Docs worktree containing the refreshed documentation:

```bash
ddocs codemap export --output .cache/docs-refresh-codemap.json
ddocs codemap benchmark --repo . --dataset .cache/docs-refresh-codemap.json --format json
```

Run the second command with the tuning-branch executable against the same repository and exported dataset. Preserve the repository revision, dataset hashes, seed, and algorithm revision when comparing later results.
