#!/usr/bin/env python3
"""Prepare fixed pre-change source snapshots for selected benchmark tasks."""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]
RESULTS = ROOT / "discovery-results"
BENCHMARKS = ROOT / "fixtures"
SELECTED = {
    ("shazow/wifitui", 163),
    ("shazow/wifitui", 167),
    ("shazow/wifitui", 178),
}

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")


def run(args: list[str], cwd: Path | None = None, timeout: int = 300) -> None:
    result = subprocess.run(
        args,
        cwd=str(cwd) if cwd else None,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        encoding="utf-8",
        errors="replace",
        timeout=timeout,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"command failed ({result.returncode}): {' '.join(args)}\n{result.stdout}\n{result.stderr}"
        )


def slug(repo: str, number: int) -> str:
    return f"{repo.replace('/', '__')}-pr-{number}"


def issue_markdown(task: dict[str, Any]) -> str:
    issue = task["issues"][0] if task.get("issues") else None
    lines = [
        f"# Benchmark Task: {task['repository']} PR #{task['number']}",
        "",
        "## Working repository",
        "",
        "Use the `source/` directory. It is pinned to the repository state before the accepted upstream change.",
        "",
        "## Task",
        "",
    ]
    if issue:
        lines.extend([f"### {issue['title']}", "", issue.get("body") or "No issue body.", ""])
    else:
        lines.extend([task["title"], "", task.get("body") or "No task body.", ""])
    lines.extend(
        [
            "## Verification",
            "",
            "```bash",
            "go test ./...",
            "```",
            "",
            "Do not inspect `oracle.json` while performing the benchmark. It exists only for evaluation after the attempt.",
        ]
    )
    return "\n".join(lines) + "\n"


def prepare(task: dict[str, Any]) -> None:
    target = BENCHMARKS / slug(task["repository"], task["number"])
    source = target / "source"
    if target.exists():
        shutil.rmtree(target)
    target.mkdir(parents=True)

    owner, name = task["repository"].split("/", 1)
    url = f"https://github.com/{owner}/{name}.git"
    run(["git", "init", str(source)])
    run(["git", "-c", "safe.directory=*", "remote", "add", "origin", url], cwd=source)
    run(
        [
            "git",
            "-c",
            "safe.directory=*",
            "-c",
            "core.longpaths=true",
            "fetch",
            "--depth",
            "1",
            "origin",
            task["base_commit"],
        ],
        cwd=source,
        timeout=600,
    )
    run(
        ["git", "-c", "safe.directory=*", "checkout", "--detach", "FETCH_HEAD"],
        cwd=source,
    )

    (target / "TASK.md").write_text(issue_markdown(task), encoding="utf-8")
    public_metadata = {
        "repository": task["repository"],
        "pull_request": task["number"],
        "base_commit": task["base_commit"],
        "issue_numbers": [issue.get("number") for issue in task.get("issues", [])],
        "verification": "go test ./...",
    }
    (target / "metadata.json").write_text(json.dumps(public_metadata, indent=2), encoding="utf-8")
    oracle = {
        "accepted_pr": task["url"],
        "accepted_title": task["title"],
        "head_commit": task["head_commit"],
        "merge_commit": task["merge_commit"],
        "changed_paths": task["changed_paths"],
        "test_paths": task["test_paths"],
        "additions": task["additions"],
        "deletions": task["deletions"],
    }
    (target / "oracle.json").write_text(json.dumps(oracle, indent=2), encoding="utf-8")
    print(f"prepared {target.name} at {task['base_commit']}")


def main() -> int:
    inventory = json.loads((RESULTS / "task_inventory.json").read_text(encoding="utf-8"))
    tasks = {
        (task.get("repository"), task.get("number")): task
        for task in inventory["tasks"]
        if "error" not in task
    }
    missing = SELECTED - set(tasks)
    if missing:
        raise RuntimeError(f"selected tasks are absent from task inventory: {sorted(missing)}")
    BENCHMARKS.mkdir(parents=True, exist_ok=True)
    for key in sorted(SELECTED):
        prepare(tasks[key])
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
