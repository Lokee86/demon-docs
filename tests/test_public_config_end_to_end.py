from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger import cli


def test_public_config_end_to_end_custom_root_index_markers_and_parent_label(
    tmp_path: Path,
    monkeypatch,
) -> None:
    project = tmp_path / "project"
    notes = project / "notes"
    notes.mkdir(parents=True)
    (project / ".doc-ledger.toml").write_text(
        """
root = "notes"
index_file = "README.md"

[markers]
prefix = "navmark"

[parent_link]
label = "Parent"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (notes / "page.md").write_text("# Page\n", encoding="utf-8")
    monkeypatch.chdir(project)

    assert cli.main(["fix"]) == 0

    readme_path = notes / "README.md"
    page_path = notes / "page.md"
    assert readme_path.exists()
    assert "Parent: [Notes](./README.md)" in page_path.read_text(encoding="utf-8")

    readme_text = readme_path.read_text(encoding="utf-8")
    assert "<!-- navmark:files:start -->" in readme_text
    assert "<!-- navmark:stubs:start -->" in readme_text
    assert "<!-- navmark:folders:start -->" in readme_text

    assert cli.main(["check"]) == 0


def test_public_config_end_to_end_legacy_index_file_still_works_via_config_discovery(
    tmp_path: Path,
    monkeypatch,
) -> None:
    project = tmp_path / "project"
    notes = project / "docs"
    stubs = notes / "stubs"
    guide = notes / "guide"
    stubs.mkdir(parents=True)
    guide.mkdir(parents=True)
    (project / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (notes / "page.md").write_text("# Page\n", encoding="utf-8")
    (stubs / "draft.md").write_text("# Draft\n", encoding="utf-8")
    (guide / "topic.md").write_text("# Topic\n", encoding="utf-8")
    monkeypatch.chdir(project)

    assert cli.main(["fix"]) == 0

    root_readme = notes / "!README.md"
    guide_readme = guide / "!README.md"
    assert root_readme.exists()
    assert guide_readme.exists()

    root_readme_text = root_readme.read_text(encoding="utf-8")
    page_text = (notes / "page.md").read_text(encoding="utf-8")
    draft_text = (stubs / "draft.md").read_text(encoding="utf-8")
    topic_text = (guide / "topic.md").read_text(encoding="utf-8")

    assert "- [page.md](page.md) - Page documentation." in root_readme_text
    assert "- [guide](guide/!README.md) - Guide documentation." in root_readme_text
    assert "Parent index: [Docs](./!README.md)" in page_text
    assert "Parent index: [Docs](../!README.md)" in draft_text
    assert "Parent index: [Guide](./!README.md)" in topic_text

    assert cli.main(["check"]) == 0


def test_public_config_end_to_end_custom_draft_folder_and_section_headings(
    tmp_path: Path,
    monkeypatch,
) -> None:
    project = tmp_path / "project"
    notes = project / "notes"
    drafts = notes / "_drafts"
    drafts.mkdir(parents=True)
    (project / ".doc-ledger.toml").write_text(
        """
root = "notes"
index_file = "README.md"

[sections.files]
heading = "Files"

[sections.stubs]
heading = "Drafts"

[sections.folders]
heading = "Folders"

[drafts]
folder = "_drafts"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (drafts / "idea.md").write_text("# Idea\n", encoding="utf-8")
    monkeypatch.chdir(project)

    assert cli.main(["fix"]) == 0

    assert not (drafts / "README.md").exists()

    readme_text = (notes / "README.md").read_text(encoding="utf-8")
    assert "## Drafts" in readme_text
    assert "- [idea.md](_drafts/idea.md) - Stub: Idea documentation." in readme_text

    idea_text = (drafts / "idea.md").read_text(encoding="utf-8")
    assert "Parent index: [Notes](../README.md)" in idea_text

    folders_section = readme_text.split("## Folders", 1)[1].split("## Related Docs", 1)[0]
    assert "_drafts" not in folders_section


def test_public_config_end_to_end_indexes_non_markdown_files_without_editing_them(
    tmp_path: Path,
    monkeypatch,
) -> None:
    project = tmp_path / "project"
    notes = project / "notes"
    notes.mkdir(parents=True)
    (project / ".doc-ledger.toml").write_text(
        """
root = "notes"
index_file = "README.md"

[files]
include_patterns = ["**/*.md", "**/*.pdf", "**/*.png", "**/*.yaml"]

[editable]
parent_index_extensions = [".md", ".mdx"]
""".strip()
        + "\n",
        encoding="utf-8",
    )
    (notes / "page.md").write_text("# Page\n", encoding="utf-8")
    (notes / "diagram.png").write_bytes(b"\x89PNG\r\n\x1a\nbinary png")
    (notes / "openapi.yaml").write_bytes(b"%YAML 1.2\n---\nopenapi: 3.0.0\n")
    (notes / "manual.pdf").write_bytes(b"%PDF-1.4\nbinary pdf\n")
    monkeypatch.chdir(project)

    assert cli.main(["fix"]) == 0

    readme_text = (notes / "README.md").read_text(encoding="utf-8")
    assert "- [page.md](page.md) - Page documentation." in readme_text
    assert "- [diagram.png](diagram.png) - Diagram documentation." in readme_text
    assert "- [openapi.yaml](openapi.yaml) - Openapi documentation." in readme_text
    assert "- [manual.pdf](manual.pdf) - Manual documentation." in readme_text

    assert "Parent index: [Notes](./README.md)" in (notes / "page.md").read_text(encoding="utf-8")
    assert (notes / "diagram.png").read_bytes() == b"\x89PNG\r\n\x1a\nbinary png"
    assert (notes / "openapi.yaml").read_bytes() == b"%YAML 1.2\n---\nopenapi: 3.0.0\n"
    assert (notes / "manual.pdf").read_bytes() == b"%PDF-1.4\nbinary pdf\n"

    assert cli.main(["check"]) == 0
