#!/usr/bin/env python3
"""Run benchmark baseline verification in the repository's natural Linux environment."""

from __future__ import annotations

import datetime as dt
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
BENCHMARKS = ROOT / "fixtures"

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")


def windows_to_wsl(path: Path) -> str:
    resolved = path.resolve()
    drive = resolved.drive.rstrip(":").lower()
    tail = resolved.as_posix().split(":", 1)[1]
    return f"/mnt/{drive}{tail}"


def main() -> int:
    summaries = []
    for task_dir in sorted(BENCHMARKS.iterdir()):
        if not task_dir.is_dir():
            continue
        source = task_dir / "source"
        if not source.is_dir():
            continue
        wsl_source = windows_to_wsl(source)
        command = f"cd {subprocess.list2cmdline([wsl_source])} && go test ./..."
        # list2cmdline is Windows-oriented, so use simple single-quote shell quoting here.
        command = "cd '" + wsl_source.replace("'", "'\\''") + "' && go test ./..."
        print(f"validating {task_dir.name}", flush=True)
        result = subprocess.run(
            ["wsl.exe", "bash", "-lic", command],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            encoding="utf-8",
            errors="replace",
            timeout=900,
        )
        record = {
            "task": task_dir.name,
            "verified_at": dt.datetime.now(dt.timezone.utc).isoformat(),
            "environment": "WSL bash login shell",
            "command": "go test ./...",
            "exit_code": result.returncode,
            "stdout": result.stdout,
            "stderr": result.stderr,
        }
        (task_dir / "baseline-validation.json").write_text(
            json.dumps(record, indent=2), encoding="utf-8"
        )
        summaries.append(record)
        print(f"  exit {result.returncode}", flush=True)
    (BENCHMARKS / "validation-summary.json").write_text(
        json.dumps(summaries, indent=2), encoding="utf-8"
    )
    return 0 if summaries and all(item["exit_code"] == 0 for item in summaries) else 1


if __name__ == "__main__":
    raise SystemExit(main())
