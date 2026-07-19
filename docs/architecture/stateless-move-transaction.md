# Stateless Move Transaction

Parent index: [Architecture](./README.md)

## Purpose

This document describes the implemented architecture of `ddocs mv`: its stateless planning and application boundary, repository containment rules, Markdown path-rewrite behavior, filesystem ordering, rollback behavior, and verification seams.

## Overview

`ddocs mv` performs one explicit file or directory move and rewrites recognized repository-local links that would otherwise change meaning. It plans against the pre-move filesystem, applies the filesystem rename, and then applies generated Markdown rewrites. The command does not require an initialized Demon Docs repository and does not create or update `.ddocs/` state.

The move boundary is intentionally separate from persistent link reconciliation. It does not infer an already-completed move from historical identity state. It receives the intended source and destination, resolves affected links before mutation, and applies only the path changes required by that explicit move.

## Code root

The primary implementation is in `internal/links/move*.go`, with command orchestration in `internal/app/move.go`.

## Responsibilities

The stateless move boundary owns:

- resolving the selected repository boundary and the two move paths;
- validating source, destination, containment, ignore policy, and move shape;
- scanning the current Markdown inventory before mutation;
- resolving affected local targets using the pre-move filesystem;
- mapping paths under a moved file or directory to their post-move paths;
- rewriting incoming links from Markdown sources that remain in place;
- rewriting relative local links inside Markdown sources that are themselves moved;
- preserving recognized link syntax while changing destination paths;
- planning deterministic generated rewrites with expected source hashes;
- handling bare wiki-link ambiguity caused by the move;
- applying the filesystem move and generated rewrites in a fixed order; and
- attempting best-effort rollback when generated rewrites fail after the move.

The public plan reports the resolved source and destination, whether the source is a directory, the Markdown files that will change, and the number of links to rewrite.

## Does not own

This boundary does not own:

- persistent `.ddocs/` identities, path history, incoming-link state, or review events;
- discovery of a move that has already happened;
- Git commits, staging, or general working-tree rollback;
- creation of destination parent directories;
- symbolic-link source moves;
- rewriting non-Markdown source files;
- changes to Markdown prose, labels, titles, aliases, fragments, queries, or unrelated destinations;
- semantic validation of headings or anchors;
- selection among ambiguous wiki targets when the current move cannot establish one deterministic target; or
- external target contents or files outside the selected repository boundary.

The common generated-rewrite boundary owns source-preserving atomic replacement. The move layer supplies the planned path and expected old content; it does not implement a second write protocol.

## Repository boundary resolution

`runMove` starts with the current working directory. If repository discovery finds an initialized Demon Docs repository, its root becomes the default boundary. Otherwise the current directory is the boundary. `--root PATH` overrides discovery; a relative override is resolved from the current directory.

The CLI resolves relative `SOURCE` and `DESTINATION` arguments from the current directory. `PlanMove` normalizes the repository root and both paths to absolute, clean paths. Direct callers may also provide relative paths, which are resolved from the supplied repository root.

Containment is checked in two stages:

1. The cleaned source and destination must be lexically inside the selected root.
2. The real root and source are resolved with `filepath.EvalSymlinks`, and the destination parent is resolved the same way. The real source and destination parent must remain inside the real repository root.

This prevents a path that appears inside the boundary from escaping through a symlinked source or destination parent. The source itself must not be a symbolic link. The destination may be reached through an existing parent only when that parent resolves inside the boundary.

## Validation

Planning rejects the move before any mutation when:

- the source cannot be `Lstat`ed;
- the source is a symbolic link;
- the source is the repository root;
- the destination is the same cleaned path as the source;
- a directory destination is inside the source directory;
- the destination parent does not exist or is not a directory;
- lexical or real-path containment leaves the repository boundary;
- the source or final destination is excluded by repository ignore policy;
- a non-directory destination already exists; or
- an affected wiki target is ambiguous in a way that includes the moved source.

If the requested destination is an existing directory, the final destination is that directory joined with the source basename. The final destination is then checked for collisions. A case-only rename is the exception to the ordinary same-path and destination-exists checks: a path with the same case-folded key but different cleaned spelling is treated as a rename rather than an overwrite.

The move does not create missing parents. A caller must create the destination parent before planning.

## Plan construction and path mapping

`PlanMove` builds a current filesystem inventory and enumerates repository Markdown sources subject to the repository traversal and ignore policy. Each source is read and parsed before the move. For every recognized link, the planner:

1. resolves the link's local path relative to the current Markdown source and preserves its syntax/style information;
2. leaves non-local links and ignored targets outside the move graph;
3. resolves the actual target using the source form and the pre-move inventory;
4. maps the Markdown source path through the move when the source is inside the moved path;
5. maps the resolved target through the move when the target is inside the moved path; and
6. renders a replacement relative to the final Markdown source while preserving the link form.

`remapMovedPath` is the common mapping rule. A path outside the source subtree is unchanged. A path inside it is converted to a relative path from the moved source root and joined below the final destination. Mapping the source and target independently is what handles both incoming links and links inside moved Markdown sources.

A moved file and a moved directory use the same rule. For a directory, the directory itself maps to the final destination and every descendant retains its relative suffix. Thus a link to `docs/guide/page.md` becomes a link to `docs/archive/guide/page.md` when `docs/guide` is moved to `docs/archive`.

The planner skips links whose source and target remain unchanged after mapping. It also avoids turning an empty destination into a self-link when the resolved target is the current Markdown source. Planned updates are sorted by final path, and generated rewrites are sorted by their final rewrite path, so plan output does not depend on filesystem enumeration order.

## Incoming-link rewrites

Markdown sources outside the moved subtree remain at their original paths. If one of their recognized local links resolves to the moved source or to a descendant of a moved directory, the target is remapped to its post-move location and the source receives a generated rewrite.

The rewrite changes only the destination path. Inline links, images, reference definitions, path-based wiki links, and supported local HTML targets use the existing parser and renderer conventions. Labels, titles, aliases, embed markers, query strings, fragments, angle wrapping, path-separator style, escaping, surrounding prose, newline style, and final-newline state are preserved by the generated rewrite boundary.

Broken links that are unrelated to the explicit move are not repaired or normalized. An ignored target is outside this move graph and is not rewritten.

## Relative links inside moved Markdown sources

A Markdown source inside the moved subtree is read at its original location, but its generated rewrite is addressed at its final location. This matters even when the linked target does not move: the relative path must be rendered from the Markdown source's new directory.

For example, moving `docs/topic.md` to `docs/archive/topic.md` changes a relative link to `../README.md` into `../../README.md`. A fragment-only link such as `#section` remains unchanged because it is local to the same Markdown source and does not name a filesystem path.

For a directory move, links from a moved descendant are mapped from the descendant's final path. A link to another descendant is rewritten according to the mapped target; a link to an external or otherwise unmoved repository path is rendered relative to the moved source's new location.

## Wiki ambiguity handling

Wiki resolution is syntax-aware. A path-specific wiki target can resolve directly, while an extensionless bare wiki target may use a unique Markdown basename candidate. The planner does not invent a target when the pre-move candidates are ambiguous.

There are two move-specific cases:

- If an affected bare wiki target has multiple candidates and one candidate is inside the moved source, planning fails with an ambiguous-wiki error. The move cannot safely determine which candidate the author meant.
- If the target is resolved uniquely before the move but the move would make a bare wiki spelling ambiguous afterward, the renderer makes the path explicit while preserving wiki syntax. The `.md` suffix is removed from the rendered path for the bare wiki form, producing a path such as `[[archive/guide]]` rather than leaving the now-ambiguous `[[guide]]`.

An ambiguity that does not involve the moved source and does not affect a link's resolved target is left outside the move plan. The command neither guesses among candidates nor rewrites unrelated wiki links.

## Case-only renames

Case-only renames are supported when the source and destination have the same path key but different cleaned spelling. Planning allows the operation and skips the ordinary destination-exists rejection for that case.

`renameMovePath` performs the filesystem step through a temporary sibling name: it creates and removes a temporary directory to reserve a unique temporary path, renames the source to that temporary path, then renames the temporary path to the requested destination. If the second rename fails, it makes a best-effort attempt to restore the temporary path to the original source path.

The normal link planner still evaluates links for the spelling change. A link such as `[Guide](Guide.md)` is rewritten to `[Guide](guide.md)` when the source is renamed accordingly.

## Source-hash preflight

A plan records an expected SHA-256 digest for every Markdown source that will receive a generated rewrite. `ApplyMove` runs `preflightMove` before renaming anything. Preflight verifies:

- the move source still exists, is not a symbolic link, and still has the planned file/directory kind;
- the non-case-only destination is still absent; and
- every planned rewrite origin still has the expected SHA-256 digest.

A changed Markdown source aborts the operation before the filesystem move. The error includes the expected and actual digest. This prevents the move from applying a stale rewrite over a concurrent authored edit.

Preflight is intentionally performed for the complete plan before the first mutation. It is not a per-file check performed after some rewrites have already been applied.

## Filesystem and rewrite ordering

The application sequence is:

```text
plan against the pre-move filesystem
-> preflight source kind, destination availability, and all old hashes
-> rename SOURCE to DESTINATION
-> apply all generated Markdown rewrites
-> report success
```

The filesystem move precedes Markdown rewriting because the moved source must exist at its final path before its generated rewrite is installed. The generated rewrite list is copied from the plan only after preflight succeeds and is passed to the common generated-write application boundary.

If the filesystem rename itself fails, no generated rewrite is attempted. If generated rewriting fails after the rename, rollback is attempted.

## Atomic source replacement

Generated Markdown writes use the shared generated-rewrite boundary. Each rewrite carries the expected old content and the new content, and the write path validates the expected source content before replacement. The replacement is performed through a same-directory temporary file and atomic replacement rather than writing directly over the authored source.

This boundary is used both for normal move rewrites and for restoring old content during rollback. The move transaction therefore relies on per-source atomic replacement, not on a repository-wide filesystem transaction.

## Rollback ordering

When `ApplyGenerated` reports a rewrite failure, `rollbackMove` restores in this order:

1. For each planned rewrite, inspect its final rewrite path. If the path exists, replace it with the recorded old bytes and original mode. If it does not exist, skip it; a missing path is treated as a rewrite that may not have been installed.
2. Rename the final destination back to the original source using the same rename helper, including the temporary sibling step for a case-only rename.
3. Aggregate any restoration failures and return them together.

Restoring rewritten bytes before moving the source back also covers Markdown files that were inside the moved subtree: their old content is restored at the post-move path, then the whole source subtree is moved back.

Rollback is best effort, not a filesystem transaction. It can fail because a rewrite path, destination, parent, permissions, or another external condition changed after the forward move. A successful rollback returns an error stating that the move was rolled back. If rollback itself fails, the error reports both the rewrite failure and the rollback failure details. Git remains the authoritative recovery mechanism for authored files.

## Statelessness

The move planner uses a fresh filesystem inventory and `FilesManifest{}` rather than persistent link identity state. Neither dry-run nor apply creates `.ddocs/`; the move command does not publish identities, incoming-link groups, fingerprints, review events, or applied-change history.

Statelessness does not mean read-only: a non-dry-run changes the requested filesystem location and affected Markdown sources. It means the command has no durable Demon Docs transaction to commit or recover. In an initialized repository, a later watcher or link-enabled reconciliation pass can refresh persistent state after the explicit move.

## Invariants and safety boundaries

The implementation maintains these invariants:

- The source and final destination remain inside the selected repository boundary, including real-path containment through existing symlinked parents.
- The repository root cannot be moved, and a directory cannot be moved into itself.
- A symbolic-link source is never moved.
- Missing destination parents are not synthesized.
- Existing non-directory destinations are never overwritten.
- All affected Markdown sources are validated against their planned content before the move starts.
- Only recognized local link destinations and the relative rendering context of moved Markdown sources are changed.
- Unrelated prose, labels, fragments, queries, and broken unaffected links remain unchanged.
- Ambiguous affected wiki targets are not guessed.
- Plans and reported updates have deterministic path order.
- A failed preflight causes no filesystem or source rewrite mutation.
- A failed post-move rewrite triggers best-effort content restoration before location restoration.
- No `.ddocs/` state is created or updated by this command.

## Failure modes

The CLI returns usage errors when `SOURCE` and `DESTINATION` are not both supplied. Planning errors cover missing sources, invalid containment, symlink sources, root/self moves, source-directory containment, ignored paths, missing or invalid destination parents, existing destination collisions, and affected wiki ambiguity.

Application errors occur when the source changes kind, the destination appears after planning, a planned Markdown source's hash changes, the filesystem rename fails, or generated rewriting fails. Hash and destination preflight failures occur before the source is moved. Rename failures occur before generated rewrites. Rewrite failures may leave a partial forward application, so the command reports whether best-effort rollback succeeded or failed.

The command does not silently recover from a failed rollback. The reported error and Git status are the recovery evidence for the caller.

## Extension seams

The implementation has explicit seams for future maintenance without changing the transaction shape:

- `PlanMove` is the planning boundary and returns the inspectable `MovePlan`.
- `MovePlan` and `MoveUpdate` are the CLI-facing summary model; private `plannedMoveRewrite` carries origin path, mode, and generated-rewrite data.
- `remapMovedPath` is the generic subtree mapping rule for files and directories.
- `resolveMoveTarget` isolates syntax-specific target resolution and wiki candidate handling.
- `renderMoveTarget` isolates syntax-preserving output and post-move bare-wiki disambiguation.
- `renameMovePath` isolates ordinary and case-only filesystem rename behavior.
- `preflightMove` isolates stale-plan checks.
- `rollbackMove` isolates best-effort restoration ordering.
- `ApplyGenerated` remains the shared source-hash and atomic-replacement seam for generated Markdown writes.

A new recognized link form should integrate through the existing parser, local-target resolver, target renderer, and generated rewrite path; the move transaction should not grow a separate syntax implementation. Persistent identity or review behavior should remain in the stateful reconciliation boundary rather than being added to `ddocs mv`.

## Verification

The implementation is verified by focused link and CLI tests. The relevant focused command is:

```bash
go test ./internal/links ./internal/app -count=1
```

The tests cover recognized inline, image, reference-definition, wiki, and HTML rewrites; stateless operation without initialization; dry-run behavior; relative links inside moved Markdown; incoming links and moved-directory descendants; bare wiki disambiguation; affected ambiguous wiki rejection; symlinked destination escape; case-only renames; changed-source hash preflight; ignored sources; outside-boundary destinations; and CLI argument validation.

A full repository verification should use the project’s normal gate after the focused tests:

```bash
make release-check
```

## Tests

Focused implementation coverage:

- `internal/links/move_test.go` — planning and application behavior, path mapping, syntax coverage, wiki ambiguity, containment, case-only rename, hash preflight, ignore policy, and statelessness.
- `internal/app/move_test.go` — CLI discovery/default boundary behavior, dry-run output and non-mutation, applied output, statelessness, and argument errors.

The tests exercise the transaction through temporary repositories and inspect both filesystem locations and exact rewritten Markdown content.

## Code map

Primary implementation:

- `internal/app/move.go` — `ddocs mv` help, flags, root discovery/override, argument path resolution, dry-run selection, apply invocation, and summary output.
- `internal/links/move.go` — `MovePlan`, `MoveUpdate`, pre-move inventory scan, link resolution, path mapping, generated rewrite planning, and deterministic sorting.
- `internal/links/move_paths.go` — real-path containment, destination normalization, target resolution, wiki ambiguity rules, absolute path conversion, and subtree remapping.
- `internal/links/move_apply.go` — stale-plan preflight, ordinary/case-only rename, generated rewrite application ordering, and rollback.

Related implementation seams used by the move files:

- the repository inventory and ignore policy used by `PlanMove`;
- the Markdown parser and target renderer used to preserve recognized link syntax; and
- the shared generated-rewrite application boundary used for expected-hash validation and atomic source replacement.

Related tests:

- `internal/links/move_test.go`
- `internal/app/move_test.go`

## Related docs

- [Stateless Document Refactoring](../guides/document-refactoring.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Repository Scope and Worktrees](repository-scope-and-worktrees.md)
- [Repository State and Transactions](repository-state-and-transactions.md)
- [Application Orchestration](application-orchestration.md)
- [CLI Reference](../reference/cli.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

This document describes the explicit move transaction, not historical move detection. The command’s rollback is bounded best effort around a filesystem rename and per-source generated writes; it does not replace Git recovery or publish persistent Demon Docs state.
