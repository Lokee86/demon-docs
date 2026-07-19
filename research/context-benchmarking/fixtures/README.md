# Historical Task Fixtures

Fixture manifests in this directory describe authentic historical OSS tasks at pinned pre-change commits.

A fixture keeps these concerns separate:

- `TASK.md`: task text visible to the agent;
- `metadata.json`: public repository, issue, base-commit, classification, and verification metadata;
- `oracle.json`: evaluator-only accepted-change metadata;
- `baseline-validation.json`: evidence that the pinned snapshot was viable before the task; and
- generated `source/` workspaces: reproducible and disposable, not committed by default.

The agent must never receive `oracle.json`, post-change repository content, or artifacts from a previous run.

## Retained Initial Fixtures

- `wifitui-pr-163`: full-width terminal layout behavior.
- `wifitui-pr-167`: access-point annotation and theme behavior.
- `wifitui-pr-178`: `NO_COLOR` behavior.

All three base snapshots passed `go test ./...` in a WSL login shell during the initial investigation. They remain candidate tasks, not a complete benchmark corpus and not proof that `wifitui` belongs in a specific quality quadrant.

`validation-summary.json` retains the combined initial baseline evidence.
