---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7310-8006-496ce3421c27
document_type: general
policy_exempt: false
summary: This document defines the documentation policy for Demon Docs.
---
# Documentation Policy

Parent index: [Demon Docs Documentation](./INDEX.md)

## Purpose

This document defines the documentation policy for Demon Docs.

It governs where documentation belongs, how current facts are separated from plans and research, how folder indexes are maintained, what shape each document type requires, when code maps are expected, and how stale or duplicated documentation is removed.

## Overview

Demon Docs documentation is organized by **documentation type and ownership**, not merely by feature name.

A single feature may have several valid documents because different documents answer different questions:

```text
Guide
= How does a user complete a task?

Reference
= What is the exact command, option, format, state, or diagnostic contract?

Architecture
= Which implemented boundary owns the behavior and how does it work?

Operations
= How is the running tool observed, recovered, and safely operated?

Research
= What evidence, experiment, corpus, or benchmark supports a claim?

Planning
= What is proposed, unresolved, sequenced, or not yet current?

Development
= How is the project built, tested, released, and changed safely?
```

Documents must make those distinctions explicit. Planning and research material must not be presented as shipped product behavior.

## Documentation types

Demon Docs uses these documentation types:

```text
Guide
Reference
Architecture
Operations
Research
Planning
Development
Limits
Agent
Notes
Legacy
```

Only durable categories with current content need physical folders. `limits/`, `agent/`, and `legacy/` should be created when the repository has material that genuinely belongs there, not pre-created as empty taxonomy placeholders.

## Guide policy

Guides are task-oriented user documentation.

Guides explain how to accomplish a concrete goal, such as:

```text
installing Demon Docs
initializing an existing repository
adding checks to CI
running foreground automation
using the repository demon
recovering from an ambiguous move
reviewing codemap evidence
operating across linked worktrees
```

Guides should:

- begin with the expected outcome;
- state prerequisites;
- present steps in execution order;
- distinguish safe read-only commands from mutating commands;
- link to reference documentation for exhaustive option details;
- link to architecture only when the ownership model matters to the task; and
- include recovery guidance for likely failures.

Guides must not become exhaustive command references or internal implementation documents.

## Reference policy

Reference documentation defines exact public contracts.

Reference docs may cover:

```text
CLI commands and flags
configuration keys and precedence
managed index markers
supported link syntax
private repository state layout
exit behavior and diagnostics
file selection and ignore rules
machine-readable output contracts
```

Reference documentation should be precise, complete, and organized for lookup. It may repeat small command signatures where needed, but explanatory workflows belong in guides.

When a reference document covers a public command or file format, it must explain:

- what the surface is for;
- who or what consumes it;
- whether it reads or mutates repository state;
- what scope it can affect;
- what failure states are reported; and
- what it explicitly does not own.

## Architecture policy

Architecture documentation describes **implemented** ownership and internal behavior.

Architecture docs should explain:

```text
code root
responsibilities
does-not-own boundaries
runtime or reconciliation flow
state ownership
invariants and safety rules
public/internal seams
failure behavior
code map
tests
```

Architecture documents must not claim future seams or planned systems are implemented. Planned architecture belongs under `planning/` until the owning boundary exists in code.

Architecture documents should identify important non-ownership boundaries. A code map alone is not sufficient; prose must explain how the implementation behaves and why the boundary exists.

## Operations policy

Operations documentation covers running behavior and recovery.

Operations docs may cover:

```text
foreground watch behavior
repository demon ownership
feeder leases and heartbeats
runtime files and logs
shutdown and stale-owner recovery
linked-worktree behavior
CI and unattended execution
performance characteristics
troubleshooting and repair workflows
```

Operations docs should distinguish correctness surfaces from convenience automation. `check` and deterministic reconciliation remain authoritative even when watch or daemon automation is enabled.

Operations documentation must state which files or processes are disposable and rebuildable, and which authored repository files may be changed.

## Research policy

Research documentation records evidence, methodology, corpora, benchmark conditions, and interpretations.

Research docs must distinguish:

```text
question
method
inputs or corpus
measured results
limitations
interpretation
retained artifacts
```

Research results are not automatically product guarantees. A benchmark from one repository or curated sample must not be generalized beyond its stated population.

Research documents may link to implemented architecture or planning, but they do not own product contracts. When research changes a product decision, the corresponding reference, architecture, or planning document must also be updated.

## Planning policy

Planning docs describe future, unresolved, proposed, partially implemented, or back-burnered work.

Planning documents must clearly state current status, such as:

```text
stub
active planning
ready for implementation
partially implemented
back-burnered
superseded
```

Planning docs should distinguish:

```text
decided direction
open decisions
expected ownership
implementation sequence
acceptance criteria
current implemented dependencies
```

Planning documents must not remain the only home for implemented facts. When work ships, current behavior must move or be rewritten into the correct guide, reference, architecture, or operations document. The planning document may retain a short status summary and links to implemented references.

## Development policy

Development documentation covers contributor-facing workflows and repository maintenance.

Development docs may cover:

```text
repository layout
build and test commands
fixture strategy
release gates
benchmark execution
generated artifacts
platform-specific development behavior
contribution boundaries
```

Development documentation should explain both the command and the reason it exists. Tests and release gates must link to the architecture or product behavior they protect where practical.

Development docs should include code maps when they describe implementation ownership or test organization.

## Limits policy

`docs/limits/` is reserved for current product limitations, known defects, blocked work, and incomplete transitional behavior.

Limits are not the same as intentional safety boundaries or architecture invariants. Permanent boundaries belong in architecture documentation. Future feature work belongs in planning. Experimental uncertainty belongs in research.

A limits document should identify:

- the affected surface;
- current impact;
- workaround or recovery path, when one exists;
- status;
- the owning implementation or plan; and
- the condition under which the entry can be removed.

## Agent policy

`docs/agent/` is reserved for agent-specific editing, testing, tool, and repository-orientation rules.

Agent docs may summarize a boundary only when that summary directly guides safe edits. Long-lived product facts must remain in canonical guide, reference, architecture, operations, or development documents.

Volatile status must not be spread through durable docs. When an agent status page is introduced, it should be explicitly non-authoritative and aggressively pruned.

## Notes policy

A top-level `docs/notes.md` may be used as a non-authoritative scratchpad when information is temporary, unclear, or not ready to classify.

Stable facts must move into the correct document type. Notes must not become a permanent substitute for creating an owning document.

## Legacy policy

`docs/legacy/` is temporary migration source material only.

Legacy documents are not current authority. Once useful facts are migrated, rewritten, or intentionally discarded, the legacy source should be deleted. Current documents must not link to legacy material as authoritative behavior.

## Folder creation policy

Create a folder only when it represents a durable boundary that will contain multiple related documents.

Do not create vague buckets such as:

```text
misc
common
general
stuff
other
```

Use an existing owning document for small additions. Use `notes.md` for temporary or unclassified material. Use a nearby `stubs/` folder when a document has a clear eventual owner but is not complete enough to be canonical.

## Index policy

Every normal documentation folder contains a `README.md` index.

`stubs/` folders are exempt from the index requirement. Empty folders named exactly `stubs/` may remain as reserved draft locations.

Every index must list:

```text
every Markdown file directly in that folder
every direct documentation subfolder
every directly indexed non-Markdown support file when it is intentionally part of the docs surface
```

Subfolder links must point to the subfolder `README.md`.

The top-level `docs/INDEX.md` is both the documentation entry point and the short rulebook. Detailed policy belongs in this document rather than being duplicated in every index.

Indexes should stay navigational. They may summarize ownership and authority, but they should not duplicate full feature documentation.

## Managed index sections

Demon Docs' own indexes should use the default managed sections:

```markdown

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
```

Hand-authored folder guidance belongs outside those markers.

## Stub policy

Incomplete documents with a clear eventual owner belong in the nearest appropriate `stubs/` folder.

A stub is non-canonical. Parent index descriptions for stubs must start with `Stub:`.

When a stub becomes canonical:

```text
move it into the owning folder
update the parent index
remove the old stub path
apply the required document shape
add related docs
add a code map when required
```

## Universal document shape

Every normal documentation file includes, at minimum:

```text
Purpose
Overview
Type-specific sections
Related docs
Notes
```

The title and parent-index link appear before `Purpose`.

`Purpose` explains why the document exists. `Overview` explains what is being documented, how it behaves, and how it fits into Demon Docs.

## Type-specific shapes

### Guide shape

```text
Purpose
Overview
Prerequisites
Procedure or workflow
Expected result
Failure and recovery
Related docs
Notes
```

### Reference shape

```text
Purpose
Overview
Exact contract sections
Defaults and precedence where applicable
Diagnostics or failure behavior
Examples
Related docs
Notes
```

### Architecture shape

```text
Purpose
Overview
Code root
Responsibilities
Does not own
Flow or lifecycle
State/data ownership
Invariants and safety boundaries
Code map
Tests
Related docs
Notes
```

### Operations shape

```text
Purpose
Overview
Operating model
Commands or controls
Runtime state and logs
Failure and recovery
Verification
Related docs
Notes
```

### Research shape

```text
Purpose
Overview
Research status
Question
Method
Corpus or inputs
Results
Limitations
Interpretation
Retained artifacts
Related docs
Notes
```

Existing research documents may combine headings when the same information remains explicit.

### Planning shape

```text
Purpose
Overview
Current status
Expected ownership or ownership boundary
Planned behavior
Implementation sequence
Acceptance criteria
Open decisions
Implemented references when applicable
Related docs
Notes
```

### Development shape

```text
Purpose
Overview
Workflow or repository boundary
Commands
Failure modes
Code map when implementation-facing
Related docs
Notes
```

## Related docs policy

Every normal document includes one `Related docs` section.

Links should point to the most relevant canonical documents. Related-doc sections should connect user workflow, exact reference, implemented architecture, operations, research, and plans without duplicating their content.

## Notes policy

Every normal document ends with `## Notes`.

Notes may contain caveats, naming history, edge cases, or small non-blocking observations. Large backlog items, core design rules, implementation ownership, and future plans belong in their owning document types.

## Code map policy

Implementation-facing architecture and development documents should include code maps.

A useful code map includes:

```text
primary implementation files or folders
related tests
source or generated artifacts
important non-ownership boundaries
```

Guides, planning documents, limits, and agent docs do not require code maps. Reference and operations documents may include them when implementation paths materially clarify the public contract.

## Active issues policy

Completed documents should not routinely contain broad `Known limits` sections.

Use `Active issues` only when a current document must point to a temporary defect, blocker, or transitional gap. The issue should link to the owning limits or planning document rather than being fully duplicated.

## Planning graduation policy

When planned work becomes current:

```text
identify the facts that are now implemented
classify those facts by current documentation type
write or update the owning current documents
leave only future, unresolved, or sequencing material in planning
add implemented references to the plan when useful
remove stale duplicate claims
update all affected indexes and related-doc links
```

There is no separate informal graduation path. Shipped behavior is not considered documented until current documentation owns it.

## Documentation correctness

Documentation changes are incomplete when any of these remain true:

```text
a moved document is absent from its owning index
a folder lacks its README.md index
a planning page is presented as current authority
research metrics are stated as universal guarantees
an implementation document has only a code list and no ownership prose
a guide duplicates the full CLI or config reference
links still point to moved or deleted documentation
current behavior exists only in the root README or roadmap
a production package or public command has no canonical current documentation owner
the documentation coverage map points only to planning or research for implemented code
```

## Implementation coverage policy

`docs/development/documentation-coverage.md` maps every production package, public command family, and independent stateful flow to canonical current documentation.

Update that map when a production package, command group, durable ownership boundary, persistent model, mutation sequence, lifecycle, concurrency boundary, or recovery seam is added, removed, renamed, or materially reassigned. Small utility packages may share the current architecture owner of the subsystem they serve. Major independent boundaries require a dedicated current owner.

Package-level coverage alone is insufficient. A broad package may contain several independently important flows, such as planning, generated source mutation, review publication, state publication, rollback, and watcher suppression. Those flows require focused owners when one umbrella page cannot explain their state transitions and failure boundaries clearly.

A map entry is not sufficient by itself. The linked document must actually explain responsibility, flow, state ownership, non-ownership boundaries, invariants, failure and recovery behavior, extension seams, and relevant tests or public contracts.

Package-level coverage is also not sufficient when one package owns several independent stateful flows. Every durable mutation boundary, persistent model, concurrency boundary, machine-readable contract, and safe-extension seam requires a canonical explanation at an appropriate level of detail. `docs/development/behavioral-contract-matrix.md` maps critical invariants to the tests and release gates that protect them.

## Related docs

- [Documentation Procedure](documentation-procedure.md)
- [Demon Docs Documentation](INDEX.md)
- [Repository Layout](development/repository-layout.md)
- [Roadmap](planning/roadmap.md)

## Notes

This policy is derived from the Space Rocks documentation policy but replaces game-specific domain, service, protocol, data, and devtools categories with categories appropriate to a repository documentation engine and CLI product.
