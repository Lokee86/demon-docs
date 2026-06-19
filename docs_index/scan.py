from __future__ import annotations

from pathlib import Path

from docs_index.model import DocsTree, FolderInfo


def scan_docs_tree(root: Path) -> DocsTree:
    folders: dict[Path, FolderInfo] = {}

    def scan_folder(folder_path: Path) -> None:
        if not folder_path.is_dir():
            return

        is_stubs = folder_path.name == "stubs"
        children = list(folder_path.iterdir())
        direct_markdown_files = sorted(
            child
            for child in children
            if child.is_file() and child.suffix == ".md" and child.name != "!README.md"
        )

        if is_stubs:
            stub_markdown_files: list[Path] = []
            direct_subfolders = sorted(child for child in children if child.is_dir())
        else:
            stub_folder = folder_path / "stubs"
            stub_markdown_files = (
                sorted(
                    child
                    for child in stub_folder.iterdir()
                    if child.is_file() and child.suffix == ".md" and child.name != "!README.md"
                )
                if stub_folder.is_dir()
                else []
            )
            direct_subfolders = sorted(
                child for child in children if child.is_dir() and child.name != "stubs"
            )

        folders[folder_path] = FolderInfo(
            path=folder_path,
            readme_path=None if is_stubs else folder_path / "!README.md",
            direct_markdown_files=direct_markdown_files,
            stub_markdown_files=stub_markdown_files,
            direct_subfolders=direct_subfolders,
            is_stubs=is_stubs,
        )

        for child in direct_subfolders:
            scan_folder(child)

        if not is_stubs:
            stub_folder = folder_path / "stubs"
            if stub_folder.is_dir():
                scan_folder(stub_folder)

    scan_folder(root)
    return DocsTree(root=root, folders=folders)
