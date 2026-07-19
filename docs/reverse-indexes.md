# Code-Folder Reverse Indexes

This document describes the code-folder reverse-index boundary and the current initial implementation. File- and folder-level codemap projection is implemented; symbol adapters, richer coverage reports, and move-aware repair remain later work.

## Purpose and Index Type

A forward documentation-folder index inventories the direct documentation children of an owning documentation folder. Its source of truth is the current folder and file inventory, and its generated entries describe that documentation tree.

A code-folder reverse index is a distinct index type. It is not an extension of an ordinary documentation-folder index and does not replace one. Its source of truth is the set of explicit documentation references resolved against code targets. It projects those references back onto code folders, files, and symbols so that code-local views can show governing documentation and visible coverage gaps.

The two directions reconcile differently:

- forward indexes start at a documentation folder and enumerate its direct documentation children;
- reverse indexes start at authored documentation references, resolve their code targets, and group the resolved references by code target.

A reverse index must not be used as input to rebuild the forward documentation tree. Its output locations are nevertheless selected by explicit recursive traversal roots: the roots decide where indexes may exist, while authored codemap references decide which documentation backlinks appear.

## Inputs

A reverse-index build uses one repository snapshot and the facts available from that snapshot:

- the repository-relative folder and file inventory, including configured Git and ignore boundaries;
- documentation files and their explicit code-folder, code-file, or code-symbol references;
- the typed repository graph, including containment and reference-resolution facts;
- code-file identities and, when enabled, normalized symbol nodes from language adapters;
- configured code roots, exclusions, and index scope; and
- Git-aware move, deletion, and rename evidence when the operation has access to it.

References remain authored inputs. A prose similarity, guessed responsibility, or inferred dependency does not create reverse coverage. A final reference syntax and the exact way a reference declares direct or descendant scope remain open decisions.

## Generated Outputs and Ownership

For each configured reverse-index scope, the planned generated projection includes:

- the indexed code target and its kind: folder, file, or symbol;
- resolved documentation references, grouped by target and retaining their source document and explicit selector;
- coverage status at the target's own level;
- visible direct-code-file coverage gaps;
- stale, missing, ambiguous, or unsupported-resolution diagnostics; and
- deterministic links or selectors back to the source documentation and target where those links are valid.

The output is owned only within explicitly managed Markdown sections or another explicitly configured generated region. Reverse-index sections have their own markers and ownership; they must not claim or rewrite the managed sections of ordinary forward indexes. Authored prose outside the reverse-index sections is preserved. A reverse-index rebuild may produce a reviewable repair plan for authored references, but does not silently rewrite them.

## Reconciliation Direction and Ordering

Reconciliation proceeds from documentation to code:

1. scan the selected documentation files and extract explicit references;
2. resolve each reference against the same repository snapshot and enabled adapters;
3. classify the result as resolved, stale, missing, ambiguous, unsupported, or a rename/move candidate;
4. project resolved edges onto the configured code-folder scopes;
5. enumerate eligible direct code children so undocumented files remain visible as gaps; and
6. compare the resulting projection with the owned generated sections and plan or apply only the managed changes.

The reverse projection is never its own source of truth. A full static rebuild recomputes it from repository inputs; incremental operation may limit the scan to affected documents, targets, adapters, or Git changes only when the result is equivalent to that rebuild.

Stable output uses normalized repository-relative paths and deterministic target kinds, qualified symbol names, source-document paths, and explicit selectors. The intended ordering is target kind and target path, then symbol qualification, then source document and reference location, with deterministic tie-breakers for selectors and diagnostics. Exact comparator and duplicate-edge rules remain open, but identical inputs must produce identical ordering and bytes.

## Coverage Levels and Scope

Folder-, file-, and symbol-level coverage are separate facts:

- a folder reference covers the referenced folder target only;
- a file reference covers the referenced file target only; and
- a symbol reference covers the referenced declaration only.

None of these levels implies another. A documented folder does not mean that every file or symbol below it is documented. A documented file does not establish documentation for its containing folder or its declarations. A documented symbol does not establish file-level or folder-level coverage.

Direct and descendant coverage must also remain distinct. A direct reference targets the named code object. A descendant projection may show references below a folder only when the reference or configuration explicitly requests descendant scope. Recursive aggregation is not automatic because it would turn broad folder references into unsupported claims about every descendant and would hide the difference between intentional coverage and incidental containment.

Every eligible undocumented direct code file in an indexed scope appears as a visible coverage gap. It is not silently omitted merely because no documentation reference points to it. Files excluded by configuration remain outside the scope; unsupported symbol analysis does not remove the file from file-level coverage.

Multiple documents may reference one code target. The reverse projection retains each distinct resolved source-to-target relationship, groups them deterministically, and does not select one document as authoritative merely because it appears first. Identical repeated edges may be deduplicated only under a documented, deterministic identity rule that preserves any distinct selectors or source locations.

## Diagnostics, Moves, and Deletions

A stale or missing reference remains visible as a diagnostic and does not create a resolved reverse edge. An ambiguous reference produces deterministic candidates and does not choose a target. An unresolved reference is never promoted from symbol-level to file- or folder-level coverage as a silent fallback.

A file-level reference remains valid without a language adapter. A symbol-level reference requires an enabled adapter that can emit and resolve the requested declaration. When the adapter is unavailable or does not support the requested symbol form, the result is an unsupported-symbol diagnostic; the file may still appear with its independent file-level status.

When Git evidence identifies an unambiguous file move, the planner may associate the current target with the prior path and propose a reference repair. A deletion removes the target from resolved current coverage but leaves references to it visible as stale or missing. A possible rename, ambiguous move, symbol rename, or fingerprint match is a candidate only. It must be reported for review rather than silently rewriting an authored reference.

## Relationship to the Repository Model

Ordinary Markdown links remain ordinary Markdown links. They are validated and represented as link edges according to the Markdown/link model; a link is not automatically code coverage unless it uses an accepted explicit code-target reference. Reverse indexes may render navigational links to ordinary documents, but generated backlinks do not replace authored links or change their target semantics.

The typed repository graph is the shared deterministic model of paths, folders, files, documents, references, containment, and resolved targets. A reverse index is a generated projection over that graph, not a second graph and not a new source of repository truth.

Language adapters contribute bounded symbol facts, source spans, and diagnostics. They do not infer conceptual symbols or semantic documentation relationships. Repositories without a supported adapter remain usable at folder and file level, with unsupported symbol references kept visible.

## Current Scope Selection

Reverse indexing has no repository-wide implicit scope. Configure one or more repository-relative roots:

```toml
[reverse_index]
roots = ["client", "services/game-server"]
```

The selected roots are traversed recursively. Source files remain visible even when they have no documentation backlink, while resolved codemap targets add file- or folder-level backlinks. Nested `.docignore` files are loaded during traversal and apply relative to the directory containing them.

Reverse indexes use the same `check`, `fix`, and `watch` command family as documentation indexes and links. Select them with `-r` / `--reverse`. `--reverse-root PATH` overrides configured roots for one invocation and may be repeated:

```bash
ddocs check -r --reverse-root services/game-server
ddocs fix -r --reverse-root /absolute/path/inside/the/repository/client
ddocs watch -r --once --reverse-root client
```

Relative override paths resolve from the current working directory. Absolute paths must remain inside the repository. Repository root, documentation-root, ignored, and nested-worktree scopes are rejected.

Codemap section headings are configured separately:

```toml
[codemap]
headings = ["Code map", "Implementation map"]
```

Reverse reconciliation fails rather than silently producing an ungrounded projection when no configured codemap section exists. A matching section with no code targets is reported as a distinct error.

## CLI and Daemon Boundary

The static CLI can build, check, fix, and foreground-watch reverse indexes through `check -r`, `fix -r`, and `watch -r`. Check must work without a running service. Fix may update only explicitly managed generated sections and must leave unresolved authored references as reviewable diagnostics or candidates.

A daemon only automates or schedules these same static operations. It may coalesce changes and retain disposable caches, but it does not own reverse-index correctness, repository truth, or a daemon-only recovery path. Removing its state must leave the CLI able to rebuild and check the projection. MCP, plugins, and other interfaces are separate adapters and need not be hosted by the daemon.

## Safety Boundaries

- No inferred semantic prose or guessed documentation relationship creates coverage.
- No ambiguous, stale, missing, or unsupported target is silently converted into a resolved edge.
- No coverage level implies a different level without an explicit rule and scope.
- No recursive descendant aggregation occurs without an explicit request or configuration.
- Undocumented eligible direct code files remain visible as gaps.
- Generated changes stay inside owned sections; authored content and source files are not rewritten by index generation.
- Reverse output is reproducible from inspectable repository inputs and disposable caches are never authoritative.
- LLM assistance, network access, and a daemon are outside the correctness path.

## Initial Acceptance Criteria

The initial design is acceptable for implementation planning when focused fixtures and repeatable CLI checks can demonstrate that:

- forward documentation-folder indexes and code-folder reverse indexes have separate types, inputs, and managed sections;
- reverse output is derived from explicit documentation references and is byte-stable for identical inputs;
- folder, file, symbol, direct, and descendant coverage are reported independently;
- undocumented direct code files appear as visible gaps;
- multiple documents for one target are retained, while stale, missing, ambiguous, and unsupported references remain visible diagnostics;
- file-level coverage works without a symbol adapter, and unsupported symbol references do not silently widen their scope;
- unambiguous moves can produce reviewable candidates, while deletions and ambiguous renames do not create guessed edges;
- ordinary Markdown links, typed-graph facts, and adapter diagnostics retain their separate meanings; and
- static CLI build, check, and fix reproduce the same result without a daemon or LLM.

## Open Decisions

- The final Markdown-compatible syntax for folder, file, symbol, and descendant selectors.
- Where reverse-index files or managed sections live, and whether code roots receive one projection or several nested projections.
- The default code-root, exclusion, and eligible-direct-file rules.
- The exact stable comparator and identity rule for duplicate references.
- Whether diagnostics for unresolved references are shown in each affected reverse index, a separate report, or both.
- How much historical Git evidence is required before presenting a move or rename candidate.
- Whether symbol-level reverse entries are colocated with file entries or exposed as a separate managed projection.
- The machine-readable export shape and limits for reverse-index diagnostics and coverage facts.
