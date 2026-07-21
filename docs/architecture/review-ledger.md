---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7a6a-a79d-308aa88e877c
document_type: general
policy_exempt: false
summary: This document describes the repository-local review ledger for suggestions, applied repairs, user decisions, undo history, and repair controls.
---
# Review Ledger

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes the repository-local review ledger for suggestions, applied repairs, user decisions, undo history, and repair controls.

## Overview

The review ledger is stored in the private `.ddocs/` Git object repository under `refs/ddocs/review`. It records an append-only audit history without creating commits in the user's normal Git history.

A suggestion is an unresolved choice. A repair is a concrete transformation ready to apply. An applied repair becomes a change event.

This page owns the ledger overview and durable concepts. Exact suggestion, decision, applied-change, undo, repair-block, stale-decision, and stale-block transitions are owned by [Review Lifecycles](review-lifecycles.md).

## Code root

```text
internal/review/
internal/app/review_*.go
internal/links/review_*.go
internal/codemap/insert.go
```

## Responsibilities

The review boundary owns:

- stable suggestion and candidate identities;
- decision fingerprints and stale-decision detection;
- declined issue and candidate persistence;
- applied-change events with before/after content and hashes;
- reconciliation-run grouping;
- file-level and repair-level transformation metadata;
- undo eligibility by depth and age;
- hash-guarded undo construction;
- repair blocks and stale-block behavior; and
- compare-and-swap advancement of the review reference.

## Does not own

It does not own:

- candidate generation or ranking;
- link target resolution;
- arbitrary semantic judgments;
- Git history for user-authored commits;
- arbitrary historical selective reverts through later edits; or
- automatic removal of authored codemap relationships.

## State model

```text
detected issue
├── deterministic repair
│   └── applied change
└── ambiguous suggestion
    ├── selected candidate
    │   └── repair
    │       └── applied change
    ├── declined candidate
    └── declined issue
```

Link repair remains automatic when one deterministic target exists. Multiple plausible targets become `link_repair` suggestions. Codemap missing-link candidates become `codemap_link` suggestions and are never inserted automatically.

Selection immediately converts a candidate into the normal repair path. There is no durable accepted-suggestion state.

## Decision persistence

Declines are keyed by stable relationship and evidence fingerprint.

```text
same relationship + same fingerprint
-> remain declined

same relationship + materially changed fingerprint
-> prior decision becomes stale and reviewable
```

This suppresses repeated unchanged suggestions without permanently hiding changed evidence.

## Applied-change events

Every generated rewrite records:

- reconciliation run;
- repair kind and selection mode;
- source identity and path;
- originating suggestion when applicable;
- before and after SHA-256 hashes;
- before and after file blobs;
- individual repair transformations; and
- related target identities and paths.

Events from one append are encoded together in `batch.json` and published through one review commit. The batch carries ordered event payloads and optional before/after snapshots; history APIs expand it back into individual event records that share the batch commit hash. Legacy commits containing `event.json` and optional `before` / `after` blobs remain readable.

## Undo model

Undo supports:

```text
one reconciliation run
one file change
one repair within a file change
```

The current file must match the recorded after hash. Run-level undo preflights every file before writing any of them. Undo appends a new event; it never deletes or rewrites prior history.

Eligibility is controlled by `[review].undo_depth` and `[review].undo_max_age_days`. Limits affect reversibility, not audit retention.

## Repair blocks

Undo alone may permit the same deterministic repair to be discovered again. A repair block records the exact source relationship and repair fingerprint.

An unchanged blocked repair is not applied. Materially changed evidence makes the block stale and reviewable rather than silently applying or permanently suppressing a different repair.

## Invariants and safety boundaries

- Existing authored codemap links are never removal suggestions.
- A suggestion decision cannot bypass normal source-hash checks.
- Undo cannot overwrite later edits.
- Review history is append-only.
- Concurrent appends use compare-and-swap reference advancement.
- Private review events do not modify the user's normal Git branch history.
- Expired undo eligibility does not erase audit records.

Automatic private-object compaction is disabled for normal review writes. The
daemon and CLI can read the same bare object store from separate processes,
and pack replacement is unsafe until those readers and writers share a
cross-process lock. Explicit storage tests still verify that controlled
single-process compaction retains review commits and undo blobs.

## Code map

- `internal/review/model.go` - review event, decision, repair, and control models.
- `internal/review/fingerprint.go` - stable evidence and relationship fingerprints.
- `internal/review/store.go` and `store_batch.go` - Git object storage, single-commit batch encoding, legacy-compatible history expansion, and compare-and-swap append.
- `internal/review/policy.go` - decision and block replay.
- `internal/review/undo.go` - bounded undo construction and eligibility.
- `internal/links/review_suggestions.go` - ambiguous link suggestion generation.
- `internal/links/review_record.go` - applied link-repair event recording.
- `internal/app/review_suggestions.go` - suggestion CLI.
- `internal/app/review_changes.go` - change inspection CLI.
- `internal/app/review_undo.go` - undo command integration.
- `internal/app/review_controls.go` - repair block controls.
- `internal/codemap/insert.go` - selected codemap candidate insertion.

## Tests

Focused coverage includes single-commit batch append/replay, legacy per-event history compatibility, nil/empty snapshot preservation, constant object growth, compaction retention, undo eligibility, suggestion CLI, link integration, codemap insertion, blocks, and concurrent history behavior.

```bash
go test ./internal/review ./internal/links ./internal/app ./internal/codemap -count=1
```

## Related docs

- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Review Lifecycles](review-lifecycles.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Configuration Reference](../reference/configuration.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Codemap Missing-Link Evidence](../research/codemap-evidence.md)
- [Repository State and Transactions](repository-state-and-transactions.md)

## Notes

The review ledger complements normal Git; it does not replace commits, branches, or repository-level rollback for broader authored work.
