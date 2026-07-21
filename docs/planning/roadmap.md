---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7366-a679-359c693253d2
document_type: general
policy_exempt: false
summary: This document summarizes shipped product areas, active work, near-term hardening, back-burnered architecture, and later product tracks without acting as the canonical reference for implemented behavior.
---
# Demon Docs Roadmap

Parent index: [Planning](./INDEX.md)

## Purpose

This document summarizes shipped product areas, active work, near-term hardening, back-burnered architecture, and later product tracks without acting as the canonical reference for implemented behavior.

## Overview

The roadmap is a sequencing and status document. Current product summaries link to canonical architecture, operations, reference, and research pages; detailed shipped behavior belongs in those documents.

## Current status

Active roadmap. The current branch includes stateless refactoring, orphan health checks, the review ledger, strict frontmatter policy, document-format schemas, reverse-index health, and explicit production codemap execution with schema-governed missing-section creation. Polyglot code intelligence and context delivery remain back-burnered or later work.

## Ownership boundary

This roadmap owns project sequencing and status summaries. It does not own exact CLI contracts, current implementation mechanics, benchmark methodology, or operational recovery procedures.

This roadmap describes the current product state and the next implementation tracks. It separates shipped behavior, active tuning work, bounded near-term work, and larger back-burnered architecture so planned work is not mistaken for released functionality.

## Current Product: Implemented in the Current Branch

### Documentation-tree reconciliation

- Go is the sole implementation and supported runtime.
- Recursive folder indexes describe direct files, draft/stub files, and child folders.
- Managed Markdown sections are the only generated regions; authored content outside them is preserved.
- Parent navigation links keep folder indexes and configured indexed documents connected to their owning index.
- `check`, `fix`, and foreground `watch` expose the same deterministic reconciliation core.
- The repository demon provides single-owner detached watcher lifecycle while shell or agent feeders remain active, without becoming a correctness dependency.
- `-d` / `--docs`, `-i` / `--indexes`, `-l` / `--links`, and `-r` / `--reverse` select reconciliation subsystems independently; `--docs` includes indexes, frontmatter, and document-body format, while `--indexes` selects indexes only.
- Existing index descriptions and link syntax are preserved where entries remain stable or moves are unambiguous.

### Document health checks

- Link-enabled `check` reports normal managed Markdown documents without a meaningful inbound link.
- Folder indexes, draft documents, self-links, and inbound links originating from indexes or drafts do not mask an orphan.
- Results are deterministic and path-sorted.
- Health checks are diagnostic only; Demon Docs does not guess where an orphan should be linked or remove it.

See [Document Health Checks](../guides/document-health-checks.md).

### Repository-local link reconciliation and refactoring

- Repository Markdown is scanned subject to `.docignore` and permanent traversal exclusions.
- Supported local forms include inline links, images, reference definitions, explicit and collapsed reference uses, path-based wiki links, wiki embeds, and common local HTML `href`, `src`, and `poster` targets.
- Stable internal file identities, path history, fingerprints, and incoming-link groups support deterministic move reconciliation without embedding IDs in source files.
- Link labels, titles, aliases, angle wrapping, query strings, fragments, source newline style, and surrounding prose are preserved.
- Undefined explicit or collapsed reference labels are reported.
- Generated rewrites use bounded concurrency while retaining deterministic planning, source-hash checks, and atomic per-file replacement.
- `ddocs mv` explicitly moves a file or directory and rewrites affected incoming and moved-source links without requiring or creating `.ddocs/` state.
- Move planning supports dry runs, repository boundaries, case-only renames, affected ambiguity refusal, source-hash preflight, and best-effort rollback.

See [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) and [Stateless Document Refactoring](../guides/document-refactoring.md).

### Reverse code-folder indexes

- Reverse indexes project authored codemap references back onto configured code folders and files.
- Recursive repository-relative roots, repeated `--reverse-root` overrides, nested `.docignore`, and configurable codemap headings are implemented.
- `check`, `fix`, and `watch` support `-r` / `--reverse` independently or alongside documentation indexes and links.
- Missing codemap sections, empty matching sections, unresolved targets, and coverage gaps remain explicit diagnostics.
- Symbol-level projection, move-aware authored-reference repair, and richer coverage reports remain later work.

See [Code-Folder Reverse Indexes](../architecture/reverse-indexes.md).

### Repository demon

- One fresh owner serves each initialized repository-local `.ddocs/` state directory.
- Shell and generic agent feeders keep the watcher active while work is in progress.
- Detached startup, stale-owner recovery, shutdown grace, status, linked-worktree bootstrap, and bounded logs are implemented.
- Bash and PowerShell hooks translate shell entry and exit into feeder registration.
- The daemon remains optional; `check`, `fix`, and foreground `watch` remain authoritative recovery and CI surfaces.

See [Repository Demon](../operations/repository-demon.md).

### Managed codemap execution and deterministic missing-link research

- Existing configured codemap sections can be inspected, dry-run, updated, and checked through an explicit foreground command family.
- Demon Docs adopts the complete section as one managed artifact rather than separating authored and generated links.
- Existing valid links remain by default; independent undiscovered-link and low-score pruning settings are available but disabled by default.
- Selected non-declined `hard_link` and `context` recommendations are added automatically.
- Shared review fingerprints suppress unchanged declined additions, while materially changed evidence may be reconsidered.
- Fenced Space Rocks-style maps remain fenced and bullet maps retain their local prefix.
- File and directory scopes are contained beneath the docs root and publish through content-addressed transactional rewrites.
- Generic reconciliation, watch, and repository-demon paths never invoke codemap generation.
- The repository corpus adapter collects paths, dependency neighbours, declared symbols, source/test relationships, related-document targets, and bounded Git co-change evidence.
- Holdout and precision tooling consume the same production ranker used by the explicit writer.
- The public codemap command now uses the document-policy schema provider to create required missing sections at deterministic schema positions.

The current curated Space Rocks sample contains 150 labeled recommendations. The retained baseline has 68 hard-link recommendations, 75.00% hard-link strict precision, 98.53% hard-link relevance, 72.86% labeled-valid hard-link recovery, 82 context recommendations, and 10/10 canonical hidden-link recovery. The ordinary cross-repository holdout recovers 11/18 links. These numbers describe pinned samples, not universal quality.

See [Managing Codemaps](../guides/managing-codemaps.md), [Codemap Managed Execution](../architecture/codemap-managed-execution.md), and [Codemap Missing-Link Evidence](../research/codemap-evidence.md).

### Suggestions, applied changes, and undo history

- Ambiguous link repairs and codemap missing-link candidates are exposed through `ddocs suggestions`.
- Selecting a candidate remains a compatibility path that converts one candidate into the normal hash-guarded recorded repair lifecycle.
- Production `codemap fix` does not require selection; it replays shared decline policy and applies the unified managed-section transaction directly.
- Declined issues and candidates persist by stable relationship and evidence fingerprint; materially changed evidence becomes stale rather than silently reappearing or remaining permanently hidden.
- Normal generated rewrites record Git-backed applied-change events under `.ddocs`, including before/after blobs, per-repair transformations, source identity, related targets, and reconciliation run.
- Unified codemap fix rewrites currently use command output and Git history as their audit surface rather than ordinary `ddocs changes` events.
- `ddocs changes` supports inspection, related-target queries, file-level undo, individual-repair undo, and whole-run undo.
- Undo refuses to overwrite later edits and may block the exact repair from being reapplied. Changed repair evidence produces a stale block that remains reviewable.
- Undo eligibility is configurable by depth and age while audit history remains inspectable.

See [Review Ledger](../architecture/review-ledger.md) and [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md).

## Active Work

### Codemap execution hardening and broader corpus validation

Near-term goals are:

- expand end-to-end schema-placement command tests and diagnostics;
- dogfood document-specific codemap placement exceptions on representative repositories;
- dogfood explicit generation on representative Demon Docs and Space Rocks sections;
- compare each scoring change against pinned precision and holdout samples;
- preserve deterministic output, evidence fingerprints, whole-section ownership, and default no-pruning behavior;
- expand evaluation beyond one repository without treating unlabeled output as ground truth; and
- use Demon Docs' own refreshed code maps as a development corpus, not as an independent benchmark.

### Daemon host adapters

The generic `agent` feeder protocol is implemented inside Demon Docs. Thin MCP, Codex, Hermes, or other host adapters still need to register before a job and unregister on success, failure, cancellation, timeout, and spawn failure. These adapters are lifecycle plumbing only; they do not require the future code graph or context builder.

## Near-Term Hardening

The following work is independent of the larger code-graph track:

### Remaining validation and link-scan performance opportunities

- **Bounded parallel cold validation:** frontmatter and document-format validation currently enumerate and process applicable Markdown documents serially on a cache miss. Introduce a bounded document-worker pool for file reads, parsing, and per-document evaluation, then merge results deterministically before duplicate-document-ID checks, immutable-state decisions, diagnostics, repair planning, and publication. Benchmark conservative Windows worker limits rather than using unbounded goroutines.
- **Validation cache invalidation by unrelated rewrites:** validation reuse currently depends on the raw whole-document SHA-256. Link repairs, generated index changes, or other body-only rewrites therefore invalidate otherwise reusable frontmatter and format results. Split cache identity by validation-owned input surface, or safely refresh affected cache entries from final published bytes, so unrelated generated rewrites do not trigger cold validation.
- **Incremental changed-source link parsing:** unchanged source fingerprints already reuse stored link records, offsets, lines, and columns, but any source-content change currently reparses the complete Markdown document. Persist line or bounded-chunk hashes and enough synchronization metadata to diff changed regions, shift stored byte and line locations for unchanged regions, and reparse only affected regions plus context. Fall back to a full parse when edits may change non-local Markdown state, including frontmatter boundaries, fenced-code delimiters, reference definitions, HTML constructs, or parser-version changes.

- stress the single-owner lease path and retain race-focused coverage;
- complete actionable watcher and reconciliation diagnostics;
- expand heading-fragment validation when a deterministic Markdown anchor model is selected;
- verify Windows, Bash, PowerShell, and linked-worktree lifecycle behavior;
- expand reverse-index diagnostics and coverage reporting;
- stress review-ledger history, undo depth, and concurrent append behavior on larger repositories; and
- keep CLI help, README examples, and focused design documents synchronized with shipped behavior.

## Back-Burnered Major Track: Polyglot Code Graph

The planned code graph is larger than a single short implementation stream and is intentionally not the immediate critical path.

The important architectural decisions are:

- the existing Markdown/link graph remains the link-reconciliation model;
- the future code graph exists to add definitions, references, calls, imports, implementations, containment, and other bounded code relationships;
- the code graph must be polyglot at the adapter boundary from its first implementation step;
- Demon Docs should normalize facts from existing parsers, compiler tooling, SCIP-style indexes, or external code-intelligence providers rather than rebuilding every language analyzer; and
- graph-derived evidence may improve the codemap algorithm and later context selection, but inferred suggestions do not become authored graph truth.

The first implementation step, when this track resumes, is the language-neutral provider and normalized fact contract. A Go-only graph embedded directly into the core is not an acceptable architectural starting point.

See [Deterministic Typed Repository Graph](./code-intelligence/repository-graph.md), [Code-Symbol References](./code-intelligence/code-symbol-references.md), and [Code, Dependency, and Entanglement Facts](./code-intelligence/code-dependency-and-entanglement.md).

## Later Track: Context Bundles and Agent Integrations

Bounded deterministic context remains planned, but it follows a stable repository/code evidence contract. The same graph and explicit repository facts may support two separate consumers:

- codemap inference, which asks what permanent authored links may be missing; and
- context projection, which asks what existing information should be shown for a temporary task.

Those scoring paths must remain distinct. A useful context item is not automatically a valid permanent codemap link.

Later work includes:

- deterministic context-request and response contracts;
- bounded ordering and token or byte budgets;
- provenance and truncation reporting;
- CLI and MCP delivery;
- thin Codex, Hermes, Claude Code, and other host adapters; and
- paired historical-task benchmarking with leakage controls.

See [Deterministic Agent Context and Integrations](./agent-context-and-integrations.md) and [Context-Injection Benchmarking](../research/context-injection-benchmarking.md).

## Optional LLM Assistance

Optional LLM assistance may eventually propose documentation changes from deterministic diffs, codemap evidence, and graph facts. It remains outside correctness and cannot be required for indexing, link repair, codemap extraction, graph construction, validation, or context delivery.

## Principles

- **Deterministic first:** identical repository inputs and configuration produce stable facts, plans, diagnostics, and ordering.
- **Managed ownership is explicit:** generated indexes, reverse indexes, and adopted codemap sections own only their declared managed surfaces.
- **Missing-link generation remains one-directional:** it does not label an existing link irrelevant; confidence pruning is a separate opt-in execution policy disabled by default.
- **Remember declines:** unchanged declined additions remain suppressed; materially changed evidence may be reconsidered.
- **Polyglot seams before language implementations:** future code-intelligence providers normalize into one contract rather than becoming core-specific special cases.
- **Reuse existing analysis:** Demon Docs should not rebuild a general parser, compiler, call-graph platform, or graph database when an adapter can consume an existing deterministic source.
- **Static core remains authoritative:** watchers, daemons, MCP, and plugins automate or expose the same rebuildable core.
- **Thin integrations:** hosts translate lifecycle and request/response concerns without creating competing repository models.
- **No semantic prose generation in core:** deterministic behavior maintains structure, paths, references, evidence, and bounded projections.

## Explicit Non-Goals

- Replacing Git, Markdown, or the repository filesystem with a proprietary authoring model.
- Treating inferred semantic relationships as equivalent to authored references.
- Building another Sourcegraph, Codebase Memory, or general multi-language analysis platform inside Demon Docs.
- Requiring a daemon, network connection, LLM, language adapter, or external indexer for baseline reconciliation.
- Applying ambiguous, non-reviewable, or broad semantic documentation changes automatically.
- Claiming that one repository's curated codemaps provide universal algorithm quality.

## Code map

- `internal/reconcile/` — forward documentation index planning and application.
- `internal/links/` — repository-local link graph, identity state, diagnostics, rewrites, and stateless move planning.
- `internal/app/move.go` — explicit repository-bounded document refactoring CLI.
- `internal/demon/` — repository-local owner, feeder, heartbeat, shutdown, and log state.
- `internal/app/demon.go` — daemon CLI and shell integration.
- `internal/codemap/` — codemap extraction, deterministic datasets, unified managed sections, rendering, and schema placement seam.
- `internal/evidence/` — missing-link evidence collection.
- `internal/codemaprecommend/` — production admission, scoring, filtering, bounds, and tiers.
- `internal/codemaprun/` — production decline replay, pruning evaluation, plans, and transaction publication.
- `internal/codemapbench/` — holdout orchestration and reports using the production ranker.
- `internal/codemapcorpus/` — repository fact adapters used by codemap analysis.
- `internal/codemapprecision/` — curated precision evaluation.
- `internal/review/` — suggestion decisions, repair controls, applied-change history, and undo data.
- `internal/app/review_*.go` — suggestion, change, undo, and block CLI contracts.
- `research/codemap-precision/` — pinned labels, reports, and evaluation artifacts.

## Implementation sequence

Near-term work should prioritize bounded document-engine capabilities, reviewable suggestion decisions, diagnostics, and hardening that do not depend on the future polyglot graph. The provider seam is the first step when code-intelligence work resumes.

## Related docs

- [Planning](INDEX.md)
- [CLI Reference](../reference/cli.md)
- [Architecture](../architecture/INDEX.md)
- [Operations](../operations/INDEX.md)
- [Codemap Evidence](../research/codemap-evidence.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Current Product Limitations](../limits/current-limitations.md)
- [Planned Code Intelligence](code-intelligence/INDEX.md)
- [Planned Agent Context and Integrations](agent-context-and-integrations.md)

## Notes

Worktree-local feature status should be reconciled into this roadmap when branches merge; the roadmap should not speculate that unmerged behavior is already on `main`.
