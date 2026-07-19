#!/usr/bin/env python3
"""Prepare and inspect a cross-repository codemap benchmark corpus."""

from __future__ import annotations

import argparse
import fnmatch
import json
import re
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path, PurePosixPath
from typing import Iterable

PATH_TOKEN = re.compile(r"`([^`\n]+)`")
MARKDOWN_LINK = re.compile(r"\[[^\]\n]+\]\(([^)\s]+)(?:\s+['\"][^'\"]*['\"])?\)")
HEADING = re.compile(r"^\s{0,3}#{1,6}\s+(.+?)\s*#*\s*$")
PLAIN_LIST_PATH = re.compile(r"^\s*[-*+]\s+([^:|]+?)(?:\s*[:|]\s+|\s+-\s+)")
TABLE_CELL = re.compile(r"^\s*\|?\s*(`?[^|`]+`?)\s*\|")
LIKELY_PATH_SUFFIXES = {
    ".c", ".cc", ".cpp", ".cs", ".go", ".h", ".hpp", ".java", ".js", ".jsx",
    ".kt", ".md", ".mpc", ".py", ".rb", ".rs", ".sh", ".sql", ".ts", ".tsx",
    ".vue", ".yaml", ".yml", ".json", ".toml", ".xml",
}


@dataclass(frozen=True)
class Pair:
    document: str
    target: str
    marker: str
    line: int


def run(args: list[str], cwd: Path | None = None) -> str:
    command = args
    if cwd is not None and args and args[0] == "git":
        command = ["git", "-c", f"safe.directory={cwd.resolve()}", *args[1:]]
    completed = subprocess.run(
        command,
        cwd=cwd,
        check=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    return completed.stdout.strip()


def ensure_clone(url: str, destination: Path, refresh: bool) -> None:
    if refresh and destination.exists():
        shutil.rmtree(destination)
    if destination.exists():
        return
    destination.parent.mkdir(parents=True, exist_ok=True)
    run(["git", "clone", "--depth", "1", "--filter=blob:none", "--no-tags", url, str(destination)])


def revision(repository: Path) -> str:
    return run(["git", "rev-parse", "HEAD"], cwd=repository)


def tracked_files(repository: Path) -> set[str]:
    output = run(["git", "ls-files"], cwd=repository)
    return {line.replace("\\", "/") for line in output.splitlines() if line}


def expand_documents(repository: Path, patterns: Iterable[str]) -> list[Path]:
    selected: set[Path] = set()
    for pattern in patterns:
        if not any(char in pattern for char in "*?["):
            candidate = repository / Path(pattern)
            if candidate.is_file():
                selected.add(candidate)
            continue
        for candidate in repository.glob(pattern):
            if candidate.is_file() and ".git" not in candidate.parts:
                selected.add(candidate)
    return sorted(selected)


def clean_heading(value: str) -> str:
    return value.strip().rstrip("#").strip().lower()


def matches_marker(value: str, wanted: set[str]) -> str | None:
    normalized = clean_heading(value)
    for marker in sorted(wanted, key=len, reverse=True):
        if normalized == marker or normalized.startswith(marker + " ") or normalized.startswith(marker + " —"):
            return marker
    return None


def likely_path(value: str) -> str | None:
    value = value.strip().strip("'\"").replace("\\", "/")
    value = value.rstrip(".,;:)")
    if not value or "://" in value or value.startswith("$"):
        return None
    if any(ch in value for ch in "*?[]"):
        return None
    if value.startswith("./"):
        value = value[2:]
    suffix = PurePosixPath(value).suffix.lower()
    if "/" not in value and suffix not in LIKELY_PATH_SUFFIXES:
        return None
    if " " in value and suffix not in LIKELY_PATH_SUFFIXES:
        return None
    return value


def extract_pairs(document: Path, docs_root: Path, markers: list[str], extraction_mode: str) -> list[Pair]:
    relative_doc = document.relative_to(docs_root).as_posix()
    wanted = {clean_heading(marker) for marker in markers}
    active_marker = ""
    active_level = 0
    pairs: list[Pair] = []
    lines = document.read_text(encoding="utf-8", errors="replace").splitlines()

    for number, line in enumerate(lines, start=1):
        heading_match = HEADING.match(line)
        if heading_match:
            heading_text = heading_match.group(1).strip()
            level = len(line) - len(line.lstrip("#"))
            matched_marker = matches_marker(heading_text, wanted)
            if matched_marker:
                active_marker = heading_text
                active_level = level
            elif active_marker and level <= active_level:
                active_marker = ""
                active_level = 0
            continue

        stripped = line.strip()
        if matches_marker(stripped.rstrip(":"), wanted):
            active_marker = stripped.rstrip(":")
            active_level = 7
            continue
        if not active_marker:
            continue

        positioned: list[tuple[int, str]] = []
        if extraction_mode != "markdown_links":
            positioned.extend((match.start(1), match.group(1)) for match in PATH_TOKEN.finditer(line))
        positioned.extend((match.start(1), match.group(1)) for match in MARKDOWN_LINK.finditer(line))
        plain_match = PLAIN_LIST_PATH.match(line)
        if plain_match and extraction_mode != "markdown_links":
            positioned.append((plain_match.start(1), plain_match.group(1)))
        table_match = TABLE_CELL.match(line)
        if table_match and extraction_mode != "markdown_links":
            positioned.append((table_match.start(1), table_match.group(1).strip("`")))

        line_targets: list[str] = []
        for _, raw in sorted(positioned, key=lambda item: item[0]):
            target = likely_path(raw)
            if target and target not in line_targets:
                line_targets.append(target)
        if extraction_mode == "leading_entry":
            line_targets = line_targets[:1]
        for target in line_targets:
            pairs.append(Pair(relative_doc, target, active_marker, number))

    unique: dict[tuple[str, str], Pair] = {}
    for pair in pairs:
        unique[(pair.document, pair.target)] = pair
    return list(unique.values())


def resolve_target(target: str, tracked: set[str]) -> str | None:
    target = target.lstrip("/")
    if target in tracked:
        return target
    matches = sorted(path for path in tracked if path.endswith("/" + target))
    if len(matches) == 1:
        return matches[0]
    return None


def process_candidate(base: Path, candidate: dict, refresh: bool) -> dict:
    clone_root = base / "checkouts" / candidate["id"]
    code_dir = clone_root / "code"
    docs_dir = code_dir
    ensure_clone(candidate["code_repository"], code_dir, refresh)
    separate_docs = candidate["docs_repository"] != candidate["code_repository"]
    if separate_docs:
        docs_dir = clone_root / "docs"
        ensure_clone(candidate["docs_repository"], docs_dir, refresh)

    tracked = tracked_files(code_dir)
    docs = expand_documents(docs_dir, candidate["documents"])
    extracted: list[Pair] = []
    for document in docs:
        extracted.extend(extract_pairs(
            document,
            docs_dir,
            candidate["section_markers"],
            candidate.get("extraction_mode", "all_paths"),
        ))

    normalized = []
    unresolved = []
    for pair in extracted:
        resolved = resolve_target(pair.target, tracked)
        payload = {
            "document": pair.document,
            "target": pair.target,
            "marker": pair.marker,
            "line": pair.line,
        }
        if resolved:
            payload["resolved_target"] = resolved
            normalized.append(payload)
        else:
            unresolved.append(payload)

    distinct_docs = sorted({pair["document"] for pair in normalized})
    distinct_targets = sorted({pair["resolved_target"] for pair in normalized})
    status = "qualifies" if len(normalized) >= 20 and len(distinct_docs) >= 1 else "diagnostic_or_reject"
    return {
        "id": candidate["id"],
        "class": candidate["class"],
        "languages": candidate["languages"],
        "extraction_mode": candidate.get("extraction_mode", "all_paths"),
        "benchmark_mode": candidate.get("benchmark_mode", "discovery_only"),
        "holdout_count": candidate.get("holdout_count", 0),
        "status": status,
        "code_repository": candidate["code_repository"],
        "docs_repository": candidate["docs_repository"],
        "code_revision": revision(code_dir),
        "docs_revision": revision(docs_dir),
        "documents_scanned": [path.relative_to(docs_dir).as_posix() for path in docs],
        "resolved_pair_count": len(normalized),
        "resolved_document_count": len(distinct_docs),
        "resolved_target_count": len(distinct_targets),
        "unresolved_pair_count": len(unresolved),
        "pairs": sorted(normalized, key=lambda row: (row["document"], row["line"], row["resolved_target"])),
        "unresolved": sorted(unresolved, key=lambda row: (row["document"], row["line"], row["target"])),
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--refresh", action="store_true", help="delete and reclone existing checkouts")
    parser.add_argument("--candidate", action="append", default=[], help="prepare only the named candidate")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    benchmark_root = Path(__file__).resolve().parents[1]
    manifest = json.loads((benchmark_root / "candidates.json").read_text(encoding="utf-8"))
    selected = set(args.candidate)
    candidates = [item for item in manifest["candidates"] if not selected or item["id"] in selected]
    unknown = selected - {item["id"] for item in manifest["candidates"]}
    if unknown:
        print(f"unknown candidates: {', '.join(sorted(unknown))}", file=sys.stderr)
        return 2

    results = []
    for candidate in candidates:
        print(f"preparing {candidate['id']}...", flush=True)
        try:
            results.append(process_candidate(benchmark_root, candidate, args.refresh))
        except (OSError, subprocess.CalledProcessError) as error:
            results.append({
                "id": candidate["id"],
                "class": candidate["class"],
                "status": "error",
                "error": str(error),
            })

    report = {
        "schema_version": 1,
        "candidate_count": len(results),
        "qualifying_count": sum(item.get("status") == "qualifies" for item in results),
        "results": results,
    }

    runs_dir = benchmark_root / "runs"
    runs_dir.mkdir(parents=True, exist_ok=True)
    full_output = runs_dir / "discovery-full.json"
    full_output.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")

    corpus_dir = benchmark_root / "corpus"
    corpus_dir.mkdir(parents=True, exist_ok=True)
    for item in results:
        if not item.get("pairs"):
            continue
        corpus = {
            key: value
            for key, value in item.items()
            if key not in {"unresolved", "unresolved_pair_count"}
        }
        (corpus_dir / f"{item['id']}.json").write_text(
            json.dumps(corpus, indent=2) + "\n", encoding="utf-8"
        )

    compact_results = []
    for item in results:
        compact = {key: value for key, value in item.items() if key not in {"pairs", "unresolved"}}
        if item.get("unresolved"):
            compact["unresolved_examples"] = item["unresolved"][:10]
        compact_results.append(compact)
    compact_report = {
        "schema_version": 1,
        "candidate_count": len(results),
        "qualifying_count": report["qualifying_count"],
        "results": compact_results,
    }
    output = benchmark_root / "discovery.json"
    output.write_text(json.dumps(compact_report, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {output}")
    print(f"wrote {full_output}")
    return 1 if any(item.get("status") == "error" for item in results) else 0


if __name__ == "__main__":
    raise SystemExit(main())
