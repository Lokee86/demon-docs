---
author: brian
created: "2026-07-19"
document_id: 2ee44fa0-7e2d-4f2c-93e3-2cc21bf2a663
document_type: general
policy_exempt: false
summary: This guide explains how to inspect, preview, update, verify, decline, and conservatively prune managed codemap sections with the explicit foreground codemap commands.
---
# Managing Codemaps

Parent index: [Guides](./README.md)

## Purpose

This guide explains how to inspect, preview, update, verify, decline, and conservatively prune managed codemap sections with the explicit foreground codemap commands.

## Overview

The production codemap workflow is separate from ordinary documentation reconciliation:

```text
inspect one file
-> review evidence and persisted declines
-> preview exact mutation scope
-> apply the unified codemap update
-> verify convergence
```

Demon Docs adopts the complete configured codemap section as one managed artifact. It retains current links by default, adds deterministic missing links, and reuses shared decline policy so rejected additions remain suppressed while their evidence is unchanged.

The command does not run from normal `ddocs fix`, `ddocs check`, `ddocs watch`, or the repository demon.

## Expected result

After a successful workflow:

- the document contains exactly one configured codemap section;
- the full section body is wrapped by one codemap-specific marker pair;
- existing valid links remain unless an explicit pruning policy selected them;
- qualified missing links are present once;
- unwanted unchanged additions remain suppressed through review policy;
- `codemap check` reports clean; and
- a second `codemap fix` is a no-op.

## Prerequisites

- Demon Docs resolves the intended repository and documentation root.
- Target documents are existing `.md` files beneath that docs root.
- Existing codemap sections use one of the configured headings.
- The repository has no unresolved concurrent edit in the selected documents.
- Review `.docignore` before running against a directory.

Verify scope first:

```bash
ddocs status
ddocs config show
```

The public command updates existing configured sections and resolves the selected document's effective schema when the section is absent. A required schema codemap section is created at its deterministic configured position. A document whose effective schema does not declare a codemap section is skipped rather than receiving an invented placement.

## Configure recognized headings

Repository configuration defines accepted headings:

```toml
[codemap]
headings = ["Code Map", "Implementation Map", "Source Map"]
remove_undiscovered_links = false
remove_low_score_links = false
```

Heading matching is case-insensitive. Heading-like text inside fenced code blocks is ignored.

Use repeated command overrides when testing another convention without changing repository configuration:

```bash
ddocs codemaps inspect \
  --root docs/architecture/runtime.md \
  --heading "Implementation Map" \
  --heading "Source Map"
```

Command-line headings replace the configured list for that invocation.

## Inspect one document

Begin with the narrowest useful scope:

```bash
ddocs codemaps inspect --root docs/architecture/runtime.md
```

Inspection is read-only. It reports:

```text
section status
whether the file would change
proposed additions
score and tier
supporting evidence
persisted decline decisions
configured removals
```

Interpret section status as:

- `existing` — one configured section was found and evaluated;
- `missing` — no configured section exists and the selected effective schema does not require one;
- `schema-created` — the selected effective schema supplied a required placement and the public command created the section.

A `missing` result can still show computed candidates, but no document update is planned without a section placement.

## Preview the write

Run a dry fix after inspection:

```bash
ddocs codemaps fix \
  --root docs/architecture/runtime.md \
  --dry-run
```

Dry-run reports how many files would change and summarizes additions, removals, adoption, and schema creation. It does not write documents or review state.

Review the current file and Git diff before applying when the section contains prose in addition to links. Complete-section adoption intentionally places all existing section-body content inside the managed region.

## Apply one document

Apply the same target without `--dry-run`:

```bash
ddocs codemaps fix --root docs/architecture/runtime.md
```

On first adoption, Demon Docs inserts a marker pair around the whole existing body:

```markdown
## Implementation Map

<!-- doc-ledger:codemap:start -->
- `src/runtime.go`
- `src/runtime_test.go`
<!-- doc-ledger:codemap:end -->
```

The marker prefix follows `[markers].prefix`.

Existing Space Rocks-style fenced maps remain fenced:

````markdown
## Code Map

<!-- doc-ledger:codemap:start -->
```text
services/game-server/internal/runtime/service.go
services/game-server/internal/runtime/service_test.go
```
<!-- doc-ledger:codemap:end -->
````

The command adds paths inside the existing fence rather than creating a second bullet list.

## Verify convergence

Run the explicit check on the same scope:

```bash
ddocs codemaps check --root docs/architecture/runtime.md
```

A clean result returns zero and prints:

```text
ddocs codemaps check passed
```

Pending deterministic changes return non-zero and list affected documents.

Also inspect the repository diff:

```bash
git diff -- docs/architecture/runtime.md
```

Run the fix a second time when validating a new repository convention:

```bash
ddocs codemaps fix --root docs/architecture/runtime.md
```

It should report zero updated files.

## Process a directory

After validating one representative document, widen the scope:

```bash
ddocs codemaps inspect --root docs/architecture
ddocs codemaps fix --root docs/architecture --dry-run
ddocs codemaps fix --root docs/architecture
ddocs codemaps check --root docs/architecture
```

Directory traversal:

- selects regular `.md` files recursively;
- honors `.docignore`;
- skips symbolic-link entries;
- skips `.worktrees/` and `.workingtrees/`; and
- processes files in stable sorted order.

To process the configured documentation root, `fix` alone may omit `--root`:

```bash
ddocs codemaps fix --dry-run
ddocs codemaps fix
```

`check` and `inspect` always require an explicit root.

## Decline an unwanted addition

Codemap execution reuses the normal suggestion ledger. It does not have a second codemap-specific decline database.

List current suggestions for the document:

```bash
ddocs suggestions docs/architecture/runtime.md
```

Inspect the suggestion identifier and candidates:

```bash
ddocs suggestions show SUGGESTION_ID
```

Decline the unwanted relationship with a reason:

```bash
ddocs suggestions decline SUGGESTION_ID --reason "Not part of this document's implementation boundary"
```

Then confirm suppression:

```bash
ddocs codemaps inspect --root docs/architecture/runtime.md
```

The candidate should be reported as declined rather than added.

A decline follows the evidence fingerprint. Unchanged evidence remains suppressed. Materially changed evidence may create a new current suggestion that requires another decision.

Reconsider a previous decision with:

```bash
ddocs suggestions reconsider SUGGESTION_ID
```

Declining a candidate does not remove an already-present codemap link. Manually deleting a link also does not create a persisted decline, so the algorithm may propose it again later.

## Retain links by default

The safe default is:

```toml
[codemap]
remove_undiscovered_links = false
remove_low_score_links = false
```

With those settings, Demon Docs does not remove a valid existing link merely because:

- the current algorithm cannot rediscover it;
- it ranks below the permanent-link tier;
- current evidence adapters do not understand the relationship; or
- the link was originally authored by a human.

This default protects nuanced relationships that deterministic evidence may not reconstruct.

## Enable conservative pruning deliberately

Only enable pruning after inspecting representative documents and accepting the algorithm as the repository's membership authority for that class of link.

```toml
[codemap]
remove_undiscovered_links = true
remove_low_score_links = false
```

`remove_undiscovered_links` removes a currently resolved entry when the planner hides it and cannot recover it from current evidence.

The more aggressive setting is:

```toml
[codemap]
remove_undiscovered_links = true
remove_low_score_links = true
```

`remove_low_score_links` also removes a hidden entry recovered only as a `context` relationship.

Before applying either setting:

```bash
ddocs codemaps inspect --root docs/architecture
ddocs codemaps fix --root docs/architecture --dry-run
```

Review every `remove` line. These settings evaluate algorithm confidence, not human semantic intent.

Broken or moved paths remain a normal link-maintenance concern. Confidence pruning considers resolved codemap entries and should not be used as a dead-link cleanup substitute.

## Adopt an existing partial managed layout

Some repositories may already have a smaller generated region beneath hand-authored links:

```markdown
## Code Map

- `src/authored.go`

<!-- doc-ledger:codemap:start -->
- `src/generated.go`
<!-- doc-ledger:codemap:end -->
```

The next successful codemap fix adopts the complete section:

```markdown
## Code Map

<!-- doc-ledger:codemap:start -->
- `src/authored.go`
- `src/generated.go`
<!-- doc-ledger:codemap:end -->
```

This migration removes the provenance split. Both links remain members of one Demon Docs-owned codemap.

## Confirm daemon exclusion

No watcher or daemon setting enables this workflow. Run it explicitly when codemap maintenance is desired.

The repository demon may observe the resulting Markdown write and reconcile other selected systems, such as link state or indexes, but it does not regenerate the codemap or enqueue another codemap execution.

For unattended verification, run the explicit check in CI:

```bash
ddocs codemaps check --root docs
```

Do not assume `ddocs check` includes codemap generation status.

## Failure and recovery

### Root is outside the docs root

The command refuses the scope before traversal.

Check:

```bash
ddocs status
ddocs config show
```

Then select a contained document or directory.

### File root is not `.md`

Select the Markdown document that owns the codemap. Directory traversal may contain other file types, but they are not selected as codemap documents.

### Section is reported missing

Confirm that the heading is configured or supply `--heading`. If the document should receive a new section, ensure its effective shared or document-specific schema declares a required codemap section and deterministic placement. Heading recognition alone does not authorize Demon Docs to invent a section.

### More than one configured section exists

Consolidate the document to one canonical codemap section. Demon Docs will not choose among multiple matching headings.

### Markers are malformed or duplicated

Ensure the section contains either no codemap markers or exactly one matching start/end pair using the configured marker prefix.

### An expected addition is absent

Check `inspect` for:

- an existing equivalent target;
- a persisted decline;
- insufficient or filtered evidence;
- a missing section; or
- an unsupported repository fact.

Use the evidence output and the codemap algorithm docs before changing thresholds.

### Fix reports that a source changed

Another process or user edited a selected file after planning. Review the edit and rerun `inspect`, dry-run, and fix. Demon Docs will not overwrite the changed source with the stale plan.

### Multi-file fix fails partway

The transaction layer attempts guarded rollback. Inspect the reported error and Git diff before retrying:

```bash
git status --short
git diff
```

### A link keeps returning after manual deletion

Manual deletion does not mean decline. Record the relationship through `ddocs suggestions decline` so unchanged evidence is suppressed.

## Related docs

- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Codemap Pipeline](../architecture/codemap-pipeline.md)
- [Codemap Evidence and Ranking](../architecture/codemap-evidence-and-ranking.md)
- [Reviewing Suggestions and Changes](reviewing-suggestions-and-changes.md)
- [Evaluating Codemap Suggestions](evaluating-codemap-suggestions.md)
- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Current Product Limitations](../limits/current-limitations.md)

## Notes

Start with one representative file and retain the default additive policy until the repository has enough reviewed evidence to justify algorithm-controlled removals.
