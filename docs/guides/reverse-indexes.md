---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-705a-914f-a3f7de68db27
document_type: general
policy_exempt: false
summary: This guide configures and verifies code-folder reverse indexes so code directories show which authored documentation pages explicitly reference them.
---
# Adopting Reverse Indexes

Parent index: [Guides](./README.md)

## Purpose

This guide configures and verifies code-folder reverse indexes so code directories show which authored documentation pages explicitly reference them.

## Overview

Reverse indexes project existing codemap targets back into selected code folders. They do not infer ownership. A documentation page must contain an explicit target under a configured codemap heading before it can appear as a backlink.

Use `check --reverse` to preview required generated changes, `fix --reverse` to apply them, and `watch --reverse` only as optional local automation.

## Prerequisites

- Demon Docs is initialized for the repository.
- The configured documentation root exists.
- At least one documentation page contains a configured codemap section with file or folder targets.
- The intended code roots are inside the repository and outside the documentation root.
- Generated code-folder index files are acceptable in the selected roots.

## Configure codemap headings

Set the headings that contain authored code targets:

```toml
[codemap]
headings = ["Code map", "Implementation map"]
```

A heading match is required. Reverse reconciliation fails rather than treating arbitrary prose links as code ownership.

## Configure reverse roots

Add repository-relative code roots:

```toml
[reverse_index]
roots = ["cmd", "internal"]
```

Roots are recursive. Nested `.docignore` files can exclude generated, vendor, fixture, or private subtrees.

The older compatibility key remains accepted:

```toml
[reverse_index]
folders = ["cmd", "internal"]
```

Prefer `roots` in new configuration.

## Preview the projection

```bash
ddocs check --reverse
```

A non-zero result may mean index files need creation or updating. It may also report configuration, scope, codemap-section, marker, or target-resolution failures.

To test a different scope without editing configuration:

```bash
ddocs check --reverse \
  --reverse-root internal/links \
  --reverse-root internal/review
```

Repeated command-line roots replace configured roots for that invocation.

## Apply and verify

```bash
ddocs fix --reverse
ddocs check --reverse
```

Review every created or changed code-folder index. Demon Docs owns only the `reverse-index` marker block. Existing prose outside that block should remain unchanged.

A generated block lists direct files and nests documentation backlinks below exact file targets. Folder-level documentation appears separately as folder documentation.

## Run with other subsystems

Without selectors, normal `fix` and `check` run documentation indexes and links, plus reverse indexes when reverse roots are configured.

Use explicit selectors for a narrow pass:

```bash
ddocs check --docs --links --reverse
ddocs fix --reverse
```

When any selector is supplied, only selected systems run.

## Use foreground automation

```bash
ddocs watch --reverse
```

Or run all enabled systems:

```bash
ddocs watch
```

Watch mode is a convenience. A later static `ddocs check --reverse` must reproduce the clean result without a running watcher.

## Expected result

- Selected code folders contain deterministic managed reverse-index blocks.
- Explicit folder and file targets link back to their source documentation.
- Eligible direct code files remain visible even when no document targets them.
- Unresolved authored targets remain diagnostics rather than guessed backlinks.
- A second `fix --reverse` changes no files.
- `check --reverse` succeeds.

## Failure and recovery

### No reverse-index roots selected

Add `[reverse_index].roots` or pass at least one `--reverse-root`.

### No codemap section found

Confirm `[codemap].headings` matches an actual heading under the configured docs root. Do not broaden headings merely to capture unrelated prose.

### Codemap section contains no targets

Add explicit file or folder targets, or remove reverse indexing from the repository until authored targets exist.

### A target is unresolved

Correct the authored target. Reverse indexes do not select among missing or ambiguous destinations.

### A root is rejected

Keep roots inside the repository, outside the docs root, outside permanently ignored directories, and outside nested Git worktrees.

### An index has incomplete markers

Repair the marker pair manually after reviewing the file. Demon Docs will not guess which authored content belongs inside a damaged generated region.

## Related docs

- [Reverse Index Architecture](../architecture/reverse-indexes.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Configuration Reference](../reference/configuration.md)
- [CLI Reference](../reference/cli.md)
- [Ignore and Traversal](../architecture/ignore-and-traversal.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Reverse indexes are file/folder projections today. Symbol-level backlinks and dependency-aware projections belong to the planned code-intelligence track.
