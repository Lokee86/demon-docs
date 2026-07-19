# Stateless Document Refactoring

`ddocs mv` moves one repository-contained file or directory and rewrites affected local link destinations. It is deliberately usable before `ddocs init`: the command does not require, create, or update `.ddocs/` state.

## Command

```bash
ddocs mv [--root PATH] [--dry-run] SOURCE DESTINATION
```

`SOURCE` and `DESTINATION` resolve from the current working directory. The repository boundary defaults to the nearest initialized Demon Docs repository root when one exists, otherwise to the current directory. `--root` provides an explicit boundary.

If `DESTINATION` is an existing directory, the source is moved beneath it using its current basename. The destination parent must already exist.

## Planning model

The command plans the complete operation before changing the filesystem:

1. validate that source and destination remain inside the repository boundary;
2. inventory repository files using the normal permanent ignores and `.docignore` policy;
3. parse recognized links in repository Markdown;
4. resolve each target against the pre-move filesystem;
5. map moved source and target paths to their post-move locations;
6. calculate the smallest destination-path rewrites needed to preserve each affected link; and
7. report the move, updated Markdown files, and rewritten-link count.

`--dry-run` stops after planning.

## Supported links

The move planner reuses the link parser and renderer used by normal reconciliation. It preserves authored labels, titles, aliases, queries, fragments, angle wrapping, path separators, URL escaping, and surrounding prose for:

- inline Markdown links and images;
- reference definitions;
- path-based wiki links and embeds; and
- supported local HTML `href`, `src`, and `poster` targets.

Moving a Markdown source can require rewrites even when its targets do not move, because relative paths are recalculated from the source's new location. Moving a directory applies the same mapping to all descendants.

## Safety boundaries

- No `.ddocs/` repository is required or created.
- Source and destination must remain inside the selected repository boundary.
- Symbolic-link move sources are rejected.
- Existing non-directory destinations are not overwritten.
- Destination parents are not created implicitly.
- Broken unaffected links remain untouched.
- An affected wiki target that resolves to multiple candidates is rejected rather than guessed.
- Every rewrite source is content-hash checked immediately before the move.
- Markdown files use atomic per-file replacement.
- A failed rewrite triggers best-effort restoration of original Markdown and the original filesystem location.

In an initialized repository, the watcher or the next link reconciliation pass refreshes persistent identity state after the explicit move.

## Code map

- `internal/app/move.go` — CLI parsing, repository-boundary selection, dry-run output, and execution.
- `internal/links/move.go` — stateless inventory scan and move-plan construction.
- `internal/links/move_paths.go` — containment checks, target remapping, and wiki-path disambiguation.
- `internal/links/move_apply.go` — preflight checks, filesystem application, case-only renames, and rollback.
- `internal/links/parser.go` — recognized Markdown destination extraction.
- `internal/links/target.go` — local target resolution and syntax-preserving path rendering.
- `internal/links/rewrite.go` — content-addressed generated rewrites and atomic file replacement.
