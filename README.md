# Archivist

[![CI](https://github.com/Lokee86/demon-docs/actions/workflows/ci.yml/badge.svg)](https://github.com/Lokee86/demon-docs/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Lokee86/demon-docs)](https://github.com/Lokee86/demon-docs/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/Lokee86/demon-docs)](go.mod)
[![License](https://img.shields.io/badge/license-PolyForm%20Shield%201.0.0-purple.svg)](LICENSE.md)

Archivist is a deterministic documentation maintenance engine for repository-owned Markdown.

[**Watch the Archivist demo on YouTube**](https://www.youtube.com/watch?v=bvfZl25QhnY)

[![Archivist demonstration](https://img.youtube.com/vi/bvfZl25QhnY/maxresdefault.jpg)](https://www.youtube.com/watch?v=bvfZl25QhnY)

It maintains folder indexes, validates and repairs local links, reports orphan documents, supports explicit link-aware moves, manages configured codemap sections, projects codemap references back onto code folders, and records reviewable repairs while limiting ownership to explicit managed surfaces.

Configured repositories can also enforce frontmatter fields and document-body structure, create documents from TOML document schemas, and resolve explicit format conflicts without rewriting authored prose.

## Warlock Toolchain

Archivist is the first available component of the [Warlock Toolchain](https://github.com/Lokee86/warlock-toolchain), a repository intelligence and governance toolchain for preserving repository knowledge, making working context portable, and enforcing durable boundaries for humans and software agents.

Archivist owns documentation integrity and maintenance. [ArcanaGraph](https://github.com/Lokee86/arcana-graph) owns language-independent repository relationships and code intelligence. The planned **Grimoire Context** tool owns context discovery, selection, packaging, and delivery. These are composable sibling tools, not unfinished Archivist feature areas.

**Preserve the lore. Bind the doctrine. Enforce the wards.**

## Current release

Version `0.3.5` adds path-scoped watcher validation for ordinary Markdown edits, independent frontmatter and document-format cache identities, selective cache refresh after generated rewrites, and an explicit `ddocs fix --all` mode for policy mutation. Bare `ddocs fix` now remains focused on indexes, links, and configured reverse indexes.

## Core behavior

Archivist can:

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
- load and prepare documentation-index and code-folder reverse-index work through bounded workers before deterministic serial merge and write application;
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

Most documentation tools treat Markdown as a collection of files located at paths. Archivist treats a repository as a versioned document graph with stable identity, history, and deterministic reconciliation.

A document can retain its identity when its path changes. Content hashes determine whether validation and inventory results remain reusable, whether evidence has materially changed, and whether a recorded repair can still be applied safely. Private Git-style objects, references, and transactions preserve repository state, path history, review decisions, and guarded undo data without requiring generated metadata in the documents themselves.

The individual techniques are familiar from version control, content-addressed storage, build systems, and databases. Their composition is the unusual part: stable document identity, content fingerprinting, repository history, managed ownership boundaries, and graph repair work together to make an ordinary Markdown repository behave like a self-maintaining document system.

## Installation

### Prebuilt release

The recommended end-user path does not require Go or repository compilation.

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

The release workflow runs the complete Go test suite on Windows and Linux, verifies checkout-based `go install`, builds both archives with `CGO_ENABLED=0`, validates checksums and archive contents, and runs the black-box smoke harness against extracted release binaries before publication. macOS does not currently have a prebuilt release asset.

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
--all          every configured reconciliation system
--docs         documentation indexes, configured frontmatter, and document-body format
--indexes      documentation indexes only
--frontmatter  configured frontmatter enforcement only
--format       document-body format enforcement only
--links        repository-local link validation, repair, and orphan checking
--reverse      code-folder reverse indexes
```

Use `ddocs <command> --help` or `ddocs <command> <subcommand> --help` for exact scoped flags and safety behavior. Bare `ddocs fix` runs indexes, links, and configured reverse indexes; use `ddocs fix --all` when frontmatter and document-format mutation should also run. Bare `check` and `watch` retain full configured validation. `demon --help` opens the repository-demon command family. See the [CLI Reference](docs/reference/cli.md) for command ownership and mutation scope.

## Incremental and private-state behavior

Unchanged clean frontmatter and document-format results can be reused from durable `.ddocs/` cache records. Frontmatter reuse is keyed to its raw leading block, policy, selected schema, immutable snapshot, and validation engine. Document-format reuse is keyed to selected schema metadata, document ID and type, and the evaluated H2+ heading tree. Ordinary prose, links, code-block content, and section body edits no longer invalidate either validation subsystem. A standalone read-only check does not initialize `.ddocs/` merely to save cache data.

Link inventory traverses the repository deterministically, reuses unchanged size/mtime metadata, and reads changed or new files through a bounded 16-worker pool. Changed Markdown link sources are also read and parsed through bounded workers; results remain indexed by source path and merge serially before target resolution, identity updates, diagnostics, review policy, and repair planning. For known target moves, each unchanged affected source independently prepares its rewrite plan through the same bounded worker pool, then results merge in source-path order before graph and diagnostic publication. When an index, frontmatter, format, or reverse-index fix changes Markdown after the initial link pass, Archivist refreshes only those changed link sources. A clean non-link fix does not run a repository-wide link scan or initialize absent link state. Explicit `--links` still runs the complete reconciliation, review, rollback, and suppression path.

Documentation-index reconciliation reads and parses existing indexes and parent-editable documents through bounded workers, retains one immutable source snapshot per file, and prepares independent folder updates concurrently. Reverse-index reconciliation likewise inventories selected folders and prepares each managed index independently through a bounded worker pool. Both systems merge updates, matched-entry claims, diagnostics, and errors in deterministic path order before any serial write application.

Automatic private-object compaction is currently disabled. The repository demon and CLI run as separate processes, and go-git pack replacement is not safe until private-state readers and writers share a cross-process lock. Normal commands therefore retain loose objects rather than risking a missing pack or referenced object.

Cold frontmatter and document-format validation, link-inventory reads, changed Markdown source reads and parsing, documentation-index preparation, and reverse-index preparation now use bounded worker pools. Frontmatter and format cache identities are isolated from unrelated body edits. A changed Markdown link source is still reparsed as a whole, while format cache candidates still require reading the file, parsing selection metadata, and scanning its heading structure. Shared command snapshots, incremental link parsing, and metadata-assisted no-read hits remain later optimizations. These remaining performance boundaries are tracked in [Current Product Limitations](docs/limits/current-limitations.md) and the [Roadmap](docs/planning/roadmap.md).

## Performance maturity

The current implementation is correctness-first. It is serviceable on modest repositories, but it is not yet optimized for low-latency operation on large or high-churn trees.

Watcher debounce only delays admission of a reconciliation pass. A burst of filesystem events resets that quiet period repeatedly, and directory moves may produce many events. Ordinary Markdown create and write events now carry changed paths into frontmatter and document-format validation, allowing untouched documents to reuse clean cache state without being read or parsed. Link and folder-index reconciliation remain broader, while schema, control-file, directory, removal, rename, overflow, and uncertain events can still force conservative full validation. Visible repair latency can therefore remain longer than the configured debounce.

The intended production direction is path-aware dirty tracking for link and index reconciliation, shared command snapshots, fewer repeated state reads and writes, and benchmark-guided scheduling. Until then, `ddocs check`, `ddocs fix`, and explicit `ddocs mv` remain the authoritative operational surfaces; the watcher and repository demon are convenience automation rather than performance guarantees.

## Safety model

Archivist owns only explicit deterministic surfaces:

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
- [Release history](CHANGELOG.md)
- [Security policy](SECURITY.md)
- [Contributors](CONTRIBUTORS.md)

Current behavior, future work, and benchmark evidence are intentionally separated. The [Roadmap](docs/planning/roadmap.md) summarizes sequencing but is not the canonical reference for shipped behavior.

## Managed Codemaps

Archivist includes an explicit foreground codemap workflow:

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

## Development

Run the complete local release gate:

```bash
make release-check
```

Run the checked-in black-box smoke harness:

```bash
go run ./tools/smoke
```

It builds fresh binaries, creates an isolated repository and user configuration, and exercises documentation-policy convergence, link repair, explicit moves, reverse indexes, and detached-daemon maintenance. Release binaries can be tested directly with `--ddocs PATH --demon PATH`.

This is a correctness smoke harness, not a load or stress test. It uses a deliberately small fixture to verify that the real CLI, storage, reconciliation, and daemon paths work together. It does not measure high-volume throughput, watcher event storms, concurrent command contention, resource ceilings, or long-running daemon stability.

Run the Go suite directly:

```bash
go test ./... -count=1
```

See [Testing and Fixtures](docs/development/testing-and-fixtures.md) and [Repository Layout](docs/development/repository-layout.md).

## Project status

Repository indexing, frontmatter enforcement, document-body format enforcement, schema-based creation, local-link reconciliation, orphan health checks, stateless moves, reverse indexes, suggestion decisions, applied-change history, watcher/demon lifecycle, schema-aware codemap execution with schema-driven missing-section placement, and codemap research tooling are implemented. Remaining Archivist work is documentation-engine hardening: diagnostics, performance, watcher scope, link and anchor validation, reverse-index reporting, and review-history resilience. Polyglot repository intelligence and deterministic task-context delivery belong to ArcanaGraph and Grimoire Context.

See [Roadmap](docs/planning/roadmap.md) for current status and sequencing.

## License

The current Archivist source tree is available under the [PolyForm Shield License 1.0.0](LICENSE.md). Competing products and services require a separate commercial license. See [LICENSING.md](LICENSING.md) for the current licensing boundary, permitted-use guidance, and earlier-version history.
