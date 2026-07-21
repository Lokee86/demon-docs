---
document_id: migration-plan
document_type: planning
created: 2026-05-12
policy_exempt: false
owner: infrastructure
---
# Migration

## System handoffs

Operations chooses the station cohort; service owners execute and verify the move.

## Purpose

Move stations from the legacy relay without interrupting queued work.

## Ownership Boundary

This plan owns sequencing and rollback criteria.

## Notes

The final cohort should not begin until recovery has been exercised.

## Related docs

- [Roadmap](roadmap.md)
- [Phase one](phase-one.md)
- [Phase two](phase-two.md)
- [API service](api-notes.md)
- [Worker service](worker-notes.md)
- [Storage service](docs/old-system/storage/storage-notes.md)
