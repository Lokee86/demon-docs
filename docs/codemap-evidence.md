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

## Safety Contract

- Existing codemap targets are never returned as missing-link candidates.
- Evidence creates a reviewable suggestion candidate, never documentation coverage or a graph edge.
- The collector has no removal or irrelevance signal.
- Declined suggestions should be stored by document, target, and evidence fingerprint. The same fingerprint remains suppressed; materially changed evidence produces a different fingerprint that may be reconsidered.
- Identical inputs produce byte-stable candidate ordering, evidence ordering, counts, and fingerprints.
