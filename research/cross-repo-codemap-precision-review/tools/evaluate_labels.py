#!/usr/bin/env python3
"""Evaluate a completed cross-repository precision review queue."""

from __future__ import annotations

import json
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any

VALID_LABELS = {"valid", "plausible", "incorrect"}


def metrics(counter: Counter[str]) -> dict[str, Any]:
    total = sum(counter.values())
    valid = counter["valid"]
    plausible = counter["plausible"]
    return {
        "reviewed": total,
        "valid": valid,
        "plausible": plausible,
        "incorrect": counter["incorrect"],
        "strict_precision": valid / total if total else 0,
        "relevance_precision": (valid + plausible) / total if total else 0,
    }


def main() -> int:
    review_root = Path(__file__).resolve().parents[1]
    manifest = json.loads((review_root / "sample-manifest.json").read_text(encoding="utf-8"))
    queue = json.loads((review_root / "labels.json").read_text(encoding="utf-8"))
    metadata = {item["id"]: item for item in manifest["suggestions"]}

    missing: list[str] = []
    counters: dict[str, Counter[str]] = defaultdict(Counter)
    by_repository: dict[str, Counter[str]] = defaultdict(Counter)
    by_evidence: dict[str, Counter[str]] = defaultdict(Counter)
    by_split: dict[str, Counter[str]] = defaultdict(Counter)
    validation_repositories = set(
        manifest.get("evaluation_split", {}).get("validation_repositories", [])
    )

    for item in queue["labels"]:
        label = item["label"]
        if not label:
            missing.append(item["id"])
            continue
        if label not in VALID_LABELS:
            raise SystemExit(f"invalid label {label!r} for {item['id']}")
        source = metadata[item["id"]]
        counters["overall"][label] += 1
        counters[source["tier"]][label] += 1
        by_repository[source["repository"]][label] += 1
        split = "validation" if source["repository"] in validation_repositories else "tuning"
        by_split[split][label] += 1
        for family in source["evidence_families"]:
            by_evidence[family][label] += 1

    if missing:
        raise SystemExit(f"unlabeled suggestions: {', '.join(missing)}")

    result = {
        "schema_version": 1,
        "algorithm_baseline": manifest["algorithm_baseline"],
        "sample_count": manifest["sample_count"],
        "overall": metrics(counters["overall"]),
        "by_tier": {
            tier: metrics(counters[tier]) for tier in ("hard_link", "context")
        },
        "by_repository": {
            repository: metrics(counter)
            for repository, counter in sorted(by_repository.items())
        },
        "by_split": {
            split: metrics(counter) for split, counter in sorted(by_split.items())
        },
        "by_evidence_family": {
            family: metrics(counter) for family, counter in sorted(by_evidence.items())
        },
    }
    (review_root / "evaluation.json").write_text(
        json.dumps(result, indent=2) + "\n", encoding="utf-8"
    )
    print(json.dumps(result, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
