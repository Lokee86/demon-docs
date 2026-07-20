# AGENTS.md

Guidance for Codex and other coding agents working in this repository.

This file is the short, always-read operating manual. Keep it practical and stable. For deeper or more temporary context, read the linked docs only when the task needs them.

## Project Snapshot

Demon Docs is a deterministic documentation maintenance engine and Go CLI.

- Canonical CLI entry point: `cmd/ddocs/`
- Alias CLI entry point: `cmd/demon/`
- Application orchestration: `internal/app/`
- Focused implementation packages: `internal/`
- Repository-level regression tests: `tests/`
- Project documentation: `docs/`
- Retained benchmark and research artifacts: `research/`
- Task-specific agent workflows: `skills/`

Demon Docs maintains explicit managed documentation surfaces while preserving authored Markdown outside those boundaries. It owns deterministic index blocks, supported path-only link rewrites, explicit repository-contained moves, configured reverse-index regions, adopted codemap sections, and private state under `.ddocs/`.

## Read First

For normal project orientation:

- `README.md`
- `docs/development/repository-layout.md`

For project memory and volatile context:

- `docs/agent/current-context.md`
- `docs/agent/session-primer.md`

For architecture and seam rules:

- `docs/agent/architecture-rules.md`
- `docs/architecture/README.md`
- `docs/development/safe-extension-procedures.md`

For documentation work:

- `docs/agent/documentation-editing.md`
- `docs/documentation-policy.md`
- `docs/documentation-procedure.md`
- `docs/development/documentation-coverage.md`
- `docs/development/behavioral-contract-matrix.md`

For testing and generated-state safety:

- `docs/agent/testing.md`
- `docs/agent/generated-files.md`
- `docs/development/testing-and-fixtures.md`
- `docs/reference/managed-files-and-state.md`

## Current Layout Notes

The Go module path is:

```text
github.com/Lokee86/demon-docs
```

Executable packages under `cmd/` stay thin. Command routing and application composition belong in `internal/app`; subsystem mechanics belong in the narrowest concrete package under `internal/`.

The canonical executable is `ddocs`. The `demon` executable is an alias backed by the same application implementation.

Normal authored repository files remain the primary product surface. Private identity, history, review, transaction, and runtime state lives under `.ddocs/`. Human-authored shared schemas under `.ddocs/schemas/` and human-editable document-specific schemas under `.ddocs/document-schemas/` are explicit exceptions to the normal implementation-owned private-state rule.

## Managed And Generated Files

Do not hand-edit tool-owned regions or internal private state as a convenience.

Demon Docs manages:

```text
content inside configured index marker pairs
configured parent-index navigation lines
path portions of deterministic supported local-link rewrites
configured reverse-index regions
complete adopted codemap section bodies
implementation-owned private state under .ddocs/
```

Authored prose outside explicit managed regions remains human-owned.

Use the owning command, configuration, schema, or implementation path when changing generated behavior. Stop active watch or demon processes before manual state recovery.

## Skills

Task-specific workflows live under `skills/*/SKILL.md`.

Use only the relevant skill for the current task. Do not load every skill for every prompt.

- `skills/micro-prompt/SKILL.md` for normal tiny implementation prompts.
- `skills/seam-first/SKILL.md` for adding or changing behavior without growing orchestration gravity wells.

## Important Conventions

- Preserve deterministic output and stable ordering.
- Preserve authored prose outside explicit managed surfaces.
- Preserve source newline style, final-newline state, and file mode where the owning subsystem guarantees them.
- Keep `check`, inspect, preview, and dry-run paths non-mutating.
- Keep mutation scope explicit and reviewable.
- Do not choose automatically among ambiguous targets.
- Keep command entry points thin.
- Keep application orchestration in `internal/app` and subsystem mechanics in their owning packages.
- Do not accumulate subsystem policy in `internal/app` merely because commands enter there.
- Prefer concrete package ownership over generic helpers, wrappers, or manager layers.
- Keep repository scope and linked-worktree behavior in `internal/repository`.
- Keep managed Markdown parsing and source-preserving transformations in `internal/markdown`.
- Keep forward index reconciliation in `internal/reconcile`.
- Keep local-link inventory, evidence, state, and rewrites in `internal/links`.
- Keep private object storage and transactions in `internal/ddrepo`.
- Keep filesystem scheduling in `internal/watch` and repository-demon ownership in `internal/demon`.
- Keep review decisions and applied-change history in their owning review seams.
- Keep codemap extraction, evidence, production ranking, execution, benchmarks, and precision evaluation separated by their existing package boundaries.
- Existing codemap links are preserved by default. Confidence pruning remains opt-in.
- A declined codemap suggestion suppresses the unchanged future suggestion; it does not remove an existing link.
- Normal reconciliation, watch, and demon automation must not invoke codemap generation.
- Always exclude nested `.worktrees/` from repository-wide scans, tests, formatters, documentation tools, and file watching.
- `.git/`, `.ddocs/`, `.obsidian/`, and `logseq/` remain hard traversal exclusions; `.docignore` adds repository-configurable exclusions.
- Do not revert unrelated user changes.
- Keep changes scoped and behavior-preserving unless behavior change is explicitly requested.

## Architecture / Seam Discipline

Read `docs/agent/architecture-rules.md` before adding ownership seams, moving packages, changing repository scope, changing mutation or transaction behavior, altering watcher/demon coordination, or editing broad command-orchestration files.

Core rules: identify the owning system first; keep policy in that owner and routing/composition thin; create the smallest concrete seam or stop and report when ownership is unclear; prefer behavior-preserving extraction; keep one seam per scoped change; preserve behavior unless explicitly authorized; and avoid unrelated cleanup, churn, refactors, or moves. Stop when work crosses seams or expands materially.

## Where To Look First

Command routing and public CLI behavior:

- `cmd/ddocs/`
- `cmd/demon/`
- `internal/app/`
- `docs/reference/cli.md`
- `docs/architecture/application-orchestration.md`

Repository and configuration selection:

- `internal/config/`
- `internal/repository/`
- `internal/ignore/`
- `docs/reference/configuration.md`
- `docs/architecture/repository-scope-and-worktrees.md`

Documentation reconciliation:

- `internal/scan/`
- `internal/markdown/`
- `internal/reconcile/`
- `docs/architecture/reconciliation-pipeline.md`
- `docs/architecture/managed-markdown-transformation.md`

Links, moves, and publication:

- `internal/links/`
- `internal/filetxn/`
- `internal/ddrepo/`
- `docs/architecture/markdown-link-reconciliation.md`
- `docs/architecture/stateless-move-transaction.md`
- `docs/architecture/generated-rewrite-publication.md`

Watch and repository demon:

- `internal/watch/`
- `internal/demon/`
- `docs/architecture/watch-scheduler.md`
- `docs/architecture/repository-demon-lease-protocol.md`
- `docs/operations/repository-demon.md`

Codemaps and research:

- `internal/codemap/`
- `internal/codemaprecommend/`
- `internal/codemaprun/`
- `internal/codemapbench/`
- `internal/codemapprecision/`
- `docs/architecture/codemap-pipeline.md`
- `docs/architecture/codemap-managed-execution.md`
- `docs/research/README.md`

## Agent Behavior Notes

- Open and read only the files needed for the requested edit.
- Follow direct references to callers, tests, state owners, and canonical docs when required to understand the named boundary.
- Do not turn a small edit into an unrequested repository-wide audit.
- Focused, safe terminal checks are allowed when useful.
- Avoid destructive Git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- Use direct workspace read/write tools for routine inspection and bounded edits.
- Do not delegate routine inspection or verification to a sub-agent.
- Do not create broad refactors when a small change solves the request.
- If a task starts to balloon, stop and report why before adding large amounts of code.
- Preserve current behavior unless the user explicitly requests a change.
- Keep implementation slices small enough for quick review.
- Update canonical documentation, documentation coverage, and behavioral contracts when the changed boundary requires them.
- Do not report tests or commands as passing unless they were actually run.
- When completing a numbered prompt, place the exact completion heading at the bottom of the report.

## Default Agent Report

```text
Changed files:
- ...

Unexpected files touched:
- none / ...

Notes:
- ...

**<NOT >COMPLETED PROMPT X**
```
