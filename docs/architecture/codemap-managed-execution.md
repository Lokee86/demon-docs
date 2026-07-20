---
author: brian
created: "2026-07-19"
document_id: e83f3463-0f5b-44d7-b67b-2a4321b8c854
document_type: general
policy_exempt: false
summary: This document defines the production codemap execution boundary that selects documents, adopts unified managed sections, filters deterministic additions through shared review policy, applies optional pruning, and publishes hash-guarded foreground rewrites.
---
# Codemap Managed Execution

Parent index: [Architecture](./README.md)

## Purpose

This document defines the production codemap execution boundary that selects documents, adopts unified managed sections, filters deterministic additions through shared review policy, applies optional pruning, and publishes hash-guarded foreground rewrites.

## Overview

Production codemap execution is an explicit foreground workflow:

```text
configured repository and docs root
-> explicit file or directory target scope
-> configured codemap heading recognition
-> existing-link extraction and repository corpus
-> production missing-link ranking
-> shared decline-policy filtering
-> optional existing-link pruning evaluation
-> unified managed-section reconciliation
-> exact before/after plan
-> hash-guarded transactional publication
```

The codemap is one artifact. Demon Docs does not maintain separate authored and generated lists after adoption. Existing links, newly discovered links, explanatory prose inside the section, and the codemap-specific markers belong to one managed section lifecycle.

The default policy is additive and conservative. Existing valid links remain even when the current algorithm does not rediscover them or ranks them below the permanent-link threshold. Removal based on confidence requires explicit repository configuration.

## Current implementation status

Existing configured codemap sections are fully supported by the public execution commands.

The internal `codemap.SectionSchema` seam for creating a missing schema-required section is implemented and tested. The application layer does not yet provide a repository file-type schema implementation to `codemaprun.Options.Schema`. Therefore the current public commands skip a file whose configured codemap section is absent. They do not yet create that missing section from repository configuration.

This distinction is important:

```text
implemented now
= recognize, adopt, inspect, check, and update existing configured sections

implemented internal seam
= accept schema placement and create a required missing section

not yet connected publicly
= resolve a repository file type and supply its codemap placement to the execution plan
```

See [Current Product Limitations](../limits/current-limitations.md) for the user-visible impact and removal condition.

## Code root

```text
internal/app/codemap_execute.go
internal/app/codemap_execute_scope.go
internal/app/codemap_execute_output.go
internal/codemaprun/
internal/codemap/managed.go
internal/codemap/managed_section.go
internal/codemap/managed_render.go
internal/codemaprecommend/
internal/review/
internal/filetxn/
internal/textio/
```

## Responsibilities

This boundary owns:

- explicit `codemap` and `codemaps` command execution;
- file-or-directory target selection beneath the configured docs root;
- configured heading overrides for one command;
- construction of one deterministic repository-wide codemap dataset and corpus;
- production recommendation generation for each selected document;
- application of persisted decline decisions and evidence fingerprints;
- evaluation of opt-in undiscovered-link and low-score removal policies;
- adoption of the complete matching section under codemap-specific markers;
- preservation of supported existing fenced or bullet rendering conventions;
- exact before/after planning;
- read-only check, inspect, and dry-run behavior; and
- hash-guarded transactional file publication for `fix`.

## Does not own

This boundary does not own:

- file-type schema definition or repository schema selection;
- creation of a new Markdown document;
- ordinary repository-local link repair;
- a universal semantic truth model for whether a human-authored link is useful;
- review-event storage and replay mechanics;
- benchmark labels or precision governance;
- reverse-index generation;
- generic `ddocs fix`, `ddocs check`, or `ddocs watch` reconciliation; or
- repository-demon scheduling.

Broken authored paths remain the responsibility of normal link maintenance. Codemap pruning is not a replacement for link repair.

## Public command boundary

The command family is:

```bash
ddocs codemaps fix [--root FILE_OR_DIRECTORY] [--dry-run]
ddocs codemaps check --root FILE_OR_DIRECTORY
ddocs codemaps inspect --root FILE_OR_DIRECTORY
```

`ddocs codemap ...` remains accepted as a singular compatibility alias for the canonical `ddocs codemaps ...` command family. The `demon` executable forwards non-demon commands to the shared application entry, so the same codemap commands are available there.

`fix` may omit `--root`; the configured documentation root becomes the target directory. `check` and `inspect` require an explicit root so a diagnostic command cannot accidentally scan every managed document.

Repeated `--heading TEXT` flags replace the configured heading set for that invocation. They do not append to it.

## Target scope resolution

A target root must remain inside the configured documentation root.

A file target:

- must already exist;
- must have the `.md` extension, case-insensitively;
- is processed as one selected document; and
- may remain unchanged when no configured codemap section exists.

A directory target:

- is walked recursively;
- includes regular `.md` files only;
- skips symbolic-link file entries;
- applies the repository `.docignore` policy;
- inherits permanent traversal exclusions from the ignore subsystem; and
- explicitly skips `.worktrees/` and `.workingtrees/` directories.

Selected paths are sorted before planning so identical repository input produces stable processing and output order.

An explicit root outside the configured docs root, a nonexistent root, or a non-Markdown file root fails before codemap planning.

## Heading recognition

The configured heading list identifies codemap sections. Matching is case-insensitive after Markdown heading cleanup. Trailing closing `#` characters and surrounding heading whitespace do not create a distinct heading.

Heading recognition is Markdown-aware:

- ATX headings outside fenced code blocks are structural candidates;
- heading-like lines inside fenced code blocks are ignored;
- the matching section ends at the next heading of the same or higher level; and
- exactly one matching section is allowed in a document.

Multiple configured codemap sections are an error. Demon Docs does not guess which one should become canonical.

An existing matching section is processed regardless of whether a future file-type schema would require a codemap for that file.

## Schema-driven missing-section seam

`codemap.SectionSchema` is the internal interface for a future repository file-type policy provider:

```go
type SectionSchema interface {
    CodemapSection(documentPath, source string) (SectionPlacement, bool, error)
}
```

A placement supplies:

```text
configured heading text
Markdown heading level from 1 through 6
byte insertion offset inside the existing document
```

The schema is consulted only when no configured codemap section exists. An existing section bypasses schema lookup and is processed normally.

When a schema requires a section, insertion validates the heading, level, and source offset, inserts the heading with bounded blank-line separation, adds that heading to the active recognition set, and verifies that the new section can be located. The codemap layer never creates the document itself.

The public application currently passes no schema provider, so this path is not yet reachable through the CLI.

## Planning lifecycle

`codemaprun.Build` constructs the complete plan before any file is written.

For each invocation it:

1. builds a codemap extraction dataset for the configured docs root;
2. builds the normalized repository corpus once;
3. loads persisted review policy once;
4. groups extracted entries by document;
5. sorts selected file paths;
6. reads each source through `textio`, retaining source encoding and line-ending behavior;
7. resolves the document path relative to the repository root;
8. computes recommendations with all current targets visible;
9. applies decline policy to each proposed addition;
10. evaluates configured pruning policies for current resolved entries;
11. reconciles the complete codemap section;
12. compares exact encoded before and after bytes; and
13. creates a prepared `filetxn.Rewrite` only when bytes differ.

A document plan records section status, existing targets, recommendations, suppressed additions, selected additions, selected removals, and exact before/after content.

## Recommendation generation

The execution path uses the production `internal/codemaprecommend` package. Benchmarks import the same package instead of carrying an independent production algorithm.

For one document:

1. current known targets are obtained from the corpus;
2. the corpus builds an evidence input with those targets visible;
3. authored codemap sections are stripped from document text before mention evidence is collected;
4. deterministic evidence candidates are collected;
5. the production ranker applies admission, scoring, ordering, bounding, negative-evidence filters, and tier assignment; and
6. current targets are excluded from missing-link output.

Stripping the codemap section prevents a link from becoming evidence for itself merely because it is already listed in the map.

## Shared decline-policy integration

Every proposed addition is converted into the shared review suggestion model with:

```text
document path
target path
score
tier
evidence set and fingerprint
```

The current review policy is replayed before reconciliation.

A candidate is suppressed when the replayed suggestion or its only candidate is declined. The unchanged evidence fingerprint keeps the decline effective. Materially changed evidence can produce a new fingerprint and make the relationship eligible for reconsideration under the normal review lifecycle.

Declines apply only to proposed additions. A decline does not delete a link already present in the codemap.

The public decline and reconsideration controls remain under `ddocs suggestions`, so codemap execution does not create a second policy store.

## Existing-link removal policy

Two independent settings control confidence-based pruning:

```toml
[codemap]
remove_undiscovered_links = false
remove_low_score_links = false
```

Both default to `false`.

Removal evaluation considers only currently extracted entries whose targets resolved to a repository path. For each such entry, the planner temporarily hides that target and reruns recommendation generation:

- when the hidden target is not recovered, `remove_undiscovered_links = true` selects it for removal;
- when the hidden target is recovered only as a `context` relationship, `remove_low_score_links = true` selects it for removal; and
- a recovered stronger relationship remains.

This is an algorithm-confidence policy, not a semantic proof. Enabling either setting intentionally delegates more codemap membership authority to the current evidence model.

Missing, ambiguous, unsupported, or otherwise unresolved entries are not selected through this confidence-pruning loop. Normal link diagnostics and repair own those path states.

## Unified section adoption

The managed marker pair is derived from `[markers].prefix`:

```markdown
<!-- doc-ledger:codemap:start -->
...
<!-- doc-ledger:codemap:end -->
```

The exact prefix is configurable.

On the first successful reconciliation of an existing section, Demon Docs:

1. locates the complete section body;
2. validates any existing codemap marker pair;
3. removes marker lines from the body while retaining the body content;
4. applies explicitly selected line removals;
5. extracts the remaining targets;
6. appends only targets not already present;
7. wraps the entire resulting body in one codemap marker pair; and
8. leaves the section heading outside the marker pair.

A prior partial layout where human-authored links sat outside a smaller generated block is unified into one managed body. Demon Docs does not retain provenance classes inside the section.

Existing prose inside the codemap section is also retained inside the managed region. This is intentional complete-section ownership, not ownership of link lines only.

Malformed or duplicated codemap markers fail the operation. Demon Docs does not repair an ambiguous ownership boundary by guessing.

## Rendering and source preservation

The renderer preserves the dominant supported map form.

For a section containing a complete fenced block:

- new targets are inserted before the first matching closing fence;
- targets remain plain repository-relative paths; and
- no redundant bullet list is created.

For a bullet-style section:

- the first extracted bullet prefix is reused, including indentation and marker form;
- new entries use backtick-wrapped repository-relative targets; and
- when no bullet prefix exists, the fallback is `- `.

Additions are normalized, deduplicated, and sorted before rendering. Existing section content is not globally re-sorted.

Selected removal deletes the extracted source line containing that entry. It does not rewrite unrelated lines.

`textio.Document.Encode` preserves the source document's line-ending and final-newline conventions when the transformed text is encoded for comparison and publication.

## Command semantics

### `inspect`

`inspect` builds the full plan and writes no repository files or review state.

For each selected document it reports:

```text
section status: missing, existing, or schema-created
whether exact bytes would change
proposed additions
persisted-decline decisions
score and tier
evidence lines
configured removals
```

A missing section can still have computed recommendations, but without a connected schema provider it remains unchanged.

### `check`

`check` builds the same plan and writes nothing.

It returns:

```text
0 when no selected document would change
1 when one or more selected documents would change
2 for command-line usage failures
non-zero for configuration, scope, extraction, planning, or read failures
```

Changed document paths are printed in sorted plan order.

### `fix --dry-run`

Dry-run builds the same plan, reports the number of files that would change and per-document add/remove/adopt/create counts, and performs no file or review-state writes.

### `fix`

`fix` applies only prepared changed rewrites. A clean plan reports zero updated files and performs no authored-file write.

## Transaction and concurrent-edit safety

`codemaprun.Apply` delegates the entire changed-file set to `filetxn.Apply`.

The transaction layer:

- accepts only prepared rewrites created by `filetxn.New`;
- rejects duplicate paths;
- verifies internal old and new digests;
- preflights every source before the first write;
- requires each source to remain a regular file;
- compares the current source digest with the planned old digest;
- preserves the existing file mode;
- performs platform-specific atomic replacement;
- rereads and verifies the expected new digest; and
- attempts guarded rollback if a later file fails.

A file changed after planning causes failure rather than overwrite. Rollback is also hash-guarded so it does not erase content created after Demon Docs' attempted write.

Codemap execution currently publishes no separate codemap state after the file transaction. Persisted decline policy is read, not mutated, by `fix`.

## Daemon and watcher exclusion

Production codemap execution has no call path from:

```text
ddocs fix
ddocs check
ddocs watch
foreground watch reconciliation
repository-demon startup
demon event handling
post-write scheduling
```

The watcher may observe files changed by an explicit codemap command in the same way it observes any external repository edit, but it does not initiate or repeat codemap generation.

This is a structural command boundary, not a configuration toggle.

## State and data ownership

Rebuildable command data:

```text
extraction dataset
repository corpus
evidence candidates
ranked recommendations
document plans
before/after rewrites
inspection output
```

Durable shared state:

```text
decline and reconsideration events under refs/ddocs/review
normal link identity and path state under refs/ddocs/state
```

The codemap command does not create a separate provenance ledger for links. The managed source section is the canonical codemap artifact.

## Invariants and safety boundaries

- One recognized section is one unified managed artifact.
- Existing and newly added links are not separated by provenance.
- Existing valid links are retained by default.
- Confidence-based pruning is explicit and disabled by default.
- Declines suppress additions only.
- An existing section bypasses file-type schema requirements.
- A missing section is never invented without schema placement.
- The codemap layer never creates a new document.
- Fenced heading examples are not treated as sections.
- Multiple matching sections and malformed ownership markers are errors.
- Identical inputs produce identical plans and bytes.
- A second successful `fix` is a no-op.
- Check, inspect, and dry-run do not write authored files.
- A concurrent source edit is never overwritten silently.
- The daemon and watcher never execute codemap generation.

## Failure behavior and recovery

### Root outside the docs tree

The command fails before traversal. Select a file or directory beneath the configured docs root or correct configuration selection.

### File root is not Markdown

The command fails. Directory traversal may contain non-Markdown files, but only regular `.md` files become codemap targets.

### No configured section exists

The current public command leaves the document unchanged. Use an existing configured heading until the repository file-type schema provider is connected.

### Multiple matching sections exist

The command fails for that plan. Consolidate the document to one canonical configured codemap section.

### Managed markers are malformed or duplicated

The command fails rather than choosing an ownership range. Repair the marker pair so exactly one matching start and end marker remain inside the section.

### A recommendation is unwanted

Decline it through the shared `ddocs suggestions` workflow. A stable evidence fingerprint suppresses the unchanged relationship on later codemap runs.

### A retained link appears weak

Leave the default pruning settings disabled unless the repository explicitly accepts algorithm-controlled removal. Manual deletion alone does not create a persisted decline against future re-addition.

### Source changed during apply

The hash preflight fails. Review the intervening edit, rerun `inspect` or `check`, and build a new plan.

### One file fails during a multi-file apply

The transaction layer attempts guarded rollback of prior writes. Inspect the reported path and current repository diff before rerunning.

## Extension seams

The intended extension points are:

- `codemap.SectionSchema` for file-type-specific missing-section placement;
- `codemap.Format.SectionHeadings` for heading recognition;
- `codemaprecommend` for production evidence admission and ranking changes;
- `review.Policy` for shared persisted decision semantics;
- `codemap.ManagedUpdate` for explicitly selected additions and removals; and
- `filetxn.Rewrite` for common content-addressed publication.

A new renderer or schema provider must preserve complete-section ownership, deterministic output, idempotency, source-hash protection, and the default no-pruning policy.

## Code map

- `internal/app/codemap_execute.go` — flag parsing, config loading, plan dispatch, command exit behavior.
- `internal/app/codemap_execute_scope.go` — root containment, file validation, recursive traversal, ignore and worktree exclusions.
- `internal/app/codemap_execute_output.go` — human-readable summaries and inspection evidence.
- `internal/codemaprun/build.go` — dataset/corpus construction, review-policy replay, recommendation planning, pruning evaluation, rewrite construction.
- `internal/codemaprun/model.go` — invocation options and per-document plan model.
- `internal/codemaprun/apply.go` — shared transaction publication boundary.
- `internal/codemap/managed.go` — schema gating, unified adoption, addition/removal reconciliation.
- `internal/codemap/managed_section.go` — Markdown-aware section location and schema insertion validation.
- `internal/codemap/managed_render.go` — marker validation, fence/bullet rendering, line removal, normalization.
- `internal/codemaprecommend/` — production ranking and evidence filters.
- `internal/review/` — shared decline and reconsideration replay.
- `internal/filetxn/` — batch preflight, atomic replacement, verification, and guarded rollback.
- `internal/textio/` — source encoding and newline preservation.

## Tests

Focused contracts include:

```text
internal/app/codemap_execute_test.go
  help and singular compatibility alias
  required roots
  single-file check/dry-run/fix convergence

internal/codemap/managed_test.go
  complete-section adoption
  idempotency
  missing-section skip without schema
  schema-required insertion
  fenced Space Rocks-style preservation
  legacy partial-region unification
  selected-entry removal

internal/codemaprun/build_test.go
  production recommendation planning
  persisted decline suppression
  opt-in pruning behavior

internal/codemaprecommend/suggestions_test.go
  production-ranker parity after benchmark extraction

internal/filetxn/apply_test.go
  batch preflight
  digest verification
  atomic replacement and rollback guards
```

Run the focused gate:

```bash
go test ./internal/codemap ./internal/codemaprecommend ./internal/codemaprun ./internal/app ./internal/filetxn -count=1
```

Run the complete release safety gates after changing ownership, rendering, ranking, or publication:

```bash
go test ./... -count=1
go vet ./...
```

## Related docs

- [Managing Codemaps](../guides/managing-codemaps.md)
- [Codemap Pipeline](codemap-pipeline.md)
- [Codemap Extraction and Dataset](codemap-extraction-and-dataset.md)
- [Codemap Evidence and Ranking](codemap-evidence-and-ranking.md)
- [Application Orchestration](application-orchestration.md)
- [Review Lifecycles](review-lifecycles.md)
- [Generated Rewrite Publication](generated-rewrite-publication.md)
- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Diagnostics and Exit Behavior](../reference/diagnostics-and-exit-behavior.md)
- [Current Product Limitations](../limits/current-limitations.md)

## Notes

The managed codemap source section is the durable product artifact. Scores, evidence, and command plans remain rebuildable explanations of how the current implementation arrived at one update.
