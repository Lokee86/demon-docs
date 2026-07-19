#!/usr/bin/env python3
"""Recreate canonical labeled audit shards from the merged benchmark."""

from __future__ import annotations

import argparse
import copy
import json
from pathlib import Path


def read(path: Path) -> dict:
    with path.open(encoding="utf-8") as stream:
        return json.load(stream)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sample", type=Path, required=True)
    parser.add_argument("--benchmark", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()

    sample = read(args.sample)
    benchmark = read(args.benchmark)
    sample_rows = sample["suggestions"]
    benchmark_rows = benchmark["suggestions"]
    if len(sample_rows) != 150 or len(benchmark_rows) != 150:
        raise ValueError("expected exactly 150 sample and benchmark rows")
    immutable = lambda row: {k: v for k, v in row.items() if k not in {"label", "rationale", "audit"}}
    if [immutable(row) for row in sample_rows] != [immutable(row) for row in benchmark_rows]:
        raise ValueError("benchmark changed sample rows")

    envelope = {key: copy.deepcopy(sample[key]) for key in ("schema_version", "corpus", "sampling")}
    args.output_dir.mkdir(parents=True, exist_ok=True)
    sizes = [18] * 8 + [6]
    start = 0
    for index, size in enumerate(sizes, 1):
        shard = copy.deepcopy(envelope)
        shard["suggestions"] = copy.deepcopy(benchmark_rows[start : start + size])
        output = args.output_dir / f"labeled-{index:02}.json"
        output.write_text(json.dumps(shard, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        start += size
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
