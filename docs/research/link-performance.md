---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7707-bda4-29732e04ffa1
document_type: general
policy_exempt: false
summary: This document records retained performance measurements for high-fanout target moves, real Space Rocks moves, and repeated full-corpus mass renames.
---
# Markdown Link Performance

Parent index: [Research](./README.md)

## Purpose

This document records retained performance measurements for high-fanout target moves, real Space Rocks moves, and repeated full-corpus mass renames.

## Overview

The measurements expose scanning, planning, storage, and generated-write regressions under realistic and synthetic link volumes. They are engineering evidence for the tested hardware, corpus, and implementation revision rather than universal latency guarantees.

## Research status

Recorded benchmark evidence. New performance changes should preserve the original artifacts and add comparable runs rather than overwriting historical results.

Demon Docs records link-reconciliation benchmarks separately from correctness tests. These results are host-specific engineering measurements, not guaranteed performance limits. Unless noted otherwise, the measurements below were taken on the Windows development host on July 19, 2026.

## High-Fanout Target Move

`BenchmarkHighFanoutTargetMove` creates 250 Markdown files that all link to one target, establishes the `.ddocs/` baseline, renames the target, and measures reconciliation plus application of every generated source rewrite.

Before generated writes used a bounded worker pool, repeated runs measured:

| Phase | Observed range |
|---|---:|
| Filesystem rewrites | 817–881 ms |
| Generated-source refresh | 30–32 ms |
| `.ddocs` publication | about 38 ms |
| Complete apply phase | 885–954 ms |
| Complete benchmark operation | 1.025–1.095 s |

Commit `12856e3` introduced a 16-worker bounded rewrite pool. Repeated runs after that change measured:

| Phase | Observed range |
|---|---:|
| Filesystem rewrites | 261–299 ms |
| Generated-source refresh | 11–12 ms |
| Complete apply phase | 322–358 ms |
| Complete benchmark operation | 505–586 ms |

The worker pool reduced filesystem rewrite time by roughly three times, complete apply time by roughly 2.7 times, and the complete benchmark operation by roughly two times. The plan remains deterministic: workers only apply already-planned, independently verified source rewrites.

Run the current synthetic benchmark with:

```bash
go test ./internal/links \
  -run '^$' \
  -bench '^BenchmarkHighFanoutTargetMove$' \
  -benchmem \
  -count=5
```

## Real Space Rocks Target Move

An earlier phase-timing run used a copied Space Rocks documentation corpus and moved `services/game-server/!INDEX.md`.

The target had 106 incoming link occurrences and required 96 Markdown source files to be rewritten.

| Reconciliation phase | Time |
|---|---:|
| State load | 57.1 ms |
| Inventory build | 12.4 ms |
| Planning | 888.4 ms |
| Total reconciliation | 957.9 ms |

| Application phase | Time |
|---|---:|
| Filesystem rewrites | 898.3 ms |
| Generated-source verification and refresh | 44.9 ms |
| `.ddocs` publication | 136.8 ms |
| Total application | 1.08 s |

This run predates the bounded parallel rewrite optimization and is retained as a real-corpus baseline rather than a direct comparison with the synthetic fixture.

## Full Space Rocks Mass Rename

The mass-rename stress test copied the Space Rocks documentation corpus, initialized and converged its link state, renamed every Markdown file in place, repaired all affected links, verified the result, and then repeated the entire rename process a second time.

Corpus and workload:

- 346 total copied files;
- 341 Markdown files renamed per pass;
- 340 Markdown source files rewritten per pass;
- 3,717 link destinations repaired per pass; and
- five independent timing iterations, all validated against the expected repair counts.

### Actual `ddocs fix -l` time

| Rename pass | Median | P95 | Mean |
|---|---:|---:|---:|
| First mass rename | 1.928 s | 1.993 s | 1.944 s |
| Second mass rename | 1.980 s | 2.013 s | 1.987 s |

Observed fix throughput:

| Rename pass | Source files/s | Link repairs/s |
|---|---:|---:|
| First mass rename | 176.32 | 1,927.61 |
| Second mass rename | 171.74 | 1,877.49 |

The updater therefore repaired more than 3,700 links across 340 source files in approximately two seconds, including CLI startup, scanning, reconciliation, source writes, `.ddocs` publication, and captured command output. Compilation and corpus copying were excluded from the fix measurement.

### Complete validation cycle

A complete cycle includes the filesystem rename, read-only pre-check, repair, post-check, and an idempotence pass.

| Scenario | Median | P95 |
|---|---:|---:|
| First complete rename cycle | 5.677 s | 5.758 s |
| Second complete rename cycle | 5.722 s | 5.792 s |
| Both rename cycles | 11.305 s | 11.550 s |
| Entire harness, including copy, initialization, and baseline convergence | 16.755 s | 17.321 s |

Both post-repair idempotence passes updated zero files. The remaining five diagnostics were inherited from the converged copied corpus; the mass renames introduced no additional unresolved links.

## Retained Evidence

The reproducible mass-rename harness and raw results are retained under:

- `research/run_space_rocks_mass_rename.py` — correctness and repeated-rename harness;
- `research/mass-rename-results/` — command logs, rename maps, and correctness summary;
- `research/benchmark_space_rocks_mass_rename.py` — five-iteration timing harness; and
- `research/mass-rename-timing/` — raw samples, JSON summary, and Markdown timing table.

Historical move measurements and their provenance are summarized under `research/link-performance/README.md`.

## Interpretation

The benchmarks exercise different workloads and should not be collapsed into one performance claim:

- the synthetic high-fanout benchmark isolates one moved target with many inbound links;
- the real Space Rocks move shows behavior on an irregular documentation graph; and
- the mass rename exercises repository-wide target identity recovery and thousands of link rewrites.

Correctness remains the release gate. Timing measurements are retained to reveal regressions in scanning, planning, filesystem writes, verification, and `.ddocs` publication.

## Limitations

Timing varies with filesystem, antivirus, hardware, repository size, link distribution, cache state, and operating system. Comparisons are most useful when the corpus and environment remain controlled.

## Related docs

- [Research](README.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)
- [Roadmap](../planning/roadmap.md)

## Notes

The benchmark suite should continue reporting phase timings, not only total duration, so regressions can be assigned to scanning, planning, storage, or writes.
