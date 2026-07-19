#!/usr/bin/env python3
"""Build Demon Docs benchmark datasets from the normalized cross-repo corpus."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def target_record(repository: Path, target: str) -> dict:
    path = repository / Path(target)
    data = path.read_bytes()
    return {
        "status": "resolved",
        "resolved_path": target,
        "exists": True,
        "size": len(data),
        "sha256": digest(data),
    }


def build_dataset(benchmark_root: Path, corpus: dict) -> tuple[dict, dict]:
    candidate_id = corpus["id"]
    repository = benchmark_root / "checkouts" / candidate_id / "code"
    if corpus["code_repository"] != corpus["docs_repository"]:
        raise ValueError(f"{candidate_id}: separate docs repositories are not benchmark-input compatible")

    deduplicated: dict[tuple[str, str], dict] = {}
    for pair in corpus["pairs"]:
        key = (pair["document"], pair["resolved_target"])
        current = deduplicated.get(key)
        if current is None or pair["line"] < current["line"]:
            deduplicated[key] = pair

    by_document: dict[str, list[dict]] = {}
    for pair in deduplicated.values():
        by_document.setdefault(pair["document"], []).append(pair)

    documents = []
    entries = []
    for document in sorted(by_document):
        document_path = repository / Path(document)
        source = document_path.read_bytes()
        lines = source.decode("utf-8", errors="replace").splitlines()
        pairs = sorted(by_document[document], key=lambda item: (item["line"], item["resolved_target"]))
        documents.append({
            "path": document,
            "size": len(source),
            "sha256": digest(source),
            "entry_count": len(pairs),
            "diagnostic_count": 0,
        })
        for pair in pairs:
            line_number = pair["line"]
            raw_line = lines[line_number - 1] if 0 < line_number <= len(lines) else ""
            target = pair["resolved_target"]
            entries.append({
                "entry": {
                    "document_path": document,
                    "heading": pair["marker"],
                    "target": target,
                    "kind": "file",
                    "syntax": "bullet",
                    "source": {
                        "line": line_number,
                        "column": 1,
                        "end_line": line_number,
                        "end_column": max(1, len(raw_line.encode("utf-8"))),
                    },
                    "raw_line": raw_line,
                },
                "resolution": target_record(repository, target),
            })

    dataset = {
        "schema_version": 1,
        "documents": documents,
        "entries": entries,
        "diagnostics": [],
    }
    plan = {
        "id": candidate_id,
        "class": corpus["class"],
        "languages": corpus["languages"],
        "benchmark_mode": corpus.get("benchmark_mode", "discovery_only"),
        "holdout_count": corpus.get("holdout_count", 0),
        "code_revision": corpus["code_revision"],
        "docs_revision": corpus["docs_revision"],
        "document_count": len(documents),
        "known_link_count": len(entries),
        "dataset": f"datasets/{candidate_id}.json",
        "repository_checkout": f"checkouts/{candidate_id}/code",
    }
    return dataset, plan


def main() -> int:
    benchmark_root = Path(__file__).resolve().parents[1]
    corpus_dir = benchmark_root / "corpus"
    datasets_dir = benchmark_root / "datasets"
    datasets_dir.mkdir(parents=True, exist_ok=True)

    plans = []
    for path in sorted(corpus_dir.glob("*.json")):
        corpus = json.loads(path.read_text(encoding="utf-8"))
        if corpus.get("benchmark_mode") in {"discovery_only", "extraction_only"}:
            continue
        dataset, plan = build_dataset(benchmark_root, corpus)
        (datasets_dir / f"{corpus['id']}.json").write_text(
            json.dumps(dataset, indent=2) + "\n", encoding="utf-8"
        )
        plans.append(plan)

    manifest = {
        "schema_version": 1,
        "algorithm_baseline": "aa6eb48c686b0423e104530418b4e9fd32e3aa78",
        "benchmarks": plans,
    }
    output = benchmark_root / "benchmark-plan.json"
    output.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {len(plans)} benchmark datasets")
    print(f"wrote {output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
