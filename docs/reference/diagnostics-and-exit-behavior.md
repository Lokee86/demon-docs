# Diagnostics and Exit Behavior

Parent index: [Reference](./README.md)

## Purpose

This document describes the behavioral contract for Demon Docs diagnostics, unresolved conditions, mutation refusal, and command success or failure.

## Overview

Demon Docs prefers explicit unresolved states over guesses. Diagnostics identify pending deterministic work, invalid configuration, broken or ambiguous references, missing baselines, coverage gaps, and runtime ownership problems. Output wording may evolve, but the safety behavior must remain stable.

## Diagnostic classes

### Pending update

The current repository differs from the deterministic result for a selected subsystem.

`check` reports the condition and returns non-zero. `fix` may apply it when the update is safe and within scope.

### Broken target

A recognized local reference resolves to no current target.

The source remains unchanged unless persistent identity or other deterministic evidence identifies exactly one destination.

### Ambiguous target

More than one candidate could satisfy a missing or moved target.

The source remains unchanged. User selection is required.

### Uninitialized state

A stateful subsystem lacks the baseline needed to infer history safely.

The first link-enabled mutating pass records current state rather than pretending to know earlier moves.

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

### Runtime ownership problem

The repository demon may report stale ownership, missing feeder activity, shutdown, or log/runtime-state problems.

Static commands remain available for verification and recovery after the active owner is stopped or recovered.

## Command behavior

### `status` and config inspection

These commands should succeed when repository and configuration selection can be inspected. Invalid or undiscoverable configuration returns failure without authored-file mutation.

### `check`

Returns success only when every selected subsystem is clean and sufficiently initialized for verification.

Returns non-zero for pending work, unresolved selected-system conditions, or orphan documents when links are selected.

### `mv`

Returns success after the explicit move and every planned rewrite complete. Dry-run success means the complete plan was valid. Boundary violations, affected ambiguity, source-hash changes, unsupported sources, destination conflicts, or failed rollback produce failure.

### `fix`

Returns success when safe planned mutations are applied and no fatal command error prevents completion. Unresolved ambiguous items may remain reported and may still require a later non-zero `check`.

### `suggestions` and `changes`

Inspection commands fail when requested identifiers cannot be resolved. Selection and undo operations fail rather than bypassing source-hash checks, undo eligibility, or later authored edits.

### `watch`

Performs an immediate reconciliation, then reports later passes. Fatal startup configuration or ownership errors prevent normal watching. Individual filesystem bursts are debounced and serialized.

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

- configuration cannot be selected safely;
- a root escapes repository scope;
- expected source content changed after planning;
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
- [Document Health Checks](../guides/document-health-checks.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)

## Notes

Exact exit-code numbers beyond zero versus non-zero should not be inferred from this page unless the CLI explicitly stabilizes and documents a numeric taxonomy.
