#!/usr/bin/env python3
"""Run and record a benchmark fixture's baseline verification command."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import subprocess
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("fixture", type=Path)
    parser.add_argument("--shell", default="bash")
    args = parser.parse_args()

    fixture = args.fixture.resolve()
    metadata = json.loads((fixture / "metadata.json").read_text(encoding="utf-8"))
    command = metadata["verification"]
    result = subprocess.run(
        [args.shell, "-lc", command],
        cwd=fixture / "source",
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        encoding="utf-8",
        errors="replace",
    )
    record = {
        "verified_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "command": command,
        "exit_code": result.returncode,
        "stdout": result.stdout,
        "stderr": result.stderr,
    }
    (fixture / "baseline-validation.json").write_text(
        json.dumps(record, indent=2), encoding="utf-8"
    )
    return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
