---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7dff-9331-c53ab2c88526
document_type: general
policy_exempt: false
summary: This document describes the Demon Docs test commands, regression fixture matrix, package coverage, performance benchmarks, research validation, CI, and release gates.
---
# Testing and Fixtures

Parent index: [Development](./INDEX.md)

## Purpose

This document describes the Demon Docs test commands, regression fixture matrix, package coverage, performance benchmarks, research validation, CI, and release gates.

## Overview

Testing is organized around deterministic behavior and source preservation. Focused package tests protect ownership boundaries, repository-level fixtures protect complete command behavior, and retained benchmarks expose performance or ranking regressions without turning research samples into universal claims.

Demon Docs is covered by focused Go package tests, filesystem integration tests, CLI fixture regressions, codemap benchmark artifacts, and cross-platform CI. Go is the sole implementation and supported runtime.

## Test Commands

For an install smoke check:

```bash
go install ./cmd/ddocs
go install ./cmd/demon
ddocs --help
ddocs --version
demon --help
demon --version
```

Run the complete local release gate from the repository root:

```bash
make release-check
```

The individual gates are:

```bash
make test-go
make regression
make vet
make build
make smoke
```

A direct full-suite run is:

```bash
go test ./... -count=1
```

## CLI Help Coverage

Help tests cover the complete public command tree, not only top-level command names. Every public command and nested subcommand must:

- return exit code `0` for `-h` and `--help`;
- write help to stdout without runtime or repository side effects;
- show usage scoped to the requested command;
- list every accepted flag, required positional identifier, and important default;
- state mutation, hash-guard, persistence, or output behavior when those details affect safe use; and
- remain reachable through both `ddocs demon ...` and the installed `demon ...` alias where applicable.

`internal/app/help_test.go` owns top-level command contracts. `internal/app/help_nested_test.go` owns review, precision, and feeder subcommand routing. `cmd/demon/main_test.go` verifies that bare `demon` and `demon --help` open daemon-specific help while shared version handling remains available.

A parent summary is not acceptable output for a requested nested command. For example, these must remain distinct:

```bash
ddocs suggestions --help
ddocs suggestions select --help
ddocs codemaps precision --help
ddocs codemaps precision sample --help
```

## Fixture Regression Matrix

`make regression` runs the Go CLI fixture matrix. It builds the binary once, then runs each retained scenario through `fix`, verifies a clean successful `check`, runs `fix` again, and requires the complete fixture tree to be byte-identical after the first and second fixes.

The scenarios cover defaults; custom headings, markers, drafts, and editable extensions; direct-to-stub transitions; stub graduation; unique and ambiguous file or folder moves; stale entry removal; and malformed managed blocks. Focused byte-level tests also verify that fenced Markdown examples are not treated as real managed structure and that the original final-newline state is preserved.

## Link-Reconciliation Coverage

Focused link tests cover:

- ordinary Markdown links and images;
- angle-wrapped destinations and titles;
- reference definitions plus explicit and collapsed label uses;
- undefined reference-label diagnostics;
- path-based wiki links, aliases, embeds, extensionless targets, and ambiguity;
- supported local HTML `href`, `src`, and `poster` attributes;
- ignored fenced code and inline code spans;
- file identity, fingerprints, case-only changes, moves, and ambiguous candidates;
- external relative and absolute filesystem targets;
- atomic generated rewrites and concurrent source changes;
- bounded rewrite-worker behavior and deterministic plans;
- initial baseline and incremental storage timing; and
- watch-event suppression for expected generated writes;
- stateless file/directory move planning, case-only renames, ambiguity refusal, and rollback; and
- review-ledger recording for deterministic and selected repairs.

## Document-Health Coverage

Orphan health tests verify that link-enabled `check`:

- reports normal managed Markdown documents without inbound links;
- excludes configured folder indexes and draft documents;
- ignores self-links;
- does not count inbound links originating from indexes or drafts;
- accepts meaningful inbound links from normal repository Markdown sources; and
- emits deterministic path-sorted diagnostics.

Focused coverage lives in `internal/app/orphans_test.go` and `internal/app/orphans_integration_test.go`.

## Frontmatter and Document-Format Coverage

Frontmatter tests cover YAML/TOML parsing, leading-block protection, deterministic rendering, configured field types, unknown-field modes, conditional requirements, immutable restoration, duplicate document IDs, docs-root containment, CRLF preservation, and idempotent repair across a path move. Document-policy tests cover protected Markdown syntax, schema selection, hierarchy and alias validation, reordering and nesting, placeholder creation, explicit ignore/merge/delete operations, stable-ID heading renames, document-specific schema invalidation, and schema-aware codemap placement.

Focused coverage lives in `internal/frontmatter/`, `internal/documentpolicy/`, `internal/app/frontmatter_integration_test.go`, `internal/app/document_policy_integration_test.go`, and `internal/app/document_policy_migration_test.go`.

## Review-Ledger Coverage

Review tests cover persisted decline decisions, stale evidence fingerprints, ambiguous link suggestions, codemap selection, applied-change events, Git-object append behavior, undo depth and age, whole-run preflight, repair-level undo, blocks, unblocks, and refusal to overwrite later edits.

Focused coverage lives in `internal/review/`, `internal/links/review_integration_test.go`, `internal/app/review_cli_test.go`, and `internal/codemap/insert_test.go`.

## Link Performance Benchmarks

Link performance is measured at both package and full-CLI levels:

- `BenchmarkInitialIndexing` measures a fresh 250-file repository;
- `BenchmarkSingleFileIncrementalUpdate` measures one changed source in a converged 500-file repository;
- `BenchmarkHighFanoutTargetMove` measures a target move requiring 250 source rewrites; and
- the copied Space Rocks mass-rename harness renames 341 Markdown files and repairs 3,717 links per pass.

The current recorded mass-rename median is 1.928 seconds for the first `ddocs fix -l` pass and 1.980 seconds for a repeated pass. The synthetic high-fanout move improved from 885–954 ms to 322–358 ms for the complete apply phase after generated source writes moved to a bounded 16-worker pool.

See [Markdown Link Performance](../research/link-performance.md) for the complete phase breakdown, throughput, methodology, historical comparison, and retained raw artifacts.

## Repository-Demon Coverage

Daemon tests cover:

- exactly-one fresh owner claims and stale-owner recovery;
- feeder registration, reuse, expiry, counting, and removal;
- shutdown requests and grace periods;
- read-only status behavior;
- bounded log rotation;
- Bash and PowerShell hook contracts;
- linked-worktree discovery and first-mutating-entry bootstrap; and
- persistent enable and disable behavior.

`TestClaimAllowsExactlyOneOwner` has now failed intermittently in more than one full-suite run by allowing a second owner after the first lease aged during the concurrent test. The same test passed 50 focused repetitions. Treat this as an unresolved suite-context reliability issue: retain focused stress coverage, reproduce the timing interaction, and do not call daemon ownership fully settled until the cause is fixed or the test contract is corrected.

## Codemap Tests and Benchmarks

Codemap coverage now spans four distinct responsibilities:

```text
extraction and target resolution
production evidence admission and ranking
explicit managed-section execution
controlled benchmark and precision evaluation
```

Focused production tests cover configured heading recognition, fenced-heading exclusion, duplicate-section failure, whole-section adoption, legacy partial-region unification, Space Rocks-style fenced rendering, bullet-prefix preservation, schema-gated creation, missing-section skip, idempotency, shared decline suppression, opt-in pruning, file/directory CLI scope, dry-run, check, transaction publication, and fix convergence.

The production ranker lives in `internal/codemaprecommend`. Benchmark packages import that same implementation, preventing a separate benchmark-only scoring path.

The committed Space Rocks precision sample contains 150 labels. The current retained baseline is:

- `hard_link` recommendations: **68**;
- `hard_link` strict precision: **75.00%** (51/68);
- `hard_link` relevance: **98.53%** (67/68);
- labeled-valid `hard_link` recovery: **72.86%** (51/70);
- `context` recommendations: **82**; and
- canonical hidden-link holdout: **10/10 recovered**.

The ordinary cross-repository holdout recovers **11/18** links (**61.11%**). The frozen cross-repository precision review retains **83** valid missing links and **34** plausible context links while suppressing all **4** demonstrated incorrect candidates, producing **70.94%** strict precision and **100%** relevance for that fixed reviewed sample only.

Run repository holdouts with:

```bash
ddocs codemaps benchmark --repo /path/to/repository --format json
```

Use `ddocs codemaps precision --help` for generation, sampling, and evaluation commands. Benchmark reports must retain the repository, revision or dataset, seed, holdout rules, labels, and command inputs needed to reproduce them.

Verify production managed execution separately:

```bash
ddocs codemaps inspect --root docs/architecture/example.md
ddocs codemaps fix --root docs/architecture/example.md --dry-run
ddocs codemaps fix --root docs/architecture/example.md
ddocs codemaps check --root docs/architecture/example.md
```

The second `fix` must be a no-op. Inspect, check, and dry-run must not write. A normal `ddocs check` does not include codemap-generation convergence.

Demon Docs' own code maps are a second development corpus. They are appropriate for extraction, portability, deterministic holdouts, and production execution tests, but they are not an independent precision benchmark because the same development process authored the docs and tunes the algorithm.

## Behavioral Contract Verification

The [Behavioral Contract Matrix](behavioral-contract-matrix.md) maps critical source-preservation, mutation, persistence, concurrency, benchmark, CLI, and compatibility guarantees to their canonical owners and focused tests.

Use it when changing a durable invariant. Package coverage alone is not sufficient: a stateful flow may contain several independently protected contracts inside one package.

## Documentation Coverage Verification

Documentation changes are verified at four levels:

```text
structure and indexes
local links and orphan reachability
implementation ownership coverage
behavior-to-test contract coverage
```

Run the repository's own documentation reconciliation:

```bash
go run ./cmd/ddocs fix --docs
go run ./cmd/ddocs check --docs
go run ./cmd/ddocs check --links
```

Then review [Documentation Coverage Map](documentation-coverage.md) against the current immediate directories under `cmd/` and `internal/`. Every production package must have a canonical current owner, and every public command family must have an exact reference or task workflow.

The coverage audit is semantic rather than a generated line-count target. A package name appearing in a code map does not count when the linked document fails to explain its responsibility, flow, non-ownership boundary, and tests.

Structural review should also confirm that normal documents contain one parent index, purpose, overview, related-docs section, and notes section. Architecture pages additionally require code root, responsibilities, does-not-own boundaries, code map, and tests. Guides require prerequisites, expected result, and failure/recovery guidance.

## Continuous Integration

`.github/workflows/ci.yml` runs:

- the complete Go suite, including `./tests`, on Linux and Windows;
- `go vet ./...`;
- both executable builds; and
- basic CLI smoke tests for `ddocs` and `demon`.

## Release Requirements

A release is eligible only when all CI jobs pass. In particular:

- Linux and Windows Go tests are green;
- the CLI fixture matrix is green;
- `go vet`, both executable builds, and CLI smoke checks pass;
- repeated reconciliation is byte-identical;
- `check` remains non-mutating;
- link and codemap reports remain deterministic for pinned fixtures; and
- repeated daemon ownership and feeder tests are free of reproducible timing failures.

## Dummy Docs Fixture Generator

`docs/development/make-dummy-docs.sh` creates a disposable nested documentation tree for manual stress testing.

```bash
./docs/development/make-dummy-docs.sh
```

Useful environment knobs include `ROOT_DIR`, `RECREATE`, `EXTENSIONS`, and the `MIN_*` / `MAX_*` folder and file counts. The default output directory is `dummy-docs/`, which is ignored by the repository.

A simple smoke flow is:

```bash
./docs/development/make-dummy-docs.sh
ddocs fix --root dummy-docs
ddocs check --root dummy-docs
```

## Context-Benchmark Research

Agent-context claims require a separate empirical harness beyond package tests and codemap holdouts. The retained research uses authentic historical OSS tasks, paired no-context and context-injected conditions, independent code/documentation quality assessment, leakage controls, and an intentionally constructed harness control.

Corpus preparation and deterministic harness validation can proceed without paid repeated model runs. See [Context-Injection Benchmarking](../research/context-injection-benchmarking.md) and `research/context-benchmarking/`.

## Code map

- `.github/workflows/ci.yml` — Linux and Windows CI.
- `tests/regression_test.go` — CLI fixture regression orchestration.
- `tests/regression_fixtures_test.go` — fixture-tree assertions.
- `internal/links/*_test.go` — link syntax, state, move, review integration, rewrite, concurrency, and timing coverage.
- `internal/review/*_test.go` — review history, policy replay, and undo coverage.
- `internal/app/help_test.go` and `help_nested_test.go` — top-level and scoped nested CLI help contracts.
- `cmd/demon/main_test.go` — standalone demon alias argument normalization.
- `internal/app/move_test.go` — stateless move CLI coverage.
- `internal/app/orphans_test.go` and `orphans_integration_test.go` — document-health rules and command behavior.
- `internal/app/review_cli_test.go` — suggestion and applied-change CLI coverage.
- `internal/codemap/insert_test.go` — compatibility selected-candidate insertion coverage.
- `internal/codemap/managed_test.go` — unified section adoption, schema gating, rendering preservation, removal, and idempotency.
- `internal/codemaprecommend/suggestions_test.go` — production ranker behavior after benchmark extraction.
- `internal/codemaprun/build_test.go` — production plans, decline suppression, additions, pruning, and rewrite construction.
- `internal/app/codemap_execute_test.go` — command aliases, required roots, dry-run, check, fix, and convergence.
- `internal/filetxn/apply_test.go` — shared batch preflight, atomic replacement, digest verification, and guarded rollback.
- `internal/frontmatter/*_test.go` — parser, evaluator, schema validation, immutable state, and repair planning.
- `internal/documentpolicy/*_test.go` — schema selection, body enforcement, explicit conflict resolution, migration, creation, and codemap placement.
- `internal/watch/*_test.go` — watcher filters, scheduling, and filesystem events.
- `internal/demon/runtime_test.go` — owner and feeder lifecycle coverage.
- `internal/app/demon_test.go` — daemon CLI and shell integration coverage.
- `internal/codemap/*_test.go` — extraction, deterministic datasets, and managed-section reconciliation coverage.
- `internal/codemaprecommend/*_test.go` — production ranking, tier, and negative-evidence coverage.
- `internal/codemaprun/*_test.go` — production execution planning coverage.
- `internal/codemapbench/*_test.go` — holdout, compatibility, classification, and report coverage using the production ranker.
- `internal/codemapprecision/*_test.go` — curated precision evaluation coverage.
- `research/codemap-precision/` — pinned labels and evaluation artifacts.
- `research/context-benchmarking/` — historical-task and harness research artifacts.
- `research/link-performance/` — historical high-fanout and real-corpus move measurements.
- `research/mass-rename-results/` — repeated whole-corpus rename correctness logs.
- `research/mass-rename-timing/` — five-run mass-rename timing samples and summaries.

## Failure modes

Repository-wide tests must exclude nested `.worktrees/` and local generated outputs. Benchmark reports should identify their corpus and conditions. A changed fixture is not accepted merely because output differs; the behavioral contract and expected bytes must be reviewed.

## Related docs

- [Development](INDEX.md)
- [Repository Layout](repository-layout.md)
- [Documentation Coverage Map](documentation-coverage.md)
- [Behavioral Contract Matrix](behavioral-contract-matrix.md)
- [Safe Extension Procedures](safe-extension-procedures.md)
- [Documentation Procedure](../documentation-procedure.md)
- [Document Health Checks](../guides/document-health-checks.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Review Ledger](../architecture/review-ledger.md)
- [Link Performance](../research/link-performance.md)
- [Codemap Evidence](../research/codemap-evidence.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Context-Injection Benchmarking](../research/context-injection-benchmarking.md)

## Notes

The complete release gate is the preferred pre-merge verification because it combines tests, regression fixtures, vetting, builds, and CLI smoke checks.
