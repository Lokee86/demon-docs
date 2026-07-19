# Demon Docs Roadmap

This roadmap describes the planned product evolution in implementation order. Each phase builds on the deterministic repository model before adding broader projections, long-running operation, or optional assistance.

## Current Foundation: Implemented

The current foundation is the stable, repository-native reconciliation layer:

- Go is the sole implementation and supported runtime.
- Recursive folder indexes describe direct files, draft/stub files, and child folders.
- Managed Markdown sections are the only generated regions; authored content outside them is preserved.
- Parent navigation links keep folder indexes and configured indexed documents connected to their owning index.
- `check`, `fix`, and `watch` provide reconciliation, verification, and local continuous maintenance.
- Existing descriptions are preserved where entries remain stable or moves are unambiguous.

This foundation establishes predictable filesystem synchronization without attempting to author semantic documentation.

## Phase 1: Repository Inventory and Typed Graph

Build a repository-wide inventory and typed graph covering:

- folders and files;
- Markdown documents, headings, and anchors;
- links and indexes; and
- explicit code-path references.

The inventory must respect Git tracked-file boundaries and `.gitignore` semantics where configured. The graph is a deterministic representation of observed repository structure and explicit references, not an inferred model of what the repository ought to mean.

See the focused design document: [Deterministic Typed Repository Graph](repository-graph.md).

## Phase 2: Links, Moves, and Incremental Reconciliation

Add deterministic Markdown link validation, including detection of broken links, ambiguous targets, and case mismatches. Add Git-aware move and rename detection so repository history and current tracked state can distinguish relocation from deletion where the evidence is available.

Use that inventory and graph to repair body links deterministically when the target is known. A full rebuild remains available as a recovery and verification path, while normal updates become incremental and touch only affected state and files.

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

## Phase 8: Operational Daemon

The daemon remains a required/planned product feature, but that product commitment is separate from runtime correctness: users must not need it for static CLI operations, CI, rebuilds, recovery, or one-shot queries.

The daemon automates and schedules static core capabilities: continuous Git/filesystem change handling, event coalescing, incremental reconciliation, and automatic maintenance. It may keep disposable caches of graph or projection state, but deleting those caches must not remove repository truth or require a separate recovery model. The daemon introduces no exclusive product capability; the static CLI remains the authoritative recovery, CI, rebuild, and debugging path and can validate or rebuild directly from repository inputs.

This phase is intentionally lower priority than the deterministic static core and its thin integrations. MCP and plugins are separate interfaces and need not be hosted by the daemon; they may use the core directly or connect to an available service. Building the daemon first would duplicate or obscure the underlying correctness model.

## Phase 9: Optional LLM Assistance

Add optional LLM assistance for proposing documentation changes from code diffs. It must consume deterministic graph and change data rather than independently discovering repository truth.

LLM output remains a proposal: it must be reviewable, attributable to its inputs, and safe to reject or edit. LLM assistance is last and outside correctness: it must never be required for core inventory, reconciliation, validation, graph correctness, generated-index operation, concept resolution, or agent integrations.

## Principles

- **Deterministic first:** repository facts, references, reconciliation, and queries have stable inputs and reproducible outputs.
- **Repository-native output:** generated results live in ordinary Markdown, source-controlled files, and explicit machine-readable exports.
- **Authored intent remains the source of truth:** generated projections do not replace or silently reinterpret hand-authored prose.
- **Projections are generated:** indexes, backlinks, maps, reports, and bundles are views of the underlying repository model.
- **Explicit resolution only:** concept resolution is deterministic only against explicit repository vocabulary, aliases, paths, symbols, headings, active files, or Git changes; ambiguity yields candidates or waits for a concrete target.
- **Daemon as an operational layer:** the daemon only automates or schedules static core capabilities and may retain disposable caches; it adds no exclusive product capability, and the static CLI remains authoritative.
- **Thin integrations:** CLI, MCP, Hermes, Claude Code, plugins, and other adapters use the same deterministic core rather than creating parallel repository models.
- **Optional adapters:** language-specific analysis is enabled deliberately and does not reduce baseline usability.
- **No semantic prose generation in core:** core Demon Docs behavior maintains structure and explicit references, not guessed explanations.

## Dependency and Order

The current foundation precedes all planned phases. Phase 1 supplies the inventory and typed graph required by Phases 2–6. Phase 2 establishes trustworthy links and incremental change handling before the distinct reverse mappings and code-folder indexes in Phase 3. Phase 4 adds optional, deterministic symbol nodes, and Phase 5 builds on the file/path/symbol model to add bounded code and dependency facts. Phase 6 consumes that completed deterministic graph for reproducible projections, including entanglement views and agent context bundles. Phase 7 then exposes those capabilities through thin adapters without creating a competing core.

Phase 8 is lower priority than Phases 1–7: it operationalizes and schedules the completed deterministic graph, reconciliation, projection, and integration capabilities as a disposable service layer. It is sequenced after the static core rather than built independently first, because an early daemon would duplicate or obscure the correctness model, and it does not own MCP or plugin hosting. Phase 9 comes last because proposals must be constrained by deterministic graph and diff data; it remains optional and outside the correctness path.

## Explicit Non-Goals

- Replacing Git, Markdown, or the repository filesystem with a proprietary storage model.
- Treating inferred semantic relationships as equivalent to explicit authored references.
- Generating or rewriting semantic prose as part of deterministic core operations.
- Requiring an always-running daemon for one-shot CLI recovery, CI, rebuilds, or correctness verification.
- Making the daemon the source of repository truth, or giving it capabilities unavailable through the static core.
- Requiring MCP or plugins to be hosted by the daemon rather than exposing them as separate interfaces.
- Requiring an LLM, a network connection, or a language adapter for baseline reconciliation.
- Deterministically resolving an unconstrained free-form concept without an explicit vocabulary, alias, path, symbol, heading, active-file, or Git-change match.
- Claiming folder-level coverage when only individual files are referenced, or vice versa.
- Applying broad, ambiguous, or non-reviewable link and documentation changes.
