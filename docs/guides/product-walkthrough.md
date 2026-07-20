---
author: brian
created: "2026-07-19"
document_id: 019f7e4d-9000-7243-ba61-27a3b45d912b
document_type: general
policy_exempt: false
summary: Adopt Demon Docs in a small existing repository, from initialization and indexes through link health, safe moves, reverse indexes, codemap suggestions, and schema-backed document creation.
---
# Product Walkthrough
Parent index: [Guides](./INDEX.md)
## Purpose
This walkthrough shows the main Demon Docs workflow in a small existing repository. It introduces each managed surface separately so the resulting changes remain reviewable.
## Overview
The sequence covers repository initialization, staged adoption, generated indexes, link state, orphan health, link-aware moves, reverse indexes, codemap execution, suggestion review, and schema-backed document creation. The command outputs were reproduced against the current CLI; paths and generated identifiers vary.
### Starting fixture
```text
docs/
  README.md
  architecture/runtime.md
  guides/installing.md
internal/runtime/service.go
```
`docs/INDEX.md` contains ordinary authored navigation:
```markdown
# Acme Service
Start with the [installation guide](guides/installing.md) and read the
[runtime overview](architecture/runtime.md).
```
`docs/architecture/runtime.md` contains an explicit code target:
```markdown
# Runtime
The runtime package owns service startup and shutdown.
## Code map
- `internal/runtime/service.go`
```
Demon Docs treats that target as authored evidence. It does not infer ownership from arbitrary prose.
### 1. Initialize
Run from the repository root:
```bash
ddocs init --root docs
ddocs status
```
Initialization creates `.ddocs/config.toml` and starter schemas. `status` confirms the repository root, docs root, config path, and `.docignore` path. Authored documentation remains normal Markdown; private identities and history live beneath `.ddocs/`.
### 2. Stage adoption for existing docs
A new config enables indexes, links, frontmatter, format policy, and repository-demon eligibility. Existing documentation may not match the starter metadata or body schemas.

For a narrow first pass, temporarily change:
```toml
[format]
enabled = false

[frontmatter]
enabled = false
```
Leave indexes and links enabled. Re-enable policy enforcement after adapting the schemas to the repository. This avoids mixing navigation adoption with a metadata and body-format migration.
### 3. Generate indexes
```bash
ddocs fix --docs
```
With frontmatter and format disabled, the fixture reports:
```text
ddocs fix updated 3 file(s)
```
Demon Docs adds managed index regions to `docs/INDEX.md` and creates indexes for `docs/architecture/` and `docs/guides/`. The guide index contains:
```markdown
## Direct Files
<!-- doc-ledger:files:start -->

- [installing.md](installing.md) - Installing documentation.
<!-- doc-ledger:files:end -->
```
Authored titles and prose remain outside the managed regions.
### 4. Establish link state
```bash
ddocs fix --links
```
The first pass records the existing graph instead of guessing historical moves:
```text
ddocs fix updated 0 file(s)
message: Link state is not initialized; this pass records a baseline and does not repair links.
```
Check health:
```bash
ddocs check --docs --links
```
The fixture initially reports:
```text
ddocs check failed
message: Orphan document: docs/guides/installing.md
```
Generated index links do not count as meaningful authored inbound evidence. A file can therefore appear in navigation and still be an orphan.
### 5. Move a document safely
Preview a rename:
```bash
ddocs mv --dry-run \
  docs/guides/installing.md \
  docs/guides/installation.md
```
The verified plan reports one move, two changed Markdown files, and two rewritten links. Apply it with the same command without `--dry-run`:
```bash
ddocs mv \
  docs/guides/installing.md \
  docs/guides/installation.md
```
The root navigation and guide index now point to `installation.md` without unrelated prose changes.

Resolve the orphan by adding a meaningful relationship to `docs/architecture/runtime.md`:
```markdown
For setup, see the [installation guide](../guides/installation.md).
```
Then verify:
```bash
ddocs fix --links
ddocs check --docs --links
```
The check passes.
### 6. Project docs back onto code
The runtime document already targets `internal/runtime/service.go` under a configured codemap heading. Build the reverse projection:
```bash
ddocs fix --reverse --reverse-root internal
ddocs check --reverse --reverse-root internal
```
Demon Docs creates `internal/runtime/README.md`:
```markdown
# Runtime
This index maps code files to their documentation.

<!-- doc-ledger:reverse-index:start -->
## Code Files
- [service.go](service.go)
  - [Runtime](../../docs/architecture/runtime.md)

<!-- doc-ledger:reverse-index:end -->
```
Reverse indexes project explicit targets; they do not invent documentation ownership.
### 7. Inspect codemap suggestions
Codemap generation is explicit. It does not run through normal reconciliation, watch, or the repository demon.
```bash
ddocs codemaps inspect --root docs/architecture/runtime.md
```
The fixture produces an additional folder candidate:
```text
docs/architecture/runtime.md
  section: existing
  changed: true
  add internal/runtime/ score=5.262 tier=context
    evidence: sibling_of_existing_target:internal/runtime/service.go:
    evidence: unique_basename_mention:runtime
```
Preview the exact write:
```bash
ddocs codemaps fix \
  --root docs/architecture/runtime.md \
  --dry-run
```
Review persisted suggestions when a relationship needs a decision:
```bash
ddocs suggestions
ddocs suggestions show SUGGESTION_ID
```
Decline an unwanted relationship with a durable reason:
```bash
ddocs suggestions decline SUGGESTION_ID \
  --reason "Not part of this document's implementation boundary"
```
An unchanged decline remains suppressed. It does not remove a link already present in the document.

For this walkthrough, apply the inspected change:
```bash
ddocs codemaps fix --root docs/architecture/runtime.md
ddocs codemaps check --root docs/architecture/runtime.md
```
The complete section becomes one managed codemap containing the file and folder targets.

Because the codemap changed, refresh its reverse projection:
```bash
ddocs fix --reverse --reverse-root internal
ddocs check --reverse --reverse-root internal
```
### 8. Create a schema-backed document
Schema-based creation works even while continuous frontmatter and format enforcement remain disabled for existing files. First set a useful default:
```toml
[frontmatter]
default_author = "Acme Docs Team"
```
Create a guide:
```bash
ddocs new general docs/guides/troubleshooting.md
```
The new file contains deterministic identity metadata and the configured structure:
```markdown
---
author: Acme Docs Team
created: "YYYY-MM-DD"
document_id: UUIDV7
document_type: general
policy_exempt: false
summary: TODO
---
# Troubleshooting
## Purpose
TODO
## Overview
TODO
```
Replace the placeholders, then update the folder index:
```bash
ddocs fix --docs
```
The new guide still needs meaningful inbound evidence. Add this to `installation.md`:
```markdown
For common failures, see [troubleshooting](troubleshooting.md).
```
Refresh and verify:
```bash
ddocs fix --links
ddocs check --docs --links
```
### 9. Final verification
Run the checks for each adopted surface:
```bash
ddocs check --docs --links
ddocs codemaps check --root docs/architecture/runtime.md
ddocs check --reverse --reverse-root internal
```
The documentation and reverse checks report `ddocs check passed`. The explicit codemap check reports `ddocs codemaps check passed`.

Repository-visible changes are limited to managed index regions, rewritten link destinations from the explicit move, one schema-created Markdown file, one managed codemap, and one managed reverse-index file. Private link, review, identity, and transaction state remains under `.ddocs/`.

Demon Docs does not replace Markdown, silently choose ambiguous targets, infer semantic ownership from arbitrary prose, or run codemap generation in the background.
## Related docs
- [Getting Started](getting-started.md)
- [Using Document Schemas](document-schemas.md)
- [Document Health Checks](document-health-checks.md)
- [Document Refactoring](document-refactoring.md)
- [Adopting Reverse Indexes](reverse-indexes.md)
- [Managing Codemaps](managing-codemaps.md)
- [Reviewing Suggestions and Changes](reviewing-suggestions-and-changes.md)
- [CLI Reference](../reference/cli.md)
- [Configuration Reference](../reference/configuration.md)
## Notes
The safest adoption order is narrow and reviewable: initialize, select one managed surface, inspect the diff, establish state, and widen enforcement only after the repository's conventions are represented explicitly.
