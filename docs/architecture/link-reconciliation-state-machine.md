---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-7c51-acc4-30369edf7f42
document_type: general
policy_exempt: false
summary: 'This document defines the implemented lifecycle of repository-local link records: how Markdown sources enter the graph, how prior identity is reused, how targets are resolved, how repair states are selected, and how records converge...'
---
# Link Reconciliation State Machine

Parent index: [Architecture](./README.md)

## Purpose

This document defines the implemented lifecycle of repository-local link records: how Markdown sources enter the graph, how prior identity is reused, how targets are resolved, how repair states are selected, and how records converge after generated rewrites.

## Overview

Link reconciliation is a stateful comparison between:

```text
previous file and link records
current repository and external-target inventory
current Markdown source contents
review decisions and repair blocks
```

The result is a deterministic `links.Plan` containing current file records, current link records, diagnostics, unresolved counts, optional generated rewrites, and optional watcher suppressions after application.

The state machine does not decide semantic relationships. It tracks filesystem-backed targets and only rewrites a destination when one deterministic target exists or a user has selected one candidate through the review workflow.

## Code root

```text
internal/links/
```

## Responsibilities

This boundary owns:

- persistent file identities and path history;
- Markdown source identities and ordered outgoing link records;
- stable link identity reuse where prior records can be matched;
- parser-version invalidation;
- exact target resolution;
- moved-target and case-only repair detection;
- bounded candidate discovery for missing targets;
- link statuses, candidates, diagnostics, and unresolved counts;
- generated rewrite planning;
- source-record refresh after generated writes; and
- publication-ready file, source, incoming-link, path, and suppression records.

## Does not own

It does not own:

- documentation-folder index generation;
- reverse-index projection;
- semantic codemap relationships;
- choosing among ambiguous candidates without a recorded user decision;
- review-history storage mechanics;
- watcher scheduling;
- cross-subsystem command transactions; or
- heading-fragment validation.

## Reconciliation inputs

Each reconciliation loads or constructs four inputs.

### Previous link state

The private `.ddocs` state reference provides:

- `FileRecord` entries keyed by stable file ID;
- `source/<source-file-id>` records containing ordered outgoing links;
- current and historical paths;
- fingerprints, sizes, modification times, and parser versions; and
- target identities used to recognize later moves.

If repository-backed state is absent, legacy `files.json` and `links.json` manifests may be read for migration. If neither state form exists, the repository is uninitialized for links.

### Current inventory

The inventory contains present repository files and directories plus explicitly referenced external targets. Repository traversal obeys permanent exclusions and `.docignore`. Explicit external filesystem targets are not governed by repository ignore rules.

The inventory reconciles current paths with previous identities. A file that moved can therefore remain the same logical target even though its path changed.

### Current Markdown sources

Present repository Markdown files are processed in deterministic path order. Link occurrences inside a source are assigned ordinals in parser order. Explicit, reference-definition, wiki, image, and supported local HTML forms all enter the same stored record model.

### Review policy

The replayed review policy can block an otherwise deterministic generated repair. An exact matching repair block produces `blocked`; changed repair evidence produces `stale_block`.

## Record identity

### File identity

A `FileRecord.ID` is the durable identity for one observed file or directory. Current path is stored separately from identity. Historical paths remain available as move evidence.

### Link identity

A new link ID is derived from:

```text
source file ID
ordinal within the parsed source
syntax kind
original target text
```

When a current occurrence matches one previous record by ordinal, target, and syntax, the previous ID is retained. If ordinal matching fails, one unique previous record with the same target and syntax can still retain its ID. Multiple possible prior matches prevent fallback reuse.

Link IDs identify occurrences, not semantic relationships. Moving or editing surrounding prose can change the ordinal and therefore prevent identity reuse when no unique target-and-syntax fallback exists.

## Source processing paths

A Markdown source follows one of three principal paths.

### Internal moved-target path

A source can bypass full reparsing when all of these are true:

- the source file identity is present in previous and current inventory;
- its scope, path, and fingerprint are unchanged;
- its prior link records contain complete rewrite metadata;
- all prior records use the current parser version and are reusable; and
- at least one prior target identity moved.

The reconciler reads the unchanged source, calculates replacements directly from stored incoming-link records, applies review-block policy, and creates updated link records and an optional generated rewrite.

This path preserves known occurrence identities and avoids interpreting a Demon Docs repair as an unrelated user edit.

### Unchanged-source reuse path

A source is reused without reading and parsing its Markdown when:

```text
source identity and current fingerprint are unchanged
parser version is current
every prior record has rewrite metadata
every prior record status is valid
```

Only `SourcePath` is refreshed. All prior outgoing records are copied into the new plan.

Any non-`valid` status forces the source through parsing on the next pass. This intentionally lets transient repair states converge to the normal current state after a successful write.

### Parsed-source path

All other sources are read and parsed from current bytes. This includes:

- new sources;
- externally edited sources;
- sources affected by a parser-version change;
- sources with incomplete legacy record metadata;
- sources containing unresolved or transient statuses; and
- sources whose fingerprints no longer match.

The parsed occurrences replace that source's outgoing graph projection in the new plan.

## Target resolution order

Each local occurrence is resolved in a deterministic sequence.

### Exact target

The syntax-specific exact resolver first checks the current inventory. Supported wiki syntax may apply extensionless Markdown and basename rules that differ from ordinary path syntax.

If the resolved filesystem path exists but is not yet represented in inventory, it is recorded as a target, including an explicit external target.

An exact target produces `valid`, unless path casing differs from the filesystem representation.

### Preferred previous identity

When an exact target is absent, the prior occurrence's target file ID is preferred. If that identity is present at a different path, it becomes the sole move candidate.

### Candidate discovery

If no preferred identity is available, candidate discovery uses syntax-aware rules. Ordinary path syntax searches after link state is initialized. Wiki syntax may use its repository-wide resolution rules during initial observation because a bare wiki target is not necessarily relative to the source directory.

Candidate discovery can use:

- matching file identity;
- basename and kind;
- previous fingerprint;
- repository inventory; and
- bounded nearby search for an external target.

Candidate lists are normalized and sorted before they are stored or displayed.

## Link statuses

`LinkRecord.Status` describes the current reconciliation outcome for one occurrence.

| Status | Meaning | Rewrite eligibility | Unresolved |
| --- | --- | --- | --- |
| `valid` | The authored destination resolves exactly under current syntax rules. | None required. | No |
| `case_mismatch` | The destination resolves, but authored casing differs from the actual path. | Automatic after initialization unless blocked. | No unless blocked |
| `moved` | One deterministic target exists at a different rendered destination. | Automatic after initialization unless blocked. | No |
| `broken` | No exact target or candidate exists. | None. | Yes |
| `ambiguous` | Multiple candidate targets exist. | Requires review selection. | Yes |
| `blocked` | A deterministic repair exactly matches an active repair block. | Suppressed until unblocked or evidence changes. | Yes |
| `stale_block` | A related repair block exists, but current evidence no longer matches exactly. | Requires review because the prior block no longer governs the changed repair. | Yes |
| `undefined_reference` | An explicit or collapsed reference use has no matching definition. | None; the missing definition is authored content. | Yes |

`case_mismatch` and `moved` describe planned or just-applied repair states. `refreshGeneratedSources` updates occurrence offsets, syntax fields, source fingerprints, sizes, and modification times, but does not rewrite the status to `valid`. Because the unchanged-source reuse path accepts only `valid` records, the next pass reparses the repaired source and normalizes the record to `valid` when the new destination resolves exactly.

## State transitions

### New or externally edited occurrence

```text
parse occurrence
-> non-local or ignored target: omit from graph
-> exact target: valid or case_mismatch
-> no exact target + zero candidates: broken
-> no exact target + one candidate: valid, moved, blocked, or stale_block
-> no exact target + multiple candidates: ambiguous
```

### Known target move with unchanged source

```text
reuse stored occurrence identity
-> calculate syntax-preserving new destination
-> unchanged rendered destination: valid metadata update
-> active exact repair block: blocked
-> changed block evidence: stale_block
-> otherwise: moved + generated rewrite
```

### Ambiguous candidate selection

```text
ambiguous or stale review suggestion
-> user selects one candidate
-> selection validates current source and candidate identity
-> normal generated rewrite path
-> applied-change event
-> refreshed source/link metadata
-> later reconciliation normalizes to valid
```

Selection does not create a separate permanent accepted state. It supplies one concrete repair to the existing rewrite and publication path.

### Repair block transitions

```text
deterministic repair
-> no matching block: moved or case_mismatch
-> exact matching block: blocked
-> same relationship with changed fingerprint: stale_block
-> explicit unblock: normal deterministic evaluation resumes
```

### Undefined references

An undefined explicit or collapsed reference becomes a stored `reference_use` record with `undefined_reference`. It has no target identity or repair candidate. Adding the missing definition changes the parsed document and causes normal reconciliation on the next pass.

## Initialization behavior

The first link-enabled `fix` or `watch` pass creates the baseline and publishes current identities. It does not apply generated repair rewrites. The plan reports that link state is uninitialized.

A read-only `check -l` reports initialization as required and fails without publishing state.

Baseline parsing still records exact targets, broken links, ambiguity, undefined references, and syntax-specific observations. Automatic path repair begins only after initialized state can provide a trusted prior identity baseline.

## Generated rewrite construction

For one source, all non-overlapping destination replacements are sorted by byte offset and validated against the exact old destination text. `NewGeneratedRewrite` encodes both old and new bytes using the source document's existing newline representation and records:

- source file ID and absolute source path;
- expected old and new SHA-256 hashes;
- affected link IDs;
- old and new destination text;
- target identity/path metadata where available;
- suggestion kind and selection mode; and
- originating suggestion ID for review-selected repairs.

The state machine only plans the rewrite. Filesystem application and multi-store publication are owned by [Generated Rewrite Publication](generated-rewrite-publication.md).

## Generated-source refresh

After source files are replaced, each rewritten source is read and parsed again for verification. Stored outgoing records are matched in ordinal order against the expected current target text. The refresh updates:

- byte start and end offsets;
- line and column;
- syntax representation;
- raw path and suffix fields;
- source fingerprint;
- source size; and
- source modification time.

Workers produce detached refresh results. Results merge back into the plan in rewrite order so concurrency cannot change stored ordering.

If an expected link cannot be found after a generated write, refresh fails and link-state publication does not occur. The authored rewrite and any already-published review event may already exist; recovery is a later reconciliation, as described in the publication architecture.

## Persistent projection

A successful state publication writes deterministic records for:

```text
meta/state
file/<file-id>
path/<path-key>
source/<source-file-id>
incoming/<target-file-id>/<source-file-id>
write/<source-file-id>
```

The new root removes obsolete records under those namespaces and publishes all desired records through one private object-repository transaction.

The `write/` records are pending watcher suppressions. A watcher consumes a suppression only when the observed source matches the expected generated after hash. Mismatched content is treated as a real external edit rather than suppressed.

## Determinism and ordering

- Markdown sources are processed in sorted repository path order.
- Link records are stored by source path and ordinal.
- Candidate display paths are sorted.
- File path history is sorted before publication.
- Generated rewrites are planned before application.
- Worker results merge in plan order.
- Diagnostic order follows deterministic source and occurrence order, not worker completion order.

## Failure behavior

### State load failure

Unsupported schema, malformed private records, or object-repository errors abort planning. No authored source is written.

### Inventory failure

Traversal, ignore-policy, fingerprint, or target-record errors abort planning. No generated writes occur.

### Source read or parse-path failure

Failure to read a required Markdown source aborts planning. Previously prepared in-memory records are discarded with the plan.

### Concurrent source edit

Expected-hash failure during generated application aborts the rewrite batch and refuses to overwrite current content. The changed source is handled as an external edit on a later reconciliation.

### Refresh failure

A generated source that cannot be verified prevents link-state publication. The source rewrite is not automatically reversed at this stage because review publication has already succeeded and the failure occurs after the generated-write transaction. A later reconciliation rebuilds current graph records from authored files.

### State publication failure

The current `refs/ddocs/state` projection remains unchanged. Authored source rewrites and review events may already be durable. Re-running reconciliation observes current source bytes and converges private link state.

## Invariants and safety boundaries

- Only repository-contained Markdown sources are rewritten.
- External targets may be observed but are never modified.
- Ignored repository targets do not participate in the graph.
- One deterministic target or one explicit user selection is required before rewriting.
- Authored label text, title text, aliases, fragments, queries, and newline representation remain unchanged.
- Expected source bytes must match before replacement.
- A parser-version change prevents stale-record reuse.
- Only `valid` prior records qualify for the unchanged-source fast path.
- Worker concurrency cannot determine result ordering.
- The link graph is a filesystem relationship graph, not a semantic documentation graph.

## Code map

- `internal/links/model.go` — file, link, and plan records.
- `internal/links/reconcile.go` — state-machine orchestration, source paths, statuses, target candidates, and rewrite planning.
- `internal/links/inventory.go` — repository and external-target inventory.
- `internal/links/filemeta.go` — file identity and fingerprint metadata.
- `internal/links/target.go` — local target resolution and candidate discovery.
- `internal/links/syntax_targets.go` — syntax-specific exact and candidate resolution.
- `internal/links/parser.go` and parser extensions — current Markdown occurrence extraction.
- `internal/links/state.go` — private record load and publication projection.
- `internal/links/apply.go` — generated-source verification and metadata refresh.
- `internal/links/review_suggestions.go` — projection of ambiguous and blocked records into review suggestions.
- `internal/links/review_selection.go` — selected-candidate conversion into concrete repair records.
- `internal/review/policy.go` — decline and repair-block replay used during reconciliation.

## Tests

Focused state-machine coverage includes:

- `internal/links/reconcile_test.go` — baseline, exact resolution, moves, ambiguity, undefined references, and identity behavior.
- `internal/links/parser_test.go` — occurrence ordering and syntax extraction.
- `internal/links/review_integration_test.go` — blocked and stale-block transitions.
- `internal/links/review_selection_test.go` — selected-candidate and transient-status handling.
- `internal/links/rewrite_test.go` — transformation construction and preservation.
- `internal/links/rewrite_transaction_test.go` — expected-hash and rollback boundaries.
- `internal/links/state` behavior exercised through reconciliation and integration tests.

Run:

```bash
go test ./internal/links -count=1
```

## Related docs

- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [Review Lifecycles](review-lifecycles.md)
- [Supported Link Syntax](../reference/supported-link-syntax.md)
- [Repository State and Transactions](repository-state-and-transactions.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)

## Notes

Status values are internal reconciliation state, not a stable machine-readable public API. User-facing behavior is defined by CLI diagnostics and review commands.
