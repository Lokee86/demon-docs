# Demon Docs

Demon Docs is a deterministic documentation maintenance engine for repository-owned Markdown.

It keeps folder indexes, local links, reverse code-folder indexes, and optional repository-local automation synchronized without taking ownership of authored prose.

## Core behavior

Demon Docs can:

- maintain recursive folder indexes inside a configured documentation root;
- preserve authored content outside explicit managed blocks;
- validate and repair supported local Markdown, wiki, reference, image, and HTML file targets;
- retain stable file identities and path history in a private `.ddocs/` repository;
- project authored codemap references back onto configured code folders and files;
- export deterministic codemap datasets and run benchmark or precision research;
- watch relevant filesystem changes in the foreground; and
- run one optional repository-local watcher through the repository demon and feeder lifecycle.

It does not silently rewrite prose, choose among ambiguous targets, remove existing codemap links as irrelevant, or treat inferred research candidates as authored relationships.

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

From the repository to manage:

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
ddocs init       initialize repository-local configuration
ddocs status     show selected repository and documentation paths
ddocs check      verify selected systems without authored-file writes
ddocs fix        apply safe deterministic reconciliation
ddocs watch      run reconciliation after relevant filesystem changes
ddocs config     inspect or initialize configuration
ddocs codemap    export and evaluate codemap evidence
ddocs demon      manage repository-local watcher lifecycle
```

Subsystem selectors:

```text
--docs     documentation folder indexes and parent navigation
--links    repository-local link validation and repair
--reverse  code-folder reverse indexes
```

Use `ddocs <command> --help` for exact flags. See the [CLI Reference](docs/reference/cli.md) for command ownership and mutation scope.

## Safety model

Demon Docs owns only explicit deterministic surfaces:

- content between managed index markers;
- configured parent-index navigation lines;
- the path portion of a recognized local link when one destination is deterministic;
- configured generated reverse-index regions; and
- private state under `.ddocs/`.

Labels, titles, aliases, queries, fragments, surrounding prose, source newline style, and final-newline state are preserved during supported link rewrites.

Ambiguous targets remain unchanged and are reported for user resolution.

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
- [Research](docs/research/README.md)
- [Planning](docs/planning/README.md)
- [Development](docs/development/README.md)

Current behavior, future work, and benchmark evidence are intentionally separated. The [Roadmap](docs/planning/roadmap.md) summarizes planned direction but is not the canonical reference for shipped behavior.

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

Demon Docs is under active development. Repository indexing, local-link reconciliation, reverse indexes, watcher/demon lifecycle, and codemap research tooling are implemented. Reviewable suggestion decisions, broader diagnostics, and the polyglot code-intelligence track remain active or planned work.

See [Roadmap](docs/planning/roadmap.md) for current status and sequencing.

## License

See [LICENSE](LICENSE).
