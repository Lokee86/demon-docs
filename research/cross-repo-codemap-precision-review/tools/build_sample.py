#!/usr/bin/env python3
"""Build a deterministic, tier-balanced manual precision review sample."""

from __future__ import annotations

import argparse
import hashlib
import json
from collections import defaultdict
from pathlib import Path
from typing import Any

INCLUDED_MODES = {"primary", "diagnostic"}
DEFAULT_PER_TIER = 25
SEED = "cross-repo-precision-review-v1"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--per-tier", type=int, default=DEFAULT_PER_TIER)
    parser.add_argument(
        "--reset-labels",
        action="store_true",
        help="replace an existing review queue with blank labels",
    )
    return parser.parse_args()


def stable_key(repository: str, suggestion: dict[str, Any]) -> str:
    material = "\0".join(
        [
            SEED,
            repository,
            suggestion["tier"],
            suggestion["document"],
            suggestion["target"],
        ]
    )
    return hashlib.sha256(material.encode("utf-8")).hexdigest()


def evidence_families(suggestion: dict[str, Any]) -> tuple[str, ...]:
    families = sorted({item.split(":", 1)[0] for item in suggestion.get("evidence", [])})
    return tuple(families) or ("none",)


def score_bucket(items: list[dict[str, Any]], item: dict[str, Any]) -> str:
    ordered = sorted(candidate["score"] for candidate in items)
    if len(ordered) < 3:
        return "all"
    low_cut = ordered[(len(ordered) - 1) // 3]
    high_cut = ordered[(2 * (len(ordered) - 1)) // 3]
    score = item["score"]
    if score <= low_cut:
        return "low"
    if score >= high_cut:
        return "high"
    return "mid"


def select_stratified(repository: str, items: list[dict[str, Any]], limit: int) -> list[dict[str, Any]]:
    if len(items) <= limit:
        return sorted(items, key=lambda item: stable_key(repository, item))

    groups: dict[tuple[str, tuple[str, ...]], list[dict[str, Any]]] = defaultdict(list)
    for item in items:
        groups[(score_bucket(items, item), evidence_families(item))].append(item)
    for group_items in groups.values():
        group_items.sort(key=lambda item: stable_key(repository, item))

    selected: list[dict[str, Any]] = []
    ordered_groups = sorted(groups)
    while len(selected) < limit:
        progressed = False
        for group in ordered_groups:
            group_items = groups[group]
            if not group_items:
                continue
            selected.append(group_items.pop(0))
            progressed = True
            if len(selected) == limit:
                break
        if not progressed:
            break
    return selected


def main() -> int:
    args = parse_args()
    if args.per_tier < 1:
        raise SystemExit("--per-tier must be positive")

    review_root = Path(__file__).resolve().parents[1]
    repository_root = review_root.parents[1]
    benchmark_root = repository_root / "research" / "cross-repo-codemap-benchmark"
    plan = json.loads((benchmark_root / "benchmark-plan.json").read_text(encoding="utf-8"))

    selected: list[dict[str, Any]] = []
    inventory: list[dict[str, Any]] = []
    sequence = 1

    for benchmark in plan["benchmarks"]:
        if benchmark["benchmark_mode"] not in INCLUDED_MODES:
            continue
        repository = benchmark["id"]
        report = json.loads(
            (benchmark_root / "reports" / f"{repository}.json").read_text(encoding="utf-8")
        )
        unmatched = report["unmatched_suggestions"]
        by_tier: dict[str, list[dict[str, Any]]] = defaultdict(list)
        for item in unmatched:
            by_tier[item["tier"]].append(item)

        for tier in ("hard_link", "context"):
            candidates = by_tier[tier]
            chosen = select_stratified(repository, candidates, args.per_tier)
            inventory.append(
                {
                    "repository": repository,
                    "tier": tier,
                    "available": len(candidates),
                    "selected": len(chosen),
                }
            )
            for item in chosen:
                review_id = f"CRP-{sequence:03d}"
                sequence += 1
                selected.append(
                    {
                        "id": review_id,
                        "repository": repository,
                        "benchmark_mode": benchmark["benchmark_mode"],
                        "document": item["document"],
                        "target": item["target"],
                        "tier": item["tier"],
                        "score": item["score"],
                        "score_bucket": score_bucket(candidates, item),
                        "evidence_families": list(evidence_families(item)),
                        "evidence": item.get("evidence", []),
                    }
                )

    manifest = {
        "schema_version": 1,
        "seed": SEED,
        "algorithm_baseline": plan["algorithm_baseline"],
        "selection_policy": {
            "included_modes": sorted(INCLUDED_MODES),
            "excluded_modes": ["stress"],
            "per_repository_per_tier_cap": args.per_tier,
            "stratification": ["tier", "score_bucket", "evidence_families"],
        },
        "evaluation_split": {
            "tuning_repositories": [
                "agent-orchestrator",
                "beads-rust",
                "genesis",
                "render-claude-context",
            ],
            "validation_repositories": ["bifrost"],
            "rationale": "Bifrost is the only sampled repository with unmatched hard-tier suggestions, so it is reserved for repository-level validation.",
        },
        "inventory": inventory,
        "sample_count": len(selected),
        "suggestions": selected,
    }
    queue = {
        "schema_version": 1,
        "sample_manifest": "sample-manifest.json",
        "labels": [
            {
                "id": item["id"],
                "repository": item["repository"],
                "document": item["document"],
                "target": item["target"],
                "label": "",
                "rationale": "",
                "reviewer": "",
            }
            for item in selected
        ],
    }

    manifest_path = review_root / "sample-manifest.json"
    labels_path = review_root / "labels.json"
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    if labels_path.exists() and not args.reset_labels:
        existing = json.loads(labels_path.read_text(encoding="utf-8"))
        existing_ids = [item["id"] for item in existing.get("labels", [])]
        generated_ids = [item["id"] for item in queue["labels"]]
        if existing_ids != generated_ids:
            raise SystemExit(
                "existing labels.json does not match the generated sample; "
                "inspect the change before using --reset-labels"
            )
        print("preserved existing labels.json")
    else:
        labels_path.write_text(json.dumps(queue, indent=2) + "\n", encoding="utf-8")

    print(f"selected {len(selected)} suggestions")
    for row in inventory:
        print(
            f"{row['repository']} {row['tier']}: "
            f"{row['selected']}/{row['available']}"
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
