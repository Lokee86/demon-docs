---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7f2b-a20a-edfaa2aa0de6
document_type: general
policy_exempt: false
summary: This document defines the implemented state transitions for suggestions, candidate decisions, applied changes, undo, repair blocks, stale decisions, and stale blocks.
---
# Review Lifecycles

Parent index: [Architecture](./README.md)

## Purpose

This document defines the implemented state transitions for suggestions, candidate decisions, applied changes, undo, repair blocks, stale decisions, and stale blocks.

## Overview

The review system records durable human decisions and generated changes without asking an algorithm to remove authored relationships or silently repeat rejected proposals.

It has three related but distinct lifecycles:

```text
suggestion lifecycle
applied-change and undo lifecycle
repair-block lifecycle
```

All durable events are appended under `refs/ddocs/review`. Current statuses are projections produced by replaying that append-only history against current evidence.

## Code root

```text
internal/review/
internal/app/review_*.go
internal/links/review_*.go
internal/codemap/insert.go
```

## Responsibilities

The review boundary owns:

- stable suggestion, relation, candidate, repair, transformation, change, run, decision, and event identities;
- evidence fingerprints and stale-decision detection;
- issue-level and candidate-level decline decisions;
- reconsider decisions;
- repair blocks and unblocks;
- append-only applied-change events with before/after blobs;
- change grouping by reconciliation run, source file, and transformation;
- undo eligibility by depth and age;
- whole-change, individual-repair, and whole-run undo construction;
- compare-and-swap publication of review event batches; and
- replay of current decision and repair policy.

## Does not own

It does not own:

- producing link or codemap candidates;
- selecting a candidate without a user command;
- deciding that an existing authored codemap link is irrelevant;
- bypassing generated-rewrite hash guards;
- arbitrary selective reversion after later user edits;
- normal user Git history;
- link-state publication; or
- cross-subsystem command rollback.

## Durable event model

The review reference points to a linear Git commit chain. Each commit contains:

```text
event.json
optional before blob
optional after blob
```

Two event types are implemented:

- `decision` — decline, reconsider, block, or unblock;
- `change` — one generated source-file transition, including undo transitions.

Events are append-only. Current state is obtained by reading history and replaying relevant decisions from oldest to newest.

## Identity model

### Suggestion identity

A suggestion includes:

- suggestion kind (`link_repair` or `codemap_link`);
- relation key identifying the underlying relationship;
- fingerprint identifying current issue evidence;
- source identity and path;
- optional link occurrence identity; and
- ordered candidates with their own fingerprints.

The relation key remains stable across evidence changes. The fingerprint changes when the material evidence for the proposed relationship changes.

### Candidate identity

Candidate decline replay uses:

```text
relation key + candidate target
```

and compares the stored candidate fingerprint with the current candidate fingerprint.

### Repair identity

A deterministic repair transformation derives a relation key and repair fingerprint from:

- source file identity;
- relation token, normally the stable link relationship token;
- old destination;
- new destination; and
- target file identity.

The transformation ID is derived from that repair relation and fingerprint.

### Change and run identity

Each generated source rewrite becomes one change ID. All rewrites published by one reconciliation batch share one run ID. A change may contain multiple transformations for the same source file.

This gives three addressable levels:

```text
run
-> source-file change
   -> individual transformation/repair
```

## Suggestion lifecycle

### Detection

Candidate-producing subsystems construct current suggestions from current evidence.

```text
current issue + candidates
-> stable relation key
-> current issue fingerprint
-> stable suggestion ID
-> candidate fingerprints
```

The review policy is then applied to produce current status and candidate flags.

### Pending

A suggestion is `pending` when no current issue-level decline or blocking state governs it.

A pending candidate can be selected unless that candidate is individually declined.

### Issue-level decline

`ddocs suggestions decline SUGGESTION` appends a `decline_issue` decision containing:

- relation key;
- current issue fingerprint;
- suggestion ID and snapshot;
- optional reason; and
- decision time.

Replay behavior is:

```text
same relation + same issue fingerprint -> declined
same relation + changed issue fingerprint -> stale
```

A declined issue is suppressed from normal pending review but remains visible through declined/history surfaces.

### Candidate-level decline

`ddocs suggestions decline SUGGESTION CANDIDATE` appends `decline_candidate` with:

- relation key;
- candidate target;
- current candidate fingerprint;
- suggestion snapshot; and
- optional reason.

Replay behavior is:

```text
same target + same candidate fingerprint -> candidate declined
same target + changed candidate fingerprint -> candidate stale
```

Candidate decline does not decline the entire issue. Other candidates remain selectable.

### Stale decisions

A stale issue or candidate means a prior decision exists for the same relationship, but current evidence differs materially.

Staleness does not silently apply or permanently suppress the changed proposal. It returns the relationship to explicit review with the prior decision visible as context.

A stale issue can be selected because it is no longer governed by the old exact fingerprint. A candidate still marked declined cannot be selected until reconsideration.

### Reconsideration

`ddocs suggestions reconsider SUGGESTION` appends `reconsider` for the relation key.

Policy replay removes:

- the issue-level decline for that relation; and
- all candidate-level declines under that relation.

The next current projection returns the suggestion to `pending` unless another independent control applies.

Reconsideration does not delete prior decisions. It appends a later event that supersedes them during replay.

### Selection

Selection requires a current suggestion and one un-declined candidate. Declined and blocked suggestions must first be reconsidered or unblocked.

For a link suggestion:

```text
select candidate
-> prepare selection link plan
-> convert candidate to concrete target repair
-> create GeneratedRewrite with selection=user
-> normal ApplyAndSave publication path
```

For a codemap suggestion, selection inserts the chosen missing relationship through the codemap insertion seam and then uses the same generated-rewrite publication machinery.

There is no durable `accepted` suggestion state. The durable result is an applied-change event whose `OriginSuggestionID` points back to the selected suggestion.

## Link-repair suggestion projection

Link records project into review suggestions when their reconciliation status is:

- `ambiguous` — normal pending/declined/stale suggestion policy applies;
- `blocked` — suggestion status is `blocked`; or
- `stale_block` — suggestion status is `stale`.

A blocked deterministic repair is not the same as an ambiguous candidate set. It represents one known repair suppressed by an active repair block.

## Applied-change lifecycle

### Preparation

Before generated source writes, one `review.Change` is prepared per source rewrite. It records:

- change and run IDs;
- suggestion kind;
- selection mode;
- optional originating suggestion;
- source identity and repository-relative path;
- before and after hashes;
- exact before and after blobs;
- individual transformations;
- related targets; and
- application time.

### Publication

After source files have been replaced and verified, all changes in the batch are appended to review history with one compare-and-swap reference update.

Callers observe the complete chain or none of it. Concurrent reference movement causes the chain to be rebuilt from the new head and retried up to three times.

If publication fails, the caller uses guarded generated-rewrite rollback. Review publication never bypasses source hash checks.

### Inspection

`ddocs changes` and related subcommands project change events by run, file, kind, selection mode, source, transformations, and related targets. Inspection reads history; it does not alter authored files or policy.

### Undo changes are changes

Undo does not delete the original event. It appends a new `change` event with:

- `SelectionUndo`;
- `UndoOf` identifying the original change; and
- optional `UndoRepairID` for a selective repair undo.

The original and compensating transitions therefore remain auditable.

## Undo eligibility

Eligibility is checked against current review history and configuration.

### Depth

`review.undo_depth` counts recent original changes, excluding undo events. A value of `0` disables undo. A positive value limits eligibility to that many recent original changes.

### Age

`review.undo_max_age_days` rejects an original change older than the configured duration. A non-positive value leaves age unrestricted.

These controls affect whether before-state may be restored. They do not remove history or retained blobs.

### Current-content guard

Before any undo is planned, the current source must match the original change's recorded after hash. If later edits changed the file, undo refuses to overwrite them.

## Whole-change undo

`ddocs changes undo CHANGE` restores the complete retained before blob for one source-file change.

```text
load original event and blobs
-> verify eligibility
-> resolve current source path
-> require current hash == recorded after hash
-> construct GeneratedRewrite(current -> before)
-> apply source rewrite
-> append undo change event
```

If review append fails, the source rewrite is rolled back through the same guarded publication compensation used by generated repairs.

## Individual-repair undo

`ddocs changes undo CHANGE --repair REPAIR` reverses one transformation while preserving other transformations from the same original source change.

`BuildUndoData` decodes the recorded after blob, adjusts offsets for earlier transformation length changes, locates the requested transformation's current new text, and substitutes its old text.

Selective undo fails when:

- the repair ID is absent;
- retained after data is unavailable;
- the calculated range is invalid; or
- the recorded new text no longer matches the after-state bytes.

The resulting generated rewrite still requires the current whole file to match the recorded original after hash before application.

## Whole-run undo

`ddocs changes undo-run RUN` targets every original change in one reconciliation run.

Before the first authored write, it preflights the entire run:

- at least one original change must exist;
- each change must be undo-eligible;
- each current source must be resolvable and readable;
- each current source hash must equal its recorded after hash;
- one run cannot contain multiple targeted changes for the same current path; and
- each reverse rewrite and review append request must be constructible.

Only after all files pass does it call `ApplyGenerated` for the complete rewrite batch. The complete undo-event batch is then published with one review-reference update.

If review publication fails, the full applied undo batch is passed to guarded rollback.

## Repair-block lifecycle

### Block creation

Undo can optionally be followed by `--block`. The application appends one or more `block_repair` decisions for the reverted transformation relationships.

A block records:

- repair relation key;
- exact repair fingerprint;
- originating change ID;
- optional reason; and
- decision time.

Blocking is a separate review event from undo. It prevents unchanged deterministic reconciliation evidence from immediately reapplying the same repair.

### Active block

During later link reconciliation:

```text
same repair relation + same fingerprint
-> active block
-> link status blocked
-> no generated rewrite
```

### Stale block

```text
same repair relation + changed fingerprint
-> stale block
-> link status stale_block
-> explicit review required
```

Changed evidence is neither silently repaired nor permanently suppressed by the old block.

### Unblock

The changes control command appends `unblock_repair` for the repair relation. Policy replay removes the active repair block. Later reconciliation evaluates the deterministic repair normally.

Unblock does not erase the original block decision.

## Decision and block replay

`review.LoadPolicy` reads visible review history newest-first, then replays decisions oldest-to-newest so later decisions supersede earlier ones.

The current in-memory policy contains maps for:

```text
issue declines by relation key
candidate declines by relation key + target
repair controls by repair relation key
```

Replay effects are:

| Decision action | Projection effect |
| --- | --- |
| `decline_issue` | Set the current issue decline. |
| `decline_candidate` | Set the current target-specific candidate decline. |
| `reconsider` | Remove issue decline and all candidate declines for the relation. |
| `block_repair` | Set the current repair control. |
| `unblock_repair` | Remove the current repair control. |

History remains append-only regardless of current projection.

## Review-store publication

`AppendBatch` validates and encodes every request before it reads and advances the review reference.

For one attempt:

```text
read refs/ddocs/review
-> use current head as first parent
-> write one event commit per request in order
-> advance reference once with compare-and-swap
```

Each commit is authored locally as `Demon Docs <ddocs@local>` and contains its event plus optional before/after blobs.

A failed preflight publishes no visible event. A compare-and-swap conflict leaves the newly written objects unreachable, retries from the current head, and does not expose a partial batch.

## Failure and recovery

### Current suggestion missing

Selection or decline against an ID no longer in the current projection fails. Re-run the suggestion listing and use current evidence. Reconsider can fall back to a historical suggestion snapshot when needed.

### Declined candidate selection

Selection refuses an individually declined candidate. Reconsider the relation first.

### Source changed after applied repair

Undo refuses because the current hash no longer matches the recorded after state. Resolve manually or use normal source-control tools; the review ledger does not perform an arbitrary merge.

### Undo outside configured limits

The command fails without modifying the source or history. Adjusting policy may make a retained event eligible, provided current content still matches.

### Source rewrite succeeds but review append fails

Guarded rollback restores source bytes only when they still match the generated after state. A refusal protects newer content and returns a joined error.

### Optional block append fails after undo event

The undo source rewrite and undo change event are already durable. The requested repair block may be absent. Re-run the explicit block control after inspecting current history.

### Review reference conflict

Append retries up to three times. Continued movement returns an error; no partial event batch becomes visible from the failed attempt.

## Invariants and safety boundaries

- Existing authored codemap links are never proposed for removal.
- Same evidence keeps an explicit decline or block active.
- Changed evidence makes the prior control stale rather than silently inheriting it.
- Candidate decline does not decline unrelated candidates.
- Reconsideration is append-only and relation-scoped.
- Selection enters the normal generated-rewrite path.
- Every applied source rewrite is represented by a change event when review publication succeeds.
- Undo cannot overwrite content changed after the recorded repair.
- Whole-run undo preflights every target before the first write.
- Undo events never erase original history.
- Review append batches are all-or-none at the reference boundary.
- Review history is separate from the user's normal Git history.

## Extension rules

When adding a suggestion kind, decision action, transformation kind, or review command:

1. define stable relation and evidence fingerprints;
2. specify which evidence change makes prior decisions stale;
3. preserve issue-level versus candidate-level scope;
4. route source mutations through generated-rewrite guards;
5. record exact before/after hashes and blobs when undo is promised;
6. define replay precedence with existing decisions;
7. keep append-only history and batch publication semantics;
8. document failure after source mutation but before later control events; and
9. add focused policy, store, CLI, and integration tests.

## Code map

- `internal/review/model.go` — suggestion, candidate, transformation, change, decision, event, status, and selection models.
- `internal/review/fingerprint.go` — stable relation, suggestion, candidate, repair, and transformation identities.
- `internal/review/policy.go` — decision replay, stale matching, reconsideration, and repair controls.
- `internal/review/store.go` — review store opening and history traversal.
- `internal/review/store_batch.go` — event validation, commit creation, before/after blobs, compare-and-swap, and retries.
- `internal/review/undo.go` — depth/age eligibility and whole/selective undo data construction.
- `internal/app/review_suggestions.go` and `review_suggestion_actions.go` — listing, selection, decline, and reconsider commands.
- `internal/app/review_changes.go` — change-history projection.
- `internal/app/review_undo.go` — change, repair, and run undo orchestration.
- `internal/app/review_controls.go` — block and unblock command integration.
- `internal/links/review_suggestions.go` — link record to suggestion projection.
- `internal/links/review_selection.go` — selected link candidate conversion.
- `internal/links/review_record.go` — applied-change and transformation construction.
- `internal/codemap/insert.go` — selected codemap candidate insertion.

## Tests

Focused coverage includes:

- `internal/review/store_test.go` — history, blobs, decline persistence, staleness, and reconsideration.
- `internal/review/store_batch_test.go` — complete batch publication and failed-preflight visibility.
- `internal/review/undo_test.go` — eligibility and selective transformation reversal.
- `internal/links/review_integration_test.go` — ambiguous, blocked, and stale-block link behavior.
- `internal/links/review_selection_test.go` — selected link candidate transitions.
- `internal/links/review_record_transaction_test.go` — review failure and source rollback.
- `internal/app/review_cli_test.go` — public suggestion, change, decline, selection, block, and undo workflows.
- `internal/app/review_undo_transaction_test.go` — whole-run preflight and transactional compensation.

Run:

```bash
go test ./internal/review ./internal/links ./internal/app ./internal/codemap -count=1
```

## Related docs

- [Review Ledger](review-ledger.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [Link Reconciliation State Machine](link-reconciliation-state-machine.md)
- [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md)
- [Repository State and Transactions](repository-state-and-transactions.md)
- [Configuration Reference](../reference/configuration.md)
- [Codemap Missing-Link Evidence](../research/codemap-evidence.md)

## Notes

Current statuses are projections, not mutable rows. The durable source of truth is the append-only event chain plus current repository evidence.
