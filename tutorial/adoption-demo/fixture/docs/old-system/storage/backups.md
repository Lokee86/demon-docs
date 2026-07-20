---
document_id: backup-service
author: Astra Team
document_type: service
created: 2026-04-28
summary: Backup and restoration expectations for relay storage.
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
