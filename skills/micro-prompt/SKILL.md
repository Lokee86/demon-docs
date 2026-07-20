# Agent Micro Refactor Skill

Use this skill for the default Demon Docs implementation workflow: one small, bounded agent task that makes a reviewable edit.

## When to use

Use this skill when the prompt asks for a surgical implementation, small refactor, file move, route-through, cleanup of a known seam, or one prompt from a numbered prompt sequence.

Do not use this skill for broad planning-only work, large architectural redesigns, or read-only reports unless the prompt explicitly asks for a read-only scan.

## Core rules

- Keep the diff small and reviewable.
- Preserve current behavior unless the prompt explicitly allows behavior change.
- Do not broaden scope beyond the named target.
- Do not do opportunistic cleanup, formatting churn, dependency upgrades, or unrelated renames.
- Do not edit generated regions or implementation-owned private state unless the prompt explicitly targets the owning workflow.
- If the task balloons, stop and report the smallest next prompt instead of continuing.
- If the prompt says read-only, do not edit files, run formatters, or perform cleanup.

## Line-count guardrails

Line count is a smell, not an automatic failure.

For hand-written production files:

- Prefer files under about 200 lines when practical.
- Treat 300+ lines as a review trigger: confirm the file still has one clear responsibility.
- Treat 350+ lines as a warning: avoid adding new responsibility unless it clearly belongs there.
- Treat 500+ lines as a refactor trigger for actively changing files.
- Files over 500 lines should usually receive behavior-preserving edits, bug fixes, routing, or extraction work, not new responsibilities.

Exemptions: generated files, fixtures, snapshots, pinned reports, and large declarative data files.

## Workflow

1. Open and read only the files needed for the requested edit.
2. Identify the smallest concrete change that satisfies the prompt.
3. Make the edit without unrelated cleanup.
4. Focused, safe terminal checks are allowed when useful for the task.
5. If verification is not requested, report what changed and whether any unexpected scope was needed.

## File reading policy

- Reading files is allowed and expected.
- Always read the named files needed for the edit.
- Reading a directly referenced file is allowed when a named file points to it through an import, interface, collaborator, test, or obvious call site.
- Do not turn a small edit into a repo-wide audit.

## Shell / verification policy

- Focused, safe terminal checks are allowed when useful for the task.
- Avoid destructive Git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- Do not run tests, generators, or repo-wide scans by default when a small edit does not require them.
- The default delegated agent job is editing only.
- Report changed files and unexpected scope changes.
- Verification commands are usually run separately by the primary agent or user.

## Report format

Report only useful execution evidence:

```text
Changed files:
- ...

Unexpected files touched:
- none / ...

Notes:
- ...

**COMPLETED PROMPT X**
```

When completing a numbered prompt, put the exact completion heading at the bottom of the report, replacing `X` with the prompt number.
