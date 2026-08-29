---
author: brian
created: "2026-08-28"
document_id: 019fb711-9205-7fe0-bdd5-9ef5244bcb3e
document_type: general
policy_exempt: false
summary: Versioned JSON diagnostic contract for all Demon Docs reconciliation checks, including stable envelope fields, subsystem diagnostic codes, severity values, and compatibility rules.
---
# Machine-Readable Diagnostics

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines the first stable machine-readable diagnostic contract exposed by Demon Docs.

## Current scope

Schema version 1 covers all reconciliation subsystems: links, documentation indexes, frontmatter, document-body format, and reverse indexes:

```bash
ddocs check --links --output-format json
ddocs check --indexes --output-format json
ddocs check --frontmatter --output-format json
ddocs check --format --output-format json
ddocs check --reverse --reverse-root services/api --output-format json
ddocs check --links --indexes --frontmatter --format --output-format json
```

Text remains the default output. JSON accepts any selected combination of reconciliation subsystems. Reverse-index checks retain their normal root and codemap preconditions; failures that prevent a reconciliation plan from being completed remain CLI/runtime errors rather than partial schema-1 reports.

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

Usage errors still return `2` through the normal CLI error surface. The completed audit also keeps invalid/unloadable configuration, missing repository or subsystem preconditions, planning/I/O failures, and private-state publication failures on stderr with exit code `2`. These conditions prevent a completed reconciliation pass and are not represented by schema 1.

## Pre-report failure boundary

Schema 1 is a reconciliation-result contract, not a generic CLI error protocol. The following classes remain outside the JSON report:

| Failure class | Surface | Reason |
|---|---|---|
| invalid flags, arguments, or output selection | stderr, exit 2 | command usage failed before reconciliation selection completed |
| configuration load or validation failure | stderr, exit 2 | the applicable reconciliation policy is not valid |
| missing docs root, reverse-index roots, codemap headings, or equivalent scope precondition | stderr, exit 2 | the requested reconciliation cannot be planned |
| planner, filesystem, resolver, or other I/O failure | stderr, exit 2 | no complete authoritative result exists |
| private-state/cache publication failure | stderr, exit 2 | the command cannot claim a successfully completed check transaction |

Consumers can therefore distinguish three states without parsing prose: exit `0` means a completed passing report, exit `1` means a completed failing report, and exit `2` means no schema-1 reconciliation report was completed. A future machine-readable command-error envelope, if needed, should be a separately defined contract rather than overloading reconciliation diagnostics.

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
  "field": "document_id",
  "section": "Purpose",
  "options": ["ignore", "delete", "repair manually"],
  "line": 42,
  "column": 17,
  "target": "../missing.md",
  "replacement": "../moved.md",
  "candidates": ["docs/a/missing.md", "docs/b/missing.md"]
}
```

Paths are repository-relative slash-separated paths for repository-owned findings. `field` identifies a frontmatter field when applicable. `section` identifies a managed index section or document heading when applicable. `options` contains explicit authored-resolution choices when a subsystem exposes them. Line and column values are one-based. Candidate arrays retain deterministic sorted order.

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
| `links.fragment_missing` | error | The target Markdown file exists, but its GitHub-style section anchors derived from parsed heading text do not contain the authored fragment. |
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

## Frontmatter diagnostic codes

| Code | Severity | Meaning |
|---|---|---|
| `frontmatter.parse_error` | error | The document's frontmatter block cannot be parsed under a supported format. |
| `frontmatter.duplicate_document_id` | error | The same `document_id` is used by more than one managed document. |
| `frontmatter.unknown_field` | warning or error | A field is not defined by the selected schema. Severity follows the configured unknown-field policy. |
| `frontmatter.immutable_mismatch` | error | An immutable field differs from its recorded value. |
| `frontmatter.missing_field` | error | A configured field is absent or empty and has a deterministic repair source. |
| `frontmatter.required_field_missing` | error | A required field is absent or empty and no deterministic source can supply it. |
| `frontmatter.invalid_value` | error | A present field does not satisfy its configured type or value contract. |
| `frontmatter.conditional_required` | error | A conditional schema rule requires a field that is absent or empty. |

Frontmatter diagnostics originate in `internal/frontmatter` and are projected into schema 1 without parsing message text. The optional `field` property identifies the affected frontmatter field. Warning-mode unknown fields do not by themselves fail the completed check; consumers must use envelope `status` and `exit_code` rather than infer command failure from diagnostic count.

## Document-format diagnostic codes

| Code | Severity | Meaning |
|---|---|---|
| `format.schema_load_error` | error | A selected shared document schema cannot be loaded or validated. |
| `format.frontmatter_parse_error` | error | Metadata required for format selection cannot be parsed. |
| `format.schema_selection_error` | error | Document metadata or path policy cannot select a valid schema. |
| `format.document_schema_load_error` | error | A document-specific schema cannot be loaded or parsed. |
| `format.document_schema_identity_mismatch` | error | A document-specific schema names a different document ID. |
| `format.document_schema_shared_mismatch` | error | A document-specific schema extends a different shared schema than metadata selects. |
| `format.schema_snapshot_load_error` | error | A retained accepted shared-schema snapshot cannot be loaded. |
| `format.schema_snapshot_missing` | error | A referenced accepted shared-schema snapshot is unavailable. |
| `format.document_schema_invalidated` | error | A document-specific schema is no longer safe after a sufficiently large shared-schema change. |
| `format.effective_schema_invalid` | error | Shared plus document-specific policy produces an invalid effective schema. |
| `format.ambiguous_section` | error | An authored heading can match more than one schema section. |
| `format.unknown_section` | error | A section is not represented by the effective schema and requires configured or authored resolution. |
| `format.duplicate_section` | error or warning | A non-repeatable section occurs more than once. `warning` is used when the effective schema explicitly accepts the duplicates. |
| `format.section_parent` | error | A known section is under the wrong schema parent or its required parent cannot be identified uniquely. |
| `format.required_section_missing` | error | A required schema section is absent. |
| `format.section_renamed` | error | A stable section ID now has a different canonical heading. |
| `format.alias_canonicalization` | error | A recognized alias should be rewritten to its configured canonical heading. |
| `format.heading_level` | error | A known section uses the wrong heading depth. |
| `format.section_order` | error | Known sections are not in deterministic schema order. |

Document-format diagnostics originate in `internal/documentpolicy`. The optional `section` field identifies the affected heading, and `options` preserves explicit manual-resolution choices for ambiguous or unknown authored sections. As with frontmatter, warning-only format diagnostics do not themselves fail a completed check.

## Reverse-index diagnostic codes

| Code | Severity | Meaning |
|---|---|---|
| `reverse_indexes.target_missing` | error | An in-scope authored codemap target does not exist. |
| `reverse_indexes.target_outside_repository` | error | An authored target resolves outside the repository boundary. |
| `reverse_indexes.target_kind_mismatch` | error | The resolved target kind does not match the authored codemap target kind. |
| `reverse_indexes.pattern_missing` | error | An in-scope authored pattern has no matches. |
| `reverse_indexes.target_ambiguous` | error | More than one target satisfies the authored codemap reference. `candidates` contains deterministic alternatives when available. |
| `reverse_indexes.target_unresolved` | error | A target cannot be projected and has no more specific schema-1 resolution code. |
| `reverse_indexes.target_unavailable` | error | A previously resolved target cannot be inspected or accepted by reverse-index scope validation. |
| `reverse_indexes.symbol_projection_error` | error | A verified symbol target cannot be projected safely because semantic identity or span evidence is inconsistent. |
| `reverse_indexes.index_missing` | warning | A required managed reverse index does not yet exist. |
| `reverse_indexes.index_out_of_date` | warning | An existing managed reverse index differs from the deterministic projection. |
| `reverse_indexes.orphan_code_file` | warning | An eligible in-scope code file has no resolved authored documentation target. |

Target diagnostics carry the authored document path, source line, target, and candidates when applicable. Reverse-index file and orphan diagnostics carry repository-relative paths. Index drift and orphan findings are warnings, but `check` still fails when they are present because reverse-index health explicitly treats either condition as check failure.

## Ordering

Diagnostics preserve deterministic reconciliation order. In a combined report, index diagnostics are emitted first, then frontmatter diagnostics, then document-format diagnostics, then reverse-index diagnostics, then link reconciliation diagnostics, then orphan-document diagnostics. Reverse-index target diagnostics are sorted deterministically before reverse-index file drift and orphan findings; frontmatter and format diagnostics retain deterministic path/field or path/section ordering; link diagnostics follow deterministic source/occurrence processing; orphan-document diagnostics are appended in sorted path order.

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
- failing frontmatter value reports with stable field evidence;
- non-failing frontmatter warning reports;
- document-format reports with stable section codes and authored-resolution options;
- reverse-index target, index-drift, and orphan-code reports;
- combined all-subsystem ordering; and
- typed reconciliation tests for link identity evidence, stale index entries, frontmatter policy findings, document-format conditions, and reverse-index projection health.

Run:

```bash
go test ./internal/links ./internal/reconcile ./internal/frontmatter ./internal/documentpolicy ./internal/reverseindex ./internal/app -count=1
```

## Related docs

- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [CLI Reference](cli.md)
- [Supported Link Syntax](supported-link-syntax.md)
- [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md)

## Notes

Schema 1 establishes one shared diagnostic envelope across every reconciliation subsystem. The runtime/configuration audit intentionally leaves failures that prevent a check plan from completing outside this report contract. A future machine-readable command-error envelope should be added only as a separate, explicit contract.
