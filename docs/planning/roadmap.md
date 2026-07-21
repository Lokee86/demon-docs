---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7366-a679-359c693253d2
document_type: general
policy_exempt: false
summary: This document summarizes shipped Demon Docs product areas, active work, and near-term documentation-engine hardening without claiming ownership of adjacent Warlock tools.
---
# Demon Docs Roadmap

Parent index: [Planning](./INDEX.md)

## Purpose

This document summarizes shipped Demon Docs product areas, active work, and near-term documentation-engine hardening without acting as the canonical reference for implemented behavior.

## Overview

The roadmap is a sequencing and status document. Current product summaries link to canonical architecture, operations, reference, and research pages; detailed shipped behavior belongs in those documents.

## Current status

Active roadmap. Version 0.3.5 includes stateless refactoring, orphan health checks, the review ledger, strict frontmatter policy, document-format schemas, reverse-index health, explicit production codemap execution with schema-governed missing-section creation, independent validation cache identities, generated-rewrite cache refresh, and path-scoped watcher validation for ordinary Markdown edits. Polyglot repository intelligence and task-context delivery are no longer Demon Docs roadmap items; they are owned by sibling Warlock tools.

## Ownership boundary

This roadmap owns Demon Docs sequencing and status summaries. It does not own exact CLI contracts, current implementation mechanics, benchmark methodology, operational recovery procedures, or the roadmaps of adjacent Warlock components.

Demon Docs owns deterministic documentation maintenance: Markdown links, indexes, schemas, frontmatter and document-format policy, authored codemaps, file/folder reverse indexes, review history, and watcher automation. [ArcanaGraph](https://github.com/Lokee86/arcana-graph) owns the language-independent repository relationship graph. **Grimoire Context** owns repository context discovery, selection, packaging, and delivery. The [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain) owns shared terminology, integration direction, and cross-tool contracts.

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
- Existing index and parent-editable document sources are loaded once through bounded workers; independent folder plans merge serially in deterministic tree order before writes.

### Document health checks

- Link-enabled `check` reports normal managed Markdown documents without a meaningful inbound link.
- Folder indexes, draft documents, self-links, and inbound links originating from indexes or drafts do not mask an orphan.
- Results are deterministic and path-sorted.
- Health checks are diagnostic only; Demon Docs does not guess where an orphan should be linked or remove it.

See [Document Health Checks](../guides/document-health-checks.md).

### Validation performance

- Cold frontmatter source reads and parsing use a bounded 16-worker pool.
- Cold document-format source reads, frontmatter parsing, Markdown parsing, and schema enforcement use the same bounded pool.
- Results remain indexed by deterministic file order and merge serially before duplicate-document-ID handling, immutable-state decisions, diagnostics, repair planning, cache publication, and schema-history publication.
- The durable clean-validation cache remains the fastest repeated path; parallel workers reduce the cost when that cache is absent or invalidated.
- Frontmatter and document-format reuse have independent identities, so unrelated prose, link, code-block, and section-body edits no longer invalidate both systems.
- Known generated rewrites refresh the final raw source hash and retain or invalidate only the validation surfaces they can affect.
- Ordinary Markdown create and write events carry changed paths into scoped watcher validation; untouched clean documents reuse cache state without being read or parsed, with conservative full-pass fallback when cache or event evidence is incomplete.

See [Validation Cache](../architecture/validation-cache.md) and [Markdown Link Performance](../research/link-performance.md).

### Repository-local link reconciliation and refactoring

- Repository Markdown is scanned subject to `.docignore` and permanent traversal exclusions.
- Supported local forms include inline links, images, reference definitions, explicit and collapsed reference uses, path-based wiki links, wiki embeds, and common local HTML `href`, `src`, and `poster` targets.
- Stable internal file identities, path history, fingerprints, and incoming-link groups support deterministic move reconciliation without embedding IDs in source files.
- Link labels, titles, aliases, angle wrapping, query strings, fragments, source newline style, and surrounding prose are preserved.
- Undefined explicit or collapsed reference labels are reported.
- Changed Markdown sources are read and parsed through bounded concurrency before deterministic serial target resolution and repair planning.
- Known-move rewrite planning and generated rewrite application use bounded concurrency while retaining deterministic source ordering, source-hash checks, and atomic per-file replacement.
- `ddocs mv` explicitly moves a file or directory and rewrites affected incoming and moved-source links without requiring or creating `.ddocs/` state.
- Move planning supports dry runs, repository boundaries, case-only renames, affected ambiguity refusal, source-hash preflight, and best-effort rollback.

See [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md) and [Stateless Document Refactoring](../guides/document-refactoring.md).

### Reverse code-folder indexes

- Reverse indexes project authored codemap references back onto configured code folders and files.
- Recursive repository-relative roots, repeated `--reverse-root` overrides, nested `.docignore`, and configurable codemap headings are implemented.
- `check`, `fix`, and `watch` support `-r` / `--reverse` independently or alongside documentation indexes and links.
- Missing codemap sections, empty matching sections, unresolved targets, and coverage gaps remain explicit diagnostics.
- Folder inventory and per-folder reverse-index reconciliation use bounded preparation workers, deterministic path-indexed results, and serial write application.
- Richer file/folder coverage reports and move-aware authored-reference repair remain possible Demon Docs hardening. General symbol identity and symbol-level repository projections belong to ArcanaGraph.

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
- Authored codemap datasets use serial deterministic discovery followed by bounded parallel document reading, hashing, extraction, and target resolution; repeated references share one per-build target-content hash.
- Repository corpus construction runs document loading, shared source analysis, and bounded Git-history collection independently. Dependency and symbol adapters share one bounded source read, then facts merge serially and deterministically.
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

The generic `agent` feeder protocol is implemented inside Demon Docs. Thin MCP, Codex, Hermes, or other host adapters may register before a job and unregister on success, failure, cancellation, timeout, and spawn failure. These adapters are lifecycle plumbing only; they do not assemble or deliver task context.

## Near-Term Hardening

The following work remains inside the Demon Docs documentation-maintenance boundary:

### Remaining validation and link-scan performance opportunities

- **Path-scoped link and index reconciliation:** ordinary Markdown edits now scope frontmatter and document-format validation, but links and folder indexes still reconcile across broader repository or documentation scope. Introduce changed-source and changed-target batches for those subsystems, retain full reconciliation as an overflow and uncertainty fallback, and benchmark end-to-end move latency separately from configured debounce.
- **Shared command snapshots and metadata-assisted cache hits:** selected validators still perform overlapping file reads and parsing, and format cache candidates still read frontmatter selection metadata plus the complete heading structure. Read each active Markdown source once per command where practical, share immutable derived data across planners, and use filesystem metadata only as a safe fast-rejection layer before authoritative hashes.
- **Incremental changed-source link parsing:** unchanged source fingerprints already reuse stored link records, offsets, lines, and columns, but any source-content change currently reparses the complete Markdown document. Persist line or bounded-chunk hashes and enough synchronization metadata to diff changed regions, shift stored byte and line locations for unchanged regions, and reparse only affected regions plus context. Fall back to a full parse when edits may change non-local Markdown state, including frontmatter boundaries, fenced-code delimiters, reference definitions, HTML constructs, or parser-version changes.

- stress the single-owner lease path and retain race-focused coverage;
- complete actionable watcher and reconciliation diagnostics;
- expand heading-fragment validation when a deterministic Markdown anchor model is selected;
- verify Windows, Bash, PowerShell, and linked-worktree lifecycle behavior;
- expand reverse-index diagnostics and coverage reporting;
- stress review-ledger history, undo depth, and concurrent append behavior on larger repositories; and
- keep CLI help, README examples, and focused design documents synchronized with shipped behavior.

## Warlock Toolchain Boundaries

Several larger ideas originally explored in this repository now have separate product owners:

- [ArcanaGraph](https://github.com/Lokee86/arcana-graph) owns normalized repository relationships, polyglot code facts, symbol identities and references, dependency and impact projections, and graph queries.
- **Grimoire Context** owns bounded task-context discovery, selection, packaging, provenance, truncation, and delivery to agents or other hosts.
- The [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain) owns shared integration contracts and cross-tool direction.

Demon Docs may export deterministic documentation facts or consume explicitly versioned facts from sibling tools, but it does not absorb their implementation responsibilities. The retained [code-intelligence](./code-intelligence/INDEX.md), [agent-context](./agent-context-and-integrations.md), and [context-benchmark](../research/context-injection-benchmarking.md) documents are historical design inputs transferred to those sibling tools, not active Demon Docs roadmap commitments.

## Optional LLM Assistance

Optional LLM assistance may eventually propose documentation changes from deterministic diffs and codemap evidence. It remains outside correctness and cannot be required for indexing, link repair, codemap extraction, validation, or any other Demon Docs operation.

## Principles

- **Deterministic first:** identical repository inputs and configuration produce stable facts, plans, diagnostics, and ordering.
- **Managed ownership is explicit:** generated indexes, reverse indexes, and adopted codemap sections own only their declared managed surfaces.
- **Missing-link generation remains one-directional:** it does not label an existing link irrelevant; confidence pruning is a separate opt-in execution policy disabled by default.
- **Remember declines:** unchanged declined additions remain suppressed; materially changed evidence may be reconsidered.
- **Tool boundaries remain explicit:** documentation maintenance stays in Demon Docs; repository graph intelligence and context delivery stay in their sibling tools.
- **Exchange facts through contracts:** integrations should use versioned, attributable data rather than duplicating another tool's internal model.
- **Static core remains authoritative:** watchers, daemons, and host adapters automate or expose the same rebuildable Demon Docs core.
- **Thin lifecycle integrations:** host adapters may maintain feeder state without becoming context engines or competing repository models.
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

Near-term work should prioritize bounded document-engine capabilities, reviewable suggestion decisions, diagnostics, performance, and operational hardening. Repository-graph and context-delivery implementation belongs in ArcanaGraph and Grimoire Context rather than this sequence.

## Related docs

- [Planning](INDEX.md)
- [CLI Reference](../reference/cli.md)
- [Architecture](../architecture/INDEX.md)
- [Operations](../operations/INDEX.md)
- [Codemap Evidence](../research/codemap-evidence.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Current Product Limitations](../limits/current-limitations.md)
- [Transferred Code-Intelligence Design](code-intelligence/INDEX.md)
- [Transferred Agent-Context Design](agent-context-and-integrations.md)

## Notes

Worktree-local feature status should be reconciled into this roadmap when branches merge; the roadmap should not speculate that unmerged behavior is already on `main`.
