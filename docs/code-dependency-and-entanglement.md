# Deterministic Code, Dependency, and Entanglement Facts

This document describes the back-burnered polyglot code-graph track: bounded code and dependency facts plus projections over those facts. It is optional analysis layered beside the existing documentation/link graph. It is not required for current documentation-folder reconciliation, link repair, codemap extraction, or the initial file/folder reverse-index implementation.

The adapter seam comes before any language implementation. Demon Docs normalizes deterministic facts from existing parsers, compiler tooling, SCIP-style indexes, language servers, or external code-intelligence providers instead of rebuilding those systems inside the project.

## Scope and Deterministic Basis

The extension consumes normalized symbols and containment facts when an enabled provider can supply them. It may add further parser, compiler, build, or bounded static-analysis facts, but it must never treat model inference, prose similarity, or an LLM interpretation as a parser fact.

A provider is language- or tool-specific; the normalized graph contract is not. Provider output must include capabilities, provenance, tool versions, repository-relative identities, and explicit unresolved or unsupported states. Consumers such as codemap ranking and context projection operate on the normalized contract rather than importing provider-specific types.

A fact is emitted only when its source tool or adapter can describe the relationship reproducibly for the configured repository snapshot. The fact records its source file or symbol, target file or symbol where known, relation kind, source span or analysis location, adapter and tool metadata, and any limitations needed to interpret it. The typed graph owns these observed facts; entanglement and impact views consume them as projections.

This document keeps language-specific call-graph and data-dependency details at the capability-contract level. Individual language designs may define how their parser or analysis tool obtains a supported fact without changing the language-neutral graph rules.

## Fact Families

An adapter may report any of the following families when it supports them:

- imports and package/module dependencies;
- calls between functions, methods, or other supported symbols;
- explicit code references between symbols or files;
- implementations, including supported interface or trait relationships;
- code containment and declaration membership;
- reads and writes of supported symbols, fields, files, or resources;
- shared-state access where the state and access path are parser-observable; and
- bounded control-flow facts, such as supported branch, dispatch, or reachability relationships, when the adapter can expose them with stated limits.

These families are not equally available in every language or repository. Containment and declaration facts may be available from syntax alone, while calls, implementations, reads/writes, shared-state access, and control-flow facts may require additional type, build, or analysis inputs. The graph reports the capability and conditions instead of presenting absent analysis as a negative fact.

A normalized observed fact has a relation kind, source and target identities when resolved, a current snapshot, source provenance, adapter/tool identity and version, and a deterministic status. It may point to folders, files, symbols, packages, or other typed nodes already defined by the repository graph.

## Adapter Capabilities and Unresolved States

Each language adapter reports a capability manifest for the current configuration. The manifest should identify supported fact families, language and parser/tool versions, required build or type information, configured scope, known precision limits, and unsupported constructs. Capability reporting is part of the graph metadata and query output.

A requested or discovered relationship uses explicit states:

- **observed:** the adapter emitted one bounded, reproducible fact and its target is resolved;
- **unresolved:** the adapter recognized a relationship or reference but could not select a current target;
- **dynamic:** the target or behavior depends on runtime dispatch, generated names, dynamic loading, or another runtime choice outside the bounded analysis;
- **reflection:** reflection or metaprogramming prevents a deterministic target from being selected under the configured analysis;
- **runtime-only:** the relationship is available only from execution or deployment behavior not included in the static inputs; and
- **unsupported:** the language, construct, tool mode, or required input is outside the adapter capability.

These states are visible diagnostics or fact records with provenance and limits. Only an observed, resolved fact creates an observed dependency edge. Unresolved, dynamic, reflection, runtime-only, and unsupported results do not create guessed edges and are not silently treated as absent dependencies.

An adapter may also report ambiguous or missing source facts using the repository graph's common diagnostic states. A candidate target produced by bounded matching or Git evidence is a reviewable candidate, not an observed relationship.

## Observed Edges, Projections, and Indicators

Observed dependency edges are normalized graph edges emitted from adapter facts. Their relation kind and direction are explicit, such as import, package dependency, call, reference, implementation, read, write, shared-state access, or a supported control-flow relation. Their provenance points to the parser or analysis output and the repository inputs that produced it.

A projection is computed from observed edges and other explicit graph facts. Cycles, hubs, neighborhoods, impact reports, and documentation mismatches are projections, not additional observed edges. A heuristic indicator is weaker still: it summarizes a pattern such as co-change or repeated proximity and must be labeled as an indicator rather than promoted into the dependency graph.

The graph must keep these categories separate in machine-readable exports and user-facing reports. A projection may cite the facts it used, but it cannot become a new source of truth merely because it was generated.

## Entanglement Projections

Entanglement views describe bounded, inspectable relationships that may help navigation, review, and impact analysis. They do not judge code quality or infer ownership. Planned projections include:

- **cycles:** strongly connected groups over a selected observed edge set, with the relation kinds and paths that form each cycle;
- **hubs:** nodes with notable observed fan-in or fan-out under a configured scope, retaining the contributing edge kinds;
- **fan-in and fan-out:** bounded counts and neighboring targets for a selected file, folder, package, or symbol;
- **boundary crossings:** observed relationships that cross configured folder, package, module, repository, or documentation/code boundaries;
- **shared mutable-state hotspots:** locations with multiple observed reads/writes or shared-state accesses, only where the adapter identifies the state and access path;
- **long neighborhoods:** bounded multi-edge neighborhoods around a changed or selected target, with explicit depth and node/edge limits;
- **Git co-change clusters:** files or symbols that change together in a selected history window; and
- **documentation/code mismatches:** differences between explicit documentation references or coverage projections and the observed code targets or changed code facts.

A cycle is a graph pattern, not proof that a cycle is harmful. A hub is a structural count, not proof that a node owns a responsibility. A shared-state hotspot is an observed access pattern, not proof of a bug. A documentation/code mismatch is a prompt for review, not proof that documentation or code is wrong.

Git co-change indicates historical coupling in the selected history and path scope. It does not establish intent, causality, ownership, or a code dependency, and it must remain a separate historical indicator from parser- or tool-observed dependency edges. Co-change clusters must identify their Git range, filtering rules, and evidence rather than appearing as ordinary dependency relationships.

## Bounded Neighborhoods and Impact Reports

Queries and reports must be bounded by explicit inputs and limits. A change or selected target may specify repository scope, edge families, direction, maximum traversal depth, maximum nodes or edges, excluded paths, history range, and adapter capabilities. If a limit or unsupported fact truncates a result, the output reports the limit and the missing or unresolved evidence instead of implying completeness.

An impact report lists potentially related observed targets, documentation references, reverse documentation edges, and relevant projections for the selected change. It must distinguish direct observed relationships from transitive projection results, heuristic indicators, and historical co-change. It must include provenance for the source snapshot, adapter/tool versions, Git range where used, configuration, traversal rules, and truncation status.

Stable ordering is required for identical inputs. Reports should order targets by relation family, direction, canonical graph identity, source location, and deterministic diagnostic or evidence fields. The exact comparator and thresholds remain open decisions, but repeated queries over identical inputs must produce the same result and export order.

## Documentation and Code Mismatches

A mismatch projection may compare explicit documentation references, reverse-index coverage, symbol availability, changed code facts, and configured scope. Examples include a documentation reference whose target is now unresolved, a changed symbol with no explicit governing documentation reference, or a documented target whose current code fact no longer matches its selector.

The projection must state which input facts caused the mismatch and which coverage level or selector was compared. It must not infer that a file should have documentation, that a document owns a responsibility, or that a particular code change requires a prose rewrite. Folder-, file-, and symbol-level coverage remain separate, as do direct and descendant scopes.

## Static CLI, Queries, and Caches

The static CLI/core owns the authoritative path for building, checking, exporting, and querying deterministic code/dependency facts and entanglement projections. It must be possible to run a bounded query or rebuild from the repository snapshot without a daemon, LLM, network service, or pre-existing cache.

Incremental operation may reuse adapter or projection results only when their relevant source files, build/type inputs, configuration, adapter/tool versions, graph schema, and history range are unchanged. Caches are disposable, attributable to those inputs, and safe to delete or rebuild. A stale or corrupt cache cannot change an observed edge or make a projection appear complete.

A daemon may watch repository and Git changes, coalesce events, trigger the same static builds, reconciliation, and projection refreshes, retain disposable caches, and report its own automation health. It does not serve ad hoc graph or impact queries, introduces no exclusive analysis capability, and is not the owner of graph truth. MCP and plugins remain separate interfaces; they may call the static core directly or use their own thin service, but need not be hosted by the daemon.

## Safety Boundaries and Non-Goals

- Parser and bounded-analysis facts are kept separate from model inference, guessed semantics, and generated prose.
- Dynamic, reflection, runtime-only, unresolved, unsupported, ambiguous, and candidate states remain visible and never become guessed edges.
- Co-change is historical evidence, not proof of dependency, intent, causality, ownership, or required design.
- Entanglement projections do not automatically claim that code is bad, owns a responsibility, should be refactored, or has a defect.
- The system does not automatically prescribe refactors, rewrite source code, or rewrite authored documentation from a graph pattern.
- Transitive neighborhoods and impact reports are bounded and labeled when incomplete.
- Heuristic indicators are not serialized or presented as observed dependency facts.
- Baseline documentation reconciliation remains usable when code/dependency analysis is disabled or unavailable.
- A daemon, persistent cache, LLM, and network service are outside the correctness path.
- Language-specific call/data algorithms and semantic dependency inference are deferred to separate designs.

## Initial Acceptance Criteria

The code-graph track is ready to resume implementation when focused fixtures and repeatable static checks can demonstrate that:

- the provider adapter boundary is implemented before a production language adapter;
- two fixture providers can normalize contrasting language/tool outputs into the same consumer contract;
- symbol discovery can remain enabled without requiring dependency analysis, and baseline documentation reconciliation works with all optional code/dependency capabilities disabled;
- adapters report per-language capabilities, versions, limits, and required inputs;
- supported imports, calls, references, implementations, containment, reads/writes, shared-state, and bounded control-flow facts are emitted only when reproducible, with provenance;
- unresolved, dynamic, reflection, runtime-only, unsupported, ambiguous, and missing outcomes remain explicit and do not create guessed edges;
- observed edges, graph projections, and heuristic indicators have separate types and exports;
- cycles, hubs, fan-in/fan-out, boundary crossings, shared-state hotspots, bounded neighborhoods, co-change clusters, and documentation/code mismatches are reproducible, bounded, and provenance-bearing;
- co-change evidence is labeled historical and cannot appear as a dependency edge;
- impact and neighborhood reports preserve stable ordering and disclose limits or missing evidence; and
- static CLI build, check, export, and query operations work without a daemon, LLM, or retained cache.

## Open Decisions

- Which existing code-intelligence source should be adapted first and how optional external binaries are discovered.
- The language-neutral relation taxonomy and normalization rules for each fact family.
- Which adapter inputs are required for each capability and how build, type, generated-code, and environment boundaries are declared.
- The representation and retention policy for unresolved, dynamic, reflection, runtime-only, and unsupported fact records.
- The precision and scope limits for control-flow, dispatch, reads/writes, and shared-state analysis.
- Whether observed facts are stored as direct graph edges, fact records plus edges, or both in machine-readable exports.
- The algorithms, thresholds, and stable ordering for cycles, hubs, fan-in/fan-out, boundary crossings, and long neighborhoods.
- The Git history range, rename handling, path filters, and minimum evidence for co-change clusters.
- The definition and default scope of documentation/code mismatch reports.
- Query syntax, export limits, cache keys, and invalidation granularity for large repositories.
- The separate language-specific designs for call graphs and data-dependency facts.
