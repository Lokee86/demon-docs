---
document_id: storage-service
document_type: service
created: 2026-04-26
policy_exempt: false
---
# Storage Notes

## Purpose

Describe durable telemetry and task-result storage.

## Overview

Storage receives validated batches and completion events from the relay services.

## Does Not Own

- Live station connectivity.
- Operator authentication.

## Data ownership

Storage owns durable telemetry batches, task results, and backup snapshots.

<a id="retention"></a>

Telemetry is retained for thirty days in the demonstration environment.

## Related docs

- [Backup notes](backups.md)
