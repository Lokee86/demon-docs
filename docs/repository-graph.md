# Planned Polyglot Repository and Code Graph

This document describes the back-burnered language-neutral graph that will join Demon Docs' existing documentation/link facts with bounded code facts from polyglot providers. It is not required for current folder-index reconciliation, link repair, codemap extraction, or the initial reverse-index implementation.

Demon Docs already has a focused repository-local Markdown link graph. The future code graph must add definitions, references, calls, imports, implementations, containment, and other reproducible code relationships without replacing that working link model or rebuilding a general code-intelligence platform from scratch.

## Observed Layers

The graph keeps three observed layers distinct:

1. **Physical repository structure:** the current checkout's repositories, folders, files, configured boundaries, and Git-visible state.
2. **Authored documentation and reference intent:** Markdown documents, headings, anchors, indexes, explicit concepts and aliases, ordinary links, and explicit references to code targets.
3. **Parsed code structure:** code files and language-adapter facts such as declarations, symbols, source spans, and containment.

A fact from one layer may be connected to a fact in another layer by an explicit or parser-observable relationship. The graph does not fill gaps between layers with semantic guesses. In particular, authored prose is not converted into a concept or code relationship merely because it sounds similar to a path, symbol, or heading.

The documentation/link layer remains owned by existing Demon Docs scanners and state. Code facts arrive through a separate provider seam and are normalized before codemap inference or context projection consumes them.

## Polyglot Provider Boundary

The provider adapter seam is the first required implementation step when this track resumes. A provider may wrap Tree-sitter analysis, compiler tooling, SCIP-style indexes, language servers, or an external code-intelligence product. The Demon Docs core must not depend on one provider's storage format or language-specific identities.

Each provider reports a capability manifest and a deterministic snapshot containing normalized nodes, edges, diagnostics, provenance, language identity, tool version, repository-relative paths, and source locations. Unsupported or unresolved relationship families remain explicit rather than being treated as negative facts.

The seam must exist before the first language implementation. A Go-only graph wired directly into core packages would make later Ruby, GDScript, Python, TypeScript, or other support unnecessarily expensive and would not satisfy the intended product scope.

Demon Docs should reuse existing deterministic analyzers where practical. It should not implement another parser platform, compiler front end, general call-graph service, or graph database merely to obtain provider facts.

## Node Types

The initial language-neutral node vocabulary includes:

- **repository:** the configured graph root and its repository metadata;
- **folder:** a repository-relative directory, including its physical containment;
- **file:** a repository-relative filesystem object with its current checkout status and type;
- **Markdown document:** a Markdown file with parsed document metadata;
- **heading or anchor:** a parser-observed heading, explicit anchor, or stable document location;
- **index:** an ordinary forward documentation-folder index or a distinct reverse code-folder index, with its ownership and scope;
- **explicit concept:** a concept declared or otherwise explicitly named by repository data;
- **alias:** an explicitly authored alternate name for a concept, path, symbol, or other supported target;
- **code symbol:** a normalized declaration emitted by an enabled language adapter; and
- **diagnostic:** a reviewable status record attached to a source fact, reference, target, or candidate set.

A file may be both a physical file node and a Markdown document node. An index is a documentation projection with an explicit index kind; a reverse index is not represented as an ordinary forward index. Diagnostics are graph records for resolution and parsing outcomes, not semantic repository objects.

Exact language-specific symbol kinds and code/dependency data structures belong in separate focused documents. This graph only requires a language-neutral symbol identity and provenance contract.

## Edge Types

The initial edge vocabulary includes:

- **containment:** repository contains folders/files; folders contain direct child folders/files; documents contain headings or anchors; symbols contain nested symbols where an adapter reports containment;
- **forward index membership:** a forward documentation index inventories a direct documentation child of its owning folder;
- **ordinary link:** a Markdown source location links to a document, heading, anchor, file, or other valid Markdown target;
- **explicit code reference:** authored documentation references a code folder, file, or symbol using an accepted explicit selector;
- **reverse documentation edge:** a resolved code target is documented by the source document/reference that explicitly names it; this is derived from an explicit code-reference edge, not authored independently;
- **symbol declaration:** a code symbol is declared in a code file; and
- **symbol containment:** a symbol is contained by its reported package, type, module, or other symbol container.

Provider facts may connect code nodes through imports, references, calls, implementations, reads/writes, or other bounded relation families when the provider declares that capability. Their normalized vocabulary and limits are defined in [Code, Dependency, and Entanglement Facts](code-dependency-and-entanglement.md). No code edge is required for current documentation reconciliation.

Edge direction and names are part of the graph schema. A reverse documentation edge points from code target to documentation for projection and query purposes, while its provenance points back to the authored explicit reference that produced it.

## Identity, Ordering, and Provenance

All current-checkout filesystem identities are based on a normalized repository-relative path and node kind. Paths use repository conventions, not host-specific absolute paths, and retain enough metadata to report the current checkout status. A repository node identifies the configured root and may include Git identity metadata when available; Git metadata does not replace the current-root identity.

A Markdown document is identified by its current file identity. Headings and anchors are identified by their containing document plus an explicit anchor or parser-observed heading location and normalized text/slug as applicable. An index identity includes its owner and explicit index kind. Explicit concepts and aliases require an authored declaration or other explicit source location; their names are not global identities merely because they occur in prose.

A symbol identity includes its current repository-relative file, adapter/language identity, declaration kind, qualified name, and any required disambiguator. Its current source span locates the current checkout and is not a permanent identity across a file move or symbol rename. A parser-derived fingerprint may support a candidate match, but it cannot authorize an identity rewrite by itself.

Git move, deletion, and rename evidence is separate from current identity. An unambiguous path move may provide provenance for a reviewable repair candidate. A possible rename or fingerprint match must not turn an old and new node into one identity or create an edge without sufficient explicit evidence.

Graph nodes and edges are ordered by deterministic tuples built from node kind, normalized repository-relative path, qualified name or selector, source location, and stable diagnostic fields. Exact ordering rules are schema decisions, but identical repository inputs, configuration, adapter versions, and Git evidence must produce identical ordering and export bytes.

An edge's normalized identity contains its relation type, source node, target node, explicit selector where applicable, and source location. Exact duplicate edges may be coalesced under that identity, but their distinct provenance locations and source records must remain available. Different selectors, locations, statuses, or derivations are not silently collapsed.

Every observed or derived fact carries provenance appropriate to its layer: repository snapshot or checkout, repository-relative source path, source span or Markdown location when available, parser or adapter identity and version for parsed code, configuration inputs, and derivation links for generated edges. A derived edge identifies the explicit or parsed facts from which it was produced.

## Resolution States

References and parser outcomes use explicit states rather than guessed targets:

- **resolved:** exactly one supported target or fact was selected, so the corresponding edge may be emitted;
- **missing:** the requested path, document, heading, anchor, file, or symbol is absent from the current snapshot;
- **ambiguous:** multiple candidates match, so the diagnostic retains a deterministic candidate list and no target edge is emitted;
- **unsupported:** no enabled parser or adapter can provide the requested fact, or the selector uses an unsupported form; and
- **candidate:** Git evidence, a bounded fingerprint, or another reviewable signal suggests a possible move or rename, but the match is not authorized as identity or an edge.

Diagnostics for non-resolved states are stable graph records and export entries. A candidate target is not a resolved target. A missing, ambiguous, unsupported, or candidate reference is not silently widened to a containing file, folder, heading, or concept. A resolved edge is created only when the graph's explicit selector and available observed facts select one target under the configured rules.

## Rebuild, Invalidation, and Caches

The static core owns graph construction, resolution, schema validation, diagnostics, and generated-projection inputs. A full rebuild reads the current repository snapshot and configured inputs, parses supported documentation and code, resolves explicit references, and reconstructs the graph without requiring persistent service state.

Incremental operation may reuse facts only when their relevant inputs are unchanged. Changes to a file, Markdown structure, explicit reference, folder containment, Git-visible state, adapter/parser version, adapter configuration, or graph schema invalidate the affected nodes, edges, diagnostics, and projections. The invalidation result must be equivalent to rebuilding the affected scope from current repository inputs.

Caches may hold parsed facts, normalized nodes, or derived projections, but they are disposable and attributable to their inputs. A missing, stale, corrupt, or deleted cache cannot change graph truth or block a static rebuild. The daemon, if used, automates these static operations; it does not own a separate graph model.

## Machine-Readable Export

The graph should have a deterministic machine-readable export for verification, integrations, and later projections. The export is planned to include:

- schema and graph-format version;
- repository identity and current-snapshot metadata;
- configuration and enabled-adapter metadata sufficient to explain the result;
- ordered nodes and edges with their kinds and canonical identities;
- provenance for observed and derived facts;
- diagnostics and candidate lists; and
- cache-independent inputs or fingerprints needed to explain invalidation.

Export metadata must identify adapter and parser versions for code facts and distinguish current-checkout identities from Git rename evidence. Export ordering must be stable, and a schema change must be explicit rather than silently changing the meaning of existing fields.

## Generated Projections

Forward indexes, code-folder reverse indexes, backlinks, repository maps, impact reports, entanglement views, context bundles, and machine-readable query results are projections of the graph. They are generated views, not graph truth and not authored intent.

A projection may retain links to its source nodes, edges, diagnostics, and provenance so that a reader can inspect why an entry exists. It must not be fed back as an authored relationship merely because it was generated. Managed Markdown sections are the only regions a projection may update, and authored content outside those sections remains the source of intent.

## Static-Core Ownership

The static CLI is the authoritative path for building, checking, validating, exporting, and rebuilding the graph and its projections. It must remain usable when daemon state is absent and must be able to discard and reconstruct disposable caches from repository inputs.

Interfaces such as MCP, Hermes, Claude Code, plugins, or other adapters may request graph facts or projections through the same core. They do not define alternate node identities, resolution rules, or correctness paths. A daemon may schedule and coalesce the same static operations, but it adds no exclusive graph capability.

## Safety Boundaries and Non-Goals

- The graph records observed structure, explicit authored references, and bounded parser facts; it does not infer repository meaning.
- Generated projections never become authored intent or a replacement for graph inputs.
- Missing, ambiguous, unsupported, and candidate states remain visible and never create guessed edges.
- Source code and authored prose are not rewritten by graph construction or export.
- Current paths and source spans are not treated as permanent identity across moves or renames without explicit reviewable evidence.
- This design does not define language-specific call graphs, data-dependency analysis, semantic dependency inference, or arbitrary conceptual symbols.
- An always-running daemon, network service, LLM, or persistent cache is not required for graph correctness.
- Free-form concepts are not resolved deterministically without explicit vocabulary, aliases, paths, symbols, headings, active files, or Git changes.

## Initial Acceptance Criteria

The graph track is ready to resume implementation when focused fixtures and repeatable static checks can demonstrate that:

- a language-neutral provider contract exists before any language-specific adapter is wired into the core;
- at least two contrasting fixture providers can emit the same normalized identities and relations without changing consumer code;
- all three observed layers are represented without collapsing authored intent into parsed or physical facts;
- the node and edge vocabulary distinguishes forward membership, ordinary links, explicit code references, derived reverse documentation edges, symbols, and diagnostics;
- canonical identities and normalized repository-relative paths produce stable ordering and exports across repeated builds;
- exact duplicate edges are handled deterministically without losing provenance;
- resolved, missing, ambiguous, unsupported, and candidate outcomes are distinguishable and only resolved outcomes create target edges;
- file moves, deletions, and rename candidates preserve current-checkout identity and do not produce guessed repairs;
- changing relevant source, configuration, adapter, or schema inputs invalidates affected facts and projections, while deleting caches does not prevent a full rebuild;
- machine-readable exports include schema/version metadata, provenance, diagnostics, and deterministic node/edge ordering; and
- static CLI rebuild and validation work without a daemon, LLM, or pre-existing cache.

## Open Decisions

- The first provider integration and whether it consumes SCIP, Tree-sitter output, a language-server index, or another existing graph product.
- The canonical provider capability manifest and normalized relation taxonomy.
- The canonical serialized schema, field names, compatibility policy, and versioning rules.
- The exact repository identity metadata and how worktrees or nested repositories are represented.
- Path normalization and case-sensitivity rules across host filesystems.
- The canonical identity and duplicate policy for headings, generated anchors, concepts, aliases, and repeated references.
- Whether diagnostics are first-class exported nodes, attached records, or both.
- The final provenance representation for derived edges and candidate rename evidence.
- The exact Git evidence and review workflow for move and rename candidates.
- The minimum adapter metadata and source-span encoding required in graph exports.
- Cache boundaries and invalidation granularity for large repositories.
- The later dependency-edge vocabulary and the separate design that will define language-specific call/data behavior.
