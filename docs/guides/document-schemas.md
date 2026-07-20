---
author: brian
created: "2026-07-19"
document_id: 3ddd217f-7c9a-4363-b182-970e8ce0480b
document_type: general
policy_exempt: false
summary: Create schema-backed documents, enforce frontmatter and Markdown structure, resolve format conflicts, and verify document-policy convergence.
---
# Using Document Schemas

Parent index: [Guides](./INDEX.md)

## Purpose

This guide explains how to install and customize document schemas, create schema-backed Markdown files, enforce frontmatter and body structure, resolve human-authored format conflicts, and verify convergence.

## Overview

A document schema serves two related purposes:

```text
creation template
-> initial frontmatter, title, parent navigation, headings, and placeholders

ongoing body policy
-> required structure, stable section identities, aliases, nesting, and explicit exceptions
```

Frontmatter enforcement and body-format enforcement remain separate operations. `--docs` runs folder indexes, frontmatter, and body format together. `--frontmatter` and `--format` run the two policy operations independently.

Demon Docs does not rewrite prose to force a schema match. Deterministic structure changes are applied automatically; unknown or duplicate human-authored sections require an explicit decision when the schema policy is `manual`.

## Prerequisites

This schema workflow intentionally uses an initialized repository because `ddocs schema init` installs repository-local starter policy under `.ddocs/schemas/`:

```bash
ddocs init --root docs
ddocs status
ddocs config show
```

Initialization is not required for ordinary index, link, health, move, or foreground-watch operations. It is required here for the schema installer and the repository-local schema paths used by the examples.

The documentation root must exist inside the selected repository. Shared schemas are human-authored project policy; review them before applying body-format enforcement across an existing documentation tree.

## Install starter schemas

`ddocs init` installs starter schemas for `general`, `service`, `planning`, and `index`. To add them to an already initialized repository, run:

```bash
ddocs schema init
```

Existing schema files are preserved. Replace the supplied starter files only when that is intentional:

```bash
ddocs schema init --force
```

Review the resulting TOML under the configured shared schema directory before using `--force` on a customized repository.

## Choose how documents select schemas

A Markdown document selects its body schema in this order:

1. `document_type` frontmatter.
2. The first matching `[[format.path_rules]]` rule.
3. `[format].default_schema`.
4. No body-format enforcement when none applies.

An explicit `document_type` is authoritative. A misspelled or missing named schema is reported rather than silently replaced by a path fallback.

Use metadata when the document type is part of the document's durable identity. Use path rules for repository conventions where location determines the expected shape.

## Customize a shared schema

Edit `.ddocs/schemas/<document-type>.toml`. A minimal schema can define:

```toml
version = 1
name = "service"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"
parent_link = true

[[sections]]
id = "purpose"
heading = "Purpose"

[[sections]]
id = "responsibilities"
heading = "Responsibilities"
placeholder = "- TODO"
```

Treat section `id` as stable. Rename a heading by changing `heading` while retaining the same `id`; this lets Demon Docs migrate the heading without inferring intent from prose.

Before broad use, validate the schema against one representative document:

```bash
ddocs check --format
ddocs fix --format
ddocs check --format
```

## Create a schema-backed document

Create a document by schema name:

```bash
ddocs new service docs/services/example.md
```

Creation loads `.ddocs/schemas/service.toml` and writes the configured frontmatter, title, optional parent link, required headings, and placeholders.

The target must remain inside the documentation root. Existing files are not overwritten by default. Interactive use asks for confirmation; noninteractive use fails unless `--force` is supplied:

```bash
ddocs new --force service docs/services/example.md
```

## Check before applying policy

Run the two policy checks separately while adopting schemas:

```bash
ddocs check --frontmatter
ddocs check --format
```

Frontmatter diagnostics can include missing or invalid fields, unknown-field policy, duplicate `document_id` values, malformed blocks, immutable-value drift, and required values without a safe repair source.

Body-format diagnostics can include missing sections, incorrect order or nesting, wrong heading levels, unknown sections, duplicate sections, invalid aliases, missing schemas, and invalidated document-specific exceptions.

Neither check applies authored-file changes.

## Apply deterministic repairs

Apply safe frontmatter repairs:

```bash
ddocs fix --frontmatter
```

Apply safe body-format repairs:

```bash
ddocs fix --format
```

Or run the complete documentation-policy group:

```bash
ddocs fix --docs
```

Safe repairs may include generated or defaulted frontmatter values, restoration of known immutable values, missing headings, deterministic ordering and nesting, heading levels, aliases, and stable-ID schema renames.

A fix may apply safe changes and still return non-zero when authored input is required. Read every diagnostic before rerunning.

## Accept an unknown section

When a useful human-authored section is not part of the shared schema, preserve it as an explicit document-specific exception:

```bash
ddocs format ignore --heading "Appendix" docs/guide.md
```

Demon Docs creates or updates:

```text
.ddocs/document-schemas/<document-id>.toml
```

The exception follows the immutable document ID across moves. The generated TOML is editable; use it to refine parent, position, aliases, or duplicate policy.

Then verify:

```bash
ddocs fix --format
ddocs check --format
```

## Resolve duplicate sections

Merge duplicate sibling sections when both bodies should remain:

```bash
ddocs format merge --heading "Notes" docs/guide.md
```

Merge preserves discovery order. Exact list-item deduplication occurs only when both complete sections are compatible lists of the same category; Demon Docs does not perform fuzzy prose merging.

Delete one explicit occurrence when it should not remain:

```bash
ddocs format delete --heading "Notes" --occurrence 2 docs/guide.md
```

To allow duplicates for this document instead, use `format ignore` on the recognized duplicate heading. The document-specific schema records `allow_duplicates = true`.

Always inspect the Git diff after merge or delete.

## Verify convergence

Run the full policy cycle:

```bash
ddocs fix --docs
ddocs check --docs
git diff -- docs .ddocs/schemas .ddocs/document-schemas
```

A second fix should report zero updated files:

```bash
ddocs fix --docs
```

When the repository demon or foreground watcher is enabled, the same selected policy systems run after relevant changes. Codemap generation remains explicit and is not part of `--docs`.

## Schema changes and existing exceptions

Demon Docs stores canonical shared-schema snapshots for migration. Stable one-to-one heading renames preserve section bodies.

Document-specific exceptions remember the shared-schema fingerprint under which they were accepted. Similarity is measured cumulatively against that accepted snapshot. When it falls below `[format].invalidation_similarity`, the exception is invalidated so stale local decisions are not silently applied to a materially different shared schema.

After a substantial schema edit:

```bash
ddocs check --format
ddocs fix --format
ddocs check --format
```

Review invalidation diagnostics and recreate only the exceptions that remain appropriate.

## Codemap section placement

A shared or document-specific schema can declare a required codemap section. Explicit `ddocs codemaps` execution may create that missing section at the schema-defined location.

Heading configuration recognizes existing codemap sections; it does not authorize Demon Docs to invent a missing section. A schema with no required codemap section leaves the document unchanged.

Use [Managing Codemaps](managing-codemaps.md) for the separate inspect, dry-run, fix, and check workflow.

## Failure and recovery

### The named schema does not exist

Confirm the `document_type`, configured schema directory, and available `.toml` files. Demon Docs does not fall back from an explicit missing metadata schema.

### Fix reports an unknown section

Choose one action deliberately:

- add the section to the shared schema when it belongs to the whole document type;
- run `format ignore` when it is a valid exception for one document;
- move or rename it manually when it should match an existing schema section; or
- run `format delete` when that exact occurrence should be removed.

### Fix reports duplicate sections

Use `format merge`, `format delete`, or a deliberate duplicate allowance. Do not rely on automatic semantic merging.

### Frontmatter remains unresolved after fix

The field may be mutable but invalid, required without a configured source, or blocked by schema policy. Correct the authored value or add an explicit safe source; Demon Docs preserves existing valid mutable values and does not guess replacements.

### A document-specific exception was invalidated

Review the changed shared schema and the generated exception file. Reapply only decisions that still fit the new structure.

### A file changes during publication

Demon Docs refuses the stale plan. Review the concurrent edit, then rerun check and fix. Guarded rollback will not overwrite content created after Demon Docs' own write.

## Related docs

- [Document Schemas And Format Enforcement](../reference/document-schemas.md)
- [Front Matter Schemas](../reference/frontmatter.md)
- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Reconciliation Command Lifecycle](../architecture/reconciliation-command-lifecycle.md)
- [Generated Rewrite Publication](../architecture/generated-rewrite-publication.md)
- [Managing Codemaps](managing-codemaps.md)
- [Getting Started](getting-started.md)

## Notes

Adopt schemas narrowly. Validate one representative document type before applying body-format policy across the full repository.