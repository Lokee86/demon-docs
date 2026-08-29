---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7b52-aa6e-731e59030442
document_type: general
policy_exempt: false
summary: This document records current Demon Docs limitations that materially affect adoption, diagnostics, or feature expectations.
---
# Current Product Limitations

Parent index: [Limits](./INDEX.md)

## Purpose

This document records current Demon Docs limitations that materially affect adoption, diagnostics, or feature expectations.

## Overview

These entries describe incomplete or deliberately narrow current surfaces. They are not permission to weaken deterministic safety rules. Permanent boundaries such as refusing ambiguous rewrites remain architecture invariants even when future interfaces improve how users resolve them.

## Initial link state has no historical move evidence

The first link-enabled mutating pass records the current repository baseline. It cannot infer where a currently broken target lived before Demon Docs began tracking identity.

Impact:

- pre-existing broken moves generally require manual repair;
- first-pass state must not be treated as historical evidence; and
- deleting `.ddocs/` resets this capability.

Workaround:

Repair current broken links, establish a clean baseline, and retain `.ddocs/` history for later moves.

Owning docs:

- [Getting Started](../guides/getting-started.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)

Removal condition:

A separate, explicit historical import mechanism is implemented. Normal baseline creation should remain non-speculative.

## Orphan health is reachability-only

The orphan check verifies that a normal managed Markdown document has at least one meaningful inbound link under its defined exclusions. It does not assess whether that link is semantically appropriate or whether the document is complete.

Impact:

- a weak but valid inbound link satisfies graph reachability;
- index and draft links intentionally do not satisfy it;
- there is no per-document semantic exemption; and
- the check does not recommend the correct owning document.

Workaround:

Use authored review to decide whether to add a meaningful relationship, move incomplete material to drafts, merge it, or remove it.

Owning docs:

- [Document Health Checks](../guides/document-health-checks.md)

Removal condition:

The reachability contract may gain explicit reviewed exemptions or richer diagnostics without claiming automated semantic judgment.

## Link checking is local, not network reachability validation

Demon Docs recognizes repository-local and supported filesystem targets. It does not fetch HTTP, HTTPS, mail, or other external destinations to test availability.

Impact:

- external URLs may be stale while `ddocs check --links` succeeds;
- network status, redirects, authentication, and rate limits are outside reconciliation; and
- external content is never edited.

Workaround:

Use a dedicated external link checker when network reachability is required.

Owning docs:

- [Supported Link Syntax](../reference/supported-link-syntax.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)

Removal condition:

A separately scoped, opt-in network checker is implemented without entering deterministic path-repair ownership.

## Reverse indexes project authored targets, not graph relationships

Reverse indexes now project authored file and folder targets plus uniquely verified exact symbol targets. Arcana owns declaration identity, parsing, and repository graph relationships; Demon Docs consumes only the verified semantic node attached to an authored codemap target.

Impact:

- exact symbol backlinks require current matching Arcana/Lexicon state and a uniquely resolved authored target;
- a path-qualified symbol falls back to its explicit backing file when semantic verification is unavailable, while a standalone `symbol:...` target cannot be projected without semantic resolution;
- rename/move awareness is surfaced through codemap semantic staleness rather than speculative reverse-index rewriting;
- dependency, call, implementation, and arbitrary graph-neighbour relationships are not part of reverse coverage; and
- generated reverse indexes must not be described as a repository code graph.

Workaround:

Use explicit file or folder targets when declaration-level resolution is unavailable. For declaration-level backlinks, prepare current Lexicon/Arcana state and use an exact symbol target that resolves uniquely.

Owning docs:

- [Reverse Index Architecture](../architecture/reverse-indexes.md)
- [Codemap Extraction and Dataset](../architecture/codemap-extraction-and-dataset.md)
- [Transferred Code-Intelligence Design](../planning/code-intelligence/INDEX.md)

Boundary:

[ArcanaGraph](https://github.com/Lokee86/arcana-graph) continues to own language-independent repository relationships and symbol identity. Demon Docs owns only the authored-document projection: raw Arcana relationships never create reverse backlinks.

## Codemap generation quality is corpus-dependent

The deterministic evidence pipeline and explicit production writer are implemented, but recorded precision and recall measurements come from pinned labeled samples. They are not universal guarantees for arbitrary repositories, languages, naming styles, or documentation conventions.

Impact:

- explicit codemap execution automatically adds only selected non-declined `hard_link` candidates; `context` remains non-mutating;
- Arcana-backed file/symbol target resolution, bounded one-hop semantic relationship evidence, and mapped-node semantic staleness are wired when current matching Arcana/Lexicon state is available; semantic staleness also requires a previously accepted baseline and retained prior Arcana snapshot, while repository-wide dependency/symbol facts still default to the built-in shallow local provider and the new role/directory coverage policy has not yet been rebenchmarked across the frozen corpora;
- a repository may receive plausible but unnecessary `context` links;
- new repository populations need independent labels;
- self-authored Demon Docs codemaps are not an independent benchmark; and
- low-quality or sparse code maps reduce useful supervision.

Workaround:

Start with one representative file, use `codemap inspect` and `fix --dry-run`, retain conservative no-pruning defaults, record declines for unwanted additions, and evaluate new corpora before changing thresholds.

Owning docs:

- [Managing Codemaps](../guides/managing-codemaps.md)
- [Codemap Missing-Link Evidence](../codemap-evidence.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)

Removal condition:

This limitation cannot be fully removed; it can be narrowed by broader validated corpora, calibrated tiers, repository-specific evaluation, and improved evidence providers.

## Agent context delivery is outside Demon Docs

The repository demon exposes lifecycle feeders for agents, but it does not build or deliver task-context bundles. Context discovery, selection, packaging, and delivery are owned by the planned **Grimoire Context** sibling tool.

Impact:

- an active `agent` feeder only keeps Demon Docs watcher automation alive;
- host adapters receive no context payload from the demon;
- Demon Docs has no context request or delivery command contract; and
- codemap `hard_link` suggestions remain permanent documentation-link candidates; `context` suggestions are non-mutating analysis/review output, not temporary task-context delivery.

Operational guidance:

Use the feeder protocol only for lifecycle integration. Context-producing hosts or sibling tools should use an explicit integration contract rather than extending the Demon Docs daemon into a context service.

Owning docs:

- [Host Adapter Feeder Integration](../operations/host-adapters.md)
- [Transferred Agent-Context Design](../planning/agent-context-and-integrations.md)
- [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain)

Boundary:

This is not a missing Demon Docs feature. Grimoire Context owns the context product boundary, while Warlock owns cross-tool integration direction.

## Machine-readable diagnostics cover reconciliation, not precondition failures

`ddocs check --output-format json` now exposes the stable schema-1 native diagnostic contract for every reconciliation subsystem: links, documentation indexes, frontmatter, document-body format, and reverse indexes. Combined reports include the selected subsystem findings plus link and reverse-index health findings where applicable. Failures that prevent reconciliation planning from completing still use the normal CLI/runtime error surface.

Impact:

- CI and agent integrations can consume stable diagnostic codes and structured evidence for every completed reconciliation check;
- reverse-index target, generated-index, and orphan-code health is included in the same envelope; and
- usage, configuration, filesystem, or runtime failures that prevent a completed plan still require the normal error surface.

Workaround:

Use the JSON contract for completed reconciliation checks. Treat exit code `2` and stderr as a command/precondition failure rather than attempting to parse it as a schema-1 reconciliation report.

Owning docs:

- [Machine-Readable Diagnostics](../reference/machine-readable-diagnostics.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [CI and Automation](../guides/ci-and-automation.md)

Removal condition:

A future contract explicitly defines machine-readable command/precondition failures where doing so is useful without conflating them with completed reconciliation findings.

## Symlink entries are not owned traversal trees

Demon Docs does not traverse symbolic-link entries as repository-owned documentation or code trees and rejects symbolic-link move sources.

Impact:

- content reachable only through a symlink is outside normal indexing and repair scope;
- `ddocs mv` cannot move a symlink source; and
- repositories that use symlinked docs must manage those paths separately.

Workaround:

Use real repository-contained paths or configure the owning repository directly.

Owning docs:

- [Managed Files and State](../reference/managed-files-and-state.md)
- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md)

Removal condition:

A complete cross-platform symlink ownership and containment policy is implemented. Silent traversal should remain prohibited.

## Scoped watcher reconciliation still retains broad evidence costs

The watcher now carries ordinary file create, write, remove, and rename paths into fail-closed scoped link and folder-index reconciliation. Affected Markdown link sources and affected index folders are planned selectively; frontmatter and document-format validation already use scoped cache-backed paths. Directory, external-target, control/schema, overflow, startup, incomplete-batch, and uncertain events still use the full path.

Scoped execution intentionally retains some repository-wide evidence work. Link reconciliation still loads persisted state, builds the authoritative inventory, and publishes the complete private projection. Folder-index reconciliation still scans the complete documentation tree and loads shared cross-folder evidence before preparing only affected folders. These costs preserve identity, ambiguity, cross-folder description, and missed-event detection without weakening correctness.

Impact:

- `debounce_seconds` remains a quiet-period setting, not a maximum repair-latency guarantee;
- ordinary file events avoid broad source/folder preparation but still pay state, inventory/tree, scheduler-polling, and publication costs;
- large directory operations and event-buffer overflow intentionally fall back to full reconciliation;
- bulk rename bursts retain the additional quiet-period policy after the first immediate observed-rename repair; and
- further latency reduction now requires narrowing evidence/state I/O rather than merely adding another path filter.

Workaround:

Use explicit `ddocs mv` for planned large directory refactors and `ddocs fix` or `ddocs check` as authoritative recovery surfaces. The watcher automatically falls back to full reconciliation when scoped evidence is incomplete.

Owning docs:

- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Markdown Link Performance](../research/link-performance.md)

Removal condition:

Measured need justifies incremental private-state loading/publication or narrower authoritative inventory/tree snapshots without weakening identity recovery, ambiguity refusal, missed-event detection, or cross-folder correctness.

## Cold validation retains serial coordination stages

Cold validation now uses a bounded 16-worker pool rather than processing every document serially. Frontmatter source reads and parsing run concurrently. Document-format source reads, frontmatter parsing, Markdown parsing, and schema enforcement also run concurrently.

Results remain indexed by deterministic file order and merge serially before operations that depend on repository-wide state. Duplicate-document-ID ownership and repair, immutable-value decisions, shared-schema and schema-history coordination, diagnostics ordering, rewrite planning, cache publication, and private-state publication remain serialized.

Impact:

- cold validation is substantially faster but does not scale linearly with CPU count;
- repositories dominated by shared-schema loading, private-state access, or publication may see smaller gains than parse-heavy corpora;
- broad policy or schema changes still invalidate many cache entries at once;
- unrelated prose, link, code-block, and section-body rewrites no longer invalidate frontmatter or format enforcement results, but each format cache candidate still requires a file read, frontmatter selection parse, and complete heading scan; and
- the worker limit is fixed at 16 rather than dynamically tuned per filesystem or machine.

Workaround:

Retain `.ddocs/` cache state, avoid deleting private state as routine cleanup, and use narrow `--frontmatter` or `--format` checks when diagnosing one policy system. The current bounded worker count is intended as a conservative cross-platform default.

Owning docs:

- [Validation Cache](../architecture/validation-cache.md)
- [Markdown Link Performance](../research/link-performance.md)
- [Roadmap](../planning/roadmap.md)

Removal condition:

Command-scoped source snapshots remove duplicate reads and parsing across selected validators, and metadata-assisted hits avoid unnecessary file reads safely. Any further parallel coordination changes preserve deterministic duplicate-ID behavior, schema-history decisions, diagnostic ordering, and publication safety under retained benchmarks.

## Changed Markdown sources are reparsed as whole documents

Unchanged source fingerprints reuse stored link records, offsets, lines, and columns. Changed sources are read and parsed concurrently through a bounded 16-worker pool, then merged in source order before serial target resolution and repair planning. Any content change still causes the complete individual Markdown source to be parsed again.

Impact:

- inserting one line into a large document reparses every link occurrence in that document;
- stored offsets are not shifted through a line- or chunk-diff model; and
- large frequently edited Markdown files can dominate incremental link-refresh cost.

Workaround:

No correctness workaround is required. Keep generated bulk content out of managed Markdown sources when practical and rely on unchanged-source reuse for files that did not change.

Owning docs:

- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Link Reconciliation State Machine](../architecture/link-reconciliation-state-machine.md)
- [Markdown Link Performance](../research/link-performance.md)

Removal condition:

Line or bounded-chunk hashes, safe offset shifting, changed-region parsing, parser-state synchronization, and conservative full-parse fallbacks are implemented and benchmarked.

## Related docs

- [Limits](INDEX.md)
- [Roadmap](../planning/roadmap.md)
- [Documentation Policy](../documentation-policy.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

This page should contain current user-visible limitations, not a general feature wishlist. Planned designs belong under `planning/`, and permanent safety rules belong in architecture.
