---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-73b7-b938-fd8d843dd5ac
document_type: general
policy_exempt: false
summary: This document defines the deterministic evidence boundary that identifies repository targets a documentation codemap may be missing without deciding existing-link irrelevance or owning source mutation.
---
# Codemap Missing-Link Evidence

Parent index: [Demon Docs Documentation](./README.md)

## Purpose

This document defines the deterministic evidence boundary that identifies repository targets a documentation codemap may be missing without deciding existing-link irrelevance or owning source mutation.

## Overview

The evidence layer receives normalized repository facts and emits candidate targets with canonical evidence records and fingerprints. It is independent of one Markdown codemap layout and does not write documents.

Production codemap execution consumes the ranked evidence through `internal/codemaprecommend` and `internal/codemaprun`. All selected non-declined tiers are eligible for insertion by the explicit foreground codemap command. Existing-link retention or pruning remains a separate execution policy.

## Inputs

The collector accepts normalized facts supplied by the codemap parser, repository inventory, Git reader, and language adapters:

- document path and visible text;
- repository file paths;
- existing codemap targets, used as seeds and exclusions;
- observed dependency edges;
- declared symbols;
- bounded Git commit path sets; and
- related documents with their current targets.

For production generation, the current codemap section is stripped from the document text before mention evidence is collected. Existing targets cannot become missing-link evidence merely by appearing in the map.

The collector remains independent of a particular section heading, marker prefix, fenced-list convention, or bullet syntax.

## Evidence signals

The implemented deterministic signals are:

- exact repository-relative path mentions;
- unique basename mentions;
- direct siblings of current targets;
- source/test counterparts;
- direct observed dependency neighbors;
- declared-symbol mentions;
- document/code and current-target/code Git co-change counts; and
- targets shared by related documents.

Each candidate retains evidence kind, source, detail, count, and a deterministic evidence fingerprint. Admission, scoring, tier assignment, decline replay, and source mutation are separate layers.

## Candidate exclusions

The evidence collector excludes:

- the current document itself;
- targets already visible in the current codemap;
- ambiguous basename or symbol matches that cannot resolve uniquely; and
- facts that cannot be established without guessing.

The ranking layer then applies its own admission and narrow negative-evidence rules.

## Confidence tiers

Ranked candidates are separated without discarding the broader relationship set:

- `hard_link` identifies a bounded stronger-confidence relationship. At most five candidates per document receive this tier.
- `context` identifies a weaker, indirect, optional, or already-explicit relationship that still survived admission and negative-evidence filtering.

Current `hard_link` qualification includes declared-symbol evidence; repeated exact paths independently corroborated by dependency or symbol evidence; supported source/test counterparts; dependency-neighbor evidence at score 18 or greater; and related-document evidence reinforced by direct document co-change.

A tier does not declare an existing link irrelevant. Both tiers are eligible for automatic addition by explicit codemap execution after shared decline-policy filtering. The tier remains visible in inspection, review state, benchmarks, and optional low-score pruning.

When `remove_low_score_links` is enabled, an existing hidden target recovered only as `context` may be selected for removal. That removal decision belongs to the execution layer, not the evidence collector.

Schema-1 reports created before tiers existed may omit the field. Evaluation treats that legacy empty value as `context` and rejects unknown non-empty tier values.

## Negative evidence

The ranking layer rejects only narrowly demonstrated incidental targets:

- dependency lockfiles when exact-path or basename evidence has no independent support;
- deeply nested assets, examples, fixtures, samples, or test data produced only by unique-basename matching; and
- children of `.github/workflows/` produced only by unique-basename matching.

Explicit path evidence preserves nested content and workflow targets. Declared symbols, dependencies, related documents, siblings, test counterparts, and Git co-change preserve supported lockfile or basename candidates.

These filters remove demonstrated noise without treating a file class as universally irrelevant.

## Fingerprints and declines

The evidence fingerprint is derived from canonical candidate evidence. It supports deterministic review replay:

- an unchanged declined relationship remains suppressed;
- materially changed evidence may produce a new fingerprint;
- a new fingerprint may be reconsidered through the shared review workflow; and
- decline state suppresses a proposed addition, not an existing codemap entry.

The evidence layer does not persist the decision. `internal/review` owns policy storage and replay.

## Safety contract

- Existing codemap targets are never returned as missing-link candidates.
- Evidence does not establish universal semantic truth or documentation completeness.
- The collector has no existing-link removal or irrelevance signal.
- Production source mutation occurs only through explicit codemap execution.
- Confidence-based existing-link pruning is separately configured and disabled by default.
- Declines are keyed by document, target, relationship kind, and evidence fingerprint.
- Identical normalized inputs produce byte-stable candidate ordering, evidence ordering, counts, and fingerprints.

## Code map

- `internal/evidence/model.go` — evidence kinds, candidates, inputs, and fingerprints.
- `internal/evidence/collect.go` — candidate aggregation and exclusions.
- `internal/evidence/mentions.go` — exact path and unique basename evidence.
- `internal/evidence/structure.go` — sibling, counterpart, and dependency evidence.
- `internal/evidence/symbols.go` — declared-symbol evidence.
- `internal/evidence/history.go` — bounded Git co-change evidence.
- `internal/codemaprecommend/` — admission, ranking, negative evidence, and tiers.
- `internal/codemaprun/` — decline replay, addition planning, and optional pruning.
- `internal/review/` — persisted decision lifecycle.

## Tests

Focused evidence and ranking coverage includes token boundaries, unique resolution, symbol ambiguity, current-target exclusion, candidate fingerprints, repeated mentions, fan-out discount, negative-evidence rules, tier limits, and deterministic ordering.

```bash
go test ./internal/evidence ./internal/codemaprecommend -count=1
```

## Related docs

- [Codemap Missing-Link Algorithm](codemap-suggestion-algorithm.md)
- [Codemap Managed Execution](architecture/codemap-managed-execution.md)
- [Codemap Evidence and Ranking](architecture/codemap-evidence-and-ranking.md)
- [Codemap Pipeline](architecture/codemap-pipeline.md)
- [Review Ledger](architecture/review-ledger.md)
- [Codemap Algorithm Development Log](codemap-algorithm-development-log.md)

## Notes

Evidence is an explainable deterministic input to policy. It is not a substitute for repository-specific review, and measured quality remains dependent on the repository population and document convention.
