---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7aa1-a158-b3f4ab6b4003
document_type: general
policy_exempt: false
summary: Demon Docs can validate and repair a configurable front matter schema for every non-ignored Markdown document under the configured documentation root. Generated folder indexes are included; generated files are not silently exempted from...
---
# Front Matter Schemas

Parent index: [Reference](./README.md)

Demon Docs can validate and repair a configurable front matter schema for every non-ignored Markdown document under the configured documentation root. Generated folder indexes are included; generated files are not silently exempted from the same repository rules.

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

## Check and fix behavior

Front matter runs with the documentation system during default reconciliation and `--docs` selection.

```bash
ddocs check --docs
ddocs fix --docs
ddocs watch --docs
```

`fix`:

- adds missing configured defaults and generated values;
- removes unknown fields only when configured for `remove`;
- preserves existing valid mutable values;
- restores immutable values from recorded Demon Docs state when possible;
- replaces an invalid immutable value only when a recorded or generated replacement exists; and
- leaves invalid mutable values unresolved for the author to correct.

A repair can write fixable fields while still returning a non-zero status for unresolved required fields, such as a missing summary with no configured default.

## Identity and private state

The default `document_id` is a generated UUIDv7. Duplicate IDs are errors and are not recorded as immutable truth. Immutable values are stored in the private `.ddocs/` object repository, keyed by a unique document ID when available and by path otherwise. This lets immutable history follow an identified document across a move without putting private state in normal Git history.

Document IDs provide a stable identity seam for future link-continuity recovery. Normal Markdown paths remain the human-facing link format.

## Rendering guarantees and limits

Demon Docs preserves the Markdown body and original line-ending style. When a front matter block must be rewritten, fields are rendered deterministically. Comments and original key ordering inside the block are not preserved.

Type-specific document policy selected through `document_type` is intentionally outside this feature and belongs to the document-policy layer.

## Code map

- `internal/frontmatter/` — parsing, validation, repair planning, duplicate-ID detection, atomic writes, and immutable state.
- `internal/config/config.go` — schema configuration and starter defaults.
- `internal/app/app.go` — `check` and `fix` integration.
- `internal/watch/` — continuous reconciliation integration.

## Related docs

- [Configuration](configuration.md)
- [CLI Reference](cli.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Managed Files and State](managed-files-and-state.md)
