---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7a71-8fff-be364ee93444
document_type: general
policy_exempt: false
summary: This document defines Demon Docs configuration selection, defaults, supported keys, repository scope behavior, ignore rules, and complete configuration examples.
---
# Demon Docs Configuration

Parent index: [Reference](./INDEX.md)

## Purpose

This document defines Demon Docs configuration selection, defaults, supported keys, repository scope behavior, ignore rules, and complete configuration examples.

## Overview

Demon Docs is configured with TOML. The primary config model lives in `internal/config/config.go` and is exercised by Go package tests and the Go CLI fixture regression matrix.

CLI help is available with `ddocs --help`, and each subcommand also supports `--help`.
Top-level version output is available with `ddocs -v` or `ddocs --version`.
Repository initialization is optional for ordinary reconciliation. Without it, `check`, `fix`, and foreground `watch` use a standalone scope based on the selected or built-in docs root.

Initialize from the repository root when repository-local configuration, a repository-wide boundary, starter schemas, feature toggles, linked-worktree daemon bootstrap, or the detached demon is needed:

```bash
ddocs init --root docs/
```

This creates `.ddocs/config.toml`. Commands run anywhere below that directory search upward for `.ddocs/`, treat its parent as the repository root, and resolve the configured docs root from there. `ddocs status` reports only this initialized-repository scope.

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
- `[index].enabled`
- `[links].enabled`
- `[reverse_index].roots`
- `[reverse_index].folders` as a compatibility alias
- `[codemap].headings`
- `[codemap].remove_undiscovered_links`
- `[codemap].remove_low_score_links`
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
- `[demon].run`
- `[review].undo_depth`
- `[review].undo_max_age_days`
- `[frontmatter].enabled`
- `[frontmatter].default_format`
- `[frontmatter].allowed_formats`
- `[frontmatter].default_author`
- `[frontmatter].unknown_fields`
- `[format].enabled`
- `[format].schema_dir`
- `[format].document_schema_dir`
- `[format].default_schema`
- `[format].invalidation_similarity`
- `[[format.path_rules]].pattern`
- `[[format.path_rules]].schema`
- `[frontmatter.fields.<name>].type`
- `[frontmatter.fields.<name>].required`
- `[frontmatter.fields.<name>].immutable`
- `[frontmatter.fields.<name>].generated`
- `[frontmatter.fields.<name>].default`
- `[frontmatter.fields.<name>].default_from`
- `[[frontmatter.rules]].when_field`
- `[[frontmatter.rules]].equals`
- `[[frontmatter.rules]].require`

## Repository Feature Toggles

Initialized repositories enable index management and automatic link maintenance by default:

```toml
[index]
enabled = true

[links]
enabled = true
```

Use repository-local commands to change either setting from anywhere inside the repository:

```bash
ddocs index enable
ddocs index disable
ddocs index status

ddocs links enable
ddocs links disable
ddocs links status
```

`--true` and `--false` are accepted aliases for `enable` and `disable`. These commands always update the nearest initialized repository's `.ddocs/config.toml`; they do not modify global configuration. When that repository has a running demon, the command requests a clean restart so the watcher reloads the changed settings.

Disabling `[index].enabled` suspends folder-index creation, insertion, repair, and index-specific tracking. Existing index files are not ignored or given special treatment. They remain ordinary document files and can still participate in link scanning, codemap extraction, and other document behavior.

Disabling `[links].enabled` suspends automatic link rewrites and user-visible link diagnostics. Demon Docs continues updating its private file identities, path history, and link graph in `.ddocs/`. Re-enabling link maintenance therefore resumes from retained state instead of rebuilding tracking from scratch.

Selectors do not override a disabled repository feature. For example, `ddocs fix -d` does not create indexes while indexing is disabled, and `ddocs fix -l` refreshes internal link state without rewriting documents while link maintenance is disabled. Configured frontmatter remains a separate docs-selected subsystem, so `-d` may still validate or repair frontmatter when indexing is disabled.

## Frontmatter Schema

Initialized repositories include a strict, project-editable frontmatter schema. Legacy configs without a `[frontmatter]` section keep frontmatter enforcement disabled.

```toml
[frontmatter]
enabled = true
default_format = "yaml"
allowed_formats = ["yaml", "toml"]
default_author = ""
unknown_fields = "remove"
```

Frontmatter applies to non-ignored `.md` files beneath the configured docs root. It is independent of `[files].include_patterns` and `[files].exclude_patterns`, which control folder-index membership rather than document policy. Generated folder indexes are included because index reconciliation completes before frontmatter reconciliation in the same `fix` or watch pass. When `document_type` is missing and format enforcement is enabled, repair uses the format path-rule/default selection rather than blindly writing the generic frontmatter default.

YAML blocks use `---`; TOML blocks use `+++`. A recognized existing block keeps its format. `default_format` is used only when a block is inserted. `allowed_formats` rejects a recognized but disallowed format instead of converting it. Malformed, unclosed, duplicate-key, or multiple leading blocks are reported without destructive rewriting.

Supported field types are:

- `string`
- `boolean` or `bool`
- `integer`
- `number`
- `string_list` or the compatibility alias `list`
- `uuid`
- `date` in `YYYY-MM-DD` form

Each field may be required, immutable, or supplied by one configured source:

```toml
[frontmatter.fields.document_id]
type = "uuid"
required = true
immutable = true
generated = true

[frontmatter.fields.author]
type = "string"
required = true
default_from = "frontmatter.default_author"

[frontmatter.fields.document_type]
type = "string"
required = true
default = "general"
```

`default`, `default_from`, and `generated` are mutually exclusive. The only current `default_from` source is `frontmatter.default_author`. Generated values are supported for `uuid` and `date`: UUIDs are created once and then preserved; generated dates use the current local calendar date. Immutable values are recorded in Demon Docs private state and restored by `fix` when later edits disagree with that known value. Existing valid mutable values are never overwritten. Existing invalid mutable values remain authored content and are reported for manual correction.

Unknown-field handling is configured with `unknown_fields`:

- `remove` is the default. `check` fails; `fix` deletes only the unknown fields.
- `warn` preserves the fields and emits warnings without failing solely because of them.
- `ignore` preserves the fields silently.

Conditional requirements use repeated rule tables:

```toml
[[frontmatter.rules]]
when_field = "policy_exempt"
equals = true
require = "policy_exempt_reason"
```

`check` never writes frontmatter or immutable state. `fix` applies deterministic repairs, then returns non-zero when required or invalid values still need authored input. The starter schema intentionally leaves `default_author` blank and `summary` without a default, so ordinary authored documents must configure those values, relax the schema, or author them explicitly. Demon Docs-owned generated folder indexes receive deterministic fallback author and summary values when no configured source exists, allowing a fresh initialized repository to converge without weakening policy for ordinary documents.

## Document Schemas And Body Format

Initialized repositories enable document-body format enforcement and write human-editable starter schemas:

```toml
[format]
enabled = true
schema_dir = ".ddocs/schemas"
document_schema_dir = ".ddocs/document-schemas"
default_schema = "general"
invalidation_similarity = 0.5
```

Shared schemas are TOML files named after `document_type`. Metadata selects the schema first. Path fallbacks are evaluated only when metadata is absent. During frontmatter repair, that same selection supplies a missing `document_type`, so later passes retain the selected path-based schema:

```toml
[[format.path_rules]]
pattern = "docs/services/**"
schema = "service"

[[format.path_rules]]
pattern = "docs/planning/**"
schema = "planning"
```

`schema_dir` stores human-authored shared policy. `document_schema_dir` stores generated but human-editable per-document exceptions keyed by immutable `document_id`. `default_schema` is used after metadata and path fallbacks. An empty default disables enforcement for unmatched documents.

`invalidation_similarity` controls when a changed shared schema invalidates document-specific exceptions. Similarity is measured cumulatively from the exact shared-schema fingerprint recorded when the exception was accepted, comparing canonical definitions by stable section ID against the larger schema's section count. The default `0.5` requires re-confirmation when similarity falls below 50 percent. `0` disables automatic invalidation.

Use `ddocs new DOCUMENT_TYPE PATH` to create a document from the corresponding schema. Use `ddocs check --format` and `ddocs fix --format` to operate on Markdown body structure independently from frontmatter. See [Document Schemas And Format Enforcement](document-schemas.md).

## Repository Demon

Initialized repositories permit the self-managing watcher by default:

```toml
[demon]
run = true
```

Use `ddocs demon run --false` to persistently disable it, or
`ddocs demon run --true` to re-enable it. `ddocs demon --status` reports the
repository-local owner and active shell/agent feeders; `ddocs demon --logs`
prints the bounded repository-local log. Demon runtime state is stored under
`.ddocs/runtime/` and is not part of document traversal or mutable object
storage.

The demon is an operational convenience around the existing watcher, not a
correctness dependency. `check`, `fix`, and foreground `watch` remain available
when it is disabled. Shell hooks use `shell` feeders; MCP and native host
adapters can use the host-neutral `agent` feeder lifecycle without moving host
logic into Demon Docs core. See [Repository Demon](../operations/repository-demon.md).

## Review and Undo

```toml
[review]
undo_depth = 100
undo_max_age_days = 30
```

`undo_depth` limits how many recent non-undo applied changes remain eligible for reversal. `0` disables undo and `-1` removes the depth limit. `undo_max_age_days` limits eligibility by age; `0` removes the age limit. These settings do not delete audit history.

Demon Docs supports undo by reconciliation run, one file change, or one repair within a file change. Every undo remains hash-guarded and refuses to overwrite later edits. See [Review Ledger](../architecture/review-ledger.md) and [Reviewing Suggestions and Changes](../guides/reviewing-suggestions-and-changes.md).

## Codemap Configuration

```toml
[codemap]
headings = ["Code map", "Codemap", "Code or source map", "Code and test map"]
remove_undiscovered_links = false
remove_low_score_links = false
```

`headings` identifies existing codemap sections. Matching is case-insensitive. The production `codemap fix`, `check`, and `inspect` commands accept repeated `--heading TEXT` overrides.

Existing links are retained by default even when the algorithm does not rediscover them or ranks them below the hard-link tier. `remove_undiscovered_links` permits removal when a hidden-link evaluation cannot recover an existing resolved target. `remove_low_score_links` permits removal when that evaluation recovers only a context-tier relationship. Both are intentionally `false` by default.

New missing links from both `hard_link` and `context` tiers are added automatically by explicit codemap execution. The codemap operation consults the shared review-decision store before writing each addition, so an unchanged declined recommendation remains suppressed and materially changed evidence may be reconsidered under the existing fingerprint policy. Tier remains visible in inspection and removal policy; it is not a per-run approval gate.

The research-oriented `codemap export` command additionally supports `--target-base repository|document`, repeated `--target-root PATH`, and `--output PATH`. `[reverse_index].roots` remains separate configuration for reverse indexes. Codemap execution is never invoked by ordinary reconciliation, foreground watch, or the repository daemon.

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

- `-d` / `--docs` selects documentation-folder indexes, configured frontmatter enforcement, and document-body format enforcement.
- `--frontmatter` selects frontmatter enforcement only.
- `--format` selects document-body format enforcement only.
- `-l` / `--links` selects Markdown link reconciliation.
- `-r` / `--reverse` selects code-folder reverse indexes.
- `-i` / `--indexes` selects documentation indexes only; `-d` / `--docs` additionally selects frontmatter and document-body format.
- When any selector is supplied, only selected systems run.
- Without selectors, indexes, configured frontmatter, document-body format, and links run; reverse indexes also run when roots are configured or supplied with `--reverse-root`.
- The selectors apply to `check`, `fix`, and `watch`.

Examples include:

```bash
ddocs check -d
ddocs check -l
ddocs check -r
ddocs fix --docs
ddocs fix --links
ddocs fix --reverse
ddocs watch -d
ddocs watch -l
ddocs watch -r
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

The following block is the initialized-repository starter configuration written by `ddocs init`; it is not identical to no-config standalone behavior. With built-in standalone defaults, indexes and links are enabled, the docs root is `docs`, frontmatter and body-format enforcement are disabled, reverse roots are empty, and the detached demon is unavailable because no initialized repository exists:

```toml
docs_root = "docs"
index_file = "INDEX.md"

[index]
enabled = true

[links]
enabled = true

[format]
enabled = true
schema_dir = ".ddocs/schemas"
document_schema_dir = ".ddocs/document-schemas"
default_schema = "general"
invalidation_similarity = 0.5

[[format.path_rules]]
pattern = "**/INDEX.md"
schema = "index"

[[format.path_rules]]
pattern = "**/README.md"
schema = "index"

[[format.path_rules]]
pattern = "**/!README.md"
schema = "index"

[[format.path_rules]]
pattern = "**/!INDEX.md"
schema = "index"

[[format.path_rules]]
pattern = "**/planning/**"
schema = "planning"

[[format.path_rules]]
pattern = "**/services/**"
schema = "service"

[reverse_index]
roots = []

[codemap]
headings = ["Code map", "Codemap", "Code or source map", "Code and test map"]
remove_undiscovered_links = false
remove_low_score_links = false

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

[demon]
run = true

[review]
undo_depth = 100
undo_max_age_days = 30

[frontmatter]
enabled = true
default_format = "yaml"
allowed_formats = ["yaml", "toml"]
default_author = ""
unknown_fields = "remove"

[frontmatter.fields.document_id]
type = "uuid"
required = true
immutable = true
generated = true

[frontmatter.fields.author]
type = "string"
required = true
default_from = "frontmatter.default_author"

[frontmatter.fields.document_type]
type = "string"
required = true
default = "general"

[frontmatter.fields.created]
type = "date"
required = true
immutable = true
generated = true

[frontmatter.fields.summary]
type = "string"
required = true

[frontmatter.fields.policy_exempt]
type = "boolean"
default = false

[frontmatter.fields.policy_exempt_reason]
type = "string"

[[frontmatter.rules]]
when_field = "policy_exempt"
equals = true
require = "policy_exempt_reason"

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

## `[index].enabled` and `[links].enabled`

`[index].enabled` controls automatic folder-index management. It defaults to `true`. When disabled, Demon Docs does not create, insert, repair, or specially track folder indexes. A file whose name matches `index_file` remains visible as an ordinary document.

`[links].enabled` controls automatic link maintenance. It defaults to `true`. When disabled, document contents are not rewritten, but persistent internal link tracking continues and is published to `.ddocs/`.

The canonical mutation commands are `ddocs index enable|disable` and `ddocs links enable|disable`. `status` reports the current repository value without changing it.

## `index_file`

`index_file` sets the folder index filename.

- Default: `INDEX.md`
- Example custom values: `README.md`, `!README.md`, or `!INDEX.md`
- Folder-index links and generated folder index paths follow this name
- Existing repositories keep their configured filename; the new default applies only when no override is selected.

Projects that want another convention should set `index_file` explicitly in config.

## `[reverse_index].roots`

`[reverse_index].roots` selects the repository folders where code-folder reverse indexes may be generated. There is no repository-wide default; an empty list requires `--reverse-root` whenever `-r` / `--reverse` is selected.

Configured roots are resolved relative to the repository root and traversed recursively. Only folders beneath those roots can receive reverse-index managed sections. Overlapping roots are collapsed to the broadest selected root.

```toml
[reverse_index]
roots = ["client", "services/game-server", "services/player-data"]
```

`--reverse-root PATH` overrides configured roots for one `check`, `fix`, or `watch` invocation and may be repeated. Relative paths resolve from the current working directory; absolute paths are accepted when they remain inside the repository.

```bash
ddocs check -r --reverse-root services/game-server
ddocs fix -r --reverse-root client --reverse-root services/player-data
cd services/game-server
ddocs watch -r --once --reverse-root .
```

`[reverse_index].folders` remains accepted as an alias for older experimental configs, but `roots` is the canonical key.

## `[codemap].headings` and removal policy

`[codemap].headings` defines the Markdown section headings recognized as codemaps. Matching is case-insensitive and ignores trailing Markdown heading markers.

```toml
[codemap]
headings = ["Implementation map", "Source map"]
remove_undiscovered_links = false
remove_low_score_links = false
```

`ddocs codemaps fix|check|inspect --heading TEXT` replaces the configured headings for that explicit codemap operation. `--codemap-heading TEXT` remains the reverse-index reconciliation override.

When a matching section exists, it is processed regardless of the document's file-type schema. The full section is managed as one unified codemap; existing links and deterministic additions are not split into separate authored and generated subsections.

The public codemap command resolves the same metadata-first effective document schema used by body-format enforcement. When that schema requires a codemap section, the command creates it at the schema-defined position. A document whose schema does not declare a codemap section remains unchanged; the command never invents a placement outside schema policy.

Removal based on algorithm confidence is opt-in through the two boolean settings. Definitively broken paths remain governed by normal link maintenance.

## `[markers].prefix`

`[markers].prefix` sets the HTML comment prefix for managed sections.

- Default: `doc-ledger`
- Managed blocks use `files`, `stubs`, `folders`, and `codemap` section ids

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

These keys control the visible headings for managed folder-index sections.

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

An initialized repository uses `.docignore` at its repository root, beside `.ddocs/`, as the base ignore policy. A standalone scope instead uses `.docignore` at its resolved docs root. It excludes paths from index traversal, frontmatter enforcement, document-body format enforcement, repository Markdown link scanning, link-target inventory, and watch events. Reverse-index traversal additionally recognizes nested `.docignore` files beneath configured roots; each nested file applies Git-ignore rules relative to its containing directory.

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

Watch mode watches the repository root for the base `.docignore`. `ddocs watch -r` watches only the selected reverse roots plus their ancestor directories, detects nested `.docignore` changes, reloads the hierarchy, and adds watches for directories that become visible. Documentation-only watch mode otherwise remains scoped to the docs root; any mode that selects link tracking observes the repository root, including tracking-only operation while automatic link maintenance is disabled.

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

These booleans control which optional sections appear in generated folder-index templates.

- `[template].include_ownership`
- `[template].include_does_not_belong`
- `[template].include_related_docs`
- `[template].include_notes`

All four default to `true`.

## Folder Indexes Only

This config uses the default split behavior and keeps parent links in folder indexes only:

```toml
root = "notes"
index_file = "INDEX.md"

[parent_link]
folder_indexes = true
indexed_files = false
```

## File-Level Parent Links

This config keeps parent links in both folder indexes and indexed files:

```toml
root = "notes"
index_file = "INDEX.md"

[parent_link]
folder_indexes = true
indexed_files = true
```

## Disable Parent Links

This config disables parent links everywhere:

```toml
root = "notes"
index_file = "INDEX.md"

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
index_file = "INDEX.md"

[drafts]
folder = "_drafts"
description_prefix = "Draft: "
```

## Non-Markdown Indexing Example

This config indexes Markdown plus image, PDF, and YAML files, while only editing parent links in `md` and `mdx` files:

```toml
root = "notes"
index_file = "INDEX.md"

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

Markdown link reconciliation is controlled by `[links].enabled`. Its persistent, schema-versioned state is stored in the active scope's private `.ddocs/` object repository: beneath the docs root in standalone mode or beneath the repository root in initialized mode. Demon Docs uses internal go-git object and reference plumbing. Link state uses `refs/ddocs/state`; suggestion decisions and applied-change history use `refs/ddocs/review`. Neither ref creates commits in the user's normal Git history or exposes a user-facing Git workflow for private state.

The first link-enabled `fix` or `watch` pass establishes this baseline without repairing links. With link maintenance enabled, `check -l` is read-only and reports a missing link-state baseline rather than creating it. With link maintenance disabled, selected and default reconciliation passes may publish tracking-only state while leaving every document unchanged. Legacy `.ddocs/files.json` and `.ddocs/links.json` state is migrated on the next successful link-state publication.

## Code map

- `internal/config/config.go` — TOML model, defaults, compatibility aliases, and selection loading.
- `internal/config/config_test.go` — config parsing, precedence, and compatibility coverage.
- `internal/config/behavior_test.go` — user-visible configuration behavior.
- `internal/app/app.go` — config commands and one-shot CLI overrides.
- `internal/repository/scope.go` — repository-root and docs-root resolution boundaries.
- `internal/repository/repository.go` — repository discovery and initialization.
- `internal/frontmatter/` — YAML/TOML parsing, schema evaluation, immutable-value state, and docs-root repair planning.
- `internal/documentpolicy/` — TOML document schemas, Markdown body parsing, structure enforcement, creation, and document-specific exceptions.
- `internal/links/` — schema-versioned link state governed by repository scope rather than TOML keys.
- `internal/review/` — review history, decisions, blocks, and undo eligibility.

## Diagnostics and failure behavior

Use `ddocs config paths` to inspect selection and `ddocs config show` to inspect resolved values. Invalid roots, escaping paths, unsupported values, or unreadable configuration fail before broad repository mutation.

## Related docs

- [Front Matter Schemas](frontmatter.md)
- [Reference](INDEX.md)
- [Getting Started](../guides/getting-started.md)
- [CLI Reference](cli.md)
- [Document Schemas And Format Enforcement](document-schemas.md)
- [Managed Files and State](managed-files-and-state.md)
- [Review Ledger](../architecture/review-ledger.md)
- [Application Orchestration](../architecture/application-orchestration.md)
- [Codemap Managed Execution](../architecture/codemap-managed-execution.md)
- [Managing Codemaps](../guides/managing-codemaps.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

Repository-local `.ddocs/config.toml` is preferred for initialized repositories. Legacy local and global names remain compatibility inputs at lower precedence.
