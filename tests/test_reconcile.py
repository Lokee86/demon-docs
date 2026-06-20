from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.model import FileUpdate
from doc_ledger.model import ReconcileResult
from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import DescriptionConfig
from doc_ledger.config import DraftConfig
from doc_ledger.config import FileConfig
from doc_ledger.config import MarkerConfig
from doc_ledger.config import ParentLinkConfig
from doc_ledger.reconcile import apply_updates
from doc_ledger.reconcile import reconcile_tree


def test_reconcile_tree_returns_reconcile_result(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)

    assert isinstance(result, ReconcileResult)
    assert len(result.updates) == 1
    assert result.messages == []
    assert result.updates[0].path == root / "README.md"


def test_reconcile_tree_plans_missing_root_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    result = reconcile_tree(root)

    assert len(result.updates) == 1
    update = result.updates[0]
    assert update.path == root / "README.md"
    assert update.old_text is None
    assert update.new_text.startswith("# Docs")
    assert "Parent index:" not in update.new_text


def test_reconcile_tree_plans_missing_child_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")
    child = root / "guide"
    child.mkdir()

    result = reconcile_tree(root)

    assert {update.path for update in result.updates} == {root / "README.md", child / "README.md"}
    child_update = next(update for update in result.updates if update.path == child / "README.md")
    assert child_update.old_text is None
    assert child_update.new_text.startswith("# Guide")
    assert "Parent index: [Docs](../README.md)" in child_update.new_text


def test_reconcile_tree_does_not_add_parent_index_to_normal_markdown_by_default(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")
    page = root / "guide.md"
    page.write_text("Guide body\n", encoding="utf-8")

    result = reconcile_tree(root)

    assert all(update.path != page for update in result.updates)


def test_reconcile_tree_does_not_add_parent_index_to_stub_markdown_by_default(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")
    stubs = root / "stubs"
    stubs.mkdir()
    draft = stubs / "draft.md"
    draft.write_text("Draft body\n", encoding="utf-8")

    result = reconcile_tree(root)

    assert all(update.path != draft for update in result.updates)


def test_reconcile_tree_uses_configured_index_file(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "guide.md").write_text("Guide body\n", encoding="utf-8")
    guide = root / "guide"
    guide.mkdir()

    result = reconcile_tree(root, DocLedgerConfig(index_file="README.md"))
    updates_by_path = {update.path: update for update in result.updates}

    assert root / "README.md" in updates_by_path
    assert guide / "README.md" in updates_by_path
    assert "- [guide.md](guide.md) - Guide documentation." in updates_by_path[root / "README.md"].new_text
    assert "- [guide](guide/README.md) - Guide documentation." in updates_by_path[root / "README.md"].new_text
    assert "Parent index: [Docs](../README.md)" in updates_by_path[guide / "README.md"].new_text


def test_reconcile_tree_uses_configured_marker_prefix(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "guide.md").write_text("Guide body\n", encoding="utf-8")
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root, DocLedgerConfig(markers=MarkerConfig(prefix="nav-ledger")))
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "<!-- nav-ledger:files:start -->" in root_update.new_text
    assert "<!-- nav-ledger:stubs:start -->" in root_update.new_text
    assert "<!-- nav-ledger:folders:start -->" in root_update.new_text


def test_reconcile_tree_uses_configured_parent_link_label(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "guide.md").write_text("Guide body\n", encoding="utf-8")

    result = reconcile_tree(
        root,
        DocLedgerConfig(parent_link=ParentLinkConfig(label="Parent directory", indexed_files=True)),
    )
    guide_update = next(update for update in result.updates if update.path == root / "guide.md")

    assert "Parent directory: [Docs](./README.md)" in guide_update.new_text


def test_reconcile_tree_removes_existing_parent_lines_when_disabled(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    guide = root / "guide"
    guide.mkdir()
    (guide / "README.md").write_text(
        "# Guide\n\nParent index: [Docs](../README.md)\n\nGuide body\n",
        encoding="utf-8",
    )
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(folder_indexes=False)))
    guide_update = next(update for update in result.updates if update.path == guide / "README.md")

    assert "Parent index:" not in guide_update.new_text
    assert "Guide body" in guide_update.new_text


def test_reconcile_tree_suppresses_child_folder_parent_index_when_folder_indexes_disabled(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")
    guide = root / "guide"
    guide.mkdir()

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(folder_indexes=False)))
    guide_update = next(update for update in result.updates if update.path == guide / "README.md")

    assert "Parent index:" not in guide_update.new_text


def test_reconcile_tree_uses_configured_draft_folder(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    drafts = root / "_drafts"
    drafts.mkdir()
    (drafts / "example.md").write_text("Example body\n", encoding="utf-8")

    result = reconcile_tree(
        root,
        DocLedgerConfig(draft=DraftConfig(folder="_drafts"), parent_link=ParentLinkConfig(indexed_files=True)),
    )
    root_update = next(update for update in result.updates if update.path == root / "README.md")
    example_update = next(update for update in result.updates if update.path == drafts / "example.md")

    assert "- [example.md](_drafts/example.md) - Stub: Example documentation." in root_update.new_text
    assert "Parent index: [Docs](../README.md)" in example_update.new_text
    assert all(update.path != drafts / "README.md" for update in result.updates)


def test_reconcile_tree_uses_configured_draft_prefix(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    drafts = root / "_drafts"
    drafts.mkdir()
    (drafts / "example.md").write_text("Example body\n", encoding="utf-8")

    config = DocLedgerConfig(draft=DraftConfig(folder="_drafts", description_prefix="Draft: "))
    result = reconcile_tree(root, config)
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [example.md](_drafts/example.md) - Draft: Example documentation." in root_update.new_text


def test_reconcile_tree_keeps_existing_child_title_for_root_parent_display(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Documentation\n", encoding="utf-8")
    (root / "alpha.md").write_text("Parent index: [Docs](./README.md)\n\nAlpha body\n", encoding="utf-8")
    guide = root / "guide"
    guide.mkdir()

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))

    guide_update = next(update for update in result.updates if update.path == guide / "README.md")
    assert "Parent index: [Docs](../README.md)" in guide_update.new_text
    assert "Parent index: [Documentation](../README.md)" not in guide_update.new_text


def test_reconcile_tree_uses_root_heading_when_no_child_parent_titles_exist(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Documentation\n", encoding="utf-8")
    guide = root / "guide"
    guide.mkdir()

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))

    guide_update = next(update for update in result.updates if update.path == guide / "README.md")
    assert "Parent index: [Documentation](../README.md)" in guide_update.new_text


def test_reconcile_tree_migrates_root_top_level_sections(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Alpha body\n", encoding="utf-8")
    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "example.md").write_text("Example body\n", encoding="utf-8")
    guide = root / "guide"
    guide.mkdir()
    (guide / "README.md").write_text("# Guide\n", encoding="utf-8")
    (root / "README.md").write_text(
        """# Docs

## Top-Level Files
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.
<!-- doc-ledger:files:end -->

## Rulebook
Keep this rulebook.

## Top-Level Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Custom guide description.
<!-- doc-ledger:folders:end -->

## Related Docs
Still here.

## Notes
More notes.
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))
    updates_by_path = {update.path: update for update in result.updates}
    root_update = updates_by_path[root / "README.md"]
    example_update = updates_by_path[stubs / "example.md"]
    guide_update = updates_by_path[guide / "README.md"]

    assert "<!-- doc-ledger:files:start -->" in root_update.new_text
    assert "<!-- doc-ledger:stubs:start -->" in root_update.new_text
    assert "<!-- doc-ledger:folders:start -->" in root_update.new_text
    assert "## Top-Level Files" not in root_update.new_text
    assert "## Top-Level Folders" not in root_update.new_text
    assert "## Direct Files" in root_update.new_text
    assert root_update.new_text.count("## Direct Files") == 1
    assert root_update.new_text.count("## Stub Files") == 1
    assert "## Direct Folders" in root_update.new_text
    assert root_update.new_text.count("## Direct Folders") == 1
    assert "- [alpha.md](alpha.md) - Custom alpha description." in root_update.new_text
    assert "- [example.md](stubs/example.md) - Stub: Example documentation." in root_update.new_text
    assert "- [Guide](guide/README.md) - Custom guide description." in root_update.new_text
    assert "## Rulebook" in root_update.new_text
    assert "## Related Docs" in root_update.new_text
    assert "## Notes" in root_update.new_text
    assert "Parent index: [Docs](../README.md)" in example_update.new_text
    assert "Parent index: [Docs](../README.md)" in guide_update.new_text


def test_reconcile_tree_adds_only_missing_stub_section_for_top_level_migration(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Alpha body\n", encoding="utf-8")
    guide = root / "guide"
    guide.mkdir()
    (guide / "README.md").write_text("# Guide\n", encoding="utf-8")
    (root / "README.md").write_text(
        """# Docs

## Top-Level Files
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.
<!-- doc-ledger:files:end -->

## Rulebook
Keep this rulebook.

## Top-Level Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Custom guide description.
<!-- doc-ledger:folders:end -->

## Related Docs
Still here.

## Notes
More notes.
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert root_update.new_text.count("## Direct Files") == 1
    assert root_update.new_text.count("## Stub Files") == 1
    assert root_update.new_text.count("## Direct Folders") == 1
    assert "## Top-Level Files" not in root_update.new_text
    assert "## Top-Level Folders" not in root_update.new_text
    assert "- [alpha.md](alpha.md) - Custom alpha description." in root_update.new_text
    assert "- [Guide](guide/README.md) - Custom guide description." in root_update.new_text
    assert "## Rulebook" in root_update.new_text
    assert "## Related Docs" in root_update.new_text
    assert "## Notes" in root_update.new_text


def test_reconcile_tree_skips_stubs_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    stubs = root / "stubs"
    stubs.mkdir()

    result = reconcile_tree(root)

    assert all(update.path != stubs / "README.md" for update in result.updates)


def test_reconcile_tree_plans_stub_file_update_once(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    example = stubs / "example.md"
    example.write_text("Example body\n", encoding="utf-8")

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))
    example_updates = [update for update in result.updates if update.path == example]

    assert len(example_updates) == 1
    assert "Parent index: [Docs](../README.md)" in example_updates[0].new_text


def test_reconcile_tree_updates_parent_indexes_for_markdown_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text(
        "# Space Docs\n\nParent index: [Old](./README.md)\n\nRoot body\n",
        encoding="utf-8",
    )

    (root / "guide.md").write_text("Guide body\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "stub.md").write_text("Stub body\n", encoding="utf-8")

    guide = root / "guide"
    guide.mkdir()
    (guide / "README.md").write_text(
        "# Guide\n\nParent index: [Wrong](../README.md)\n\nGuide body\n",
        encoding="utf-8",
    )

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))
    updates_by_path = {update.path: update for update in result.updates}

    assert updates_by_path[root / "README.md"].new_text.startswith("# Space Docs\n\nRoot body")
    assert "Parent index:" not in updates_by_path[root / "README.md"].new_text
    assert "Parent index: [Space Docs](./README.md)" in updates_by_path[root / "guide.md"].new_text
    assert "Parent index: [Space Docs](../README.md)" in updates_by_path[stubs / "stub.md"].new_text
    assert "Parent index: [Space Docs](../README.md)" in updates_by_path[guide / "README.md"].new_text


def test_reconcile_tree_renders_non_markdown_files_when_included(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "architecture.png").write_text("png body\n", encoding="utf-8")
    (root / "openapi.yaml").write_text("yaml body\n", encoding="utf-8")

    drafts = root / "stubs"
    drafts.mkdir()
    (drafts / "draft.pdf").write_text("pdf body\n", encoding="utf-8")

    (root / "README.md").write_text(
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

    result = reconcile_tree(
        root,
        DocLedgerConfig(file=FileConfig(include_patterns=["**/*.md", "**/*.png", "**/*.yaml", "**/*.pdf"])),
    )
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [architecture.png](architecture.png) - Custom architecture description." in root_update.new_text
    assert "- [openapi.yaml](openapi.yaml) - Openapi documentation." in root_update.new_text
    assert "- [draft.pdf](stubs/draft.pdf) - Stub: Draft documentation." in root_update.new_text
    assert "Parent index" not in (root / "architecture.png").read_text(encoding="utf-8")
    assert "Parent index" not in (root / "openapi.yaml").read_text(encoding="utf-8")
    assert "Parent index" not in (drafts / "draft.pdf").read_text(encoding="utf-8")


def test_reconcile_tree_uses_forward_slashes_for_generated_link_targets(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")
    (root / "alpha.md").write_text("Alpha body\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "draft.md").write_text("Draft body\n", encoding="utf-8")

    guide = root / "guide"
    guide.mkdir()

    result = reconcile_tree(root)
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [alpha.md](alpha.md) - Alpha documentation." in root_update.new_text
    assert "- [draft.md](stubs/draft.md) - Stub: Draft documentation." in root_update.new_text
    assert "- [guide](guide/README.md) - Guide documentation." in root_update.new_text
    assert "\\" not in root_update.new_text


def test_reconcile_tree_does_not_read_or_rewrite_non_editable_included_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    diagram = root / "diagram.png"
    diagram.write_bytes(b"\x89PNG\r\n\x1a\nbinary png data")
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root, DocLedgerConfig(file=FileConfig(include_patterns=["**/*.md", "**/*.png"])))

    assert all(update.path != diagram for update in result.updates)
    root_update = next(update for update in result.updates if update.path == root / "README.md")
    assert "- [diagram.png](diagram.png) - Diagram documentation." in root_update.new_text


def test_reconcile_tree_updates_parent_index_for_configured_editable_extension(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    page = root / "page.mdx"
    page.write_text("# Page\n", encoding="utf-8")

    config = DocLedgerConfig(
        file=FileConfig(include_patterns=["**/*.md", "**/*.mdx"]),
        parent_link=ParentLinkConfig(indexed_files=True),
    )
    config.file.editable_parent_index_extensions = [".md", ".mdx"]

    result = reconcile_tree(root, config)
    updates_by_path = {update.path: update for update in result.updates}

    assert "Parent index: [Docs](./README.md)" in updates_by_path[page].new_text
    assert "- [page.mdx](page.mdx) - Page documentation." in updates_by_path[root / "README.md"].new_text


def test_reconcile_tree_populates_managed_sections(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")
    (root / "alpha.md").write_text("Alpha body\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "stub.md").write_text("Stub body\n", encoding="utf-8")

    child = root / "guide"
    child.mkdir()
    (child / "topic.md").write_text("Topic body\n", encoding="utf-8")

    child_readme = child / "README.md"
    child_readme.write_text(
        "# Guide\n\n## Direct Files\n<!-- doc-ledger:files:start -->\n<!-- doc-ledger:files:end -->\n\n## Stub Files\n<!-- doc-ledger:stubs:start -->\n<!-- doc-ledger:stubs:end -->\n\n## Direct Folders\n<!-- doc-ledger:folders:start -->\n<!-- doc-ledger:folders:end -->\n",
        encoding="utf-8",
    )

    result = reconcile_tree(root)
    updates_by_path = {update.path: update for update in result.updates}

    assert "- [alpha.md](alpha.md) - Alpha documentation." in updates_by_path[root / "README.md"].new_text
    assert "- [stub.md](stubs/stub.md) - Stub: Stub documentation." in updates_by_path[root / "README.md"].new_text
    assert "- [guide](guide/README.md) - Guide documentation." in updates_by_path[root / "README.md"].new_text
    assert "doc-ledger:files:start" in updates_by_path[child_readme].new_text
    assert "doc-ledger:stubs:start" in updates_by_path[child_readme].new_text
    assert "doc-ledger:folders:start" in updates_by_path[child_readme].new_text


def test_reconcile_tree_uses_configured_file_description_template_for_new_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Alpha body\n", encoding="utf-8")

    result = reconcile_tree(
        root,
        DocLedgerConfig(description=DescriptionConfig(file_template="File: {title}.")),
    )
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [alpha.md](alpha.md) - File: Alpha." in root_update.new_text


def test_reconcile_tree_uses_configured_folder_description_template_for_new_folders(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    child = root / "guide-folder"
    child.mkdir()

    result = reconcile_tree(
        root,
        DocLedgerConfig(description=DescriptionConfig(folder_template="Folder: {title}.")),
    )
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [guide-folder](guide-folder/README.md) - Folder: Guide Folder." in root_update.new_text


def test_reconcile_tree_uses_configured_draft_prefix_for_new_stub_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "draft.md").write_text("Draft body\n", encoding="utf-8")

    result = reconcile_tree(root, DocLedgerConfig(draft=DraftConfig(description_prefix="Draft: ")))
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [draft.md](stubs/draft.md) - Draft: Draft documentation." in root_update.new_text


def test_reconcile_tree_preserves_existing_description_with_configured_fallbacks(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Parent index: [Docs](./README.md)\n\nAlpha body", encoding="utf-8")
    (root / "README.md").write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.
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

    result = reconcile_tree(
        root,
        DocLedgerConfig(description=DescriptionConfig(file_template="File: {title}.")),
    )
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [alpha.md](alpha.md) - Custom alpha description." in root_update.new_text
    assert "File: Alpha." not in root_update.new_text


def test_fix_preserves_existing_direct_file_description(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Parent index: [Docs](./README.md)\n\nAlpha body", encoding="utf-8")
    (root / "README.md").write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.
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

    assert apply_updates(reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))) == 1

    readme_text = (root / "README.md").read_text(encoding="utf-8")
    assert "Custom alpha description." in readme_text
    assert "Alpha documentation." not in readme_text


def test_reconcile_tree_promotes_stub_description_when_file_graduates(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "foo.md").write_text("Foo body\n", encoding="utf-8")
    (root / "README.md").write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
- [foo.md](stubs/foo.md) - Stub: lower-case promoted description.
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root)
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [foo.md](foo.md) - Lower-case promoted description." in root_update.new_text
    assert "stubs/foo.md" not in root_update.new_text
    assert "Stub: " not in root_update.new_text


def test_reconcile_tree_preserves_description_when_file_moves_into_stubs(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "stubs").mkdir()
    (root / "stubs" / "foo.md").write_text("Foo body\n", encoding="utf-8")
    (root / "README.md").write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
- [foo.md](foo.md) - Custom foo description.
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

    result = reconcile_tree(root)
    root_update = next(update for update in result.updates if update.path == root / "README.md")

    assert "- [foo.md](stubs/foo.md) - Stub: Custom foo description." in root_update.new_text
    assert "foo.md](foo.md) - Custom foo description." not in root_update.new_text


def test_reconcile_tree_preserves_description_when_file_moves_across_folders(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "README.md").write_text(
        """# Alpha

## Direct Files
<!-- doc-ledger:files:start -->
- [foo.md](foo.md) - Custom alpha description.
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

    beta = root / "beta"
    beta.mkdir()
    (beta / "foo.md").write_text("Foo body\n", encoding="utf-8")

    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    beta_update = next(update for update in result.updates if update.path == beta / "README.md")

    assert "- [foo.md](foo.md) - Custom alpha description." in beta_update.new_text
    assert "Alpha documentation." not in beta_update.new_text


def test_reconcile_tree_does_not_reuse_stale_description_for_ambiguous_file_moves(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "README.md").write_text(
        """# Alpha

## Direct Files
<!-- doc-ledger:files:start -->
- [foo.md](foo.md) - Custom alpha description.
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

    beta = root / "beta"
    beta.mkdir()
    (beta / "foo.md").write_text("Foo body\n", encoding="utf-8")

    gamma = root / "gamma"
    gamma.mkdir()
    (gamma / "foo.md").write_text("Foo body\n", encoding="utf-8")

    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    updates_by_path = {update.path: update for update in result.updates}

    assert "- [foo.md](foo.md) - Foo documentation." in updates_by_path[beta / "README.md"].new_text
    assert "- [foo.md](foo.md) - Foo documentation." in updates_by_path[gamma / "README.md"].new_text
    assert "Custom alpha description." not in updates_by_path[beta / "README.md"].new_text
    assert "Custom alpha description." not in updates_by_path[gamma / "README.md"].new_text


def test_reconcile_tree_uses_generated_fallback_when_same_filename_is_ambiguous(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "README.md").write_text(
        """# Alpha

## Direct Files
<!-- doc-ledger:files:start -->
- [foo.md](foo.md) - Custom alpha description.
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

    beta = root / "beta"
    beta.mkdir()
    (beta / "README.md").write_text(
        """# Beta

## Direct Files
<!-- doc-ledger:files:start -->
- [foo.md](foo.md) - Custom beta description.
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

    gamma = root / "gamma"
    gamma.mkdir()
    (gamma / "foo.md").write_text("Foo body\n", encoding="utf-8")
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    gamma_update = next(update for update in result.updates if update.path == gamma / "README.md")

    assert "- [foo.md](foo.md) - Foo documentation." in gamma_update.new_text
    assert "Custom alpha description." not in gamma_update.new_text
    assert "Custom beta description." not in gamma_update.new_text


def test_reconcile_tree_preserves_description_when_folder_moves_across_folders(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "README.md").write_text(
        """# Alpha

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Custom guide description.
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "guide").mkdir()
    (beta / "guide" / "README.md").write_text("# Guide\n", encoding="utf-8")

    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    beta_update = next(update for update in result.updates if update.path == beta / "README.md")

    assert "- [Guide](guide/README.md) - Custom guide description." in beta_update.new_text
    assert "Guide documentation." not in beta_update.new_text


def test_reconcile_tree_does_not_reuse_stale_description_for_ambiguous_folder_moves(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "README.md").write_text(
        """# Alpha

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Custom alpha guide description.
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "guide").mkdir()
    (beta / "guide" / "README.md").write_text("# Guide\n", encoding="utf-8")

    gamma = root / "gamma"
    gamma.mkdir()
    (gamma / "guide").mkdir()
    (gamma / "guide" / "README.md").write_text("# Guide\n", encoding="utf-8")

    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    updates_by_path = {update.path: update for update in result.updates}

    assert "- [guide](guide/README.md) - Guide documentation." in updates_by_path[beta / "README.md"].new_text
    assert "- [guide](guide/README.md) - Guide documentation." in updates_by_path[gamma / "README.md"].new_text
    assert "Custom alpha guide description." not in updates_by_path[beta / "README.md"].new_text
    assert "Custom alpha guide description." not in updates_by_path[gamma / "README.md"].new_text


def test_reconcile_tree_uses_generated_fallback_when_folder_name_is_ambiguous(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "README.md").write_text(
        """# Alpha

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Custom alpha guide description.
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "README.md").write_text(
        """# Beta

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Custom beta guide description.
<!-- doc-ledger:folders:end -->
""",
        encoding="utf-8",
    )

    gamma = root / "gamma"
    gamma.mkdir()
    (gamma / "guide").mkdir()
    (gamma / "guide" / "README.md").write_text("# Guide\n", encoding="utf-8")
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    gamma_update = next(update for update in result.updates if update.path == gamma / "README.md")

    assert "- [guide](guide/README.md) - Guide documentation." in gamma_update.new_text
    assert "Custom alpha guide description." not in gamma_update.new_text
    assert "Custom beta guide description." not in gamma_update.new_text


def test_reconcile_tree_keeps_stub_folders_excluded_from_direct_folder_entries(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "README.md").write_text("# Docs\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "README.md").write_text("# Stubs\n", encoding="utf-8")

    result = reconcile_tree(root)

    root_update = next(update for update in result.updates if update.path == root / "README.md")
    assert "stubs/README.md" not in root_update.new_text
    assert "## Direct Folders" in root_update.new_text


def test_reconcile_tree_removes_stale_managed_entries(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    readme = root / "README.md"
    readme.write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
- [gone.md](gone.md) - Gone direct doc.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
- [stale.md](stubs/stale.md) - Stub: Gone stub doc.
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
- [Gone](gone/README.md) - Gone folder docs.
<!-- doc-ledger:folders:end -->

## Notes

Keep this note.
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root)

    assert len(result.updates) == 1
    assert len(result.messages) == 3
    assert any("Removed stale files entry" in message for message in result.messages)
    assert any("Removed stale stubs entry" in message for message in result.messages)
    assert any("Removed stale folders entry" in message for message in result.messages)

    assert apply_updates(result) == 1
    rewritten = readme.read_text(encoding="utf-8")
    assert "gone.md" not in rewritten
    assert "stale.md" not in rewritten
    assert "gone/README.md" not in rewritten
    assert "Keep this note." in rewritten


def test_reconcile_tree_repairs_missing_end_marker(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Parent index: [Docs](./README.md)\n\nAlpha body", encoding="utf-8")
    readme = root / "README.md"
    readme.write_text(
        """# Docs

## Direct Files
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.

## Notes

Keep this note.
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root, DocLedgerConfig(parent_link=ParentLinkConfig(indexed_files=True)))

    assert len(result.updates) == 1
    assert apply_updates(result) == 1
    rewritten = readme.read_text(encoding="utf-8")
    assert "<!-- doc-ledger:files:end -->" in rewritten
    assert "Custom alpha description." in rewritten
    assert "Keep this note." in rewritten


def test_apply_updates_creates_missing_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    result = reconcile_tree(root)

    assert apply_updates(result) == 1
    assert (root / "README.md").exists()


def test_apply_updates_skips_unchanged_existing_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    readme = root / "README.md"
    readme.write_text("# Docs\n", encoding="utf-8")

    result = ReconcileResult(
        updates=[FileUpdate(path=readme, old_text="# Docs\n", new_text="# Docs\n")],
        messages=[],
    )

    assert apply_updates(result) == 0
    assert readme.read_text(encoding="utf-8") == "# Docs\n"
