---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7bda-afb0-bc282fdd47c1
document_type: general
policy_exempt: false
summary: This document defines the implemented watcher scheduling, debounce, single-run ownership, cross-watcher serialization, error propagation, cancellation, and self-write convergence contracts.
---
# Watch Scheduler and Reconciliation Serialization

Parent index: [Architecture](./INDEX.md)

## Purpose

This document defines the implemented watcher scheduling, debounce, single-run ownership, cross-watcher serialization, error propagation, cancellation, and self-write convergence contracts.

## Overview

Demon Docs watchers separate filesystem observation from reconciliation execution.

Filesystem events do not run reconciliation directly. They mark work as pending. A scheduler waits for the debounce window, grants one run at a time, clears the current pending batch, and invokes the selected reconciliation function. Events that arrive while a reconciliation is running accumulate as new pending work and cause a later follow-up run.

When forward indexes or links are watched together with reverse indexes, two observers may remain active concurrently, but their reconciliation functions share one run lock. This prevents the two mutation pipelines from writing repository files at the same time.

The scheduler owns timing and run admission. It does not own filesystem scope, reconciliation planning, file replacement, link-state transactions, reverse-index rendering, repository-demon leases, or process lifetime.

## Primary ownership

The scheduling boundary owns:

- recording that one or more relevant events occurred;
- remembering the time of the most recent event;
- enforcing the configured debounce interval;
- admitting at most one scheduler-owned reconciliation run at a time;
- clearing the admitted event batch before execution;
- retaining events that arrive during execution as later pending work;
- returning reconciliation errors to the watcher loop;
- accepting an optional shared run lock for serialization with another watcher; and
- allowing context cancellation and observer shutdown to terminate the surrounding watch loop cleanly.

## Explicit non-ownership

The scheduling boundary does not decide:

- whether an event is relevant;
- which repository paths are observed;
- which reconciliation features are selected;
- whether generated writes are suppressed;
- how indexes, links, or reverse indexes are planned;
- how authored files are replaced or rolled back;
- how private state is published;
- whether the process runs in the foreground or under the repository demon; or
- whether a failed reconciliation should later be retried by a new process.

Those responsibilities remain with `internal/watch`, `internal/reverseindex`, the selected reconciliation packages, and the calling application lifecycle.

## Scheduler state model

`internal/watch.Scheduler` maintains five pieces of mutable state under a mutex:

```text
pending   number of relevant events admitted since the last run began
running   whether this scheduler currently owns a reconciliation run
last      timestamp of the most recent admitted event
debounce  required quiet interval after the most recent event
run       reconciliation callback
```

The effective states are:

```text
idle
  pending == 0
  running == false

waiting
  pending > 0
  running == false
  quiet interval not yet satisfied

ready
  pending > 0
  running == false
  quiet interval satisfied

running
  running == true
  admitted pending batch has been reset to zero
  later events may increase pending again
```

There is no dedicated goroutine inside `Scheduler`. The outer watcher loop periodically calls `RunIfPending`.

## Event admission

`MarkChanged` performs one atomic state update:

```text
lock
pending++
last = now()
unlock
```

Every admitted event resets the quiet-period reference point. The pending count records that work exists; it is not used to run reconciliation once per event.

A burst of many events therefore becomes one run after the repository has remained quiet for the configured debounce interval.

## Run admission

`RunIfPending` refuses a run when any of these conditions apply:

- another run owned by the same scheduler is active;
- no event is pending; or
- the configured debounce interval has not elapsed since the most recent event.

When a run is admitted, the scheduler performs this transition under the mutex:

```text
running = true
pending = 0
```

It then releases the mutex before invoking the reconciliation callback. Filesystem event processing can continue to call `MarkChanged` while reconciliation runs.

After the callback returns, the scheduler reacquires the mutex and sets:

```text
running = false
```

The callback result is returned to the watcher loop. A non-nil error normally terminates the watch operation. Transient filesystem races caused by files moving or changing during a scheduled run are the bounded exception: the watcher marks another full run pending and waits for the next quiet interval rather than applying the stale plan.

## Follow-up behavior

Clearing `pending` before invoking the callback is intentional.

If no relevant events arrive during the run, the scheduler returns to `idle`.

If one or more relevant events arrive during the run, `pending` becomes non-zero while `running` remains true. Calls to `RunIfPending` refuse admission until the current run ends. A later ticker call admits one follow-up run after the new event batch satisfies the debounce interval.

This contract provides convergence without overlapping scheduler-owned runs:

```text
event burst A
-> run A begins
-> event burst B arrives during run A
-> run A ends
-> debounce B
-> one follow-up run B
```

The scheduler does not guarantee that every intermediate filesystem state is reconciled. It guarantees eventual processing of the latest observed state when events stop arriving and no run fails.

## Initial reconciliation ordering

The normal watcher performs one stable reconciliation synchronously before creating the filesystem observer.

```text
construct selected reconciliation callback
-> run callback
-> if a source moved or changed during planning/application, discard the stale plan
-> wait a bounded quiet delay and retry until success or cancellation
-> return immediately for --once
-> otherwise load ignore policy
-> create observer
-> add watched paths
-> start event loop
```

This ordering prevents an observer from reacting to generated writes produced by the initial baseline before suppression and state have converged. It also means observer-construction failure can occur only after the initial reconciliation has completed successfully. A daemon started in the middle of a folder migration therefore waits for one stable plan rather than exiting on a generated-rewrite hash mismatch.

The initial run is not admitted through `Scheduler`; it directly invokes the same reconciliation callback through the transient-race retry boundary.

## Polling cadence

The base watcher checks scheduler readiness with a ticker.

- With no positive debounce, the ticker interval is 100 milliseconds.
- With a positive debounce, the interval is half the debounce interval, capped at 250 milliseconds.

The ticker is a readiness poll, not the debounce source of truth. `RunIfPending` compares the current time with the timestamp of the most recent admitted event.

This may start a run slightly after the exact debounce boundary, bounded by the ticker interval.

## Selected-feature execution

The base reconciliation callback can run documentation indexes, frontmatter, document-body format, links, or any selected combination. The application owns their ordering; the scheduler only serializes the callback.

The order is fixed:

```text
link reconciliation or identity tracking, when selected
-> prepare missing generated indexes for policy input
-> frontmatter, when selected
-> document-body format, when selected
-> reverse-index application, when selected in the base callback
-> final folder-index convergence, when selected
-> refresh link state from final authored bytes
-> output summary and individual diagnostics
```

One callback execution owns that full selected sequence. The scheduler does not interleave another base run between forward-index and link application.

Reverse-index watching is a separate observer and execution loop. Mixed watch mode coordinates it through a shared lock rather than merging it into the base scheduler.

## Shared cross-watcher run lock

When reverse indexes are selected together with documentation indexes, frontmatter, document format, or links, `internal/app.runSelectedWatch` creates:

- one cancellation context shared by both watcher loops;
- one synchronized output writer;
- one `sync.Mutex` used as a shared reconciliation run lock; and
- an error channel with one result slot per watcher.

The base watcher receives the lock through `RootSelectedWithRunLock`. The reverse watcher receives the same lock through `reverseindex.WatchWithRunLock`.

Each reconciliation callback acquires the lock immediately before planning and applying its own changes and releases it only after that callback completes.

The invariant is:

```text
base reconciliation callback
and
reverse-index reconciliation callback
never execute concurrently in the same mixed-watch process
```

The observers, event queues, debounce mechanisms, and watch-scope refresh work may still run concurrently. Only the mutation-capable reconciliation callbacks are serialized.

The shared lock does not provide a transaction across the two pipelines. A base run may complete, release the lock, and then a reverse-index run may fail. Earlier writes are not rolled back as one combined operation.

## Reverse-index scheduler differences

The reverse-index watcher uses a resettable `time.Timer` rather than `internal/watch.Scheduler`.

Each relevant event stops and resets the timer. When the timer fires, the reverse-index reconciliation callback runs under the optional shared run lock. After a successful run, the watcher requests a refresh of discovered watch directories.

The two watcher implementations share these behavioral guarantees:

- an immediate initial run;
- debounce after the most recent relevant event;
- no overlapping mutation callback when the shared run lock is present;
- surfaced callback errors; and
- clean context cancellation.

They do not share scheduler state or pending counters.

## Output serialization

Mixed watch mode passes both watcher loops a synchronized writer. Each `Write` call holds a writer mutex so status and diagnostic writes from the two goroutines do not race at the byte level.

This protects output integrity only. It does not establish run ordering; the shared run lock owns mutation serialization.

## Self-write convergence

Link application can generate filesystem events for Markdown files that Demon Docs just rewrote.

Before normal event relevance processing, the link-enabled watcher asks the link subsystem to consume a pending suppression for the event path.

A suppression is accepted only when the observed file still matches the generated write expectation. A matching event is skipped. A mismatch invalidates the suppression and the event continues through normal relevance processing.

The scheduler therefore does not contain self-write logic. Its contract assumes that event intake has already filtered matching generated writes while preserving concurrent user changes.

## Error propagation

The watch loop treats these as terminal errors:

- non-transient selected reconciliation planning or application failure;
- external-watch addition failure;
- ignore-policy loading or reload failure;
- dynamic directory-watch addition failure;
- suppression-consumption failure;
- observer creation failure;
- observer error-channel values other than event-buffer overflow; and
- reverse-index refresh or reconciliation failure.

A transient filesystem race discards the stale plan and retries after the repository settles. An `fsnotify` event-buffer overflow means individual event detail was lost, not that the observer or repository is unusable; the watcher logs the overflow, marks a complete reconciliation pending, and continues. Other terminal errors return to the foreground command or repository-demon owner.

Closed observer event or error channels terminate the base watcher cleanly when no explicit error is available.

## Cancellation and shutdown

Context cancellation is the normal clean-stop signal.

The base watcher returns nil when its context is done. Deferred cleanup stops the ticker and closes the observer.

The reverse-index watcher cancels its directory-refresh worker, stops using the timer channel, and closes and drains the fsnotify watcher. Draining prevents shutdown from blocking when the platform observer still has buffered events or errors.

In mixed mode, the first watcher result cancels the shared context. The application waits for the second watcher to exit before returning. The first non-nil result takes precedence; otherwise the second result is returned.

## Invariants

The scheduler and serialization design must preserve these invariants:

- Relevant event bursts produce bounded reconciliation work rather than one run per event.
- One `Scheduler` never invokes its callback concurrently with itself.
- Events arriving during a run are not discarded.
- Mixed base and reverse reconciliation callbacks never overlap.
- One stable initial reconciliation completes before observer creation; transient move races are retried until success or cancellation.
- Event-buffer overflow schedules a complete reconciliation instead of terminating the watcher.
- Generated-write suppression cannot hide content that no longer matches the generated expectation.
- Reconciliation errors are surfaced and terminate the current watch lifecycle.
- Cancellation does not leave an observer or refresh worker intentionally running.
- Serialization does not imply a cross-subsystem transaction or rollback boundary.

## Extension rules

### Adding a selected base feature

A new feature executed by the base watcher must be added inside the single base reconciliation callback if it mutates the same authored repository surface and requires sequential ordering with forward indexes or links.

Document its fixed order relative to existing features. Do not start an uncoordinated mutation goroutine from event handling.

### Adding a separate watcher

A separate observer is justified only when its watch scope or refresh lifecycle is materially independent. If its reconciliation can write repository files, mixed mode must share the same run lock or use a stronger explicit transaction seam.

### Changing debounce behavior

Preserve the latest-event quiet-period rule and follow-up behavior for events arriving during a run. Tests must cover both burst coalescing and mid-run event admission.

### Adding retry behavior

Retries must not be added inside the scheduler without defining idempotency, source preconditions, private-state publication, diagnostic repetition, and cancellation. Current retry ownership is deliberately narrow: initial and scheduled reconciliation may retry only recognized transient filesystem races, and observer overflow may request one complete rebuild because event detail was lost. Other errors still end the watch lifecycle.

## Verification

Focused verification:

```bash
go test ./internal/watch ./internal/reverseindex ./internal/app -count=1
```

Important contracts include:

- scheduler debounce and follow-up runs;
- base reconciliation serialization through the run lock;
- initial reconciliation before observer creation;
- clean cancellation;
- observer-error propagation and non-terminal event-buffer overflow recovery;
- transient initial-plan retry during active moves;
- generated-write convergence without self-write loops; and
- combined base and reverse watch coordination.

## Code map

Primary implementation:

- `internal/watch/scheduler.go` — pending count, latest-event timestamp, debounce admission, and single-run state.
- `internal/watch/watch.go` — immediate initial run, base event loop, scheduler polling, suppression intake, cancellation, and optional run lock.
- `internal/watch/features.go` — selected-feature scope and external target watch helpers.
- `internal/app/reverse_index.go` — mixed-watch goroutines, synchronized output, shared run lock, cancellation, and result joining.
- `internal/reverseindex/watch.go` — reverse-index debounce timer, watch refresh worker, optional run lock, and draining shutdown.
- `internal/links/suppression.go` — generated-write event suppression consumed before scheduling.

Focused tests:

- `internal/watch/watch_test.go`
- `internal/watch/watcher_contract_test.go`
- `internal/watch/filter_test.go`
- `internal/watch/external_test.go`
- `internal/reverseindex/watch_test.go`
- `internal/app/reverse_index_test.go`

Non-ownership boundaries:

- `internal/reconcile/` owns forward-index planning and application.
- `internal/links/` owns link planning, generated rewrites, private state, and suppression records.
- `internal/reverseindex/` owns reverse-index planning and application.
- `internal/demon/` owns detached lifecycle and repository-local demand leases.

## Related docs

- [Watcher and Automation](../operations/watcher-and-automation.md)
- [Dynamic Watch Scope](../operations/dynamic-watch-scope.md)
- [Repository Demon](../operations/repository-demon.md)
- [Repository Demon Lease Protocol](repository-demon-lease-protocol.md)
- [Reconciliation Pipeline](reconciliation-pipeline.md)
- [Markdown Link Reconciliation](markdown-link-reconciliation.md)
- [Reverse Index Architecture](reverse-indexes.md)

## Notes

The base and reverse watchers currently use different debounce implementations. Their required behavioral contracts are aligned, but they should not be described as one shared scheduler implementation.
