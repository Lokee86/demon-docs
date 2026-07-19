# Markdown Link Reconciliation

Demon Docs maintains a repository-scoped graph of local Markdown links. This is a focused link graph for validation and path repair; it is not the later repository, code, symbol, or agent-context graph.

## Scope

Markdown source files and repository-local targets are scanned throughout the Demon Docs repository root, subject to `.docignore` and the permanent traversal exclusions. A link to an ignored repository path is left outside the link graph. Explicit targets outside the repository are not governed by the repository's `.docignore`.

Local targets may be:

- Markdown files;
- images, PDFs, archives, source files, and other non-Markdown files;
- directories;
- relative paths that resolve outside the repository;
- absolute filesystem paths; or
- `file://` URLs.

Web URLs and other non-filesystem schemes are not part of the local link graph.

Demon Docs only rewrites Markdown source files inside the repository. A target outside the repository can be checked and used as reconciliation evidence, but the external target itself is never modified.

## Supported Markdown Forms

The link scanner handles:

- inline links such as `[Guide](guide.md)`;
- images such as `![Diagram](assets/diagram.png)`;
- angle-wrapped destinations such as `[File](<files/a b.pdf>)`; and
- reference definitions such as `[guide]: docs/guide.md`.

Link-like text inside fenced code blocks and inline code spans is ignored. Heading fragments and query strings are preserved when a path is rewritten. Heading-anchor existence is not yet validated. HTML links, wiki-link syntax, and undefined reference labels are outside this first implementation.

## Persistent State

`.ddocs/` is a private Demon Docs repository, independent of the project's `.git/`. It uses go-git object, tree, reference, and filesystem-storage plumbing internally, but exposes no staging, branch, merge, commit-history, or manual repository workflow.

State is stored as deterministic records for file identities, current paths, Markdown sources and outgoing links, incoming-link groups, fingerprints, and pending generated writes. Record names are distributed across 16 content-addressed shards. A state reference atomically publishes the new root tree after all affected shard objects exist.

A single-file change rewrites only its affected shard or shards; unchanged objects and root entries are reused. The old `.ddocs/files.json` and `.ddocs/links.json` manifests are read only for migration and are removed after the first successful repository-backed publication.

The state is implementation-owned and schema-versioned. Source files are not modified to embed Demon Docs IDs.

## First Scan

The first link-enabled `fix` or `watch` pass establishes the baseline state and does not repair links. Existing broken links are reported.

`check -l` remains read-only. When link state has not been initialized, it reports that initialization is required and exits non-zero.

After the baseline exists, later passes can repair links using recorded identity and current filesystem evidence.

## Reconciliation Evidence

Demon Docs prefers deterministic evidence in this order:

1. the previous target file ID still resolves to a present file;
2. the target remains at the recorded path, including a case-only correction;
3. an exact, unique content fingerprint identifies a moved file;
4. a unique filename candidate exists inside the repository; or
5. a bounded search near a missing external target finds a unique candidate.

A unique candidate can be rewritten automatically. Multiple candidates are recorded and reported, and the source link remains unchanged for the user to resolve.

Relative links remain relative. Absolute filesystem links remain absolute. Link labels, titles, query strings, fragments, angle wrapping, and the source file's newline style are preserved; only the filesystem path is replaced.

## External Edits and Generated Rewrites

User-authored Markdown changes and Demon Docs-generated repairs follow separate paths.

For an external edit, Demon Docs fingerprints the changed source, parses its current Markdown, compares the resulting outgoing links with the stored source record, and replaces that source's graph edges.

For a known target move, Demon Docs queries stored incoming links by target identity, calculates exact destination replacements from the existing link records, and constructs a generated rewrite without first treating the result as a user edit. Each generated rewrite records the source file ID, expected old and new content hashes, affected link IDs, and old and new destinations.

Before writing, every source must still match its expected old hash. Writes use a same-directory temporary file and atomic replacement. The known graph mutation is then published directly. Reparsing the rewritten source is limited to verifying the expected links and refreshing byte offsets, line numbers, and fingerprints.

If a source changed concurrently, the generated rewrite aborts without overwriting the user's content. The next reconciliation processes that source through the external-edit path.

Unchanged files reuse stored fingerprints when path, size, and modification time agree. Current benchmarks cover initial indexing and a single-file incremental update so storage and scanning regressions remain visible.

## Commands and Feature Selection

With no selector flags, index and link reconciliation both run:

```bash
ddocs check
ddocs fix
ddocs watch
```

Run only one subsystem with either the short or long selector:

```bash
ddocs check -d
ddocs check --docs
ddocs check -l
ddocs check --links
ddocs check -r
ddocs check --reverse

ddocs fix -d
ddocs fix -l
ddocs fix -r

ddocs watch -d
ddocs watch -l
ddocs watch -r
```

Supplying selectors runs only those systems. Without selectors, documentation indexes and links run; reverse indexes also run when reverse roots are configured or supplied.

`check` reports pending rewrites, broken links, ambiguous links, and missing baseline state without modifying files. `fix` applies repository-contained source rewrites and saves the resulting state. `watch` uses the same reconciliation path automatically after relevant filesystem events.

When links are enabled, watch mode observes the repository root because moves of non-Markdown targets can require Markdown updates. It also watches the nearest existing parent directories of explicitly linked external targets, so an external rename or removal can trigger the same bounded reconciliation attempt. Documentation-only watch mode remains scoped to the configured docs root. Reverse-only watch mode remains scoped to configured or supplied reverse roots.

## Related Files

- `internal/links/`
- `internal/app/app.go`
- `internal/watch/watch.go`
- `internal/reconcile/reconcile.go`
