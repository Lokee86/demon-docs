---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-734f-a812-8f4c708fb696
document_type: general
policy_exempt: false
summary: This document defines how Demon Docs publishes one planned batch of generated Markdown rewrites across authored source files, the append-only review ledger, refreshed link metadata, and private link state.
---
# Generated Rewrite Publication

Parent index: [Architecture](./README.md)

## Purpose

This document defines how Demon Docs publishes one planned batch of generated Markdown rewrites across authored source files, the append-only review ledger, refreshed link metadata, and private link state.

## Overview

A generated rewrite is a complete before-and-after byte transition for one Markdown source. It carries expected hashes and exact transformations so Demon Docs can reject stale plans rather than reinterpreting current content during application.

Publication spans three physically separate surfaces:

```text
authored Markdown files
refs/ddocs/review
refs/ddocs/state
```

Each surface has a strong local publication boundary. They do not form one shared atomic transaction. The implemented sequence uses preflight checks and targeted rollback to keep the surfaces consistent where possible, then relies on a later reconciliation to converge state after failures that occur beyond the rollback boundary.

## Code root

```text
internal/filetxn/
internal/links/rewrite.go
internal/links/rewrite_transaction.go
internal/links/apply.go
internal/links/review_record.go
internal/links/state.go
internal/review/store_batch.go
internal/ddrepo/
```

## Responsibilities

The shared `internal/filetxn` boundary owns the content-addressed filesystem transaction used by both link and frontmatter rewrites. Link publication layers review history, graph refresh, and durable link suppressions on top of it.

The shared filesystem boundary owns:

- validating the internal consistency of prepared rewrites;
- checking every source before the first replacement;
- checking each source again immediately before its replacement;
- preserving source permissions and exact newline-encoded bytes;
- same-directory temporary-file writes and OS-specific atomic replacement;
- post-replacement hash verification;
- rollback of an attempted source batch after a filesystem failure.

The link-specific publication layer additionally owns:

- preparation and append-only publication of applied-change events;
- rollback of source files when review-history publication fails;
- refreshing generated link offsets and source metadata;
- publishing the complete private link-state projection; and
- making watcher suppressions durable with that state projection.

Frontmatter uses the shared filesystem boundary but owns its own diagnostics, immutable-value projection, and rollback when private-state publication fails. Documentation-body format uses the same rewrite boundary and owns schema-history snapshots and invalidation recovery.

## Does not own

It does not own:

- deciding which target is correct;
- generating ambiguous suggestions;
- parsing normal external edits;
- cross-subsystem rollback for documentation or reverse indexes;
- watcher event scheduling;
- review-policy replay;
- arbitrary user-file recovery after content diverges from both recorded states; or
- a global transaction spanning filesystem files and private Git references.

## Publication inputs

`links.ApplyAndSave` receives a completed `links.Plan`.

The plan can contain:

- zero or more `GeneratedRewrite` values;
- current file and link manifests;
- current diagnostics and unresolved counts;
- repository root; and
- no durable suppressions yet.

Each generated rewrite contains:

```text
source file ID
absolute source path
expected old SHA-256
expected new SHA-256
exact old and new bytes
ordered link transformations
suggestion kind and selection mode
optional originating suggestion ID
```

The old and new byte slices are populated only by the package constructors. This prevents a caller from supplying hashes that describe different content than the bytes later written.

## End-to-end sequence

The implemented publication sequence is:

```text
1. prepare applied-change records and review append requests
2. preflight the complete source batch
3. replace and verify each source in plan order
4. append the complete applied-change batch to refs/ddocs/review
5. attach watcher suppressions to the in-memory plan
6. re-read and verify every generated source
7. refresh link offsets and source file metadata
8. publish the complete link-state projection to refs/ddocs/state
9. remove migrated legacy JSON state files
```

No later step begins when an earlier step returns an error.

## Change preparation

Before authored files are touched, `prepareGeneratedChanges` constructs one review change per rewrite.

All changes in the plan share one new reconciliation run ID. Each change receives:

- a change ID;
- suggestion kind;
- selection mode (`deterministic`, `user`, or `undo` as applicable);
- optional originating suggestion ID;
- source file identity and repository-relative path;
- before and after hashes;
- applied timestamp;
- individual transformation records; and
- sorted related-target records.

Each transformation receives a stable relation key, evidence fingerprint, and transformation ID derived from the source relationship and before/after destination evidence. The corresponding append request retains exact before and after file blobs for later inspection and undo.

Preparation opens the review store and validates enough metadata to build the batch, but does not advance the review reference.

## Source-batch preflight

`links.ApplyGenerated` validates link-specific metadata, then delegates the prepared byte transitions to `filetxn.Apply`. Frontmatter plans build the same `filetxn.Rewrite` values and call the same apply path.

It rejects:

- empty paths;
- duplicate paths in one batch;
- missing hashes;
- rewrites not created by the package constructors;
- old bytes that do not match the expected old hash; and
- new bytes that do not match the expected new hash.

It then preflights every current source through a bounded worker pool. Each source must:

- exist;
- be a regular file;
- be readable; and
- match its expected old hash.

This is a batch-wide concurrency barrier. A stale source prevents every authored replacement in the batch.

Worker completion order does not affect result order. Preflight results remain indexed by rewrite position.

## Per-file replacement

After the complete batch passes, sources are processed in plan order.

Immediately before each replacement, Demon Docs repeats that source's stat, regular-file, read, and expected-old-hash checks. This closes the interval between the parallel batch preflight and the actual write.

The source mode is retained. Replacement then uses:

```text
create temporary file in the source directory
-> apply original permission bits
-> write exact new bytes
-> sync and close the temporary file
-> perform the platform-specific atomic replacement
-> re-read the destination
-> verify the expected new hash
```

A watcher suppression is constructed only after the new hash verifies. Suppressions remain in rewrite order and carry source identity, both hashes, affected link IDs, and old/new destinations.

## Filesystem failure and rollback

If the second preflight, replacement, read-back, or new-hash verification fails, Demon Docs attempts to roll back every attempted source in reverse order.

For each attempted source, current bytes are classified as:

| Current hash | Rollback action |
| --- | --- |
| Expected old hash | No action; the source is already restored or was never changed. |
| Expected new hash | Atomically replace it with the recorded old bytes, then verify the old hash. |
| Any other hash | Refuse to overwrite it because newer or unknown content exists. |

Rollback errors are joined with the original apply error. Refusal is deliberate: recovering batch symmetry must not destroy a user edit that arrived after the generated replacement.

The same guarded mechanism is exposed as `RollbackGenerated` for a successfully applied batch. It reverses old/new expectations and still refuses to overwrite content outside the recorded after state.

## Review-ledger publication

After every source verifies in its new state, `recordGeneratedChanges` publishes the prepared applied-change batch.

`review.Store.AppendBatch` first validates and encodes every request. It then:

```text
reads refs/ddocs/review
-> writes one commit per event, chained in request order
-> includes event.json and optional before/after blobs
-> advances refs/ddocs/review once with compare-and-swap
```

The reference update exposes either the complete event chain or none of it. Unreferenced objects written during a failed compare-and-swap are not part of visible review history.

When another process changes the review reference concurrently, the store rebuilds the chain from the new head and retries, up to three attempts. Exhaustion returns `review history changed during append`.

### Review publication failure

If append fails, Demon Docs immediately calls `RollbackGenerated` for the authored source batch.

- Successful rollback leaves authored files in their recorded old state and the review reference unchanged.
- Failed or refused rollback returns both errors. Some source files may remain in generated state, and no new review events are visible.

Link-state refresh and publication do not run after review publication failure.

## Suppression staging

Only after review publication succeeds are the generated suppressions assigned to `plan.Suppressions`.

At this point suppressions are still memory-only. They become durable only when the later link-state publication writes `write/<source-file-id>` records under `refs/ddocs/state`.

This ordering means a failed review append cannot leave durable suppressions for source writes that were rolled back.

## Generated-source refresh

Demon Docs re-reads every generated source and parses its current link occurrences. Refresh verifies that each stored outgoing link can be found in ordinal order with its expected current target text.

Detached worker results contain:

- refreshed byte offsets;
- line and column;
- syntax fields;
- raw path and suffix;
- source fingerprint;
- source size; and
- source modification time.

Results merge into the plan in rewrite order to keep state deterministic.

Refresh does not republish review history and does not attempt source rollback on failure. At this stage the authored bytes and review events are already durable and agree about the applied change.

### Refresh failure

If an expected occurrence cannot be found or a rewritten source cannot be read, fingerprinted, or stated:

- authored source rewrites remain;
- applied-change events remain visible;
- `refs/ddocs/state` is not advanced; and
- suppressions are not made durable by this run.

A later link reconciliation parses current source bytes and rebuilds the state projection. The failure is recoverable through convergence, not by silently erasing the recorded applied change.

## Link-state publication

After refresh succeeds, `links.Save` constructs the complete desired private-state record set:

```text
meta/state
file/<file-id>
path/<path-key>
source/<source-file-id>
incoming/<target-file-id>/<source-file-id>
write/<source-file-id>
```

The state transaction deletes obsolete records in the owned namespaces and writes desired records in sorted name order. `internal/ddrepo` rewrites only dirty shards and advances `refs/ddocs/state` with compare-and-set against the transaction's base root.

One successful state-reference update publishes current identities, paths, outgoing links, incoming groups, refreshed source metadata, and pending suppressions together.

After publication, legacy `.ddocs/files.json` and `.ddocs/links.json` are removed when present.

### State publication failure

If the private state transaction conflicts or otherwise fails:

- authored source rewrites remain;
- review events remain;
- the previous `refs/ddocs/state` root remains authoritative;
- new suppressions are not durable; and
- later reconciliation must converge state from current authored files and review history.

The implementation does not roll authored content back after state-publication failure. Doing so would require appending a compensating review event or rewriting visible history; neither is part of this boundary.

## Atomicity matrix

| Surface | Local guarantee | Cross-surface guarantee |
| --- | --- | --- |
| One authored source | Same-directory atomic replacement plus hash verification. | None by itself. |
| Authored rewrite batch | All sources preflight before first write; attempted writes roll back on filesystem failure where current hashes permit. | Review and state are not included. |
| Review event batch | One compare-and-swap exposes the full commit chain or none. | Authored files are rolled back if review append fails. |
| Link-state projection | One private root-reference update publishes all owned records. | Authored files and review history are not rolled back if state publication fails. |
| All three surfaces | Ordered publication and targeted compensation. | No single atomic transaction spans them. |

## Recovery expectations

### Source changed before apply

Re-run reconciliation. The source enters the external-edit path; no generated content from the stale plan is written.

### Filesystem apply failed but rollback succeeded

Correct the filesystem error and re-run. No review or state publication occurred.

### Filesystem apply and rollback both failed

Inspect the joined error and affected sources. Preserve any newer user content. Run `ddocs check -l` after manual resolution to obtain a current plan.

### Review append failed and source rollback succeeded

Retry after resolving review-store contention or corruption. No applied-change event or authored rewrite from the failed batch remains.

### Review append failed and rollback was refused

Treat current authored content as authoritative until inspected. The generated change is not recorded in visible review history. Resolve the file manually, then reconcile.

### Refresh or state publication failed

Do not undo the authored change merely because private state is stale. The review event records the completed rewrite. Re-run link reconciliation so current source contents can refresh and publish the graph.

## Invariants and safety boundaries

- Every source in a batch must match its planned old bytes before the first replacement.
- Each source must still match immediately before its own replacement.
- Only exact constructor-owned new bytes are written.
- Successful replacement is verified by content hash.
- Rollback never overwrites content matching neither recorded old nor recorded new state.
- Review events are visible as one complete batch.
- Review publication failure triggers guarded source rollback.
- Refresh and state publication occur only after review history records the rewrite.
- Pending suppressions become durable only with link-state publication.
- Worker concurrency cannot change plan, event, suppression, refresh, or diagnostic order.
- No documentation may describe the three surfaces as one atomic transaction.

## Code map

- `internal/filetxn/` — shared rewrite model, parallel batch preflight, atomic replacement, verification, and guarded rollback.
- `internal/links/rewrite.go` — link transformation validation, adaptation to the shared transaction, and link-specific suppression construction.
- `internal/links/rewrite_transaction.go` — link-facing guarded rollback adapter.
- `internal/links/apply.go` — publication sequence, review-failure compensation, refresh, and state handoff.
- `internal/links/review_record.go` — run/change/transformation construction and review append requests.
- `internal/links/state.go` — desired link-state projection and suppression records.
- `internal/filetxn/replace_unix.go` and `replace_windows.go` — shared platform replacement seam.
- `internal/review/store_batch.go` — review event preparation, commit-chain creation, compare-and-swap, and retries.
- `internal/ddrepo/transaction.go` — sharded state-root transaction and conflict detection.
- `internal/watch/` and `internal/links/suppression.go` — later suppression consumption; not publication ownership.

## Tests

Focused coverage includes:

- `internal/filetxn/apply_test.go` — shared batch preflight, apply, rollback, and newer-content refusal.
- `internal/links/rewrite_test.go` — transformation ranges, exact bytes, and source preservation.
- `internal/links/rewrite_concurrency_test.go` — complete preflight barrier and suppression order.
- `internal/links/rewrite_transaction_test.go` — rollback after write failure and refusal over changed content.
- `internal/links/review_record_transaction_test.go` — source rollback when review publication fails.
- `internal/review/store_batch_test.go` — complete history-chain publication and batch preflight.
- `internal/review/store_test.go` — retained before/after blobs and visible review history.
- `internal/ddrepo/repository_test.go` — private-state transactions and conflict behavior.
- `internal/links/reconcile_test.go` and integration tests — end-to-end link-state convergence.

Run:

```bash
go test ./internal/filetxn ./internal/links ./internal/frontmatter ./internal/review ./internal/ddrepo -count=1
```

## Related docs

- [Link Reconciliation State Machine](link-reconciliation-state-machine.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Review Lifecycles](review-lifecycles.md)
- [Review Ledger](review-ledger.md)
- [Repository State and Transactions](repository-state-and-transactions.md)
- [Reconciliation Command Lifecycle](reconciliation-command-lifecycle.md)
- [Watcher and Automation](../operations/watcher-and-automation.md)

## Notes

The ordered compensation model is intentionally narrower than a distributed transaction. Current source bytes and append-only review history are preserved rather than hidden behind an unsupported claim of global atomicity.
