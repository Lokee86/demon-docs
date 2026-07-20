---
author: brian
created: "2026-07-19"
document_id: 8f7b7799-630e-4538-a33b-973fe8a4b4fe
document_type: general
policy_exempt: false
summary: This document tells agents how to handle Demon Docs managed regions, generated artifacts, and private state safely.
---
# Generated Files and Managed State

Parent index: [Agent](./INDEX.md)

## Purpose

This document tells agents how to handle Demon Docs managed regions, generated artifacts, and private state safely.

## Overview

Generated or implementation-owned outputs must be changed through their source configuration, schema, command, or implementation path rather than by editing the resulting artifact as a convenience.

## Rules

- Do not hand-edit content inside managed documentation index markers as the source of truth.
- Do not hand-edit generated reverse-index regions as the source of truth.
- Treat an adopted codemap section body as one unified managed region.
- Do not hand-edit implementation-owned `.ddocs/` objects, refs, transaction state, review state, runtime ownership, leases, heartbeats, or logs while the system is active.
- Human-authored `.ddocs/schemas/` files are policy inputs and may be edited deliberately.
- Generated `.ddocs/document-schemas/` files are explicit human-editable exceptions after creation.
- Use the owning `ddocs` command or implementation path to update generated surfaces.
- Stop active watcher and demon processes before manual recovery.
- Do not commit ordinary local binaries, caches, dummy fixtures, or runtime logs.
- Preserve authored prose outside explicit managed regions.

## Related docs

- [Managed Files and State](../reference/managed-files-and-state.md)
- [Document Schemas and Format Enforcement](../reference/document-schemas.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Testing](testing.md)

## Notes

This is an agent safety guide. Exact managed-surface and private-state contracts remain in the reference and architecture documentation.
