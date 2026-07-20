---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-76de-a665-333edc85e448
document_type: general
policy_exempt: false
summary: This document defines which repository files Demon Docs may manage, which portions remain authored, what private state is stored under .ddocs/, and which source-preservation guarantees apply.
---
# Managed Files and State

Parent index: [Reference](./README.md)

## Purpose

This document defines which repository files Demon Docs may manage, which portions remain authored, what private state is stored under `.ddocs/`, and which source-preservation guarantees apply.

## Overview

Demon Docs does not replace Markdown or Git with a proprietary authoring store. Authored files remain ordinary repository files. Generated ownership is narrow and marked, while private state records deterministic identities, history, reverse indexes, transactions, and runtime coordination needed to reconcile repository changes.

## Managed index blocks

Documentation folder indexes use HTML marker pairs. By default:

```markdown

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
```

Demon Docs owns content between matching markers. Prose outside managed blocks remains authored.

Heading- and marker-like text inside fenced code blocks is treated as code content, not document structure.

## Parent links

Parent navigation can be enabled independently for folder indexes and indexed files. The default label is `Parent index`.

Demon Docs edits only configured Markdown-like file types for parent links. The root index has no parent link.

## Frontmatter

When `[frontmatter].enabled` is true, Demon Docs validates every non-ignored `.md` file beneath the configured docs root, including generated folder indexes. Existing YAML (`---`) and TOML (`+++`) blocks keep their format. Missing blocks use the configured default format.

`fix` may insert configured defaults or generated UUID/date values, restore immutable fields from private state, and remove unknown fields when configured. It does not replace an existing valid mutable value. Malformed blocks, invalid mutable values, and required fields without a repair source remain authored problems and are not guessed.

The Markdown body, newline convention, and final-newline behavior remain preserved during frontmatter replacement.

## Document schemas and body format

Human-authored shared TOML schemas live under `.ddocs/schemas/`. Generated document-specific TOML schemas live under `.ddocs/document-schemas/`; they are explicit exceptions and remain human-editable.

Body-format enforcement uses ordinary Markdown headings rather than managed-region markers. `fix` may move complete sections, change heading levels, propagate schema heading renames, and create missing sections with placeholder text. It does not rewrite prose. Unknown or duplicate human-authored sections block body changes until an explicit ignore, merge, delete, or manual repair decision is made.

## Managed codemap sections

Explicit codemap execution adopts the complete configured codemap section under one marker pair derived from `[markers].prefix`:

```markdown
## Code Map

<!-- doc-ledger:codemap:start -->
- `src/runtime.go`
- `src/runtime_test.go`
<!-- doc-ledger:codemap:end -->
```

The heading remains outside the marker pair. Everything in the section body, including existing links and explanatory prose, becomes part of the unified managed region. Demon Docs does not preserve separate authored and generated provenance groups inside the section.

On first adoption, a prior partial generated block is expanded to cover the whole section. Existing marker lines are removed and one canonical pair is rendered. Malformed or duplicated codemap markers are an error rather than an automatically guessed repair.

Supported rendering behavior is:

- a complete fenced path list remains fenced and receives new targets before its closing fence;
- a bullet map retains the first recognized bullet prefix and indentation;
- additions are normalized, deduplicated, and sorted;
- existing section content is not globally re-sorted;
- selected removal deletes only the extracted entry line; and
- source line endings, final-newline state, and file mode are preserved through the shared text and transaction layers.

Existing valid links are retained by default. `[codemap].remove_undiscovered_links` and `[codemap].remove_low_score_links` permit confidence-based pruning only when explicitly enabled. Declined recommendations suppress additions but do not remove an existing entry.

When a selected document schema requires a codemap section, explicit codemap execution may create the missing section at its schema-defined position before adopting it. Schemas without a codemap section leave missing sections unchanged.

## Local link rewrites

Link reconciliation may rewrite the resolved path portion of recognized local links in repository Markdown sources.

It preserves:

- link labels and image alt text;
- Markdown titles;
- wiki aliases;
- query strings and fragments;
- angle wrapping;
- surrounding prose;
- source newline style; and
- whether the source ended with a final newline.

Only one deterministic destination permits an automatic rewrite. Ambiguous links remain unchanged.

## Reverse-index outputs

Configured reverse-index roots may receive generated documentation projections. Generated regions remain explicit and deterministic. Authored codemap sections are inputs; inferred candidates do not become authored truth automatically.

## `.docignore`

`.docignore` lives at the repository root and uses Git ignore syntax. It applies to index scanning, frontmatter enforcement, link scanning, reverse-index traversal, and watcher event filtering as implemented by each subsystem.

These directories are always pruned at every depth:

```text
.git/
.ddocs/
.obsidian/
logseq/
```

Repository-specific exclusions belong in `.docignore`.

## `.ddocs/` private state

The initialized repository stores private Demon Docs state under `.ddocs/`.

State families include:

```text
configuration
object and identity records
path history and fingerprints
recorded immutable frontmatter values
shared schema history used for deterministic heading migrations
human-authored shared schemas under `schemas/`
human-editable document-specific schemas under `document-schemas/`
incoming-link and reverse-index state
transaction and generated-write metadata
review decisions and repair controls under `refs/ddocs/review`
applied-change events with before/after blobs and hashes
repository demon runtime ownership
feeder leases and heartbeats
bounded logs
```

Object state, runtime state, and internal refs are implementation-owned and should not be hand-edited while commands or watchers are active. The TOML files under `.ddocs/schemas/` and `.ddocs/document-schemas/` are the deliberate exceptions: both are editable policy inputs, although document-specific files are initially generated by explicit user decisions. Link identity, immutable-frontmatter records, and shared-schema history use `refs/ddocs/state`; suggestion decisions and applied-change history use `refs/ddocs/review`. Neither creates commits on the user's normal Git branch.

## Rebuildability

The filesystem and authored repository remain the primary rebuild source. Much of `.ddocs/` can be reconstructed, but deleting it loses historical evidence used for move reconciliation and resets daemon/runtime ownership.

Delete or reset `.ddocs/` only as a deliberate recovery action after stopping active processes and preserving any diagnostics needed to understand the failure.

## Stateless moves

`ddocs mv` does not require, create, or update `.ddocs/`. It performs one explicit repository-contained move and rewrites affected Markdown directly. In an initialized repository, a later watcher or link reconciliation pass refreshes persistent identity state.

## Mutation boundaries

Demon Docs does not:

- rewrite arbitrary prose;
- change link labels, titles, or aliases to improve style;
- edit binary targets;
- edit external target files;
- traverse symbolic-link entries as owned repository trees;
- choose among ambiguous targets;
- remove existing codemap links based on algorithm confidence unless the repository explicitly enables that pruning policy;
- treat a decline as an instruction to remove an already-present codemap link; or
- run codemap generation through normal reconciliation, watch, or daemon paths.

## Diagnostics

Unexpected writes should be investigated through:

```bash
ddocs config paths
ddocs config show
ddocs check
demon --status
demon --logs
```

Stop active automation before performing manual recovery.

## Examples

A path-only rewrite:

```markdown
[Configuration](../old/configuration.md#selection)
```

may become:

```markdown
[Configuration](../reference/configuration.md#selection)
```

The label and fragment remain unchanged.

## Related docs

- [Configuration Reference](configuration.md)
- [Document Schemas And Format Enforcement](document-schemas.md)
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Review Ledger](../architecture/review-ledger.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Private state is not a substitute for source control. Git remains the authoritative history for authored files and reviewed generated changes.
