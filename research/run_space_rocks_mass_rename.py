from __future__ import annotations

import json
import re
import shutil
import subprocess
import tempfile
from collections import Counter
from pathlib import Path

WORKTREE = Path(r"D:\!bin\demon-docs-mass-rename-tests")
SOURCE = Path(r"D:\!bin\space-rocks\docs")
BINARY = WORKTREE / "bin" / "ddocs.exe"
REPORT_DIR = WORKTREE / "research" / "mass-rename-results"
Path(r"C:\tmp").mkdir(parents=True, exist_ok=True)
TARGET = Path(tempfile.mkdtemp(prefix="demon-docs-space-rocks-mass-rename-", dir=r"C:\tmp"))


def run(name: str, *args: str) -> dict[str, object]:
    result = subprocess.run(
        [str(BINARY), *args], cwd=TARGET, text=True, capture_output=True,
        encoding="utf-8", errors="replace",
    )
    text = result.stdout + result.stderr
    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    (REPORT_DIR / f"{name}.txt").write_text(text, encoding="utf-8")
    updated_match = re.search(r"ddocs fix updated (\d+) file\(s\)", text)
    unresolved_match = re.search(r"unresolved (\d+) link\(s\)", text)
    update_paths = [line for line in text.splitlines() if re.match(r"^[A-Za-z]:\\", line)]
    return {
        "name": name,
        "exit_code": result.returncode,
        "updated": int(updated_match.group(1)) if updated_match else None,
        "reported_update_paths": len(update_paths),
        "repairs": text.count("Repair link in "),
        "broken": text.count("Broken link in "),
        "ambiguous": text.count("Ambiguous link in "),
        "undefined_reference": text.count("Undefined reference label in "),
        "unresolved_summary": int(unresolved_match.group(1)) if unresolved_match else None,
        "errors": [line for line in text.splitlines() if line.startswith("ddocs error:")],
    }


def rename_all_markdown(suffix: str) -> dict[str, str]:
    markdown = sorted(TARGET.rglob("*.md"), key=lambda path: path.as_posix())
    mapping: dict[str, str] = {}
    temporary: list[tuple[Path, Path]] = []
    for index, old in enumerate(markdown):
        final = old.with_name(f"{old.stem}{suffix}{old.suffix}")
        temp = old.with_name(f".__ddocs_mass_rename_{index:04d}__.tmp")
        mapping[old.relative_to(TARGET).as_posix()] = final.relative_to(TARGET).as_posix()
        old.rename(temp)
        temporary.append((temp, final))
    for temp, final in temporary:
        temp.rename(final)
    return mapping


def remaining_old_markdown_names(mapping: dict[str, str]) -> list[dict[str, object]]:
    old_names = {Path(old).name for old in mapping}
    counts: Counter[str] = Counter()
    files: dict[str, list[str]] = {}
    name_pattern = re.compile(r"[A-Za-z0-9_!.-]+\.md")
    for path in TARGET.rglob("*.md"):
        text = path.read_text(encoding="utf-8", errors="replace")
        for name in name_pattern.findall(text):
            if name in old_names:
                counts[name] += 1
                files.setdefault(name, []).append(path.relative_to(TARGET).as_posix())
    return [
        {"basename": name, "count": count, "files": files[name][:5]}
        for name, count in counts.most_common(25)
    ]


def main() -> None:
    shutil.rmtree(REPORT_DIR, ignore_errors=True)
    shutil.copytree(SOURCE, TARGET, dirs_exist_ok=True)
    steps: list[dict[str, object]] = []

    steps.append(run("01-init", "init", "--root", "."))
    steps.append(run("02-baseline-initialize", "fix", "-l"))
    steps.append(run("03-baseline-repair", "fix", "-l"))
    steps.append(run("04-baseline-idempotent", "fix", "-l"))
    steps.append(run("05-baseline-check", "check", "-l"))

    first_map = rename_all_markdown("-mass-renamed")
    steps.append(run("06-first-rename-prefixed-check", "check", "-l"))
    steps.append(run("07-first-rename-fix", "fix", "-l"))
    steps.append(run("08-first-rename-post-check", "check", "-l"))
    steps.append(run("09-first-rename-idempotent", "fix", "-l"))

    second_map = rename_all_markdown("-again")
    steps.append(run("10-second-rename-prefixed-check", "check", "-l"))
    steps.append(run("11-second-rename-fix", "fix", "-l"))
    steps.append(run("12-second-rename-post-check", "check", "-l"))
    steps.append(run("13-second-rename-idempotent", "fix", "-l"))

    baseline, first_post, second_post = steps[4], steps[7], steps[11]
    summary = {
        "sandbox": str(TARGET),
        "source_files": sum(1 for path in SOURCE.rglob("*") if path.is_file()),
        "first_markdown_renamed": len(first_map),
        "second_markdown_renamed": len(second_map),
        "final_markdown_files_present": sum(1 for _ in TARGET.rglob("*.md")),
        "old_files_remaining_after_first": [old for old in first_map if (TARGET / old).exists()],
        "old_files_remaining_after_second": [old for old in second_map if (TARGET / old).exists()],
        "steps": steps,
        "unresolved_comparison": {
            "baseline": {"broken": baseline["broken"], "ambiguous": baseline["ambiguous"]},
            "first_post": {"broken": first_post["broken"], "ambiguous": first_post["ambiguous"]},
            "second_post": {"broken": second_post["broken"], "ambiguous": second_post["ambiguous"]},
        },
        "remaining_first_pass_names": remaining_old_markdown_names(first_map),
        "remaining_second_pass_names": remaining_old_markdown_names(second_map),
    }
    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    (REPORT_DIR / "first-rename-map.json").write_text(json.dumps(first_map, indent=2), encoding="utf-8")
    (REPORT_DIR / "second-rename-map.json").write_text(json.dumps(second_map, indent=2), encoding="utf-8")
    (REPORT_DIR / "summary.json").write_text(json.dumps(summary, indent=2), encoding="utf-8")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
