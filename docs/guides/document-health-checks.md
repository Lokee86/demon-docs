---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7608-9064-9752b1288671
document_type: general
policy_exempt: false
summary: This guide explains the orphan-document health check performed by ddocs check when link reconciliation is selected.
---
# Document Health Checks

Parent index: [Guides](./INDEX.md)

## Purpose

This guide explains the orphan-document health check performed by `ddocs check` when link reconciliation is selected.

## Overview

An orphan is a managed Markdown document in the configured documentation tree that has no meaningful inbound link from another normal document. Orphan detection is read-only: it reports navigation gaps and makes `check` fail, but it does not guess where a document should be linked or remove the file.

The health check runs whenever links are selected, including the default `ddocs check` behavior and explicit `ddocs check --links`.

## Prerequisites

- The repository has a configured documentation root.
- Link reconciliation is initialized when using stateful repository checks.
- Canonical documents are distinguishable from configured folder indexes and draft/stub documents.
- The current working tree is suitable for reviewing reported navigation gaps.

## Run the health check

```bash
ddocs check
ddocs check --links
```

An orphan is reported as:

```text
message: Orphan document: docs/path/to/file.md
```

Results are sorted by repository-relative path.

## What counts as a candidate

Candidates are present Markdown-family files indexed beneath the configured docs root, including `.md`, `.markdown`, `.mdown`, `.mkd`, and `.mdx`.

The following are excluded as orphan candidates:

- configured folder index files;
- files under any configured draft/stub folder; and
- non-Markdown indexed assets.

## What counts as an inbound link

A link counts when another present repository Markdown source resolves to the candidate.

These do not count:

- the document linking to itself;
- links originating from configured folder index files; and
- links originating from draft/stub documents.

This prevents generated navigation and unfinished drafts from masking that a canonical document is disconnected from the authored documentation graph.

## Resolve an orphan

Choose the action that reflects authored intent:

```text
link the document from a relevant canonical document
move incomplete material into the configured draft folder
merge its useful content into an owning document and delete it
remove an obsolete document
```

Demon Docs does not select the owning document automatically.

After the authored change:

```bash
ddocs fix --links
ddocs check --links
```

## Expected result

A clean check means every normal managed Markdown document has at least one meaningful inbound authored link, excluding index, draft, and self-links.

It does not prove semantic quality, completeness, or relevance. It verifies graph reachability under the defined rules.

## Failure and recovery

### A document is intentionally standalone

Link it from an appropriate canonical overview or exclude it from the managed documentation scope. There is no per-document semantic exemption in the current health check.

### An index link exists but the document is still reported

Index links are intentionally excluded. Add a meaningful link from another canonical document.

### A draft links to it but it is still reported

Draft sources are intentionally excluded because unfinished material should not establish canonical reachability.

### The docs root is absent

Link-only checks can operate on repository links, but orphan detection requires an existing configured docs root. Verify configuration with `ddocs status` and `ddocs config show`.

## Related docs

- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [CLI Reference](../reference/cli.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Getting Started](getting-started.md)
- [CI and Automation](ci-and-automation.md)

## Notes

The health check identifies missing inbound reachability only. It does not recommend removing an existing link or judge whether a linked document is semantically useful.
