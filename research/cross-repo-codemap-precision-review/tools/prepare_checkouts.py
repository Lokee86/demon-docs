#!/usr/bin/env python3
"""Prepare the exact pinned repository revisions used by the precision review."""

from __future__ import annotations

import json
import shutil
import subprocess
from pathlib import Path

INCLUDED_MODES = {"primary", "diagnostic"}


def run(args: list[str], cwd: Path | None = None) -> str:
    command = args
    if cwd is not None and args and args[0] == "git":
        command = ["git", "-c", f"safe.directory={cwd.resolve()}", *args[1:]]
    completed = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    return completed.stdout.strip()


def checkout_revision(url: str, revision: str, destination: Path) -> None:
    if destination.exists():
        try:
            current = run(["git", "rev-parse", "HEAD"], cwd=destination)
        except (OSError, subprocess.CalledProcessError):
            shutil.rmtree(destination)
        else:
            if current == revision:
                return
            shutil.rmtree(destination)

    destination.parent.mkdir(parents=True, exist_ok=True)
    run(["git", "clone", "--filter=blob:none", "--no-checkout", "--no-tags", url, str(destination)])
    run(["git", "fetch", "--depth", "1", "origin", revision], cwd=destination)
    run(["git", "checkout", "--detach", "FETCH_HEAD"], cwd=destination)
    current = run(["git", "rev-parse", "HEAD"], cwd=destination)
    if current != revision:
        raise RuntimeError(f"{destination}: got {current}, expected {revision}")


def main() -> int:
    review_root = Path(__file__).resolve().parents[1]
    repository_root = review_root.parents[1]
    benchmark_root = repository_root / "research" / "cross-repo-codemap-benchmark"
    plan = json.loads((benchmark_root / "benchmark-plan.json").read_text(encoding="utf-8"))
    candidates = json.loads((benchmark_root / "candidates.json").read_text(encoding="utf-8"))
    urls = {item["id"]: item["code_repository"] for item in candidates["candidates"]}
    checkout_root = benchmark_root / "checkouts"

    for benchmark in plan["benchmarks"]:
        if benchmark["benchmark_mode"] not in INCLUDED_MODES:
            continue
        repository = benchmark["id"]
        destination = checkout_root / repository / "code"
        print(f"preparing {repository} at {benchmark['code_revision']}...", flush=True)
        checkout_revision(urls[repository], benchmark["code_revision"], destination)

    print(f"prepared pinned checkouts under {checkout_root}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
