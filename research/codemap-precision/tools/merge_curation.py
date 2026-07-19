#!/usr/bin/env python3
"""Validate and merge codemap precision curation shards."""

from __future__ import annotations

import argparse
import hashlib
import json
import subprocess
from pathlib import Path
from typing import Any

LABELS = {"valid_missing_link", "plausible_but_unnecessary", "incorrect"}
AUDIT_FIELDS = {
    "document_section",
    "document_ref",
    "document_excerpt",
    "target_ref",
    "target_excerpt",
    "target_sha256",
    "target_kind",
}
MUTABLE_FIELDS = {"label", "rationale", "audit"}


def read(path: Path) -> dict[str, Any]:
    with path.open(encoding="utf-8") as stream:
        value = json.load(stream)
    if not isinstance(value, dict) or not isinstance(value.get("suggestions"), list):
        raise ValueError(f"invalid benchmark JSON: {path}")
    return value


def pair(item: dict[str, Any]) -> tuple[str, str]:
    return str(item.get("document", "")), str(item.get("target", ""))


def immutable(item: dict[str, Any]) -> dict[str, Any]:
    return {key: value for key, value in item.items() if key not in MUTABLE_FIELDS}


def file_hash(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def directory_hash(root: Path, directory: Path) -> str:
    relative = directory.relative_to(root).as_posix()
    result = subprocess.run(
        ["git", "ls-files", "--", relative],
        cwd=root,
        text=True,
        capture_output=True,
        check=True,
    )
    rows: list[str] = []
    for name in sorted(line.strip() for line in result.stdout.splitlines() if line.strip()):
        child = root.joinpath(*name.split("/"))
        if child.is_file():
            rows.append(f"{name}:{file_hash(child)}")
    return hashlib.sha256("\n".join(rows).encode()).hexdigest()


def verify_target(root: Path, item: dict[str, Any]) -> None:
    audit = item["audit"]
    target = root.joinpath(*str(item["target"]).split("/"))
    kind = audit["target_kind"]
    expected = audit["target_sha256"]
    if target.is_file():
        if kind != "file":
            raise ValueError(f"{pair(item)} target kind {kind!r}, expected file")
        actual = file_hash(target)
    elif target.is_dir():
        if kind != "directory":
            raise ValueError(f"{pair(item)} target kind {kind!r}, expected directory")
        actual = directory_hash(root, target)
    else:
        if kind != "missing":
            raise ValueError(f"{pair(item)} target does not exist but kind is {kind!r}")
        actual = ""
    if expected != actual:
        raise ValueError(f"{pair(item)} target SHA mismatch: {expected!r} != {actual!r}")


def merge(source: Path, shards: list[Path], repository: Path, output: Path, reviewed_at: str) -> dict[str, int]:
    template = read(source)
    envelope = {key: template[key] for key in ("schema_version", "corpus", "sampling")}
    expected = {pair(item): item for item in template["suggestions"]}
    if len(expected) != len(template["suggestions"]):
        raise ValueError("template contains duplicate document-target pairs")

    merged: dict[tuple[str, str], dict[str, Any]] = {}
    labels = {label: 0 for label in sorted(LABELS)}
    for shard_path in shards:
        shard = read(shard_path)
        shard_metadata = {key: shard[key] for key in envelope}
        expected_metadata = dict(envelope)
        shard_metadata["corpus"] = dict(shard_metadata["corpus"])
        expected_metadata["corpus"] = dict(expected_metadata["corpus"])
        shard_metadata["corpus"]["reviewed_at"] = ""
        expected_metadata["corpus"]["reviewed_at"] = ""
        if shard_metadata != expected_metadata:
            raise ValueError(f"metadata changed in {shard_path}")
        for item in shard["suggestions"]:
            key = pair(item)
            if key not in expected:
                raise ValueError(f"unexpected suggestion in {shard_path}: {key}")
            if key in merged:
                raise ValueError(f"duplicate curated suggestion: {key}")
            if immutable(item) != immutable(expected[key]):
                raise ValueError(f"immutable fields changed for {key}")
            label = item.get("label")
            if label not in LABELS:
                raise ValueError(f"invalid or missing label for {key}: {label!r}")
            rationale = str(item.get("rationale", "")).strip()
            if len(rationale) < 20:
                raise ValueError(f"rationale too short for {key}")
            audit = item.get("audit")
            if not isinstance(audit, dict) or set(audit) != AUDIT_FIELDS:
                raise ValueError(f"invalid audit shape for {key}")
            for field in AUDIT_FIELDS - {"target_sha256"}:
                if not isinstance(audit[field], str) or not audit[field].strip():
                    raise ValueError(f"empty audit field {field} for {key}")
            verify_target(repository, item)
            labels[label] += 1
            merged[key] = item

    missing = sorted(set(expected) - set(merged))
    if missing:
        raise ValueError(f"missing {len(missing)} curated suggestions; first: {missing[:3]}")
    result = dict(envelope)
    result["corpus"] = dict(result["corpus"])
    result["corpus"]["reviewed_at"] = reviewed_at
    result["suggestions"] = [merged[pair(item)] for item in template["suggestions"]]
    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open("w", encoding="utf-8", newline="\n") as stream:
        json.dump(result, stream, indent=2, ensure_ascii=False)
        stream.write("\n")
    return labels


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--repository", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--reviewed-at", required=True)
    parser.add_argument("shards", nargs="+", type=Path)
    args = parser.parse_args()
    labels = merge(args.source, args.shards, args.repository.resolve(), args.output, args.reviewed_at)
    print(json.dumps({"output": str(args.output), "labels": labels}, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
