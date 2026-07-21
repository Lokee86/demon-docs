# Deployment

## Purpose

Describe the operator-facing deployment sequence.

## Rollout Checklist

- Confirm the release artifact checksum.
- Confirm station compatibility.
- Confirm rollback ownership.

## Overview

Deploy the relay before updating stations, then verify telemetry continuity.

## Notes

Production rollout requires an explicit rollback owner.

## Related docs

- [Local setup](local-setup.md)
- [Monitoring](monitoring.md)
- [Release plan](release-plan.md)
- [Migration plan](migration.md)
- [API service notes](api-notes.md)
- [Storage retention](docs/old-system/storage/storage-notes.md#retention)
