# Demon Docs Roadmap

This roadmap describes the planned product evolution in implementation order. The focused filesystem and Markdown-link reconciliation core comes first; broader repository, code, projection, and agent-context graphs build on it later.

## Current Foundation: Implemented

The current foundation is the stable, repository-native reconciliation layer:

- Go is the sole implementation and supported runtime.
- Recursive folder indexes describe direct files, draft/stub files, and child folders.
- Managed Markdown sections are the only generated regions; authored content outside them is preserved.
- Parent navigation links keep folder indexes and configured indexed documents connected to their owning index.
- `check`, `fix`, and foreground `watch` provide reconciliation, verification, and explicit continuous local maintenance.
- The repository demon provides single-owner detached watcher lifecycle while shell or agent feeders remain active, without becoming a correctness dependency.
- `-d` / `--docs`, `-l` / `--links`, and `-r` / `--reverse` select reconciliation subsystems independently; `-i` / `--indexes` remains a compatibility alias for `--docs`.
- A focused repository-root Markdown link graph tracks local links to Markdown, assets, directories, absolute filesystem paths, and accessible external targets.
- Stable internal file IDs, path history, and fingerprints support deterministic move reconciliation without modifying source files to embed IDs.
- Existing index descriptions and link syntax are preserved where entries remain stable or moves are unambiguous.

This foundation establishes predictable filesystem synchronization without attempting to author semantic documentation or requiring the later repository/context graph.

## Phase 1: Markdown Links, Moves, and Static Reconciliation

Complete and harden the focused Markdown-link subsystem independently of the broader repository graph:

- validate local inline links, images, and reference definitions;
- track Markdown and non-Markdown targets;
- detect broken links, case mismatches, moves, and ambiguous candidates;
- preserve relative or absolute link style, labels, titles, queries, and fragments;
- use stable file IDs and fingerprints when exact path evidence disappears;
- keep first-scan baseline creation separate from later repair passes; and
- expose identical behavior through `check`, `fix`, and `watch`.

A unique deterministic candidate may be repaired automatically. Multiple plausible candidates remain unchanged and are reported for user resolution. The static CLI remains sufficient for rebuilding and reconciliation; `watch` provides continuous observation of the same core operations.

See [Markdown Link Reconciliation](markdown-links.md).

## Phase 2: Repository Inventory and Typed Graph

Build the broader repository-wide typed graph needed for advanced reverse indexing, code analysis, projections, and agent context. It may cover:

- folders and files beyond the focused link-state requirements;
- Markdown documents, headings, anchors, concepts, and aliases;
- indexes, ordinary link edges, and explicit code-path references; and
- provenance required by later projections and queries.

This graph is a deterministic representation of observed repository structure and explicit references, not an inferred model of what the repository ought to mean. It is deliberately sequenced after reliable static link updating rather than being required to implement it.

See the focused design document: [Deterministic Typed Repository Graph](repository-graph.md).

## Phase 3: Reverse Documentation Mapping and Code-Folder Indexes

Make documentation coverage navigable in both directions. Explicit codemap or code-path references produce code-to-documentation backlinks and local navigation.

Introduce code-folder indexes as a distinct reverse-index type, separate from ordinary forward documentation-folder indexes. These indexes are derived from documentation references into code folders, files, and (when available) symbols; they project which documentation covers each code target. They reconcile from those documentation references and their current targets, rather than being maintained as ordinary forward folder indexes or as an extension of the every-folder index model. Keep folder-level coverage distinct from file-level coverage: documenting a code folder does not imply that every file within it is documented, and a file reference does not automatically establish folder coverage.

See the focused design document: [Code-Folder Reverse Indexes](reverse-indexes.md).

## Phase 4: Deterministic Code-Symbol References

Add optional language adapters for deterministic code-symbol references. Adapters turn declarations into symbol nodes with source spans and connect explicit documentation references to those nodes.

Go is the first adapter. Each adapter must expose bounded, reproducible facts from the language parser or analysis tool; arbitrary semantic inference is out of scope. Repositories without an enabled adapter remain fully usable with file- and path-level references.

See the focused design document: [Code-Symbol References](code-symbol-references.md).

## Phase 5: Deterministic Code and Dependency Graph

Build a deterministic code and dependency graph from bounded parser, build, import, module, and package facts, together with explicit repository references. Track code folders, files, symbols, containment, and dependency edges where an enabled adapter can expose reproducible facts. The graph must remain inspectable and reproducible; semantic dependency inference is out of scope. This graph becomes the later source for dependency impact and entanglement projections.

See the focused design document: [Code, Dependency, and Entanglement Facts](code-dependency-and-entanglement.md).

## Phase 6: Graph Projections, Entanglement, and Queries

Expose useful projections and queries over the deterministic graph:

- backlinks and reverse references;
- orphaned documents and navigation gaps;
- impact reports for changed code or documentation;
- deterministic code/dependency graph views;
- bounded entanglement projections across documentation, code targets, and dependencies;
- freshness warnings based on explicit, inspectable signals;
- repository maps;
- machine-readable graph export; and
- bounded context bundles for agents.

These are generated views over repository state. They must remain reproducible from the same inputs and must not silently overwrite authored intent.

## Phase 7: Agent Context and Thin Integrations

Expose the same deterministic core and bounded projections through agent context bundles and thin adapters for the CLI, MCP, Hermes, Claude Code, and other clients. Adapters translate requests, responses, and host-specific concerns around the same core; they do not maintain a competing repository model or invent separate correctness rules.

Free-form concept resolution is deterministic only when a request matches explicit repository vocabulary, aliases, paths, symbols, headings, active files, or Git changes. Ambiguous prompts return candidates or wait for a concrete target rather than silently choosing one. Agent-facing context must identify its deterministic inputs and preserve the distinction between facts, projections, and authored intent.

See the focused design document: [Deterministic Agent Context and Integrations](agent-context-and-integrations.md).

Future evidence should compare paired historical tasks with and without Demon Docs context across repositories representing good/poor code and good/poor documentation. An intentionally constructed repository should validate the harness separately. This research is documented in [Context-Injection Benchmarking](context-injection-benchmarking.md) and is not a prerequisite for designing the deterministic core.

### Phase 7 research gate

Before claiming that agent context improves implementation work, develop the benchmark corpus and harness described in [Context-Injection Benchmarking](context-injection-benchmarking.md). This research is not required to begin deterministic context implementation and does not require immediate paid model trials. Corpus preparation, pinned historical tasks, control fixtures, and deterministic bundle inspection can proceed first.

## Phase 8: Repository Demon and Operational Expansion

The self-managing repository demon now wraps the same watcher used by foreground `ddocs watch`. It provides one fresh owner per local `.ddocs/` repository, shell and agent feeder heartbeats, detached startup, stale-owner recovery, shutdown grace, status, bounded logs, and independent linked-worktree runtime state.

Foreground `ddocs watch` remains available for explicit terminal-controlled operation. The static CLI remains authoritative for CI, rebuilds, recovery, and debugging, and deleting runtime state or disposable caches must never remove repository truth.

Further operational work may add MCP and native-plugin feeder adapters, packaging and installation improvements, broader cross-platform lifecycle tests, incremental scheduling, or graph-cache reuse. Those integrations remain outside Demon Docs core and must translate host lifecycle into the generic agent-feeder protocol rather than creating a competing repository model or daemon implementation.

## Phase 9: Optional LLM Assistance

Add optional LLM assistance for proposing documentation changes from code diffs. It must consume deterministic graph and change data rather than independently discovering repository truth.

LLM output remains a proposal: it must be reviewable, attributable to its inputs, and safe to reject or edit. LLM assistance is last and outside correctness: it must never be required for core inventory, reconciliation, validation, graph correctness, generated-index operation, concept resolution, or agent integrations.

## Principles

- **Deterministic first:** repository facts, references, reconciliation, and queries have stable inputs and reproducible outputs.
- **Repository-native output:** generated results live in ordinary Markdown, source-controlled files, and explicit machine-readable exports.
- **Authored intent remains the source of truth:** generated projections do not replace or silently reinterpret hand-authored prose.
- **Projections are generated:** indexes, backlinks, maps, reports, and bundles are views of the underlying repository model.
- **Explicit resolution only:** concept resolution is deterministic only against explicit repository vocabulary, aliases, paths, symbols, headings, active files, or Git changes; ambiguity yields candidates or waits for a concrete target.
- **One core, two watcher surfaces:** foreground `ddocs watch` and the self-managing repository demon run the same static core; neither adds exclusive product capability, and the static CLI remains authoritative.
- **Thin integrations:** CLI, MCP, Hermes, Claude Code, plugins, and other adapters use the same deterministic core rather than creating parallel repository models.
- **Optional adapters:** language-specific analysis is enabled deliberately and does not reduce baseline usability.
- **No semantic prose generation in core:** core Demon Docs behavior maintains structure and explicit references, not guessed explanations.

## Dependency and Order

The current foundation and Phase 1 establish trustworthy static Markdown-link updating, file identity, and continuous watch operation without waiting for the advanced repository graph. Phase 2 then expands those bounded filesystem facts into the typed repository model required by later reverse mappings, code-folder indexes, symbols, dependencies, projections, and agent context.

Phase 3 adds the distinct reverse documentation mappings and code-folder indexes. Phase 4 adds optional deterministic symbol nodes, and Phase 5 adds bounded code and dependency facts. Phase 6 consumes the broader deterministic graph for reproducible projections, including entanglement views and context bundles. Phase 7 exposes those capabilities through thin adapters without creating a competing core.

Phase 8 establishes and expands the self-managing lifecycle around the existing watcher, including generic shell and agent feeders, while keeping host adapters separate. Phase 9 comes last because proposals must be constrained by deterministic graph and diff data; it remains optional and outside the correctness path.

## Explicit Non-Goals

- Replacing Git, Markdown, or the repository filesystem with a proprietary storage model.
- Treating inferred semantic relationships as equivalent to explicit authored references.
- Generating or rewriting semantic prose as part of deterministic core operations.
- Requiring the repository demon for one-shot CLI recovery, CI, rebuilds, or correctness verification.
- Making the repository demon or foreground watcher the source of repository truth, or giving either capabilities unavailable through the static core.
- Requiring MCP or plugins to be hosted by the daemon rather than exposing them as separate interfaces.
- Requiring an LLM, a network connection, or a language adapter for baseline reconciliation.
- Deterministically resolving an unconstrained free-form concept without an explicit vocabulary, alias, path, symbol, heading, active-file, or Git-change match.
- Claiming folder-level coverage when only individual files are referenced, or vice versa.
- Applying broad, ambiguous, or non-reviewable link and documentation changes.
