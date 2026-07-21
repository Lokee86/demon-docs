---
author: brian
created: "2026-07-20"
document_id: 019f7d55-31e4-7d8f-9bb4-4e2d9cf9e6b1
document_type: general
policy_exempt: false
summary: Durable incremental cache for clean frontmatter and document-format validation.
---
# Validation Cache

Parent index: [Architecture](./INDEX.md)

## Purpose

Frontmatter and document-body format validation use a durable incremental cache for unchanged clean documents. The cache is an optimization boundary: authored Markdown, diagnostics, repair decisions, immutable records, and schema history remain authoritative.

## Record identity

Each record is stored through `ddrepo` under `.ddocs/` and is addressed by the normalized repository-relative path. A record is reusable only when all of these values match:

- normalized path;
- raw document content SHA-256;
- validation engine version;
- effective frontmatter policy hash;
- effective selected shared/document schema hash; and
- the immutable-value snapshot used by frontmatter validation.

The selected schema identity is retained in the record so an unchanged document can verify the current shared and document-specific schema sources without reparsing its Markdown frontmatter. Shared and document-specific schema source changes therefore invalidate the record; changing the engine version also invalidates all records.

Scoped watcher validation uses a normalized-path lookup that does not require a
new content hash for untouched files. Frontmatter scoped reuse requires a
current `FrontmatterClean` entry and reuses its cached document identity for
global duplicate-ID detection. Document-format scoped reuse requires a current
`FormatClean` entry and reuses its cached schema name and document ID. If any
active untouched document lacks the required clean state, the watcher reruns
that subsystem with the full builder.

Because the cache currently keys reuse to the raw whole-document SHA-256, a generated link rewrite, index rewrite, or other body-only edit invalidates both frontmatter and document-format cache entries even when their relevant metadata and heading structure are unchanged. The following validation pass therefore performs a cold parse for that document. A future optimization may split cache identity by owned input surface or refresh the affected validation records from the final published bytes after generated rewrites.

## Clean-only reuse

Only a document with no diagnostics and no pending repair is recorded as clean. A cache hit restores no authored output; it contributes the same empty diagnostic result and skips document parsing and evaluation. Within one validation pass, selected schema-source hashes, shared schemas, and schema history are memoized and reused across documents selecting the same inputs. A later pass creates a fresh snapshot and therefore observes schema-file changes.

Frontmatter cache candidates retain the document ID and immutable values needed to preserve duplicate-ID detection and immutable publication. If duplicate IDs are present, all affected candidates are reparsed before evaluation. A fix can also publish immutable values retained by a clean check cache hit.

YAML and TOML behavior, diagnostic ordering, and deterministic repair ordering are unchanged. Cache failure is an operational error because a corrupt or unavailable private-state repository must not be silently mistaken for durable validation state.

## Mutation boundary

`check` may update cache records only when the repository already has an initialized `.ddocs` private object store; standalone checks do not initialize one merely to persist cache data. It never writes authored Markdown or schema files. `fix` writes the same cache records in addition to its existing guarded authored-file and immutable-state publication. Cache records use the existing `ddrepo` transaction and do not create commits on the user's normal Git branch.

Frontmatter and document-format `check` planners share one command-scoped cache store. The store synchronizes lookup, retention, and merge operations while the two planners run concurrently. Cache entries are cloned at the store boundary so callers cannot mutate shared map state. The merged dirty set is published once, serially, only after all selected check planners finish successfully.

Each validation pass also removes cache records whose normalized paths are no longer in the active Markdown scope. A rename or deletion therefore does not leave permanently reachable stale records. Re-merging an identical cache entry does not publish a private-state transaction, and a repeated clean frontmatter fix does not republish immutable values that already match durable state.

## Related docs

- [Front Matter Schemas](../reference/frontmatter.md)
- [Document Schemas And Format Enforcement](../reference/document-schemas.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Private Object Repository](private-object-repository.md)

## Notes

The cache is an optimization over authoritative source and private-state records; deleting it is safe after active validation processes have stopped, although the next validation pass reparses every document.
