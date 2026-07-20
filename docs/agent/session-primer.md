---
author: brian
created: "2026-07-19"
document_id: d3a026ea-a99c-4e5e-9497-cf2a81dff02a
document_type: general
policy_exempt: false
summary: This document provides a short, stable orientation layer for new agent sessions working on Demon Docs.
---
# Session Primer

Parent index: [Agent](./INDEX.md)

## Purpose

This document provides a short, stable orientation layer for new agent sessions working on Demon Docs.

## Overview

Use this primer to get oriented quickly, then move to canonical documentation for current implementation details.

## Rules

- Demon Docs is a deterministic documentation maintenance engine and Go CLI.
- Authored repository files remain the primary product surface.
- Managed ownership is narrow and explicit.
- Read-only commands and dry-run paths must remain non-mutating.
- Ambiguous repairs remain unchanged until a user selects or resolves them.
- Executable entry points stay thin; subsystem mechanics belong in focused internal packages.
- Use canonical guide, reference, architecture, operations, research, planning, development, and limits docs for current facts.
- Use direct repository inspection before guessing from filenames.
- Prefer small ownership moves over broad edits.
- Do not hand-edit managed regions or private state as a convenience.
- Do not treat agent memory as source of truth.

## Canonical docs to use

- [Guides](../guides/INDEX.md)
- [Reference](../reference/INDEX.md)
- [Architecture](../architecture/INDEX.md)
- [Operations](../operations/INDEX.md)
- [Research](../research/INDEX.md)
- [Planning](../planning/INDEX.md)
- [Development](../development/INDEX.md)
- [Limits](../limits/INDEX.md)

## Related docs

- [Current Context](current-context.md)
- [Architecture Rules](architecture-rules.md)
- [Documentation Editing](documentation-editing.md)
- [Prompting and Reporting](prompting-and-reporting.md)
- [Repo Hygiene](repo-hygiene.md)
- [Generated Files](generated-files.md)

## Notes

Keep stable facts in canonical documentation.
