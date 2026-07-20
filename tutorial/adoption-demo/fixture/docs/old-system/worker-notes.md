---
document_id: service-core
author: Astra Team
document_type: service
created: 2026-04-18
summary: Working notes for asynchronous task execution.
policy_exempt: false
---
# Worker Notes

## Purpose

Describe asynchronous station task execution.

## Responsibilities

- Pull validated work from the relay queue.
- Dispatch work to connected stations.

## Responsibilities

- Retry transient station failures.
- Emit completion events.

## Overview

The worker consumes API output and reports task state back to [API notes](api-notes.md).

## Does not own

- Request authentication.
- Long-term telemetry retention.

## Notes

The two responsibility sections were written by different maintainers.
