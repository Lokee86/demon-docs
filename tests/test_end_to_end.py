from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger import cli


def test_fix_on_empty_docs_root_creates_root_readme(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    assert (docs_root / "README.md").exists()


def test_new_normal_doc_updates_direct_files_and_parent_index(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    guide_text = (docs_root / "guide.md").read_text(encoding="utf-8")

    assert "- [guide.md](guide.md) - Guide documentation." in root_readme
    assert "Parent index:" not in guide_text


def test_new_stub_doc_updates_stub_files_and_parent_index(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir(parents=True)
    (stubs_dir / "guide.md").write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    stub_text = (stubs_dir / "guide.md").read_text(encoding="utf-8")

    assert "- [guide.md](stubs/guide.md) - Stub: Guide documentation." in root_readme
    assert "Parent index:" not in stub_text


def test_stub_doc_graduation_preserves_description(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir(parents=True)
    stub_path = stubs_dir / "archive.md"
    stub_path.write_text("# Archive\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    readme_path = docs_root / "README.md"
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
    assert "Parent index:" not in moved_text


def test_canonical_doc_moving_into_stubs_preserves_description(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir(parents=True)
    doc_path = docs_root / "guide.md"
    doc_path.write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    readme_path = docs_root / "README.md"
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
    assert "Parent index:" not in moved_text


def test_cross_folder_move_preserves_unique_description(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    alpha_dir = docs_root / "alpha"
    beta_dir = docs_root / "beta"
    alpha_dir.mkdir(parents=True)
    beta_dir.mkdir(parents=True)
    source_path = alpha_dir / "notes.md"
    source_path.write_text("# Notes\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    alpha_readme = alpha_dir / "README.md"
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

    beta_readme = beta_dir / "README.md"
    beta_text = beta_readme.read_text(encoding="utf-8")
    moved_text = moved_path.read_text(encoding="utf-8")

    assert "- [notes.md](notes.md) - Alpha notes for migration." in beta_text
    assert "Parent index:" not in moved_text


def test_deleting_doc_removes_index_entry(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    guide_path = docs_root / "guide.md"
    guide_path.write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    guide_path.unlink()

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    assert "guide.md" not in root_readme


def test_fix_then_check_returns_clean(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    assert cli.main(["check", "--root", str(docs_root)]) == 0


def test_default_config_preserves_space_rocks_doc_conventions(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    stubs_dir = docs_root / "stubs"
    stubs_dir.mkdir()
    (stubs_dir / "example.md").write_text("# Example\n", encoding="utf-8")
    guide_dir = docs_root / "guide"
    guide_dir.mkdir()
    (guide_dir / "README.md").write_text("# Guide\n", encoding="utf-8")
    (docs_root / "README.md").write_text(
        """# Docs

## Top-Level Files
<!-- doc-ledger:files:start -->
- [page.md](page.md) - Page documentation.
<!-- doc-ledger:files:end -->

## Top-Level Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Guide documentation.
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--root", str(docs_root)]) == 0
    assert cli.main(["check", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    page_text = (docs_root / "page.md").read_text(encoding="utf-8")
    example_text = (stubs_dir / "example.md").read_text(encoding="utf-8")
    guide_text = (guide_dir / "README.md").read_text(encoding="utf-8")

    assert root_readme.startswith("# Docs")
    assert "## Top-Level Files" not in root_readme
    assert "## Top-Level Folders" not in root_readme
    assert "## Direct Files" in root_readme
    assert "## Stub Files" in root_readme
    assert "## Direct Folders" in root_readme
    assert "<!-- doc-ledger:files:start -->" in root_readme
    assert "<!-- doc-ledger:stubs:start -->" in root_readme
    assert "<!-- doc-ledger:folders:start -->" in root_readme
    assert "- [page.md](page.md) - Page documentation." in root_readme
    assert "- [example.md](stubs/example.md) - Stub: Example documentation." in root_readme
    assert "- [Guide](guide/README.md) - Guide documentation." in root_readme
    assert "Parent index:" not in page_text
    assert "Parent index:" not in example_text
    assert "Parent index: [Docs](../README.md)" in guide_text


def test_default_config_indexes_markdown_only_and_leaves_png_untouched(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    original_png = b"\x89PNG\r\n\x1a\nbinary\x00data"
    (docs_root / "diagram.png").write_bytes(original_png)

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    page_text = (docs_root / "page.md").read_text(encoding="utf-8")
    diagram_bytes = (docs_root / "diagram.png").read_bytes()

    assert "- [page.md](page.md) - Page documentation." in root_readme
    assert "diagram.png" not in root_readme
    assert "Parent index:" not in page_text
    assert diagram_bytes == original_png


def test_fix_with_configured_index_file_uses_readme_md(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")
    (docs_root / "guide").mkdir()

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
index_file = "README.md"

[parent_link]
indexed_files = true
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    guide_text = (docs_root / "guide.md").read_text(encoding="utf-8")
    guide_readme = (docs_root / "guide" / "README.md").read_text(encoding="utf-8")

    assert "- [guide.md](guide.md) - Guide documentation." in root_readme
    assert "- [guide](guide/README.md) - Guide documentation." in root_readme
    assert "README.md" not in root_readme.split("## Direct Files", 1)[1].split("## Stub Files", 1)[0]
    assert "Parent index: [Docs](./README.md)" in guide_text
    assert "Parent index: [Docs](../README.md)" in guide_readme


def test_fix_with_configured_legacy_index_file_uses_bang_readme_md(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")
    (docs_root / "guide").mkdir()

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
index_file = "!README.md"

[parent_link]
indexed_files = true
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "!README.md").read_text(encoding="utf-8")
    guide_text = (docs_root / "guide.md").read_text(encoding="utf-8")
    guide_readme = (docs_root / "guide" / "!README.md").read_text(encoding="utf-8")

    assert (docs_root / "!README.md").exists()
    assert (docs_root / "guide" / "!README.md").exists()
    assert "- [guide.md](guide.md) - Guide documentation." in root_readme
    assert "Parent index: [Docs](./!README.md)" in guide_text
    assert "Parent index: [Docs](../!README.md)" in guide_readme


def test_fix_with_folder_indexes_disabled_suppresses_child_folder_parent_links(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")
    (docs_root / "guide").mkdir()

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
index_file = "README.md"

[parent_link]
folder_indexes = false
indexed_files = true
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    guide_text = (docs_root / "guide.md").read_text(encoding="utf-8")
    guide_readme = (docs_root / "guide" / "README.md").read_text(encoding="utf-8")

    assert "Parent index: [Docs](./README.md)" in guide_text
    assert "Parent index:" not in guide_readme
    assert "- [guide.md](guide.md) - Guide documentation." in root_readme


def test_fix_with_configured_marker_prefix_uses_nav_ledgers(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text('[markers]\nprefix = "nav-ledger"\n', encoding="utf-8")

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    assert "<!-- nav-ledger:files:start -->" in root_readme
    assert "<!-- nav-ledger:stubs:start -->" in root_readme
    assert "<!-- nav-ledger:folders:start -->" in root_readme


def test_fix_with_configured_draft_folder_uses_drafts_section(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    drafts_dir = docs_root / "_drafts"
    drafts_dir.mkdir()
    (drafts_dir / "example.md").write_text("# Example\n", encoding="utf-8")

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text('[drafts]\nfolder = "_drafts"\n', encoding="utf-8")

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root)]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    example_text = (drafts_dir / "example.md").read_text(encoding="utf-8")

    assert "- [example.md](_drafts/example.md) - Stub: Example documentation." in root_readme
    assert "Parent index:" not in example_text
    assert not (drafts_dir / "README.md").exists()
    assert "_drafts" not in root_readme.split("## Direct Folders", 1)[1].split("## Related Docs", 1)[0]


def test_fix_renders_non_markdown_files_when_included(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "architecture.png").write_text("png body\n", encoding="utf-8")
    (docs_root / "openapi.yaml").write_text("yaml body\n", encoding="utf-8")

    drafts_dir = docs_root / "stubs"
    drafts_dir.mkdir()
    (drafts_dir / "draft.pdf").write_text("pdf body\n", encoding="utf-8")

    readme_path = docs_root / "README.md"
    readme_path.write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
- [architecture.png](architecture.png) - Custom architecture description.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
[files]
include_patterns = ["**/*.md", "**/*.png", "**/*.yaml", "**/*.pdf"]
""".strip(),
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root)]) == 0

    root_readme = readme_path.read_text(encoding="utf-8")
    assert "- [architecture.png](architecture.png) - Custom architecture description." in root_readme
    assert "- [openapi.yaml](openapi.yaml) - Openapi documentation." in root_readme
    assert "- [draft.pdf](stubs/draft.pdf) - Stub: Draft documentation." in root_readme
    assert "Parent index" not in (docs_root / "architecture.png").read_text(encoding="utf-8")
    assert "Parent index" not in (docs_root / "openapi.yaml").read_text(encoding="utf-8")
    assert "Parent index" not in (drafts_dir / "draft.pdf").read_text(encoding="utf-8")
