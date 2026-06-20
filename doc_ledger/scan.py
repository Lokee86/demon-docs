from __future__ import annotations

import fnmatch
from pathlib import Path

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import default_config
from doc_ledger.model import DocsTree, FolderInfo
from doc_ledger.path_format import posix_relative_path


def scan_docs_tree(root: Path, config: DocLedgerConfig | None = None) -> DocsTree:
    config = config or default_config()
    folders: dict[Path, FolderInfo] = {}
    index_file = config.index_file
    draft_folder_name = config.draft.folder
    root = root.resolve(strict=False)

    def scan_folder(folder_path: Path) -> None:
        if not folder_path.is_dir():
            return

        is_stubs = folder_path.name == draft_folder_name
        children = list(folder_path.iterdir())
        direct_markdown_files = sorted(
            child
            for child in children
            if child.is_file() and _is_indexable_file(root, child, config)
        )

        if is_stubs:
            stub_markdown_files: list[Path] = []
            direct_subfolders = sorted(child for child in children if child.is_dir())
        else:
            stub_folder = folder_path / draft_folder_name
            stub_markdown_files = (
                sorted(
                    child
                    for child in stub_folder.iterdir()
                    if child.is_file() and _is_indexable_file(root, child, config)
                )
                if stub_folder.is_dir()
                else []
            )
            direct_subfolders = sorted(
                child for child in children if child.is_dir() and child.name != draft_folder_name
            )

        folders[folder_path] = FolderInfo(
            path=folder_path,
            readme_path=None if is_stubs else folder_path / index_file,
            direct_markdown_files=direct_markdown_files,
            stub_markdown_files=stub_markdown_files,
            direct_subfolders=direct_subfolders,
            is_stubs=is_stubs,
        )

        for child in direct_subfolders:
            scan_folder(child)

        if not is_stubs:
            stub_folder = folder_path / draft_folder_name
            if stub_folder.is_dir():
                scan_folder(stub_folder)

    scan_folder(root)
    return DocsTree(root=root, folders=folders)


def _is_indexable_file(root: Path, path: Path, config: DocLedgerConfig) -> bool:
    if path.name == config.index_file:
        return False

    path_text = posix_relative_path(path.resolve(strict=False), root)

    if any(_matches_root_relative_pattern(path_text, pattern) for pattern in config.file.exclude_patterns):
        return False

    return any(_matches_root_relative_pattern(path_text, pattern) for pattern in config.file.include_patterns)


def _matches_root_relative_pattern(path_text: str, pattern: str) -> bool:
    if fnmatch.fnmatch(path_text, pattern):
        return True
    if pattern.startswith("**/") and fnmatch.fnmatch(path_text, pattern[3:]):
        return True
    return False
