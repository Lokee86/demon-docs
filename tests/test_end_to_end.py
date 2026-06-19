from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index import cli


def test_fix_on_empty_docs_root_creates_root_readme(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    assert (docs_root / "!README.md").exists()


def test_new_normal_doc_updates_direct_files_and_parent_index(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "!README.md").read_text(encoding="utf-8")
    guide_text = (docs_root / "guide.md").read_text(encoding="utf-8")

    assert "- [guide.md](guide.md) - Guide documentation." in root_readme
    assert "Parent index: [Docs](./!README.md)" in guide_text


def test_new_stub_doc_updates_stub_files_and_parent_index(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir(parents=True)
    (stubs_dir / "guide.md").write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "!README.md").read_text(encoding="utf-8")
    stub_text = (stubs_dir / "guide.md").read_text(encoding="utf-8")

    assert "- [guide.md](stubs/guide.md) - Stub: Guide documentation." in root_readme
    assert "Parent index: [Docs](../!README.md)" in stub_text


def test_stub_doc_graduation_preserves_description(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir(parents=True)
    stub_path = stubs_dir / "archive.md"
    stub_path.write_text("# Archive\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    readme_path = docs_root / "!README.md"
    readme_text = readme_path.read_text(encoding="utf-8")
    readme_path.write_text(
        readme_text.replace(
            "- [archive.md](stubs/archive.md) - Stub: Archive documentation.",
            "- [archive.md](stubs/archive.md) - Stub: Legacy migration notes.",
        ),
        encoding="utf-8",
    )

    moved_path = docs_root / "archive.md"
    stub_path.rename(moved_path)

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = readme_path.read_text(encoding="utf-8")
    moved_text = moved_path.read_text(encoding="utf-8")

    assert "- [archive.md](archive.md) - Legacy migration notes." in root_readme
    assert "Parent index: [Docs](./!README.md)" in moved_text


def test_canonical_doc_moving_into_stubs_preserves_description(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir(parents=True)
    doc_path = docs_root / "guide.md"
    doc_path.write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    readme_path = docs_root / "!README.md"
    readme_text = readme_path.read_text(encoding="utf-8")
    readme_path.write_text(
        readme_text.replace(
            "- [guide.md](guide.md) - Guide documentation.",
            "- [guide.md](guide.md) - Migration guide for operators.",
        ),
        encoding="utf-8",
    )

    moved_path = stubs_dir / "guide.md"
    doc_path.rename(moved_path)

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = readme_path.read_text(encoding="utf-8")
    moved_text = moved_path.read_text(encoding="utf-8")

    assert "- [guide.md](stubs/guide.md) - Stub: Migration guide for operators." in root_readme
    assert "Parent index: [Docs](../!README.md)" in moved_text


def test_cross_folder_move_preserves_unique_description(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    alpha_dir = docs_root / "alpha"
    beta_dir = docs_root / "beta"
    alpha_dir.mkdir(parents=True)
    beta_dir.mkdir(parents=True)
    source_path = alpha_dir / "notes.md"
    source_path.write_text("# Notes\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    alpha_readme = alpha_dir / "!README.md"
    alpha_readme.write_text(
        alpha_readme.read_text(encoding="utf-8").replace(
            "- [notes.md](notes.md) - Notes documentation.",
            "- [notes.md](notes.md) - Alpha notes for migration.",
        ),
        encoding="utf-8",
    )

    moved_path = beta_dir / "notes.md"
    source_path.rename(moved_path)

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    beta_readme = beta_dir / "!README.md"
    beta_text = beta_readme.read_text(encoding="utf-8")
    moved_text = moved_path.read_text(encoding="utf-8")

    assert "- [notes.md](notes.md) - Alpha notes for migration." in beta_text
    assert "Parent index: [Beta](./!README.md)" in moved_text


def test_deleting_doc_removes_index_entry(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    guide_path = docs_root / "guide.md"
    guide_path.write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    guide_path.unlink()

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "!README.md").read_text(encoding="utf-8")
    assert "guide.md" not in root_readme


def test_fix_then_check_returns_clean(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    assert cli.main(["check", "--root", str(docs_root)]) == 0
