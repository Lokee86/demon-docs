---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-74a7-b456-2b3b0aa24a6e
document_type: general
policy_exempt: false
summary: This document is the durable chronological record of the codemap missing-link algorithm, its benchmark program, tuning decisions, rejected experiments, and measured baselines.
---
# Codemap Algorithm Development Log

Parent index: [Archivist Documentation](./INDEX.md)

## Purpose

This document is the durable chronological record of the codemap missing-link algorithm, its benchmark program, tuning decisions, rejected experiments, and measured baselines.

The maintained description of current behavior lives in [Codemap Suggestion Algorithm](codemap-suggestion-algorithm.md). Research artifacts remain under `research/` and are linked below.

The work was developed on July 18–19, 2026. The final algorithm baseline covered by this log is `b7dfc598c9a158e29ba9e9167dbf2fa6016b80d1`.

## Overview

This is a historical development record, not the current codemap command contract. Current extraction, ranking, managed execution, and evaluation behavior belongs to the linked architecture, reference, and research owners below; retained measurements remain tied to their stated repositories and pinned artifacts.

## Fixed Design Decisions

The development program began with several constraints that remained unchanged:

- Suggest potentially missing semantic links only.
- Never suggest that an existing link is irrelevant or should be removed.
- Preserve declined suggestions and suppress the same evidence fingerprint in future; reconsider only after material evidence change.
- Separate deterministic evidence collection from scoring, tiering, human judgment, and graph mutation.
- Treat high-confidence direct-link review and useful agent context as different outputs.
- Keep monolithic per-file indexes separate from ordinary codemap calculations.
- Prefer narrow evidence rules over repository-specific special cases or global threshold chasing.

## Development Phases

### Phase 1: Format inventory and trusted evidence foundation

The first phase established that the system could learn from Space Rocks without hard-coding the Space Rocks Markdown layout.

The work inventoried codemap conventions, normalized authored document-to-code pairs, created a trusted Space Rocks review set, and added the deterministic evidence collector.

Evidence included explicit paths, unique basenames, accepted-target siblings, source/test counterparts, dependency neighbors, Git co-change, and targets inherited from related documents.

Key outcome: evidence became a reviewable fact layer rather than an automatic coverage decision.

### Phase 2: Benchmark harness and repository adapters

The second phase built a repeatable hidden-link benchmark:

1. Extract known authored links.
2. Hide a deterministic subset.
3. Run the evidence and ranking pipeline without those links.
4. Measure which hidden links reappear and at what tier.
5. Export canonical text and JSON reports.

The benchmark CLI, repository corpus adapter, report exporters, orchestration, and evidence validation landed as independent streams and were then reconciled.

Key outcome: tuning could be measured against frozen inputs rather than anecdotal inspection.

### Phase 3: Ranking and stronger semantic evidence

The first ranked algorithm assigned evidence weights, applied logarithmic repetition and fanout controls, and bounded output per document.

Declared-symbol evidence was then added as the strongest direct semantic signal. This allowed a document that names an implementation symbol to identify its owner without relying only on paths or repository shape.

Key outcome: the algorithm moved from an unordered evidence inventory to a deterministic suggestion surface.

### Phase 4: Space Rocks authored-links precision benchmark

The hidden-link benchmark measured recovery but could not measure user-facing precision because every suggestion outside the hidden positive set looked false.

A separate authored-links benchmark therefore left existing links visible, generated genuinely new suggestions, and manually labeled a deterministic 150-item sample:

- 70 `valid_missing_link`;
- 62 `plausible_but_unnecessary`;
- 18 `incorrect`.

Initial aggregate results:

| Metric | Initial result |
|---|---:|
| Strict precision | 46.67% (70/150) |
| Relevance/acceptance | 88.00% (132/150) |
| Precision@1 | 60.00% |
| Precision@3 | 54.67% |
| Precision@5 | 52.80% |

The result showed that one binary suggestion list conflated direct missing links with useful but optional context.

### Phase 5: Hard-link and context tiers

The first tuning pass retained the whole relationship set but divided it into:

- `hard_link`: bounded direct-link review; and
- `context`: weaker, indirect, optional, or already-explicit relationships.

Initial hard-link qualification used declared symbols, test counterparts, or sufficiently strong dependency evidence, with at most five hard links per document.

Results:

| Metric | Pass 1 |
|---|---:|
| Hard-link suggestions | 81 |
| Hard strict precision | 64.20% (52/81) |
| Hard relevance | 95.06% (77/81) |
| Labeled-valid hard recovery | 74.29% (52/70) |
| Full-pool hard links | 602 |
| Hidden-link holdout | 10/10 |

Key decision: useful context should not be discarded merely because it is not appropriate for permanent insertion.

### Phase 6: Corroborated structural qualification

The second tuning pass tightened several paths:

- exact-path mentions stayed context instead of automatically becoming hard links;
- source/test counterparts required independent dependency, related-document, or sibling support;
- related-document targets could qualify when direct Git document co-change corroborated them; and
- the five-item cap counted qualifying candidates rather than allowing weak context items to consume hard-link positions.

Results:

| Metric | Pass 2 |
|---|---:|
| Hard-link suggestions | 70 |
| Hard strict precision | 72.86% (51/70) |
| Hard relevance | 97.14% (68/70) |
| Labeled-valid hard recovery | 72.86% (51/70) |
| Full-pool hard links | 631 |
| Hidden-link holdout | 10/10 |

Strict precision increased substantially while losing one labeled-valid hard candidate.

### Phase 7: Directional counterpart confidence

The third tuning pass distinguished test verification targets from production implementation targets:

- dependency-only hard qualification required score 18 rather than 16;
- supported test counterparts could still qualify directly; and
- non-test implementation counterparts required score 20 in addition to independent support.

Results:

| Metric | Pass 3 |
|---|---:|
| Hard-link suggestions | 68 |
| Hard strict precision | 75.00% (51/68) |
| Hard relevance | 98.53% (67/68) |
| Labeled-valid hard recovery | 72.86% (51/70) |
| Full-pool hard links | 621 |
| Hidden-link holdout | 10/10 |

This became the stable Space Rocks precision baseline.

### Phase 8: Cross-repository benchmark

Space Rocks was not sufficient evidence for a repository-agnostic claim. A separate corpus was built from pinned open-source repositories with explicit document-to-code mappings.

The calculation corpus covered five ordinary repositories and multiple languages. A sixth repository, gbrain, used one monolithic per-file index with hundreds of targets. It was classified as a stress case because hiding the index removed nearly all topical evidence.

Initial ordinary-corpus recovery:

- 18 hidden links;
- 11 recovered;
- one hard recovery;
- ten context recoveries.

The first cross-repository tuning rule promoted an exact path only when:

- it appeared at least twice; and
- dependency or declared-symbol evidence independently corroborated it.

After tuning:

| Metric | Before | After |
|---|---:|---:|
| Total recovered | 11/18 | 11/18 |
| Hard recovered | 1 | 4 |
| Context recovered | 10 | 7 |
| Primary recovery | 6/8 | 6/8 |
| Primary hard recovery | 0 | 2 |
| Stress recovery | 3/10 context | 3/10 context |

This was a confidence-tier improvement, not a recall increase.

### Phase 9: Cross-repository manual precision review

Positive-link recovery still could not determine whether unmatched suggestions were useful. A deterministic sample of 121 unmatched suggestions was manually labeled across the five ordinary repositories.

The split was frozen before tuning:

- tuning: agent-orchestrator, beads-rust, Genesis, render-claude-context;
- validation: Bifrost.

Initial results:

| Scope | Reviewed | Valid | Plausible | Incorrect | Strict | Relevance |
|---|---:|---:|---:|---:|---:|---:|
| Overall | 121 | 83 | 34 | 4 | 68.60% | 96.69% |
| Tuning | 93 | 64 | 25 | 4 | 68.82% | 95.70% |
| Bifrost | 28 | 19 | 9 | 0 | 67.86% | 100.00% |
| Hard tier | 3 | 3 | 0 | 0 | 100.00% | 100.00% |
| Context tier | 118 | 80 | 34 | 4 | 67.80% | 96.61% |

The three hard suggestions all came from Bifrost, so the 100% result was explicitly not treated as a broad hard-tier precision estimate.

The four errors were:

- unsupported `Cargo.lock`;
- two deeply nested asset/source directories matched by generic basenames; and
- workflow scripts matched by the generic basename `scripts`.

### Phase 10: Narrow incidental-target rejection

The final pass added negative evidence for only the demonstrated failure classes.

Rules:

- suppress dependency lockfiles without evidence beyond exact path or unique basename;
- suppress deeply nested asset/example/fixture/sample/test-data targets produced only by unique-basename matching; and
- suppress `.github/workflows/` children produced only by unique-basename matching.

Nested-content and workflow targets with explicit paths remain. Lockfiles and basename matches remain when independently corroborated.

Final fixed-sample result:

| Scope | Valid retained | Plausible retained | Incorrect suppressed | Retained strict | Retained relevance |
|---|---:|---:|---:|---:|---:|
| Overall | 83 | 34 | 4 | 70.94% | 100.00% |
| Tuning | 64 | 25 | 4 | 71.91% | 100.00% |
| Bifrost | 19 | 9 | 0 | 67.86% | 100.00% |

No reviewed valid or plausible suggestion changed tier or disappeared.

Four replacement context candidates surfaced after the removals: `rust-toolchain.toml` and three Genesis solver implementations. Manual inspection found all four to be direct owners explicitly named by their documents.

Cross-repository recovery remained 11/18 with four hard recoveries. Bifrost remained 2/3, both hard. The separate index stress result remained 3/10, all context.

Space Rocks remained unchanged at:

- 75.00% hard strict precision;
- 98.53% hard relevance;
- 51/70 labeled-valid hard recovery;
- 621 hard and 3,872 context candidates in the full source pool; and
- 10/10 canonical hidden-link recovery.

### Phase 11: Production policy and authored-provenance repair

August 27 dogfooding against Archivist exposed that the retained research tiers had been connected to a broader mutation policy than their labels justified. `context` recommendations were being written as permanent links, and resolved glob members were flattened into ordinary existing-file seeds. An authored pattern such as `internal/app/codemap_*.go` could therefore fan out through siblings, dependencies, related documents, and Git history.

The production repair separated these concerns:

- only non-declined `hard_link` recommendations are eligible for automatic insertion;
- `context` remains inspectable but non-mutating;
- file, directory, pattern, and symbol provenance survives corpus projection;
- authored directories cover their descendants instead of causing per-file re-suggestions;
- resolved patterns cover their matches without turning each match into an independent expansion seed;
- only explicitly authored resolved files seed sibling, dependency, test-counterpart, and target-history expansion; and
- basename-only patterns constrain inferred non-matching siblings in their literal parent directory unless the current document supplies direct path, basename, or symbol evidence.

The immediate Archivist self-test on `docs/architecture/codemap-pipeline.md` moved from 30 would-be additions before the repair, to five hard-link additions after the mutation-policy split, to zero permanent additions after provenance and scope-boundary preservation. Context candidates remain visible for later ranking work.

The older precision and recovery numbers above are retained as historical baselines. They have not yet been regenerated against the repaired production algorithm and must not be presented as post-repair measurements.

### Phase 12: Code-intelligence provider boundary

The next step separated repository semantics from Archivist's local language parsers without changing ranking policy. `internal/codemapcorpus` now exposes a narrow `CodeIntelligenceProvider` contract that returns repository-local dependency and declared-symbol facts. The existing shallow parsers are wrapped as the default local provider rather than being hard-wired into corpus construction.

The corpus retains ownership of trust and determinism at the boundary: provider paths are validated against the current repository-file inventory, malformed or out-of-scope paths fail construction, facts are normalized/deduplicated/sorted before publication, and provider failures are not silently mixed with fallback results. `BuildContext` propagates caller cancellation through production, benchmark, and precision paths; the older `Build` entry point remains as a background-context compatibility wrapper.

No Arcana or Lexicon transport was added in this phase. The purpose of the seam is to make that integration replace the semantic fact source later without moving evidence scoring, review policy, codemap coverage, or mutation authority out of Archivist.

### Phase 13: Pinned Arcana file and symbol resolution

The next phase wired authored target resolution to Arcana without yet consuming Arcana relationship edges. Archivist now speaks the `arcana.query.v1` JSONL protocol through `resolve_file` and `resolve_symbol`, retaining Arcana's durable external node identity and source span when a target resolves uniquely.

Trust is deliberately narrower than snapshot existence. `.arcana/CURRENT` must equal `.lexicon/CURRENT`; the Arcana snapshot must name that Lexicon snapshot; and the content-addressed Lexicon manifest must verify against its published ID. Path-qualified queries additionally require the current source SHA-256 to equal the Lexicon manifest content ID. Standalone/global symbol resolution requires a clean repository at the Git head recorded when Lexicon state was prepared. Stale or unavailable semantic state therefore degrades to the pre-existing `symbol_unverified` or `unsupported` outcomes instead of being treated as current truth.

Path-qualified and standalone symbols can now produce real `resolved`, `missing`, or `ambiguous` dataset states. Plain file targets remain filesystem-authoritative but can carry Arcana semantic identity when the snapshot can safely attest them. Once a current Arcana protocol session is open, transport or query failure is fatal rather than silently mixing partial semantic results with fallback behavior.

Arcana relationship expansion remains intentionally deferred to the following evidence phase.

### Phase 14: Bounded Arcana relationship evidence

Arcana graph relationships are now consumed through a separate per-document `RelationshipProvider` rather than being injected into the repository-wide code-intelligence corpus. This preserves the Step 3 provider boundary and prevents a hidden benchmark answer from influencing which graph neighborhood is fetched.

Only currently visible exact file targets and Step 4-verified symbol targets seed Arcana expansion. Directory and pattern targets do not. The provider projects one-hop `calls`, `imports`, `depends-on`, `implements`, `extends`, `overrides`, `uses-trait`, `includes`, and `tests` relationships back to current repository file pairs. Source-content freshness is rechecked for both seed and neighbor paths.

Expansion is deliberately bounded: at most 128 relation-capable nodes per seed and 128 neighbors per node/direction. A truncated seed neighborhood is discarded rather than partially trusted. Returned relationships become a distinct `semantic_relationship` evidence kind with weight 3. They can surface and rank context recommendations, but they do not qualify a `hard_link` and their score is excluded from numeric hard-link thresholds.

Live dogfooding used a freshly prepared matching Lexicon/Arcana snapshot for Archivist. `codemap-pipeline.md` still produced zero additions/removals. On `codemap-extraction-and-dataset.md`, Arcana emitted real call/implements relationship evidence and changed context scores/order, while the automatic hard-link set remained exactly the same five files as the fallback run. This is a mutation-isolation check, not a precision claim.

### Phase 15: Deterministic candidate roles

The next phase separated **relationship type** from **confidence tier**. Every admitted suggestion now carries one of five deterministic roles:

- `primary_implementation`;
- `supporting_implementation`;
- `verification_test`;
- `interface_boundary`; or
- `context_only`.

Classification is evidence-derived and deliberately non-probabilistic. Recognized test/spec paths and incoming Arcana `tests` edges become verification. Outgoing `implements`, `extends`, `overrides`, `uses-trait`, and `includes` edges identify the candidate as the contract/base/trait side and therefore `interface_boundary`. Direct symbol/path/basename evidence identifies primary implementation. Dependency, semantic, counterpart, sibling, and related-document evidence identifies supporting implementation when no stronger role applies. Remaining retained candidates are context-only.

The role field is orthogonal to `hard_link`/`context`. This phase does not change weights, ranking order, hard-link thresholds, caps, or automatic mutation eligibility. The purpose is to provide the semantic buckets required for the following coverage-aware selection rewrite without combining that rewrite with classification mechanics.

Roles are emitted by inspect and benchmark report surfaces. Precision evaluation now includes `by_role` metrics and role sampling coverage, while legacy schema-1 suggestions with no role are interpreted as `context_only` for role-level evaluation. The precision helper was also corrected to recognize `semantic_relationship` at its implemented weight when selecting a primary evidence kind.

### Phase 16: Coverage-aware selection

The next phase replaced the flat top-30 cutoff with deterministic coverage-aware selection while leaving evidence weights unchanged. Candidates remain raw-score ordered into coarse `floor(log2(score))` bands. Higher bands are exhausted before lower ones; only candidates with comparable evidence magnitude compete on coverage.

Within one band, selection prefers uncovered non-context roles and uncovered target directories before redundant candidates. Target parent directory is used only as a lightweight implementation-seam proxy. The repeated exact-path reserve remains independent, and final inspect/report order remains raw score plus target path.

Hard-link qualification predicates and numeric thresholds were retained, but allocation became more conservative: `context_only` cannot promote, no semantic role may consume more than two of the five hard-link slots, and no target directory may consume more than three. This prevents one test family, one implementation role, or one dense package from monopolizing the permanent map while still allowing a package-centered document to retain several distinct seam representatives.

Focused synthetic coverage pins both sides of the policy: same-band role/directory candidates can survive a crowded cutoff, while lower score bands cannot displace stronger-band candidates.

Live Archivist dogfooding against a freshly rebuilt matching Lexicon/Arcana snapshot showed the intended hard-link redistribution on `docs/architecture/codemap-extraction-and-dataset.md`. Before the selection rewrite, its five hard links were four `verification_test` candidates plus one `primary_implementation`. After the rewrite, the five slots were two verification candidates, one primary implementation, and two supporting implementations. `docs/architecture/codemap-pipeline.md` still produced zero additions and zero removals. This is a behavioral/convergence check, not a precision claim. Fresh cross-corpus precision and recall measurement remains deferred to the final benchmark/tuning phase.

### Phase 17: Semantic staleness from Arcana snapshot diffs

Codemap generation now records a per-document semantic validation baseline in `.ddocs`: the accepted Arcana snapshot ID plus the document digest at that point. On later runs, unchanged documents are compared against the current Arcana snapshot with the protocol `diff` operation.

The detector is intentionally scoped to adopted mapped semantics. It reports mapped targets whose Arcana node disappeared, moved, changed identity/qualified ownership, changed definition metadata, became ambiguous, or changed outgoing logical relationships. Directory and glob abstractions are not promoted into semantic nodes for this purpose.

The result is advisory staleness, not pruning authority. `inspect` emits the mapped target and deterministic change kind. `check` treats a semantic-stale document as failing even when the managed codemap text is unchanged. Existing removal policy remains separate.

Baseline advancement is conservative. A missing baseline is initialized on successful `fix`; unchanged mapped semantics advance to the newest snapshot automatically. If mapped semantics changed while the document itself stayed byte-identical, `fix` does not advance the baseline, so repeated runs cannot clear the warning. Once the document is edited and `fix` succeeds, the current snapshot becomes its new accepted baseline.

### Phase 18: Symbol-level reverse projection from adopted targets

Reverse indexes now consume the same optional current Arcana target resolver used by codemap dataset construction. Ownership remains unchanged: only explicit targets already authored in codemap sections can create documentation backlinks. Raw Arcana graph relationships, neighbourhoods, dependencies, calls, and implementation edges are never reverse-index inputs.

A uniquely resolved authored symbol target is projected beneath its exact backing file using Arcana's stable node key/external identity, declaration kind, qualified name, and current source span. The symbol target satisfies reverse-index coverage for that backing file without being flattened into a generic file-level documentation backlink. Path-qualified symbols retain the existing explicit file-path fallback when semantic state is unavailable; standalone `symbol:...` targets require unique semantic resolution and ambiguous in-scope candidates remain diagnostics.

A live disposable Go repository with matching prepared Lexicon/Arcana state verified `service/runtime.go#Run` renders as a `function` symbol at line 3 with its documentation nested underneath. Removing Arcana state from the same fixture produced the conservative file-level fallback. The second Arcana-backed `check --reverse` was clean.

## Rejected or Revised Experiments

### Pooling the monolithic index with ordinary repositories

**Observed:** gbrain uses one index document with hundreds of file mappings. Hiding its authored index removes the document's topical evidence and tests a different retrieval problem.

**Decision:** exclude `stress` mode from ordinary aggregate calculations and report it separately.

### Broad dependency-lockfile suppression

**Observed:** the first negative rule also removed a Space Rocks `go.sum` candidate that had sibling and Git co-change support and had been reviewed as relevant context.

**Decision:** suppress lockfiles only when exact-path or basename evidence is unsupported. Corroborated lockfiles remain context.

### Generic test-counterpart demotion

**Observed:** broadly demoting test counterparts left hard precision at 75.00% but reduced labeled-valid hard recovery from 51/70 to 33/70.

**Decision:** reject the experiment. Test counterparts retain their prior behavior and require independent support for hard qualification.

### Global score reduction or broad context deletion

**Observed:** the manual review showed 34 plausible context candidates and only four outright errors. Broad deletion would improve strict precision by discarding useful relationships rather than improving semantic discrimination.

**Decision:** retain the context tier and tune only demonstrated negative-evidence classes.

### Further tuning against the same four errors

**Risk:** repeated tuning on the same fixed sample would overfit the existing corpus.

**Decision:** stop algorithm tuning and move to early product implementation testing. New repositories, scoped feature documents, and real accept/decline outcomes should precede another pass.

## Commit Ledger

The following commits form the algorithm and benchmark development chain.

| Commit | Change |
|---|---|
| `683a15d9` | Added deterministic missing-link evidence collection. |
| `b071e162` | Added the trusted Space Rocks codemap review set. |
| `529bed1a` | Inventoried Space Rocks codemap formats. |
| `b02686f8` | Added the hidden-link benchmark harness. |
| `09d587db` | Added repository dataset extraction and export. |
| `3fc77e5c` | Merged codemap format inventory. |
| `ba9a96f2` | Merged evidence collector stream. |
| `f246d023` | Merged benchmark harness stream. |
| `750085a0` | Merged trusted review-set stream. |
| `548cbaf0` | Reconciled extraction and benchmark streams. |
| `2e43bbc0` | Added benchmark report exports. |
| `447db213` | Added benchmark orchestration. |
| `5ba7dda7` | Added benchmark CLI contract. |
| `18a7f26d` | Added evidence-signal validation. |
| `0115aa11` | Added repository corpus adapter. |
| `278b0ec1` | Merged benchmark runner. |
| `b07278e8` | Merged benchmark reports. |
| `d5561011` | Merged evidence validation. |
| `a3c99fa2` | Merged benchmark CLI. |
| `e47c7ff9` | Integrated benchmark command. |
| `904ce82b` | Added deterministic ranked suggestions. |
| `802db193` | Added declared-symbol evidence. |
| `2420deaf` | Added the first curated precision benchmark. |
| `486aabd8` | Added the labeled precision benchmark implementation. |
| `d477759f` | Finalized authored-links precision artifacts. |
| `c7d09234` | Kept the large precision source report temporary. |
| `0ed2cd62` | Added `hard_link` and `context` tiers. |
| `6acbbbc5` | Tightened structural hard-link qualification. |
| `aa6eb48c` | Tightened directional counterpart confidence. |
| `95b3ed43` | Added the cross-repository benchmark corpus. |
| `6ea39964` | Promoted corroborated repeated references. |
| `a5f095f7` | Separated index stress and recorded wider tuning. |
| `2dac7740` | Added cross-repository manual precision review. |
| `3c98fedb` | Added incidental-target rejection. |
| `215cef7b` | Narrowed lockfile handling to preserve supported context. |
| `b7dfc598` | Reverted the harmful broad test-counterpart penalty. |
| `73657346` | Recorded final incidental-target tuning results. |

The commit subjects are not the complete specification. The current behavior is defined by the code and [Codemap Suggestion Algorithm](codemap-suggestion-algorithm.md).

## Artifact Registry

### Initial review and evidence validation

- `research/codemap-inventory/`: codemap format fixtures and normalized inventory.
- `research/codemap-review/`: trusted Space Rocks links and review findings.
- `research/codemap-evidence-validation/`: evidence validation record.

### Space Rocks precision

- `research/codemap-precision/space-rocks-precision-sample-150.json`: deterministic sample.
- `research/codemap-precision/space-rocks-precision-benchmark.json`: labels, rationales, references, and hashes.
- `research/codemap-precision/evaluation.json`: current precision evaluation.
- `research/codemap-precision/README.md`: pass-by-pass Space Rocks results and reproduction.

### Cross-repository recovery

- `research/cross-repo-codemap-benchmark/candidates.json`: discovery shortlist and extraction modes.
- `research/cross-repo-codemap-benchmark/corpus/`: normalized explicit mappings.
- `research/cross-repo-codemap-benchmark/datasets/`: benchmark inputs.
- `research/cross-repo-codemap-benchmark/reports/`: per-repository outputs.
- `research/cross-repo-codemap-benchmark/evaluation.json`: aggregate evaluation.
- `research/cross-repo-codemap-benchmark/results.md`: readable results.
- `research/cross-repo-codemap-benchmark/README.md`: workflow, modes, and pass history.

### Cross-repository precision

- `research/cross-repo-codemap-precision-review/sample-manifest.json`: frozen stratified sample with evidence metadata.
- `research/cross-repo-codemap-precision-review/labels.json`: blind review queue and completed labels.
- `research/cross-repo-codemap-precision-review/RUBRIC.md`: fixed label definitions.
- `research/cross-repo-codemap-precision-review/evaluation.json`: pre-tuning manual precision.
- `research/cross-repo-codemap-precision-review/tuning-pass-3-review-comparison.json`: fixed-sample survival check.
- `research/cross-repo-codemap-precision-review/tuning-pass-3-summary.json`: final validation summary.
- `research/cross-repo-codemap-precision-review/FINDINGS.md`: analysis and tuning conclusions.

## Verification Gate

The final algorithm and research baseline passed:

```text
go test ./... -count=1
go vet ./...
go build ./cmd/ddocs ./cmd/demon
```

The benchmark scripts also regenerated the pinned cross-repository reports, validated Bifrost only after rules were frozen, reproduced Space Rocks precision, and recovered the canonical Space Rocks holdout 10/10.

## Current Readiness Decision

The algorithm is accepted for early implementation testing.

The next work should implement the review workflow, persistent declines, and real repository dogfooding. The system must continue to present suggestions as reviewable evidence rather than automatic truth.

Another tuning pass should begin only after collecting materially new evidence from additional repositories or actual user accept/decline outcomes.

## Related docs

- [Codemap Suggestion Algorithm](codemap-suggestion-algorithm.md)
- [Codemap Missing-Link Evidence](codemap-evidence.md)
- [Codemap Managed Execution](architecture/codemap-managed-execution.md)
- [Codemap Pipeline](architecture/codemap-pipeline.md)
- [Codemap Benchmark Methodology](research/codemap-benchmark-methodology.md)
- [Codemap Precision Governance](research/codemap-precision-governance.md)

## Notes

This log preserves development history and rejected experiments. It must not be used as the sole authority for current product behavior or as a universal quality guarantee.
