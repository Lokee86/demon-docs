# Deterministic Agent Context and Integrations

This document describes the planned graph-based context retrieval layer and its host integrations. Context assembly is not implemented and follows the back-burnered polyglot code-graph provider contract. The repository demon already exposes a generic operational `agent` feeder, but feeder lifecycle is separate from context delivery.

When implemented, one static core will assemble context from deterministic documentation/link facts plus normalized polyglot code facts. CLI, MCP, native plugins, and editor or agent adapters will remain different delivery surfaces around that core. This is not conventional embedding/vector RAG and does not create a second repository model.

## One Core, Thin Adapters

The same static graph and projection APIs should be exposed through:

- the Demon Docs CLI;
- MCP for broad interoperability;
- Hermes;
- Claude Code;
- Codex;
- Cursor;
- Gemini; and
- other thin adapters that translate a host request into core inputs and return the resulting context, candidates, diagnostics, or status.

These integrations do not own repository truth, target resolution, graph identity, or correctness rules. A native plugin may add host hooks, commands, UI, or automatic delivery when the host provides those extension points. MCP provides a broadly interoperable protocol surface. Neither is required to host or replace the static core, and no claim is made that every listed host currently exposes identical plugin hooks.

The daemon, when available, watches repository and Git changes, coalesces events, and triggers the same static rebuild, reconciliation, and projection operations. It does not deliver agent context or host integrations. MCP and native plugins call the static core or their own thin service/interface and remain separate from the daemon.

Operational feeder registration is also separate from context delivery. An MCP or native host adapter may register a generic `agent` feeder while a job or session is active and remove it on every terminal path. That feeder only keeps the repository watcher alive; it does not make the daemon an MCP server, context service, or host-specific integration.

## Context Inputs

A context request may provide any combination of the following deterministic inputs:

- explicit repository-relative paths, files, folders, symbols, or selectors;
- the caller's current directory;
- active, read, or edited files supplied by the host;
- current Git changes and an explicitly selected history range;
- exact repository vocabulary, including declared concepts and observed names;
- headings and anchors from Markdown documents; and
- explicitly configured aliases.

The request also declares a context mode, repository scope, allowed graph facts, and applicable size or traversal limits. Host activity is input evidence, not an instruction to infer a responsibility or to search the whole repository. An adapter must preserve which inputs were explicit and which came from host state.

## Deterministic Resolution

Explicit paths, symbols, headings, aliases, and concepts are resolved against the current graph snapshot and configured repository boundaries. A free-form prompt may produce deterministic candidates only through lexical normalization and matches against explicit repository vocabulary. Lexical normalization may define case, separators, punctuation, and other documented token rules; it does not provide magical semantic resolution and does not use embeddings to invent a target.

A request that does not identify one target returns candidates or waits for a concrete target. It must not silently choose a vaguely similar folder, file, symbol, or document. Candidate results include the matched vocabulary or selector, canonical graph identity, source location, match class, and provenance.

Ranking follows fixed rule tiers so that the same inputs produce the same candidate order. The planned precedence is:

1. an explicit exact path, symbol, heading, or selector;
2. an exact configured alias or explicitly declared concept;
3. an exact normalized repository vocabulary or heading match; and
4. a documented lexical token match within repository vocabulary.

Within a tier, candidates are ordered by explicit scope, active or edited file relevance, current-directory proximity, Git-change relevance, and canonical graph identity, using fixed tie-breakers. The exact numeric scores, if scores are exposed at all, remain open decisions. A score never overrides an ambiguity rule or authorizes a guessed edge.

Every candidate and selected seed carries provenance for the request input, normalization rule, graph snapshot, repository path or symbol, vocabulary/alias declaration, host activity signal, and ranking tier. An empty result distinguishes no explicit or lexical match from an unavailable graph or unsupported adapter.

## Context Modes

Modes select bounded inputs and projections; they do not change graph truth or resolution rules.

- **orientation:** repository/folder maps, nearby indexes, headings, direct documentation, and selected target identity for understanding where the request sits;
- **implementation:** selected files or symbols, relevant authored references, declarations, containment, and explicitly supported dependency edges within configured limits;
- **change-impact:** current Git changes, affected documentation and code targets, bounded graph neighborhoods, reverse documentation edges, and impact diagnostics;
- **documentation:** governing documents, headings, forward and reverse indexes, explicit code references, coverage gaps, and unresolved reference diagnostics; and
- **entanglement:** bounded cycles, hubs, fan-in/fan-out, boundary crossings, shared-state indicators, long neighborhoods, co-change indicators, and documentation/code mismatches, each labeled by evidence type.

A mode may omit unavailable capabilities without treating them as negative facts. The response reports which requested inputs, adapters, or projections were unavailable or truncated.

## Context Assembly and Bounds

Assembly starts with resolved explicit seeds and deterministic candidates. It traverses only configured graph edge types and directions for the selected mode. Each traversal has explicit limits such as maximum depth, nodes, edges, files, source bytes, and output tokens or serialized size.

The assembler:

1. resolves request inputs and reports candidates or ambiguity before broad traversal;
2. selects permitted seeds and mode-specific edge families;
3. traverses the typed graph and selected projections within the configured bounds;
4. suppresses duplicates by canonical graph identity while retaining provenance for every contributing source;
5. orders facts, authored references, projections, and indicators deterministically;
6. applies token, byte, and item budgets; and
7. reports truncation, omitted categories, unresolved inputs, and unavailable capabilities explicitly.

Stable ordering uses source category, relation or projection kind, canonical graph identity, source location, and deterministic diagnostic fields. Authored documentation, observed parser facts, generated projections, and historical or heuristic indicators remain distinguishable in the assembled context. A context bundle is a view over the graph, not new graph truth.

The system must never inject ambiguous or oversized context silently. Ambiguous seeds require candidate presentation or a concrete target. A budget overrun produces a visible truncation result or refuses automatic delivery according to the caller's policy; it does not silently drop the most inconvenient evidence or claim completeness.

Context may refresh after concrete agent activity such as selecting a candidate, changing the current directory, reading or editing a file, changing an active-file set, or changing Git state. Refresh invalidates affected seeds and projections and reassembles from the updated graph inputs. It must not treat an unqualified follow-up as permission to broaden scope indefinitely.

## Illustrative Public Operations

Names and final schemas are open, but the static CLI may expose operations shaped like:

```text
ddocs context --mode orientation --path docs/README.md --format json
ddocs context --mode implementation --symbol src/ledger/index.go::Reconcile --budget 12000
ddocs context --mode change-impact --git-range HEAD~1..HEAD --format json
ddocs context --mode documentation --concept "reverse index" --candidates

ddocs graph query --path src/ledger --edges references,contains --depth 2
ddocs graph export --format json
```

An MCP adapter may expose corresponding tools such as `doc_ledger_context`, `doc_ledger_candidates`, `doc_ledger_graph_query`, and `doc_ledger_status`. These names and request/response schemas are illustrative. The adapter should return the same canonical identities, provenance, ordering, diagnostics, and truncation metadata as the CLI rather than implementing a parallel resolver.

## Automatic Delivery and Plugin Status

Automatic context injection is opt-in per host and request policy. A native plugin may deliver a context bundle after a supported host event, but it must apply the same ambiguity, scope, size, provenance, and privacy rules as an explicit request. A host without the relevant hook remains usable through its CLI, MCP, or another supported adapter; integrations must not assume identical hook or event support across Hermes, Claude Code, Codex, Cursor, Gemini, or other hosts.

A plugin or integration status view may report separate health categories:

- **installation:** whether the adapter or plugin is installed and its version;
- **repository:** whether the configured repository and permissions are available;
- **graph:** whether the static graph is present, current, valid, or needs a rebuild; and
- **daemon:** whether the optional automation loop is reachable, current, or healthy; this reports daemon automation health only.

Daemon status is operational only. A missing, stopped, or unhealthy daemon must not imply that the static graph or context core is unavailable. Static CLI/core availability is the correctness baseline.

## LLM Boundary

An optional LLM may summarize an already assembled deterministic context or propose a change after the caller reviews its inputs. It must not resolve ambiguous concepts, invent graph edges, select hidden context, establish repository truth, or become necessary for context correctness. LLM use, provider, prompt, and output provenance must remain visible to the caller, and remote delivery requires an explicit policy.

## Security and Privacy

Context assembly should respect repository-relative scope, configured exclusions, filesystem permissions, host authorization, and the caller's requested budgets. It must not use a plugin or MCP request to bypass repository access controls or inject unrelated repository content. Repository content is data for the resolver, not an instruction to change resolver scope or security policy.

Remote MCP consumers, host plugins, and optional LLM providers require explicit transport and data-sharing policy. Requests and bundles should identify included paths, source locations, provenance, truncation, and delivery target so operators can audit what was exposed. Redaction or secret-exclusion policy remains configurable and must not be represented as a guarantee that arbitrary secrets have been detected. Automatic delivery should default to refusal rather than broadening scope when authorization, target resolution, or size policy is unclear.

## Static Core, Cache, and Daemon Boundary

The static CLI and core own candidate resolution, ranking, graph traversal, context assembly, serialization, diagnostics, and refresh semantics. They must be able to build and query context from repository inputs without a daemon, LLM, network service, or retained cache.

Caches may retain parsed graph facts, resolved candidates, or bounded context projections only when they are keyed and attributable to the relevant repository snapshot, graph schema, adapter/configuration, host activity inputs, and request limits. Caches are disposable and safe to delete; stale caches cannot silently produce context for the wrong checkout or bypass a refresh.

The daemon may watch repository and Git changes, coalesce events, trigger the same static rebuild, reconciliation, and projection operations, and retain disposable caches. It reports its own automation health but does not deliver agent context, host integrations, or own a query service, and it does not own repository truth. MCP and native plugins call the static core or their own thin service/interface and remain separate from the daemon.

## Safety Boundaries and Non-Goals

- Graph-based context retrieval is deterministic from explicit inputs, repository vocabulary, graph facts, and fixed bounds; it is not embedding/vector RAG.
- Free-form lexical matches produce candidates only and never authorize a guessed target.
- Observed facts, authored intent, generated projections, and historical or heuristic indicators remain labeled and separate.
- Ambiguous or oversized context is never injected silently, and truncation is always reported.
- A plugin, MCP server, daemon, or host adapter does not define alternate graph truth or resolution rules.
- Context assembly does not infer ownership, responsibility, correctness, refactor needs, or semantic intent from proximity or vocabulary alone.
- Automatic delivery does not bypass permissions, configured exclusions, repository scope, or host policy.
- The LLM boundary is optional and outside graph, resolution, ranking, and delivery correctness.
- No claim is made that every supported host has identical native hooks, lifecycle events, or plugin APIs.

## Evaluation Boundary

Context assembly correctness can be tested deterministically with fixtures, but claims that injected context improves agent implementation require a separate empirical benchmark. The future benchmark should compare matched no-context and context-injected runs across repositories with independently assessed code and documentation quality. Benchmark-specific oracle data must remain outside context generation and agent-visible inputs.

See [Context-Injection Benchmarking](context-injection-benchmarking.md).

## Benchmark and Validation Direction

The context system should eventually be tested against authentic historical OSS tasks across independently assessed code-quality and documentation-quality quadrants. Each treatment run should be paired with the same repository snapshot and task without Demon Docs context, while a deliberately constructed repository validates the harness itself.

The benchmark is future research rather than a current implementation gate. Corpus preparation, pinned task manifests, deterministic bundle inspection, leakage checks, and harness dry runs can proceed before repeated paid model trials. See [Context-Injection Benchmarking](context-injection-benchmarking.md).

## Initial Acceptance Criteria

The design is ready for implementation planning when focused fixtures and adapter checks can demonstrate that:

- CLI, MCP, Hermes, Claude Code, Codex, Cursor, Gemini, and other adapters can request the same core context without defining competing repository models, while host-hook differences remain explicit;
- explicit paths, symbols, concepts, current directory, active/read/edited files, Git changes, headings, vocabulary, and aliases are preserved as distinct context inputs with provenance;
- lexical normalization and explicit vocabulary produce stable candidate lists, fixed ranking tiers, and visible ambiguity without magical semantic resolution;
- each context mode applies declared edge families, budgets, duplicate suppression, stable ordering, and truncation reporting;
- observed facts, authored intent, projections, and heuristic indicators remain distinguishable in context output;
- context refresh after concrete agent activity invalidates affected data and does not silently broaden scope;
- automatic injection refuses ambiguous or oversized bundles and reports the reason;
- plugin status separates installation, repository, graph, and operational daemon health, with daemon absence not blocking static use;
- static CLI build/query/export works without a daemon, LLM, or retained cache; and
- optional LLM summarization or proposal behavior cannot change candidates, graph facts, or correctness decisions.

## Open Decisions

- Final CLI command names, MCP tool names, request/response schemas, and context serialization format.
- The exact lexical normalization rules, fixed ranking tiers, scope-proximity rules, and whether numeric scores are exposed.
- Which host activity signals count as active, read, or edited files and how long they remain relevant.
- Default modes, graph edge allowlists, traversal directions, token/byte/item budgets, and truncation policies.
- The canonical duplicate identity and provenance format for contexts assembled from multiple seeds.
- Refresh triggers, cache keys, invalidation granularity, and behavior after a candidate is selected.
- Automatic-delivery consent, authorization, redaction, audit, and remote data-sharing policies.
- Plugin status schema and the optional daemon connection protocol.
- The minimum common adapter surface versus host-specific hooks for Hermes, Claude Code, Codex, Cursor, Gemini, and other hosts.
- LLM provider configuration and the provenance format for summaries or proposals derived from deterministic context.
