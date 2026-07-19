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

Link state lives under the repository-owned `.ddocs/` directory:

```text
.ddocs/files.json
.ddocs/links.json
```

`files.json` assigns stable internal IDs to tracked files and records path history and content fingerprints. `links.json` records source locations, target paths, target file IDs when known, status, and ambiguous candidates.

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

## Commands and Feature Selection

With no selector flags, index and link reconciliation both run:

```bash
ddocs check
ddocs fix
ddocs watch
```

Run only one subsystem with either the short or long selector:

```bash
ddocs check -i
ddocs check --indexes
ddocs check -l
ddocs check --links

ddocs fix -i
ddocs fix -l

ddocs watch -i
ddocs watch -l
```

Supplying both selectors runs both systems, the same as supplying neither.

`check` reports pending rewrites, broken links, ambiguous links, and missing baseline state without modifying files. `fix` applies repository-contained source rewrites and saves the resulting state. `watch` uses the same reconciliation path automatically after relevant filesystem events.

When links are enabled, watch mode observes the repository root because moves of non-Markdown targets can require Markdown updates. It also watches the nearest existing parent directories of explicitly linked external targets, so an external rename or removal can trigger the same bounded reconciliation attempt. Index-only watch mode remains scoped to the configured docs root.

## Related Files

- `internal/links/`
- `internal/app/app.go`
- `internal/watch/watch.go`
- `internal/reconcile/reconcile.go`
