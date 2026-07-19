# Demon Docs

Demon Docs is a deterministic documentation maintenance engine for repository-owned Markdown.

It maintains folder indexes, validates and repairs local links, reports orphan documents, supports explicit link-aware moves, projects authored codemaps back onto code folders, and records reviewable repairs without taking ownership of authored prose.

## Core behavior

Demon Docs can:

- maintain recursive folder indexes inside a configured documentation root;
- preserve authored content outside explicit managed blocks;
- validate and repair supported Markdown, wiki, reference, image, and local HTML targets;
- report managed Markdown documents with no meaningful inbound links;
- move a repository-contained file or directory and rewrite affected links without initialization;
- retain stable file identities and path history in a private `.ddocs/` repository;
- expose ambiguous repairs and codemap candidates for accept, decline, or reconsider decisions;
- record applied repairs with bounded, hash-guarded undo and repair blocks;
- project authored codemap references onto configured code folders and files;
- export deterministic codemap datasets and run benchmark or precision research;
- watch relevant filesystem changes in the foreground; and
- run one optional repository-local watcher through the repository demon and feeder lifecycle.

It does not silently rewrite prose, choose among ambiguous targets, recommend removing existing codemap links as irrelevant, or treat inferred research candidates as authored relationships.

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

Use the stateless refactoring command without initializing a repository:

```bash
ddocs mv --dry-run docs/old.md docs/new.md
ddocs mv docs/old.md docs/new.md
```

For persistent indexes, link history, health checks, review history, reverse indexes, and automation, initialize the repository:

```bash
ddocs init --root docs/
ddocs fix
ddocs fix
ddocs check
```

The first link-enabled mutating pass establishes private identity and history state. A second `fix` verifies idempotence before the read-only `check` gate.

Inspect repository selection at any time:

```bash
ddocs status
ddocs config paths
ddocs config show
```

See [Getting Started](docs/guides/getting-started.md) for adoption, ignore rules, subsystem selection, and recovery guidance.

## Primary commands

```text
ddocs init         initialize repository-local configuration
ddocs status       show selected repository and documentation paths
ddocs mv           move a file or directory and rewrite affected links
ddocs check        verify selected systems and report document-health failures
ddocs fix          apply safe deterministic reconciliation
ddocs watch        run reconciliation after relevant filesystem changes
ddocs suggestions  inspect and decide unresolved repair suggestions
ddocs changes      inspect, undo, block, or unblock applied repairs
ddocs config       inspect or initialize configuration
ddocs codemap      export and evaluate codemap evidence
ddocs demon        manage repository-local watcher lifecycle
```

Subsystem selectors:

```text
--docs     documentation folder indexes and parent navigation
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
- configured generated reverse-index regions; and
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

Watch and demon modes are convenience layers. `ddocs check` remains the authoritative CI and recovery surface.

See [CI and Automation](docs/guides/ci-and-automation.md) and [Repository Demon](docs/operations/repository-demon.md).

## Documentation

- [Documentation index](docs/README.md)
- [Documentation policy](docs/documentation-policy.md)
- [Guides](docs/guides/README.md)
- [Reference](docs/reference/README.md)
- [Architecture](docs/architecture/README.md)
- [Operations](docs/operations/README.md)
- [Current limitations](docs/limits/README.md)
- [Research](docs/research/README.md)
- [Planning](docs/planning/README.md)
- [Development](docs/development/README.md)

Current behavior, future work, and benchmark evidence are intentionally separated. The [Roadmap](docs/planning/roadmap.md) summarizes sequencing but is not the canonical reference for shipped behavior.

## Experimental Codemap Suggestions

Demon Docs includes a deterministic codemap missing-link research pipeline. It collects repository evidence, ranks targets, and separates a bounded `hard_link` review surface from broader `context` relationships.

The current baseline is suitable for early implementation testing and dogfooding. Suggestions remain reviewable evidence: the tool does not automatically insert permanent codemap links, does not recommend removing existing links, and does not treat a candidate as semantic truth.

See:

- [Codemap Suggestion Algorithm](docs/codemap-suggestion-algorithm.md) for current behavior and measured readiness;
- [Codemap Algorithm Development Log](docs/codemap-algorithm-development-log.md) for the full benchmark and tuning history; and
- [Codemap Missing-Link Evidence](docs/codemap-evidence.md) for the evidence and safety boundary.

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

Repository indexing, local-link reconciliation, orphan health checks, stateless moves, reverse indexes, suggestion decisions, applied-change history, watcher/demon lifecycle, and codemap research tooling are implemented. Broader diagnostics, polyglot code intelligence, and deterministic agent context remain active or planned work.

See [Roadmap](docs/planning/roadmap.md) for current status and sequencing.

## License

See [LICENSE](LICENSE).
