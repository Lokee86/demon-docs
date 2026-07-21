---
document_id: 1c1ecf26-df39-4ba0-a3bf-fb9d9bd30481
author: ""
document_type: general
created: 2026-04-02
summary: Lifecycle states for relay work and station telemetry.
policy_exempt: false
---
# Lifecycle

## Notes

A station may disconnect and rejoin without losing its durable work queue.

## Purpose

Describe startup, readiness, degradation, recovery, and shutdown states.

## Overview

Each lifecycle transition is explicit and observable by the operations layer.

## Related docs

- [Architecture](architecture.md)
- [Terminology](terminology.md)
- [API service overview](api-notes.md#overview)
- [Worker responsibilities](worker-notes.md#responsibilities)
- [Monitoring guide](monitoring.md)
