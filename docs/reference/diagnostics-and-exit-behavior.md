---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-798b-bde2-333a82ea7a83
document_type: general
policy_exempt: false
summary: This document describes the behavioral contract for Demon Docs diagnostics, unresolved conditions, mutation refusal, and command success or failure.
---
# Diagnostics and Exit Behavior

Parent index: [Reference](./INDEX.md)

## Purpose

This document describes the behavioral contract for Demon Docs diagnostics, unresolved conditions, mutation refusal, and command success or failure.

## Overview

Demon Docs prefers explicit unresolved states over guesses. Diagnostics identify pending deterministic work, invalid configuration, broken or ambiguous references, missing baselines, coverage gaps, and runtime ownership problems. Output wording may evolve, but the safety behavior must remain stable.

## Diagnostic classes

### Pending update

The current repository differs from the deterministic result for a selected subsystem.

`check` reports the condition and returns non-zero. `fix` may apply it when the update is safe and within scope.

### Frontmatter violation

A Markdown file beneath the configured docs root has missing, invalid, immutable, conditional, duplicate-ID, malformed-block, disallowed-format, or unknown-field state under the selected schema.

`check` reports the path and field without writing. `fix` removes unknown fields only when `unknown_fields = "remove"`, inserts configured defaults or generated values, and restores known immutable values. Existing valid mutable values are preserved. Existing invalid mutable values and required fields without a repair source remain unresolved and make `fix` return non-zero after any safe partial repairs.

Warning-mode unknown fields are preserved and reported without failing solely because of the warning.

### Document-format violation

A Markdown document does not match its selected TOML document schema. Violations include missing sections, incorrect order or nesting, incorrect heading levels, unresolved unknown human-authored sections, duplicate sections, invalid aliases, invalid document-specific schemas, and invalidated document-specific exceptions.

`check` reports without writing. `fix` applies deterministic structure changes, but an unresolved unknown or duplicate section blocks all body-format mutation for that document during the run. Diagnostics list the explicit actions available through `ddocs format`: ignore into the document-specific schema, merge, delete an occurrence, or repair manually.

### Broken target

A recognized local reference resolves to no current target.

The source remains unchanged unless persistent identity or other deterministic evidence identifies exactly one destination.

### Ambiguous target

More than one candidate could satisfy a missing or moved target.

The source remains unchanged. User selection is required.

### Missing baseline

A stateful subsystem lacks the persisted baseline needed to infer history safely. This is not equivalent to an uninitialized Demon Docs repository.

The first link-enabled mutating pass records current state rather than pretending to know earlier moves. It can do so in standalone or initialized mode.

### Orphan document

A normal managed Markdown document has no meaningful inbound link. Index files, draft documents, self-links, and inbound links originating from indexes or drafts do not satisfy the health check.

`check` reports `message: Orphan document: PATH` and returns non-zero. No automatic fix is attempted because Demon Docs does not decide which canonical document should own the relationship.

### Stale review decision or repair block

A persisted decline or block refers to an older evidence fingerprint. The old decision remains auditable, but changed evidence is surfaced for review rather than silently suppressed or applied.

### Invalid configuration or scope

A selected root, path, pattern, or setting is invalid, escapes the permitted boundary, or cannot be resolved consistently.

The command fails before broad repository mutation.

### Coverage or extraction gap

A codemap or reverse-index input is missing, empty, unresolved, or unsupported.

The gap is reported. It is not silently converted into a semantic conclusion.

### Codemap section missing

A selected Markdown document contains no configured codemap heading.

The codemap commands consult the selected effective document schema. If it requires a codemap section, `check` reports the schema-created change and `fix` may create it at the declared position. Otherwise the section remains `missing`; this is not permission to invent a heading or placement.

### Multiple codemap sections

A document contains more than one heading matching the active configured codemap heading set.

Planning fails. Demon Docs does not choose one section or merge them automatically.

### Malformed codemap ownership markers

A codemap section contains duplicated or unbalanced configured codemap marker lines.

Planning fails before publication. The ownership range must be made unambiguous manually.

### Codemap source changed before apply

A selected document no longer matches the source digest used to build the rewrite plan.

The transaction fails rather than overwriting the intervening edit. Rebuild the plan against current content.

### Runtime ownership problem

The repository demon may report stale ownership, missing feeder activity, shutdown, or log/runtime-state problems.

Static commands remain available for verification and recovery after the active owner is stopped or recovered.

## Command behavior

### `status` and config inspection

These commands should succeed when repository and configuration selection can be inspected. Invalid or undiscoverable configuration returns failure without authored-file mutation.

### `check`

Returns success only when every selected subsystem is clean and has the baseline or state required for verification.

Returns non-zero for pending work, unresolved selected-system conditions, frontmatter or document-format violations, orphan documents when links are selected, or reverse-index orphans when reverse indexes are selected. Frontmatter warnings may be printed while the command still succeeds.

### `mv`

Returns success after the explicit move and every planned rewrite complete. Dry-run success means the complete plan was valid. Boundary violations, affected ambiguity, source-hash changes, unsupported sources, destination conflicts, or failed rollback produce failure.

### `fix`

Returns success when safe planned mutations are applied and no unresolved selected-system condition remains. Frontmatter repair may apply deterministic fields and still return non-zero when authored input is required. Body-format repair may apply safe changes to some documents while returning non-zero for documents blocked by unresolved human-authored sections. Ambiguous link items remain unchanged and non-zero.

### `suggestions` and `changes`

Inspection commands fail when requested identifiers cannot be resolved. Selection and undo operations fail rather than bypassing source-hash checks, undo eligibility, or later authored edits.

### `watch`

Performs an immediate reconciliation, then reports later passes. Fatal startup configuration or ownership errors prevent normal watching. Individual filesystem bursts are debounced and serialized.

### Production codemap commands

`codemap inspect` is read-only and returns the computed section status, byte-change status, additions, declines, evidence, tiers, and configured removals.

`codemap check` returns zero when no selected document would change and one when the production plan contains one or more rewrites. Usage errors return two. Configuration, scope, read, extraction, marker, section, schema, and planning failures return non-zero.

`codemap fix --dry-run` reports the same production plan without writing. `codemap fix` applies prepared rewrites through batch hash preflight and atomic replacement. A clean plan succeeds with zero updated files.

A missing section is a no-op rather than a failure when the selected effective schema does not require a codemap section. A schema-required missing section is planned as `schema-created` and may be created at its declared position. Multiple matching sections, malformed markers, roots outside the docs tree, non-Markdown file roots, and concurrent source changes fail.

### Codemap research commands

Export, benchmark, and precision commands report extraction, dataset, or evaluation failures. They do not mutate authored codemap relationships as a side effect of successful analysis.

## Output expectations

Diagnostics should identify enough context to act safely:

```text
subsystem
source or affected path
problem class
candidate or expected target when known
whether a write occurred
whether user selection is required
```

Machine-readable output is not implied unless a command explicitly documents such a format.

## Failure safety

Demon Docs should fail without broad mutation when:

- configuration, the selected frontmatter schema, or the effective document schema cannot be validated safely;
- a root escapes repository scope;
- expected source content changed after planning;
- a codemap has multiple configured sections or malformed ownership markers;
- a codemap target root escapes the configured documentation tree;
- atomic replacement cannot complete;
- multiple targets are plausible;
- required state cannot be decoded; or
- daemon ownership cannot be established safely.

Per-source expected-content hashes and atomic replacement protect generated link rewrites from overwriting concurrent edits.

## Examples

A clean verification:

```bash
ddocs check
```

A narrow diagnostic pass:

```bash
ddocs check --links
```

Configuration inspection before retrying a failed command:

```bash
ddocs config paths
ddocs config show
```

## Related docs

- [CLI Reference](cli.md)
- [Managed Files and State](managed-files-and-state.md)
- [Configuration Reference](configuration.md)
- [Document Schemas And Format Enforcement](document-schemas.md)
- [Document Health Checks](../guides/document-health-checks.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)

## Notes

Exact exit-code numbers beyond zero versus non-zero should not be inferred from this page unless the CLI explicitly stabilizes and documents a numeric taxonomy.
