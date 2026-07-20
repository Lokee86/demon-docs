---
author: brian
created: "2026-07-19"
document_id: 019f7d55-2e95-72b2-a536-7659111e2ff5
document_type: general
policy_exempt: false
summary: This document describes the implemented transformation boundary that preserves authored Markdown while Demon Docs creates and updates managed folder indexes, parent-index lines, descriptions, and generated entries.
---
# Managed Markdown Transformation

Parent index: [Architecture](./README.md)

## Purpose

This document describes the implemented transformation boundary that preserves authored Markdown while Demon Docs creates and updates managed folder indexes, parent-index lines, descriptions, and generated entries.

## Overview

Managed Markdown is not rewritten as a normalized document. Demon Docs builds a structural view of the source, identifies narrowly owned spans, plans replacement text for those spans, and preserves everything outside them.

The complete forward-index path crosses three distinct ownership boundaries:

```text
internal/markdown
  structural recognition and in-memory source transformation

internal/reconcile
  filesystem-to-index matching, description preservation, and update planning

internal/textio
  newline-aware decoding and byte-preserving re-encoding
```

Those boundaries are intentionally separate. Parsing a managed block does not decide which files belong in it. Matching a moved file does not choose how source bytes are encoded. Applying a plan does not reinterpret Markdown.

## Code root

```text
internal/markdown/
internal/reconcile/
internal/textio/
internal/scan/
internal/model/
internal/pathutil/
```

## Responsibilities

This boundary owns:

- recognition of real Markdown headings and fenced code ranges;
- detection, migration, creation, and replacement of managed sections;
- parsing and rendering of generated file, stub, and folder entries;
- insertion, replacement, and removal of parent-index lines;
- creation of missing folder-index templates;
- preservation of existing entry descriptions when identity is sufficiently clear;
- direct-to-stub, stub-to-direct, and unambiguous cross-folder transition handling;
- deterministic planning and stale-entry diagnostics;
- preservation of LF, CRLF, mixed line endings, and final-newline state; and
- repository-containment checks before forward-index application.

## Does not own

This boundary does not own:

- arbitrary Markdown formatting or prose normalization;
- semantic document placement;
- repository-wide local-link identity and repair;
- reverse-index projection into code folders;
- codemap suggestion evidence;
- cross-file atomic publication; or
- rollback of earlier forward-index writes when a later write fails.

## Structural recognition

`internal/markdown` uses Goldmark to identify headings and fenced code blocks. It then performs source-span operations against the original string rather than serializing Goldmark's AST.

This distinction is critical:

- Goldmark determines whether a heading or marker-like line is real Markdown structure.
- Original byte offsets determine what Demon Docs may replace.
- Text inside fenced code remains authored example content even when it resembles a heading, parent link, or managed marker.
- Inline code is not treated as a managed section boundary.

The scanner recognizes three managed section identities:

```text
files
stubs
folders
```

Their displayed headings and marker prefix are configurable. Section identity remains internal and stable even when headings are renamed.

## Managed-section lifecycle

### Existing canonical section

When both canonical start and end markers exist outside fenced code, `ReplaceManaged` replaces only that marker span.

The replacement consists of:

```text
start marker
optional blank line and generated entries
end marker
```

The source prefix before the start marker and suffix after the end marker are reused unchanged.

### Legacy heading without markers

When a configured current or legacy section heading exists but no managed markers exist, `EnsureManaged` wraps the legacy section body with canonical markers. The heading is normalized to the current configured heading.

The legacy body extends to the next parsed heading. Fence-contained heading text does not terminate it.

### Mixed legacy and canonical source

When managed markers already exist but a recognized legacy heading remains, Demon Docs normalizes the heading text without rebuilding the rest of the document.

### Missing section

Missing sections are created in the fixed internal order:

```text
files
stubs
folders
```

New managed sections are inserted before the first real `Related Docs` or `Notes` heading when present. Otherwise, they are appended. Heading-like examples inside fences are ignored.

### Malformed section

When a start marker exists without a matching end marker, replacement stops at the next real heading, the next managed section start, or end of file. This bounds repair and prevents one malformed block from consuming later authored sections.

## Entry parsing and rendering

Managed entries use the exact generated form:

```markdown
- [link text](target) - description
```

`ParseEntries` records:

- owning index path;
- section identity;
- link text;
- link target;
- description; and
- original source line.

Only entries inside a recognized managed section are parsed. A matching bullet outside a managed section remains ordinary prose.

Rendering is deliberately simple. The reconciliation layer decides the canonical relative target and the description to preserve; the Markdown layer only renders the resulting entry.

## Parent-index lifecycle

`DesiredParent` determines whether a file should contain a parent-index line:

- the managed root index has no parent;
- child folder indexes point to the parent folder index when `folder_indexes` is enabled;
- normal indexed files point to their folder index when `indexed_files` is enabled;
- draft files point to the owning folder index one level above the draft folder; and
- non-configured editable extensions are never changed.

`UpdateParent` then performs one of three operations:

```text
existing line + desired line -> replace exact line
existing line + no desired line -> remove line and adjacent managed spacing
no existing line + desired line -> insert after first real heading
```

Parent-shaped lines inside fenced code are ignored. Insertion preserves whether the source originally ended with a newline.

## Forward-index planning

`internal/reconcile.TreeWithIgnoreRoot` performs the forward-index plan in these stages:

```text
scan the managed documentation tree
-> read existing indexes with newline metadata
-> parse existing generated entries
-> derive stable folder and root display titles
-> create in-memory templates for missing indexes
-> calculate direct, stub, and child-folder membership
-> match stable and transitioned entries
-> render each managed section
-> plan parent-index changes
-> sort stale-entry diagnostics
-> return file updates without writing
```

The plan contains complete replacement text for each changed file. `check` consumes the plan diagnostically; `fix` applies it.

## Description preservation and transitions

Description reuse is conservative.

### Stable entry

An entry that still resolves to the same target in the same section keeps its link text and description while its relative target is canonicalized.

### Direct-to-stub transition

A direct file moved into the configured draft folder keeps its description and gains the configured draft prefix when absent.

### Stub graduation

A draft file moved into the owning folder keeps its description, removes the configured draft prefix, and capitalizes the first rune when the stripped description began lowercase.

### Cross-folder file move

A stale file entry may supply its description to a same-basename destination only when both are globally unique among unmatched entries and unmatched filesystem files.

### Cross-folder folder move

The same uniqueness rule applies to child folders, using the child folder basename and its index target.

### Ambiguous transition

When more than one stale entry or destination shares the basename, Demon Docs does not guess. The destination receives the configured generated description and the stale entries are removed and reported.

## Newline and byte preservation

`internal/textio` normalizes CRLF to LF for in-memory transformations while retaining the original representation.

For uniform files:

- LF remains LF;
- CRLF remains CRLF; and
- final-newline state follows the transformed text.

For mixed-ending files, unchanged lines are matched between the old and new normalized text. Their original raw bytes, including original line endings and trailing spaces, are reused. Newly inserted or changed lines use the first observed line-ending style.

New files use the host platform's normal line ending through `EncodeNew`.

This is a preservation guarantee for unchanged source lines, not a general guarantee that malformed managed content retains its original bytes inside the owned span.

## Application boundary

`ApplyWithin` validates every planned path against the configured docs root before the first write. If any path escapes the root, no update is applied.

After containment preflight, forward-index updates are written sequentially:

```text
create parent directory if required
-> reread existing file when replacing it
-> encode with its original newline representation
-> write the complete file
-> continue to the next update
```

Forward-index application is not a multi-file transaction. If a later write fails, earlier writes remain. Re-running reconciliation is the recovery mechanism because planning is deterministic and idempotent.

This differs from generated link rewrites, which use hash guards, temporary files, and a separate rollback boundary.

## State and data ownership

- `internal/scan` owns the current filesystem inventory below the managed root.
- `internal/markdown` owns structural source recognition and in-memory transformations.
- `internal/reconcile` owns matching filesystem facts to existing generated entries and producing `model.FileUpdate` plans.
- `internal/textio` owns newline-aware decoding and encoding.
- `internal/pathutil` owns relative path rendering used in generated targets.
- No persistent private `.ddocs` state is required for forward-index matching.

## Invariants and safety boundaries

- Text outside the selected managed block or parent-index line remains authored content.
- Fence-contained examples are never treated as real managed structure.
- A managed replacement cannot intentionally cross into the next real section.
- Stable descriptions are preserved.
- Ambiguous basename-based transitions are not guessed.
- Planning is deterministic for identical filesystem and source inputs.
- A converged second plan produces no updates.
- `check` does not apply the plan.
- All forward-index paths are containment-checked before mutation.
- Mixed line endings and trailing spaces on unchanged lines are preserved.
- Application may partially complete across files; it does not claim cross-file atomicity.

## Failure behavior

Planning fails when required files cannot be read, the documentation tree cannot be scanned, or a managed replacement cannot be formed. No planned update is applied by planning itself.

Application fails on directory creation, reread, encoding input, or file-write errors. The returned changed count identifies how many earlier files completed. The safe recovery procedure is to correct the filesystem or permission problem and run `ddocs fix --docs` again.

Malformed blocks are repaired only within the bounded section range. Unsupported authored bullet shapes are not imported as generated entries and may be replaced when they fall inside an owned managed block.

## Code map

- `internal/markdown/markdown.go` — structural recognition, managed blocks, entries, templates, titles, descriptions, and parent lines.
- `internal/textio/textio.go` — uniform and mixed line-ending preservation.
- `internal/reconcile/reconcile.go` — tree planning, transition matching, deterministic diagnostics, containment, and application.
- `internal/scan/` — managed documentation-tree inventory.
- `internal/model/model.go` — folder, entry, update, and reconciliation result structures.
- `internal/pathutil/` — generated relative path rendering.

## Tests

Focused contracts include:

- `internal/markdown/source_preservation_test.go` — byte preservation outside each managed block and custom marker examples inside fences;
- `internal/markdown/markdown_test.go` — Goldmark structural boundaries, parent links, templates, and final-newline state;
- `internal/markdown/behavior_test.go` — migration, malformed boundaries, custom headings, and parent insertion/replacement/removal;
- `internal/reconcile/transitions_moves_test.go` — direct/stub and cross-folder description transitions;
- `internal/reconcile/preservation_repair_test.go` — stale removal and bounded malformed-block repair;
- `internal/reconcile/line_endings_test.go` — mixed-ending unmanaged-byte preservation;
- `internal/reconcile/determinism_test.go` — stable plans and clean second passes;
- `internal/reconcile/scope_test.go` — containment preflight; and
- `internal/textio/textio_test.go` — newline decoding and re-encoding.

Run the focused suite with:

```bash
go test ./internal/markdown ./internal/reconcile ./internal/textio ./internal/scan -count=1
```

## Related docs

- [Reconciliation Model](reconciliation-pipeline.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Configuration Reference](../reference/configuration.md)
- [Behavioral Contract Matrix](../development/behavioral-contract-matrix.md)
- [Extending Reconciliation and State](../development/extending-reconciliation-and-state.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)

## Notes

The term “source-preserving” means Demon Docs limits its ownership and preserves unchanged source bytes where implemented. It does not mean generated spans retain arbitrary formatting that conflicts with the canonical managed format.
