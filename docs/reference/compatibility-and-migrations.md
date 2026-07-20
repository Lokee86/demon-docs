---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-75d2-93b1-1542c56d90bf
document_type: general
policy_exempt: false
summary: This document defines the current compatibility inputs, automatic migrations, and refusal behavior used when Demon Docs reads older configuration, managed indexes, codemap syntax, or link state.
---
# Compatibility and Migrations

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines the current compatibility inputs, automatic migrations, and refusal behavior used when Demon Docs reads older configuration, managed indexes, codemap syntax, or link state.

## Overview

Demon Docs accepts a bounded set of previous names and formats so repositories can upgrade without an all-at-once rewrite. Compatibility inputs are read into the current model. New files and examples should use current names.

Automatic migration is permitted only when the old format can be translated deterministically. Unsupported private-state schemas fail explicitly rather than being guessed.

## Configuration filenames

Selection accepts current and compatibility paths in this order:

```text
1. explicit --config PATH
2. nearest .ddocs/config.toml found upward
3. ./.demon-docs.toml
4. ./demon-docs.toml
5. ./.doc-ledger.toml
6. ./doc-ledger.toml
7. user config: demon-docs/config.toml
8. user compatibility config: doc-ledger/config.toml
9. built-in defaults
```

Repository `.ddocs/config.toml` is the current initialized-repository format. Compatibility local files are selected only from the current directory unless bounded discovery is needed inside an existing `.ddocs` marker.

Local and global files are not merged.

## Configuration key aliases

Current keys and accepted compatibility forms include:

```text
docs_root                         preferred repository root key
root                              standalone compatibility root key

[parent_link].folder_indexes      current folder-index switch
[parent_link].indexed_files       current file switch
[parent_link].enabled             compatibility switch used for unspecified current switches

[reverse_index].roots             current root list
[reverse_index].folders           compatibility root list when roots is absent

[editable].parent_index_extensions current key
[editable].extensions               compatibility key

[sections.<kind>].heading         current heading field
[sections.<kind>].title/name      compatibility heading fields
```

`[aliases].files` and `[aliases].folders` configure legacy managed-index headings that may be recognized during transition.

When both a current key and its compatibility form are present, the current key takes precedence where the loader defines an explicit precedence.

## Legacy managed index headings

The default legacy headings are:

```text
Top-Level Files
Top-Level Folders
```

When a legacy index has recognized headings and no managed markers, reconciliation can wrap the existing entries in current managed sections. When markers already exist, legacy headings are normalized without treating fenced-code text as document structure.

Description preservation and source newline rules still apply during migration. Review the first `fix --docs` diff before accepting it.

## Codemap syntax compatibility

The codemap extractor recognizes current Markdown target forms and retained legacy inline or indented entries. The normalized dataset records the syntax kind and source location, so downstream evidence and reverse-index consumers receive one deterministic model.

Compatibility parsing does not make every prose path a codemap target. Entries must still occur under a configured codemap heading and satisfy extractor rules.

## Legacy link-state migration

Older link state may exist as:

```text
.ddocs/files.json
.ddocs/links.json
```

When current object state is absent, Demon Docs can load both legacy manifests as one initialized link baseline. The next successful link-state publication writes current records into the private `.ddocs` object repository and removes the two legacy JSON files.

Migration occurs through a normal successful state save, such as a link-enabled `fix` or watcher pass. A read-only `check` does not publish migrated state.

Both legacy files must be present and decodable. A partial pair or invalid JSON fails explicitly.

## Current private-state schemas

Current link state records include schema versions. An unsupported stored schema returns an error such as an unsupported link-state schema rather than silently rebuilding over the unreadable state.

Preserve the failing `.ddocs/` directory and command output before reset. Reinitializing private state loses historical identity evidence and should remain a deliberate recovery step.

Review-ledger events are append-only Git objects under the private review reference. Undo eligibility settings may change without deleting audit history.

## Linked-worktree bootstrap compatibility

A linked Git worktree whose primary worktree is initialized can copy the primary `.ddocs/config.toml` during the first mutating demon entry. It initializes fresh local object storage and does not copy runtime state, link history, review history, or owner/feeders.

This is a bootstrap operation, not ongoing synchronization.

## Command aliases

Current compatibility aliases include:

```text
-i, --indexes   aliases --docs
```

`demon` and `ddocs demon` expose the same repository-demon application boundary.

Aliases may remain supported without being preferred in new documentation.

## Upgrade and downgrade behavior

Demon Docs supports deterministic forward migrations described on this page. It does not promise that an older binary can read state written by a newer schema.

Before changing binary versions:

- stop active watcher automation;
- preserve Git status and relevant `.ddocs/` diagnostics;
- review release notes or repository changes;
- run configuration inspection; and
- use a branch or clean worktree for the first mutating pass.

## Diagnostics and failure behavior

Migration refuses to continue when:

- a selected compatibility config cannot be decoded;
- a root escapes repository scope;
- only one legacy link-state JSON file exists;
- legacy state JSON is invalid;
- a current private-state schema is unsupported;
- managed marker pairs are incomplete; or
- linked-worktree metadata cannot be resolved safely.

The correct response is to preserve evidence and diagnose the specific format. Do not delete `.ddocs/` as the first troubleshooting step.

## Related docs

- [Upgrading Demon Docs](../guides/upgrading.md)
- [Configuration Reference](configuration.md)
- [Managed Files and State](managed-files-and-state.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Repository Scope and Worktrees](../architecture/repository-scope-and-worktrees.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Compatibility support exists to preserve deterministic behavior during transition. Deprecated names should not be copied into new examples merely because the loader still accepts them.
