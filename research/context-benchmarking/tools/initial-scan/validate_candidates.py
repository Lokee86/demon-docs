#!/usr/bin/env python3
"""Materialize reproducible issue/PR benchmark-task inventories for shortlisted repos."""

from __future__ import annotations

import datetime as dt
import json
import re
import subprocess
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]
RESULTS = ROOT / "discovery-results"

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

CANDIDATES: dict[str, list[int]] = {
    "shazow/wifitui": [174, 173, 163, 178, 167],
    "crossplane-contrib/provider-sql": [379, 377, 290, 361],
    "mercuretechnologies/expo-open-ota": [71, 41, 60],
    "openstack-exporter/openstack-exporter": [413, 378, 323, 541],
}

ISSUE_REF_RE = re.compile(
    r"(?i)\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\s+(?:https://github\.com/[^\s]+/issues/)?#?(\d+)"
)
LOOSE_REF_RE = re.compile(r"(?<![\w/])#(\d+)\b")


def run_json(args: list[str]) -> Any:
    result = subprocess.run(
        ["gh", *args],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        encoding="utf-8",
        errors="replace",
        timeout=180,
    )
    if result.returncode != 0:
        raise RuntimeError(f"gh {' '.join(args)} failed: {result.stderr.strip()}")
    return json.loads(result.stdout)


def pull_request(repo: str, number: int) -> dict[str, Any]:
    owner, name = repo.split("/", 1)
    raw = run_json(["api", f"repos/{owner}/{name}/pulls/{number}"])
    files = run_json(["api", "--paginate", f"repos/{owner}/{name}/pulls/{number}/files?per_page=100"])
    body = raw.get("body") or ""
    title = raw.get("title") or ""
    refs = set(ISSUE_REF_RE.findall(body))
    refs.update(LOOSE_REF_RE.findall(title))
    # Do not treat the PR number itself as an issue reference.
    refs.discard(str(number))

    issues: list[dict[str, Any]] = []
    for ref in sorted(refs, key=int):
        try:
            issue = run_json(["api", f"repos/{owner}/{name}/issues/{ref}"])
        except Exception as exc:  # noqa: BLE001
            issues.append({"number": int(ref), "error": str(exc)})
            continue
        if "pull_request" in issue:
            continue
        issues.append(
            {
                "number": issue.get("number"),
                "title": issue.get("title"),
                "body": issue.get("body") or "",
                "state": issue.get("state"),
                "created_at": issue.get("created_at"),
                "closed_at": issue.get("closed_at"),
                "url": issue.get("html_url"),
                "labels": [label.get("name") for label in issue.get("labels", [])],
            }
        )

    changed_paths = [item.get("filename") for item in files]
    test_paths = [path for path in changed_paths if path and ("test" in path.lower() or path.endswith("_test.go"))]
    generated_paths = [
        path
        for path in changed_paths
        if path
        and (
            "zz_generated" in path.lower()
            or "/generated/" in path.lower()
            or path.lower().endswith((".gen.go", ".generated.go"))
        )
    ]
    return {
        "repository": repo,
        "number": number,
        "title": title,
        "url": raw.get("html_url"),
        "body": body,
        "state": raw.get("state"),
        "merged_at": raw.get("merged_at"),
        "base_branch": raw.get("base", {}).get("ref"),
        "base_commit": raw.get("base", {}).get("sha"),
        "head_commit": raw.get("head", {}).get("sha"),
        "merge_commit": raw.get("merge_commit_sha"),
        "changed_files": raw.get("changed_files"),
        "additions": raw.get("additions"),
        "deletions": raw.get("deletions"),
        "changed_paths": changed_paths,
        "test_paths": test_paths,
        "generated_paths": generated_paths,
        "issues": issues,
    }


def task_quality(task: dict[str, Any]) -> tuple[int, list[str], list[str]]:
    score = 0
    strengths: list[str] = []
    concerns: list[str] = []
    issues = task["issues"]
    changed_files = int(task.get("changed_files") or 0)
    delta = int(task.get("additions") or 0) + int(task.get("deletions") or 0)

    if issues:
        score += 25
        strengths.append("has original issue context")
        if any(len((issue.get("body") or "").strip()) >= 100 for issue in issues):
            score += 10
            strengths.append("issue contains substantive reproduction or requirements")
    else:
        concerns.append("no linked issue recovered")

    if 2 <= changed_files <= 12:
        score += 20
        strengths.append("bounded multi-file change")
    elif changed_files == 1:
        score += 8
        concerns.append("mostly localized task")
    elif changed_files > 20:
        concerns.append("large patch surface")

    if 20 <= delta <= 600:
        score += 20
        strengths.append("reviewable implementation size")
    elif delta > 1000:
        concerns.append("large implementation delta")

    if task["test_paths"]:
        score += 20
        strengths.append("accepted patch changes tests")
    else:
        concerns.append("accepted patch has no obvious test change")

    if task["generated_paths"]:
        score -= 10
        concerns.append("generated-file churn may dominate the patch")

    unique_roots = {path.split("/", 1)[0] for path in task["changed_paths"] if path}
    if len(unique_roots) >= 2 or changed_files >= 4:
        score += 10
        strengths.append("requires broader repository context")

    if task.get("base_commit"):
        score += 5
    else:
        concerns.append("missing reproducible base commit")

    return max(0, min(100, score)), strengths, concerns


def main() -> int:
    RESULTS.mkdir(parents=True, exist_ok=True)
    tasks: list[dict[str, Any]] = []
    for repo, numbers in CANDIDATES.items():
        print(f"{repo}")
        for number in numbers:
            print(f"  PR #{number}", flush=True)
            try:
                task = pull_request(repo, number)
                score, strengths, concerns = task_quality(task)
                task["task_quality_score"] = score
                task["strengths"] = strengths
                task["concerns"] = concerns
                tasks.append(task)
            except Exception as exc:  # noqa: BLE001
                tasks.append({"repository": repo, "number": number, "error": str(exc)})

    tasks.sort(key=lambda item: (item.get("task_quality_score", -1), item["repository"]), reverse=True)
    payload = {
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "tasks": tasks,
    }
    (RESULTS / "task_inventory.json").write_text(json.dumps(payload, indent=2), encoding="utf-8")

    lines = [
        "# Candidate Historical Task Inventory",
        "",
        f"Generated: {payload['generated_at']}",
        "",
        "> Base commits are the pull request base SHAs reported by GitHub. Before using a task, verify that the issue was open against that exact state and run baseline tests at the pinned commit.",
        "",
    ]
    current_repo = None
    for task in sorted(tasks, key=lambda item: (item["repository"], -item.get("task_quality_score", -1))):
        if task["repository"] != current_repo:
            current_repo = task["repository"]
            lines.extend([f"## {current_repo}", ""])
        if "error" in task:
            lines.extend([f"### PR #{task['number']} — retrieval failed", "", task["error"], ""])
            continue
        lines.extend(
            [
                f"### [PR #{task['number']}]({task['url']}) — {task['title']}",
                "",
                f"- Task quality: {task['task_quality_score']}/100",
                f"- Base commit: `{task['base_commit']}`",
                f"- Patch: {task['changed_files']} files, +{task['additions']}/-{task['deletions']}",
                f"- Test paths: {', '.join(task['test_paths']) if task['test_paths'] else 'none detected'}",
                f"- Strengths: {', '.join(task['strengths']) if task['strengths'] else 'none'}",
                f"- Concerns: {', '.join(task['concerns']) if task['concerns'] else 'none'}",
                "",
                "Linked issues:",
                "",
            ]
        )
        if task["issues"]:
            for issue in task["issues"]:
                if issue.get("error"):
                    lines.append(f"- Issue #{issue['number']}: retrieval failed")
                else:
                    summary = " ".join((issue.get("body") or "").split())[:260]
                    lines.append(f"- [#{issue['number']}]({issue['url']}) — {issue['title']}: {summary}")
        else:
            lines.append("- None recovered.")
        lines.extend(["", "Changed paths:", ""])
        lines.extend(f"- `{path}`" for path in task["changed_paths"])
        lines.append("")

    (RESULTS / "task_inventory.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
