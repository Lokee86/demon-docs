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

## Markdown anchors are not validated

Demon Docs preserves query strings and fragments while repairing paths, but it does not currently verify that a Markdown heading fragment exists or matches a renderer-specific anchor algorithm.

Impact:

- a file path may resolve while `#fragment` is stale;
- `check --links` does not prove heading-level validity; and
- different Markdown renderer slug rules remain outside the current contract.

Workaround:

Review fragment-bearing links manually or with a renderer-specific checker.

Owning docs:

- [Supported Link Syntax](../reference/supported-link-syntax.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)

Removal condition:

A deterministic supported anchor model, diagnostics, tests, and reference contract are implemented.

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

## Reverse indexes are file and folder level

Current reverse indexes project authored targets onto files and folders. They do not resolve authored declaration identities or produce symbol-level backlinks.

Impact:

- overloaded files cannot identify one declaration as the documented target;
- rename-aware symbol repair is unavailable;
- dependency and call relationships are not part of reverse coverage; and
- generated output should not be described as a code graph.

Workaround:

Target the narrowest current file or folder and use prose to name the declaration when needed.

Owning docs:

- [Reverse Index Architecture](../architecture/reverse-indexes.md)
- [Planned Code Intelligence](../planning/code-intelligence/INDEX.md)

Removal condition:

The language-neutral provider seam, symbol identity contract, authored syntax, resolution diagnostics, tests, and projections are implemented.

## Codemap generation quality is corpus-dependent

The deterministic evidence pipeline and explicit production writer are implemented, but recorded precision and recall measurements come from pinned labeled samples. They are not universal guarantees for arbitrary repositories, languages, naming styles, or documentation conventions.

Impact:

- explicit codemap execution automatically adds selected non-declined candidates from both tiers;
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

## Agent context delivery is not implemented

The repository demon exposes lifecycle feeders for agents, but it does not currently build or deliver deterministic task-context bundles.

Impact:

- an active `agent` feeder only keeps watcher automation alive;
- host adapters receive no context payload from the demon;
- there is no current CLI/MCP context request contract; and
- codemap suggestions must not be conflated with temporary task context.

Workaround:

Use the feeder protocol only for lifecycle integration and rely on existing host tooling for context until the planned static context core exists.

Owning docs:

- [Host Adapter Feeder Integration](../operations/host-adapters.md)
- [Deterministic Agent Context and Integrations](../planning/agent-context-and-integrations.md)

Removal condition:

A deterministic bounded context core, request/response schema, delivery adapters, provenance, truncation reporting, and evaluation are implemented.

## Machine-readable health output is not a stable public contract

Current `check`, orphan, and reconciliation diagnostics are human-readable text. JSON or SARIF output is not currently documented as a stable command contract.

Impact:

- integrations should primarily use process success/failure and preserve full text output;
- parsing individual message wording may be brittle; and
- CI annotations require adapter logic.

Workaround:

Treat zero versus non-zero as the stable automation boundary and archive stdout/stderr for diagnosis.

Owning docs:

- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [CI and Automation](../guides/ci-and-automation.md)

Removal condition:

A versioned machine-readable schema, command flag, tests, and compatibility policy are implemented.

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

## Watcher latency is dominated by broad reconciliation

The configured watcher debounce is only the quiet interval before work is admitted. It is not an end-to-end repair-latency target.

Current event handling records that relevant work exists but does not retain a complete path-scoped dirty set for every reconciliation subsystem. Repeated filesystem events reset the quiet interval, directory moves can emit large create, rename, and remove bursts, and bulk file renames use an additional quiet period of at least 500 milliseconds. Only the first recognized file rename in one batch receives the immediate targeted repair path; remaining changes converge through the normal selected reconciliation callback.

That callback may scan broad repository scope and run link reconciliation, frontmatter, document format, folder indexes, and final link-state refresh serially. Execution time is added after debounce and scheduler polling, so a sub-second configuration can still appear to take several seconds.

Impact:

- `debounce_seconds` should not be interpreted as maximum repair latency;
- lowering debounce further may increase churn without materially reducing broad-pass execution time;
- large directory moves or rapid editor-generated changes may schedule expensive follow-up passes;
- detached demon logs report completion after reconciliation, which can make processing time look like debounce time; and
- current watcher performance is suitable for a correctness-first hackathon prototype and modest repositories, not a production low-latency claim for large or high-churn trees.

Workaround:

Use explicit `ddocs mv` for planned moves, select only the required subsystem when running foreground watch, and use `ddocs fix` or `ddocs check` as the authoritative recovery and verification surfaces. Treat the watcher and repository demon as convenience automation.

Owning docs:

- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Watch Scheduler and Reconciliation Serialization](../architecture/watch-scheduler.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Roadmap](../planning/roadmap.md)

Removal condition:

Watch events produce path-aware dirty sets, each subsystem can reconcile only affected sources and targets, repeated state reads and writes are reduced or batched, large moves have bounded targeted handling, and retained benchmarks establish end-to-end latency expectations across representative repository sizes and event bursts.

## Cold validation retains serial coordination stages

Cold validation now uses a bounded 16-worker pool rather than processing every document serially. Frontmatter source reads and parsing run concurrently. Document-format source reads, frontmatter parsing, Markdown parsing, and schema enforcement also run concurrently.

Results remain indexed by deterministic file order and merge serially before operations that depend on repository-wide state. Duplicate-document-ID ownership and repair, immutable-value decisions, shared-schema and schema-history coordination, diagnostics ordering, rewrite planning, cache publication, and private-state publication remain serialized.

Impact:

- cold validation is substantially faster but does not scale linearly with CPU count;
- repositories dominated by shared-schema loading, private-state access, or publication may see smaller gains than parse-heavy corpora;
- broad policy or schema changes still invalidate many cache entries at once;
- link, index, or other body-only rewrites no longer invalidate frontmatter cache entries, but they still change document format's current whole-document identity and force a fresh format scan; and
- the worker limit is fixed at 16 rather than dynamically tuned per filesystem or machine.

Workaround:

Retain `.ddocs/` cache state, avoid deleting private state as routine cleanup, and use narrow `--frontmatter` or `--format` checks when diagnosing one policy system. The current bounded worker count is intended as a conservative cross-platform default.

Owning docs:

- [Validation Cache](../architecture/validation-cache.md)
- [Markdown Link Performance](../research/link-performance.md)
- [Roadmap](../planning/roadmap.md)

Removal condition:

Document format identity is narrowed to schema-selection metadata and heading structure, and affected cache records are safely refreshed after final generated rewrites, so body-only link changes do not cause format cold validation. Any further parallel coordination changes preserve deterministic duplicate-ID behavior, schema-history decisions, diagnostic ordering, and publication safety under retained benchmarks.

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
