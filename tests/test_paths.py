from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index.paths import resolve_root


def test_resolve_root_uses_explicit_cwd_for_relative_paths() -> None:
    cwd = Path("/tmp/workspace")

    result = resolve_root("docs", cwd=cwd)

    assert result == Path("/tmp/workspace/docs")
    assert result.is_absolute()


def test_resolve_root_returns_absolute_path_for_absolute_input() -> None:
    result = resolve_root("/var/tmp/docs")

    assert result == Path("/var/tmp/docs")
    assert result.is_absolute()


def test_resolve_root_does_not_require_path_to_exist() -> None:
    result = resolve_root("missing/docs", cwd=Path("/tmp/workspace"))

    assert result == Path("/tmp/workspace/missing/docs")
    assert not result.exists()
