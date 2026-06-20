from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import FileConfig
from doc_ledger.scan import scan_docs_tree


def test_scan_docs_tree_includes_nested_markdown_files_with_posix_patterns(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    nested = root / "guide" / "deep"
    nested.mkdir(parents=True)
    target = nested / "setup.md"
    target.write_text("Setup body\n", encoding="utf-8")

    tree = scan_docs_tree(root, DocLedgerConfig(file=FileConfig(include_patterns=["**/*.md"])))

    assert tree.folders[root / "guide" / "deep"].direct_markdown_files == [target]
    assert (root / "README.md") not in tree.folders[root].direct_markdown_files


def test_scan_docs_tree_excludes_nested_markdown_files_with_posix_patterns(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    allowed = root / "guide" / "setup.md"
    allowed.parent.mkdir(parents=True)
    allowed.write_text("Guide body\n", encoding="utf-8")

    ignored = root / "ignored" / "deep" / "skip.md"
    ignored.parent.mkdir(parents=True)
    ignored.write_text("Ignore body\n", encoding="utf-8")

    tree = scan_docs_tree(
        root,
        DocLedgerConfig(file=FileConfig(include_patterns=["**/*.md"], exclude_patterns=["ignored/**/*.md"])),
    )

    assert tree.folders[root / "guide"].direct_markdown_files == [allowed]
    assert tree.folders[root / "ignored" / "deep"].direct_markdown_files == []
    assert ignored not in tree.folders[root / "ignored" / "deep"].direct_markdown_files
