---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-76d7-9f86-feea837c424e
document_type: general
policy_exempt: false
summary: This document defines the safe contributor workflow for extending link syntax, reconciliation statuses, private state records, schemas, review events, and suggestion kinds.
---
# Extending Reconciliation and State

Parent index: [Development](./README.md)

## Purpose

This document defines the safe contributor workflow for extending link syntax, reconciliation statuses, private state records, schemas, review events, and suggestion kinds.

## Overview

Reconciliation extensions are high-risk because they can affect authored source, persisted identity, review history, and future repair decisions. The parser, resolver, renderer, state model, publication mechanism, and review policy are separate seams and must remain consistent.

## Adding supported link syntax

### Ownership

A syntax extension normally touches:

```text
parse occurrence
-> split path from preserved suffix
-> resolve local target
-> store syntax and offsets
-> render only the path portion
-> participate in normal reconciliation
-> participate in stateless move planning
```

Primary files are under `internal/links/`, especially parser dispatch and syntax-specific files.

### Required decisions

Define:

- exact recognized grammar;
- protected contexts where it must not be parsed;
- whether the target can be local, external, directory, or extensionless;
- which source bytes are the replaceable path span;
- what suffix, alias, title, quoting, or wrapper must be preserved;
- ambiguity behavior;
- whether existing stored records require a parser-version bump; and
- whether stateless `ddocs mv` supports the same syntax immediately.

### Required tests

Add positive, negative, and rewrite tests for:

- every accepted form;
- malformed and web/unsupported forms;
- fenced and inline code exclusion;
- path plus query/fragment preservation;
- normal moved-target repair;
- ambiguity refusal;
- stateless file and directory moves; and
- state reload after a converged pass.

Update [Supported Link Syntax](../reference/supported-link-syntax.md) and the behavioral matrix.

## Adding a link status

Statuses are persisted strings in link records and are consumed by diagnostics, review suggestion projection, selection, and later reconciliation.

Before adding a status, define:

- the exact state transition that creates it;
- whether it means resolved, unresolved, reviewable, blocked, or stale;
- whether it fails `check`;
- whether it may generate an automatic rewrite;
- whether it may generate a review suggestion;
- what evidence fingerprint controls staleness;
- how it transitions back to `valid`; and
- compatibility behavior when older binaries read the record.

Add transition tests in `internal/links`, CLI diagnostic tests, and review-policy tests when applicable. Do not reuse an existing status for a materially different meaning.

## Adding a private state record

`internal/ddrepo` stores opaque named records; the owning subsystem defines record names and payload schemas.

### Procedure

1. Choose a stable namespaced record name.
2. Validate that it fits the record-name restrictions and deterministic shard model.
3. Define a payload schema version when the record may outlive one binary release.
4. Classify the record as rebuildable or historically irreplaceable.
5. Read and write through one `ddrepo.Transaction` when the record must publish with related state.
6. Decide how an absent record is distinguished from an empty record.
7. Add round-trip, stale-transaction, and corruption tests.
8. Update [Managed Files and State](../reference/managed-files-and-state.md) and [Repository State and Transactions](../architecture/repository-state-and-transactions.md).

### Publication rule

A transaction commits by compare-and-set against the state reference captured at `Begin`. On conflict, reload current state and recompute; do not blindly retry an old derived payload.

Only dirty shards should change. Stable record serialization and sorted collections are required so equivalent data produces equivalent objects.

## Changing a private schema

Choose one explicit compatibility path:

- additive field with safe zero-value behavior;
- automatic migration from known older schema;
- rebuild from repository facts;
- unsupported-schema refusal; or
- new record name/schema with an upgrade procedure.

Add fixtures for the previous schema. A schema migration is incomplete without tests proving both old-state handling and stable new-state serialization.

Never silently reinterpret a field, status, fingerprint, or path base.

## Adding a review suggestion kind

A new `review.SuggestionKind` requires more than an enum value.

Define:

- relation-key construction;
- suggestion and candidate fingerprint inputs;
- producer ownership;
- candidate ordering;
- issue-level and candidate-level decline semantics;
- when prior decisions become stale;
- whether selection writes authored files;
- the generated `review.Change` and transformations;
- undo behavior;
- block/unblock behavior; and
- CLI rendering and filtering.

Add producer tests, policy replay tests, selection preflight tests, history publication tests, and undo tests. Update the review architecture and user guide.

## Adding a review event or decision action

Review history is append-only under its own Git reference. A new event or action must remain replayable in commit order.

Required decisions:

- whether the existing event schema can represent the addition additively;
- what older readers do with it;
- how current policy projection changes;
- whether batch append remains all-or-nothing;
- whether event blobs require additional before/after content; and
- whether undo or pruning rules change.

Add store round-trip, batch publication, projection, and CLI tests. Do not create a side channel that bypasses review-store append and policy replay.

## Commands

Focused verification commonly includes:

```bash
go test ./internal/links ./internal/review ./internal/ddrepo ./internal/app -count=1
go test ./... -count=1
go vet ./...
```

Run link reconciliation against repository fixtures when syntax or status behavior changes.

## Failure modes

- Parser recognizes a syntax but move rendering corrupts its wrapper or suffix.
- Stored offsets are reused after parser behavior changes without invalidation.
- A new status is never cleared and becomes permanent false drift.
- A private record is published outside the transaction that owns related state.
- Conflict retry reuses stale derived data.
- A suggestion kind has no stable relation key, so declines do not persist.
- An event is written but policy replay ignores it.
- Undo records insufficient before/after data to reverse one transformation safely.

## Code map

- `internal/links/parser.go` and syntax-specific parser files — recognized source forms.
- `internal/links/reconcile.go`, `model.go`, and `state.go` — status transitions and persistence.
- `internal/links/rewrite*.go` and `move*.go` — generated path replacement and stateless moves.
- `internal/ddrepo/` — sharded private object repository and compare-and-set transaction.
- `internal/review/model.go`, `store*.go`, `policy.go`, and `undo.go` — append-only history and policy projection.
- `internal/app/review_*.go` — public review command behavior.

## Related docs

- [Safe Extension Procedures](safe-extension-procedures.md)
- [Supported Link Syntax](../reference/supported-link-syntax.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Review Ledger](../architecture/review-ledger.md)
- [Behavioral Contract Matrix](behavioral-contract-matrix.md)

## Notes

A new syntax or state value should be added only when its lifecycle can be stated precisely. Ambiguous ownership is a reason to add a seam before adding behavior.
