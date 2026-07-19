# Repository State and Transactions

Parent index: [Architecture](./README.md)

## Purpose

This document describes the private `.ddocs/` object repository, durable identity and history state, transaction boundaries, and recovery assumptions used by deterministic reconciliation.

## Overview

Demon Docs keeps authored files in the ordinary repository filesystem. Private `.ddocs/` state supplements those files with stable identities, path history, fingerprints, incoming-link groups, reverse-index state, generated-write metadata, review decisions, applied-change history, and repository-demon runtime data.

The private repository exists to make later reconciliation deterministic. It is not an alternate authoring model and does not replace Git.

## Code root

```text
internal/ddrepo/
internal/links/
internal/reverseindex/
internal/review/
internal/demon/
```

## Responsibilities

The private repository and transaction boundary own:

- encoding and decoding private objects;
- stable object identity independent of current path;
- transactional updates to related state;
- persisted path history and fingerprints;
- durable link inventory and incoming-reference groups;
- generated-write records used to distinguish owned updates;
- suggestion decisions, applied-change events, undo eligibility, and repair controls;
- reverse-index state needed for deterministic checks;
- isolation of runtime demon state under `.ddocs/runtime/`; and
- failure behavior when expected state cannot be read or committed safely.

## Does not own

Private state does not own:

- authored Markdown prose;
- Git commit history;
- semantic decisions about documentation ownership;
- ambiguous target selection;
- external target-file contents;
- codemap relationship authorship; or
- product configuration precedence outside the configuration package.

## State model

The broad relationship is:

```text
repository files
+ selected configuration
+ private identities/history
-> deterministic inventory and plan
-> authored-file writes when authorized
-> transactional private-state update
```

State records are implementation details. Public behavior is defined by stable command and safety contracts rather than by requiring users to edit object files directly.

## Transaction flow

A mutating reconciliation should:

```text
read a coherent starting state
scan current repository files
plan deterministic changes
verify expected source content before replacement
apply atomic per-source writes
persist corresponding private state transactionally
report unresolved items without inventing state
```

Link rewrites use source-content hashes and same-directory atomic replacement. This prevents a generated write from silently overwriting a source changed after planning.

## Worktree boundaries

Linked Git worktrees share repository history but have distinct working directories. Runtime and first-mutating-entry behavior must avoid treating one worktree's active filesystem state as another worktree's current paths.

Repository discovery and worktree handling live under `internal/repository/`. The repository demon and mutating commands use those boundaries when selecting local state.

## Rebuildability

Authored files and configuration remain the rebuild source. Private state can be reconstructed, but reconstruction loses historical identity evidence that no longer exists in the current filesystem.

Deleting `.ddocs/` therefore changes capability:

```text
current files can be re-indexed
current links can be validated
past move identity may be lost
first-pass repair history resets
runtime demon ownership resets
```

Recovery should prefer stopping active owners and diagnosing state before deletion.

## Invariants and safety boundaries

- Private state must stay inside the selected repository boundary.
- Object decoding failures must not trigger guessed authored-file rewrites.
- Transactions must not partially claim writes that were not applied.
- Current source content must match the planned expectation before replacement.
- Runtime ownership and durable reconciliation state must not be conflated.
- `.ddocs/` is always excluded from normal documentation/link traversal.
- Git remains the authoritative review and rollback mechanism for authored changes.

## Code map

Primary implementation:

- `internal/ddrepo/codec.go` - private object encoding and decoding.
- `internal/ddrepo/objects.go` - object model.
- `internal/ddrepo/repository.go` - private repository access.
- `internal/ddrepo/transaction.go` - transactional state updates.
- `internal/links/state.go` - durable link state.
- `internal/links/filemeta.go` - file identity and metadata.
- `internal/links/inventory.go` - repository link inventory.
- `internal/links/apply.go` - generated write application.
- `internal/review/store.go` - append-only review events under `refs/ddocs/review`.
- `internal/review/policy.go` - decision and block replay.
- `internal/review/undo.go` - bounded undo construction.
- `internal/repository/repository.go` - repository discovery.
- `internal/repository/worktree.go` - linked-worktree behavior.
- `internal/demon/` - runtime ownership state under `.ddocs/runtime/`.

Related tests:

- `internal/ddrepo/*_test.go`
- `internal/links/reconcile_test.go`
- `internal/links/rewrite_test.go`
- `internal/repository/worktree_test.go`
- `internal/demon/runtime_test.go`

## Tests

Run focused coverage:

```bash
go test ./internal/ddrepo ./internal/links ./internal/review ./internal/repository ./internal/demon -count=1
```

The complete gate remains:

```bash
make release-check
```

## Related docs

- [Managed Files and State](../reference/managed-files-and-state.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Reconciliation Pipeline](reconciliation-pipeline.md)
- [Review Ledger](review-ledger.md)
- [Review Lifecycles](review-lifecycles.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [Repository Demon](../operations/repository-demon.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

The exact on-disk object representation is internal and may evolve. Stable user guarantees concern deterministic behavior, mutation scope, failure safety, and rebuildability.
