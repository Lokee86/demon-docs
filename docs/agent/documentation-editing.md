---
author: brian
created: "2026-07-19"
document_id: 1ff3410e-249d-423c-a8a2-7a4f0d8494cd
document_type: general
policy_exempt: false
summary: This document guides agents editing Demon Docs documentation under the repository documentation taxonomy.
---
# Documentation Editing

Parent index: [Agent](./INDEX.md)

## Purpose

This document guides agents editing Demon Docs documentation under the repository documentation taxonomy.

## Overview

Agents must classify documentation by type before editing so each fact lands in the correct owning document set. The documentation policy defines ownership and required shapes; the procedure defines the editing workflow.

## Rules

- Classify facts before writing.
- Update owning `README.md` indexes through the managed reconciliation path.
- Use `stubs/` only for incomplete documents with clear eventual homes.
- Do not put implemented facts only in planning documents.
- Do not present research evidence as a shipped product guarantee.
- Do not put long-lived product facts in agent documents.
- Do not create vague folders such as `misc`, `common`, or `general`.
- Reuse the existing canonical owner before creating another document.
- Implementation-facing architecture and development documents require useful code maps.
- Package coverage does not replace explanation of independent stateful flows, mutation boundaries, concurrency, failure, recovery, and tests.
- Update the documentation coverage map when production ownership changes.
- Update the behavioral contract matrix when a durable invariant or its protecting tests change.
- Keep the root README as a product entry point rather than a duplicate complete manual.

## Related docs

- [Documentation Policy](../documentation-policy.md)
- [Documentation Procedure](../documentation-procedure.md)
- [Documentation Coverage Map](../development/documentation-coverage.md)
- [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)

## Notes

This document is agent workflow guidance, not a replacement for the policy, procedure, or canonical implementation documentation.
