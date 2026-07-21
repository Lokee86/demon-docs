# Demon Docs

Demon Docs is a deterministic documentation maintenance engine for repository-owned Markdown.

It maintains folder indexes, validates and repairs local links, reports orphan documents, supports explicit link-aware moves, manages configured codemap sections, projects codemap references back onto code folders, and records reviewable repairs while limiting ownership to explicit managed surfaces.

Configured repositories can also enforce frontmatter fields and document-body structure, create documents from TOML document schemas, and resolve explicit format conflicts without rewriting authored prose.

## Hackathon judge quick path

- Review the [hackathon scope, prior-work boundary, and AI engineering process](HACKATHON.md).
- Install the [latest prebuilt release](https://github.com/Lokee86/demon-docs/releases/latest), or build from source below.
- Run the [adoption walkthrough](tutorial/adoption-demo/README.md) against its disposable fixture.
- Review the [current limitations](#current-limitations) and verify the repository with `go test ./... -count=1` and `go run ./cmd/ddocs check`.

## Core behavior

Demon Docs can:

- maintain recursive folder indexes inside a configured documentation root;
- preserve authored content outside explicit managed blocks;
- validate and repair supported Markdown, wiki, reference, image, and local HTML targets;
- report managed Markdown documents with no meaningful inbound links;
- move a repository-contained file or directory and rewrite affected links without initialization;
- retain stable file identities and path history in private `.ddocs/` state, under the standalone docs root or the initialized repository root;
- reuse durable clean-validation results for unchanged frontmatter and document-body format checks;
- read changed or new link-inventory content through a bounded worker pool while preserving deterministic traversal and merge order;
- read and parse changed Markdown link sources through bounded workers before serial deterministic target resolution and repair planning;
- plan independent source rewrites for known target moves through bounded workers before deterministic merge;
- retain private `.ddocs/` objects without automatic compaction until readers and writers share a cross-process lock;
- expose ambiguous repairs and codemap candidates for decline, reconsider, or compatibility selection decisions;
- record applied normal repairs with bounded, hash-guarded undo and repair blocks;
- explicitly inspect, preview, update, and verify unified managed codemap sections;
- preserve existing codemap links by default while supporting opt-in confidence pruning;
- project codemap references onto configured code folders and files;
- export deterministic codemap datasets through bounded per-document workers with per-build target-hash deduplication;
- build codemap corpora through bounded shared source analysis while document loading and Git-history collection proceed independently;
- run codemap benchmark or precision research;
- watch relevant filesystem changes in the foreground; and
- run one optional repository-local watcher through the repository demon and feeder lifecycle.

It does not silently rewrite prose outside explicit managed regions, choose among ambiguous targets, remove codemap links by confidence unless configured to do so, or invoke codemap generation through normal watch or daemon automation.

## Documentation as a versioned graph

Most documentation tools treat Markdown as a collection of files located at paths. Demon Docs treats a repository as a versioned document graph with stable identity, history, and deterministic reconciliation.

A document can retain its identity when its path changes. Content hashes determine whether validation and inventory results remain reusable, whether evidence has materially changed, and whether a recorded repair can still be applied safely. Private Git-style objects, references, and transactions preserve repository state, path history, review decisions, and guarded undo data without requiring generated metadata in the documents themselves.

The individual techniques are familiar from version control, content-addressed storage, build systems, and databases. Their composition is the unusual part: stable document identity, content fingerprinting, repository history, managed ownership boundaries, and graph repair work together to make an ordinary Markdown repository behave like a self-maintaining document system.

## Installation

### Prebuilt release

The recommended judge and end-user path does not require Go or repository compilation.

Download the latest release from [GitHub Releases](https://github.com/Lokee86/demon-docs/releases/latest):

- `demon-docs_windows_amd64.zip` for 64-bit Windows 10 or 11;
- `demon-docs_linux_amd64.tar.gz` for 64-bit Linux; and
- `checksums.txt` to verify the downloaded archive.

Extract the archive, then either run the binaries from that directory or place them on your `PATH`.

Windows PowerShell verification:

```powershell
.\ddocs.exe --version
.\ddocs.exe --help
.\demon.exe --help
```

Linux installation and verification:

```bash
tar -xzf demon-docs_linux_amd64.tar.gz
sudo install demon-docs_linux_amd64/ddocs /usr/local/bin/ddocs
sudo install demon-docs_linux_amd64/demon /usr/local/bin/demon
ddocs --version
ddocs --help
demon --help
```

The release workflow runs the complete Go test suite on Windows and Linux, executes the Windows and Linux command-line artifacts, runs the vet gate, builds both archives with `CGO_ENABLED=0`, and publishes SHA-256 checksums. macOS does not currently have a prebuilt release asset.

### Build from source

Go 1.26.5 or newer is required to build the current source tree.

```bash
git clone https://github.com/Lokee86/demon-docs.git
cd demon-docs
go install ./cmd/ddocs
go install ./cmd/demon
```

Or build repository-local binaries:

```bash
go build -o bin/ddocs ./cmd/ddocs
go build -o bin/demon ./cmd/demon
```

`ddocs` is the canonical executable. `demon` is an alias backed by the same application implementation.

## Quick start

Run index, link, health, move, and foreground-watch operations without initializing a repository:

```bash
ddocs fix --root docs --indexes
ddocs fix --root docs --links
ddocs watch --root docs --once
ddocs mv --dry-run docs/old.md docs/new.md
ddocs mv docs/old.md docs/new.md
ddocs check --root docs --indexes --links
```

In standalone mode, the resolved docs root is also the scope boundary. The first link-enabled mutating pass creates private identity and history state beneath `docs/.ddocs/`; it does not create `.ddocs/config.toml`.

Initialize only when repository-level configuration or lifecycle features are needed:

```bash
ddocs init --root docs/
ddocs fix
ddocs fix
ddocs check
```

Initialization establishes a stable repository-wide boundary and enables repository discovery, `ddocs status`, feature toggles, starter schemas, linked-worktree bootstrap, reverse projections outside the docs root, and the detached repository demon. A second `fix` verifies idempotence before the read-only `check` gate.

Inspect configuration selection at any time with `ddocs config paths` and `ddocs config show`. `ddocs status` specifically reports an initialized repository.

See [Getting Started](docs/guides/getting-started.md) for adoption, ignore rules, subsystem selection, and recovery guidance.

## Primary commands

```text
ddocs init         optionally initialize repository-local configuration and daemon scope
ddocs status       show selected repository and documentation paths
ddocs mv           move a file or directory and rewrite affected links
ddocs new          create a document from a configured document schema
ddocs format       resolve an explicit document-body format conflict
ddocs schema       install starter document schemas
ddocs check        verify selected systems and report document-health failures
ddocs fix          apply safe deterministic reconciliation
ddocs watch        run reconciliation after relevant filesystem changes
ddocs suggestions  inspect and decide unresolved repair suggestions
ddocs changes      inspect, undo, block, or unblock applied repairs
ddocs config       inspect or initialize configuration
ddocs codemaps      manage codemap sections and run codemap research
ddocs demon        manage repository-local watcher lifecycle
```

Subsystem selectors:

```text
--docs         documentation indexes, configured frontmatter, and document-body format
--indexes      documentation indexes only
--frontmatter  configured frontmatter enforcement only
--format       document-body format enforcement only
--links        repository-local link validation, repair, and orphan checking
--reverse      code-folder reverse indexes
```

Use `ddocs <command> --help` or `ddocs <command> <subcommand> --help` for exact scoped flags and safety behavior. `demon --help` opens the repository-demon command family. See the [CLI Reference](docs/reference/cli.md) for command ownership and mutation scope.

## Incremental and private-state behavior

Unchanged clean frontmatter and document-format results can be reused from durable `.ddocs/` cache records. Content, policy, schema, immutable-state, duplicate-identity, or validation-engine changes invalidate reuse automatically. A standalone read-only check does not initialize `.ddocs/` merely to save cache data.

Link inventory traverses the repository deterministically, reuses unchanged size/mtime metadata, and reads changed or new files through a bounded 16-worker pool. Changed Markdown link sources are also read and parsed through bounded workers; results remain indexed by source path and merge serially before target resolution, identity updates, diagnostics, review policy, and repair planning. For known target moves, each unchanged affected source independently prepares its rewrite plan through the same bounded worker pool, then results merge in source-path order before graph and diagnostic publication. When an index, frontmatter, format, or reverse-index fix changes Markdown after the initial link pass, Demon Docs refreshes only those changed link sources. A clean non-link fix does not run a repository-wide link scan or initialize absent link state. Explicit `--links` still runs the complete reconciliation, review, rollback, and suppression path.

Automatic private-object compaction is currently disabled. The repository demon and CLI run as separate processes, and go-git pack replacement is not safe until private-state readers and writers share a cross-process lock. Normal commands therefore retain loose objects rather than risking a missing pack or referenced object.

Cold frontmatter and document-format validation, link-inventory reads, and changed Markdown source reads and parsing now use bounded worker pools. A changed Markdown source is still reparsed as a whole, and body-only generated rewrites can still invalidate validation cache entries. These remaining performance boundaries are tracked in [Current Product Limitations](docs/limits/current-limitations.md) and the [Roadmap](docs/planning/roadmap.md).

## Performance maturity

The current implementation is a correctness-first hackathon prototype. It is serviceable on modest repositories, but it is not yet optimized for low-latency operation on large or high-churn trees.

Watcher debounce only delays admission of a reconciliation pass. A burst of filesystem events resets that quiet period repeatedly, directory moves may produce many events, and the admitted callback can still perform broad repository scanning and several selected reconciliation stages serially. As a result, a configured debounce measured in fractions of a second can still produce visible repair latency measured in seconds.

The intended production direction is path-aware dirty tracking, feature-specific incremental reconciliation, fewer repeated state reads and writes, and benchmark-guided scheduling. Until then, `ddocs check`, `ddocs fix`, and explicit `ddocs mv` remain the authoritative operational surfaces; the watcher and repository demon are convenience automation rather than performance guarantees.

## Safety model

Demon Docs owns only explicit deterministic surfaces:

- content between managed index markers;
- configured parent-index navigation lines;
- the path portion of a recognized local link when one destination is deterministic;
- explicitly requested repository-contained moves;
- configured generated reverse-index regions;
- the complete body of an adopted configured codemap section; and
- private identity, review, and runtime state under `.ddocs/`.

Labels, titles, aliases, queries, fragments, surrounding prose, source newline style, and final-newline state are preserved during supported link rewrites.

Ambiguous targets remain unchanged and are reported for user selection. Undo refuses to overwrite files changed after the recorded repair.

## Automation

Foreground automation:

```bash
ddocs watch
```

Repository-local detached ownership:

```bash
demon run
demon --status
demon --logs
```

Foreground `ddocs watch` works in standalone or initialized mode. The detached repository demon requires an initialized repository because its configuration, ownership, feeders, and logs are repository-local. Both are convenience layers: `ddocs check` remains the authoritative normal reconciliation CI and recovery surface. Codemap-generation convergence requires the separate read-only `ddocs codemaps check --root ...` command.

See [CI and Automation](docs/guides/ci-and-automation.md) and [Repository Demon](docs/operations/repository-demon.md).

## Documentation

- [Documentation index](docs/INDEX.md)
- [Documentation policy](docs/documentation-policy.md)
- [Agent guidance](docs/agent/INDEX.md)
- [Guides](docs/guides/INDEX.md)
- [Reference](docs/reference/INDEX.md)
- [Architecture](docs/architecture/INDEX.md)
- [Operations](docs/operations/INDEX.md)
- [Current limitations](docs/limits/INDEX.md)
- [Research](docs/research/INDEX.md)
- [Planning](docs/planning/INDEX.md)
- [Development](docs/development/INDEX.md)

Current behavior, future work, and benchmark evidence are intentionally separated. The [Roadmap](docs/planning/roadmap.md) summarizes sequencing but is not the canonical reference for shipped behavior.

## Managed Codemaps

Demon Docs includes an explicit foreground codemap workflow:

```bash
ddocs codemaps inspect --root docs/architecture/example.md
ddocs codemaps fix --root docs/architecture/example.md --dry-run
ddocs codemaps fix --root docs/architecture/example.md
ddocs codemaps check --root docs/architecture/example.md
```

The command adopts the complete configured section as one managed artifact, preserves existing valid links by default, automatically adds selected non-declined `hard_link` and `context` recommendations, and uses content-addressed transactional writes. Persisted declines suppress unchanged future additions. Confidence pruning is separately configurable and disabled by default.

Existing configured sections are supported. When the selected effective document schema requires a codemap section, the public command creates it at the schema-defined deterministic position; documents without that schema authority remain unchanged.

Codemap generation never runs through generic `fix`, generic `check`, watch, or the repository demon.

See:

- [Managing Codemaps](docs/guides/managing-codemaps.md) for the operational workflow;
- [Codemap Managed Execution](docs/architecture/codemap-managed-execution.md) for ownership, planning, rendering, transactions, and failure behavior;
- [Codemap Missing-Link Algorithm](docs/codemap-suggestion-algorithm.md) for ranking and measured readiness;
- [Codemap Algorithm Development Log](docs/codemap-algorithm-development-log.md) for benchmark and tuning history; and
- [Codemap Missing-Link Evidence](docs/codemap-evidence.md) for the evidence boundary.

## Built with GPT-5.6 and Codex

Demon Docs was built through an AI-native engineering workflow designed to use GPT-5.6 and Codex as an implementation team while retaining explicit human control over product direction, architecture, risk, and scope.

None of the current Go implementation was written by hand. Implementation, testing, debugging, documentation, repository operations, and many architectural drafts were produced through ChatGPT-5.6, Codex, and delegated agent workflows.

This was not a single-prompt generation process. The project was developed through many small implementation streams, repeated reviews, test failures, rejected approaches, benchmark-driven corrections, and deliberate consolidation.

### How GPT-5.6 was used

GPT-5.6 was used primarily through ChatGPT with a customized Model Context Protocol server connected directly to the repository.

That environment allowed GPT-5.6 to inspect and edit files, manage branches and Git worktrees, run bounded verification tasks, reconcile parallel implementation streams, and maintain context across the project. It acted as the primary engineering interface for:

- turning product requirements into concrete architecture;
- dividing large features into bounded implementation steps;
- implementing and reviewing repository changes;
- identifying missing ownership boundaries and architectural seams;
- coordinating parallel work;
- diagnosing failed tests and integration problems;
- reviewing documentation against the implemented product surface;
- maintaining repository and worktree hygiene; and
- preparing the project for release and submission.

GPT-5.6 also entered the project with accumulated context from working with me on Space Rocks, a substantially larger software project. That context included established preferences for deterministic behaviour, explicit ownership, conservative mutation, early architectural seams, isolated Git worktrees, small implementation tasks, direct verification, and comprehensive documentation.

This meant the development relationship did not begin from a blank prompt. GPT-5.6 already understood many of the engineering standards, workflow constraints, and failure patterns I expected it to account for.

### How Codex was used

Codex provided additional implementation capacity, particularly when work could be divided into isolated or parallel streams.

Codex sessions were used for bounded implementation, debugging, testing, documentation, and corrective passes. Hermes was also used to coordinate Codex agents and sub-agents during portions of development where parallel execution was useful. Hermes was deployed with 5.6-luna.

Work was generally divided by ownership boundary or feature surface, completed in isolated branches or worktrees where practical, then reviewed, tested, merged, and corrected through the primary GPT-5.6 workflow.

Codex was not treated as an autonomous product owner. Its work operated within requirements, constraints, and architectural decisions already established for the project.

### Human ownership

I acted as the product designer and lead engineer.

I identified the original problem, defined the product, selected and rejected features, established behavioural and safety constraints, reviewed plans and implementation results, prioritized work, resolved conflicting approaches, and decided what needed to be cut or postponed. Innovative approaches, such as the novel codemap extraction algorithm, and the use of Git-style tracking, were also human-based and driven, though developed and deployed by AI.

Important human decisions included:

- rebuilding the project in Go rather than extending the original Python prototype;
- making deterministic behaviour a core product constraint;
- limiting automatic mutation to explicit managed surfaces;
- preserving authored prose outside those surfaces;
- separating safe daemon operations from riskier explicit commands;
- retaining ambiguous repairs for human review rather than guessing;
- building review, decline, block, and guarded-undo workflows;
- using a private Git-style object and transaction system for repository state;
- treating codemap generation as experimental rather than presenting early results as production guarantees; and
- deferring RepoGraph and agent-context injection when their scope threatened the submission deadline.

AI wrote the code and much of the architecture, but it did so inside a managed engineering process. Plans were challenged, abstractions were rejected, tests changed implementation direction, benchmark failures changed evidence rules, and features were cut when they could not be completed responsibly.

The intended model was not “prompt once and accept the result.” It was to use AI as an engineering team under active technical direction.

### Development evidence

The repository includes the [raw hackathon development logs](.codex-hackathon/sessions/).

These JSONL files are preserved as unmodified session records rather than edited excerpts. They show implementation prompts, tool activity, agent responses, failures, corrections, and work distributed across multiple sessions.

They should not be interpreted as a complete transcript of the project. Much of the primary development happened through GPT-5.6 in ChatGPT using the repository MCP server, while the included files primarily preserve the Codex-facing portion of the workflow. Together with the Git history, they provide a direct record of how the project was built rather than only a retrospective description.

The repository history provides the other major evidence boundary. It preserves both the earlier Python prototype and the subsequent Go rebuild, allowing the project’s pre-hackathon state and hackathon development to be distinguished directly.

### Prior work and hackathon scope

Before the hackathon, the project existed as a small Python utility called **Doc Ledger**. It generated documentation indexes and included an early watcher-daemon concept, but remained a narrow, backburnered tool.

During the hackathon, it was rebuilt in Go and renamed Demon Docs.

The rebuild added repository-scoped identity and state, link reconciliation after ordinary filesystem moves, link-aware file operations, review and undo history, document and frontmatter schemas, reverse indexes, orphan health checks, the upgraded daemon lifecycle, experimental codemap suggestions, broader testing, and the current documentation system.

The original ideas of automated index maintenance and a watcher daemon predate the hackathon. The present architecture and nearly the entire current product surface were developed during the submission period.

## Development

Run the complete local release gate:

```bash
make release-check
```

Run the Go suite directly:

```bash
go test ./... -count=1
```

See [Testing and Fixtures](docs/development/testing-and-fixtures.md) and [Repository Layout](docs/development/repository-layout.md).

## Project status

Repository indexing, frontmatter enforcement, document-body format enforcement, schema-based creation, local-link reconciliation, orphan health checks, stateless moves, reverse indexes, suggestion decisions, applied-change history, watcher/demon lifecycle, schema-aware codemap execution with schema-driven missing-section placement, and codemap research tooling are implemented. Broader diagnostics, polyglot code intelligence, and deterministic agent context remain incomplete or planned.

See [Roadmap](docs/planning/roadmap.md) for current status and sequencing.

## License

See [LICENSE](LICENSE).
