---
document_id: service-node
document_type: service
created: 2026-04-22
policy_exempt: false
legacy_status: active
---
# API Notes

## Responsibilities

- Accept station telemetry.
- Publish validated batches.

## Purpose

Document the API process that currently fronts Astra Relay.

## Does Not Own

- Durable station scheduling.
- Long-term analytics storage.

## Overview

The API validates incoming payloads and forwards accepted work to [[worker-notes|the worker service]]. Storage behavior is described in [storage notes](storage/archive/storage-notes.md#retention).

![System overview](system-overview.jpg)

The same asset is also embedded for Obsidian users: ![[system-overview.jpg]].

## Notes

The file and folder names in this section predate the current service taxonomy.

## Related docs

- [System architecture](architecture.md)
- [Deployment](deployment.md)
- [Monitoring](monitoring.md)
- [Release plan](release-plan.md)
- [[worker-notes|Worker service]]
