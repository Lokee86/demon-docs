from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index import cli
from docs_index.readme_io import make_readme_template


def test_fix_accepts_root(tmp_path: Path) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0
    assert (tmp_path / "!README.md").exists()


def test_fix_reports_summary(tmp_path: Path, capsys) -> None:
    readme_path = tmp_path / "!README.md"
    marker_block = "<!-- docs-index:files:start -->\n<!-- docs-index:files:end -->"
    readme_text = make_readme_template(tmp_path, tmp_path, None).replace(
        marker_block,
        "<!-- docs-index:files:start -->\n\n- [ghost.md](ghost.md) - Ghost documentation.\n\n<!-- docs-index:files:end -->",
    )
    readme_path.write_text(readme_text, encoding="utf-8")

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0

    output = capsys.readouterr().out
    assert "docs-index fix updated 1 file(s)" in output
    assert "docs-index fix reconciliation messages: 1" in output


def test_check_accepts_root(tmp_path: Path, capsys) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0
    capsys.readouterr()

    assert cli.main(["check", "--root", str(tmp_path)]) == 0
    output = capsys.readouterr().out
    assert output.strip() == "docs-index check passed"


def test_check_reports_failure_output(tmp_path: Path, capsys) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["check", "--root", str(tmp_path)]) == 1
    output = capsys.readouterr().out
    assert "docs-index check failed" in output
    assert str(tmp_path / "!README.md") in output


def test_fix_rejects_missing_root(tmp_path: Path) -> None:
    missing_root = tmp_path / "missing"

    try:
        cli.main(["fix", "--root", str(missing_root)])
    except SystemExit as exc:
        assert exc.code == 2
    else:  # pragma: no cover - defensive
        raise AssertionError("expected fix to fail for missing root")


def test_check_rejects_missing_root(tmp_path: Path) -> None:
    missing_root = tmp_path / "missing"

    try:
        cli.main(["check", "--root", str(missing_root)])
    except SystemExit as exc:
        assert exc.code == 2
    else:  # pragma: no cover - defensive
        raise AssertionError("expected check to fail for missing root")


def test_watch_accepts_root_and_once(tmp_path: Path) -> None:
    assert cli.main(["watch", "--root", str(tmp_path), "--once"]) == 0
