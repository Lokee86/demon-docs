# Demon Docs adoption demo

This is a deterministic, documentation-only fixture for a 10–15 minute product walkthrough. It contains no nested Git repository and no initialized `.ddocs` state.

The committed fixture is excluded from the Demon Docs project's own scans through `.docignore`, but it remains tracked by Git. Reset scripts copy it to a clean sibling workspace so `ddocs init` can establish an independent repository boundary.

## Reset the workspace

From the Demon Docs checkout:

```bash
bash tutorial/adoption-demo/reset-demo.sh
cd ../demon-docs-adoption-demo
```

PowerShell:

```powershell
.\tutorial\adoption-demo\reset-demo.ps1
Set-Location ..\demon-docs-adoption-demo
```

Both scripts replace the target completely. An alternate target path may be passed as the first argument.

## Intended starting condition

The fixture contains roughly thirty Markdown documents across four levels. It deliberately includes:

- no folder indexes;
- missing and incomplete YAML frontmatter;
- weak handwritten and duplicated document IDs;
- unknown frontmatter fields;
- missing and disordered schema sections;
- recognized heading aliases;
- one useful unknown section requiring an explicit exception;
- duplicate list sections suitable for an explicit merge;
- ordinary Markdown links, image links, wiki links, aliases, and fragments;
- one ambiguous bare wiki link;
- one meaningful orphan document;
- one ignored private-notes folder;
- one service area requiring a folder move and several file renames.

The bundled image was visually screened before inclusion. It is a safe pixel-art spacecraft asset used to demonstrate image-link and wiki-embed updates.

## Walkthrough outline

Install and verify the commands:

```bash
go install ./cmd/ddocs
go install ./cmd/demon
ddocs --version
```

Initialize the copied workspace:

```bash
ddocs init --root docs
ddocs status
ddocs config paths
```

Set `frontmatter.default_author` in `.ddocs/config.toml` to `Astra Operations`, and add `default = "TODO"` under `[frontmatter.fields.summary]`. These are the only fixture-specific policy values. Then inspect and apply the shipped documentation schemas:

```bash
ddocs check --docs
ddocs fix --docs
```

Most issues are repaired deterministically. Resolve the two authored structural decisions:

```bash
ddocs format ignore --heading "Rollout Checklist" docs/guides/deployment.md
ddocs format merge --heading "Responsibilities" docs/old-system/worker-notes.md
ddocs fix --docs
```

Establish link state and inspect the remaining human decisions:

```bash
ddocs fix --links
ddocs suggestions
ddocs check --links
```

Resolve the ambiguous `[[overview]]` suggestion by selecting the intended candidate. The orphan report should identify `docs/notes/launch-retrospective.md`; add one meaningful authored link to it from `docs/home.md`.

Preview and perform the service reorganization:

```bash
ddocs mv --dry-run docs/old-system docs/services
ddocs mv docs/old-system docs/services
ddocs mv docs/services/api-notes.md docs/services/api-service.md
ddocs mv docs/services/worker-notes.md docs/services/worker-service.md
ddocs mv docs/services/storage/storage-notes.md docs/services/storage/storage-service.md
ddocs mv docs/services/assets/system-overview.jpg docs/services/assets/service-overview.jpg
```

Inspect Markdown links and wiki links after the move. The fixture includes a labeled wiki link, a bare wiki link, a Markdown image, a wiki embed, and a fragmented storage link that should remain semantically intact.

Create one correctly structured service document:

```bash
ddocs new service docs/services/scheduler-service.md
```

Link it from an appropriate service document, then inspect generated history:

```bash
ddocs changes
ddocs changes log
```

Finish with a bounded watcher pass and convergence checks:

```bash
ddocs watch --once
ddocs fix
ddocs check
ddocs fix
```

The final `fix` should produce no changes. Codemap generation and reverse code indexes are intentionally outside this demo.
