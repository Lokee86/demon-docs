#!/usr/bin/env python3
"""Collect repository metadata for future context-benchmark candidate review.

This tool intentionally does not rank repositories by stars or assign code/documentation
quality automatically. It gathers evidence for later human classification against the
2x2 benchmark matrix documented in docs/context-injection-benchmarking.md.
"""

from __future__ import annotations

import argparse
import json
import subprocess
from pathlib import Path
from typing import Any


def gh_json(args: list[str]) -> Any:
    result = subprocess.run(
        ["gh", *args],
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        encoding="utf-8",
        errors="replace",
    )
    return json.loads(result.stdout)


def inspect_repository(name: str) -> dict[str, Any]:
    repo = gh_json(
        [
            "repo",
            "view",
            name,
            "--json",
            "nameWithOwner,url,description,defaultBranchRef,licenseInfo,isArchived,isFork,"
            "stargazerCount,updatedAt,issues,pullRequests",
        ]
    )
    repo["candidate_note"] = (
        "Stars are retained as metadata only and must not affect quadrant assignment."
    )
    return repo


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("repositories", nargs="+", help="OWNER/REPO candidates")
    parser.add_argument("--output", default="candidate-metadata.json")
    args = parser.parse_args()

    payload = {
        "schema": 1,
        "classification_status": "unreviewed",
        "repositories": [inspect_repository(name) for name in args.repositories],
    }
    Path(args.output).write_text(json.dumps(payload, indent=2), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
