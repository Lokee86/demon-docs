---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-7727-9e69-a5563ef1b1a8
document_type: general
policy_exempt: false
summary: This document describes the implemented ignore policy and filesystem traversal boundary. It covers repository-root .docignore, nested .docignore files used by reverse-index traversal, permanent exclusions, path-scope safety, and the...
---
# Ignore and Traversal

Parent index: [Architecture](./README.md)

## Purpose

This document describes the implemented ignore policy and filesystem traversal boundary. It covers repository-root `.docignore`, nested `.docignore` files used by reverse-index traversal, permanent exclusions, path-scope safety, and the consumers that apply these decisions.

## Overview

`internal/ignore/` supplies two related policies:

- `Policy` loads one `.docignore` file at a supplied root and evaluates repository-relative paths.
- `Hierarchy` loads the repository-root file and, when traversal reaches them, nested `.docignore` files whose rules are rooted at their containing directories.

The policy is an input to inventory and event filtering. It does not decide which files a subsystem can index, how generated indexes are rendered, or which watcher-only temporary files are interesting.

## Code root

```text
internal/ignore/
```

## Responsibilities

The ignore boundary owns:

- loading `.docignore` files;
- parsing each line with go-git's Git-ignore pattern parser;
- applying ordered pattern matching and negation;
- applying the permanent directory exclusions;
- converting an absolute path into ignore-domain-relative components; and
- refusing paths or nested policy directories outside the policy root.

`Policy` also exposes the repository-root control-file predicate. `Hierarchy` tracks which policy directories have already been loaded so callers can safely load ancestors and visited directories more than once.

## Does not own

The ignore boundary does not own:

- Git trackedness, `.gitignore`, Git status, or Git history;
- configured file include/exclude patterns;
- index-file or draft-folder semantics;
- link parsing, target repair, or external-target watching;
- reverse-index root selection or code-file eligibility; or
- watcher-only ignored directories, suffixes, and editor temporary-file rules.

Those consumers combine ignore results with their own scope and file-type rules.

## Repository scope and policy domains

An initialized repository uses the repository root as the base ignore root. Its `.docignore` sits beside `.ddocs/`. Calls such as `scan.TreeWithIgnoreRoot(docsRoot, repositoryRoot, c)` therefore scan the configured documentation root while evaluating paths against the repository-root policy. Legacy standalone configurations use the managed root as the ignore root.

The base `Policy` has one Git-ignore domain: every pattern is parsed relative to the policy root. A missing `.docignore` produces an empty matcher and does not exclude ordinary paths. The policy is independent of Git's own ignore files: a Git-tracked path can be excluded by `.docignore`, and a Git-ignored path can remain visible to Demon Docs.

`Hierarchy` begins with the repository-root domain. As a reverse-index walk accepts a directory, it loads that directory's `.docignore`, if present, with the directory as the pattern domain. A pattern in `services/api/.docignore` therefore describes paths below `services/api`, not paths relative to the repository root. Ancestor files are loaded from the repository root through the relevant directory before a reverse root is evaluated.

Patterns use the Git-ignore syntax provided by go-git, including comments, anchored paths, wildcards, directory patterns, and `!` negation. Within the ordered matcher, a later matching rule determines the result. Nested patterns are appended after the already-loaded ancestor patterns, so a matching rule in a deeper loaded domain can override an earlier matching ancestor rule. This is precedence of ignore rules, not a new independent policy: all loaded domains contribute to the same hierarchy matcher.

Only reverse-index traversal currently loads nested `.docignore` files. The documentation scanner, link inventory, ordinary watcher, codemap corpus, and codemap dataset use a root `Policy`; a nested file in those paths is not automatically a new domain.

## Permanent exclusions

The following directory names are ignored at every depth:

```text
.git/
.ddocs/
.obsidian/
logseq/
```

The check is performed before the Git-ignore matcher. A `!` rule cannot re-include these directories or anything below them. This also means that a policy cannot expose `.ddocs/` private Demon Docs state, a Git worktree/control directory, Obsidian metadata, or Logseq metadata by negation.

On Windows, these four permanent directory names are compared with `strings.EqualFold`, so case variants such as `.GIT` and `LogSeq` remain permanently excluded. This case folding is limited to the permanent-directory check. The implementation does not normalize all `.docignore` patterns or all control-file names to case-insensitive form.

## Path safety and out-of-root behavior

Before matching, `Policy` and `Hierarchy` convert the candidate to a cleaned path relative to their root. A candidate outside that root is an error, including a path that escapes through `..`. Loading a nested policy directory outside the hierarchy root is rejected in the same way. The root itself is represented by an empty relative component list and is not permanently ignored.

This is a boundary for ignore evaluation, not a prohibition on every external filesystem reference. The link inventory applies the policy only when a target is inside the repository; explicit external link targets remain outside the repository `.docignore` domain and are handled by the link subsystem's external-target rules.

Reverse-index root selection has additional scope checks in `internal/reverseindex/`: roots must be existing directories inside the repository, cannot be the repository root, cannot overlap the docs root, cannot be inside a worktree control directory, and cannot already be ignored.

## Re-inclusion and pruning rules

A normal path can be re-included when a later matching `!` rule changes its matcher result. The following limits are important:

- Permanent directories and their descendants cannot be re-included at all.
- A negation cannot make a directory traversable after a consumer has already pruned that directory. If an ignored directory is returned as `SkipDir`, no descendant `.docignore` is loaded and no later descendant rule is evaluated by that walk.
- A nested rule can re-include a path excluded by an ancestor rule only when traversal reaches the containing directory and loads the nested policy. A nested `.docignore` inside a directory that was pruned by an earlier directory match is never available to rescue that directory.
- A re-included path still has to pass the consumer's own filters. For example, a scanner re-included by `.docignore` must also match configured file include patterns and must not match configured exclude patterns; a reverse-index target must remain inside a selected root and be an eligible code file.

The direct `Policy.Ignored` and `Hierarchy.Ignored` methods evaluate the supplied path; they do not themselves walk parents or create missing directories. The traversal caller determines whether a false result permits descent.

## Control-file treatment

`.docignore` is a control-plane input, not a permanent exclusion. `Ignored` does not special-case the control file, so an ignore pattern can match it as an ordinary path. Control-file detection is separate:

- `Policy.IsControlFile` recognizes exactly the repository-root `.docignore`.
- `Hierarchy.IsControlFile` recognizes a path whose basename is `.docignore`; reverse-index traversal also explicitly omits `.docignore` from generated code-folder file inventories.
- The ordinary watcher treats the repository-root control file as relevant even when normal include patterns would not select it, reloads the root policy, and rescans its watched tree.
- The reverse-index watcher treats any `.docignore` event in its selected scope or ancestors as a reason to refresh the nested hierarchy and watched directories.

The forward scanner does not index `.docignore` merely because it exists: after ignore evaluation, `scan.IsIndexable` still applies the configured index filename and file include/exclude rules. The link inventory and other repository-wide inventories apply their own file-inventory rules rather than treating the control file as authored Markdown.

## Consumer boundaries

### Documentation scanner

`internal/scan/scan.go` recursively builds the documentation folder tree with a root `Policy`. It checks each directory before recording or descending into it; an ignored directory is pruned. It separately applies the configured include/exclude globs, skips symlinks, omits the configured index file, and gives the configured draft folder its own traversal behavior. The scanner owns documentation-tree shape; `.docignore` only decides whether a candidate is traversable or visible to that scan.

### Link inventory and moves

`internal/links/inventory.go` walks the repository with a root `Policy`, skips ignored directories with `filepath.SkipDir`, and excludes ignored files from the repository target inventory. Its `ignored` helper intentionally returns no repository-policy result for paths outside the repository, allowing external link targets to remain an independent link concern. `internal/links/move.go` uses the same inventory policy to refuse an ignored move source or destination and to avoid rewriting links whose resolved targets are ignored.

The links boundary owns link syntax, target resolution, identity state, and source-preserving rewrites. It does not reinterpret a negated or ignored path as a link repair decision.

### Reverse-index traversal

`internal/reverseindex/scope.go`, `traversal.go`, `inventory.go`, and `targets.go` use `Hierarchy`. They select and validate configured roots, load ancestor policies, prune ignored directories, load a directory's nested `.docignore` after accepting that directory, and ignore excluded files and resolved targets. A selected root that is already ignored is rejected rather than silently producing an empty projection. Reverse-index code then applies its own code-file extension/name eligibility and generated-index ownership rules.

`internal/reverseindex/watch.go` watches selected roots and their ancestors. `.docignore` events request hierarchy refresh; newly visible directories are added to the watch set before the next reconciliation. Reverse-index traversal and watching therefore share nested-policy behavior that the ordinary documentation/link watcher does not provide.

### Ordinary watcher

`internal/watch/watch.go` loads a root `Policy`. Its initial `addTree` walk prunes permanently or explicitly ignored directories. Event relevance combines the policy with the selected docs/repository scope and watcher-only configuration: `[watch].ignored_dirs`, ignored suffixes, and `.#` editor files. Those watcher-only filters are not repository-wide ignore rules.

A root `.docignore` event is always relevant and reloads the policy before the watched tree is refreshed. When a directory is created, the watcher adds it only if the root policy and watcher-only filters allow it. Documentation-only watching remains scoped to the docs root; link-enabled watching observes the repository root and also manages external-link parent-directory watches.

## Runtime flow

```text
load repository-root Policy
    -> select a consumer scope
    -> evaluate each candidate with permanent exclusions and Git-ignore rules
    -> prune ignored directories or omit ignored files
    -> apply consumer-specific include, type, and ownership rules
    -> on control-file change, reload policy and refresh affected watches
```

Reverse-index traversal replaces the single-policy load with:

```text
load root Hierarchy
    -> validate selected root against loaded ancestors
    -> walk accepted directories
    -> load each directory's nested .docignore
    -> evaluate children with accumulated domains
    -> refresh hierarchy and watches when a nested control file changes
```

## Safety invariants

- Ignore evaluation never authorizes a path outside its policy root.
- Permanent exclusions are stronger than all `.docignore` negations.
- A consumer must prune ignored directories if it wants descendant rules to be unavailable after pruning.
- `.gitignore` state and Git trackedness never substitute for `.docignore`.
- A path becoming visible through negation does not bypass the consumer's own scope, type, or write-ownership checks.
- Watcher-only exclusions do not silently become static repository exclusions.
- Ignore policy is selection input; it is not a source of semantic documentation relationships.

## Code map

- `internal/ignore/ignore.go` — root `Policy`, Git-ignore loading, permanent exclusions, control-file detection, and root containment.
- `internal/ignore/hierarchy.go` — repository-root and nested `.docignore` domains for reverse-index traversal.
- `internal/scan/scan.go` — documentation-tree traversal and consumer-specific indexability.
- `internal/links/inventory.go` — repository target inventory filtered by the root policy.
- `internal/links/move.go` — ignored source, destination, and target checks during stateless moves.
- `internal/reverseindex/scope.go` — reverse-root validation and ancestor policy loading.
- `internal/reverseindex/traversal.go` — nested-policy traversal and directory pruning.
- `internal/reverseindex/inventory.go` — code-folder inventory filtering.
- `internal/reverseindex/watch.go` — reverse watch refreshes for nested control files.
- `internal/watch/watch.go` — ordinary watcher filtering, root policy reload, and watch-tree pruning.

## Tests

Focused tests cover permanent exclusions, Git-ignore matching and negation, nested policy domains, deeper negation, repository-owned scanner scope, link-target exclusion, reverse nested `.docignore`, reverse watch refresh, and ordinary watcher control-file handling.

```bash
go test ./internal/ignore ./internal/scan ./internal/links ./internal/reverseindex ./internal/watch -count=1
```

## Related docs

- [Architecture](README.md)
- [Configuration Reference](../reference/configuration.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Reconciliation Model](reconciliation-pipeline.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Code-Folder Reverse Indexes](reverse-indexes.md)
- [Adopting Reverse Indexes](../guides/reverse-indexes.md)
- [Repository Scope and Worktrees](repository-scope-and-worktrees.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)

## Notes

Nested `.docignore` domains are currently a reverse-index traversal capability. Other scanners continue using one root policy unless their owning implementation explicitly adopts hierarchical loading.
