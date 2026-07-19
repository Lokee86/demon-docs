# Codemap Corpus and Adapters

Parent index: [Architecture](./README.md)

## Purpose

This document describes how Demon Docs builds normalized repository facts for codemap evidence, including repository files, dependency edges, symbols, related documents, and bounded Git history.

## Overview

The corpus is a deterministic fact boundary between the authored codemap dataset and evidence collection.

```text
repository + codemap dataset
-> tracked/visible repository paths
-> authored document text and resolved targets
-> language-specific local dependencies
-> declared symbols
-> related-document targets
-> bounded co-change history
-> normalized Corpus
```

Adapters report facts they can establish from supported syntax. They do not rank candidates or infer that a document should link to a target.

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
- resolved authored targets grouped by document;
- local dependency extraction for supported languages;
- declared symbol extraction for supported languages;
- related-document relationships derived from local Markdown links;
- bounded non-merge Git commit path sets; and
- normalized, deduplicated, sorted corpus output.

## Does not own

It does not own:

- code-map Markdown extraction;
- suggestion evidence weights or ranking;
- a complete language semantic graph;
- external package resolution;
- runtime call graphs;
- symbol-reference resolution; or
- persistent review decisions.

## Corpus construction flow

`Build` receives repository scope and a codemap dataset.

It performs:

```text
validate every dataset document exists
-> discover repository files
-> derive repository directories
-> read dataset document text
-> project resolved authored targets
-> collect dependencies
-> extract symbols
-> resolve related documents
-> collect bounded history
-> sort and return normalized facts
```

A dataset document missing from the repository is an error because later evidence would otherwise be built from incomplete or mismatched inputs.

## Repository file discovery

File discovery tries these sources in order:

1. Git CLI tracked files;
2. the go-git index; and
3. an ignore-aware filesystem walk.

The fallback walk avoids nested Git repositories. Shared permanent exclusions and `.docignore` apply. Paths are normalized to repository-relative slash form and sorted.

Parent directory paths are derived from files so evidence can reason about repository directories even when Git does not track empty directory objects.

## Authored target projection

Only resolved authored targets from the dataset become existing-target facts. Their normalization is stable and document-scoped.

Unresolved, ambiguous, unsupported, or pattern-only records do not become an exact existing target merely because they appeared in authored Markdown.

This projection is used to exclude already-authored relationships from missing-link candidates.

## Dependency adapter contract

Dependency adapters emit `evidence.DependencyEdge` values with:

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

## Symbol extraction

Current symbol facts include:

- Go named types;
- exported Go functions;
- exported Go methods on exported receivers;
- GDScript `class_name`; and
- qualified GDScript functions.

Generic, unexported, common, or ambiguous declarations are filtered according to the implemented extractor rules. A symbol that maps to multiple paths cannot become unique symbol evidence.

Symbol declarations are facts about definitions, not references from arbitrary code.

## Related-document facts

Local Markdown links among dataset documents establish related-document facts. Each related record exposes the related document's resolved authored code targets.

This lets one document's explicit map inform another document's evidence without treating arbitrary proximity as a relationship.

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

## State and data ownership

The corpus is an in-memory, rebuildable projection. It does not persist a second repository graph.

- repository files and directories come from current scope;
- document text and authored targets come from the selected dataset;
- dependency and symbol facts come from supported source adapters;
- related facts come from current local Markdown links; and
- history facts come from bounded current Git history.

## Invariants and safety boundaries

- Every emitted path is normalized and repository-relative.
- Current ignore policy applies to corpus file discovery.
- Nested repositories are not traversed by fallback walking.
- Missing dataset documents fail construction.
- Dependency adapters emit local facts only.
- Ambiguous resolutions are omitted rather than guessed.
- Symbol ambiguity prevents unique symbol evidence.
- Collections are deduplicated and sorted.
- Hidden holdout targets must be removed from related-document inputs before generation.

## Failure behavior

Corpus construction fails when required documents or repository files cannot be read, Git/index discovery fails without a usable fallback, a supported adapter encounters an I/O error, or history collection cannot produce the required bounded facts.

Unsupported language syntax normally results in no fact rather than a fatal error. Consumers must not interpret missing adapter evidence as proof of no relationship.

## Code map

- `internal/codemapcorpus/build.go` and `model.go` — corpus assembly and model.
- `files.go`, `paths.go`, and `gitcli.go` — repository discovery and normalized paths.
- `dependencies.go` — adapter dispatch and local target index.
- `dependency_go.go` — Go modules and imports.
- `dependency_scripts.go` — GDScript, JavaScript/TypeScript, Ruby, and Python adapters.
- `symbols.go` — current declaration extractors.
- `related.go` — local document relationships.
- `history.go` — bounded commit facts.

## Tests

Focused tests cover:

- complete corpus input assembly and missing-document refusal;
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
