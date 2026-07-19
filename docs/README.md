# Demon Docs Documentation

The repo root [README.md](../README.md) is the starting point for Demon Docs.
This `docs/` folder holds deeper design, operational, research, and maintenance references for the tool.

## References

- [Roadmap](roadmap.md): Current shipped behavior, active tuning work, near-term work, and the back-burnered polyglot code-graph track.
- [Codemap Missing-Link Evidence](codemap-evidence.md): Implemented codemap export, evidence, holdout benchmarking, precision evaluation, current metrics, and decision-safety rules.
- [Suggestions, Repairs, and Change History](review-ledger.md): Suggestion decisions, Git-backed applied-change history, undo granularity, and repair blocks.
- [Markdown Link Reconciliation](markdown-links.md): Repository-scoped link validation, persistent identity state, supported syntax, and deterministic path repair.
- [Stateless Document Refactoring](document-refactoring.md): Explicit file and directory moves with dry-run planning and affected-link rewrites, without requiring initialization.
- [Repository Demon](repository-demon.md): Self-managed watcher ownership, shell and agent feeders, worktrees, shutdown, recovery, and logs.
- [Configuration](configuration.md): Config selection, repository settings, defaults, and supported CLI overrides.
- [Reconciliation Model](reconciliation-model.md): How Demon Docs scans, plans, applies, and verifies deterministic index and link updates.
- [Watcher and Automation](watcher-and-automation.md): Foreground watch behavior and its relationship to the repository demon.
- [Testing and Fixtures](testing-and-fixtures.md): Release gates, fixture strategy, link/codemap regression coverage, and research benchmarks.
- [Markdown Link Performance](link-performance.md): High-fanout move benchmarks, real Space Rocks move timings, and repeated 3,717-link mass-rename measurements.
- [Code-Folder Reverse Indexes](reverse-indexes.md): Implemented file/folder reverse documentation projections and the remaining symbol, repair, and reporting boundaries.
- [Deterministic Typed Repository Graph](repository-graph.md): Back-burnered architecture for joining the existing documentation/link graph with normalized polyglot code facts.
- [Code-Symbol References](code-symbol-references.md): Planned declaration-level references behind the same polyglot adapter seam as the future code graph.
- [Code, Dependency, and Entanglement Facts](code-dependency-and-entanglement.md): Planned polyglot code-graph provider contract and bounded graph projections.
- [Deterministic Agent Context and Integrations](agent-context-and-integrations.md): Planned graph-based context retrieval and thin host integrations.
- [Context-Injection Benchmarking](context-injection-benchmarking.md): Research corpus, paired no-context controls, leakage rules, and historical-task fixtures.
- [Dummy Docs Fixture Generator](make-dummy-docs.sh): Manual fixture and stress generator for recursive docs-tree testing.

## Direct Files

<!-- doc-ledger:files:start -->

- [agent-context-and-integrations.md](agent-context-and-integrations.md) - Agent Context And Integrations documentation.
- [code-dependency-and-entanglement.md](code-dependency-and-entanglement.md) - Code Dependency And Entanglement documentation.
- [code-symbol-references.md](code-symbol-references.md) - Code Symbol References documentation.
- [codemap-evidence.md](codemap-evidence.md) - Codemap Evidence documentation.
- [configuration.md](configuration.md) - Configuration documentation.
- [context-injection-benchmarking.md](context-injection-benchmarking.md) - Future context-injection benchmarking research plan.
- [document-refactoring.md](document-refactoring.md) - Stateless file and directory moves with affected-link repair.
- [link-performance.md](link-performance.md) - Recorded link reconciliation, move, and mass-rename performance.
- [markdown-links.md](markdown-links.md) - Markdown Links documentation.
- [reconciliation-model.md](reconciliation-model.md) - Reconciliation Model documentation.
- [repository-demon.md](repository-demon.md) - Repository Demon documentation.
- [repository-graph.md](repository-graph.md) - Repository Graph documentation.
- [reverse-indexes.md](reverse-indexes.md) - Reverse Indexes documentation.
- [review-ledger.md](review-ledger.md) - Suggestions, repairs, applied-change history, undo, and repair blocks.
- [roadmap.md](roadmap.md) - Roadmap documentation.
- [testing-and-fixtures.md](testing-and-fixtures.md) - Testing And Fixtures documentation.
- [watcher-and-automation.md](watcher-and-automation.md) - Watcher And Automation documentation.
<!-- doc-ledger:files:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Notes

- `docs/README.md` is the default docs-tree index file.
- Implemented architecture documents include `## Code map` sections so Demon Docs can export and hold out its own authored relationships as a development corpus.
- Self-authored Demon Docs links are useful for extraction and regression testing but are not an independent precision benchmark.
