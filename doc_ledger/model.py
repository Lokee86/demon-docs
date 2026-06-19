from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path


@dataclass
class FolderInfo:
    path: Path
    readme_path: Path | None
    direct_markdown_files: list[Path]
    stub_markdown_files: list[Path]
    direct_subfolders: list[Path]
    is_stubs: bool


@dataclass
class DocsTree:
    root: Path
    folders: dict[Path, FolderInfo]


@dataclass
class IndexEntry:
    readme_path: Path
    section: str
    link_text: str
    link_target: str
    description: str
    original_line: str


@dataclass
class FileUpdate:
    path: Path
    old_text: str | None
    new_text: str


@dataclass
class ReconcileResult:
    updates: list[FileUpdate]
    messages: list[str]
