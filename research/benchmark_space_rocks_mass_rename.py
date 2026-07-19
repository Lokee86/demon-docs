from __future__ import annotations

import csv
import json
import math
import re
import shutil
import statistics
import subprocess
import tempfile
import time
from pathlib import Path
from typing import Callable, TypeVar

WORKTREE = Path(r"D:\!bin\demon-docs-mass-rename-tests")
SOURCE = Path(r"D:\!bin\space-rocks\docs")
BINARY = WORKTREE / "bin" / "ddocs.exe"
RESULT_DIR = WORKTREE / "research" / "mass-rename-timing"
ITERATIONS = 5
T = TypeVar("T")


def timed(call: Callable[[], T]) -> tuple[T, float]:
    started = time.perf_counter_ns()
    result = call()
    elapsed_ms = (time.perf_counter_ns() - started) / 1_000_000
    return result, elapsed_ms


def command(root: Path, *args: str) -> dict[str, object]:
    result = subprocess.run(
        [str(BINARY), *args],
        cwd=root,
        text=True,
        capture_output=True,
        encoding="utf-8",
        errors="replace",
    )
    output = result.stdout + result.stderr
    updated = re.search(r"ddocs fix updated (\d+) file\(s\)", output)
    return {
        "exit_code": result.returncode,
        "updated": int(updated.group(1)) if updated else None,
        "repairs": output.count("Repair link in "),
        "broken": output.count("Broken link in "),
        "ambiguous": output.count("Ambiguous link in "),
        "errors": [line for line in output.splitlines() if line.startswith("ddocs error:")],
    }


def rename_all(root: Path, suffix: str) -> int:
    files = sorted(root.rglob("*.md"), key=lambda path: path.as_posix())
    temporary: list[tuple[Path, Path]] = []
    for index, old in enumerate(files):
        final = old.with_name(f"{old.stem}{suffix}{old.suffix}")
        temp = old.with_name(f".__ddocs_timing_{index:04d}__.tmp")
        old.rename(temp)
        temporary.append((temp, final))
    for temp, final in temporary:
        temp.rename(final)
    return len(files)


def run_iteration(index: int) -> dict[str, object]:
    root = Path(tempfile.mkdtemp(prefix=f"ddocs-mass-rename-timing-{index:02d}-", dir=r"C:\tmp"))
    stages: dict[str, float] = {}
    observations: dict[str, object] = {}
    iteration_started = time.perf_counter_ns()

    _, stages["copy_ms"] = timed(lambda: shutil.copytree(SOURCE, root, dirs_exist_ok=True))

    for stage, args in [
        ("init_ms", ("init", "--root", ".")),
        ("baseline_initialize_ms", ("fix", "-l")),
        ("baseline_repair_ms", ("fix", "-l")),
        ("baseline_idempotent_ms", ("fix", "-l")),
        ("baseline_check_ms", ("check", "-l")),
    ]:
        result, elapsed = timed(lambda args=args: command(root, *args))
        stages[stage] = elapsed
        observations[stage.removesuffix("_ms")] = result

    first_count, stages["first_filesystem_rename_ms"] = timed(lambda: rename_all(root, "-mass-renamed"))
    observations["first_files_renamed"] = first_count
    for stage, args in [
        ("first_precheck_ms", ("check", "-l")),
        ("first_fix_ms", ("fix", "-l")),
        ("first_postcheck_ms", ("check", "-l")),
        ("first_idempotent_ms", ("fix", "-l")),
    ]:
        result, elapsed = timed(lambda args=args: command(root, *args))
        stages[stage] = elapsed
        observations[stage.removesuffix("_ms")] = result

    second_count, stages["second_filesystem_rename_ms"] = timed(lambda: rename_all(root, "-again"))
    observations["second_files_renamed"] = second_count
    for stage, args in [
        ("second_precheck_ms", ("check", "-l")),
        ("second_fix_ms", ("fix", "-l")),
        ("second_postcheck_ms", ("check", "-l")),
        ("second_idempotent_ms", ("fix", "-l")),
    ]:
        result, elapsed = timed(lambda args=args: command(root, *args))
        stages[stage] = elapsed
        observations[stage.removesuffix("_ms")] = result

    stages["first_rename_cycle_ms"] = sum(stages[key] for key in (
        "first_filesystem_rename_ms", "first_precheck_ms", "first_fix_ms", "first_postcheck_ms", "first_idempotent_ms"
    ))
    stages["second_rename_cycle_ms"] = sum(stages[key] for key in (
        "second_filesystem_rename_ms", "second_precheck_ms", "second_fix_ms", "second_postcheck_ms", "second_idempotent_ms"
    ))
    stages["two_rename_cycles_ms"] = stages["first_rename_cycle_ms"] + stages["second_rename_cycle_ms"]
    stages["total_ms"] = (time.perf_counter_ns() - iteration_started) / 1_000_000

    first_fix = observations["first_fix"]
    second_fix = observations["second_fix"]
    assert isinstance(first_fix, dict) and isinstance(second_fix, dict)
    valid = (
        first_count == 341
        and second_count == 341
        and first_fix.get("updated") == 340
        and second_fix.get("updated") == 340
        and first_fix.get("repairs") == 3717
        and second_fix.get("repairs") == 3717
        and not first_fix.get("errors")
        and not second_fix.get("errors")
    )

    result = {
        "iteration": index,
        "sandbox": str(root),
        "valid": valid,
        "stages_ms": stages,
        "observations": observations,
    }
    shutil.rmtree(root, ignore_errors=True)
    return result


def percentile(values: list[float], percentile_value: float) -> float:
    ordered = sorted(values)
    if len(ordered) == 1:
        return ordered[0]
    position = (len(ordered) - 1) * percentile_value
    lower = math.floor(position)
    upper = math.ceil(position)
    if lower == upper:
        return ordered[lower]
    fraction = position - lower
    return ordered[lower] * (1 - fraction) + ordered[upper] * fraction


def summarize(iterations: list[dict[str, object]]) -> dict[str, object]:
    stage_names = list(iterations[0]["stages_ms"].keys())
    stages: dict[str, object] = {}
    for stage in stage_names:
        values = [float(item["stages_ms"][stage]) for item in iterations]
        stages[stage] = {
            "samples_ms": [round(value, 3) for value in values],
            "mean_ms": round(statistics.mean(values), 3),
            "median_ms": round(statistics.median(values), 3),
            "p95_ms": round(percentile(values, 0.95), 3),
            "min_ms": round(min(values), 3),
            "max_ms": round(max(values), 3),
            "stddev_ms": round(statistics.pstdev(values), 3),
        }

    first_fix_seconds = float(stages["first_fix_ms"]["median_ms"]) / 1000
    second_fix_seconds = float(stages["second_fix_ms"]["median_ms"]) / 1000
    two_cycles_seconds = float(stages["two_rename_cycles_ms"]["median_ms"]) / 1000
    return {
        "iterations": len(iterations),
        "all_valid": all(bool(item["valid"]) for item in iterations),
        "corpus": {"total_files": 346, "markdown_files": 341, "rewritten_source_files_per_pass": 340, "link_repairs_per_pass": 3717},
        "stages": stages,
        "throughput_from_medians": {
            "first_fix_files_per_second": round(340 / first_fix_seconds, 2),
            "first_fix_links_per_second": round(3717 / first_fix_seconds, 2),
            "second_fix_files_per_second": round(340 / second_fix_seconds, 2),
            "second_fix_links_per_second": round(3717 / second_fix_seconds, 2),
            "two_cycles_renamed_files_per_second": round(682 / two_cycles_seconds, 2),
            "two_cycles_link_repairs_per_second": round(7434 / two_cycles_seconds, 2),
        },
    }


def write_markdown(summary: dict[str, object]) -> None:
    stages = summary["stages"]
    selected = [
        "copy_ms", "baseline_initialize_ms", "baseline_repair_ms", "baseline_check_ms",
        "first_filesystem_rename_ms", "first_precheck_ms", "first_fix_ms", "first_postcheck_ms", "first_idempotent_ms", "first_rename_cycle_ms",
        "second_filesystem_rename_ms", "second_precheck_ms", "second_fix_ms", "second_postcheck_ms", "second_idempotent_ms", "second_rename_cycle_ms",
        "two_rename_cycles_ms", "total_ms",
    ]
    lines = [
        "# Space Rocks Mass-Rename Timing",
        "",
        f"Measured iterations: {summary['iterations']}",
        f"All iterations valid: {summary['all_valid']}",
        "",
        "| Stage | Median | Mean | P95 | Min | Max |",
        "|---|---:|---:|---:|---:|---:|",
    ]
    for name in selected:
        item = stages[name]
        lines.append(
            f"| {name} | {item['median_ms']:.3f} ms | {item['mean_ms']:.3f} ms | {item['p95_ms']:.3f} ms | {item['min_ms']:.3f} ms | {item['max_ms']:.3f} ms |"
        )
    throughput = summary["throughput_from_medians"]
    lines.extend([
        "",
        "## Throughput",
        "",
        f"- First fix: {throughput['first_fix_files_per_second']} files/s; {throughput['first_fix_links_per_second']} link repairs/s",
        f"- Second fix: {throughput['second_fix_files_per_second']} files/s; {throughput['second_fix_links_per_second']} link repairs/s",
        f"- Complete two-cycle scenario: {throughput['two_cycles_renamed_files_per_second']} renamed files/s; {throughput['two_cycles_link_repairs_per_second']} applied link repairs/s",
        "",
    ])
    (RESULT_DIR / "summary.md").write_text("\n".join(lines), encoding="utf-8")


def main() -> None:
    Path(r"C:\tmp").mkdir(parents=True, exist_ok=True)
    shutil.rmtree(RESULT_DIR, ignore_errors=True)
    RESULT_DIR.mkdir(parents=True, exist_ok=True)

    iterations: list[dict[str, object]] = []
    for index in range(1, ITERATIONS + 1):
        result = run_iteration(index)
        iterations.append(result)
        print(f"iteration {index}/{ITERATIONS}: valid={result['valid']} total={result['stages_ms']['total_ms']:.3f} ms", flush=True)

    summary = summarize(iterations)
    (RESULT_DIR / "iterations.json").write_text(json.dumps(iterations, indent=2), encoding="utf-8")
    (RESULT_DIR / "summary.json").write_text(json.dumps(summary, indent=2), encoding="utf-8")
    write_markdown(summary)

    with (RESULT_DIR / "samples.csv").open("w", newline="", encoding="utf-8") as handle:
        stage_names = list(iterations[0]["stages_ms"].keys())
        writer = csv.DictWriter(handle, fieldnames=["iteration", "valid", *stage_names])
        writer.writeheader()
        for item in iterations:
            writer.writerow({"iteration": item["iteration"], "valid": item["valid"], **item["stages_ms"]})

    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
