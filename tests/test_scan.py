from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import DraftConfig
from doc_ledger.config import FileConfig
from doc_ledger.scan import scan_docs_tree


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


def test_scan_docs_tree_uses_configured_index_file(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")

    child = root / "guide"
    child.mkdir()
    (child / "README.md").write_text("child\n", encoding="utf-8")

    tree = scan_docs_tree(root, DocLedgerConfig(index_file="README.md"))

    root_info = tree.folders[root]
    assert root_info.readme_path == root / "README.md"
    assert root_info.direct_markdown_files == [root / "guide.md"]
    assert root_info.direct_subfolders == [child]

    child_info = tree.folders[child]
    assert child_info.readme_path == child / "README.md"


def test_scan_docs_tree_uses_configured_draft_folder(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")

    drafts = root / "_drafts"
    drafts.mkdir()
    (drafts / "example.md").write_text("example\n", encoding="utf-8")
    (drafts / "nested").mkdir()

    tree = scan_docs_tree(root, DocLedgerConfig(draft=DraftConfig(folder="_drafts")))

    root_info = tree.folders[root]
    assert root_info.stub_markdown_files == [drafts / "example.md"]
    assert drafts not in root_info.direct_subfolders

    drafts_info = tree.folders[drafts]
    assert drafts_info.is_stubs is True
    assert drafts_info.readme_path is None
    assert drafts_info.direct_markdown_files == [drafts / "example.md"]


def test_scan_docs_tree_default_indexes_markdown_only(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")
    (root / "diagram.png").write_text("png\n", encoding="utf-8")
    (root / "guide.pdf").write_text("pdf\n", encoding="utf-8")

    tree = scan_docs_tree(root)

    assert tree.folders[root].direct_markdown_files == [root / "guide.md"]


def test_scan_docs_tree_can_include_non_markdown_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")
    (root / "guide.pdf").write_text("pdf\n", encoding="utf-8")
    (root / "image.png").write_text("png\n", encoding="utf-8")

    tree = scan_docs_tree(
        root,
        DocLedgerConfig(file=FileConfig(include_patterns=["**/*.md", "**/*.pdf", "**/*.png"])),
    )

    assert tree.folders[root].direct_markdown_files == [root / "guide.md", root / "guide.pdf", root / "image.png"]


def test_scan_docs_tree_can_include_arbitrary_direct_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")
    (root / "notes.txt").write_text("notes\n", encoding="utf-8")

    tree = scan_docs_tree(root, DocLedgerConfig(file=FileConfig(include_patterns=["**/*"])))

    assert tree.folders[root].direct_markdown_files == [root / "guide.md", root / "notes.txt"]


def test_scan_docs_tree_can_exclude_tmp_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")
    (root / "scratch.tmp").write_text("tmp\n", encoding="utf-8")

    tree = scan_docs_tree(
        root,
        DocLedgerConfig(file=FileConfig(include_patterns=["**/*"], exclude_patterns=["**/*.tmp"])),
    )

    assert tree.folders[root].direct_markdown_files == [root / "guide.md"]


def test_scan_docs_tree_excludes_configured_index_file_from_direct_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("root\n", encoding="utf-8")
    (root / "guide.md").write_text("guide\n", encoding="utf-8")

    tree = scan_docs_tree(root, DocLedgerConfig(index_file="README.md", file=FileConfig(include_patterns=["**/*.md", "**/*.txt"])))

    assert tree.folders[root].readme_path == root / "README.md"
    assert tree.folders[root].direct_markdown_files == [root / "guide.md"]


def test_scan_docs_tree_applies_include_exclude_rules_inside_draft_folder(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("root\n", encoding="utf-8")
    drafts = root / "stubs"
    drafts.mkdir()
    (drafts / "example.md").write_text("example\n", encoding="utf-8")
    (drafts / "diagram.png").write_text("png\n", encoding="utf-8")
    (drafts / "scratch.tmp").write_text("tmp\n", encoding="utf-8")

    tree = scan_docs_tree(
        root,
        DocLedgerConfig(file=FileConfig(include_patterns=["**/*.md", "**/*.png"], exclude_patterns=["**/*.tmp"])),
    )

    assert tree.folders[root].stub_markdown_files == [drafts / "diagram.png", drafts / "example.md"]
