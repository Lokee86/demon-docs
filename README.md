# doc-ledger

doc-ledger is a repo-local documentation index maintenance tool.
It scans a configurable docs root and reconciles folder index files, direct file entries, draft/stub file entries, direct folder entries, and `Parent index` links.

## What doc-ledger does

- Keeps folder index files in sync with the filesystem.
- Adds or updates direct file entries for indexed files.
- Adds or updates stub/draft entries for files inside `stubs/`-style draft folders.
- Adds or updates direct folder entries for child folders.
- Keeps `Parent index` links aligned with the current folder layout.
- Preserves existing descriptions when a target is still present and recognizable.

## Quick Start

The default docs root is `docs`.
The default index file is `README.md`.
The default stub/draft folder is `stubs/`.

Run the tool from the repo root:

```bash
python3 main.py fix --root docs
python3 main.py check --root docs
python3 main.py watch --root docs
python3 main.py watch --root docs --once
```

Config files are optional:

```bash
python3 main.py fix --config .doc-ledger.toml
python3 main.py check --config .doc-ledger.toml
```

## Commands

- `fix`: Reconciles the docs tree and writes any needed updates.
- `check`: Reconciles the docs tree without writing files. This is the verification gate.
- `watch`: Watches the docs tree and reruns reconciliation when relevant files change.
- `watch --once`: Runs one reconciliation pass and exits.

## What It Edits

- Folder index files in the managed docs root.
- Direct file entries in managed README files.
- Stub/draft file entries in managed README files.
- Direct folder entries in managed README files.
- `Parent index` lines in editable indexed files.

## What It Does Not Edit

- Hand-authored content outside the managed sections.
- `stubs/` folders themselves as index targets.
- Non-doc files that are not included by configuration.
- Non-editable indexed files when parent-link editing is disabled by extension.
- `.gitignore`-protected Python cache files and other generated cache artifacts.

## Managed README Sections

doc-ledger manages three section ids:

- `files`
- `stubs`
- `folders`

Those sections render as HTML-comment-controlled blocks in each managed README.
The default headings are:

- `Direct Files`
- `Stub Files`
- `Direct Folders`

Legacy headings such as `Top-Level Files` and `Top-Level Folders` are migrated forward when present.

## Parent Index Links

Each editable indexed file gets a `Parent index` line that points back to the owning README.

- Normal files use `./README.md`.
- Files inside the draft/stub folder use `../README.md`.
- Child folder README files use `../README.md`.
- The label is configurable, but the default label is `Parent index`.

## Stubs/Drafts

- The default draft folder is `stubs/`.
- Draft folders do not get their own index file.
- Draft file descriptions are prefixed with `Stub: ` by default.
- Draft entries stay in the parent README’s stub section.
- When a stub file graduates into the canonical folder, doc-ledger keeps the description when it can.

## Watch Mode

Watch mode is convenience for local automation, not the correctness gate.
`check` remains the command to rely on in scripts and CI.

- Watch startup lines include a timestamp and the running PID.
- Watch output shows reconciliation summaries.
- Relevant filesystem changes trigger the same reconciliation path as `fix`.
- The watcher ignores common editor junk and configured ignored paths.

## Configuration

doc-ledger reads TOML configuration from `.doc-ledger.toml` or `doc-ledger.toml`.
When no config file is supplied, the built-in defaults apply.

Important settings include:

- `root`: docs root directory, default `docs`
- `index_file`: folder index file name, default `README.md`
- `markers.prefix`: managed marker prefix, default `doc-ledger`
- `parent_link.label`: parent-link label, default `Parent index`
- `sections.files.heading`, `sections.stubs.heading`, `sections.folders.heading`
- `aliases.files`, `aliases.folders`
- `drafts.folder`: draft folder name, default `stubs`
- `drafts.description_prefix`: default `Stub: `
- `files.include_patterns` and `files.exclude_patterns`
- `editable.parent_index_extensions`
- `descriptions.file_template` and `descriptions.folder_template`
- `watch.debounce_seconds`, `watch.ignored_dirs`, and `watch.ignored_suffixes`
- `template.include_ownership`, `template.include_does_not_belong`, `template.include_related_docs`, `template.include_notes`

## Testing

Run the doc-ledger tests from the repo root:

```bash
python3 -m pytest tests
```

The repo also keeps Python cache artifacts out of commits with `.gitignore`, and the hygiene test fails if the working tree contains `__pycache__`, `.pyc`, or `.pyo` files.

## More Docs

- [Docs tree entry point](docs/README.md)
- [docs/configuration.md](docs/configuration.md)
- [docs/reconciliation-model.md](docs/reconciliation-model.md)
- [docs/watcher-and-automation.md](docs/watcher-and-automation.md)
- [docs/testing-and-fixtures.md](docs/testing-and-fixtures.md)

The docs tree uses `docs/README.md` as its index by default.
Projects that want the legacy filename can opt in with `index_file = "!README.md"` in `.doc-ledger.toml`.
