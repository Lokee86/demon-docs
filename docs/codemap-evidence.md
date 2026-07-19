# Codemap Missing-Link Evidence

This package boundary collects deterministic evidence that a documentation codemap may be missing a repository target. It does not parse codemaps, decide that a link is semantically correct, create repository-graph edges, or recommend removing an existing link.

## Inputs

The collector accepts normalized facts supplied by the codemap parser, repository inventory, Git reader, and later language adapters:

- the document path and text;
- repository file paths;
- existing codemap targets, used only as seeds and exclusions;
- observed dependency edges;
- bounded Git commit path sets; and
- related documents with their already accepted targets.

The collector is intentionally independent of any one Markdown codemap layout. Space Rocks codemaps can supply the same input model as any future adapter.

## Evidence Signals

The initial deterministic signals are:

- exact repository-relative path mentions;
- unique basename mentions;
- direct siblings of accepted targets;
- source/test counterparts;
- direct observed dependency neighbors;
- document/code and accepted-target/code Git co-change counts; and
- accepted targets shared by related documents.

Each candidate retains its evidence kind, source, detail, count, and a deterministic evidence fingerprint. Scoring and acceptance are separate layers.

## Suggestion Tiers

Ranked candidates are separated without discarding the broader relationship set:

- `hard_link` identifies a bounded, high-confidence surface for direct codemap review. At most five qualifying candidates are selected per document from the retained 30-candidate pool, even when weaker context candidates rank above them. A candidate currently qualifies through a declared-symbol mention; a test target with independent dependency, related-document, or sibling support; a non-test implementation counterpart with the same independent support and score at least 20; dependency-neighbor evidence with score at least 18; or related-document evidence reinforced by direct Git co-change with the current document.
- `context` identifies weaker, indirect, or already-explicit relationships that can still improve bounded agent context. Exact path mentions remain context because the document already exposes the target directly, unsupported source/test counterparts remain context rather than qualifying by filename structure alone, and lower-scoring implementation counterparts remain context until their combined evidence clears the stricter production-file threshold.

A tier is not an automatic write decision. It does not declare an existing link irrelevant, and it does not bypass persisted declines or human review. Schema-1 reports created before tiers existed may omit the field; evaluation treats that legacy empty value as `context` and rejects unknown non-empty tier values.

## Safety Contract

- Existing codemap targets are never returned as missing-link candidates.
- Evidence creates a reviewable suggestion candidate, never documentation coverage or a graph edge.
- The collector has no removal or irrelevance signal.
- Declined suggestions should be stored by document, target, and evidence fingerprint. The same fingerprint remains suppressed; materially changed evidence produces a different fingerprint that may be reconsidered.
- Identical inputs produce byte-stable candidate ordering, evidence ordering, counts, and fingerprints.
