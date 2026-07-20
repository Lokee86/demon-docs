---
author: brian
created: "2026-07-19"
document_id: 3b4f8d66-b4d2-4c86-b478-dd844c1fe101
document_type: general
policy_exempt: false
summary: This document owns concise testing and verification guidance for agents working on Demon Docs.
---
# Agent Testing Rules

Parent index: [Agent](./README.md)

## Purpose

This document owns concise testing and verification guidance for agents working on Demon Docs.

## Overview

Use the smallest focused verification that covers the changed boundary, then use the complete release gate before merge when appropriate. The detailed test inventory and benchmark contracts remain in the development documentation.

## Rules

- Focused, safe terminal checks are allowed when useful.
- Do not report a command as passing unless it was actually run.
- Avoid destructive Git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- Exclude nested `.worktrees/` from repository-wide commands.
- Preserve deterministic test inputs and byte-level fixtures unless the intended contract changed.
- Investigate changed fixture output; do not accept it merely because a generator produced it.
- Read-only command tests must verify that no repository or runtime state was written.
- Mutation tests should verify preflight, publication, rollback, and source-preservation behavior when those boundaries are relevant.
- Update the behavioral contract matrix when a durable invariant or its protecting test changes.

## Focused checks

Run a focused package test with:

```bash
go test ./internal/<package> -count=1
```

Run the complete Go suite with:

```bash
go test ./... -count=1
```

Run the complete local release gate with:

```bash
make release-check
```

Verify repository documentation with:

```bash
go run ./cmd/ddocs fix --docs
go run ./cmd/ddocs check --docs
go run ./cmd/ddocs check --links
```

A second documentation fix should be a no-op after convergence.

## Current caution

`TestClaimAllowsExactlyOneOwner` has failed intermittently in full-suite context while passing focused repetitions. Preserve and report that distinction rather than treating either result as conclusive proof that the timing issue is resolved.

## Related docs

- [Testing and Fixtures](../development/testing-and-fixtures.md)
- [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md)
- [Documentation Procedure](../documentation-procedure.md)
- [Repo Hygiene](repo-hygiene.md)
- [Generated Files](generated-files.md)

## Notes

Detailed test catalogs, benchmark methodology, and release requirements remain in [Testing and Fixtures](../development/testing-and-fixtures.md).
