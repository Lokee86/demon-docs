---
author: brian
created: "2026-07-19"
document_id: 8ebd03b5-09be-4ace-bbc5-b1c27c6ebf73
document_type: general
policy_exempt: false
summary: This document owns agent-facing repository safety and hygiene rules for Demon Docs.
---
# Repo Hygiene

Parent index: [Agent](./INDEX.md)

## Purpose

This document owns agent-facing repository safety and hygiene rules for Demon Docs.

## Overview

Demon Docs worktrees may contain unrelated user changes, generated artifacts, private state, or benchmark outputs. Agents must handle repository state carefully while keeping scoped work focused.

## Rules

- Assume the repository may be dirty.
- Do not clean, reset, or revert unrelated user changes casually.
- Always exclude nested `.worktrees/` from broad scans, tests, formatters, documentation tools, and file watching.
- Do not commit normal local binaries, caches, dummy fixture trees, or runtime logs.
- Treat `.ddocs/` runtime and implementation state as private state, not ordinary authored source.
- Do not delete `.ddocs/` as casual cleanup; doing so loses reconciliation history and resets runtime ownership.
- Keep research artifacts separate from production fixtures unless a test deliberately pins them.
- Avoid broad cleanup during scoped work.
- Review generated documentation diffs rather than assuming generated means correct.

## Related docs

- [Repository Layout](../development/repository-layout.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Testing](testing.md)
- [Generated Files](generated-files.md)

## Notes

Permanent repository hygiene belongs here. Temporary warnings belong in `current-context.md`.
