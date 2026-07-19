# Demon Docs

## Motivation
## Usage
## Contributing

`Demon Docs` keeps folder indexes and local Markdown links synchronized with the filesystem.

It maintains folder indexes inside the configured docs root and a focused local-link graph across Markdown files in the repository. Link targets may be Markdown files, non-Markdown assets, directories, repository-external relative paths, or absolute filesystem paths.

You keep owning the actual files and hand-written content. Index reconciliation owns only clearly marked managed sections, while link reconciliation changes only the target path inside an existing Markdown link. `check` reports pending index or link work, `fix` applies it, and `watch` runs the same operations automatically after relevant filesystem changes.

The result is a file tree that can be moved, split, expanded, or reorganized without leaving stale index pages behind.

## What it manages

`Demon Docs` reconciles:

- folder index files, such as `README.md`
- direct file entries in each folder index
- draft/stub file entries from a configured draft folder
- direct child-folder entries
- `Parent index` links in folder indexes by default, and in indexed files when configured
- local Markdown links to files and directories inside or outside the repository
- stable file identities, link history, reverse indexes, and generated-write state in the private `.ddocs/` object repository

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

From the root of the repository you want Demon Docs to manage:

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
ddocs fix --help
ddocs check --help
ddocs watch --help
ddocs config paths
ddocs config show
ddocs config init --local
ddocs config init --global
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
ddocs check -i      # indexes only
ddocs check -l      # links only
ddocs fix --indexes
ddocs fix --links
ddocs watch -i
ddocs watch -l
```

Using both selectors, or neither selector, runs both systems.

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

If a manual `fix` reports `0 file(s)` changed after files changed, a watcher may already have reconciled the tree. Check the watcher log.

## Automation example

A shell startup file can launch the watcher with a PID guard:

```bash
DDOCS_ROOT="${DDOCS_ROOT:-docs}"

ddocs_pid_file="$PWD/.cache/ddocs-watch.pid"
ddocs_log_file="$PWD/.cache/ddocs-watch.log"

mkdir -p "$PWD/.cache"

ddocs_watch_is_running() {
  [ -s "$ddocs_pid_file" ] || return 1

  local watcher_pid
  watcher_pid="$(cat "$ddocs_pid_file" 2>/dev/null)" || return 1

  case "$watcher_pid" in
    ''|*[!0-9]*) return 1 ;;
  esac

  kill -0 "$watcher_pid" 2>/dev/null || return 1
  ps -p "$watcher_pid" -o args= 2>/dev/null | grep -Fq "ddocs watch"
}

start_ddocs_watch() {
  setsid bash -c '
    cd "$1" || exit 1
    exec ddocs watch --root "$2" </dev/null >>"$3" 2>&1
  ' _ "$PWD" "$DDOCS_ROOT" "$ddocs_log_file" >/dev/null 2>&1 &

  echo $! > "$ddocs_pid_file"
}

if ! ddocs_watch_is_running; then
  rm -f "$ddocs_pid_file"
  start_ddocs_watch
fi

unset ddocs_pid_file
unset ddocs_log_file
```

For `direnv`, source process startup scripts outside any `set -a` block so helper variables are not exported.

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

- validate semantic documentation quality
- decide what a folder should own
- validate heading-anchor existence yet
- modify local link targets, binary files, or files outside the repository
- inspect Git status
- guess when more than one link target is plausible

Index writes are confined to the configured docs root. Link rewrites are confined to Markdown source files inside the repository root. Symbolic-link entries are not traversed or edited.

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
