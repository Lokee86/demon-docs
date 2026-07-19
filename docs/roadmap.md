# Demon Docs Roadmap

This roadmap describes the current product state and the next implementation tracks. It separates shipped behavior, active tuning work, bounded near-term work, and larger back-burnered architecture so planned work is not mistaken for released functionality.

## Current Product: Implemented on `main`

### Documentation-tree reconciliation

- Go is the sole implementation and supported runtime.
- Recursive folder indexes describe direct files, draft/stub files, and child folders.
- Managed Markdown sections are the only generated regions; authored content outside them is preserved.
- Parent navigation links keep folder indexes and configured indexed documents connected to their owning index.
- `check`, `fix`, and foreground `watch` expose the same deterministic reconciliation core.
- The repository demon provides single-owner detached watcher lifecycle while shell or agent feeders remain active, without becoming a correctness dependency.
- `-d` / `--docs`, `-l` / `--links`, and `-r` / `--reverse` select reconciliation subsystems independently; `-i` / `--indexes` remains a compatibility alias for `--docs`.
- Existing index descriptions and link syntax are preserved where entries remain stable or moves are unambiguous.

### Repository-local link reconciliation

- Repository Markdown is scanned subject to `.docignore` and permanent traversal exclusions.
- Supported local forms include inline links, images, reference definitions, explicit and collapsed reference uses, path-based wiki links, wiki embeds, and common local HTML `href`, `src`, and `poster` targets.
- Stable internal file identities, path history, fingerprints, and incoming-link groups support deterministic move reconciliation without embedding IDs in source files.
- Link labels, titles, aliases, angle wrapping, query strings, fragments, source newline style, and surrounding prose are preserved.
- Undefined explicit or collapsed reference labels are reported.
- Generated rewrites use bounded concurrency while retaining deterministic planning, source-hash checks, and atomic per-file replacement.

See [Markdown Link Reconciliation](markdown-links.md).

### Reverse code-folder indexes

- Reverse indexes project authored codemap references back onto configured code folders and files.
- Recursive repository-relative roots, repeated `--reverse-root` overrides, nested `.docignore`, and configurable codemap headings are implemented.
- `check`, `fix`, and `watch` support `-r` / `--reverse` independently or alongside documentation indexes and links.
- Missing codemap sections, empty matching sections, unresolved targets, and coverage gaps remain explicit diagnostics.
- Symbol-level projection, move-aware authored-reference repair, and richer coverage reports remain later work.

See [Code-Folder Reverse Indexes](reverse-indexes.md).

### Repository demon

- One fresh owner serves each initialized repository-local `.ddocs/` state directory.
- Shell and generic agent feeders keep the watcher active while work is in progress.
- Detached startup, stale-owner recovery, shutdown grace, status, linked-worktree bootstrap, and bounded logs are implemented.
- Bash and PowerShell hooks translate shell entry and exit into feeder registration.
- The daemon remains optional; `check`, `fix`, and foreground `watch` remain authoritative recovery and CI surfaces.

See [Repository Demon](repository-demon.md).

### Codemap extraction and deterministic missing-link research

- Authored codemap sections can be exported as a deterministic JSON dataset.
- The repository corpus adapter collects paths, dependency neighbours, declared symbols, source/test relationships, related-document targets, and bounded Git co-change evidence.
- Holdout benchmarking measures whether known authored targets are recovered.
- Precision tooling generates, samples, and evaluates ranked candidate links.
- Suggestions are divided into `hard_link` and `context` tiers.
- Existing links are never returned as missing-link candidates, and there is no removal or irrelevance signal.

The current curated Space Rocks sample contains 150 labeled suggestions. The recorded baseline has 60% precision at rank 1, 64.2% valid-link precision for the `hard_link` tier, 95.1% non-junk acceptance for that tier, and 74.3% sample recall of valid links. These numbers describe the pinned labeled sample, not a universal quality claim.

See [Codemap Missing-Link Evidence](codemap-evidence.md).

## Active Work

### Codemap tuning and broader corpus validation

The current tuning work is isolated from `main`. Near-term goals are:

- compare each scoring change against the pinned precision sample;
- preserve deterministic output and evidence fingerprints;
- expand evaluation beyond one repository without treating unlabeled output as ground truth;
- use Demon Docs' own refreshed code maps as a development corpus, not as an independent benchmark; and
- merge only changes that demonstrate a measured improvement or a clearly safer candidate surface.

### User-facing suggestion decisions

The evidence and ranking machinery exists, but the complete review lifecycle remains unfinished. Planned bounded work includes:

- list strong missing-link candidates with their evidence;
- persist accepted and declined decisions by document, target, and evidence fingerprint;
- suppress an unchanged declined suggestion;
- reconsider it only when the underlying evidence materially changes; and
- keep accepted changes reviewable rather than silently rewriting authored codemaps.

### Daemon host adapters

The generic `agent` feeder protocol is implemented inside Demon Docs. Thin MCP, Codex, Hermes, or other host adapters still need to register before a job and unregister on success, failure, cancellation, timeout, and spawn failure. These adapters are lifecycle plumbing only; they do not require the future code graph or context builder.

## Near-Term Hardening

The following work is independent of the larger code-graph track:

- stress the single-owner lease path and retain race-focused coverage;
- complete actionable watcher and reconciliation diagnostics;
- expand heading-fragment validation when a deterministic Markdown anchor model is selected;
- verify Windows, Bash, PowerShell, and linked-worktree lifecycle behavior;
- expand reverse-index diagnostics and coverage reporting;
- expose the codemap accept/decline workflow; and
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

See [Deterministic Typed Repository Graph](repository-graph.md), [Code-Symbol References](code-symbol-references.md), and [Code, Dependency, and Entanglement Facts](code-dependency-and-entanglement.md).

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

See [Deterministic Agent Context and Integrations](agent-context-and-integrations.md) and [Context-Injection Benchmarking](context-injection-benchmarking.md).

## Optional LLM Assistance

Optional LLM assistance may eventually propose documentation changes from deterministic diffs, codemap evidence, and graph facts. It remains outside correctness and cannot be required for indexing, link repair, codemap extraction, graph construction, validation, or context delivery.

## Principles

- **Deterministic first:** identical repository inputs and configuration produce stable facts, plans, diagnostics, and ordering.
- **Authored intent remains authoritative:** generated indexes, reverse indexes, candidates, and projections do not silently replace hand-authored meaning.
- **Only suggest missing relationships:** the codemap system never recommends that an existing link is irrelevant or should be removed.
- **Remember declines:** unchanged declined suggestions remain suppressed; materially changed evidence may be reconsidered.
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
- `internal/links/` — repository-local link graph, identity state, diagnostics, and rewrites.
- `internal/demon/` — repository-local owner, feeder, heartbeat, shutdown, and log state.
- `internal/app/demon.go` — daemon CLI and shell integration.
- `internal/codemap/` — authored codemap extraction and deterministic datasets.
- `internal/evidence/` — missing-link evidence collection.
- `internal/codemapbench/` — holdout orchestration, ranking, tiers, and reports.
- `internal/codemapcorpus/` — repository fact adapters used by codemap analysis.
- `internal/codemapprecision/` — curated precision evaluation.
- `research/codemap-precision/` — pinned labels, reports, and evaluation artifacts.
