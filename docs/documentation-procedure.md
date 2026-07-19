# Documentation Procedure

Parent index: [Demon Docs Documentation](./README.md)

## Purpose

This document defines the standard process for creating, updating, moving, graduating, and removing Demon Docs documentation.

## Overview

Use this procedure for all repository documentation work, including user guides, reference material, implemented architecture, operations, research, planning, development guidance, indexes, stubs, limits, and legacy cleanup.

Follow these steps in order:

```text
1. Classify the information.
2. Choose the owning document and folder.
3. Decide whether a new file or folder is justified.
4. Create or update every affected README.md index.
5. Apply the stub rule when the document is incomplete.
6. Write the required document shape.
7. Add related docs, code maps, recovery information, and notes.
8. Graduate implemented facts out of planning and research.
9. Remove stale, duplicated, or legacy material.
10. Update implementation-to-document coverage.
11. Verify structure, links, and product accuracy.
```

## 1. Classify the information

Classify information by what it is, not where it was first discussed.

```text
Guide        = task-oriented user workflow
Reference    = exact public contract or lookup material
Architecture = implemented ownership and internal behavior
Operations   = running behavior, automation, recovery, and troubleshooting
Research     = experiment, benchmark, corpus, evidence, or evaluation
Planning     = future, unresolved, proposed, or not-yet-current work
Development  = contributor workflow, tests, fixtures, release, repository layout
Limits       = current defect, blocker, or incomplete transitional behavior
Agent        = agent-specific editing and tool rules
Notes        = temporary or unclassified scratch material
Legacy       = temporary migration source only
```

When a change spans several types, update each owning document rather than forcing everything into one page.

## 2. Choose the owner

Use this mapping:

```text
Guide        -> docs/guides/
Reference    -> docs/reference/
Architecture -> docs/architecture/
Operations   -> docs/operations/
Research     -> docs/research/
Planning     -> docs/planning/
Development  -> docs/development/
Limits       -> docs/limits/ when needed
Agent        -> docs/agent/ when needed
Notes        -> docs/notes.md when needed
Legacy       -> docs/legacy/ temporarily
```

Place information where it is owned, not merely where it is consumed.

Examples:

```text
"Run ddocs check in CI"                     -> guide
"check exits non-zero for pending changes"  -> reference
"app.Run coordinates selected subsystems"   -> architecture
"recover a stale demon owner"               -> operations
"150-suggestion precision sample results"   -> research
"future polyglot provider contract"          -> planning
"make release-check composition"             -> development
```

## 3. Reuse before creating

Before creating a file, check whether an existing document owns the fact.

Create a new file only when:

- the topic has a clear durable owner;
- it has enough substance to stand alone;
- adding it to an existing document would blur ownership; and
- readers are likely to look for it independently.

Create a new folder only when it is a durable boundary expected to contain multiple documents.

Do not create vague buckets. Use `notes.md` for temporary material and `stubs/` for incomplete documents with a known owner.

## 4. Update indexes first

Every normal documentation folder must contain `README.md`.

When adding a file:

```text
add it to the parent README.md Direct Files section
use a one-line description that states its ownership
```

When adding a folder:

```text
create the child README.md
add the folder to the parent Direct Folders section
link to child-folder/README.md
```

When moving or deleting a document, update all affected indexes in the same change.

Indexes remain navigational. Do not copy the document body into the index.

## 5. Apply the stub rule

Use a nearby `stubs/` folder when a document has a clear eventual owner but is not complete enough to be canonical.

Stub index descriptions begin with `Stub:`. Empty `stubs/` folders may remain without their own index. Other empty documentation folders are not retained.

When a stub graduates, move it to the owning folder, apply the full document shape, update indexes and links, and delete the old stub path.

## 6. Write the required shape

Every normal document begins with:

```markdown
# Title

Parent index: [Owning Folder](./README.md)

## Purpose

## Overview
```

Use the type-specific shape from [Documentation Policy](documentation-policy.md).

The body should answer ownership and behavior questions in prose. Tables, command blocks, file lists, and code maps support the explanation; they do not replace it.

## 7. Add supporting sections

### Related docs

Add one `## Related docs` section linking to the most relevant canonical documents.

Prefer links that bridge document types:

```text
guide -> reference and operations
reference -> guide and architecture
architecture -> reference, operations, tests, and plans
research -> architecture and planning
planning -> implemented dependencies and research
```

### Code maps

Add a code map to implementation-facing architecture and development documents. Include primary code, tests, generated/source files, and important non-ownership boundaries.

### Failure and recovery

Guides and operations docs must cover likely failure states and the safe recovery path. Reference docs must describe diagnostics and mutation scope.

### Notes

End every normal document with `## Notes`. Keep this section small.

## 8. Graduate current facts

When planned work ships:

```text
move current behavior into guide/reference/architecture/operations docs
update the plan's Current status
link to implemented references
remove future-tense duplicates that are no longer true
leave unresolved and later work in planning
```

When research produces a decision:

```text
retain the evidence in research
update the owning current or planning document with the decision
state the benchmark's population and limitations
avoid turning sample results into universal product claims
```

## 9. Remove stale material

Check for:

- root README sections duplicated by detailed docs;
- planning pages that still describe shipped behavior as future;
- research pages treated as product reference;
- old paths left after document moves;
- duplicate current facts across architecture and operations;
- stale index entries;
- fully replaced legacy docs;
- graduated stubs still present under `stubs/`; and
- empty non-stub folders.

Delete fully replaced material rather than maintaining indefinite redirects inside the repository. Git history remains the recovery source.

## 10. Update implementation coverage

When production code or public commands change, update [Documentation Coverage Map](development/documentation-coverage.md).

Confirm that:

```text
every production package has a canonical current owner
every public command family has a task or reference owner
small utility packages map to the subsystem they concretely serve
implemented behavior is not owned only by research or planning
new limits are recorded when the implementation remains incomplete
```

A coverage-table link does not replace substantive documentation. Open the linked owner and verify that it actually describes the changed responsibility.

## 11. Verify

Before completion, verify:

```text
The document type is correct.
The owner and folder are correct.
Every folder has README.md unless it is stubs/.
Every file and direct folder is indexed.
Parent-index links are correct.
The document has Purpose, Overview, Related docs, and Notes.
Type-specific sections are present.
Current behavior is not owned only by planning or research.
Implementation docs explain ownership and include useful code maps.
Guides state prerequisites, expected result, and recovery.
Moved links resolve.
The root README remains a product entry point rather than a complete manual.
```

Run the project documentation checks and normal test gate after structural changes. Document-only moves can still break repository links, examples, fixture assumptions, or codemap extraction.

## Related docs

- [Documentation Policy](documentation-policy.md)
- [Demon Docs Documentation](README.md)
- [Testing and Fixtures](development/testing-and-fixtures.md)
- [Repository Layout](development/repository-layout.md)

## Notes

Demon Docs uses its own default `README.md` index convention. Reconciliation-generated index changes should be reviewed like any other documentation edit because descriptions and authored guidance remain human-owned.
