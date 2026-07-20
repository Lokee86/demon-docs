---
author: brian
created: "2026-07-19"
document_id: 9aaf9bd3-bfe8-4d5a-8812-897e89d2de30
document_type: general
policy_exempt: false
summary: This document owns prompt and report expectations for implementation agents working on Demon Docs.
---
# Prompting and Reporting

Parent index: [Agent](./INDEX.md)

## Purpose

This document owns prompt and report expectations for implementation agents working on Demon Docs.

## Overview

Demon Docs agent work should stay small, bounded, and easy to review.

## Rules

- Prompts should be small and specific.
- Implementation tasks should usually target under two minutes of delegated agent work.
- Each prompt should have one clear edit goal.
- Avoid mixing unrelated refactors.
- Stop if the task balloons.
- Reports should include changed files.
- Reports should mention unexpected files touched.
- Numbered completion headings should be placed at the bottom when requested.
- Command output should only be reported when the command was actually run.

## Related docs

- [Testing](testing.md)
- [Documentation Editing](documentation-editing.md)
- [Repo Hygiene](repo-hygiene.md)
- [Micro-prompt skill](../../skills/micro-prompt/SKILL.md)

## Notes

This document does not replace task-specific instructions.
