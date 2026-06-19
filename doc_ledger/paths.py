from __future__ import annotations

from pathlib import Path


def resolve_root(root_arg: str, cwd: Path | None = None) -> Path:
    base = cwd if cwd is not None else Path.cwd()
    root = Path(root_arg)
    if root.is_absolute():
        return root.resolve(strict=False)
    return (base / root).resolve(strict=False)
