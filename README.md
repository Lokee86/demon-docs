# doc-ledger

## Motivation
## Usage
## Contributing

`doc-ledger` keeps folder index files in sync with a file tree.

Point it at a root folder, and it scans the folders and files inside it. For each folder, it creates or updates a local index file that lists the folder’s direct files, draft/stub files, and child folders. It can also add parent-index links so readers can move back up the tree.

You keep owning the actual files and any hand-written index content. `doc-ledger` only owns clearly marked managed sections. `check` reports when the indexes no longer match the filesystem, and `fix` updates them. `watch` starts a persistent process that will watch a root folder and all its children and automatically updates indexes with any changes.

The result is a file tree that can be moved, split, expanded, or reorganized without leaving stale index pages behind.

## What it manages

`doc-ledger` reconciles:

- folder index files, such as `README.md`
- direct file entries in each folder index
- draft/stub file entries from a configured draft folder
- direct child-folder entries
- `Parent index` links in folder indexes by default, and in indexed files when configured

It preserves hand-authored content outside managed index blocks.

## Installation

The Go implementation is the primary doc-ledger executable. Install it from a checkout:

```bash
git clone https://github.com/Lokee86/doc-ledger.git
cd doc-ledger
go install ./cmd/doc-ledger
```

Or build a repository-local binary:

```bash
go build -o bin/doc-ledger ./cmd/doc-ledger
```

Ensure the Go install directory is on `PATH`, then verify the executable:

```bash
doc-ledger --help
doc-ledger --version
```

## Quick Start

From the `doc-ledger` repo:

```bash
doc-ledger fix --root docs
doc-ledger check --root docs
```

`fix` writes needed updates.

`check` verifies the same reconciliation without writing files.

## Development

Run the complete local release gate, including the retained Python suite and exact-byte Python/Go parity:

```bash
python -m pip install -e ".[dev]"
make release-check
```

The release gate runs focused Go tests, the retained Python suite, exact-byte Python/Go parity, `go vet`, an executable build, and CLI smoke checks. GitHub Actions repeats Go and parity coverage on Linux and Windows.

Build and run directly from the checkout:

```bash
go run ./cmd/doc-ledger fix
go run ./cmd/doc-ledger check
```

Installed usage is:

```bash
doc-ledger fix
doc-ledger check
doc-ledger watch
```

The original Python implementation remains under `doc_ledger/` only as a legacy parity reference. It is not a canonical installed entrypoint. When differential testing requires it, use:

```bash
python main.py fix
```

Two intentional corrections from the legacy reference are part of the Go contract: headings and marker-like comments inside fenced code blocks are treated as code, and source files retain their original final-newline state. Both behaviors have focused byte-level tests and are not hidden by parity normalization.

CLI help is available at the top level and for each subcommand:

```bash
doc-ledger --help
doc-ledger -v
doc-ledger --version
doc-ledger fix --help
doc-ledger check --help
doc-ledger watch --help
doc-ledger config paths
doc-ledger config show
doc-ledger config init --local
doc-ledger config init --global
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
doc-ledger fix --root docs
```

Reconciles the docs tree and writes updates.

```bash
doc-ledger check --root docs
```

Verifies that the docs tree is already reconciled. Returns non-zero if `fix` would change files.

```bash
doc-ledger watch --root docs
```

Runs one reconciliation immediately, then watches the docs tree for relevant filesystem changes.

```bash
doc-ledger watch --root docs --once
```

Runs the watcher path once and exits.

A config file can replace repeated command flags:

```bash
doc-ledger fix --config .doc-ledger.toml
doc-ledger check --config .doc-ledger.toml
```

Config commands:

```bash
doc-ledger config paths
doc-ledger config show
doc-ledger config init --local
doc-ledger config init --global
```

## Config Selection

doc-ledger selects one base config before applying command-specific CLI overrides.

Selection order:

1. `--config PATH`
2. current-directory `.doc-ledger.toml`
3. current-directory `doc-ledger.toml`
4. global user config
5. built-in defaults

There is no upward parent-directory search and no merge between local and global config files.

`--root` still overrides the selected base config root.

CLI override examples:

```bash
doc-ledger fix --root docs --index-file "!README.md"
doc-ledger fix --root docs --draft-folder "_drafts"
doc-ledger fix --root docs --include "**/*.png"
doc-ledger fix --root docs --exclude "**/*.tmp"
doc-ledger fix --root docs --marker-prefix "nav-ledger"
doc-ledger fix --root docs --parent-label "Back to Index"
doc-ledger fix --root docs --parent-link-folder-indexes
doc-ledger fix --root docs --no-parent-link-folder-indexes
doc-ledger fix --root docs --parent-link-indexed-files
doc-ledger fix --root docs --no-parent-link-indexed-files
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

`doc-ledger` maintains:

```text
docs/README.md
docs/guides/README.md
```

It indexes `future-topic.md` from the parent folder’s stub section, not from `stubs/README.md`.

## Managed sections

`doc-ledger` owns only the content between its marker comments.

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

`doc-ledger` maintains parent navigation lines where configured.

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
doc-ledger fix --root docs --parent-link-indexed-files
doc-ledger fix --root docs --no-parent-link-folder-indexes
```

## Description preservation

`doc-ledger` tries to preserve existing index descriptions.

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

If a stale entry no longer maps to a current file or folder, `doc-ledger` removes it and reports a reconciliation message.

## Configuration

`doc-ledger` looks for config files named:

```text
.doc-ledger.toml
doc-ledger.toml
```

Selection order:

1. `--config PATH`
2. current-directory `.doc-ledger.toml`
3. current-directory `doc-ledger.toml`
4. global user config
5. built-in defaults

Local config lookup is current-directory only.
There is no upward parent-directory search.
Local and global config files are not merged.
CLI flags override the selected config.

`--root` overrides the configured root.

Minimal config:

```toml
root = "docs"
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

## Watch mode

Watch mode is for local convenience. It is not a replacement for `check`.

The watcher:

- runs one reconciliation immediately on startup
- watches the configured root recursively
- reacts to relevant file and directory events
- debounces noisy event bursts
- runs one reconciliation at a time
- schedules a follow-up pass if changes arrive during a run
- logs timestamps and process IDs so watcher/fix races are visible

Example:

```bash
doc-ledger watch --root docs
```

If a manual `fix` reports `0 file(s)` changed after files changed, a watcher may already have reconciled the tree. Check the watcher log.

## Automation example

A shell startup file can launch the watcher with a PID guard:

```bash
DOC_LEDGER_ROOT="${DOC_LEDGER_ROOT:-docs}"

doc_ledger_pid_file="$PWD/.cache/doc-ledger-watch.pid"
doc_ledger_log_file="$PWD/.cache/doc-ledger-watch.log"

mkdir -p "$PWD/.cache"

doc_ledger_watch_is_running() {
  [ -s "$doc_ledger_pid_file" ] || return 1

  local watcher_pid
  watcher_pid="$(cat "$doc_ledger_pid_file" 2>/dev/null)" || return 1

  case "$watcher_pid" in
    ''|*[!0-9]*) return 1 ;;
  esac

  kill -0 "$watcher_pid" 2>/dev/null || return 1
  ps -p "$watcher_pid" -o args= 2>/dev/null | grep -Fq "doc-ledger watch"
}

start_doc_ledger_watch() {
  setsid bash -c '
    cd "$1" || exit 1
    exec doc-ledger watch --root "$2" </dev/null >>"$3" 2>&1
  ' _ "$PWD" "$DOC_LEDGER_ROOT" "$doc_ledger_log_file" >/dev/null 2>&1 &

  echo $! > "$doc_ledger_pid_file"
}

if ! doc_ledger_watch_is_running; then
  rm -f "$doc_ledger_pid_file"
  start_doc_ledger_watch
fi

unset doc_ledger_pid_file
unset doc_ledger_log_file
```

For `direnv`, source process startup scripts outside any `set -a` block so helper variables are not exported.

## Testing

Run the test suite:

```bash
go test ./...
```

Run only the cross-implementation parity test with `make parity`. The legacy Python suite can still be run with `python -m pytest tests -q`; it is retained to document the pre-port behavior and is not the primary release gate.

Useful manual smoke flow:

```bash
doc-ledger fix --root docs
doc-ledger check --root docs
```

For fixture stress testing, see:

```text
docs/make-dummy-docs.sh
```

That script generates a synthetic documentation tree for manual reconciliation tests.

## Safety boundaries

`doc-ledger` does not:

- validate semantic documentation quality
- decide what a folder should own
- rewrite arbitrary links inside document bodies
- modify binary or non-editable files
- inspect Git status
- guarantee perfect rename detection

It only reconciles the filesystem against the configured index model.

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
doc-ledger
doc-ledger.exe
__pycache__/
*.pyc
.pytest_cache/
.cache/
dummy-docs/
```

The repo `.gitignore` should exclude those paths.
