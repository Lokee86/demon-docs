# Demon Docs Documentation

The repo root [README.md](../README.md) is the starting point for Demon Docs.
This `docs/` folder holds deeper design, operational, and maintenance references for the tool.

## References

- [Roadmap](roadmap.md): Planned implementation order from the deterministic repository graph through self-managed watcher operations and optional LLM assistance.
- [Deterministic Typed Repository Graph](repository-graph.md): Focused design for the language-neutral deterministic repository graph.
- [Code-Folder Reverse Indexes](reverse-indexes.md): Focused design for documentation coverage projected onto code targets.
- [Code, Dependency, and Entanglement Facts](code-dependency-and-entanglement.md): Focused design for deterministic code facts and bounded entanglement projections.
- [Deterministic Agent Context and Integrations](agent-context-and-integrations.md): Focused design for graph-based agent context retrieval and thin integrations.
- [Context-Injection Benchmarking](context-injection-benchmarking.md): Future research plan for a four-quadrant OSS corpus, paired no-context controls, and an intentionally constructed harness control.
- [Configuration](configuration.md): Config file shape, defaults, and supported overrides.
- [Repository Demon](repository-demon.md): Self-managed watcher ownership, shell and agent feeders, worktrees, shutdown, recovery, and logs.
- [Markdown Link Reconciliation](markdown-links.md): Repository-scoped local link validation, persistent identity state, and path repair.
- [Code-Symbol References](code-symbol-references.md): Focused Phase 4 design for deterministic declaration-level documentation references and language adapters.
- [Reconciliation Model](reconciliation-model.md): How Demon Docs scans, plans, and applies index updates.
- [Watcher and Automation](watcher-and-automation.md): Watch mode behavior, timestamps, PID output, and automation guidance.
- [Testing and Fixtures](testing-and-fixtures.md): Test layout, fixture strategy, and regression coverage for Demon Docs.
- [Dummy Docs Fixture Generator](make-dummy-docs.sh): Manual fixture and stress generator for recursive docs-tree testing.

## Direct Files

<!-- doc-ledger:files:start -->

- [agent-context-and-integrations.md](agent-context-and-integrations.md) - Agent Context And Integrations documentation.
- [code-dependency-and-entanglement.md](code-dependency-and-entanglement.md) - Code Dependency And Entanglement documentation.
- [code-symbol-references.md](code-symbol-references.md) - Code Symbol References documentation.
- [codemap-evidence.md](codemap-evidence.md) - Codemap Evidence documentation.
- [configuration.md](configuration.md) - Configuration documentation.
- [context-injection-benchmarking.md](context-injection-benchmarking.md) - Future context-injection benchmarking research plan.
- [markdown-links.md](markdown-links.md) - Markdown Links documentation.
- [reconciliation-model.md](reconciliation-model.md) - Reconciliation Model documentation.
- [repository-demon.md](repository-demon.md) - Repository Demon documentation.
- [repository-graph.md](repository-graph.md) - Repository Graph documentation.
- [reverse-indexes.md](reverse-indexes.md) - Reverse Indexes documentation.
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