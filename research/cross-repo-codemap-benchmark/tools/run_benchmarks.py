#!/usr/bin/env python3
"""Run the frozen Demon Docs algorithm against prepared cross-repo datasets."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
from pathlib import Path

CALCULATION_MODES = {"primary", "diagnostic"}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--checkout-root",
        type=Path,
        help="reuse an external directory containing candidate checkout folders",
    )
    return parser.parse_args()


def resolve_repository(benchmark_root: Path, configured: str, checkout_root: Path | None) -> Path:
    relative = Path(configured)
    if checkout_root is None:
        return benchmark_root / relative
    if relative.parts and relative.parts[0] == "checkouts":
        relative = Path(*relative.parts[1:])
    return checkout_root.resolve() / relative


def aggregate(items: list[dict]) -> dict:
    completed = [item for item in items if item.get("status") == "completed"]
    hidden = sum(item["holdout_count"] for item in completed)
    recovered = sum(item["recovered_link_count"] for item in completed)
    suggestions = sum(item["unique_suggestion_count"] for item in completed)
    hard = sum(item["hard_recovered_count"] for item in completed)
    context = sum(item["context_recovered_count"] for item in completed)
    return {
        "repository_count": len(completed),
        "hidden_link_count": hidden,
        "recovered_link_count": recovered,
        "hard_recovered_count": hard,
        "context_recovered_count": context,
        "unique_suggestion_count": suggestions,
        "recall": recovered / hidden if hidden else 0,
        "positive_only_precision": recovered / suggestions if suggestions else 0,
    }


def summary_row(label: str, item: dict) -> str:
    return (
        f"| {label} | {item['repository_count']} | {item['hidden_link_count']} | "
        f"{item['recovered_link_count']} | {item['hard_recovered_count']} | "
        f"{item['context_recovered_count']} | {item['recall']:.2%} | "
        f"{item['positive_only_precision']:.2%} |"
    )


def main() -> int:
    args = parse_args()
    benchmark_root = Path(__file__).resolve().parents[1]
    repository_root = benchmark_root.parents[1]
    plan = json.loads((benchmark_root / "benchmark-plan.json").read_text(encoding="utf-8"))
    reports_dir = benchmark_root / "reports"
    reports_dir.mkdir(parents=True, exist_ok=True)

    env = os.environ.copy()
    env["GOCACHE"] = str(repository_root / ".cache" / "cross-repo-go-build")
    env["GIT_CONFIG_COUNT"] = "1"
    env["GIT_CONFIG_KEY_0"] = "safe.directory"
    env["GIT_CONFIG_VALUE_0"] = "*"

    summaries = []
    failed = False
    for benchmark in plan["benchmarks"]:
        candidate_id = benchmark["id"]
        repository = resolve_repository(benchmark_root, benchmark["repository_checkout"], args.checkout_root)
        dataset = benchmark_root / benchmark["dataset"]
        report_path = reports_dir / f"{candidate_id}.json"
        command = [
            "go", "run", "./cmd/ddocs", "codemap", "benchmark",
            "--repo", str(repository),
            "--dataset", str(dataset),
            "--seed", f"cross-repo-{candidate_id}-v1",
            "--holdout-count", str(benchmark["holdout_count"]),
            "--format", "json",
            "--output", str(report_path),
        ]
        print(f"benchmarking {candidate_id}...", flush=True)
        completed = subprocess.run(
            command,
            cwd=repository_root,
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            encoding="utf-8",
            errors="replace",
        )
        if completed.returncode != 0:
            failed = True
            summaries.append({
                **benchmark,
                "status": "error",
                "exit_code": completed.returncode,
                "stderr": completed.stderr.strip(),
            })
            continue

        report = json.loads(report_path.read_text(encoding="utf-8"))
        tier_counts = {"hard_link": 0, "context": 0}
        for suggestion in report.get("recovered_suggestions", []):
            tier = suggestion.get("tier") or "context"
            tier_counts[tier] = tier_counts.get(tier, 0) + 1
        summaries.append({
            **benchmark,
            "status": "completed",
            "raw_suggestion_count": report["raw_suggestion_count"],
            "unique_suggestion_count": report["unique_suggestion_count"],
            "recovered_link_count": len(report["recovered_links"]),
            "missed_link_count": len(report["missed_links"]),
            "hard_recovered_count": tier_counts.get("hard_link", 0),
            "context_recovered_count": tier_counts.get("context", 0),
            "precision": report["precision"],
            "recall": report["recall"],
        })

    aggregates = {
        "calculation_corpus": aggregate([
            item for item in summaries if item.get("benchmark_mode") in CALCULATION_MODES
        ]),
        "primary": aggregate([
            item for item in summaries if item.get("benchmark_mode") == "primary"
        ]),
        "diagnostic": aggregate([
            item for item in summaries if item.get("benchmark_mode") == "diagnostic"
        ]),
        "stress": aggregate([
            item for item in summaries if item.get("benchmark_mode") == "stress"
        ]),
    }
    result = {
        "schema_version": 2,
        "algorithm_baseline": plan["algorithm_baseline"],
        "calculation_policy": {
            "included_modes": sorted(CALCULATION_MODES),
            "excluded_modes": ["stress"],
            "reason": "monolithic per-file indexes test a different retrieval problem and are reported separately",
        },
        "aggregates": aggregates,
        "benchmarks": summaries,
    }
    summary_json = benchmark_root / "evaluation.json"
    summary_json.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")

    lines = [
        "# Cross-repository benchmark results",
        "",
        f"Frozen algorithm baseline: `{plan['algorithm_baseline']}`.",
        "",
        "Monolithic per-file indexes are excluded from the calculation corpus. They remain visible as a separate stress result.",
        "",
        "## Calculation corpus",
        "",
        "| Scope | Repositories | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |",
        "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
        summary_row("Primary + diagnostic", aggregates["calculation_corpus"]),
        summary_row("Primary", aggregates["primary"]),
        summary_row("Diagnostic", aggregates["diagnostic"]),
        "",
        "## Index stress corpus",
        "",
        "| Scope | Repositories | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |",
        "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
        summary_row("Monolithic per-file index", aggregates["stress"]),
        "",
        "## Repository details",
        "",
        "| Repository | Mode | Language(s) | Known | Hidden | Recovered | Hard | Context | Recall | Positive-only precision |",
        "| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |",
    ]
    for item in summaries:
        languages = ", ".join(item["languages"])
        if item["status"] != "completed":
            lines.append(
                f"| {item['id']} | {item['benchmark_mode']} | {languages} | {item['known_link_count']} | "
                f"{item['holdout_count']} | error | — | — | — | — |"
            )
            continue
        lines.append(
            f"| {item['id']} | {item['benchmark_mode']} | {languages} | {item['known_link_count']} | "
            f"{item['holdout_count']} | {item['recovered_link_count']} | {item['hard_recovered_count']} | "
            f"{item['context_recovered_count']} | {item['recall']:.2%} | {item['precision']:.2%} |"
        )
    lines.extend([
        "",
        "The precision columns are positive-only holdout precision: unmatched suggestions are counted as false because this corpus has not yet been manually labeled for genuinely new links. They must not be compared with the manually reviewed Space Rocks precision benchmark.",
        "",
        "The gbrain result is a stress test: one document owns hundreds of targets, and redacting the authored index removes nearly all topical prose. It is not used to tune or summarize ordinary document-to-code behavior.",
        "",
        "This run measures recovery only. Cross-repository precision still requires manual labeling of sampled unmatched suggestions.",
    ])
    summary_md = benchmark_root / "results.md"
    summary_md.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"wrote {summary_json}")
    print(f"wrote {summary_md}")
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
