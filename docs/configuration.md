# doc-ledger Configuration

Parent index: [Docs](./README.md)

doc-ledger is configured with TOML. The config model lives in `doc_ledger/config.py` and is exercised by `tests/test_config.py` and `tests/test_public_config_end_to_end.py`.

## What Configuration Controls

The supported keys are:

- `root`
- `index_file`
- `[markers].prefix`
- `[parent_link].label`
- `[parent_link].enabled`
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

## Discovery

doc-ledger discovers configuration by walking upward from the current working directory.

- `.doc-ledger.toml` is preferred over `doc-ledger.toml` in the same directory.
- A nearer config wins over a farther config.
- `--config` overrides discovery and uses the explicit file.
- `--root` overrides the configured `root` value.

If no config file is found, doc-ledger uses the built-in defaults.

## Default Configuration

The defaults reflect the standalone repo behavior:

```toml
root = "docs"
index_file = "README.md"

[markers]
prefix = "doc-ledger"

[parent_link]
label = "Parent index"
enabled = true

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
ignored_dirs = [".git", ".cache", "__pycache__"]
ignored_suffixes = ["~", ".swp", ".tmp", ".bak"]

[template]
include_ownership = true
include_does_not_belong = true
include_related_docs = true
include_notes = true
```

## `root`

`root` sets the docs tree root.

- Default: `docs`
- Used when `--root` is omitted and no config override is provided
- If `--root` is omitted and a config file exists, the configured `root` is resolved relative to that config file

## `index_file`

`index_file` sets the folder index filename.

- Default: `README.md`
- Example custom value: `!README.md`
- Folder README links and generated folder index paths follow this name
- To keep the legacy filename, set `index_file = "!README.md"`.

## `[markers].prefix`

`[markers].prefix` sets the HTML comment prefix for managed sections.

- Default: `doc-ledger`
- The managed blocks use `files`, `stubs`, and `folders` section ids

## `[parent_link].label` and `[parent_link].enabled`

`[parent_link].label` sets the text used for parent index lines.

- Default: `Parent index`
- Example: `Parent`

`[parent_link].enabled` toggles parent index line management.

- Default: `true`
- When `false`, doc-ledger removes existing parent index lines and does not insert new ones

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

`[files].include_patterns` controls which files are indexed.

- Default: `["**/*.md"]`
- Patterns are matched relative to the managed root
- The index file itself is excluded even if it matches the include patterns

`[files].exclude_patterns` removes files from indexing.

- Default: `[]`
- Excludes are also matched relative to the managed root

Example:

```toml
[files]
include_patterns = ["**/*.md", "**/*.png", "**/*.pdf", "**/*.yaml"]
exclude_patterns = ["**/*.tmp"]
```

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

- Default: `[".git", ".cache", "__pycache__"]`

`[watch].ignored_suffixes` lists filename suffixes the watcher ignores.

- Default: `["~", ".swp", ".tmp", ".bak"]`

## `[template].include_*`

These booleans control which optional sections appear in generated README templates.

- `[template].include_ownership`
- `[template].include_does_not_belong`
- `[template].include_related_docs`
- `[template].include_notes`

All four default to `true`.

## Custom Index Example

This config keeps the legacy `!README.md` folder index filename:

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

## Related Files

- `doc_ledger/config.py`
- `tests/test_config.py`
- `tests/test_public_config_end_to_end.py`
