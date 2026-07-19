# Suggestions, Repairs, and Change History

Demon Docs keeps one repository-local review ledger for unresolved suggestions, applied repairs, user decisions, and undo history. The ledger is stored in the private `.ddocs/` Git object repository under `refs/ddocs/review`; it does not create commits in the user's normal Git history.

## State Model

A suggestion is an unresolved choice. A repair is a concrete change ready to apply. An applied repair becomes a change record.

```text
detected issue
├── deterministic repair
│   └── applied change
└── ambiguous suggestion
    ├── selected candidate
    │   └── repair
    │       └── applied change
    ├── declined candidate
    └── declined issue
```

Link repair remains automatic when exactly one deterministic target exists. Multiple plausible targets become `link_repair` suggestions. Codemap missing-link candidates always become `codemap_link` suggestions and are never inserted automatically.

Selecting a suggestion immediately converts the candidate into the normal hash-guarded repair path. There is no lasting accepted-suggestion state.

## Suggestion Commands

```bash
ddocs suggestions [FILE]
ddocs suggestions declined [FILE]
ddocs suggestions log [FILE]
ddocs suggestions show SUGGESTION
ddocs suggestions select SUGGESTION [CANDIDATE]
ddocs suggestions decline SUGGESTION [CANDIDATE] --reason "..."
ddocs suggestions reconsider SUGGESTION
```

`ddocs suggestions` regenerates current candidates from repository state and joins them with persisted decisions. A candidate may be selected by its displayed number or target path. Omitting the candidate is allowed when the suggestion has exactly one candidate.

Declining a candidate suppresses only that candidate. Declining without a candidate suppresses the entire issue. Decisions are keyed by the stable relationship and evidence fingerprint. The same evidence remains declined; materially changed evidence is shown as stale and may be reconsidered.

Existing authored codemap links are never presented as removal or irrelevance suggestions.

## Applied Change Commands

```bash
ddocs changes [FILE]
ddocs changes related FILE
ddocs changes show CHANGE
ddocs changes log [FILE]
ddocs changes undo CHANGE [--repair REPAIR] [--block] [--reason "..."]
ddocs changes undo-run RUN [--block] [--reason "..."]
ddocs changes block CHANGE [--repair REPAIR] [--reason "..."]
ddocs changes unblock CHANGE [--repair REPAIR]
```

Every generated rewrite records:

- the reconciliation run;
- repair kind and selection mode;
- source file identity and path;
- originating suggestion when applicable;
- before and after SHA-256 hashes;
- before and after file blobs;
- individual repair transformations; and
- related target identities and paths.

`changes related FILE` answers which source files Demon Docs rewrote because the named target moved or changed. `changes show` prints the stored transformation metadata and a unified before/after diff.

## Undo Granularity

Demon Docs supports three bounded undo levels:

1. reconciliation run;
2. one file change; and
3. one repair within a file change.

It deliberately does not attempt arbitrary historical selective reversal through later user edits. Use the repository's normal version-control system for that case.

Before any undo, Demon Docs verifies that the current file still matches the recorded after hash. Run-level undo preflights every affected file before writing any of them. A mismatch stops the operation rather than overwriting newer work.

Undo is recorded as a new change event. Existing history is never deleted or rewritten.

## Repair Blocks

Undo alone permits the same deterministic repair to be discovered again. `--block`, or a separate `changes block`, records a control decision for the exact source relationship and repair fingerprint.

An unchanged blocked repair is not applied. It remains visible as blocked. If the target relationship or evidence changes, the old block becomes stale; Demon Docs still does not apply it silently and instead surfaces it for review.

## Undo Configuration

```toml
[review]
undo_depth = 100
undo_max_age_days = 30
```

`undo_depth` limits how many recent non-undo changes are eligible for reversal. `0` disables undo and `-1` allows unlimited depth. `undo_max_age_days` limits eligibility by age; `0` disables the age limit.

Audit history remains inspectable even when a change is outside the configured undo window.

## Storage and Safety

The review ledger uses Git commits, trees, and blobs inside `.ddocs/`. Each event commit points to its predecessor and contains `event.json` plus optional `before` and `after` blobs. Concurrent appends advance `refs/ddocs/review` with compare-and-swap semantics.

Repairs and undo use the same atomic replacement machinery as link reconciliation. Source hashes are checked before writing, replacement is same-directory and atomic, and the resulting hash is verified afterward.

## Code map

- `internal/review/` — review models, fingerprints, Git history, policy replay, and undo construction.
- `internal/links/review_suggestions.go` — ambiguous link suggestions and user-selected repair conversion.
- `internal/links/review_record.go` — applied repair event recording.
- `internal/app/review_suggestions.go` — suggestion CLI.
- `internal/app/review_changes.go` — change inspection CLI.
- `internal/app/review_undo.go` — undo and repair controls.
- `internal/codemap/insert.go` — bounded authored codemap insertion for selected candidates.
