---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7ad0-aaf4-6046c8448538
document_type: general
policy_exempt: false
summary: This guide upgrades the Demon Docs binaries and repository state while preserving authored files, configuration, link identity evidence, review history, and a recoverable pre-upgrade snapshot.
---
# Upgrading Demon Docs

Parent index: [Guides](./INDEX.md)

## Purpose

This guide upgrades the Demon Docs binaries and repository state while preserving authored files, configuration, link identity evidence, review history, and a recoverable pre-upgrade snapshot.

## Overview

Demon Docs keeps authored documentation in normal repository files and private state under `.ddocs/`. Most upgrades require replacing the binaries and running normal verification. Some versions may deterministically migrate compatibility configuration, managed index headings, codemap syntax, or legacy link-state files during the next successful mutating pass.

Use a clean branch or worktree for the first upgrade pass. Do not upgrade while a watcher or repository demon is actively writing.

## Version 0.3.4 performance update

Version 0.3.4 parallelizes cold or invalidated frontmatter and document-format validation through a bounded 16-worker pool. Frontmatter source reads and parsing now run concurrently. Document-format source reads, frontmatter parsing, Markdown parsing, and per-document schema enforcement also run concurrently.

Repository-wide coordination remains deterministic and serial where required: duplicate document-ID handling, immutable-state decisions, shared-schema and schema-history decisions, diagnostic ordering, final rewrite ordering, cache publication, and private-state publication.

No configuration or private-state migration is required. Existing validation-cache entries remain compatible. On the retained 1,000-document Windows fixture, warmed-host cold-pass means improved from 157.5 ms to 56.1 ms for frontmatter and from 314.4 ms to 171.6 ms for document format.

## Version 0.3.3 hotfix

Version 0.3.3 disables automatic private-object compaction. Version 0.3.2 could repack `.ddocs` from the daemon while a separate CLI process was reading the same object store, leaving references pointed at missing packfiles or objects. Loose objects are retained until private-state readers and writers share a cross-process lock.

The watcher also schedules one stabilization pass after filesystem watch registration. This closes the startup handoff gap in which a file could change after initial reconciliation but before the operating-system event reader was fully active. Startup readiness checks also use the fresh owner heartbeat to tolerate atomic owner-file replacement without falsely reporting that the daemon stopped.

Validation-cache publication now retries bounded optimistic repository conflicts. A foreground check or fix no longer fails merely because the daemon advanced private state during the same cache write.

Repositories that report `packfile not found`, `object not found`, or an unreadable state-root hash should stop the daemon and preserve the damaged `.ddocs` directory before rebuilding its private Git metadata. Keep `config.toml` and authored schemas.

## Version 0.3.2 behavior changes

Version 0.3.2 changes several execution and private-state details without requiring an authored-document migration:

- `-i` and `--indexes` now select folder indexes only. Use `-d` or `--docs` when the command should also run frontmatter and document-body format policy.
- Link repair runs before the other selected authored-file systems, folder-index convergence writes last, and non-link systems refresh link state only for Markdown sources they actually changed. Clean index-, frontmatter-, or format-only fixes no longer perform a repository-wide link scan or initialize absent link state.
- Missing generated indexes are prepared with their complete planned heading and navigation content before frontmatter or document-format planning. Generated index `author` and `summary` repair defaults do not depend on document-format enforcement being enabled.
- Clean frontmatter and document-format results can be reused from durable validation-cache records. Content, policy, schema, immutable-state, duplicate-identity, or validation-engine changes invalidate reuse.
- Link inventory reads changed and new files through a bounded 16-worker pool while retaining serial deterministic traversal and ordered result merging.
- Review events created by one reconciliation run are stored in one `batch.json` review commit. Existing per-event `event.json` commits remain readable; no destructive review-history conversion is required.
- When one live file and stale absent private records share a `document_id`, reconciliation collapses the stale aliases into the live identity, remaps links, merges path history, and uses that historical path evidence before generic filename guessing.
- Existing pending watcher suppressions are retained and merged with suppressions from a new generated rewrite batch.

In version 0.3.2, cold frontmatter and format validation remained serial, and changed Markdown sources were still reparsed as complete documents. Version 0.3.4 parallelizes cold validation; changed Markdown link sources are still reparsed as complete documents.

## Prerequisites

- The repository's current Git status is understood.
- The currently installed Demon Docs version still runs.
- Active external feeder tokens can be released.
- The Go toolchain required by the target revision is installed when building from source.

## Record the current state

```bash
ddocs --version
ddocs status
ddocs config paths
ddocs config show
demon --status
```

Preserve:

- the Git diff;
- the selected configuration path and resolved values;
- relevant demon logs;
- any current unresolved diagnostics; and
- a copy of `.ddocs/` when private-state migration is expected or the repository is difficult to reconstruct.

Do not commit runtime owner, feeder, heartbeat, or log files merely as an upgrade backup.

## Stop automation

Release external adapter tokens, close shell feeders, and disable the repository demon:

```bash
demon run --false
```

Stop any foreground `ddocs watch` process separately.

Disabling persists `[demon].run = false`, removes feeders, and requests shutdown. Re-enable it after verification when desired.

## Install the target version

From the target checkout:

```bash
go install ./cmd/ddocs
go install ./cmd/demon
```

Or build repository-local binaries:

```bash
go build -o bin/ddocs ./cmd/ddocs
go build -o bin/demon ./cmd/demon
```

Verify that the shell resolves the intended binary:

```bash
ddocs --version
ddocs --help
```

## Inspect configuration before mutation

```bash
ddocs config paths
ddocs config show
ddocs status
```

Confirm that the selected repository, docs root, index filename, codemap headings, reverse roots, parent-link settings, and review limits match expectations.

Current configuration names should be preferred even when compatibility names remain accepted.

## Run read-only verification

Start with narrow checks when the upgrade affects a known subsystem:

```bash
ddocs check --indexes
ddocs check --frontmatter
ddocs check --format
ddocs check --links
ddocs check --reverse
```

Use `ddocs check --docs` when indexes, frontmatter, and document-body format should be verified together.

A link check can read legacy JSON state, but it does not publish the current object format. Reverse checking requires configured roots and matching codemap sections.

Review every diagnostic before running a broad mutating command.

## Apply deterministic migrations

Run the necessary subsystem fixes:

```bash
ddocs fix --indexes
ddocs fix --frontmatter
ddocs fix --format
ddocs fix --links
ddocs fix --reverse
```

Use `ddocs fix --docs` when all three documentation-policy systems should migrate together.

Or run the configured default set:

```bash
ddocs fix
```

Important migrations include:

- recognizing that the built-in folder-index default is now `INDEX.md`; existing repositories should keep an explicit setting such as `index_file = "README.md"` or `index_file = "!README.md"` when they do not intend to rename established index files;
- wrapping or normalizing recognized legacy managed-index headings;
- reading compatibility configuration keys into the current model;
- normalizing retained codemap entry syntax into the current dataset; and
- publishing legacy `.ddocs/files.json` and `.ddocs/links.json` into current object state, then removing the old files after a successful save;
- reading legacy per-event review commits alongside current batched review commits without rewriting the old history; and
- collapsing stale duplicate private file identities only when one present file with the same `document_id` is unambiguous.

Review the Git diff and command output. Private-state publication must not be mistaken for permission to accept unexpected authored-file changes.

## Verify idempotence

```bash
ddocs fix
ddocs check
```

The second fix should report zero changed files. A clean check confirms the selected subsystems are reconciled under the target version.

Run the repository's normal test or release gate when upgrading Demon Docs inside its own checkout or when generated fixtures depend on exact output.

## Re-enable automation

```bash
demon run --true
demon --status
```

External adapters should reacquire new feeder tokens. Do not reuse pre-upgrade tokens.

## Expected result

- The intended binary version is active.
- Configuration selection remains correct.
- Authored files contain only reviewed deterministic changes.
- Current private state and legacy review commits are readable.
- Legacy link JSON is removed only after successful current-state publication.
- Validation-cache records are reused only when all identity inputs still match.
- Normal state and review writes do not trigger private-object compaction.
- A second fix is idempotent.
- `ddocs check` succeeds.
- Watcher automation is re-enabled only after static verification.

## Failure and recovery

### Unsupported state schema

Stop. Preserve `.ddocs/`, the old binary, and the exact error. Do not run repeated fixes or delete state before determining whether a supported migration path exists.

### Configuration selection changed

Use `ddocs config paths` to identify the newly selected file. Remove unintended duplicate local configs or pass an explicit `--config` while correcting repository policy.

### The first fix changes too many files

Restore from Git, narrow the subsystem, inspect changed defaults and configuration compatibility, and retry only after the plan is understood.

### Legacy link-state migration fails

Both `.ddocs/files.json` and `.ddocs/links.json` must exist and decode. Preserve the pair and diagnose the malformed or missing file. Do not fabricate a partial baseline.

### An older binary no longer reads the repository

Downgrade compatibility is not guaranteed after a newer schema writes state. Restore the preserved pre-upgrade `.ddocs/` snapshot together with the corresponding authored-file state, or continue with a supported newer binary.

## Related docs

- [Compatibility and Migrations](../reference/compatibility-and-migrations.md)
- [Configuration Reference](../reference/configuration.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)
- [Repository State and Transactions](../architecture/repository-state-and-transactions.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)

## Notes

Git remains the authoritative rollback mechanism for authored files. A `.ddocs/` backup preserves private evidence but does not replace a matching source snapshot.
