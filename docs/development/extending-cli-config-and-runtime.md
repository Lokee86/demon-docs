---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7058-a264-1a2d6a542dc0
document_type: general
policy_exempt: false
summary: This document defines the safe workflow for adding public commands, nested help, configuration keys and aliases, watcher features, and repository-demon or host integration behavior.
---
# Extending CLI, Configuration, and Runtime

Parent index: [Development](./INDEX.md)

## Purpose

This document defines the safe workflow for adding public commands, nested help, configuration keys and aliases, watcher features, and repository-demon or host integration behavior.

## Overview

CLI, configuration, and runtime extensions are user-visible even when the implementation change is small. Safe changes preserve command selection, non-mutating help, configuration precedence, shared watcher serialization, and lease/token ownership.

## Adding a configuration key

A normal configuration key requires coordinated updates to:

```text
public Config structure
default value
raw TOML pointer field
Load mapping
starter configuration when users should see it
CLI override when one exists
configuration reference
focused tests
```

Use pointer fields in the raw TOML structure when absence must differ from an explicit zero, false, or empty value.

Define:

- default behavior;
- valid values and validation point;
- precedence relative to CLI flags;
- whether an empty value disables behavior or is invalid;
- whether it changes repository scope or mutation scope; and
- whether the starter config should include it.

Do not depend on a Go zero value accidentally matching the desired default.

## Adding a compatibility alias

An alias must have a reason and a removal policy.

Specify:

- canonical key or command;
- alias precedence when both are supplied;
- whether the alias is permanent or transitional;
- diagnostic behavior for conflicts;
- migration documentation; and
- tests covering alias-only, canonical-only, and both-present cases.

Canonical explicit values should normally win over compatibility aliases. Never allow file-order accidents to decide precedence.

## Adding a public command

Public command ownership begins in `internal/app`; executable packages should remain thin argument and I/O adapters.

Required work:

1. Add command routing with explicit usage-error handling.
2. Define read-only versus mutating behavior.
3. Define repository/config discovery requirements.
4. Define stdout, stderr, and exit-code contracts.
5. Add scoped `-h` and `--help` before runtime work begins.
6. Add command-specific parser tests and integration tests.
7. Update the CLI reference and a task guide when normal users need a workflow.
8. Add the command family to documentation coverage.

Nested commands must return their own help. Falling back to the parent summary is a defect.

## Adding an executable alias

An alias such as `demon` may normalize arguments before entering the shared application. It must not fork command semantics.

Test:

- bare invocation;
- help and version handling;
- nested help;
- error-code parity; and
- side-effect parity with the canonical `ddocs` form.

## Adding a watcher feature

A watcher feature is more than an event filter. It must define:

- selection through `watch.Features` or an equivalent concrete seam;
- initial reconciliation before observation;
- relevant files, directories, and control files;
- ignore-policy behavior;
- dynamic directory additions and deletions;
- external watch roots when applicable;
- participation in the shared run lock;
- follow-up scheduling for events during a run;
- generated-write suppression when the feature writes watched files; and
- cancellation and observer-error propagation.

Do not start an independent reconciliation goroutine that bypasses the shared scheduler. Cross-feature serialization is a correctness boundary.

Add filter, scheduler, dynamic-scope, self-loop, error-propagation, and combined-feature tests. Update operations documentation.

## Adding repository-demon behavior

Demon behavior must preserve the single-owner lease and token-safe lifecycle.

Define:

- which runtime record owns the behavior;
- whether it belongs to owner, feeder, heartbeat, shutdown request, or log state;
- freshness and expiry rules;
- token validation;
- behavior under stale-owner recovery;
- read-only status behavior;
- shutdown and no-feeder grace interaction; and
- cross-platform process seams.

Do not make status mutate expired runtime records. Cleanup belongs to the owning lifecycle path.

## Adding a host adapter or feeder

Host integrations should use the feeder contract rather than owning a second daemon process model.

Document:

- client identifier and kind;
- registration, heartbeat, leave, and expiry;
- token storage and secrecy boundary;
- duplicate-session reuse behavior;
- shell or host shutdown cleanup;
- failure when the repository demon is absent; and
- installation/removal procedure.

Keep host-specific scripts outside core lease ownership.

## Output and exit behavior

Every command change must classify outcomes as:

```text
successful result
completed check with drift or threshold failure
usage/configuration error
runtime execution error
```

Use existing command-family conventions rather than inventing new exit meanings. Machine-readable output must remain separate from explanatory diagnostics when the command supports JSON.

## Commands

Focused verification commonly includes:

```bash
go test ./internal/config ./internal/app ./internal/watch ./internal/demon ./cmd/ddocs ./cmd/demon -count=1
make smoke
go test ./... -count=1
go vet ./...
```

Watcher or demon timing changes should also run focused repeated tests to expose contention and expiry regressions.

## Failure modes

- Raw config field cannot distinguish absence from explicit false or zero.
- Alias overrides the canonical key unexpectedly.
- Help performs repository discovery or mutation before returning.
- Nested help prints the parent command.
- New watcher work overlaps another selected subsystem.
- Event filter ignores the control file that changes scope.
- Generated writes trigger an endless watch loop.
- Status deletes runtime state.
- Host adapter creates duplicate feeders or leaks tokens.
- Executable alias drifts from canonical command behavior.

## Code map

- `internal/config/config.go` — defaults, TOML decoding, aliases, selection, and config mutation.
- `internal/app/app.go` and command-specific files — parsing, help, orchestration, output, and exit behavior.
- `cmd/ddocs/` and `cmd/demon/` — executable adapters.
- `internal/watch/` — feature selection, filtering, dynamic scope, and scheduling.
- `internal/demon/` — owner lease, feeders, heartbeats, runtime records, and logs.
- `internal/app/demon*.go` — demon and host-facing command integration.

## Related docs

- [Safe Extension Procedures](safe-extension-procedures.md)
- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Repository Demon](../operations/repository-demon.md)
- [Host Adapter Feeder Integration](../operations/host-adapters.md)
- [Behavioral Contract Matrix](behavioral-contract-matrix.md)

## Notes

A CLI or config addition is not complete when only the parser accepts it. The exact help, precedence, side effects, diagnostics, and recovery behavior are part of the feature.
