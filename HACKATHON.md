# Demon Docs

**The problem:** Your Markdown links are broken and stale. Again. Because someone moved the files outside Obsidian—or because VS Code predictably failed to update them again.

This is the problem that birthed Demon Docs. It began as a simple index generator with limited link maintenance, then grew into full Markdown link reconciliation and, from there, lightweight document management.

Demon Docs brings the link and navigation maintenance people expect from tools like Obsidian to ordinary Markdown folders—without requiring every file operation to happen inside a particular editor or through tool-specific commands.

It is designed to work alongside common editors and publishing tools, maintaining the files beneath them rather than replacing them.

## What makes Demon Docs different

While Markdown renderers, static-site generators, editors, and publishing tools are everywhere, reliable link management and maintenance remain surprisingly rare. The few applications and utilities that provide it generally require file operations to happen through their own interfaces or specific commands.

Demon Docs operates at the filesystem level through its own Git-style repository and transaction system, with an optional self-managed daemon. It can reconcile links during explicit Demon Docs moves—an active daemon can also reconcile during live operations—or after files have already been moved or renamed through Explorer, Git, scripts, agents, editors, or ordinary filesystem operations; this requires an initialized Demon Docs repository.

At a time when developers are increasingly reliant on AI tools for execution and automation, Demon Docs is a rare new developer utility built around strict determinism, with strong human-facing use cases. Its core operations do not depend on, or drive an LLM, external service, or probabilistic output.

The same repository system maintains folder indexes, enforces document and frontmatter schemas, detects documentation-health problems, generates reverse indexes from authored code relationships, and records review and recovery history. Authored content is preserved, changes remain inspectable, and repairs can be reviewed, declined, blocked, or undone.

## What it does

Demon Docs currently provides:

- managed folder indexes that preserve authored prose;
- link tracking and repair after ordinary filesystem moves;
- document and frontmatter schemas through the CLI or Git hooks;
- a Git-backed review ledger with declines, blocks, and guarded undo;
- reverse indexes derived from authored documentation codemaps;
- experimental missing-link suggestions ranked from deterministic evidence; and
- a self-managed daemon that maintains links and indexes automatically.

The daemon operations are intentionally limited while riskier and experimental operations remain explicit. The static `ddocs` CLI and `demon` alias remain authoritative.

## AI-native engineering

Demon Docs was built through a workflow designed to get the most out of GPT-5.6 and Codex while limiting the problems agents typically run into on larger projects.

None of the implementation code was written by hand.

Coding work was performed either by GPT-5.6 through ChatGPT and a customized MCP server, or through Codex and Hermes agents and sub-agents, depending on the stage of development and whether parallel work was useful.

This ChatGPT account is used almost entirely for software development. Project decisions, architectural rules, workflow preferences, previous failures, and unresolved work persist across separate conversations rather than being rebuilt from scratch every time.

As a result, GPT-5.6 also entered Demon Docs with substantial context from working with me on a much larger previous software project. It already understood how I divide implementation work, which architectural shortcuts I reject, how I expect changes to be tested and verified, and where agent-built systems tend to lose coherence. Demon Docs benefited from that accumulated working history from the beginning.

I acted as the product designer and lead engineer. I defined the problems, made product decisions, set behavioural and safety constraints, reviewed plans, prioritized implementation order, rejected redundant suggestions, orchestrated parallel work, and cut or deferred features that threatened the deadline.

The project was divided into small implementation streams, isolated in worktrees where practical, then tested, reviewed, merged, and corrected. Weak abstractions were rejected, benchmark failures changed the evidence rules, and larger features were postponed when necessary.

AI wrote the code and much of the architecture, but within a managed engineering process. The goal was to use AI as an engineering team, not an unreviewed code generator.

## Prior work and hackathon scope

Before the hackathon, I had a small Python project called **Doc Ledger** that generated documentation indexes and included an early daemon concept. It solved a real problem in my own repository but remained a narrow, backburnered utility.

After the hackathon began, the project was entirely rebuilt in Go and renamed Demon Docs. Before that rebuild, the project contained only basic Markdown index maintenance and an early watcher daemon that could refresh those indexes. Link identity and repair, repository-scoped state, link-aware moves, review and undo history, reverse indexes, schemas, the upgraded daemon, the codemap generation experiment, and the present test and documentation system were all added during the event.

The repository history preserves the original Python prototype and the full Go rebuild, making the pre-hackathon work easy to distinguish from the work completed during the submission period.

Primary Codex feedback ID: `019f7928-4500-70e0-9bf9-6b20ad53c6a7`.

## Codemap generation

My personal documentation procedure requires the presence of "codemaps" inside every implementation document. These are simply a list of code files that the document is relevant to. These codemaps are a necessary deterministic component of the reverse-indexing feature.

Since this has been evaluated as an uncommon convention, Demon Docs also introduces an experimental codemap generation algorithm that has been evaluated through known-link recovery and automated-AI review of new suggestions.

In the current cross-repository review, 121 suggestions from five external repositories were examined. Eighty-three were judged valid missing relationships, thirty-four were plausible but unnecessary, and four were incorrect. A narrow tuning pass removed the observed incorrect target patterns without removing reviewed valid or plausible suggestions from that sample.

A separate benchmark recovered 11 of 18 hidden authored targets across the external corpus. Testing against Space Rocks recovered all ten relationships in its canonical holdout set.

These are small and uneven datasets, not production guarantees. The useful result is a measurable way to improve the algorithm without silently turning suggestions into repository truth.

Simply put, Demon Docs can attempt to deterministically generate codemaps from existing documentation based on a variety of criteria. This feature remains experimental and is excluded from the primary demonstration.

## Where it goes next

Reverse indexes already provide a deterministic route from source code to the authored documentation that describes it. A broader polyglot repository graph and bounded agent-context bundler are planned, but they are not part of this submission.

Installation, testing instructions, benchmark methodology, and current limitations are documented in the repository README and supporting documentation.
