# Benchmark Repository Recommendation

Generated: 2026-07-18

## Decision

Use **`shazow/wifitui` as the first demon-docs context-injection benchmark**.

Use **`crossplane-contrib/provider-sql` as the harder second-stage benchmark** after the experiment harness and scoring criteria are stable.

This is a more useful outcome than selecting one repository and forcing it to cover every difficulty level.

## Primary: shazow/wifitui

Why it fits:

- 8,703 handwritten Go LOC across 9 package boundaries.
- 18 test files and a clean `go test ./...` result on the current snapshot.
- Only one Markdown file: a user-facing README with no architecture or ownership documentation.
- Multiple backends, OS-specific composition, TUI state/layout, CLI behavior, D-Bus integration, themes, and mocks.
- At least nine issue-linked merged PRs in the sampled history.
- Many accepted changes are bounded, multi-file, and test-backed.
- Small enough to understand and evaluate manually without making context selection trivial.

The code is not necessarily poor-quality. The useful defect is the lack of repository-level architecture and ownership context. That is the condition demon-docs is intended to improve.

### Prepared benchmark tasks

Three source snapshots are already pinned to the exact pre-change commits:

| Task | Base commit | Accepted change shape | Baseline |
|---|---|---|---|
| PR #163 — width-aware TUI layout | `4583f965beaac68ed8de4cdfebd614645fcbac8a` | 12 files, +398/-52, six test files | `go test ./...` passes in WSL |
| PR #167 — access-point annotation/theme behavior | `86e3912f192617d5cbbc001f0ca059f710ddbe3d` | 5 files, +67/-1, two test files | `go test ./...` passes in WSL |
| PR #178 — `NO_COLOR` behavior | `111f53dee103724c0bbafd155cdcb51f8ab2a731` | 2 files, +24/-0 | `go test ./...` passes in WSL |

Each task directory contains:

- `source/` — detached pre-change repository snapshot.
- `TASK.md` — original issue text and verification command.
- `metadata.json` — public benchmark metadata.
- `oracle.json` — accepted PR information for post-run evaluation only.
- `baseline-validation.json` — recorded WSL verification result.

Prepared locations:

```text
benchmarks/shazow__wifitui-pr-163/
benchmarks/shazow__wifitui-pr-167/
benchmarks/shazow__wifitui-pr-178/
```

Historical PRs #173 and #174 were not prepared as independent tasks because both address issue #171 from the same base commit. They are better treated later as a composite or multi-patch benchmark rather than pretending each is a complete independent oracle.

## Hard secondary: crossplane-contrib/provider-sql

Why it fits:

- 31,784 handwritten Go LOC across 58 package boundaries.
- 27 test files.
- No architecture documentation and only two Markdown files in the scanned snapshot.
- Cluster-scoped and namespaced implementations mirror each other across APIs, controllers, examples, CRDs, and generated outputs.
- Database behavior, Kubernetes reconciliation, reference resolution, generated code, and schema artifacts create real context-selection pressure.
- Eleven issue-linked PRs and nineteen benchmark-sized PRs in the sampled history.
- The current full `go test ./...` suite passes.

Strong historical tasks include:

- PR #379 / issue #378: conditional PostgreSQL schema requirements across APIs, examples, CRDs, duplicate reconcilers, and tests.
- PR #290 / issue #285: namespaced database-reference resolution with generated resolver and CRD consequences.
- PR #361 / issue #359: PostgreSQL 16 `WITH INHERIT FALSE` support across cluster and namespaced ownership surfaces.

This repository is a stronger test of whether demon-docs prevents agents from missing mirrored ownership and generated artifacts. It is less suitable for the first harness run because dependency installation and the Kubernetes/Crossplane domain add noise to failure analysis.

## Other finalists

### openstack-exporter/openstack-exporter

Good issue/PR inventory and unit-test coverage. The default full suite fails without OpenStack credentials and configuration (`OS_AUTH_URL` and `clouds.yaml`), although normal package tests passed before the integration package failed. Keep it as a future external-service benchmark, not the initial corpus.

### mercuretechnologies/expo-open-ota

The initial scanner underrated its documentation because the documentation site is nested under `apps/docs/`. It has useful cross-stack tasks, but it is not genuinely documentation-poor. Its full Go suite also failed on Windows because test cleanup attempted to remove files still held open by another process. It may still work as a WSL-only cross-stack benchmark later.

## Scanner corrections made

The discovery tooling was corrected after the first run to:

- recognize documentation directories nested below the repository root;
- count `.github/workflows` without treating workflow files as product docs;
- use short hashed clone-directory names;
- enable Git long-path handling per command;
- isolate every scan in a timestamped clone directory;
- avoid changing global Git safe-directory configuration;
- reduce bias toward high-star organization repositories in preliminary ranking.

The original first-run results remain useful, but future runs will use the corrected heuristics.

## Next benchmark experiment

Run the three prepared `wifitui` tasks under identical agent/tool budgets:

1. No injected repository context.
2. Basic repository map only.
3. Full deterministic demon-docs context.
4. Full context with deliberate irrelevant-document noise.

Capture repository searches, files opened, tool calls, input/output tokens, changed files, tests, task completion, and divergence from the accepted upstream patch. The agent must only receive `TASK.md` and `source/`; `oracle.json` is evaluation-only.
