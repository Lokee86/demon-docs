# Context Benchmarking Research

This directory preserves exploratory work for the future Demon Docs context-injection benchmark described in [Context-Injection Benchmarking](../../docs/context-injection-benchmarking.md).

This is research material, not a current product subsystem, release requirement, or commitment to fund model trials now. The immediate purpose is to develop a credible experimental design, retain reproducible discovery evidence, and avoid losing useful historical-task work.

## Benchmark Shape

The authentic OSS corpus should eventually cover four independently reviewed repository quadrants:

- good code / good documentation;
- good code / poor documentation;
- poor code / good documentation; and
- poor code / poor documentation.

A deliberately constructed repository should separately act as a harness control with a completely known graph, expected context, and implementation surface.

The synthetic repository is not the experimental control condition. The experimental control is an agent run against the same authentic repository snapshot and task without Demon Docs context. The treatment run uses the same model, task, tools, limits, and snapshot with Demon Docs context injected.

Stars are metadata only. They may help locate projects with sufficient issue and pull-request history, but they must not influence code-quality or documentation-quality classification.

## Contents

- `tools/` contains current corpus-preparation helpers and the preserved initial scanner.
- `discovery-results/` contains retained reports and machine-readable evidence from the initial GitHub search.
- `fixtures/` contains historical task text, public metadata, evaluator-only oracle data, and recorded baseline validation.

Large shallow clones, temporary run directories, dependency caches, generated agent workspaces, and fixture `source/` directories are intentionally not source-controlled. Historical source snapshots are reproducible from the repository and base commit recorded in each task manifest.

## Current Status

The initial search focused on active Go repositories with usable tests and historical pull-request data. It found plausible tasks and produced three validated `wifitui` fixtures, but it did not implement the final benchmark design.

Known limitations of that work:

- it primarily searched for under-documented repositories;
- star bands affected candidate-pool construction;
- its combined score did not independently classify code and documentation quality;
- nested documentation sites could be missed; and
- only a small number of candidates received manual validation.

The original scanner is retained under `tools/initial-scan/` for provenance and reproducibility, not as the recommended future candidate-selection method. The star-neutral helper scripts in `tools/` are the starting point for later corpus development.

No retained repository is formally assigned to a quadrant yet. `wifitui` remains a plausible good-code/poor-documentation candidate, and `provider-sql` remains a possible harder ownership/generated-surface candidate, but both require explicit rubric-based review.

## Affordable Work That Can Continue

Useful preparation does not require paid agent trials:

- refine separate code-quality and documentation-quality rubrics;
- identify and review candidate repositories for each quadrant;
- preserve authentic issue/PR tasks at pinned pre-change commits;
- validate baseline builds and tests;
- define public task manifests and evaluator-only oracle manifests;
- build the deterministic context bundles; and
- dry-run workspace reset, leakage prevention, logging, and scoring against the synthetic control fixture.

Paired model runs, repeated trials, and statistical conclusions can remain deferred until they are affordable.
