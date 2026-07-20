---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7fe6-b7fe-3d23205115ac
document_type: general
policy_exempt: false
summary: This guide moves one repository-contained file or directory and rewrites affected local links with ddocs mv, including use before repository initialization.
---
# Stateless Document Refactoring

Parent index: [Guides](./README.md)

## Purpose

This guide moves one repository-contained file or directory and rewrites affected local links with `ddocs mv`, including use before repository initialization.

## Overview

`ddocs mv` is an explicit, stateless refactoring command. It scans the selected repository boundary, resolves affected links against the pre-move filesystem, calculates the required destination-path rewrites, and applies the move without creating or updating `.ddocs/` state.

Use it when the intended source and destination are already known. Use normal link reconciliation when a move has already happened and persistent identity history should discover the new target.

## Prerequisites

- The source exists inside the selected repository boundary.
- The destination parent already exists.
- The destination does not overwrite an existing non-directory path.
- The working tree is clean or its current changes are intentionally understood.

The boundary is the nearest initialized Demon Docs repository root when one exists. Otherwise it is the current directory. `--root PATH` selects an explicit boundary.

## Preview the move

```bash
ddocs mv --dry-run SOURCE DESTINATION
```

The dry run reports the filesystem move, Markdown files that would change, and the number of affected links. It does not write files or create private state.

Review the complete plan before applying broad directory moves.

## Apply the move

```bash
ddocs mv SOURCE DESTINATION
```

When `DESTINATION` is an existing directory, the source moves beneath it using its current basename.

The command supports:

- files and directories;
- case-only renames;
- inline Markdown links and images;
- reference definitions;
- path-based wiki links and embeds; and
- supported local HTML `href`, `src`, and `poster` targets.

Moving a Markdown source may require rewriting its own relative links even when the targets do not move. Moving a directory applies the same path mapping to descendants.

## Preservation guarantees

Rewrites preserve authored labels, titles, aliases, queries, fragments, angle wrapping, path-separator style, URL escaping, surrounding prose, newline style, and final-newline state.

Broken unaffected links remain untouched.

## Safety and rollback

The command:

- rejects paths outside the selected boundary;
- rejects symbolic-link move sources;
- refuses existing non-directory destinations;
- does not create destination parents implicitly;
- rejects affected ambiguous wiki targets rather than guessing;
- checks every affected Markdown source hash immediately before applying;
- uses same-directory atomic Markdown replacement; and
- attempts best-effort restoration of original content and location if a rewrite fails.

In an initialized repository, the watcher or next link-enabled reconciliation refreshes persistent identity state after the explicit move.

## Expected result

The source appears at the requested destination, affected links still resolve, unrelated links and prose remain byte-stable, and no `.ddocs/` state is created solely by the move command.

## Failure and recovery

### Destination parent is missing

Create the intended parent explicitly, then rerun the dry run. The command does not infer directory structure.

### An affected wiki target is ambiguous

Resolve the wiki link to a path-specific authored target or narrow the repository naming ambiguity before retrying.

### A source changed after planning

The hash preflight stops the operation. Review the concurrent edit and rerun the dry run from current state.

### Apply fails after the filesystem move

Inspect the reported rollback result and Git status. The command attempts restoration, but Git remains the authoritative recovery mechanism for authored files.

## Related docs

- [CLI Reference](../reference/cli.md)
- [Stateless Move Transaction](../architecture/stateless-move-transaction.md)
- [Markdown Link Reconciliation](../architecture/markdown-link-reconciliation.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

`ddocs mv` is intentionally explicit and stateless. It is not a general historical rename detector and does not replace persistent link identity for already-completed moves.
