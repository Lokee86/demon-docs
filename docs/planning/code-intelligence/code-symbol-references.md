---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-71b2-a949-b6557ef7de3c
document_type: general
policy_exempt: false
summary: This document plans declaration-level code-symbol references, provider resolution states, graph edges, user-facing syntax direction, and move/rename behavior.
---
# Code-Symbol References

Parent index: [Planned Code Intelligence](./INDEX.md)

## Purpose

This document plans declaration-level code-symbol references, provider resolution states, graph edges, user-facing syntax direction, and move/rename behavior.

## Overview

Symbol references would let documentation identify code declarations more precisely than file or folder paths while preserving deterministic unresolved and ambiguous states.

## Current status

Back-burnered architecture plan. Current reverse indexes and codemap evidence operate at existing file/folder and extracted-symbol evidence boundaries; authored declaration references are not yet a supported product contract.

## Expected ownership

The normalized provider seam should own declaration identities and resolution. Markdown syntax and generated projections should consume those facts without becoming language-specific parsers.

This document defines the planned deterministic references from authored documentation to specific declarations in source code. Symbol references are not implemented and belong to the back-burnered polyglot code-graph track. They extend file- and path-level references without replacing them.

The language-neutral provider seam must be implemented before any production language adapter. Symbol facts will feed the projections described in [Code-Folder Reverse Indexes](../../architecture/reverse-indexes.md) and the broader [Code, Dependency, and Entanglement Facts](./code-dependency-and-entanglement.md) work.

## Goals

A documentation page or codemap should be able to target a declaration rather than an entire source file. Precise symbol targets enable:

- reverse links from code back to the governing documentation;
- impact reports that identify the documentation affected by a changed declaration; and
- bounded agent context containing the relevant declarations and their governing docs/indexes.

This is especially useful in legacy or spaghetti code, where a file can contain many unrelated responsibilities and file-level coverage is too broad to guide a reader safely.

## Version 1 Scope

Version 1 recognizes deterministic declaration regions emitted by a language adapter:

- package or module;
- type, class, or interface;
- function;
- method;
- field; and
- constant or variable, where the language adapter supports it.

A declaration is a parser-observable region with a precise source span. Arbitrary conceptual blocks inside a function are not inferred. A comment that describes a conceptual block does not create a symbol unless an explicit, supported mechanism says otherwise.

## Deterministic Basis

Adapters consume language-parser facts: AST nodes, declaration kinds, names, containment, and exact source spans. The adapter does not ask an LLM to decide what a symbol means or where it begins and ends.

No first production adapter is selected yet. The architecture must accept multiple languages from the outset and may normalize output from compiler tooling, Tree-sitter analysis, SCIP-style indexes, language servers, or external code-intelligence providers. A Go provider may use `go/parser`, `go/ast`, `go/token`, `go/types`, or `go/packages`, but those implementation choices must not define the core contract.

The normalized graph is derived from provider output and repository paths, not from formatting conventions or guessed prose.

## Adapter Contract

Every language or tool provider implements the same language-neutral contract:

1. Accept a repository root, a set of candidate source files, and adapter configuration.
2. Parse supported files without modifying them.
3. Emit declaration facts for supported v1 kinds, including containment and exact current source spans.
4. Normalize each fact into a symbol node with the fields below.
5. Report deterministic parse, unsupported-feature, and qualification diagnostics.
6. Return stable ordering for identical repository inputs.

An adapter must identify its language and version or parser mode in the graph metadata. It must not emit inferred conceptual symbols, silently widen a declaration span, or hide an unresolved qualification problem.

### Normalized Symbol Node

Each symbol node contains:

- **repository-relative file path:** normalized from the repository root, using repository path conventions;
- **language:** the adapter language identifier;
- **kind:** one of the supported declaration kinds, or a later adapter-defined extension;
- **qualified name:** the declaration name with deterministic package, module, type, or namespace qualification where available;
- **optional signature/disambiguator:** a normalized signature, receiver, overload marker, or equivalent detail when names alone are insufficient;
- **container:** the containing package/module, type, class, interface, or other declaration node when applicable;
- **current source span:** exact start and end positions for the current file, with enough information to locate the declaration reliably; and
- **optional stable fingerprint:** a parser-derived fingerprint used only for diagnostics and candidate matching across changes, never as proof of identity by itself.

A node's current path and span describe the current checkout. They are not a promise that a symbol has a permanent identity across renames.

## Graph Edges

The Phase 4 graph adds these edges:

- `document references symbol`;
- `symbol declared in file`;
- `symbol member of container`; and
- `file contained by folder`.

Declaration and symbol extraction are Phase 4 groundwork. Calls, general references, imports, later dependency/data edges, and entanglement projections are optional capabilities after this phase; they are not required for code-symbol references, reverse documentation links, or the initial Go adapter. Their separate planning boundary is described in [Code, Dependency, and Entanglement Facts](./code-dependency-and-entanglement.md).

## Resolution Flow and Diagnostics

Resolution is deterministic and follows the same repository snapshot used to build the graph:

1. Select the referenced repository-relative file or candidate file set.
2. Select the enabled language adapter from the file type and configuration.
3. Parse the file and enumerate supported declaration facts.
4. Normalize candidate nodes and compare the explicit selector against path, kind, qualified name, and optional signature/disambiguator.
5. Create a `document references symbol` edge only when exactly one candidate is selected.
6. Emit a stable diagnostic for every non-resolved result.

The diagnostic categories are:

- **resolved:** exactly one supported declaration matched; include the normalized node and current span.
- **missing:** the file, selector, or declaration does not exist in the current repository snapshot.
- **unsupported-language:** no enabled adapter can inspect the target file, or the adapter does not support the requested declaration form.
- **ambiguous:** more than one candidate matches; include deterministic candidate paths, qualified names, signatures where available, and spans.

Diagnostics are reviewable data, not silent fallback behavior. An ambiguous reference is not treated as a file-level reference and does not produce a guessed edge.

## User-Facing Syntax: Not Final

The reference syntax is intentionally not finalized by this design. A non-final example could be a readable path plus symbol selector:

```text
src/ledger/index.go :: package ledger :: function Reconcile
```

A later representation could use YAML/front-matter-style structured metadata next to an ordinary Markdown link:

```yaml
---
# Illustrative only; non-binding.
path: src/ledger/index.go
kind: function
name: Reconcile
---
```

```markdown
[Reconcile](src/ledger/index.go)
```

These examples are illustrative only and are not a compatibility commitment. Any final syntax must be:

- repository-native and easy to review in a diff;
- readable without specialized tooling;
- unambiguous for supported declaration kinds;
- round-trippable without losing selectors or diagnostics;
- valid alongside ordinary Markdown; and
- free of required source-code markers.

The syntax must also permit a file-level link to remain a valid, less precise reference when no symbol selector is supplied.

## Rename and Move Behavior

Line movement is naturally handled by reparsing the current file and recording the declaration's current source span. A file move can be repaired through Git-aware path tracking when the tracked-file evidence identifies the old and new path.

A symbol rename cannot be safely auto-repaired from parsing alone. A changed name can represent a rename, deletion plus addition, or two unrelated declarations. A stable fingerprint may produce reviewable rename candidates, but it must not silently rewrite an ambiguous or merely similar reference. Unresolved candidates remain diagnostics until an author confirms the change.

## Generated Projections

Symbol references feed generated projections without replacing authored documentation:

- resolved symbol references feed distinct code-folder reverse indexes, separate from ordinary forward documentation-folder indexes, and symbol-level reverse documentation projections; reverse-index scope and reconciliation are defined in [Code-Folder Reverse Indexes](../../architecture/reverse-indexes.md);
- symbol-level documentation lists show the documents that explicitly target each declaration;
- impact reports identify affected symbol references and governing documents for a changed declaration; and
- deterministic context bundles contain only the selected declarations plus the governing documents and indexes needed to interpret them.

Symbol resolution does not widen coverage. A symbol-level reference does not imply file- or folder-level coverage, a file-level reference does not imply symbol- or folder-level coverage, and a folder-level reference does not imply coverage of its files, symbols, or descendants. Reverse projections retain these levels and any direct-versus-descendant scope explicitly.

Context bundles are bounded by explicit selectors and configured size or depth limits. They do not expand into arbitrary neighboring code or inferred prose.

## Arbitrary-Region Policy

Declaration references are the default. If a repository needs a region that is not a supported declaration:

- line ranges may be supported as explicitly marked, fragile references;
- named source-comment regions may be offered as an optional escape hatch; and
- source code must not require HTML comments for the feature.

Line ranges move poorly and must produce freshness or staleness diagnostics when surrounding content changes. Named regions, if implemented, remain authored markers with deterministic boundaries; they are not inferred from comment meaning. LLM inference of arbitrary regions is outside the deterministic core.

## Incremental and Cache Behavior

The static CLI must be able to rebuild symbol facts from the repository without persistent service state. Incremental operation may reuse cached adapter output when the relevant file content, adapter version/configuration, and parser inputs are unchanged.

A file change invalidates its declarations and any edges or projections derived from them. A tracked file move invalidates path-based references and may invoke Git-aware repair. Configuration, adapter, or parser-version changes invalidate the affected cache scope. Cache entries must be disposable, attributable to their inputs, and safe to rebuild after corruption or omission.

## Safety Boundaries

- Parsing and normalization never edit source files.
- Generated changes are limited to explicitly managed documentation projections or reviewable repair plans.
- Missing, unsupported, and ambiguous references remain visible diagnostics.
- A symbol fingerprint is a candidate aid, not an authorization to rewrite.
- Ordinary file-level references continue to work without a symbol adapter.
- No network service, daemon, or LLM is required for resolution or correctness.
- Authored prose and explicit selectors remain the source of intent.

## Provider Rollout Strategy

The language-neutral provider contract and fixture providers come first. Production providers may then roll out independently behind explicit capability and configuration checks. No language implementation establishes or privately extends the core contract; each provider must emit equivalent normalized nodes, diagnostics, provenance, and deterministic ordering.

Unsupported languages remain valid repository files and may still participate in file-level links, folder indexes, and file-level reverse references. Provider-specific extensions must not change the language-neutral graph contract.

## Illustrative Go Provider Acceptance Criteria

A later Go provider would be acceptable when it can, deterministically and with focused tests:

- discover package declarations and supported declarations for types, interfaces, functions, methods, fields, and supported constants/variables;
- emit repository-relative paths, qualified names, containment, and exact source spans;
- use `go/parser`, `go/ast`, and `go/token`, with `go/types` or `go/packages` only for documented qualification needs;
- resolve one explicit selector and distinguish resolved, missing, unsupported-language, and ambiguous cases;
- preserve ordinary file-level references and work without a daemon or LLM;
- invalidate and rebuild facts after source edits, file moves, and adapter/configuration changes; and
- produce stable machine-readable graph facts and reviewable diagnostics across repeated runs.

The acceptance suite must include legacy-style files with multiple declarations, duplicate short names in different containers, declarations moved within a file, and references that cannot be safely repaired after a rename.

## Open Decisions

- The final Markdown-compatible selector syntax and whether structured metadata is preferred over an inline selector.
- The canonical source-span encoding: line/column only, byte offsets, or both.
- The exact qualified-name and signature normalization rules for overloaded or generic languages.
- Whether fingerprints are persisted in graph exports or retained only in diagnostics and repair candidates.
- The supported line-range syntax and freshness policy for fragile references.
- The format and limits of deterministic context bundles.
- Which code-intelligence provider should be adapted first after the language-neutral seam is stable.
- Which Go qualification cases require `go/types` or `go/packages`, and how failures are surfaced in a later Go provider.
- The configuration boundary between enabled providers and automatic file-type discovery.

## Implementation sequence

```text
finalize provider and symbol identity contract
-> select non-binding author syntax
-> implement deterministic resolution diagnostics
-> add rename/move behavior
-> add generated projections
-> validate with at least one real provider
```

## Related docs

- [Planned Code Intelligence](INDEX.md)
- [Repository Graph](repository-graph.md)
- [Code Dependency and Entanglement](code-dependency-and-entanglement.md)
- [Reverse Indexes](../../architecture/reverse-indexes.md)
- [Roadmap](../roadmap.md)

## Notes

Illustrative syntax in this plan is non-binding until the provider identity and resolution contract is stable.
