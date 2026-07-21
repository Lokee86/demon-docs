# Demon Docs adoption walkthrough

> [!IMPORTANT]
> **Do not open `tutorial/`, `tutorial/adoption-demo/`, or `tutorial/adoption-demo/fixture/` as the Obsidian vault.**
>
> `tutorial/adoption-demo/fixture/` is a tracked, read-only source template containing synthetic Astra Relay documentation. It is not the working tutorial repository, it is not Demon Docs product documentation, and it is not Space Rocks documentation.
>
> Run the reset script first. It creates a separate disposable sibling directory named `demon-docs-adoption-demo`. **Open and edit only that sibling directory.** Running the reset script again deletes and recreates that sibling workspace from the tracked fixture.

This walkthrough adopts Demon Docs into a deliberately inconsistent documentation repository, repairs what can be decided safely, preserves authored decisions, reorganizes a service area without breaking references, and finishes by enabling automatic maintenance.

The fixture is deterministic and contains no nested Git repository or initialized `.ddocs` state. Its top-level `README.md` discloses every intentional starting problem and navigability shortcoming. The demonstration uses YAML frontmatter; TOML frontmatter is also supported but is outside this walkthrough.

Codemap generation, reverse code indexes, and source-code integration are intentionally excluded so the walkthrough stays focused on the core documentation-maintenance workflow.

## Source template versus working repository

```text
Demon Docs checkout/
└── tutorial/adoption-demo/fixture/   tracked source template; do not use as the vault

Sibling of the Demon Docs checkout/
└── demon-docs-adoption-demo/         generated disposable repository; use this as the vault
```

The reset scripts never repair or regenerate files in place under `tutorial/`. They copy the tracked source template into the sibling working repository. The scripts refuse targets located inside the Demon Docs checkout.

## 1. Install Demon Docs

From the Demon Docs checkout:

```bash
go install ./cmd/ddocs
go install ./cmd/demon
ddocs --version
```

## 2. Generate or reset the disposable demonstration repository

Run the script **from the Demon Docs checkout**:

```bash
bash tutorial/adoption-demo/reset-demo.sh
cd ../demon-docs-adoption-demo
```

PowerShell equivalent:

```powershell
.\tutorial\adoption-demo\reset-demo.ps1
Set-Location ..\demon-docs-adoption-demo
```

The script copies:

```text
tutorial/adoption-demo/fixture/
```

into this separate sibling directory:

```text
../demon-docs-adoption-demo/
```

The sibling directory is the actual tutorial repository. The script deletes and recreates it on every run, so edits made there are intentionally disposable. It does **not** modify the tracked source fixture inside the Demon Docs checkout.

Before continuing, confirm that the current directory ends with:

```text
demon-docs-adoption-demo
```

Open that exact directory as the Obsidian vault. Do not open the Demon Docs checkout, its `tutorial/` directory, or the tracked `fixture/` directory as the vault.

## 3. Review the disclosed starting condition

Open the repository's top-level `README.md`.

It provides a complete inventory of the fixture's intentional state:

- twenty-three documents with existing YAML frontmatter, ranging from valid to incomplete, weakly identified, duplicated, empty-valued, invalidly dated, and unknown-field-bearing;
- five documents with no frontmatter at all;
- two stale Markdown paths;
- one ambiguous wiki link and one orphaned document;
- an older service area that needs to be moved and renamed;
- eleven folders without local indexes, reducing navigability without constituting broken Markdown;
- 175 authored local link occurrences across Markdown links, wiki links, aliases, fragments, a reference-style link, an image link, and a wiki image embed;
- one ignored private-notes file that is intentionally outside Demon Docs management.

The individual files remain available for manual inspection, but the walkthrough now lets Demon Docs diagnose the repository itself.

## 4. Initialize without starting automatic maintenance

The daemon is enabled in the generated configuration by default. Disable it in the same command chain as initialization so it cannot repair the demonstration before the explicit steps are shown:

```bash
ddocs init --root docs && ddocs demon run --false
ddocs status
ddocs demon --status
ddocs config paths
```

Initialization creates the repository-local configuration, default schemas, and private state under `.ddocs/`. The authored documentation remains in its original condition.

## 5. Set the repository's policy defaults

Open `.ddocs/config.toml`.

Set the default author in the existing `[frontmatter]` section:

```toml
[frontmatter]
default_author = "Astra Operations"
```

Add a safe placeholder under the existing summary field definition:

```toml
[frontmatter.fields.summary]
default = "TODO"
```

These are the only fixture-specific policy values. The shipped `general`, `service`, `planning`, and `index` schemas remain unchanged.

## 6. Diagnose the documentation

```bash
ddocs check --docs
```

The report should independently identify the conditions disclosed in the fixture README, including:

- missing folder indexes;
- absent, invalid, duplicated, and unknown frontmatter fields;
- missing and disordered required sections;
- a duplicated authored section;
- an authored section that is not present in the shared schema.

The output is intentionally extensive. It demonstrates repository-wide diagnosis; it does not need to be read line by line.

## 7. Apply deterministic repairs

```bash
ddocs fix --docs
```

This pass generates local indexes, normalizes metadata, assigns stable document IDs, removes unknown policy fields, orders recognized sections, and creates required missing sections.

The command is expected to remain non-clean because two authored structural decisions cannot be made safely without instruction.

Inspect two representative results:

```bash
cat docs/INDEX.md
cat docs/old-system/api-notes.md
```

The first shows generated local navigation. The second shows metadata and service-schema normalization while preserving the document's authored content and references.

## 8. Resolve the two authored structural decisions

Preserve the deployment document's useful custom checklist by creating a document-specific schema exception:

```bash
ddocs format ignore \
  --heading "Rollout Checklist" \
  docs/guides/deployment.md
```

Merge the worker document's two authored responsibility lists:

```bash
ddocs format merge \
  --heading "Responsibilities" \
  docs/old-system/worker-notes.md
```

Finish and verify the documentation-policy pass:

```bash
ddocs fix --docs
ddocs check --docs
```

The documentation check should now pass.

## 9. Repair links and establish trusted history

Start with a read-only link check:

```bash
ddocs check --links
```

The initial report should include two genuine broken references:

- `docs/getting-started.md` points to `concepts/archive/configuration.md`, while the repository contains exactly one `configuration.md` at `concepts/configuration.md`;
- `docs/old-system/api-notes.md` points to `storage/archive/storage-notes.md#retention`, while the repository contains exactly one `storage-notes.md` at `storage/storage-notes.md#retention`.

The first repair pass records the repository's link baseline. The second can use the now-initialized state and the unique filename candidates to repair both stale paths deterministically:

```bash
ddocs fix --links
ddocs fix --links
ddocs suggestions docs/home.md
```

Scoping `suggestions` to `docs/home.md` keeps unrelated suggestion types out of the walkthrough output.

The remaining relationship issues should be:

- `docs/home.md` contains an ambiguous `[[overview|project overview]]` wiki link because two files are named `overview.md`;
- `docs/notes/launch-retrospective.md` is useful but has no incoming authored link.

Inspect the ambiguous suggestion using the identifier printed by the scoped `ddocs suggestions` command:

```bash
ddocs suggestions show <suggestion-id>
```

Select the intended target by path:

```bash
ddocs suggestions select <suggestion-id> docs/concepts/overview.md
```

Demon Docs can resolve the ambiguous target after that explicit decision. It should not invent a semantic relationship for the orphaned retrospective.

Add this authored entry under `Related docs` in `docs/home.md`:

```markdown
- [Launch retrospective](notes/launch-retrospective.md)
```

Verify that the adopted repository now has a clean link baseline:

```bash
ddocs check --links
```

Next, create a new broken-link condition with an ordinary filesystem rename while the daemon is still disabled:

```bash
mv \
  docs/guides/local-setup.md \
  docs/guides/workstation-setup.md
```

The authored references still point to `local-setup.md`. Confirm the failure without modifying anything:

```bash
ddocs check --links
```

Because the clean baseline recorded the document's prior identity and location, Demon Docs can now repair the static filesystem change deterministically:

```bash
ddocs fix --links
grep -R -n "workstation-setup" docs
ddocs check --links
```

The old path should be replaced across the affected documents and generated indexes before the final link check passes.

## 10. Preview the service reorganization

The service notes still live under `docs/old-system`. Preview the largest move before writing anything:

```bash
ddocs mv --dry-run docs/old-system docs/services
```

The preview reports the planned move and a broad set of affected Markdown files and rewritten links without mutating the repository.

## 11. Move and rename the service area

```bash
ddocs mv docs/old-system docs/services
ddocs mv docs/services/api-notes.md docs/services/api-service.md
ddocs mv docs/services/worker-notes.md docs/services/worker-service.md
ddocs mv \
  docs/services/storage/storage-notes.md \
  docs/services/storage/storage-service.md
ddocs mv \
  docs/services/assets/system-overview.jpg \
  docs/services/assets/service-overview.jpg
```

The service tree is referenced throughout the concepts, guides, troubleshooting, planning, stub, and service documents. The folder move and subsequent API, worker, storage, and image renames rewrite dozens of references across the documentation tree, exercising a substantial portion of the 175-link graph rather than one isolated source file.

Inspect `docs/services/api-service.md` after the move. In one document, the fixture verifies preservation and rewriting of:

- a labeled wiki link;
- a Markdown link with a heading fragment;
- a Markdown image reference;
- a wiki image embed.

Then search the broader graph for rewritten service paths:

```bash
grep -R -n "services/\|api-service\|worker-service\|storage-service" docs
```

Verify all managed references again:

```bash
ddocs check --links
```

## 12. Create a policy-compliant document

Create a new service document from the shipped service schema:

```bash
ddocs new service docs/services/scheduler-service.md
```

Open the generated file. It begins with policy-compliant YAML frontmatter and the service schema's required sections.

Add this authored entry under `Related docs` in `docs/services/api-service.md`:

```markdown
- [Scheduler service](scheduler-service.md)
```

Reconcile the addition:

```bash
ddocs fix
```

## 13. Inspect the review ledger

```bash
ddocs changes
ddocs changes log
```

The ledger records ordinary generated repairs and provides inspectable before-and-after history. To inspect one entry in detail, use an identifier from `ddocs changes`:

```bash
ddocs changes show <change-id>
```

Demon Docs also supports bounded hash-guarded undo, but performing an undo is unnecessary for the main walkthrough.

## 14. Enable automatic maintenance

Enable the repository daemon only after the explicit repair and reorganization demonstrations are complete:

```bash
ddocs demon run --true
ddocs demon --status
```

Now rename the new service with an ordinary filesystem command rather than `ddocs mv`:

```bash
mv \
  docs/services/scheduler-service.md \
  docs/services/task-scheduler.md
```

Allow the watcher to process the filesystem event:

```bash
sleep 2
```

Inspect the affected navigation and authored reference:

```bash
grep -n "scheduler" \
  docs/services/INDEX.md \
  docs/services/api-service.md
```

The index and the link from `api-service.md` should now reference `task-scheduler.md`, demonstrating that ordinary editor or filesystem changes can be reconciled automatically.

## 15. Prove convergence

```bash
ddocs check
ddocs fix
```

The complete check should pass. The final fix should report zero changed files, showing that the repository has reached a clean, idempotent state.

The walkthrough's progression is:

**disclose → diagnose → repair → decide → reorganize → automate → verify**
