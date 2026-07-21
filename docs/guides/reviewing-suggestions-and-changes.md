---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-72df-b917-0041b8f11041
document_type: general
policy_exempt: false
summary: This guide reviews unresolved repair suggestions, records accept or decline decisions, inspects applied changes, and performs bounded undo or repair blocking.
---
# Reviewing Suggestions and Changes

Parent index: [Guides](./INDEX.md)

## Purpose

This guide reviews unresolved repair suggestions, records accept or decline decisions, inspects applied changes, and performs bounded undo or repair blocking.

## Overview

Demon Docs separates unresolved choices, persisted decisions, and concrete repairs. Ambiguous link targets appear through `ddocs suggestions`. Selecting a candidate converts it into the compatibility hash-guarded repair path and records the applied repair in the private review ledger.

Codemap recommendations are generated only by the explicit `ddocs codemaps` command family. They are reviewed through codemap command output and Git rather than being mixed into ordinary link suggestions or published as `ddocs changes` events.

Deterministic single-target link repairs remain automatic and are recorded through the normal generated-repair lifecycle.

## Prerequisites

- The repository is initialized and has usable `.ddocs/` state.
- Run from within the intended repository.
- Review the working tree before selecting or undoing a repair.

## List current suggestions

```bash
ddocs suggestions
ddocs suggestions FILE
ddocs suggestions show SUGGESTION
```

The list is regenerated from current repository state and joined with persisted decisions.

Inspect declined and historical decisions:

```bash
ddocs suggestions declined [FILE]
ddocs suggestions log [FILE]
```

## Select a candidate

```bash
ddocs suggestions select SUGGESTION [CANDIDATE]
```

A candidate may be identified by its displayed number or target path. The candidate may be omitted when only one choice exists.

Selection immediately creates and applies the concrete compatibility repair. There is no permanent accepted-suggestion state.

## Decline or reconsider

```bash
ddocs suggestions decline SUGGESTION [CANDIDATE] --reason "..."
ddocs suggestions reconsider SUGGESTION
```

Declining a candidate suppresses only that candidate. Declining without a candidate suppresses the whole issue.

Decisions are keyed by stable relationship and evidence fingerprint. Unchanged evidence remains suppressed. Materially changed evidence becomes stale and is shown for review rather than silently reappearing or remaining permanently hidden.

## Inspect applied changes

```bash
ddocs changes [FILE]
ddocs changes related FILE
ddocs changes show CHANGE
ddocs changes log [FILE]
```

`changes related FILE` finds source files rewritten because the named target moved or changed. `changes show` presents transformation metadata and a unified before/after diff.

## Undo

Supported granularity:

```text
one reconciliation run
one file change
one repair within a file change
```

Commands:

```bash
ddocs changes undo CHANGE [--repair REPAIR] [--block] [--reason "..."]
ddocs changes undo-run RUN [--block] [--reason "..."]
```

Undo requires the current file to match the recorded after hash. Run-level undo preflights every affected file before any write. Later user edits cause refusal rather than overwrite.

## Block or unblock a repair

```bash
ddocs changes block CHANGE [--repair REPAIR] [--reason "..."]
ddocs changes unblock CHANGE [--repair REPAIR]
```

A block prevents the exact repair fingerprint from being applied again. Changed relationship or evidence makes the old block stale and reviewable; it is not silently reused.

## Expected result

- Ambiguous choices remain explicit until selected.
- Declines remain suppressed while evidence is unchanged.
- Every applied normal generated repair is inspectable in `ddocs changes`.
- Explicit codemap fix output and Git history remain the audit surface for unified codemap rewrites.
- Undo creates a new history event rather than deleting prior audit history.
- Later authored edits are never overwritten by historical undo.

## Failure and recovery

### A suggestion no longer exists

Regenerate the current list. Repository state or evidence may have changed since the identifier was displayed.

### Undo is ineligible

Check configured depth and age limits. Audit history remains inspectable even when undo eligibility expires.

### Undo reports an after-hash mismatch

The file changed after the recorded repair. Use normal Git history or manually integrate the intended reversal; Demon Docs does not perform arbitrary historical selective reverts through later edits.

### A repair returns after undo

Use `--block` during undo or add a separate repair block when the exact deterministic repair should remain suppressed.

## Related docs

- [CLI Reference](../reference/cli.md)
- [Review Ledger Architecture](../architecture/review-ledger.md)
- [Codemap Missing-Link Evidence](../research/codemap-evidence.md)
- [Managing Codemaps](managing-codemaps.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Configuration Reference](../reference/configuration.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Missing-link suggestions never present an existing codemap entry as irrelevant. Optional confidence pruning is a separate explicit codemap execution policy and remains disabled by default.
