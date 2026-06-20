from __future__ import annotations

import sys
from pathlib import Path, PurePosixPath, PureWindowsPath


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.path_format import posix_relative_path


def test_posix_relative_path_returns_forward_slashes_for_windows_style_paths() -> None:
    result = posix_relative_path(
        PureWindowsPath("docs\\guide\\setup.md"),
        PureWindowsPath("docs"),
    )

    assert result == "guide/setup.md"
    assert "\\" not in result


def test_posix_relative_path_handles_windows_style_nested_bases() -> None:
    assert posix_relative_path(
        PureWindowsPath("docs\\guide\\README.md"),
        PureWindowsPath("docs\\guide"),
    ) == "README.md"


def test_posix_relative_path_handles_extended_windows_prefix_mismatch() -> None:
    assert posix_relative_path(
        PureWindowsPath(r"\\?\D:\repo\docs\design\architecture.md"),
        PureWindowsPath(r"D:\repo\docs"),
    ) == "design/architecture.md"


def test_posix_relative_path_handles_reverse_extended_windows_prefix_mismatch() -> None:
    assert posix_relative_path(
        PureWindowsPath(r"D:\repo\docs\design\architecture.md"),
        PureWindowsPath(r"\\?\D:\repo\docs"),
    ) == "design/architecture.md"


def test_posix_relative_path_keeps_posix_behavior_unchanged() -> None:
    assert posix_relative_path(
        PurePosixPath("docs/guide/setup.md"),
        PurePosixPath("docs"),
    ) == "guide/setup.md"
