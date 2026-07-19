#!/usr/bin/env python3
"""Prepare a reproducible pre-change source workspace for one historical OSS task."""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
from pathlib import Path


def run(args: list[str], cwd: Path | None = None) -> None:
    subprocess.run(args, cwd=cwd, check=True)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("repository", help="OWNER/REPO")
    parser.add_argument("base_commit")
    parser.add_argument("target", type=Path)
    parser.add_argument("--verification", default="go test ./...")
    args = parser.parse_args()

    target = args.target.resolve()
    source = target / "source"
    if target.exists():
        shutil.rmtree(target)
    target.mkdir(parents=True)

    run(["git", "init", str(source)])
    run(
        ["git", "remote", "add", "origin", f"https://github.com/{args.repository}.git"],
        source,
    )
    run(["git", "fetch", "--depth", "1", "origin", args.base_commit], source)
    run(["git", "checkout", "--detach", "FETCH_HEAD"], source)

    metadata = {
        "schema": 1,
        "repository": args.repository,
        "base_commit": args.base_commit,
        "verification": args.verification,
        "classification_status": "unreviewed",
    }
    (target / "metadata.json").write_text(json.dumps(metadata, indent=2), encoding="utf-8")
    (target / "TASK.md").write_text(
        "# Benchmark Task\n\nAdd the original issue text here before running an agent.\n",
        encoding="utf-8",
    )
    (target / "oracle.json").write_text(
        json.dumps({"warning": "Evaluator-only. Add accepted-change metadata here."}, indent=2),
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
