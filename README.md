# Demon Docs

## Motivation
## Usage
## Contributing

`Demon Docs` keeps documentation indexes, repository-local links, and authored code maps synchronized with the repository.

It maintains folder indexes inside the configured docs root, a focused local-link graph across repository Markdown, and deterministic codemap datasets used to evaluate possible missing documentation-to-code links. Supported local-link forms include ordinary Markdown links and images, reference definitions and uses, path-based wiki links, and common local HTML targets.

You keep owning the actual files and hand-written content. Index reconciliation owns only clearly marked managed sections. Link reconciliation changes only a resolved destination path while preserving authored labels, titles, aliases, queries, fragments, and surrounding prose. Codemap analysis exports and ranks evidence but does not silently add or remove authored code-map links.

`check` reports pending or unresolved work, `fix` applies safe deterministic updates, `watch` runs the same reconciliation core after relevant filesystem changes, and `demon` provides optional repository-local background lifecycle around that watcher. `mv` is a separate stateless refactoring command that moves a file or directory and rewrites affected links without requiring initialization.

## What it manages

`Demon Docs` reconciles:

- folder index files, such as `README.md`
- direct file entries in each folder index
- draft/stub file entries from a configured draft folder
- direct child-folder entries
- `Parent index` links in folder indexes by default, and in indexed files when configured
- local Markdown links, images, reference definitions, wiki links, and common HTML file targets
- explicit file and directory moves with affected-link rewrites, dry-run planning, and no initialization requirement
- undefined explicit and collapsed Markdown reference labels
- stable file identities, link history, reverse indexes, and generated-write state in the private `.ddocs/` object repository
- deterministic codemap extraction, evidence collection, holdout benchmarking, precision sampling, and tiered missing-link candidates
- repository-local suggestion decisions, Git-backed applied-change history, bounded undo, and repair blocks
- optional repository-local watcher ownership, feeder heartbeats, linked-worktree state, and bounded logs

It preserves hand-authored content outside managed index blocks and preserves link labels, titles, query strings, and fragments when updating a path.

## Installation

Go is the sole implementation and supported runtime for Demon Docs. Install it from a checkout:

```bash
git clone https://github.com/Lokee86/demon-docs.git
cd demon-docs
go install ./cmd/ddocs
go install ./cmd/demon
```

Or build a repository-local binary:

```bash
go build -o bin/ddocs ./cmd/ddocs
go build -o bin/demon ./cmd/demon
```

Ensure the Go install directory is on `PATH`, then verify the executable:

```bash
ddocs --help
ddocs --version
demon --help
demon --version
```

`ddocs` is the canonical command. `demon` is an installed alias backed by the same internal app implementation and has identical behavior.

## Quick Start

Use Demon Docs on an ordinary Markdown repository without initializing it:

```bash
ddocs mv --dry-run docs/old.md docs/new.md
ddocs mv docs/old.md docs/new.md
```

The command scans from the current directory by default, or from the nearest initialized repository root when one exists. It moves files or directories and rewrites affected relative and incoming links without creating `.ddocs/`.

For persistent index, link-history, watcher, and repository-management features, initialize the repository:

```bash
ddocs init --root docs/
ddocs fix
ddocs check
```

`init` creates `.ddocs/config.toml`, records `docs/` as the docs root, and makes the current directory the repository root.

`fix` writes needed updates. The first link-enabled `fix` establishes the private `.ddocs/` object repository without repairing links; later passes can use that baseline to reconcile moves. `check` verifies the same reconciliation without writing files. Both commands can then be run from anywhere inside the repository.

## Development

Run the complete local Go release gate:

```bash
make release-check
```

The release gate runs focused Go tests, the Go CLI fixture regression matrix, `go vet`, both executable builds, and CLI smoke checks for `ddocs` and `demon`. GitHub Actions runs the complete Go suite on Linux and Windows.

Build and run directly from the checkout:

```bash
go run ./cmd/ddocs fix
go run ./cmd/ddocs check
```

Installed usage is:

```bash
ddocs fix
ddocs check
ddocs watch
```

The installed alias behaves identically:

```bash
demon fix
demon check
demon watch
```

Two intentional compatibility corrections are part of the Go contract: headings and marker-like comments inside fenced code blocks are treated as code, and source files retain their original final-newline state. Both behaviors have focused byte-level tests.

CLI help is available at the top level and for each subcommand:

```bash
ddocs --help
ddocs -v
ddocs --version
ddocs init --help
ddocs status --help
ddocs mv --help
ddocs fix --help
ddocs check --help
ddocs watch --help
ddocs config paths
ddocs config show
ddocs config init --local
ddocs config init --global
ddocs codemap --help
ddocs codemap export --help
ddocs codemap benchmark --help
ddocs codemap precision --help
ddocs suggestions --help
ddocs changes --help
ddocs demon --help
```

`-v` and `--version` are top-level version flags.

Default conventions:

```text
docs root:       docs
index file:      README.md
draft folder:    stubs
parent label:    Parent index
marker prefix:   doc-ledger
```

## Commands

```bash
ddocs init --root docs/
```

Initializes the current directory as the repository root and writes `.ddocs/config.toml`. The specified docs root must already exist inside the repository.

```bash
ddocs status
```

Shows the detected repository root, docs root, config path, and repository-owned `.docignore` path.

```bash
ddocs mv [--root PATH] [--dry-run] SOURCE DESTINATION
```

Moves one repository-contained file or directory and rewrites affected local link destinations. It supports ordinary Markdown links and images, reference definitions, path-based wiki links and embeds, and supported local HTML targets. The command does not require or create `.ddocs/`; destination parents must already exist. When `DESTINATION` is an existing directory, the source is moved beneath it.

See [Stateless Document Refactoring](docs/document-refactoring.md) for planning, safety, and rollback behavior.

```bash
ddocs fix
```

Reconciles indexes and links, writes repository-contained updates, and persists link state.

```bash
ddocs check
```

Verifies indexes and links without writing files. It returns non-zero for pending updates, broken or ambiguous links, or uninitialized link state.

```bash
ddocs watch
```

Runs one reconciliation immediately, then watches for relevant filesystem changes. Link-enabled watch mode observes the repository root and the parent directories of explicitly linked external targets so moved non-Markdown targets can trigger repairs.

Select one subsystem explicitly when needed:

```bash
ddocs check -d      # documentation indexes only
ddocs check -l      # links only
ddocs check -r      # reverse indexes only
ddocs fix --docs
ddocs fix --links
ddocs fix --reverse
ddocs watch -d
ddocs watch -l
ddocs watch -r
```

Supplying selectors runs only those systems. Without selectors, documentation indexes and links run; reverse indexes also run when reverse roots are configured or supplied.

```bash
ddocs watch --root docs --once
```

Runs the watcher path once and exits.

A config file can replace repeated command flags:

```bash
ddocs fix --config .demon-docs.toml
ddocs check --config .demon-docs.toml
```

Config commands:

```bash
ddocs config paths
ddocs config show
ddocs config init --local
ddocs config init --global
```

Codemap analysis commands:

```bash
ddocs codemap export --output .ddocs/codemap.json
ddocs codemap benchmark --help
ddocs codemap precision --help
```

`codemap export` scans configured codemap headings and writes a deterministic dataset containing documents, normalized targets, resolution diagnostics, and content hashes. `benchmark` removes known authored targets in controlled holdouts and measures whether the deterministic evidence system recovers them. `precision` generates, samples, and evaluates ranked suggestions against curated labels. Candidates are divided into `hard_link` and `context` tiers; neither tier authorizes automatic documentation edits.

See [Codemap Missing-Link Evidence](docs/codemap-evidence.md) for evidence signals, current benchmark results, safety rules, and research artifacts.

Suggestion review and applied-change history:

```bash
ddocs suggestions [FILE]
ddocs suggestions show SUGGESTION
ddocs suggestions select SUGGESTION [CANDIDATE]
ddocs suggestions decline SUGGESTION [CANDIDATE] --reason "..."
ddocs suggestions reconsider SUGGESTION

ddocs changes [FILE]
ddocs changes related FILE
ddocs changes show CHANGE
ddocs changes undo CHANGE [--repair REPAIR] [--block]
ddocs changes undo-run RUN [--block]
ddocs changes block CHANGE [--repair REPAIR]
ddocs changes unblock CHANGE [--repair REPAIR]
```

Ambiguous link repairs and codemap missing-link candidates are suggestions. Selecting one converts it into the normal hash-guarded repair path. Deterministic single-target link repairs remain automatic but are recorded as inspectable changes. Undo is available by reconciliation run, file change, or individual repair while the recorded after-state still matches. See [Suggestions, Repairs, and Change History](docs/review-ledger.md).

## Config Selection

Demon Docs selects one base config before applying command-specific CLI overrides.

Selection order:

1. `--config PATH`
2. nearest `.ddocs/config.toml`, found by searching upward
3. current-directory `.demon-docs.toml`
4. current-directory `demon-docs.toml`
5. legacy local compatibility fallbacks
6. canonical global user config at `demon-docs/config.toml`
7. legacy global compatibility fallback at `doc-ledger/config.toml`
8. built-in defaults

Repository config is discovered upward. Legacy standalone local configs remain current-directory only. Local and global config files are not merged.

`--root` still overrides the selected docs root for a single command. In an initialized repository, relative overrides resolve from the repository root and cannot escape it.

CLI override examples:

```bash
ddocs fix --root docs --index-file "!README.md"
ddocs fix --root docs --draft-folder "_drafts"
ddocs fix --root docs --include "**/*.png"
ddocs fix --root docs --exclude "**/*.tmp"
ddocs fix --root docs --marker-prefix "nav-ledger"
ddocs fix --root docs --parent-label "Back to Index"
ddocs fix --root docs --parent-link-folder-indexes
ddocs fix --root docs --no-parent-link-folder-indexes
ddocs fix --root docs --parent-link-indexed-files
ddocs fix --root docs --no-parent-link-indexed-files
```

## Folder indexes

Every normal folder under the managed root gets an index file.

By default, the index file is:

```text
README.md
```

Draft folders do not get their own index. By default, the draft folder is:

```text
stubs/
```

For this tree:

```text
docs/
  README.md
  overview.md
  stubs/
    future-topic.md
  guides/
    README.md
    setup.md
```

`Demon Docs` maintains:

```text
docs/README.md
docs/guides/README.md
```

It indexes `future-topic.md` from the parent folder’s stub section, not from `stubs/README.md`.

## Managed sections

`Demon Docs` owns only the content between its marker comments.

Default managed sections:

```markdown
## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
```

Content outside those marker blocks remains hand-authored.

## Parent links

`Demon Docs` maintains parent navigation lines where configured.

Default shape:

```markdown
Parent index: [Folder Name](./README.md)
```

Rules:

- child folder indexes point to `../README.md`
- normal files do not get a parent link by default
- files inside `stubs/` do not get a parent link by default
- the root index has no parent link

The label and index filename are configurable. `indexed_files` turns file-level parent links on when you want them.

Parent-link override flags:

- `--parent-link-folder-indexes`
- `--no-parent-link-folder-indexes`
- `--parent-link-indexed-files`
- `--no-parent-link-indexed-files`

Examples:

```bash
ddocs fix --root docs --parent-link-indexed-files
ddocs fix --root docs --no-parent-link-folder-indexes
```

## Description preservation

`Demon Docs` tries to preserve existing index descriptions.

It preserves descriptions when:

- a file remains in place
- a folder remains in place
- a stub graduates into the parent folder
- a canonical file moves into the stub folder
- a cross-folder move can be matched unambiguously

Stub graduation removes the configured stub prefix:

```markdown
- [topic.md](stubs/topic.md) - Stub: Topic documentation.
```

becomes:

```markdown
- [topic.md](topic.md) - Topic documentation.
```

Moving a canonical file into the stub folder applies the reverse transformation.

If a stale entry no longer maps to a current file or folder, `Demon Docs` removes it and reports a reconciliation message.

## Configuration

The canonical repository config is:

```text
.ddocs/config.toml
```

It is created by `ddocs init --root <docs-root>` and discovered by searching upward from the current directory.

Legacy standalone and global configs remain supported:

```text
.demon-docs.toml
demon-docs.toml
demon-docs/config.toml
```

Compatibility fallbacks remain supported at lower priority:

```text
.doc-ledger.toml
doc-ledger.toml
doc-ledger/config.toml
```

Selection order:

1. `--config PATH`
2. nearest `.ddocs/config.toml`, found by searching upward
3. current-directory `.demon-docs.toml`
4. current-directory `demon-docs.toml`
5. legacy local compatibility fallbacks
6. canonical global user config at `demon-docs/config.toml`
7. legacy global compatibility fallback at `doc-ledger/config.toml`
8. built-in defaults

Repository config is discovered upward. Legacy local config lookup is current-directory only. Local and global config files are not merged. CLI flags override the selected config.

`--root` overrides the configured docs root.

Minimal repository config:

```toml
docs_root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = true
indexed_files = false
```

Use the legacy compatibility switch if you want the older single flag:

```toml
root = "docs"

[parent_link]
enabled = true
```

Use file-level parent links:

```toml
root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = true
indexed_files = true
```

Disable all parent links:

```toml
root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = false
indexed_files = false
```

Use a custom draft folder:

```toml
[drafts]
folder = "_drafts"
description_prefix = "Draft: "
```

Customize section headings:

```toml
[sections.files]
heading = "Files"

[sections.stubs]
heading = "Drafts"

[sections.folders]
heading = "Folders"
```

Customize marker prefix:

```toml
[markers]
prefix = "docs-index"
```

Index non-Markdown files without editing them:

```toml
[files]
include_patterns = ["**/*.md", "**/*.png", "**/*.pdf", "**/*.yaml"]

[editable]
parent_index_extensions = [".md", ".mdx"]
```

## Ignoring paths

Place `.docignore` at the repository root, beside `.ddocs/`, to exclude files and directories from index scanning, link scanning, and watch events. It uses Git ignore syntax and is independent from `.gitignore`.

Patterns are relative to the repository root. Legacy standalone configurations continue treating the docs root as the ignore root.

```gitignore
# Generated exports inside the docs root
/docs/generated/

# Private notes anywhere below docs/
docs/**/*.private.md
docs/scratch/**
!docs/scratch/README.md
```

These directory names are always pruned at any depth and cannot be re-included:

```text
.git/
.ddocs/
.obsidian/
logseq/
```

Watch mode watches and reloads the repository-root `.docignore` even when the docs root is a nested directory.

## Repository demon

Initialized repositories permit the self-managing repository watcher by
default. One fresh demon owner serves each local `.ddocs/` repository while
shell or agent feeders remain active:

```bash
demon run
demon --status
demon --logs
demon acquire --client mcp
demon heartbeat --token TOKEN
demon release --token TOKEN
```

The same lifecycle commands are available as `ddocs demon ...`. `demon acquire`, `heartbeat`, and `release` form the host-neutral feeder interface used by MCP, Codex, Hermes, and other agent adapters.

Install automatic shell entry and exit tracking in Bash with:

```bash
eval "$(ddocs demon __shell-hook bash)"
```

Or in a PowerShell profile with:

```powershell
Invoke-Expression (& ddocs demon __shell-hook powershell)
```

Shell feeding is implemented by the CLI. MCP and native agent integrations use
the public host-neutral lifecycle: acquire a token with a client name, refresh
it before feeder expiry, and release it on every terminal path. The demon does
not host those integrations or deliver agent context.

Runtime ownership, feeder heartbeats, shutdown requests, and bounded logs live
under `.ddocs/runtime/`. The existing `ddocs watch` command remains a foreground
watcher for explicit terminal-controlled use. See
[Repository Demon](docs/repository-demon.md).

## Watch mode

Watch mode is for local convenience. It is not a replacement for `check`.

The watcher:

- runs one reconciliation immediately on startup
- watches the docs root for index-only operation, or the repository root when links are enabled
- reacts to relevant file and directory events
- debounces noisy event bursts
- runs one reconciliation at a time
- schedules a follow-up pass if changes arrive during a run
- logs timestamps and process IDs so watcher/fix races are visible

Example:

```bash
ddocs watch --root docs
```

If a manual `fix` reports `0 file(s)` changed after files changed, a watcher may already have reconciled the tree. Check the watcher log. Generated link rewrites use a bounded worker pool while preserving deterministic planning and per-source atomic replacement.

## Watcher automation

Do not add a second PID-file or `setsid` wrapper around `ddocs watch` when the
repository demon is enabled. The demon already owns detached startup,
single-owner coordination, feeder heartbeats, shutdown grace, and logs.

Use foreground `ddocs watch` only when you deliberately want the process
attached to the current terminal. See [Watcher and
Automation](docs/watcher-and-automation.md) for the distinction.

## Testing

Run the test suite:

```bash
go test ./... -count=1
```

Run the focused Go package tests and the Go CLI fixture regression matrix separately when needed:

```bash
make test-go
make regression
```

The regression matrix retains ten fixture scenarios and validates each with `fix`, a clean successful `check`, a second `fix`, and byte-identical complete fixture trees after the first and second fixes.

Useful manual smoke flow:

```bash
ddocs init --root docs/
ddocs fix
ddocs check
```

For fixture stress testing, see:

```text
docs/make-dummy-docs.sh
```

That script generates a synthetic documentation tree for manual reconciliation tests.

## Safety boundaries

`Demon Docs` does not:

- decide semantic documentation quality or ownership
- validate heading-anchor existence yet
- rewrite link labels, titles, aliases, surrounding prose, binary files, or external target files
- move paths outside the selected `ddocs mv` repository boundary or overwrite an existing non-directory destination
- treat a codemap suggestion as an authored relationship
- recommend removing an existing codemap link as irrelevant
- guess when more than one link or code target is plausible
- perform arbitrary historical selective reverts through later user edits

Index writes are confined to the configured docs root. Link rewrites are confined to repository Markdown source files and require one deterministic target. Stateless moves are confined to their selected repository boundary, verify affected Markdown content before applying, and refuse affected ambiguous wiki targets rather than guessing. Codemap commands remain export, benchmark, and review tooling; they do not silently modify authored codemap sections. Symbolic-link entries are not traversed or edited.

## Using `!README.md`

Some repos prefer `!README.md` so folder indexes sort first in file explorers.

That is supported through config:

```toml
root = "docs"
index_file = "!README.md"
```

With that config, parent links use `!README.md` automatically:

```markdown
Parent index: [Guides](./!README.md)
```

## Repository hygiene

Do not commit runtime artifacts:

```text
bin/
ddocs
demon
ddocs.exe
demon.exe
.cache/
dummy-docs/
```

The repo `.gitignore` should exclude those paths.
