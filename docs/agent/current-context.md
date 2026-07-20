---
author: brian
created: "2026-07-19"
document_id: 5e940090-17d9-430b-968f-0f3e18c6a059
document_type: general
policy_exempt: false
summary: This file is volatile project memory for agents working on Demon Docs.
---
# Agent Current Context

Parent index: [Agent](./INDEX.md)

## Purpose

This file is volatile project memory for agents working on Demon Docs.

## Overview

Keep this file short and prune it aggressively. Stable facts belong in canonical documentation.

## Current warnings

- The repository may contain unrelated user changes; do not clean or revert them casually.
- Exclude nested `.worktrees/` from broad scans, tests, formatters, documentation tools, and file watching.
- Do not hand-edit implementation-owned `.ddocs/` state while commands, watchers, or the repository demon are active.
- `TestClaimAllowsExactlyOneOwner` has shown intermittent full-suite timing behavior even when focused repetitions pass; do not hide or casually normalize that result.
- Normal reconciliation, watch, and demon paths must not invoke codemap generation.

## Current active gaps

- Broader diagnostics remain incomplete.
- Polyglot code intelligence remains planned.
- Deterministic agent-context integration remains planned.

Use [Current Limitations](../limits/current-limitations.md) and [Roadmap](../planning/roadmap.md) for canonical status.

## Related docs

- [Session Primer](session-primer.md)
- [Repo Hygiene](repo-hygiene.md)
- [Generated Files](generated-files.md)
- [Current Limitations](../limits/current-limitations.md)
- [Roadmap](../planning/roadmap.md)

## Notes

Prune resolved warnings and move durable facts into their canonical owners.
