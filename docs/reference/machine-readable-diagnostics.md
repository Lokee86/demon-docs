---
author: brian
created: "2026-08-28"
document_id: 019fb711-9205-7fe0-bdd5-9ef5244bcb3e
document_type: general
policy_exempt: false
summary: Versioned JSON diagnostic contract for link and index verification, including stable envelope fields, diagnostic codes, severity values, and compatibility rules.
---
# Machine-Readable Diagnostics

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines the first stable machine-readable diagnostic contract exposed by Demon Docs.

## Current scope

Schema version 1 currently covers links and documentation indexes:

```bash
ddocs check --links --output-format json
ddocs check --indexes --output-format json
ddocs check --links --indexes --output-format json
```

Text remains the default output. JSON accepts any selected combination of the migrated link and index subsystems. Frontmatter, document-body format, or reverse-index selection is rejected with usage exit code `2` rather than returning a partial machine report.

`fix` and `watch` do not yet expose this JSON diagnostic contract.

## Envelope

A completed supported check writes exactly one JSON object to stdout followed by a newline:

```json
{
  "schema_version": 1,
  "command": "check",
  "status": "failed",
  "exit_code": 1,
  "diagnostics": []
}
```

Stable envelope fields:

- `schema_version` — integer schema identifier. Version 1 is the contract documented here.
- `command` — currently `check`.
- `status` — `passed` or `failed`.
- `exit_code` — `0` for a clean completed check and `1` for completed verification with findings.
- `diagnostics` — ordered array of structured diagnostic objects; clean reports use an empty array rather than `null`.

Usage errors still return `2` through the normal CLI error surface. Configuration, filesystem, or runtime failures that prevent a completed diagnostic pass are not represented by schema 1.

## Diagnostic object

Every diagnostic contains:

```json
{
  "code": "links.broken",
  "severity": "error",
  "subsystem": "links",
  "message": "Local link target does not exist"
}
```

Optional structured evidence is emitted when available:

```json
{
  "path": "docs/guide.md",
  "line": 42,
  "column": 17,
  "target": "../missing.md",
  "replacement": "../moved.md",
  "candidates": ["docs/a/missing.md", "docs/b/missing.md"]
}
```

Paths are repository-relative slash-separated paths when the underlying link state is repository-owned. Line and column values are one-based. Candidate arrays retain deterministic sorted order.

Consumers should branch on `code`, not parse `message`. `message` is still part of the schema-1 contract and remains stable within that schema version, but it is display text rather than a secondary identifier.

## Severity values

Schema 1 defines three severity strings:

- `error` — verification cannot treat the condition as resolved without additional state or authored input.
- `warning` — deterministic work or a health issue exists and therefore the check may still fail, but the condition is not an ambiguous destructive decision.
- `info` — structured informational record that does not itself imply unresolved state.

Command success is defined by the envelope `status` and `exit_code`, not by counting severities.

## Link diagnostic codes

| Code | Severity | Meaning |
|---|---|---|
| `links.state_uninitialized` | error | Link verification lacks the persisted baseline required to verify move history safely. |
| `links.broken` | error | A recognized local link has no current target. |
| `links.ambiguous` | error | More than one candidate can satisfy the missing target. |
| `links.undefined_reference` | error | An explicit or collapsed Markdown reference label has no definition. |
| `links.repair` | warning | A deterministic moved-target repair is available. `replacement` contains the proposed destination. |
| `links.case_repair` | warning | Target casing differs from the authored path and can be repaired deterministically. |
| `links.repair_blocked` | error | An exact active review block prevents an otherwise deterministic repair. |
| `links.repair_block_stale` | error | A related repair block exists, but its evidence is stale and requires review. |
| `links.repair_selected` | info | A review-selected candidate has been converted into the normal guarded repair path. |
| `links.orphan_document` | warning | A managed Markdown document has no meaningful inbound link under the documented orphan rules. |

All reconciliation codes in the table except `links.orphan_document` are emitted directly by `internal/links`. Orphan-document diagnostics are projected by command orchestration from the same link-health pass. `links.repair_selected` is part of the typed link diagnostic model for review-selected repairs even though public schema-1 JSON is exposed by `check`.

## Index diagnostic codes

| Code | Severity | Meaning |
|---|---|---|
| `indexes.missing` | warning | A managed documentation index does not exist and would be created. |
| `indexes.out_of_date` | warning | An existing documentation index differs from the deterministic managed result. |
| `indexes.stale_entry` | warning | A managed index entry no longer corresponds to current repository content. `section` and `target` identify the stale entry when available. |

Index diagnostics are emitted directly by `internal/reconcile`. The optional schema-1 `section` field identifies the managed index section for diagnostics such as stale entries.

## Ordering

Diagnostics preserve deterministic reconciliation order. In a combined report, index diagnostics are emitted first, then link reconciliation diagnostics, then orphan-document diagnostics. Link diagnostics follow deterministic source/occurrence processing; orphan-document diagnostics are appended in sorted path order.

Consumers must not infer priority from array position. Use `code`, `severity`, path, and position.

## Compatibility policy

Within schema version 1:

- existing field meanings do not change;
- existing diagnostic-code meanings do not change;
- new optional fields may be added;
- new diagnostic codes may be added;
- consumers must ignore unknown optional fields and unknown codes they do not act on; and
- removing or repurposing an existing required field or code meaning requires a schema-version change.

Human-readable output is a separate interface and is not the machine contract.

## Verification

Contract coverage includes:

- clean JSON reports with an empty diagnostic array;
- failing broken-link reports with stable code, severity, path, line, column, and target;
- missing and out-of-date index reports;
- combined link/index reports;
- rejection when an unmigrated subsystem is selected; and
- typed reconciliation tests for link identity evidence and stale index entries.

Run:

```bash
go test ./internal/links ./internal/reconcile ./internal/app -count=1
```

## Related docs

- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [CLI Reference](cli.md)
- [Supported Link Syntax](supported-link-syntax.md)
- [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md)

## Notes

Schema 1 establishes one shared diagnostic envelope across migrated subsystems. Frontmatter, document-body format, reverse indexes, and suitable runtime/configuration failures should adopt this envelope rather than inventing independent JSON formats.
