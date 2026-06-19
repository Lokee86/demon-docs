from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index.scan import scan_docs_tree


def test_scan_docs_tree_includes_every_folder_and_splits_normal_vs_stub_files(
    tmp_path: Path,
) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")
    (root / "notes.txt").write_text("ignore\n", encoding="utf-8")

    subdir = root / "guide"
    subdir.mkdir()
    (subdir / "!README.md").write_text("guide readme\n", encoding="utf-8")
    (subdir / "topic.md").write_text("topic\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "stub-a.md").write_text("stub a\n", encoding="utf-8")
    nested = stubs / "nested"
    nested.mkdir()
    (nested / "nested.md").write_text("nested\n", encoding="utf-8")

    tree = scan_docs_tree(root)

    assert tree.root == root
    assert set(tree.folders) == {root, subdir, stubs, nested}

    root_info = tree.folders[root]
    assert root_info.is_stubs is False
    assert root_info.readme_path == root / "!README.md"
    assert root_info.direct_markdown_files == [root / "guide.md"]
    assert root_info.stub_markdown_files == [stubs / "stub-a.md"]
    assert root_info.direct_subfolders == [subdir]

    subdir_info = tree.folders[subdir]
    assert subdir_info.is_stubs is False
    assert subdir_info.readme_path == subdir / "!README.md"
    assert subdir_info.direct_markdown_files == [subdir / "topic.md"]
    assert subdir_info.stub_markdown_files == []
    assert subdir_info.direct_subfolders == []

    stubs_info = tree.folders[stubs]
    assert stubs_info.is_stubs is True
    assert stubs_info.readme_path is None
    assert stubs_info.direct_markdown_files == [stubs / "stub-a.md"]
    assert stubs_info.stub_markdown_files == []
    assert stubs_info.direct_subfolders == [nested]

    nested_info = tree.folders[nested]
    assert nested_info.is_stubs is False
    assert nested_info.direct_markdown_files == [nested / "nested.md"]
