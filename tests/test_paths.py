from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.paths import resolve_root


def test_resolve_root_uses_explicit_cwd_for_relative_paths(tmp_path: Path) -> None:
    cwd = tmp_path / "workspace"

    result = resolve_root("docs", cwd=cwd)

    assert result == cwd / "docs"
    assert result.is_absolute()


def test_resolve_root_returns_absolute_path_for_absolute_input(tmp_path: Path) -> None:
    absolute_root = tmp_path / "docs"

    result = resolve_root(absolute_root)

    assert result == absolute_root
    assert result.is_absolute()


def test_resolve_root_keeps_absolute_root_absolute_with_explicit_cwd(tmp_path: Path) -> None:
    absolute_root = tmp_path / "docs"

    result = resolve_root(absolute_root, cwd=tmp_path / "other-workspace")

    assert result == absolute_root
    assert result.is_absolute()


def test_resolve_root_does_not_require_path_to_exist(tmp_path: Path) -> None:
    cwd = tmp_path / "workspace"

    result = resolve_root("missing/docs", cwd=cwd)

    assert result == cwd / "missing" / "docs"
    assert not result.exists()
