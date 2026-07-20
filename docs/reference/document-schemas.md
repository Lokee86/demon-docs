---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7d22-8914-45d3864f4421
document_type: general
policy_exempt: false
summary: This document defines TOML document schemas, schema selection, document creation, body-format enforcement, document-specific exceptions, and schema migrations.
---
# Document Schemas And Format Enforcement

Parent index: [Reference](./README.md)

## Purpose

This document defines TOML document schemas, schema selection, document creation, body-format enforcement, document-specific exceptions, and schema migrations.

## Overview

A document schema is both the creation template and the ongoing Markdown structure policy for one `document_type`. Frontmatter enforcement and document-body format enforcement remain separate reconciliation operations, but both belong to the document-policy system. The same effective schema can authorize explicit codemap execution to create a missing required codemap section at a deterministic position; it does not authorize heading invention when no codemap section is declared.

Shared, human-authored schemas live in:

```text
.ddocs/schemas/<document-type>.toml
```

Generated, human-editable document-specific schemas live in:

```text
.ddocs/document-schemas/<document-id>.toml
```

`ddocs init` writes Space Rocks-derived starter schemas for `general`, `service`, `planning`, and `index`. `ddocs schema init` adds missing starter schemas to an initialized repository; `--force` replaces existing starter files.

## Schema Selection

For each Markdown document, Demon Docs selects the body schema in this order:

1. The `document_type` frontmatter value.
2. The first matching `[[format.path_rules]]` entry.
3. `[format].default_schema`.
4. No format enforcement when all three are absent.

An explicit metadata value is authoritative. Demon Docs does not fall back to a path rule when `document_type` names a missing schema; it reports that missing schema.

A document-specific schema is keyed by immutable `document_id`, so repository moves do not detach accepted exceptions.

## Creating Documents

```bash
ddocs new DOCUMENT_TYPE PATH
ddocs new --force DOCUMENT_TYPE PATH
```

The schema name and `document_type` are the same value. Creation:

- loads `.ddocs/schemas/<document-type>.toml`;
- creates configured frontmatter through the repository frontmatter schema;
- creates the title, optional parent-index line, required headings, and configured placeholders;
- refuses paths outside the docs root; and
- refuses an existing file unless the interactive overwrite warning is accepted or `--force` is supplied.

The overwrite prompt defaults to no. Noninteractive execution fails without writing unless `--force` is present.

## Shared Schema Structure

```toml
version = 1
name = "service"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"
parent_link = true

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false

[[sections]]
id = "purpose"
heading = "Purpose"

[[sections]]
id = "responsibilities"
heading = "Responsibilities"
placeholder = "- TODO"
aliases = ["Owned Responsibilities"]

[[sections]]
id = "nested-example"
heading = "Nested Example"
parent = "responsibilities"
optional = true
```

Section IDs are stable schema identities. They permit deterministic heading renames and parent/order changes without inferring intent from document prose.

Supported section fields are:

- `id`: required stable identity within the schema.
- `heading`: required canonical Markdown heading text.
- `parent`: optional parent section ID; absent means top level.
- `after`: positioning override used primarily by document-specific schemas.
- `placeholder`: text inserted when a required section is missing.
- `aliases`: accepted alternative heading texts.
- `optional`: prevents missing-section diagnostics and insertion.
- `allow_duplicates`: accepts multiple sibling occurrences.
- `canonicalize_aliases`: makes `fix` replace an accepted alias with the canonical heading.

The heading tree is configurable. A schema may enforce only top-level sections or a complete nested hierarchy.

## Body-Format Operations

```bash
ddocs check --format
ddocs fix --format
ddocs watch --format
```

`--docs` is the umbrella selector for folder indexes, frontmatter, and body format. `--frontmatter` and `--format` select the two policy operations independently.

The format parser uses ordinary Markdown symbols. It does not add or require managed-region markers. Headings inside fenced code, blockquotes, and HTML blocks are ignored completely.

`check` reports violations without writing. `fix` may:

- reorder complete human-authored sections;
- change heading levels to match the schema;
- propagate deterministic schema heading renames;
- canonicalize aliases when configured;
- add missing headings with placeholder text; and
- apply configured duplicate or unknown-section policies.

It does not rewrite prose. Unresolved human-authored sections block body mutation for that document during the current run.

## Unknown Sections

The default `unknown_sections = "manual"` reports an unknown human-authored section and leaves the whole document unchanged. The explicit choices are:

```bash
ddocs format ignore --heading "Appendix" docs/guide.md
ddocs format delete --heading "Appendix" --occurrence 1 docs/guide.md
```

Manual repair means making no format change during that run.

`ignore` creates or updates the document-specific schema. The accepted section becomes a real part of that document's effective schema. By default it is placed after the shared schema's known siblings; multiple accepted unknown siblings retain discovery order. The generated TOML remains editable, so its parent, position, aliases, and duplicate rule may be changed later.

## Duplicate Sections

The shared default `duplicate_sections = "manual"` reports duplicates and blocks document-format mutation. Configurable shared policies are:

- `manual`
- `merge`
- `delete-first`
- `delete-last`
- `keep` or `allow`

The explicit merge operation is always available:

```bash
ddocs format merge --heading "Notes" docs/guide.md
```

Merge retains one heading and combines bodies in discovery order. Exact duplicate list items are removed only when both complete sections are compatible whole-list sections of the same category:

- unordered list with unordered list;
- ordered list with ordered list; or
- task list with task list.

Mixed content, non-list content, or different list categories are concatenated without deduplication. Demon Docs performs no fuzzy or semantic content matching.

Choosing `ignore` for a recognized duplicate records `allow_duplicates = true` in the document-specific schema.

## Schema Renames And Change Invalidation

Heading renames are schema changes only. A section retains its stable `id` while its `heading` changes. Demon Docs stores canonical shared-schema snapshots by fingerprint in its private Git-backed object state, plus the latest successfully reconciled version for migration. A deterministic one-to-one rename changes only the Markdown heading text and preserves the complete section body.

Demon Docs does not infer renames from document wording. Ambiguous structural changes remain unresolved.

Document-specific exceptions record the exact shared-schema fingerprint under which they were accepted. `[format].invalidation_similarity` defaults to `0.5`. Similarity is always measured cumulatively against that accepted snapshot: it is the proportion of stable section IDs whose canonical definitions remain unchanged, divided by the larger schema's section count. When similarity falls below the configured value, `check` reports invalidation and `fix` deletes the document-specific schema before requiring new decisions. Set the threshold to `0` to disable automatic invalidation.

Formatting-only TOML changes do not affect canonical comparison.

## Document-Specific Schema Structure

```toml
version = 1
document_id = "019f7d55-31e4-7d22-8914-45d3864f4421"
shared_schema = "general"
shared_fingerprint = "..."

[[sections]]
id = "local-appendix-1234abcd"
heading = "Appendix"
after = "notes"

[[sections]]
id = "notes"
heading = "Notes"
allow_duplicates = true
```

Document-specific sections extend matching shared section IDs or add local IDs. Shared and document-level aliases are combined into the effective schema.

## Code map

- `internal/documentpolicy/schema.go` — shared and document-specific TOML schema loading and effective-schema composition.
- `internal/documentpolicy/validation.go` — schema identity, hierarchy, policy, alias, and ordering validation.
- `internal/documentpolicy/canonical.go` — canonical fingerprints and similarity comparison.
- `internal/documentpolicy/state.go` — exact schema snapshots and latest migration state in the private Git-backed repository.
- `internal/documentpolicy/selection.go` — docs-root traversal, metadata-first selection, and path fallbacks.
- `internal/documentpolicy/markdown.go` — source-preserving Markdown heading parsing.
- `internal/documentpolicy/classification.go` — section identity, location, and reparenting analysis.
- `internal/documentpolicy/enforce.go` — ordering, placeholder insertion, aliases, heading levels, and schema renames.
- `internal/documentpolicy/merge.go` — duplicate-section merge and list-only exact deduplication.
- `internal/documentpolicy/plan.go` — invalidation, check/fix planning, and transactional application.
- `internal/documentpolicy/resolve.go` — explicit ignore, merge, and delete operations.
- `internal/documentpolicy/codemap.go` — effective-schema codemap placement for explicit codemap execution.
- `internal/documentpolicy/create.go` — schema-based document creation.
- `internal/documentpolicy/files.go` — starter-schema publication and transactional writes.

## Related docs

- [Using Document Schemas](../guides/document-schemas.md)
- [CLI Reference](cli.md)
- [Demon Docs Configuration](configuration.md)
- [Frontmatter](frontmatter.md)
- [Managed Files and State](managed-files-and-state.md)
- [Application Orchestration](../architecture/application-orchestration.md)

## Notes

Shared schemas are project policy. Document-specific schemas are explicit, persistent exceptions rather than silent parser suppression.
