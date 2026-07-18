# Reconciliation Model

Demon Docs keeps the docs tree in a predictable shape by scanning folders, reading existing managed index blocks, and planning the smallest set of file updates needed to bring the tree back into sync.

## Scan Model

The scanner starts from the configured managed root and builds a tree of folders.

- The managed root is the folder Demon Docs owns, such as `docs/` by default.
- Normal folders are folders that can have their own index file.
- Draft folders, also called stub folders in the implementation, are the configured draft folder name such as `stubs/` by default.
- Direct files are indexed files that live directly inside a normal folder.
- Stub files are indexed files that live directly inside the draft folder for a normal folder.
- Direct folders are child folders of a normal folder, excluding the draft folder itself.
- Draft folders do not get their own index file.

The scan model is descriptive only. It records what exists on disk and where Demon Docs should look for managed content.

## README Index Behavior

Demon Docs treats README files as structured documents with managed sections.

- Managed blocks are wrapped in HTML comment markers.
- The managed sections are Direct Files, Stub Files, and Direct Folders.
- Human-authored content outside the managed markers is preserved.
- Existing managed entries are parsed from those marker blocks before reconciliation rewrites them.
- The default index filename is `README.md`, and `index_file = "!README.md"` keeps the legacy filename.
- Folder index files get `Parent index` links by default.
- Indexed files do not get `Parent index` links unless `indexed_files = true` is set.

If a README already has the expected managed sections, Demon Docs updates only the content inside those managed blocks.

Goldmark determines which headings and HTML comments are Markdown structure. Heading- and marker-like text inside fenced code blocks is code content and is never treated as a managed section. Parent-link-shaped lines inside fenced code are likewise examples rather than editable parent links. This is an intentional compatibility correction: fenced examples are not treated as real headings or managed sections.

## Missing README Creation

During reconciliation, Demon Docs creates missing index files where they belong.

- Normal folders get an index file if one is missing.
- The root folder gets an index file if one is missing.
- Draft folders do not get an index file.

The generated README template includes the managed sections so reconciliation can fill them in on the first pass.

## Parent Index Behavior

Demon Docs maintains parent index lines according to the configured parent-link toggles.

- The root index file has no parent index line.
- Child folder index files point to the parent folder using `../<index file>`.
- Normal docs point to their folder index using `./<index file>` when `indexed_files = true`.
- Stub docs point to the owning parent folder index using `../<index file>` when `indexed_files = true`.
- `folder_indexes = false` disables parent links in child folder indexes.
- `indexed_files = false` disables parent links in indexed files.

The parent index line is only written for file types that are configured as editable for parent links.

Parent-link insertion, replacement, and removal preserve whether the source document ended with a newline. This is an intentional compatibility guarantee for source preservation.

## Entry Preservation

Reconciliation prefers to preserve stable, existing index content when the target still belongs in the same place.

- Stable entries keep their existing descriptions.
- Graduating a stub file into a normal file removes a leading `Stub:` prefix when present.
- Moving a canonical file into the draft folder adds a `Stub:` prefix when needed.
- Unambiguous cross-folder file and folder moves preserve descriptions.
- Stale entries are removed from managed blocks and reported as reconciliation messages.

This preservation is intentionally narrow. Demon Docs matches by the current filesystem model and existing managed entries; it does not try to guess every historical rename pattern.

## Safety Boundaries

Demon Docs is a reconciliation tool, not a semantic documentation author.

- It does not decide which folder should own a topic.
- It does not rewrite arbitrary body links inside doc content.
- It does not edit non-editable file types even if they are indexed.
- `check` reports pending reconciliation, but it does not inspect git status.

Those boundaries keep the tool predictable and keep hand-authored prose under human control.

## Related Files

- `internal/scan/scan.go`
- `internal/markdown/markdown.go`
- `internal/reconcile/reconcile.go`
- `internal/model/model.go`
