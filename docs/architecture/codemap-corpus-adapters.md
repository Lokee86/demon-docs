---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-7a47-b5c2-b8cd161a8a8b
document_type: general
policy_exempt: false
summary: This document describes how Archivist builds normalized repository facts for codemap evidence, including repository files, dependency edges, symbols, related documents, and bounded Git history.
---
# Codemap Corpus and Adapters

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes how Archivist builds normalized repository facts for codemap evidence, including repository files, dependency edges, symbols, related documents, and bounded Git history.

## Overview

The corpus is a deterministic fact boundary between the authored codemap dataset, code-intelligence providers, and evidence collection.

```text
repository + codemap dataset
-> tracked/visible repository paths
-> authored document text, target provenance, and concrete coverage
-> CodeIntelligenceProvider
   -> dependency facts
   -> declared-symbol facts
-> related-document targets
-> bounded co-change history
-> corpus-owned normalization and validation
-> normalized Corpus
-> per-document InputContext
   -> currently visible exact file/symbol seeds
   -> optional RelationshipProvider
   -> bounded semantic relationship facts
```

The providers report facts only. They do not rank candidates, decide documentation coverage, or mutate authored maps. The repository-wide `CodeIntelligenceProvider` defaults to Archivist's shallow language adapters. The per-document `RelationshipProvider` is optional and is currently backed by Arcana when a current verified snapshot is available.

## Code root

```text
internal/codemapcorpus/
internal/evidence/model.go
```

## Responsibilities

This boundary owns:

- repository file discovery under the ignore policy;
- derived repository directory paths;
- document source text for dataset documents;
- authored target provenance grouped by document, preserving file, directory, pattern, symbol, and unknown kinds;
- concrete authored coverage used to suppress already-covered missing-link candidates;
- the code-intelligence provider contract for repository-local dependency and declared-symbol facts;
- the per-document relationship-provider contract for bounded graph facts from currently visible exact targets;
- deterministic validation, normalization, deduplication, and ordering of provider facts;
- shallow local dependency and symbol extraction as the default fallback provider;
- related-document relationships derived from local Markdown links;
- bounded non-merge Git commit path sets; and
- normalized, deduplicated, sorted corpus output.

## Does not own

It does not own:

- code-map Markdown extraction;
- suggestion evidence weights or ranking;
- provider-specific graph storage, snapshot lifecycle, or transport;
- a complete language semantic graph;
- external package resolution in the local fallback;
- runtime call graphs in the local fallback;
- symbol-reference resolution; or
- persistent review decisions.

## Corpus construction flow

`Build` receives repository scope and a codemap dataset.

It performs:

```text
discover repository files
-> derive repository directories, authored target provenance, and concrete coverage
-> concurrently load dataset documents, request code-intelligence facts, and collect bounded history
-> validate/normalize provider facts against current repository files
-> resolve related documents from the completed document set
-> merge, deduplicate, sort, and return normalized facts
```

Document loading, code-intelligence collection, and history collection each have an independent bounded task. The caller context is propagated through `BuildContext` to the provider so an external process boundary can be cancelled cleanly. The compatibility `Build` entry point uses a background context.

The default local provider retains the existing bounded 16-worker source pass: each supported source is read once, and its immutable bytes feed every applicable fallback adapter before results merge serially in source-path order. Repository discovery, provider-output validation, final deduplication, related-document projection, and corpus publication remain deterministic.

Subsystem errors are checked after collection in stable document, source-fact, then history order. A dataset document missing from the repository remains an error because later evidence would otherwise be built from incomplete or mismatched inputs.

## Repository file discovery

File discovery tries these sources in order:

1. Git CLI tracked files;
2. the go-git index; and
3. an ignore-aware filesystem walk.

The fallback walk avoids nested Git repositories. Shared permanent exclusions and `.docignore` apply. Paths are normalized to repository-relative slash form and sorted.

Parent directory paths are derived from files so evidence can reason about repository directories even when Git does not track empty directory objects.

## Authored target projection

The corpus preserves the authored abstraction level instead of flattening every resolution into an equivalent file seed. Each document retains target provenance for explicit files, directories, glob/pattern families, symbols, and unknown targets together with any concrete repository paths established by resolution.

Concrete coverage and outward evidence expansion are separate concepts:

- exact authored files cover their resolved file and may seed bounded structural, dependency, and target-history evidence;
- authored directories cover the directory and its descendants but do not turn every descendant into an outward expansion seed;
- resolved patterns cover their matched paths but remain one authored pattern boundary rather than a set of independently authored files; and
- symbol targets preserve their backing-path resolution without becoming generic file-neighborhood seeds.

For basename-only patterns such as `internal/app/codemap_*.go`, the literal parent directory is also an intentional scope boundary. Inferred evidence does not promote non-matching siblings from that directory merely because they are related through history, dependencies, tests, or another document. A direct path, basename, or declared-symbol mention in the current document may still surface such a target.

Unresolved, ambiguous, unsupported, or kind-mismatched records do not become safe outward expansion seeds merely because they appeared in authored Markdown.

## Code-intelligence provider seam

`CodeIntelligenceProvider` is the narrow replaceable boundary for semantic repository facts. A provider receives only the absolute repository root, the current normalized repository-file inventory, and the caller context. It returns:

```text
DependencyEdge[]
SymbolDeclaration[]
```

Archivist validates every returned path against the current repository-file inventory, rejects provider facts that point outside that inventory, trims relation/symbol identifiers, removes self-edges and empty facts, deduplicates equivalent facts, and publishes deterministic ordering. Provider order is therefore not observable by evidence or ranking.

The provider does **not** receive document text, existing codemap targets, evidence weights, review state, or mutation authority. This keeps future Lexicon/Arcana integration on the fact side of the boundary rather than allowing an external graph to decide what belongs in documentation.

`codemapcorpus.Options.CodeIntelligence` selects the provider. A nil provider uses the built-in local fallback. `codemaprun.Options.CodeIntelligence` forwards the same seam into production planning. Production, benchmark, and precision entry points use `BuildContext`; current CLI paths still leave the provider nil, so external provider selection is not wired until the concrete integration step.

Provider failures fail corpus construction rather than silently mixing incomplete semantic facts with fallback facts. Explicit stale/unavailable degradation policy belongs to the concrete external provider integration and is not implemented by this seam alone.

## Relationship-provider seam

`RelationshipProvider` is deliberately separate from repository-wide code-intelligence collection. It is invoked only by `Corpus.InputContext` after the caller has supplied the exact authored targets that remain visible for one document. The request contains the repository root, current repository-file inventory, and normalized exact file/symbol seeds for that document.

This placement is a benchmark-safety boundary. A held-out codemap target is removed from the visible target set before relationship collection, so that hidden answer cannot influence which Arcana neighborhood is queried. Directory and pattern targets never become relationship seeds. A symbol target becomes a seed only when Step 4 resolved it to an exact semantic node.

The current Arcana provider queries only a one-hop allowlist:

```text
calls
imports
depends-on
implements
extends
overrides
uses-trait
includes
tests
```

File seeds expand only relation-capable nodes whose source path is exactly the authored file. Verified symbol seeds resolve back to their exact Arcana identity and expand only that symbol. Each seed is bounded to at most 128 relation-capable nodes, and each node/direction is bounded to 128 neighbors. If Arcana reports truncation at either boundary, that seed contributes no semantic relationship evidence rather than publishing an arbitrary partial neighborhood.

Both the seed path and returned neighbor path must still match the Lexicon content identity verified by the Step 4 Arcana resolver. Relationships are projected to repository file pairs, normalized against the current repository-file inventory, deduplicated, and sorted before evidence collection.

Arcana relationship facts are a distinct `semantic_relationship` evidence kind. They may produce an inspectable `context` recommendation but do not satisfy any `hard_link` promotion rule by themselves. Step 6 now uses their direction and relation type as deterministic candidate-role input; stronger role-aware ranking or promotion remains deferred.

## Local fallback dependency adapters

The default local provider uses the existing dependency adapters to emit `evidence.DependencyEdge` values with:

```text
repository-relative source
repository-relative target
relation identifier
```

Edges are local, deduplicated, self-edge-free, and sorted.

Current supported source extensions are:

```text
Go:          .go
GDScript:    .gd
JavaScript:  .js .jsx .mjs .cjs
TypeScript:  .ts .tsx
Ruby:        .rb
Python:      .py
```

### Go

Go imports are resolved through repository `go.mod` module paths. Local package imports project to local non-test Go files. External modules are not converted into repository targets.

The adapter does not construct a full Go package or symbol graph.

### GDScript

The adapter recognizes local `preload`, `load`, and quoted `extends` references, including `res://` resources. Godot roots are discovered from `project.godot` files, with more specific roots considered before broader roots.

Dynamic expressions and runtime resource construction are unsupported.

### JavaScript and TypeScript

The adapter recognizes relative imports, side-effect imports, and `require` forms implemented by the parser. It tries supported extensions and index-file forms for relative paths.

Package names, configured path aliases, bundler aliases, and arbitrary resolver plugins are not treated as local without an implemented adapter seam.

### Ruby

The adapter recognizes `require_relative`. General `require`, load-path manipulation, autoloading, and framework conventions are not resolved.

### Python

The adapter recognizes relative `from` imports. Absolute module imports, namespace packages, dynamic imports, and environment-dependent import paths are not resolved.

## Adapter fallback and ambiguity

Adapters only emit edges to paths present in the repository file index. Candidate extension and index-file fallbacks are tested in deterministic order and deduplicated.

When supported syntax cannot identify one current repository target, the adapter omits the edge rather than selecting an arbitrary file.

Unsupported syntax is a documented absence of evidence, not evidence that no dependency exists.

## Local fallback symbol extraction

The default local provider's current symbol facts include:

- Go named types;
- exported Go functions;
- exported Go methods on exported receivers;
- GDScript `class_name`; and
- qualified GDScript functions.

Generic, unexported, common, or ambiguous declarations are filtered according to the implemented extractor rules. A symbol that maps to multiple paths cannot become unique symbol evidence.

Symbol declarations are facts about definitions, not references from arbitrary code.

## Related-document facts

Local Markdown links among dataset documents establish related-document facts. Related-document projection exposes direct resolved file targets; directory, pattern, and symbol abstractions are not flattened into inherited per-file targets.

This lets one document's explicit file map inform another document's evidence without turning broader authored abstractions into transitive file-level authority.

During controlled holdouts, the benchmark orchestrator sanitizes these target lists so hidden answers are not exposed indirectly.

## Git history facts

History collection uses Git CLI when available and go-git otherwise.

Defaults:

```text
maximum commits examined: 1000
maximum changed paths admitted per commit: 200
merge commits: excluded
minimum repository files in a contributing commit: 2
```

The result is a set of normalized commit IDs and repository paths. Large bulk commits are excluded to reduce broad, low-specificity co-change evidence.

History is bounded evidence, not ownership truth. Squashes, rebases, generated commits, and repository age affect what can be observed.

## Corpus construction performance

A retained Windows benchmark constructs a corpus from 384 Go files, 128 GDScript files, and 96 dataset documents. Every source contributes dependency facts, Go and GDScript sources also contribute symbol facts, and the repository intentionally has no Git history so the measurement includes the normal empty-history fallback. The comparison used `GOMAXPROCS=16`, five one-iteration runs, and commit `41b6f3e` as the serial baseline.

Excluding the first sample to reduce host filesystem and antivirus warm-up noise, mean full-corpus construction improved from 1.546 seconds to 523.3 milliseconds, a 2.95x speedup and 66.2% latency reduction. The retained benchmark is `BenchmarkBuildCorpus` in `internal/codemapcorpus/build_benchmark_test.go`.

```bash
go test ./internal/codemapcorpus -run '^$' -bench '^BenchmarkBuildCorpus$' -benchmem -count=5 -benchtime=1x
```

## State and data ownership

The corpus is an in-memory, rebuildable projection. It does not persist a second repository graph.

- repository files and directories come from current scope;
- document text and authored targets come from the selected dataset;
- dependency and symbol facts come through the selected `CodeIntelligenceProvider`;
- the default provider derives those facts from the built-in shallow source adapters;
- bounded per-document semantic relationship facts come through the optional `RelationshipProvider` and are not persisted in the corpus;
- related facts come from current local Markdown links; and
- history facts come from bounded current Git history.

## Invariants and safety boundaries

- Every emitted path is normalized and repository-relative.
- Current ignore policy applies to corpus file discovery.
- Nested repositories are not traversed by fallback walking.
- Missing dataset documents fail construction.
- Provider dependency and symbol paths must resolve to current repository files or corpus construction fails.
- The default local dependency adapters emit local facts only.
- Ambiguous local-fallback resolutions are omitted rather than guessed.
- Symbol ambiguity prevents unique symbol evidence.
- Source workers read immutable repository bytes and publish only indexed detached results.
- Worker completion order cannot affect returned ordering or first-error selection.
- Collections are deduplicated and sorted serially after collection.
- Hidden holdout targets must be removed from related-document inputs and relationship-provider seeds before generation.
- Arcana relationship neighborhoods are one-hop, allowlisted, source-current, and bounded; truncated seed neighborhoods are discarded rather than partially trusted.
- Directory and pattern coverage suppresses already-covered descendants or matches without making them independent expansion seeds.
- Basename-only authored patterns constrain inferred sibling evidence outside the pattern unless the current document supplies direct mention or symbol evidence.

## Failure behavior

Corpus construction fails when required documents or repository files cannot be read, Git/index discovery fails without a usable fallback, a supported adapter encounters an I/O error, or history collection cannot produce the required bounded facts.

Unsupported language syntax normally results in no fact rather than a fatal error. Consumers must not interpret missing adapter evidence as proof of no relationship.

## Code map

- `internal/codemapcorpus/build.go` and `model.go` — corpus assembly, provider selection, and model.
- `code_intelligence.go` — repository-wide provider contract, default local provider, and corpus-owned provider-output validation.
- `relationship_intelligence.go` and `input.go` — visible-seed relationship contract, benchmark-safe per-document collection, and relationship normalization.
- `build_collections.go` — concurrent document, code-intelligence, and history collection with deterministic error priority.
- `source_facts.go` and `workers.go` — default-provider bounded source reads, fallback adapter dispatch, indexed results, and serial normalized merge.
- `files.go`, `paths.go`, and `gitcli.go` — repository discovery and normalized paths.
- `dependencies.go` — dependency adapter dispatch and local target index.
- `dependency_go.go` — Go modules and imports.
- `dependency_scripts.go` — GDScript, JavaScript/TypeScript, Ruby, and Python adapters.
- `symbols.go` — current declaration extractors.
- `related.go` — local document relationships.
- `history.go` — bounded commit facts.

## Tests

Focused tests cover:

- complete corpus input assembly and missing-document refusal;
- injected provider usage, context propagation, deterministic provider normalization, invalid-path rejection, and provider error propagation;
- visible relationship-seed projection, hidden-holdout isolation, verified-symbol seed identity, and normalized relationship output;
- bounded local-fallback source concurrency, deterministic indexed errors, and one read per supported source;
- retained large-corpus construction performance;
- repository paths and tracked parent directories;
- every supported dependency adapter and local-only resolution;
- Go and GDScript declarations plus ambiguity/filtering;
- related-document target projection;
- bounded history behavior; and
- ignore-aware deterministic file discovery.

```bash
go test ./internal/codemapcorpus -count=1
```

## Related docs

- [Codemap Pipeline](codemap-pipeline.md)
- [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md)
- [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md)
- [Ignore and Traversal](ignore-and-traversal.md)
- [Extending Codemap Analysis](../development/extending-codemap-analysis.md)

## Notes

The corpus is intentionally polyglot but not language-complete. A small explicit adapter with tested refusal behavior is preferable to broad heuristic dependency guessing.
