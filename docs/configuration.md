# Demon Docs Configuration

Demon Docs is configured with TOML. The primary config model lives in `internal/config/config.go` and is exercised by Go package tests and the Go CLI fixture regression matrix.

CLI help is available with `ddocs --help`, and each subcommand also supports `--help`.
Top-level version output is available with `ddocs -v` or `ddocs --version`.
Initialize a repository from its root directory:

```bash
ddocs init --root docs/
```

This creates `.ddocs/config.toml`. Commands run anywhere below that directory search upward for `.ddocs/`, treat its parent as the repository root, and resolve the configured docs root from there. `ddocs status` prints the resolved repository, docs root, config, and `.docignore` paths.

The `config` subcommand provides:

- `ddocs config paths`
- `ddocs config show`
- `ddocs config init --local`
- `ddocs config init --global`

## What Configuration Controls

The supported keys are:

- `docs_root`
- `root` as a legacy standalone-config alias
- `index_file`
- `[markers].prefix`
- `[parent_link].label`
- `[parent_link].folder_indexes`
- `[parent_link].indexed_files`
- `[parent_link].enabled` for compatibility with older configs
- `[sections.files].heading`
- `[sections.stubs].heading`
- `[sections.folders].heading`
- `[aliases].files`
- `[aliases].folders`
- `[drafts].folder`
- `[drafts].description_prefix`
- `[files].include_patterns`
- `[files].exclude_patterns`
- `[editable].parent_index_extensions`
- `[descriptions].file_template`
- `[descriptions].folder_template`
- `[watch].debounce_seconds`
- `[watch].ignored_dirs`
- `[watch].ignored_suffixes`
- `[template].include_ownership`
- `[template].include_does_not_belong`
- `[template].include_related_docs`
- `[template].include_notes`

## Selection

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

Compatibility fallbacks remain supported at lower priority:

- `.doc-ledger.toml`
- `doc-ledger.toml`
- `doc-ledger/config.toml`

`--root` still overrides the selected docs root for `fix`, `check`, and `watch`.

`ddocs config show` prints the selected base config.
`ddocs config paths` prints the discovered repository config, current-directory legacy candidates, and global user config paths.
`ddocs config init --local` writes `.demon-docs.toml` in the current directory.
`ddocs config init --global` writes the global config file and creates parent directories as needed.

CLI flags override the selected base config. The reconciliation selectors are operational flags rather than persistent configuration:

- `-i` / `--indexes` runs index reconciliation only when used without the link selector.
- `-l` / `--links` runs link reconciliation only when used without the index selector.
- Supplying both selectors, or neither selector, runs both systems.
- The selectors apply to `check`, `fix`, and `watch`.

Examples include:

```bash
ddocs check -i
ddocs check -l
ddocs fix --indexes
ddocs fix --links
ddocs watch -i
ddocs watch -l
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

## Default Configuration

The defaults reflect the standalone repo behavior:

```toml
docs_root = "docs"
index_file = "README.md"

[markers]
prefix = "doc-ledger"

[parent_link]
label = "Parent index"
folder_indexes = true
indexed_files = false

[sections.files]
heading = "Direct Files"

[sections.stubs]
heading = "Stub Files"

[sections.folders]
heading = "Direct Folders"

[aliases]
files = ["Top-Level Files"]
folders = ["Top-Level Folders"]

[drafts]
folder = "stubs"
description_prefix = "Stub: "

[files]
include_patterns = ["**/*.md"]
exclude_patterns = []

[editable]
parent_index_extensions = [".md"]

[descriptions]
file_template = "{title} documentation."
folder_template = "{title} documentation."

[watch]
debounce_seconds = 0.75
ignored_dirs = [".cache", "__pycache__"]
ignored_suffixes = ["~", ".swp", ".tmp", ".bak"]

[template]
include_ownership = true
include_does_not_belong = true
include_related_docs = true
include_notes = true
```

## `docs_root` and legacy `root`

`docs_root` sets the documentation tree relative to the repository root containing `.ddocs/`.

- `ddocs init --root docs/` writes `docs_root = "docs"`
- Commands can run from any descendant of the repository root
- `--root` overrides the selected docs root for a single command, resolves relative to the repository root, and cannot escape it

Legacy standalone config files may continue using `root`; both keys load into the same docs-root setting, with `docs_root` taking precedence when both are present.

## `index_file`

`index_file` sets the folder index filename.

- Default: `README.md`
- Example custom value: `!README.md`
- Folder README links and generated folder index paths follow this name
- To keep the legacy filename, set `index_file = "!README.md"`.

Projects that want `!README.md` should set `index_file = "!README.md"` in config.

## `[markers].prefix`

`[markers].prefix` sets the HTML comment prefix for managed sections.

- Default: `doc-ledger`
- The managed blocks use `files`, `stubs`, and `folders` section ids

## `[parent_link].label` and `[parent_link].folder_indexes` / `[parent_link].indexed_files`

`[parent_link].label` sets the text used for parent index lines.

- Default: `Parent index`
- Example: `Parent`

`[parent_link].folder_indexes` controls parent links in folder index files.

- Default: `true`
- When `false`, Demon Docs does not insert or update parent links in child folder index files

`[parent_link].indexed_files` controls parent links in indexed files such as `page.md` and `topic.md`.

- Default: `false`
- When `true`, Demon Docs inserts or updates parent links in editable indexed files

`[parent_link].enabled` is a compatibility alias for older configs.

- If `enabled` is present and `folder_indexes` or `indexed_files` are not present, the alias applies to both behaviors.
- If `folder_indexes` or `indexed_files` are present, they override the alias for that behavior.

CLI override flags can change parent-link behavior for a single run.

Supported override flags:

- `--parent-link-folder-indexes`
- `--no-parent-link-folder-indexes`
- `--parent-link-indexed-files`
- `--no-parent-link-indexed-files`

Examples:

```bash
ddocs fix --root docs --parent-link-indexed-files
ddocs fix --root docs --no-parent-link-folder-indexes
```

## `[sections.*].heading`

These keys control the visible headings for managed README sections.

- `[sections.files].heading` defaults to `Direct Files`
- `[sections.stubs].heading` defaults to `Stub Files`
- `[sections.folders].heading` defaults to `Direct Folders`

Legacy aliases are configurable through `[aliases]`:

- `[aliases].files` defaults to `["Top-Level Files"]`
- `[aliases].folders` defaults to `["Top-Level Folders"]`

Those aliases are accepted during migration and normalized into the configured managed section headings.

## `[drafts].folder` and `[drafts].description_prefix`

`[drafts].folder` sets the draft folder name.

- Default: `stubs`
- Example custom value: `_drafts`
- Draft folders do not get their own index file
- Files inside the draft folder are indexed in the owning parent folder’s stub section

`[drafts].description_prefix` sets the prefix for draft file descriptions.

- Default: `Stub: `
- Example custom value: `Draft: `

Example:

```toml
[drafts]
folder = "_drafts"
description_prefix = "Draft: "
```

## `[files].include_patterns` and `[files].exclude_patterns`

`[files].include_patterns` controls which files appear in generated folder indexes.

- Default: `["**/*.md"]`
- Patterns are matched relative to the managed docs root
- The index file itself is excluded even if it matches the include patterns
- These patterns do not limit link targets: link reconciliation tracks every non-ignored local target type referenced by repository Markdown

`[files].exclude_patterns` removes files from indexing.

- Default: `[]`
- Excludes are also matched relative to the managed root

Example:

```toml
[files]
include_patterns = ["**/*.md", "**/*.png", "**/*.pdf", "**/*.yaml"]
exclude_patterns = ["**/*.tmp"]
```

## `.docignore`

An initialized repository owns one `.docignore` file at its repository root, beside `.ddocs/`. It excludes paths from index traversal, repository Markdown link scanning, link-target inventory, and watch events.

Rules use Git ignore syntax, including comments, anchored paths, `*`, `**`, directory patterns, and `!` negation. Patterns are relative to the repository root. Legacy standalone configurations continue using the docs root as the ignore root. `.docignore` is independent from `.gitignore`: a Git-tracked file may be excluded from Demon Docs, and a Git-ignored file may still be indexed.

Example:

```gitignore
# Generated exports inside docs/
/docs/generated/

# Private working files below docs/
docs/**/*.private.md
docs/scratch/**

# Re-include one file from an ignored pattern
!docs/scratch/README.md
```

The following directory names are permanently excluded at any depth and cannot be re-included with `!`:

- `.git/`
- `.ddocs/`
- `.obsidian/`
- `logseq/`

Watch mode watches the repository root for `.docignore` changes, reloads the rules, and adds watches for directories that become visible. Index-only watch mode otherwise remains scoped to the docs root; link-enabled watch mode observes the repository root.

## `[editable].parent_index_extensions`

`[editable].parent_index_extensions` controls which indexed files can receive parent index lines.

- Default: `[".md"]`
- Matching is exact and includes the leading dot
- Use this to allow additional editable file types such as `.mdx`

Example:

```toml
[editable]
parent_index_extensions = [".md", ".mdx"]
```

With the example above:

- `page.md` gets a parent index line
- `page.mdx` gets a parent index line
- `diagram.png` can be indexed, but it does not receive a parent index line

## `[descriptions].file_template` and `[descriptions].folder_template`

These templates control fallback descriptions.

- `[descriptions].file_template` defaults to `{title} documentation.`
- `[descriptions].folder_template` defaults to `{title} documentation.`
- `{title}` is replaced with a title-cased name

Examples:

```toml
[descriptions]
file_template = "File: {title}."
folder_template = "Folder: {title}."
```

## `[watch].debounce_seconds`, `[watch].ignored_dirs`, and `[watch].ignored_suffixes`

`[watch].debounce_seconds` controls how quickly the watcher reruns reconciliation after changes.

- Default: `0.75`

`[watch].ignored_dirs` lists directory names the watcher ignores.

- Default: `[".cache", "__pycache__"]`
- These are watcher-only exclusions; shared traversal exclusions belong in `.docignore`

`[watch].ignored_suffixes` lists filename suffixes the watcher ignores.

- Default: `["~", ".swp", ".tmp", ".bak"]`

## `[template].include_*`

These booleans control which optional sections appear in generated README templates.

- `[template].include_ownership`
- `[template].include_does_not_belong`
- `[template].include_related_docs`
- `[template].include_notes`

All four default to `true`.

## Folder Indexes Only

This config uses the default split behavior and keeps parent links in folder indexes only:

```toml
root = "notes"
index_file = "README.md"

[parent_link]
folder_indexes = true
indexed_files = false
```

## File-Level Parent Links

This config keeps parent links in both folder indexes and indexed files:

```toml
root = "notes"
index_file = "README.md"

[parent_link]
folder_indexes = true
indexed_files = true
```

## Disable Parent Links

This config disables parent links everywhere:

```toml
root = "notes"
index_file = "README.md"

[parent_link]
folder_indexes = false
indexed_files = false
```

## Legacy Compatibility Example

This config keeps the legacy `!README.md` folder index filename and uses the compatibility alias:

```toml
root = "notes"
index_file = "!README.md"

[markers]
prefix = "navmark"

[parent_link]
label = "Parent"
enabled = true
```

## Custom Draft Folder Example

This config uses `_drafts` as the draft folder:

```toml
root = "notes"
index_file = "README.md"

[drafts]
folder = "_drafts"
description_prefix = "Draft: "
```

## Non-Markdown Indexing Example

This config indexes Markdown plus image, PDF, and YAML files, while only editing parent links in `md` and `mdx` files:

```toml
root = "notes"
index_file = "README.md"

[files]
include_patterns = ["**/*.md", "**/*.png", "**/*.pdf", "**/*.yaml"]
exclude_patterns = ["**/*.tmp"]

[editable]
parent_index_extensions = [".md", ".mdx"]
```

In that setup:

- `page.md` is indexed and gets a parent index line
- `page.mdx` is indexed and gets a parent index line
- `diagram.png`, `manual.pdf`, and `openapi.yaml` are indexed
- non-editable files are left untouched by parent-link editing

## Link State

Markdown link reconciliation has no required TOML keys. Its persistent, schema-versioned state is stored in the initialized repository's private `.ddocs/` object repository. Demon Docs uses internal go-git object and reference plumbing, but exposes no Git workflow for this state.

The first link-enabled `fix` or `watch` pass establishes this baseline without repairing links. `check -l` is read-only and reports uninitialized state rather than creating it. Legacy `.ddocs/files.json` and `.ddocs/links.json` state is migrated on the next successful link-state publication.

## Related Files

- `internal/config/config.go`
- `internal/links/`
- `internal/config/config_test.go`
