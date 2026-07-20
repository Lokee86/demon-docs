# Seam-First Feature Skill

Use this skill when adding a new command behavior, reconciliation rule, mutation path, review behavior, watcher or demon behavior, schema behavior, codemap behavior, or other visible product behavior.

Do not use this skill for purely mechanical renames, path moves, formatting, generated-region updates, or behavior-preserving extractions where another workflow is more specific.

## Goal

Add features through the smallest correct owning seam instead of growing orchestration gravity wells.

Behavior may be small. Ownership should still be explicit.

## Core rules

- Identify the owning system before editing.
- If no clear owner exists, stop and report the missing seam.
- Do not place new behavior in a command router, application coordinator, watcher scheduler, demon lifecycle, review coordinator, or transaction coordinator only because it is convenient.
- Defer mechanics, not ownership.
- Prefer a tiny owner or seam now over extracting mature behavior later.
- Preserve existing behavior except for the requested feature slice.
- Do not create vague buckets such as `utils`, `helpers`, `common`, `misc`, or generic managers.
- Good seams have concrete responsibilities: repository scope, configuration selection, managed Markdown transformation, index reconciliation, link evidence, path rewriting, transaction publication, review decisions, watcher scheduling, demon ownership, codemap extraction, ranking, execution, or benchmark evaluation.
- If the feature needs policy, put policy in the owning system, not in command or orchestration glue.
- If the feature only needs routing, keep the routing thin.

## File-size rules

For hand-written production files:

- Prefer files under 200 lines.
- Treat 300+ lines as a hard architecture review point.
- If a feature task would add code to an already-large file, create or use a smaller owning file or seam instead.
- Do not add new behavior to large coordination or gravity-well files for convenience.

Generated files, fixtures, snapshots, pinned reports, and large declarative data files are exempt from strict line-count rules.

## Gravity-well warning files

Be especially cautious before adding behavior to:

- broad `internal/app` command coordinators;
- reconciliation or publication coordinators that already cross several systems;
- watcher or demon lifecycle coordinators;
- review command coordinators;
- broad transaction or repository-state owners; and
- CLI entry points under `cmd/`.

If a file is already large or coordinates multiple systems, do not add a new responsibility there unless the prompt explicitly says to.

## Feature workflow

1. Name the feature behavior in one sentence.
2. Identify the owning system.
3. Check whether an existing seam already owns it.
4. If an existing seam owns it, add the smallest behavior there.
5. If no seam owns it, create the smallest concrete owner first.
6. Route existing command or orchestration code through that owner.
7. Keep compatibility adapters only when required by existing contracts or call sites.
8. Do not combine unrelated seams in the same edit.
9. Keep verification focused on the changed seam, then use the broader release gate separately when appropriate.

## Stop conditions

Stop and report instead of editing if:

- The correct owner is unclear.
- The prompt requires adding behavior to a gravity-well file without a clear owner.
- The feature appears to involve multiple seams at once.
- The task requires generated-region, schema, private-state, or public CLI changes not mentioned in the prompt.
- The edit would require touching materially more files than the prompt allows.
- The feature would add another responsibility to an already over-broad production file.

## Terminal policy

- Focused, safe terminal checks are allowed when useful for the task.
- Avoid destructive Git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- Do not run tests, generators, or repo-wide scans by default when a small edit does not require them.
- Broad verification happens separately after the edit unless the prompt explicitly includes it.

## Report format

Report only:

```text
Changed files:
- ...

Owning seam used or created:
- ...

Unexpected files touched:
- none / ...

Notes:
- ...

**COMPLETED PROMPT X**
```

When completing a numbered prompt, put the exact completion heading at the bottom of the report, replacing `X` with the prompt number.
