from __future__ import annotations

from pathlib import PurePath
from pathlib import PureWindowsPath


def posix_relative_path(target_path: PurePath, base_path: PurePath) -> str:
    return _normalize_windows_extended_prefix(target_path).relative_to(
        _normalize_windows_extended_prefix(base_path)
    ).as_posix()


def _normalize_windows_extended_prefix(path: PurePath) -> PurePath:
    if isinstance(path, PureWindowsPath):
        text = str(path)
        if text.startswith("\\\\?\\"):
            return PureWindowsPath(text.removeprefix("\\\\?\\"))
    return path
