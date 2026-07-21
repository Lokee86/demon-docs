---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7aa1-a158-b3f4ab6b4003
document_type: general
policy_exempt: false
summary: Demon Docs can validate and repair a configurable front matter schema for every non-ignored Markdown document under the configured documentation root. Generated folder indexes are included; generated files are not silently exempted from...
---
# Front Matter Schemas

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines the configurable frontmatter formats, field schema, repair sources, immutable-value behavior, and diagnostics used by Demon Docs.

## Overview

Frontmatter enforcement plans without authored-file mutation during `check` and performs guarded repairs during `fix` or `watch`. It applies a configurable field schema to every non-ignored Markdown document beneath the configured docs root, including generated folder indexes, while preserving the document body and selected existing frontmatter format.

## Formats

YAML and TOML front matter are supported:

```markdown
---
document_id: 019c7c64-87e8-7f45-a7d9-4af639634a2e
author: Documentation Team
document_type: guide
created: 2026-07-19
summary: Explains the repository setup workflow.
policy_exempt: false
---
```

```markdown
+++
document_id = "019c7c64-87e8-7f45-a7d9-4af639634a2e"
author = "Documentation Team"
document_type = "guide"
created = 2026-07-19
summary = "Explains the repository setup workflow."
policy_exempt = false
+++
```

YAML is the default generated format. Existing TOML or YAML blocks retain their current format during repair. Demon Docs never converts a document between formats automatically.

Front matter must be the first block in the file. Malformed blocks, duplicate keys, unsupported formats, and multiple leading blocks are errors.

## Enabling enforcement

Existing configurations without `[frontmatter]` remain behavior-compatible and leave enforcement disabled. New repository starter configs include the default schema explicitly.

```toml
[frontmatter]
enabled = true
default_format = "yaml"
allowed_formats = ["yaml", "toml"]
default_author = "Documentation Team"
unknown_fields = "remove"
```

The configured docs root, `.docignore`, and permanent traversal exclusions define the enforcement scope. `[files].include_patterns` and `[files].exclude_patterns` affect folder-index membership, not frontmatter policy.

## Default schema

```toml
[frontmatter.fields.document_id]
type = "uuid"
required = true
immutable = true
generated = true

[frontmatter.fields.author]
type = "string"
required = true
default_from = "frontmatter.default_author"

[frontmatter.fields.document_type]
type = "string"
required = true
default = "general"

[frontmatter.fields.created]
type = "date"
required = true
immutable = true
generated = true

[frontmatter.fields.summary]
type = "string"
required = true

[frontmatter.fields.policy_exempt]
type = "boolean"
default = false

[frontmatter.fields.policy_exempt_reason]
type = "string"

[[frontmatter.rules]]
when_field = "policy_exempt"
equals = true
require = "policy_exempt_reason"
```

Projects may replace this schema. No document type, status system, team model, or project-specific policy is hard-coded into Demon Docs.

When document-body format enforcement is enabled and `document_type` is missing, frontmatter repair resolves the configured format path rules and then `default_schema`, and writes that selected schema name. Existing non-empty `document_type` metadata remains authoritative. This keeps generated indexes and path-classified planning or service documents from being stamped with the generic frontmatter default before body-format enforcement runs.

Generated folder indexes are Demon Docs-owned files. When such an index lacks required `author` or `summary` values and those fields have no configured source, repair uses `TODO` for the author and `Generated documentation folder index.` for the summary. A configured literal default or non-empty `default_author` takes precedence. These generated-index defaults apply whenever frontmatter enforcement is selected; they do not depend on document-body format enforcement being enabled.

## Field definitions

Each `[frontmatter.fields.<name>]` table supports:

- `type`: `string`, `boolean`, `integer`, `number`, `string_list`, `date`, or `uuid`;
- `required`: rejects missing or empty values;
- `immutable`: records accepted values in private `.ddocs/` state and detects later changes;
- `generated`: currently generates UUIDv7 values for `uuid` fields and current local calendar dates for `date` fields;
- `default`: supplies a literal repair value; and
- `default_from`: currently supports `frontmatter.default_author`.

A field may have at most one value source: `default`, `default_from`, or `generated`.

## Conditional rules

Rules are data, not hard-coded behavior:

```toml
[[frontmatter.rules]]
when_field = "policy_exempt"
equals = true
require = "policy_exempt_reason"
```

The required field must exist and be non-empty when the condition matches. Other repositories can omit this rule or define different conditions.

## Unknown fields

`unknown_fields` supports three modes:

- `remove` — default. `check` reports unknown fields and `fix` removes them.
- `warn` — preserves unknown fields and reports warnings.
- `ignore` — preserves unknown fields silently.

`check` never modifies documents.

## Incremental validation cache

Unchanged clean documents may be served from the durable validation cache under `.ddocs/`. A reusable entry requires the normalized path, content SHA-256, validation engine version, effective frontmatter policy hash, selected shared/document schema hash, and the current immutable-value snapshot to match. Only zero-diagnostic documents are cached. Duplicate document IDs and immutable changes invalidate or bypass reuse, so cache hits do not suppress those diagnostics.

`check` may write cache records inside an already initialized `.ddocs/` store; it does not initialize private state for a standalone repository and does not write authored Markdown, schema inputs, or generated document content. Content, policy, schema, immutable-state, or engine changes invalidate the entry automatically, and records for documents no longer in scope are removed.

## Check and fix behavior

Front matter runs with the documentation system during default reconciliation and `--docs` selection. It can also run independently with `--frontmatter`; `--indexes` does not select frontmatter.

```bash
ddocs check --frontmatter
ddocs fix --frontmatter
ddocs watch --frontmatter

# indexes + frontmatter + document-body format
ddocs check --docs
ddocs fix --docs
ddocs watch --docs
```

`fix`:

- adds missing configured defaults and generated values;
- removes unknown fields only when configured for `remove`;
- preserves existing valid mutable values;
- restores immutable values from recorded Demon Docs state when possible;
- replaces an invalid immutable value only when a recorded or generated replacement exists;
- resolves duplicate generated `document_id` values by preserving the recorded owner, or the lexicographically first path when no owner is recorded, and assigning new UUIDs to the other documents; and
- leaves invalid mutable values unresolved for the author to correct.

A repair can write fixable fields while still returning a non-zero status for unresolved required fields, such as a missing summary with no configured default.

## Identity and private state

The default `document_id` is a generated UUIDv7. `check` reports duplicate IDs without writing. `fix` can safely resolve duplicates when `document_id` remains a generated UUID field: it preserves the document already recorded as the ID owner, falls back to the lexicographically first path when no owner is recorded, and assigns fresh UUIDs to every other duplicate. Duplicate values are never recorded as shared immutable truth. Immutable values are stored in the private `.ddocs/` object repository, keyed by a unique document ID when available and by path otherwise. This lets immutable history follow an identified document across a move without putting private state in normal Git history.

Document IDs provide a stable identity seam used by link inventory and reconciliation. An unambiguous live document can retain identity across a content-changing move, and stale absent private aliases with the same ID can collapse into that live record with merged path history. Normal Markdown paths remain the human-facing link format.

## Rendering guarantees and limits

Demon Docs preserves the Markdown body and original line-ending style. When a front matter block must be rewritten, fields are rendered deterministically. Comments and original key ordering inside the block are not preserved.

Type-specific document policy selected through `document_type` belongs to the separate document-body format operation in the shared document-policy system. See [Document Schemas And Format Enforcement](document-schemas.md).

## Code map

- `internal/frontmatter/` — parsing, validation, repair planning, duplicate-ID detection, atomic writes, and immutable state.
- `internal/config/config.go` and `internal/config/format_selection.go` — schema configuration, starter defaults, and shared metadata/path-rule selection.
- `internal/app/app.go` — `check` and `fix` integration.
- `internal/watch/` — continuous reconciliation integration.

## Related docs

- [Using Document Schemas](../guides/document-schemas.md)
- [Configuration](configuration.md)
- [CLI Reference](cli.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Managed Files and State](managed-files-and-state.md)

## Notes

Frontmatter policy is separate from document-body format enforcement. The two operations share repository scope and the file-transaction boundary but have different schemas, diagnostics, and private-state records.
