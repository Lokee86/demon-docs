---
author: brian
created: "2026-07-19"
document_id: 2e484814-3e3d-4a96-90a4-1ff820160580
document_type: general
policy_exempt: false
summary: This document defines mandatory architecture and seam-editing guardrails for agents changing Demon Docs code, structure, or ownership boundaries.
---
# Architecture and Seam Editing Rules

Parent index: [Agent](./README.md)

## Purpose

This document defines mandatory architecture and seam-editing guardrails for agents changing Demon Docs code, structure, or ownership boundaries.

## Overview

Use these rules to decide where a change belongs before editing. They govern ownership, scope, extraction, mutation safety, and responsibility boundaries; they do not replace canonical architecture documentation.

## Rules

- Identify the owning system before editing.
- If no clear owner exists, create the smallest concrete ownership seam or stop and report the missing seam.
- Defer mechanics, not ownership.
- Keep executable entry points thin.
- Keep command routing and application composition in `internal/app`; keep subsystem policy and mechanics in the owning package.
- Prefer behavior-preserving extraction before behavior change.
- Do not add subsystem responsibility to broad command, reconciliation, watcher, demon, review, or publication coordinators merely because they are convenient call sites.
- Avoid vague buckets or wrappers such as `helpers`, `utils`, `common`, `misc`, or generic managers unless a concrete responsibility truly requires one.
- Keep one ownership seam per scoped change.
- Preserve deterministic ordering and source-preservation guarantees.
- Preserve behavior unless behavior change is explicitly authorized.
- Do not include unrelated cleanup, formatting churn, opportunistic refactors, or package/folder moves.
- Generated-region, schema, private-state, transaction, and public CLI changes must be explicitly in scope.

### High-risk seam verification

New seams at trust-sensitive boundaries require explicit verification expectations before the seam is complete. This includes repository selection, source-preserving Markdown transformation, link rewriting, filesystem mutation, private object storage, review publication, transaction rollback, watcher scheduling, demon ownership, document schemas, and codemap publication.

For each new high-risk seam, identify:

- the owning system;
- the bypass or reach-through behavior that must remain forbidden;
- focused behavioral or contract tests for the seam;
- the failure, rollback, or recovery boundary; and
- whether the invariant belongs in the behavioral contract matrix.

Package coverage alone is not sufficient when one package owns several independently stateful or mutating flows. Update the canonical focused architecture owner when a new durable lifecycle or mutation boundary appears.

### Line-count guardrails

For handwritten production files:

- Prefer files under roughly 200 lines when practical.
- Around 300 lines, review whether the file still has one clear responsibility.
- Around 350 lines, avoid adding new responsibility unless it clearly belongs there.
- Around 500 lines, treat actively changing files as split candidates.
- For files above 500 lines, prefer extraction or routing over adding more responsibility.
- Generated files, fixtures, pinned reports, and large declarative data files are exempt.

Stop and report before proceeding when ownership is unclear, the work crosses multiple seams, or the required scope expands materially beyond the request.

## Related docs

- [Safe Extension Procedures](../development/safe-extension-procedures.md)
- [Repository Layout](../development/repository-layout.md)
- [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md)
- [Application Orchestration](../architecture/application-orchestration.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Seam-first skill](../../skills/seam-first/SKILL.md)

## Notes

Keep current implementation facts in their owning canonical documents instead of duplicating them here.
