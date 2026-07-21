---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7e4a-a048-7d24a81a2014
document_type: general
policy_exempt: false
summary: Historical context-injection benchmark design transferred from Demon Docs to the planned Grimoire Context tool.
---
# Context-Injection Benchmarking

Parent index: [Research](./INDEX.md)

## Purpose

This document preserves the historical-task research design for evaluating deterministic context injection against no-context controls without benchmark leakage. Ownership has transferred to the planned **Grimoire Context** sibling tool.

## Overview

Grimoire Context will need evidence that deterministic context bundles improve real agent work. Context delivery is outside Demon Docs, but historical corpus preparation artifacts, three pinned task fixtures, validation manifests, and fixture-preparation tools remain here as transferred research provenance. Paid or large-scale paired model trials remain deferred.

## Research status

Transferred research design with retained discovery artifacts and validated historical-task fixtures. It does not describe shipped or planned Demon Docs behavior.

## Benchmark framing

The benchmark asks whether bounded deterministic repository context improves task performance, not whether an agent can solve a task after seeing its historical solution. Authentic pre-change snapshots, independent oracles, paired conditions, and leakage controls are therefore central to the design.

## Research Question

The primary question is:

> Does bounded Grimoire Context output reduce repository discovery work and improve implementation correctness without adding harmful noise?

The benchmark must distinguish context quality from model quality, repository familiarity, task difficulty, and documentation quality.

## Repository Matrix

Repository selection should cover a two-axis matrix rather than ranking projects by stars or documentation volume alone.

| Code quality | Documentation quality | Intended test |
|---|---|---|
| Good | Good | Detect unnecessary or noisy context when the repository already communicates itself well. |
| Good | Poor | Test whether Grimoire Context reconstructs missing architectural and ownership context. |
| Poor | Good | Test whether accurate documentation helps an agent navigate tangled or inconsistent implementation. |
| Poor | Poor | Stress-test discovery, prioritization, ambiguity handling, and incomplete repository knowledge. |

A fifth repository or task set should act as a **harness control fixture**. It may be a small intentionally constructed repository where the expected graph, selected context, and implementation surface are exactly known. This validates workspace reset, graph construction, context delivery, leakage prevention, logging, and scoring; it does not replace authentic OSS repositories.

The harness control is distinct from the **experimental control condition**. For every authentic task, the experimental control is the same repository snapshot, task, model, prompt, tools, and limits with no injected Grimoire Context output. The treatment differs only by receiving the deterministic context bundle.

Stars are metadata only. They may help identify repositories with enough issue and pull-request history, but they are not evidence of code quality and must not affect quadrant assignment.

## Independent Classification

Code quality and documentation quality must be assessed independently.

### Code-quality evidence

Possible evidence includes:

- package or module cohesion;
- dependency direction and coupling;
- repeated or mirrored implementation surfaces;
- generated-code ownership hazards;
- test reliability and build friction;
- large files or functions;
- inconsistent conventions;
- stale or dead paths;
- cross-package bug history; and
- whether ordinary changes routinely require unrelated edits.

No single metric establishes that a repository is good or bad. Classification should be recorded as a reviewable rationale with supporting observations.

### Documentation-quality evidence

Possible evidence includes:

- architecture and subsystem maps;
- ownership boundaries;
- invariants and failure behavior;
- setup and verification instructions;
- terminology consistency;
- generated-source rules;
- cross-package workflows;
- freshness relative to the selected historical commit; and
- whether a developer can locate the correct change surface from the documentation.

Documentation volume is not documentation quality. A large documentation site may still omit implementation ownership, while a short design document may be sufficient.

## Authentic Historical Tasks

The preferred tasks come from merged OSS issues and pull requests:

1. Pin the repository to the parent commit immediately before an accepted change.
2. Preserve the original issue text as the agent task.
3. Confirm the baseline builds and its relevant tests pass at that commit.
4. Keep the accepted patch hidden from the agent.
5. Use the accepted patch, tests, and review discussion as evaluation evidence.

A useful task should require repository discovery rather than isolated syntax work. The corpus should eventually include localized bugs, cross-package changes, configuration or persistence changes, and ownership/refactor changes.

The accepted upstream implementation is evidence, not necessarily the only valid solution. Evaluation should prioritize behavior, correct ownership, verification, and unnecessary changes before textual patch similarity.

## Experimental Conditions

At minimum, each task should be attempted under two matched conditions:

1. **Control:** repository and task only, with no injected Grimoire Context bundle.
2. **Treatment:** the same repository snapshot, task, model, prompt, limits, and tools, with the deterministic Grimoire Context bundle injected.

A later third condition may compare Grimoire Context with ordinary repository search or a generic retrieval system:

3. **Naive retrieval:** repository search or generic RAG without deterministic ArcanaGraph-backed context assembly.

Runs must not share conversational memory or prior task discoveries. Model, version, temperature or equivalent sampling controls, prompt, token budget, tool access, timeout, and verification commands should remain fixed within a comparison.

## Measurements

Useful measurements include:

- task completion and test results;
- correct files and ownership boundaries selected;
- missed dependent or mirrored surfaces;
- unnecessary files changed;
- repository searches, files opened, and tool calls;
- prompt, context, and completion tokens;
- elapsed execution time where reliably available;
- context bytes or tokens delivered;
- ambiguity and truncation diagnostics;
- factual errors caused by stale or irrelevant context; and
- qualitative review against the accepted change.

Cost should be recorded but not treated as the sole success metric. A smaller context that causes a wrong edit is worse than a larger context that reliably identifies the required change surface.

## Preventing Benchmark Leakage

The context bundle must be generated from the pinned pre-change repository only. It must not contain:

- the accepted patch;
- post-change documentation;
- issue comments added after the solution was known, unless deliberately included as part of the task;
- oracle metadata;
- paths or summaries manually tailored to reveal the expected solution; or
- artifacts produced by an earlier agent run.

Task text, public metadata, context input, agent workspace, and evaluator-only oracle data should remain separate.

## Cost-Conscious Staging

The research can advance without funding full model trials.

### Stage 1: Design and corpus preparation

- define the repository matrix and classification rubric;
- preserve discovery scripts and candidate reports;
- identify historical tasks and pin base commits;
- validate baseline builds and tests;
- define task manifests and evaluator-only oracle manifests; and
- test deterministic context generation without asking an agent to implement anything.

### Stage 2: Harness dry runs

- use the intentionally constructed control repository;
- verify workspace reset, context delivery, logging, and scoring;
- use local or low-cost models only to catch harness defects; and
- do not draw product conclusions from these runs.

### Stage 3: Small paired pilot

- select one task from each repository quadrant;
- run one control and one treatment attempt per task;
- inspect whether the metrics are meaningful; and
- revise the protocol before spending on repeated trials.

### Stage 4: Repeated benchmark

When affordable, run multiple independent repetitions per condition, randomize condition order, retain complete run artifacts, and report failures and inconclusive results rather than selecting only successful examples.

## Current Research Artifacts

GitHub discovery work is preserved under `research/context-benchmarking/`. It includes current star-neutral corpus-preparation helpers, the original scanner retained for provenance, candidate reports, and three historical `wifitui` task manifests with baseline validation and evaluator-only oracle data.

The prepared fixtures are `wifitui` pull requests 163, 167, and 178. Each fixture separates the agent-visible task and metadata from evaluator-only oracle data and records baseline validation. `validate_fixture.py` and the committed validation summary check the retained fixture contract. These fixtures establish harness inputs; they do not yet constitute a paired context experiment.

The first scan was biased toward active, moderately sized Go projects and included star bands as a candidate-pool heuristic. Its rankings must not be treated as the final benchmark design. In particular:

- `wifitui` appears useful as a possible good-code/poor-docs candidate;
- `provider-sql` may be useful as a harder repository with mirrored and generated ownership surfaces;
- neither has been formally classified under the two-axis rubric; and
- the remaining quadrants still require deliberate candidate discovery and review.

The scanner should be revised before another broad search so code quality, documentation quality, task quality, and operational feasibility are reported separately.

## Open Decisions

- Final code-quality and documentation-quality rubrics.
- The control repository and exact harness contract.
- Required number of tasks and repetitions per quadrant.
- Which models and hosts to compare.
- Whether human review is blinded to the experimental condition.
- How to score valid implementations that differ from the accepted patch.
- Which tool-call and token metrics can be collected consistently across hosts.
- How context injection is delivered at the system or host-integration layer without leaking evaluator data.

## Non-Goals

- Proving that one model is generally better than another.
- Treating GitHub stars as a quality score.
- Calling an under-documented repository bad code without independent evidence.
- Manufacturing all benchmark tasks in synthetic fixtures.
- Requiring expensive benchmark runs before the context feature can be designed.
- Using benchmark-specific summaries as production context rules.

## Code map

- `research/context-benchmarking/tools/discover_candidates.py` — star-neutral repository and task candidate discovery.
- `research/context-benchmarking/tools/prepare_historical_fixture.py` — pinned historical-task fixture construction.
- `research/context-benchmarking/tools/validate_fixture.py` — fixture separation and integrity checks.
- `research/context-benchmarking/tools/initial-scan/` — original discovery scripts retained for provenance.
- `research/context-benchmarking/discovery-results/` — candidate inventories, findings, and recommendations.
- `research/context-benchmarking/fixtures/` — agent-visible tasks, metadata, baseline validation, and evaluator-only oracle records.

## Limitations

The available fixture set is intentionally small and cost-conscious. Results must be stratified by task and repository characteristics, and no-context controls must use the same model, prompt, tools, and stopping rules.

## Related docs

- [Research](INDEX.md)
- [Planned Agent Context and Integrations](../planning/agent-context-and-integrations.md)
- [Roadmap](../planning/roadmap.md)
- [Testing and Fixtures](../development/testing-and-fixtures.md)

## Notes

Context usefulness and permanent codemap-link validity are separate questions and require separate scoring and evaluation paths.
