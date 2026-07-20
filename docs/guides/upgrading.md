---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7ad0-aaf4-6046c8448538
document_type: general
policy_exempt: false
summary: This guide upgrades the Demon Docs binaries and repository state while preserving authored files, configuration, link identity evidence, review history, and a recoverable pre-upgrade snapshot.
---
# Upgrading Demon Docs

Parent index: [Guides](./README.md)

## Purpose

This guide upgrades the Demon Docs binaries and repository state while preserving authored files, configuration, link identity evidence, review history, and a recoverable pre-upgrade snapshot.

## Overview

Demon Docs keeps authored documentation in normal repository files and private state under `.ddocs/`. Most upgrades require replacing the binaries and running normal verification. Some versions may deterministically migrate compatibility configuration, managed index headings, codemap syntax, or legacy link-state files during the next successful mutating pass.

Use a clean branch or worktree for the first upgrade pass. Do not upgrade while a watcher or repository demon is actively writing.

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
ddocs check --docs
ddocs check --links
ddocs check --reverse
```

A link check can read legacy JSON state, but it does not publish the current object format. Reverse checking requires configured roots and matching codemap sections.

Review every diagnostic before running a broad mutating command.

## Apply deterministic migrations

Run the necessary subsystem fixes:

```bash
ddocs fix --docs
ddocs fix --links
ddocs fix --reverse
```

Or run the configured default set:

```bash
ddocs fix
```

Important migrations include:

- wrapping or normalizing recognized legacy managed-index headings;
- reading compatibility configuration keys into the current model;
- normalizing retained codemap entry syntax into the current dataset; and
- publishing legacy `.ddocs/files.json` and `.ddocs/links.json` into current object state, then removing the old files after a successful save.

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
- Current private state is readable.
- Legacy link JSON is removed only after successful current-state publication.
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
