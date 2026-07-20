---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-77b0-a469-9dcfdc8e196a
document_type: general
policy_exempt: false
summary: This guide adds Demon Docs verification to CI and explains when to use foreground watch mode or the repository demon during local work.
---
# CI and Automation

Parent index: [Guides](./INDEX.md)

## Purpose

This guide adds Demon Docs verification to CI and explains when to use foreground watch mode or the repository demon during local work.

## Overview

`ddocs check` is the authoritative automation surface because it computes the normal reconciliation plan without writing repository files. Link-enabled checks also fail for orphan managed Markdown documents. Watchers and the repository demon improve local feedback but are not correctness dependencies.

## Prerequisites

- Demon Docs is initialized for the repository.
- A local `ddocs fix` followed by `ddocs check` succeeds.
- CI can install or build the Go command.
- Repository configuration and `.docignore` are committed when they are intended project policy.

## Add the CI check

Build or install the CLI, then run:

```bash
ddocs check
```

A typical job should:

```text
checkout the repository
install the supported Go toolchain
build or install ddocs
run ddocs check from the repository
fail the job for pending reconciliation, unresolved links, reverse-index failures, or orphan documents
```

Do not run `ddocs fix` as an unreviewed CI mutation. CI should report required changes; developers should apply and review them locally.

Generic `ddocs check` does not include production codemap generation. Repositories that require codemap convergence in CI must add an explicit contained check:

```bash
ddocs codemaps check --root docs
```

That command remains read-only and fails when one or more selected codemaps would change. It also reports section, marker, scope, schema-placement, and planning failures. A document whose selected effective schema requires a missing codemap section is stale; a schema without a codemap section leaves the document unchanged.

## Narrow CI adoption

Subsystem selectors permit staged adoption:

```bash
ddocs check --docs
ddocs check --links
ddocs check --reverse
```

Use a temporary narrow check only while introducing the tool. The intended steady state should verify every enabled repository subsystem.

## Local foreground watch

Use foreground watch when one terminal should visibly own automation:

```bash
ddocs watch
```

The watcher performs one immediate reconciliation, observes relevant filesystem events, debounces bursts, and serializes reconciliation passes.

Use `Ctrl+C` or the terminal's normal process control to stop it.

## Repository demon

Use the repository demon when shell or agent activity should keep one repository-local watcher alive without dedicating a terminal:

```bash
demon run
demon --status
demon --logs
```

Install shell hooks only when automatic shell entry/exit feeding is desired. MCP and agent adapters use the feeder acquire, heartbeat, and release contract rather than embedding watcher logic.

Do not run an additional detached wrapper around `ddocs watch` when the repository demon owns the repository.

## Expected result

- CI reports pending documentation or link reconciliation without mutating the checkout.
- Repositories that opt into codemap convergence run explicit `ddocs codemaps check --root ...` alongside generic checks.
- Local watch or daemon automation shortens feedback loops.
- A plain `ddocs check` remains sufficient for normal reconciliation after automation is stopped, but it does not verify codemap generation.

## Failure and recovery

### CI passes locally but fails remotely

Compare configuration selection, working directory, case sensitivity, ignored paths, generated files, and platform-specific path behavior. Run `ddocs config paths` and `ddocs config show` in both environments.

### CI reports uninitialized link state

Establish and commit the intended authored repository changes locally. Private `.ddocs/` runtime/object state is repository-local but normally not a portable CI artifact; CI should not be expected to infer pre-baseline move history.

### A watcher appears to undo manual work

Stop the watcher or demon, inspect logs and the pending reconciliation plan, then make the authored/configuration change that changes the deterministic result. Repeatedly editing generated managed blocks against the configured model will be reconciled back.

### More than one watcher appears active

Use `demon --status` and `demon --logs`. The repository demon has single-owner coordination; unmanaged foreground watchers remain the operator's responsibility.

## Related docs

- [Getting Started](getting-started.md)
- [CLI Reference](../reference/cli.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Repository Demon](../operations/repository-demon.md)
- [Document Health Checks](document-health-checks.md)
- [Managing Codemaps](managing-codemaps.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

CI should verify authored repository state, not conceal drift by committing automatic fixes from the build environment.
