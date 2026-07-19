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

`.docignore` lives at the repository root and uses Git ignore syntax. It applies to index scanning, link scanning, reverse-index traversal, and watcher event filtering as implemented by each subsystem.

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
incoming-link and reverse-index state
transaction and generated-write metadata
review decisions and repair controls under `refs/ddocs/review`
applied-change events with before/after blobs and hashes
repository demon runtime ownership
feeder leases and heartbeats
bounded logs
```

Private state is implementation-owned and should not be hand-edited while commands or watchers are active. Link identity state uses `refs/ddocs/state`; suggestion decisions and applied-change history use `refs/ddocs/review`. Neither creates commits on the user's normal Git branch.

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
- remove existing codemap links as allegedly irrelevant; or
- treat research candidates as authored relationships.

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
- [Diagnostics and Exit Behavior](diagnostics-and-exit-behavior.md)
- [Reconciliation Pipeline](../architecture/reconciliation-pipeline.md)
- [Review Ledger](../architecture/review-ledger.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Private state is not a substitute for source control. Git remains the authoritative history for authored files and reviewed generated changes.
