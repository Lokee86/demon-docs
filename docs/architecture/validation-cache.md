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

Each record is stored through `ddrepo` under `.ddocs/` and is addressed by the normalized repository-relative path. The record contains independent reusable identities for the two validation subsystems.

Frontmatter reuse requires:

- normalized path;
- validation engine version;
- raw leading-frontmatter identity;
- effective frontmatter policy hash;
- effective selected shared/document schema hash; and
- the immutable-value snapshot used by frontmatter validation.

Document-format reuse requires:

- normalized path;
- validation engine version;
- format-source identity;
- format-selection policy hash; and
- effective selected shared/document schema hash.

The raw whole-document SHA-256 remains recorded as source-state metadata and as a stale-write guard, but it is no longer the shared reuse identity for both validators. Frontmatter identity hashes the exact leading YAML or TOML block, including its source line endings. Ordinary body prose, link, and heading edits therefore do not invalidate a clean frontmatter result. The identity also distinguishes no block, unterminated blocks, and a newly introduced second leading block so cached clean results cannot hide parser diagnostics.

Document format owns a separate structural identity and separate policy/schema hashes. Its identity includes the selected schema name, document ID and type, validation engine version, and the evaluated H2+ heading tree. The nested tree records heading text, levels, hierarchy, order, and duplicate occurrences. H1 titles, ordinary prose, links, fenced or protected headings, code blocks, and section body content are excluded because the current schema engine does not evaluate them.

Known generated rewrites refresh the raw content hash from the exact final published bytes. Rewrite owners also declare which validation surfaces they may affect:

- link destination rewrites retain both clean frontmatter and document-format results;
- index generation retains frontmatter results and invalidates document-format results because managed headings may change;
- document-format repairs retain frontmatter results and invalidate format results; and
- frontmatter repairs invalidate both results because document identity, type, and schema selection may change.

The refresh is guarded by the expected old content hash, so a stale cache record is never carried across an unrelated edit. If every clean result is invalidated, the record is removed; otherwise the unaffected subsystem identity remains reusable under the final raw content hash.

Scoped watcher validation uses a normalized-path lookup that does not require reading or hashing untouched files. Frontmatter scoped reuse requires a current `FrontmatterClean` entry with the active frontmatter policy and reuses its cached document identity for global duplicate-ID detection. Document-format scoped reuse requires a current `FormatClean` entry with the active format policy and reuses its cached schema name and document ID. If any active untouched document lacks the required clean state, or a duplicate ID spans a changed and untouched document, the watcher reruns that subsystem with the full builder.

The selected schema metadata remains in the record so either subsystem can verify current shared and document-specific schema sources before reuse. Shared or document-specific schema changes invalidate only the matching subsystem identity. The structural fingerprint has its own version marker, so older whole-document format identities miss safely and rebuild without another durable-record schema change. Validation cache schema version 2 continues to discard older combined-identity entries safely.

## Clean-only reuse

Only a document with no diagnostics and no pending repair is recorded as clean. A cache hit restores no authored output; it contributes the same empty diagnostic result and skips document parsing and evaluation. Within one validation pass, selected schema-source hashes, shared schemas, and schema history are memoized and reused across documents selecting the same inputs. A later pass creates a fresh snapshot and therefore observes schema-file changes.

Frontmatter cache candidates retain the document ID and immutable values needed to preserve duplicate-ID detection and immutable publication. If duplicate IDs are present, all affected candidates are reparsed before evaluation. A fix can also publish immutable values retained by a clean check cache hit.

YAML and TOML behavior, diagnostic ordering, and deterministic repair ordering are unchanged. Cache failure is an operational error because a corrupt or unavailable private-state repository must not be silently mistaken for durable validation state.

## Mutation boundary

`check` may update cache records only when the repository already has an initialized `.ddocs` private object store; standalone checks do not initialize one merely to persist cache data. It never writes authored Markdown or schema files. `fix` writes the same cache records in addition to its existing guarded authored-file and immutable-state publication. Cache records use the existing `ddrepo` transaction and do not create commits on the user's normal Git branch.

Generated-rewrite refresh happens only after the authored bytes and the rewrite owner's required durable state have published successfully. Cache refresh failure is reported as an operational error, but the cache remains derived state: a missing or stale record only forces later revalidation and never authorizes an authored-file write.

Frontmatter and document-format `check` planners share one command-scoped cache store. The store synchronizes lookup, retention, and merge operations while the two planners run concurrently. Cache entries are cloned at the store boundary so callers cannot mutate shared map state. The merged dirty set is published once, serially, only after all selected check planners finish successfully.

Each validation pass also removes cache records whose normalized paths are no longer in the active Markdown scope. A rename or deletion therefore does not leave permanently reachable stale records. Re-merging an identical cache entry does not publish a private-state transaction, and a repeated clean frontmatter fix does not republish immutable values that already match durable state.

## Related docs

- [Front Matter Schemas](../reference/frontmatter.md)
- [Document Schemas And Format Enforcement](../reference/document-schemas.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Private Object Repository](private-object-repository.md)

## Notes

The cache is an optimization over authoritative source and private-state records; deleting it is safe after active validation processes have stopped, although the next validation pass reparses every document.
