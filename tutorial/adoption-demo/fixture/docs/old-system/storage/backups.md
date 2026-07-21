---
document_id: 612f8d14-d36f-4912-bf30-7629387a7f77
author: Platform Team
document_type: service
created: 2026-04-28
summary: Snapshot and restoration procedures.
policy_exempt: false
---
# Backups

## Overview

Nightly snapshots are retained independently from the primary telemetry store.

## Purpose

Define recoverability expectations for durable relay data.

## Responsibilities

- Create encrypted snapshots.
- Exercise restoration procedures.

## Does not own

- Application-level retry behavior.

## Related docs

- [Storage service](docs/old-system/storage/storage-notes.md)
- [Migration plan](migration.md)
- [Release plan](release-plan.md)
- [Phase two](phase-two.md)
- [[worker-notes|Worker service]]
