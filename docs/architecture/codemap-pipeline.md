# Codemap Pipeline

Parent index: [Architecture](./README.md)

## Purpose

This document describes the currently implemented codemap pipeline: how authored Markdown code maps are extracted, how repository facts are assembled, how deterministic evidence becomes ranked missing-link suggestions, and how the benchmark and precision commands evaluate that behavior.

## Overview

The current system is a one-directional analysis pipeline:

```text
Markdown code-map sections
  -> internal/codemap extraction and target resolution
  -> versioned codemap dataset
  -> internal/codemapcorpus repository facts
  -> internal/evidence candidate evidence and fingerprints
  -> internal/codemapbench ranking and suggestion tiers
  -> current suggestions, controlled holdout reports, or precision evaluation
  -> internal/app command output and review integration
```

The system only suggests potentially missing document-to-code links. It does not suggest removal, irrelevance, or cleanup of an existing authored link. A ranked candidate is review data, not an authored relationship and not an automatic write decision.

## Code root

```text
internal/codemap/
internal/codemapcorpus/
internal/evidence/
internal/codemapbench/
internal/codemapprecision/
internal/app/app.go
internal/app/codemap_benchmark.go
internal/app/codemap_benchmark_engine.go
internal/app/codemap_precision.go
internal/app/codemap_precision_source.go
internal/app/review_codemap.go
```

## Responsibilities

The codemap pipeline owns:

- parsing configured Markdown code-map sections without normalizing the authored source first;
- normalizing authored target paths and classifying target kinds;
- resolving targets and exporting deterministic dataset records;
- collecting repository files, document text, resolved authored targets, dependency edges, symbol declarations, related documents, and bounded Git history;
- converting those facts into candidate evidence records and stable fingerprints;
- scoring, ordering, and tiering potentially missing-link suggestions;
- hiding controlled holdout links from benchmark inputs;
- sampling and evaluating manually labeled precision data; and
- exposing export, benchmark, precision, and review-time current-suggestion flows through `internal/app`.

## Does not own

The codemap pipeline does not own:

- general semantic code-graph truth;
- the meaning, completeness, or necessity of authored documentation;
- persisted review decisions, declines, blocks, or undo policy;
- automatic authorship of ranked relationships;
- removal or irrelevance judgments for existing codemap links; or
- arbitrary prose interpretation beyond the implemented deterministic extractors.

## Flow and lifecycle

### Extraction and dataset

`internal/codemap/extractor.go` parses a document using the configured section-heading aliases. The built-in aliases are `code map`, `codemap`, `code or source map`, and `code and test map`; repository configuration or the export command can replace them. Extraction stops at the next heading at the same or higher level.

The extractor accepts the current authored forms covered by its tests:

- Markdown list entries whose first code span is a target;
- fenced map content with target lines and arrow, equals, leading-path, or indented descriptions;
- indented legacy targets with inline or following descriptions; and
- nested headings and colon-terminated group labels as context.

It preserves the document path, map heading, target, target kind, syntax kind, context, description, source span, and raw line in each `codemap.Entry`. Targets are normalized to slash-separated, cleaned paths while preserving a trailing directory slash. Target kinds distinguish files, directories, globs, symbols, and unknown targets. Text outside a configured map section, prose-only bullets, and TODO-only map content are not treated as authored links.

`internal/codemap/dataset.go` builds a schema-1 `codemap.Dataset` by walking Markdown files below the selected docs root under the repository ignore policy. It skips directories, symlinks, and non-Markdown files. Each document records its path, byte size, SHA-256, section count, entry count, and diagnostic count. Each entry carries a `TargetRecord` with resolution status, resolved path or pattern matches, existence, size, and file hash where applicable.

Target resolution uses the configured repository-relative or document-relative base plus optional repository-relative target roots. It keeps outside-repository, missing, kind-mismatch, symbol-unverified, pattern-resolved, pattern-missing, ambiguous, and unsupported states explicit. It does not guess an ambiguous target. `MarshalDataset` and `ExportDataset` emit indented JSON with stable ordering. The dataset schema includes diagnostics, although the current extractor tests exercise the normal path with no emitted diagnostics; unrecognized authored lines are skipped rather than invented as entries.

`internal/codemap/strip.go` removes configured authored map sections while retaining non-map text and line structure. The benchmark engine uses this before evidence collection so a hidden authored link cannot be read back directly from the document.

`internal/codemap/insert.go` is the narrow write primitive for a selected candidate. It appends a bullet to the first configured map section, rejects a target that is already authored, and returns byte offsets and inserted text for the generated repair. It does not select candidates or decide whether a link is semantically required.

### Corpus facts

`internal/codemapcorpus/build.go` constructs a `Corpus` from a repository and dataset. The corpus contains:

- repository files and derived repository file and directory paths;
- the source text of dataset documents;
- resolved authored targets grouped by document;
- deterministic dependency edges;
- bounded commit path sets;
- related documents with their visible authored targets; and
- declared symbols.

Repository file discovery tries `git ls-files` first, then the go-git index, then an ignored-aware filesystem walk. The walk avoids nested Git repositories. `.docignore` and the shared ignore policy can exclude files from the corpus. All paths are normalized and sorted.

`internal/codemapcorpus/dependencies.go`, `dependency_go.go`, and `dependency_scripts.go` collect only local dependency facts that the current adapters recognize:

- Go imports resolved through repository `go.mod` module paths to local non-test Go files;
- Godot `preload`, `load`, and quoted `extends` references, including `res://` resources;
- relative JavaScript, JSX, TypeScript, and TSX imports, side-effect imports, and `require` references;
- Ruby `require_relative`; and
- Python relative `from` imports.

`internal/codemapcorpus/symbols.go` extracts Go named types, exported functions, exported methods on exported receivers, and GDScript `class_name` and qualified functions. It rejects generic or ambiguous symbol matches rather than assigning a symbol to multiple paths.

`internal/codemapcorpus/related.go` resolves local Markdown links among documents in the dataset. A related-document fact carries only the related document's resolved authored targets. `internal/codemapcorpus/history.go` and `gitcli.go` collect non-merge commits, using Git CLI when available and go-git otherwise. The defaults examine at most 1,000 commits and ignore commits changing more than 200 paths; only commits with at least two repository files contribute.

The corpus is a provider of normalized facts. It does not infer a missing relationship itself.

### Evidence

`internal/evidence/collect.go` receives one document's text, repository paths, existing visible targets, dependency edges, commit facts, related-document targets, and symbol declarations. It excludes the document itself and every supplied existing target before adding a candidate. Each candidate contains a target path, one or more evidence records, and a SHA-256 fingerprint derived from the target and canonical evidence fields.

The current evidence kinds are:

- exact repository-relative path mention;
- unique basename mention;
- declared-symbol mention when the symbol resolves to one path;
- sibling of an existing target;
- source/test counterpart;
- dependency neighbor in either direction;
- Git co-change with the document;
- Git co-change with an existing target; and
- a target already authored by a related document.

Mentions are token-boundary checked. Basename evidence is used only when the basename is unique in the repository path set. Structural evidence considers same-directory siblings and test/spec naming or directory counterparts. Dependency, related-document, and history evidence are taken from the normalized corpus facts rather than from arbitrary prose. Evidence counts are retained and evidence items and candidates are emitted in deterministic order.

Evidence is explanatory input to ranking. It is not proof that a target is required, and it does not create a graph edge or mutate a document.

### Current suggestions, ranking, and tiers

`internal/codemapbench/suggestions.go` converts evidence candidates into `Suggestion` values. The current scorer assigns these base weights:

```text
exact path mention             6
unique basename mention       4
declared symbol mention       7
test counterpart              6
dependency neighbor            4
related-document target        4
sibling target                 2
Git target co-change          1.5
Git document co-change         1
```

Repeated evidence uses a logarithmic occurrence factor, except exact-path and unique-basename mentions, which use a fixed occurrence factor. A repeated evidence atom is discounted by the logarithm of its fan-out across candidates so a broad directory or commit does not dominate the ranking. A candidate normally needs two evidence kinds, or one admitted current evidence kind, to enter ranking. The selected list is capped at 30 suggestions per document, with up to two additional repeated exact-path mentions reserved when they occur at least twice.

Selected suggestions are ordered by descending score with target-path tie-breaking. The first five positions are eligible for `hard_link` when they have a declared-symbol mention, a test counterpart, or a dependency-neighbor score of at least 16. All other selected suggestions are `context`. `hard_link` is the current link-review surface; `context` preserves weaker or indirect relationships that may be useful for bounded context but are not strong enough for the current permanent-link tier.

The tier is descriptive ranking metadata. It does not automatically write a link, establish semantic coverage, or authorize removal of an existing link.

### Controlled holdouts

`internal/codemapbench/holdout.go` and `run.go` implement the controlled recovery benchmark. `ResolvedLinksFromDataset` uses only exact entries with `ResolutionResolved`; pattern families, unresolved symbols, stale targets, and other non-exact resolution states are not exact-link holdout answers. A reviewed trusted-link JSON file can be used instead of the dataset's resolved links.

Known links are normalized, deduplicated, and sorted. The default seed is `demon-docs-codemap-benchmark-v1`. The default holdout is 20 percent of known links, rounded up, unless an exact count or fraction is supplied. A SHA-256 hash of the seed and document-target key chooses the hidden subset independently of input order.

The runner sends the generator only the document list and visible links. For each document, `internal/app/codemap_benchmark_engine.go` obtains corpus input, replaces authored document text with `StripAuthoredSections`, and sanitizes related-document targets to visible links. Hidden links therefore do not appear in the document text, existing-target seeds, or related-document evidence provided to the generator.

The report classifies normalized generator output as recovered hidden links, missed hidden links, unmatched suggestions, already-linked suggestions, duplicates, or invalid suggestions. It calculates benchmark precision as recovered divided by recovered plus unmatched plus already-linked suggestions, and recall as recovered divided by hidden links. Reports are canonical, schema-versioned, and available as text or JSON. These are measurements of the selected corpus, holdout, and generator run; they are not universal product guarantees.

`SuggestCurrent` is the non-holdout path used for current suggestions and precision sampling. It treats every exact authored link supplied by the corpus as visible and asks only for additional candidates. It does not convert current suggestions into holdout truth.

### Precision evaluation

`internal/codemapprecision/precision.go` evaluates manually labeled candidates from a source suggestion report. The labels are:

- `valid_missing_link`;
- `plausible_but_unnecessary`; and
- `incorrect`.

`CandidatesFromReport` uses unmatched suggestions only, ranks them within each document by score and target, and excludes recovered trusted links from precision samples. `BuildBenchmark` creates an unlabeled schema-1 template with repository and revision metadata. `Sample` uses a required seed and deterministic balancing: it selects representative documents across available areas, subsystems, score buckets, and primary evidence kinds, keeps top-five candidates where available, and fills from lower ranks and the remaining candidate pool until the requested count is reached.

`Evaluate` requires complete labels, rationales, and document/target audit references. It verifies that every sampled candidate still exists in the source report with the same score, evidence, and rank before computing results. The evaluation includes overall precision and non-junk acceptance, precision at ranks 1, 3, and 5, per-document results, evidence-kind, score-bucket, rank-bucket, tier, and sampling-coverage breakdowns, plus hard-link sample valid recall and suggestions per document.

Precision results describe the pinned labeled benchmark and its curation rules. They are regression and tuning evidence, not a guarantee for unrelated repositories or future corpora.

## Command orchestration

`internal/app/app.go` dispatches `ddocs codemap` to the three current command families:

- `ddocs codemap export` resolves repository/configuration scope, headings, target base, and target roots, calls `codemap.BuildDataset`, and writes deterministic JSON to stdout or a requested output file;
- `ddocs codemap benchmark` is parsed by `internal/app/codemap_benchmark.go`; `internal/app/codemap_benchmark_engine.go` loads or builds the dataset, builds the corpus, selects dataset or trusted-review links, runs `codemapbench.Run`, and encodes the report. It supports seed, count/fraction, text/JSON, output, and minimum precision/recall thresholds. Exit code 0 means completion with thresholds passing, 1 means a completed run failed a requested threshold, and 2 means invalid arguments or execution failure; and
- `ddocs codemap precision` is parsed by `internal/app/codemap_precision.go`. `source` is implemented in `internal/app/codemap_precision_source.go` and builds a corpus plus `SuggestCurrent` report, with optional document-prefix exclusions. `sample` loads a source report and writes an unlabeled precision template. `evaluate` loads a labeled benchmark and source report, evaluates them, and writes text or JSON results.

At review time, `internal/app/review_common.go` runs the same current dataset/corpus/`SuggestCurrent` path and converts unmatched codemap suggestions into review candidates. `internal/app/review_codemap.go` applies a user-selected candidate through `codemap.InsertTarget`, then records the generated hash-guarded rewrite through the review/link machinery. The analysis pipeline itself never applies a suggestion merely because it was ranked.

## State and data ownership

- `internal/codemap` owns authored-entry extraction, target normalization, resolution records, dataset serialization, map stripping, and the narrow selected-target insertion primitive.
- `internal/codemapcorpus` owns repository fact collection and the corpus passed to evidence consumers.
- `internal/evidence` owns candidate construction, evidence aggregation, deterministic ordering, and fingerprints.
- `internal/codemapbench` owns suggestion scoring, tiers, current-vs-holdout orchestration, classification, and benchmark reports.
- `internal/codemapprecision` owns precision sample schemas, deterministic sampling, label validation, and metric aggregation.
- `internal/app` owns command parsing, configuration and repository scope resolution, pipeline assembly, output, and process exit behavior.
- `internal/review` and the review-facing app files own persisted decisions, declines, applied changes, and undo behavior.

No package treats an evidence candidate as authored documentation truth.

## Invariants and safety boundaries

- Only potentially missing links are suggested.
- Existing authored links are never described as irrelevant and are never returned as removal candidates.
- The candidate collector excludes the document and the visible existing-target set supplied by the orchestrator.
- Benchmark holdouts are absent from document map text, visible-target seeds, and sanitized related-document targets.
- Exact-link recovery benchmarks do not use unresolved, ambiguous, pattern, or stale dataset records as exact answers.
- Evidence, score, and tier are review metadata; none is an automatic write instruction.
- A selected write is rejected when the target is already authored and is recorded through the normal bounded review rewrite path.
- Normalized paths, evidence, candidates, holdout selection, samples, and reports have deterministic ordering for identical inputs.
- Ambiguous or unsupported extraction and target-resolution cases remain non-authoritative rather than being guessed into a link.
- Sample or benchmark metrics must be interpreted only with their corpus, revision, labels, seed, and sampling/holdout method.

## Code map

Primary implementation:

- `internal/codemap/model.go` - authored entry, target-kind, syntax, format, and resolution model types.
- `internal/codemap/extractor.go` - configured-section Markdown extraction and target normalization.
- `internal/codemap/dataset.go` - ignored-aware Markdown scan, target resolution, hashes, and dataset export.
- `internal/codemap/strip.go` - removal of authored map content from benchmark evidence input.
- `internal/codemap/insert.go` - selected candidate insertion primitive.
- `internal/codemapcorpus/build.go` and `model.go` - corpus assembly and input projection.
- `internal/codemapcorpus/files.go`, `paths.go`, and `gitcli.go` - repository paths and bounded Git facts.
- `internal/codemapcorpus/dependencies.go`, `dependency_go.go`, and `dependency_scripts.go` - local dependency adapters.
- `internal/codemapcorpus/symbols.go` - Go and GDScript declaration extraction.
- `internal/codemapcorpus/related.go` - related-document target facts.
- `internal/evidence/model.go` and `collect.go` - evidence model, candidate collection, and fingerprints.
- `internal/evidence/mentions.go`, `structure.go`, `symbols.go`, and `history.go` - current evidence signal collectors.
- `internal/codemapbench/orchestrator.go` and `run.go` - corpus/generator seams and holdout execution.
- `internal/codemapbench/current.go` - current non-holdout suggestion execution.
- `internal/codemapbench/suggestions.go` - scoring, admission, ranking, and tier assignment.
- `internal/codemapbench/holdout.go`, `adapters.go`, and report files - holdout selection, trusted links, and report serialization.
- `internal/codemapprecision/model.go` and `precision.go` - labeled benchmark schema, sampling, validation, and evaluation.
- `internal/app/app.go` - codemap dispatch and export orchestration.
- `internal/app/codemap_benchmark.go` and `codemap_benchmark_engine.go` - benchmark command and engine integration.
- `internal/app/codemap_precision.go` and `codemap_precision_source.go` - precision command flows.
- `internal/app/review_common.go` and `review_codemap.go` - current review suggestions and selected codemap repair integration.

## Tests

The implementation is covered by focused tests for:

- extraction forms, configured headings, prose boundaries, TODO-only maps, fixture inventories, dataset resolution, target roots, stable JSON, stripping, duplicate insertion rejection, and selected insertion (`internal/codemap/*_test.go`);
- ignored-aware corpus files, resolved target projection, related documents, supported dependency adapters, repository paths, symbol extraction/filtering, bounded history, and missing dataset documents (`internal/codemapcorpus/*_test.go`);
- deterministic evidence, repeated mentions, unique basenames, fingerprints, symbol ambiguity, every current signal in the pinned validation cases, and exclusion of existing targets (`internal/evidence/*_test.go`);
- weighted scoring, repeated-mention reservation, fan-out discounting, weak-signal admission, per-document bounds, hard-link caps, current visible-link behavior, deterministic holdouts, hidden-link isolation, orchestration errors, and canonical text/JSON reports (`internal/codemapbench/*_test.go`);
- deterministic ranking and stratified sampling, complete audit validation, schema/trailing-data checks, tier validation, overall/top-k/per-document metrics, and breakdowns (`internal/codemapprecision/precision_test.go`); and
- export output and configuration, benchmark flags/thresholds, benchmark answer isolation, precision source filtering, precision sampling/evaluation, and command help (`internal/app/codemap_*_test.go`).

## Related docs

- [Architecture](README.md)
- [Application Orchestration](application-orchestration.md)
- [Codemap Missing-Link Evidence](../research/codemap-evidence.md)
- [Evaluating Codemap Suggestions](../guides/evaluating-codemap-suggestions.md)
- [Current Product Limitations](../limits/current-limitations.md)
- [Review Ledger](review-ledger.md)
- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)
- [Repository Layout](../development/repository-layout.md)

## Notes

The current implementation is deterministic and review-oriented, but evidence strength and evaluation metrics remain bounded by the repository facts, selected revision, holdout or sample construction, and human labels supplied to each run. The permanent safety rule is unchanged: suggest potentially missing links only; never suggest removal or irrelevance of existing links.
