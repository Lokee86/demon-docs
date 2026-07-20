---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7fd7-9f86-3ef6e942ed96
document_type: general
policy_exempt: false
summary: 'This document describes the implemented repository boundary in internal/repository/: how Demon Docs discovers initialized repositories, resolves documentation scope, protects docs-root containment, and handles linked Git worktrees for...'
---
# Repository Scope and Worktrees

Parent index: [Architecture](./INDEX.md)

## Purpose

This document describes the implemented repository boundary in `internal/repository/`: how Demon Docs discovers initialized repositories, resolves documentation scope, protects docs-root containment, and handles linked Git worktrees for the repository demon.

## Overview

Demon Docs treats the ordinary filesystem as the repository-discovery authority. An initialized repository is identified by a non-directory `.ddocs/config.toml` path found while walking from the requested path toward its ancestors. Scope resolution then turns the selected configuration and root setting into a repository root, docs root, configuration path, and `.docignore` path.

Linked Git worktrees are the one narrow exception to the Git-independent discovery model. The worktree adapter can read Git's linked-worktree metadata to identify a primary initialized worktree and, only on a mutating demon entry, create independent local `.ddocs/` state for the linked worktree. It does not make Git the general source of Demon Docs repository truth.

## Code root

```text
internal/repository/
  repository.go
  scope.go
  worktree.go
```

## Responsibilities

The repository boundary owns:

- ancestor-based discovery of initialized `.ddocs/config.toml` repositories;
- detection of a `.ddocs` marker for initialization and configuration-boundary checks;
- deriving a repository root from a config path with the expected `.ddocs/config.toml` shape;
- initialization of `.ddocs/config.toml` and fresh private object storage;
- conversion of configured or overridden docs-root values into absolute paths;
- lexical and, when resolvable, symlink-aware containment checks;
- construction of the command-scoped `Scope` including `.docignore`;
- read-only linked-worktree detection; and
- first-mutating-entry bootstrap of local linked-worktree configuration and object storage.

The returned `Scope` is the boundary consumed by application orchestration and reconciliation. It identifies where documentation files, repository-local ignore rules, and private Demon Docs state are selected for the current operation.

## Does not own

The repository boundary does not own:

- TOML parsing, configuration precedence, or legacy/global config selection;
- documentation scanning, managed index planning, Markdown parsing, or link reconciliation;
- private object encoding or transactional state mechanics in `internal/ddrepo`;
- demon leases, feeders, owner heartbeats, shutdown, or runtime logs in `internal/demon`;
- Git history, commits, review, checkout, or general Git repository management;
- watcher scheduling in `internal/watch`; or
- the content or semantic ownership of authored documentation files.

Git awareness is intentionally limited to `worktree.go`. Normal `Discover`, `FindMarker`, `Initialize`, and `ResolveScope` operation does not invoke Git or require a Git repository.

## Discovery and configuration markers

`Discover(start)` normalizes the starting path with `filepath.Abs`. If the path names an existing non-directory, discovery begins at its parent. It then walks upward until it finds a non-directory `.ddocs/config.toml`, returning its containing repository root and config path. A `.ddocs` directory by itself is not enough for initialized discovery.

`FindMarker(start)` performs the same ancestor walk but checks only whether `.ddocs` can be found with `os.Stat`. The application uses this broader marker check to reject initialization inside an already marked repository. The configuration package also uses it as the boundary when looking for legacy local configuration, so a legacy config cannot be selected by walking above an initialized Demon Docs repository.

`RootForConfig(configPath)` accepts only the path shape whose final component is `config.toml` and whose parent directory is `.ddocs`. It returns the directory above `.ddocs`; arbitrary config files do not acquire an initialized repository root through this helper.

`Initialize(repoRoot, configText)` creates the repository's `.ddocs` directory, writes `config.toml`, and calls `ddrepo.Init` to create fresh private object storage. It rejects an existing marker, removes partially created state when a later write or object-store initialization fails, and returns the config path. The CLI `ddocs init` performs the additional user-facing checks that the docs root already exists and is inside the current repository before calling this function.

## Scope resolution

`ResolveScope` produces:

```text
Scope{
  RepositoryRoot
  DocsRoot
  ConfigPath
  IgnorePath
  Initialized
}
```

For an initialized config path, `RootForConfig` establishes the repository root and makes it the base for a configured root or an explicit root override. Relative roots therefore resolve from the initialized repository root, including when a command is launched from a nested docs directory. An explicit root override replaces the configured value rather than being layered on top of it.

For a standalone or legacy scope, the config path is not in the exact `.ddocs/config.toml` shape, or no config path is supplied. Relative roots resolve from the working directory, or from the directory containing the supplied non-marker config when no root override is present; a root override uses the working directory as its base. The resolved docs root becomes the standalone scope's repository root, and `Initialized` remains false. This behavior lets stateless or legacy configuration operate without claiming an enclosing initialized repository.

`ConfigPath` is cleaned when present. `IgnorePath` is always `<RepositoryRoot>/.docignore`, not automatically the docs root for initialized scopes. `DocsRootExists` is a separate directory check; scope construction does not require the docs root to exist.

## Containment and symlink safety

`Contains(root, path)` performs absolute lexical containment using `filepath.Rel`. The path is accepted when the relative path is neither `..` nor prefixed by `..` plus a path separator. This rejects configured roots that lexically escape the repository, while allowing the repository root itself.

Initialized scope resolution applies the lexical check first and then calls `validateRealContainment`. That helper evaluates symlinks for both repository and docs roots and rejects the scope if the resolved docs root is outside the resolved repository root. The same real-containment rule is used by `ResolveDocsRoot`, which also returns a repository-relative, slash-normalized root for configuration writing and an absolute path for filesystem use.

If either path cannot be evaluated with `EvalSymlinks`—for example, because the target does not exist yet—the real-path check returns without adding an error. The lexical check still applies. Existence-sensitive callers separately use `DocsRootExists`; `ddocs init` requires an existing directory before initialization, while normal command paths fail when a selected feature needs a missing docs root.

## Linked-worktree detection

`DetectLinkedWorktree(start)` is read-only. It walks ancestors looking for a `.git` file whose first trimmed line begins with `gitdir:` case-insensitively. It resolves that Git directory, reads its `commondir` entry, and derives the primary worktree root from the common Git directory. Detection succeeds only when the primary worktree already contains `.ddocs/config.toml`.

The returned `Location` always uses the linked worktree directory as `Root`. If that worktree already has a `.ddocs/config.toml`, its local config is returned. Otherwise, detection returns the primary worktree's config path as a source for a later bootstrap. A malformed or non-directory linked `.ddocs` marker is reported as an error rather than silently replaced. Detection does not create `.ddocs`, object storage, runtime state, or config files.

This adapter is deliberately narrower than ordinary discovery:

```text
ordinary operation: filesystem .ddocs/config.toml discovery
linked-worktree adapter: read .git file -> gitdir -> commondir -> primary .ddocs/config.toml
```

It does not discover arbitrary Git repositories, inspect commits, or synchronize worktree state.

## Mutating demon bootstrap and local isolation

`BootstrapLinkedWorktree(start)` is used only through demon paths that permit mutation. It first runs read-only linked-worktree detection. If the linked worktree already has its own config, it returns that location without copying or reinitializing it. If no local config exists, it:

1. creates the linked worktree's `.ddocs` directory;
2. reads the primary worktree's `config.toml`;
3. writes the same config text to the linked worktree;
4. initializes fresh linked-worktree object storage with `ddrepo.Init`; and
5. returns the linked config path.

Only configuration text is copied. Runtime owner, feeder, heartbeat, shutdown, and log files are not copied. If fresh object-store initialization fails, the newly prepared marker is removed. A marker that is a non-directory is rejected.

The result is per-worktree isolation:

```text
primary worktree: .ddocs/config.toml + objects + runtime
linked worktree:  .ddocs/config.toml + fresh objects + runtime
```

The configs initially match, but each worktree subsequently reads and mutates its own local config, object store, and demon runtime. `demon.New(location.Root)` therefore addresses the current worktree's `.ddocs/runtime/` rather than the primary worktree's runtime. The worktrees share Git history through Git itself; they do not share mutable Demon Docs state.

The application boundary makes the mutation distinction explicit. `demonLocation` calls ordinary `Discover` first. `ddocs demon run` and the internal `__enter` path allow `BootstrapLinkedWorktree`, because they can create local state. Read-only status and logs use detection without bootstrap. The detached `__serve` and feeder paths use the already selected location and do not independently bootstrap a worktree.

## Flow and lifecycle

Normal application flow is:

```text
config.Select / explicit config
-> repository.Discover when selecting local initialized config
-> config.Load
-> repository.ResolveScope
-> docs-root existence check where the selected feature requires it
-> reconcile, links, reverse indexes, or foreground watch
```

`ddocs init` uses `FindMarker`, `ResolveDocsRoot`, an existing-directory check, and then `Initialize`. `ddocs status` uses `Discover`, loads the config, resolves the scope, and reports the repository root, docs root, config, and `.docignore`.

Repository demon flow is:

```text
mutating demon entry
-> ordinary Discover
-> linked DetectLinkedWorktree
-> BootstrapLinkedWorktree only when needed
-> config.Load for the selected local config
-> demon.New(current worktree root)
-> watcher reconciliation against a scope rooted at that worktree
```

The demon reuses the normal watcher/reconciliation core. Repository scope selection does not itself acquire leases, start processes, or perform document reconciliation.

## State and safety invariants

- An initialized repository is identified by `.ddocs/config.toml`, not by Git metadata.
- Relative roots in initialized scopes cannot escape the repository lexically or through resolvable symlinks.
- A standalone scope owns its resolved docs root and does not infer an enclosing initialized repository.
- Read-only linked-worktree detection does not create or copy Demon Docs state.
- Only mutating demon entry points may bootstrap a linked worktree.
- Each worktree owns its own `.ddocs` config, object storage, and demon runtime state.
- Linked-worktree bootstrap copies config text but never primary runtime state or primary object contents.
- The `.docignore` path belongs to the resolved repository root for the current scope.
- Git remains an adapter for linked-worktree topology, not the general repository authority.

## Code map

Primary implementation:

- `internal/repository/repository.go` - `Location`, `.ddocs` discovery, marker checks, docs-root resolution, initialization, and config-path root derivation.
- `internal/repository/scope.go` - initialized versus standalone scope resolution, lexical containment, symlink-aware containment, and docs-root existence checks.
- `internal/repository/worktree.go` - read-only linked-worktree detection and mutating local bootstrap.

Relevant callers:

- `internal/config/config.go` - local config selection and initialized-repository boundary handling.
- `internal/app/app.go` - `init`, `status`, tree commands, scope resolution, and docs-root checks.
- `internal/app/demon.go` - demon location selection, mutating bootstrap gates, local runtime root selection, and watcher startup.
- `internal/ddrepo/` - private object-store initialization used by repository initialization and worktree bootstrap.

## Tests

Focused repository coverage is in:

- `internal/repository/repository_test.go` - initialization, child-directory discovery, marker lookup, docs-root containment, and config-path shape.
- `internal/repository/scope_test.go` - initialized scope values, repository-relative overrides, repository escapes, and standalone scope ownership.
- `internal/repository/worktree_test.go` - config-only linked-worktree bootstrap, fresh local objects, read-only detection, and the absence of copied runtime state.
- `internal/app/demon_test.go` - application-level demon location and read-only status behavior.

Run the focused package tests with:

```bash
go test ./internal/repository ./internal/app -count=1
```

## Related docs

- [Application Orchestration](application-orchestration.md)
- [Repository State and Transactions](repository-state-and-transactions.md)
- [Using Linked Git Worktrees](../guides/linked-worktrees.md)
- [Repository Demon](../operations/repository-demon.md)
- [Host Adapter Feeder Integration](../operations/host-adapters.md)
- [Configuration Reference](../reference/configuration.md)
- [Ignore and Traversal](ignore-and-traversal.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Documentation Policy](../documentation-policy.md)

## Notes

The repository package provides boundary and location decisions; it does not guarantee that a selected docs root exists. Existence checks and mutation policy remain explicit at the application or demon call site.
