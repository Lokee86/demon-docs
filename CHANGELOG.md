# Changelog

This file records notable user-facing changes to Demon Docs. Exact command contracts and current behavior remain documented in the README and the canonical documentation tree.

## [0.3.5] - 2026-07-21

### Initial public release

Demon Docs `v0.3.5` is the initial public release of the Go implementation: a deterministic documentation-maintenance engine for repository-owned Markdown.

### Core capabilities

- Recursive managed folder indexes with preserved authored prose and parent navigation.
- Local Markdown, wiki, reference, image, and supported HTML link validation and repair.
- Repair after ordinary filesystem moves and explicit link-aware `ddocs mv` operations.
- Repository-scoped document identity, path history, private state, and transactional publication.
- Frontmatter policy enforcement and schema-driven document-body formatting.
- Schema-based document creation and explicit format-conflict resolution.
- Orphan-document health checks.
- Authored codemap management and code-folder reverse indexes.
- Deterministic codemap suggestion research with explicit foreground execution.
- Persisted review decisions, applied-change history, repair blocks, and hash-guarded undo.
- Foreground filesystem watching and an optional self-managed repository demon.

### Reliability and performance

- Atomic, source-hash-guarded rewrites with deterministic ordering.
- Bounded worker pools for validation, link-source parsing, index preparation, reverse-index preparation, and codemap dataset construction.
- Independent frontmatter and document-format cache identities.
- Path-scoped watcher validation for ordinary Markdown frontmatter and format changes.
- Selective validation-cache refresh after generated rewrites.
- A checked-in black-box smoke harness covering source-built and packaged binaries.

### Command behavior

- Bare `ddocs fix` reconciles indexes, links, and configured reverse indexes.
- Frontmatter and document-format mutation requires `--docs`, the specific subsystem selector, or `--all`.
- Generic `fix`, `check`, watcher, and daemon paths do not invoke experimental codemap generation.

### Distribution

- Prebuilt 64-bit Windows and Linux archives.
- SHA-256 checksums for release artifacts.
- Windows and Linux CI, packaged-binary verification, Go tests, vet, and smoke-harness gates.
- The `v0.3.5` release artifacts are distributed under the MIT License. The repository was relicensed under Apache License 2.0 after this tag.

### Known limitations

- No prebuilt macOS artifact is currently published.
- Markdown heading fragments are preserved but not validated.
- Link and folder-index watcher reconciliation can still operate across broader scope than one changed file.
- Codemap suggestion quality is corpus-dependent and remains experimental.
- The smoke harness verifies integration correctness; it is not a stress, soak, or performance test.

[0.3.5]: https://github.com/Lokee86/demon-docs/releases/tag/v0.3.5
