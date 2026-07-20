# Demon Docs

Demon Docs is a deterministic documentation maintenance engine for repository-owned Markdown.

It maintains folder indexes, validates and repairs local links, reports orphan documents, supports explicit link-aware moves, manages configured codemap sections, projects codemap references back onto code folders, and records reviewable repairs while limiting ownership to explicit managed surfaces.

Configured repositories can also enforce frontmatter fields and document-body structure, create documents from TOML document schemas, and resolve explicit format conflicts without rewriting authored prose.

## Core behavior

Demon Docs can:

- maintain recursive folder indexes inside a configured documentation root;
- preserve authored content outside explicit managed blocks;
- validate and repair supported Markdown, wiki, reference, image, and local HTML targets;
- report managed Markdown documents with no meaningful inbound links;
- move a repository-contained file or directory and rewrite affected links without initialization;
- retain stable file identities and path history in private `.ddocs/` state, under the standalone docs root or the initialized repository root;
- expose ambiguous repairs and codemap candidates for decline, reconsider, or compatibility selection decisions;
- record applied normal repairs with bounded, hash-guarded undo and repair blocks;
- explicitly inspect, preview, update, and verify unified managed codemap sections;
- preserve existing codemap links by default while supporting opt-in confidence pruning;
- project codemap references onto configured code folders and files;
- export deterministic codemap datasets and run benchmark or precision research;
- watch relevant filesystem changes in the foreground; and
- run one optional repository-local watcher through the repository demon and feeder lifecycle.

It does not silently rewrite prose outside explicit managed regions, choose among ambiguous targets, remove codemap links by confidence unless configured to do so, or invoke codemap generation through normal watch or daemon automation.

## Installation

Go is the supported implementation and runtime.

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

Verify installation:

```bash
ddocs --version
ddocs --help
demon --help
```

`ddocs` is the canonical executable. `demon` is an alias backed by the same application implementation.

## Quick start

Run index, link, health, move, and foreground-watch operations without initializing a repository:

```bash
ddocs fix --root docs --docs
ddocs fix --root docs --links
ddocs watch --root docs --once
ddocs mv --dry-run docs/old.md docs/new.md
ddocs mv docs/old.md docs/new.md
ddocs check --root docs --docs --links
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
--frontmatter  configured frontmatter enforcement only
--format       document-body format enforcement only
--links    repository-local link validation, repair, and orphan checking
--reverse  code-folder reverse indexes
```

Use `ddocs <command> --help` or `ddocs <command> <subcommand> --help` for exact scoped flags and safety behavior. `demon --help` opens the repository-demon command family. See the [CLI Reference](docs/reference/cli.md) for command ownership and mutation scope.

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
