---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-72b1-a1b3-39ffc6b36427
document_type: general
policy_exempt: false
summary: This document routes common Demon Docs extensions to the correct ownership boundary, required documentation, and minimum verification.
---
# Safe Extension Procedures

Parent index: [Development](./README.md)

## Purpose

This document routes common Demon Docs extensions to the correct ownership boundary, required documentation, and minimum verification.

## Overview

Many Demon Docs changes cross more than one file because public behavior, persistent state, source preservation, and deterministic reports are separate seams. Safe extension starts by identifying the owning seam instead of adding behavior to the first convenient package.

Use the focused procedures for implementation details:

- [Extending Reconciliation and State](extending-reconciliation-and-state.md)
- [Extending Codemap Analysis](extending-codemap-analysis.md)
- [Extending CLI, Configuration, and Runtime](extending-cli-config-and-runtime.md)

## Extension routing

| Change | Primary owner | Required companion updates |
| --- | --- | --- |
| New local link syntax | `internal/links` parser, resolver, renderer, state | syntax reference, parser/rewrite/move tests, parser-version decision |
| New link status | `internal/links` reconciliation state model | diagnostics/reference, review projection, state compatibility tests |
| New private state record | owning subsystem plus `internal/ddrepo` transaction | managed-state reference, rebuildability classification, migration tests |
| Private schema change | owning codec/model | compatibility/migration policy, old-state fixture, refusal or migration test |
| New review suggestion kind | `internal/review` model and producer | review policy/projection, CLI output, history compatibility, undo decision |
| New review event or decision action | `internal/review` store/model | replay behavior, schema compatibility, batch publication tests |
| New codemap evidence kind | `internal/evidence` | ranking admission/weight decision, benchmark and precision evaluation |
| Ranking weight, admission, cap, or tier change | `internal/codemapbench` | pinned evaluation, methodology record, evidence/ranking architecture |
| New language dependency adapter | `internal/codemapcorpus` | explicit supported syntax, false-positive fixtures, corpus architecture |
| New report field or meaning | owning report package | schema compatibility decision, canonical serialization tests, reference |
| New configuration key | `internal/config` | defaults, decode, starter config when appropriate, reference and tests |
| Compatibility alias | `internal/config` or CLI parser | migration reference, precedence test, removal condition if transitional |
| New public command or subcommand | `internal/app` and executable entry when needed | scoped help, CLI reference, task guide, exit/diagnostic tests |
| New watcher feature | `internal/watch` plus owning subsystem | event relevance, run-lock participation, dynamic scope, operations docs |
| New daemon/host integration | `internal/demon` / `internal/app` | lease/feeder lifecycle, runtime files, shell/host operations docs |

## Universal workflow

1. Identify the canonical architecture or reference owner.
2. Write down the new invariant and non-ownership boundary.
3. Decide whether persisted or machine-readable state changes.
4. Decide whether old repositories and reports remain readable.
5. Add the smallest concrete seam in the owning package.
6. Add focused tests for positive behavior, refusal behavior, and determinism.
7. Update public help/reference when users can observe the surface.
8. Update the [Behavioral Contract Matrix](behavioral-contract-matrix.md) when the new invariant is durable.
9. Run focused tests, the full Go suite, vet, and documentation checks.

## Compatibility decision

Every extension that changes persisted state or machine-readable output must explicitly choose one of these outcomes:

```text
additive compatible
automatic migration
rebuild from current repository facts
explicit unsupported-schema refusal
breaking schema version with documented upgrade
```

Do not silently reinterpret existing fields or reuse a stored status with a new meaning.

## Determinism decision

Any extension that collects, ranks, renders, or serializes multiple values must define ordering. Map iteration order, filesystem enumeration order, Git history order, and concurrent worker completion order are not acceptable output order.

Sort at the owning boundary and test order independence where inputs may arrive in different sequences.

## Documentation decision

Update the document type that owns the changed fact:

- guide for a new user workflow;
- reference for exact flags, keys, formats, or diagnostics;
- architecture for ownership, flow, state, and invariants;
- operations for runtime behavior and recovery;
- research for methodology or measured evidence;
- limits for a current incomplete surface; and
- planning only for work that is not yet implemented.

A roadmap note or code map is not sufficient current documentation.

## Commands

Focused commands depend on the extension. The minimum final gate is:

```bash
go test ./... -count=1
go vet ./...
go run ./cmd/ddocs fix --docs
go run ./cmd/ddocs check --docs
go run ./cmd/ddocs check --links
```

Changes to executables or CLI routing should also run:

```bash
make smoke
```

Changes to managed indexes should run:

```bash
make regression
```

Changes to codemap ranking or evidence require the pinned benchmark and precision workflow described in [Codemap Precision Governance](../research/codemap-precision-governance.md).

## Failure modes

Unsafe extensions commonly fail by:

- adding a parser case without renderer, move, state, or help coverage;
- changing a stored meaning without a schema decision;
- admitting a new evidence signal without measuring false positives;
- adding a config key without a default or explicit zero-value meaning;
- adding a nested command that falls back to parent help;
- running a new watcher path outside the shared run lock;
- depending on map or filesystem iteration order; or
- documenting an implementation fact only in planning or a pull request.

## Code map

- `internal/app/` — public command routing and orchestration.
- `internal/config/` — configuration contracts and compatibility aliases.
- `internal/links/`, `internal/review/`, `internal/ddrepo/` — authored mutation and private state.
- `internal/codemap*`, `internal/evidence/` — deterministic analysis, reports, and evaluation.
- `internal/watch/`, `internal/demon/` — runtime observation and ownership.
- `docs/reference/`, `docs/architecture/`, `docs/operations/` — canonical current documentation owners.

## Related docs

- [Behavioral Contract Matrix](behavioral-contract-matrix.md)
- [Testing and Fixtures](testing-and-fixtures.md)
- [Documentation Procedure](../documentation-procedure.md)
- [Compatibility and Migrations](../reference/compatibility-and-migrations.md)

## Notes

These procedures describe the minimum safe surface. A change may require additional subsystem-specific tests or documentation when it crosses more than one ownership boundary.
