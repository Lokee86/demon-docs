from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index.model import FileUpdate
from docs_index.model import ReconcileResult
from docs_index.reconcile import apply_updates
from docs_index.reconcile import reconcile_tree


def test_reconcile_tree_returns_reconcile_result(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)

    assert isinstance(result, ReconcileResult)
    assert len(result.updates) == 1
    assert result.messages == []
    assert result.updates[0].path == root / "!README.md"


def test_reconcile_tree_plans_missing_root_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    result = reconcile_tree(root)

    assert len(result.updates) == 1
    update = result.updates[0]
    assert update.path == root / "!README.md"
    assert update.old_text is None
    assert update.new_text.startswith("# Docs")
    assert "Parent index:" not in update.new_text


def test_reconcile_tree_plans_missing_child_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")
    child = root / "guide"
    child.mkdir()

    result = reconcile_tree(root)

    assert {update.path for update in result.updates} == {root / "!README.md", child / "!README.md"}
    child_update = next(update for update in result.updates if update.path == child / "!README.md")
    assert child_update.old_text is None
    assert child_update.new_text.startswith("# Guide")
    assert "Parent index: [Docs](../!README.md)" in child_update.new_text


def test_reconcile_tree_skips_stubs_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    stubs = root / "stubs"
    stubs.mkdir()

    result = reconcile_tree(root)

    assert all(update.path != stubs / "!README.md" for update in result.updates)


def test_reconcile_tree_updates_parent_indexes_for_markdown_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text(
        "# Space Docs\n\nParent index: [Old](./!README.md)\n\nRoot body\n",
        encoding="utf-8",
    )

    (root / "guide.md").write_text("Guide body\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "stub.md").write_text("Stub body\n", encoding="utf-8")

    guide = root / "guide"
    guide.mkdir()
    (guide / "!README.md").write_text(
        "# Guide\n\nParent index: [Wrong](../!README.md)\n\nGuide body\n",
        encoding="utf-8",
    )

    result = reconcile_tree(root)
    updates_by_path = {update.path: update for update in result.updates}

    assert updates_by_path[root / "!README.md"].new_text.startswith("# Space Docs\n\nRoot body")
    assert "Parent index:" not in updates_by_path[root / "!README.md"].new_text
    assert "Parent index: [Space Docs](./!README.md)" in updates_by_path[root / "guide.md"].new_text
    assert "Parent index: [Space Docs](../!README.md)" in updates_by_path[stubs / "stub.md"].new_text
    assert "Parent index: [Space Docs](../!README.md)" in updates_by_path[guide / "!README.md"].new_text


def test_reconcile_tree_populates_managed_sections(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")
    (root / "alpha.md").write_text("Alpha body\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "stub.md").write_text("Stub body\n", encoding="utf-8")

    child = root / "guide"
    child.mkdir()
    (child / "topic.md").write_text("Topic body\n", encoding="utf-8")

    child_readme = child / "!README.md"
    child_readme.write_text(
        "# Guide\n\n## Direct Files\n<!-- docs-index:files:start -->\n<!-- docs-index:files:end -->\n\n## Stub Files\n<!-- docs-index:stubs:start -->\n<!-- docs-index:stubs:end -->\n\n## Direct Folders\n<!-- docs-index:folders:start -->\n<!-- docs-index:folders:end -->\n",
        encoding="utf-8",
    )

    result = reconcile_tree(root)
    updates_by_path = {update.path: update for update in result.updates}

    assert "- [alpha.md](alpha.md) - Alpha documentation." in updates_by_path[root / "!README.md"].new_text
    assert "- [stub.md](stubs/stub.md) - Stub: Stub documentation." in updates_by_path[root / "!README.md"].new_text
    assert "- [guide](guide/!README.md) - Guide documentation." in updates_by_path[root / "!README.md"].new_text
    assert "docs-index:files:start" in updates_by_path[child_readme].new_text
    assert "docs-index:stubs:start" in updates_by_path[child_readme].new_text
    assert "docs-index:folders:start" in updates_by_path[child_readme].new_text


def test_fix_preserves_existing_direct_file_description(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Parent index: [Docs](./!README.md)\n\nAlpha body", encoding="utf-8")
    (root / "!README.md").write_text(
        """# Docs

## Direct Files
<!-- docs-index:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    assert apply_updates(reconcile_tree(root)) == 1

    readme_text = (root / "!README.md").read_text(encoding="utf-8")
    assert "Custom alpha description." in readme_text
    assert "Alpha documentation." not in readme_text


def test_reconcile_tree_promotes_stub_description_when_file_graduates(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "foo.md").write_text("Foo body\n", encoding="utf-8")
    (root / "!README.md").write_text(
        """# Docs

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
- [foo.md](stubs/foo.md) - Stub: lower-case promoted description.
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root)
    root_update = next(update for update in result.updates if update.path == root / "!README.md")

    assert "- [foo.md](foo.md) - Lower-case promoted description." in root_update.new_text
    assert "stubs/foo.md" not in root_update.new_text
    assert "Stub: " not in root_update.new_text


def test_reconcile_tree_preserves_description_when_file_moves_into_stubs(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "stubs").mkdir()
    (root / "stubs" / "foo.md").write_text("Foo body\n", encoding="utf-8")
    (root / "!README.md").write_text(
        """# Docs

## Direct Files
<!-- docs-index:files:start -->
- [foo.md](foo.md) - Custom foo description.
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root)
    root_update = next(update for update in result.updates if update.path == root / "!README.md")

    assert "- [foo.md](stubs/foo.md) - Stub: Custom foo description." in root_update.new_text
    assert "foo.md](foo.md) - Custom foo description." not in root_update.new_text


def test_reconcile_tree_preserves_description_when_file_moves_across_folders(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "!README.md").write_text(
        """# Alpha

## Direct Files
<!-- docs-index:files:start -->
- [foo.md](foo.md) - Custom alpha description.
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "foo.md").write_text("Foo body\n", encoding="utf-8")

    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    beta_update = next(update for update in result.updates if update.path == beta / "!README.md")

    assert "- [foo.md](foo.md) - Custom alpha description." in beta_update.new_text
    assert "Alpha documentation." not in beta_update.new_text


def test_reconcile_tree_uses_generated_fallback_when_same_filename_is_ambiguous(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "!README.md").write_text(
        """# Alpha

## Direct Files
<!-- docs-index:files:start -->
- [foo.md](foo.md) - Custom alpha description.
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "!README.md").write_text(
        """# Beta

## Direct Files
<!-- docs-index:files:start -->
- [foo.md](foo.md) - Custom beta description.
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    gamma = root / "gamma"
    gamma.mkdir()
    (gamma / "foo.md").write_text("Foo body\n", encoding="utf-8")
    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    gamma_update = next(update for update in result.updates if update.path == gamma / "!README.md")

    assert "- [foo.md](foo.md) - Foo documentation." in gamma_update.new_text
    assert "Custom alpha description." not in gamma_update.new_text
    assert "Custom beta description." not in gamma_update.new_text


def test_reconcile_tree_preserves_description_when_folder_moves_across_folders(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "!README.md").write_text(
        """# Alpha

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
- [Guide](guide/!README.md) - Custom guide description.
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "guide").mkdir()
    (beta / "guide" / "!README.md").write_text("# Guide\n", encoding="utf-8")

    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    beta_update = next(update for update in result.updates if update.path == beta / "!README.md")

    assert "- [Guide](guide/!README.md) - Custom guide description." in beta_update.new_text
    assert "Guide documentation." not in beta_update.new_text


def test_reconcile_tree_uses_generated_fallback_when_folder_name_is_ambiguous(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    alpha = root / "alpha"
    alpha.mkdir()
    (alpha / "!README.md").write_text(
        """# Alpha

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
- [Guide](guide/!README.md) - Custom alpha guide description.
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    beta = root / "beta"
    beta.mkdir()
    (beta / "!README.md").write_text(
        """# Beta

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
- [Guide](guide/!README.md) - Custom beta guide description.
<!-- docs-index:folders:end -->
""",
        encoding="utf-8",
    )

    gamma = root / "gamma"
    gamma.mkdir()
    (gamma / "guide").mkdir()
    (gamma / "guide" / "!README.md").write_text("# Guide\n", encoding="utf-8")
    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")

    result = reconcile_tree(root)
    gamma_update = next(update for update in result.updates if update.path == gamma / "!README.md")

    assert "- [guide](guide/!README.md) - Guide documentation." in gamma_update.new_text
    assert "Custom alpha guide description." not in gamma_update.new_text
    assert "Custom beta guide description." not in gamma_update.new_text


def test_reconcile_tree_keeps_stub_folders_excluded_from_direct_folder_entries(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "!README.md").write_text("# Docs\n", encoding="utf-8")

    stubs = root / "stubs"
    stubs.mkdir()
    (stubs / "!README.md").write_text("# Stubs\n", encoding="utf-8")

    result = reconcile_tree(root)

    root_update = next(update for update in result.updates if update.path == root / "!README.md")
    assert "stubs/!README.md" not in root_update.new_text
    assert "## Direct Folders" in root_update.new_text


def test_reconcile_tree_removes_stale_managed_entries(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    readme = root / "!README.md"
    readme.write_text(
        """# Docs

## Direct Files
<!-- docs-index:files:start -->
- [gone.md](gone.md) - Gone direct doc.
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
- [stale.md](stubs/stale.md) - Stub: Gone stub doc.
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
- [Gone](gone/!README.md) - Gone folder docs.
<!-- docs-index:folders:end -->

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
    assert "gone/!README.md" not in rewritten
    assert "Keep this note." in rewritten


def test_reconcile_tree_repairs_missing_end_marker(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    (root / "alpha.md").write_text("Parent index: [Docs](./!README.md)\n\nAlpha body", encoding="utf-8")
    readme = root / "!README.md"
    readme.write_text(
        """# Docs

## Direct Files
<!-- docs-index:files:start -->
- [alpha.md](alpha.md) - Custom alpha description.

## Notes

Keep this note.
""",
        encoding="utf-8",
    )

    result = reconcile_tree(root)

    assert len(result.updates) == 1
    assert apply_updates(result) == 1
    rewritten = readme.read_text(encoding="utf-8")
    assert "<!-- docs-index:files:end -->" in rewritten
    assert "Custom alpha description." in rewritten
    assert "Keep this note." in rewritten


def test_apply_updates_creates_missing_readme(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()

    result = reconcile_tree(root)

    assert apply_updates(result) == 1
    assert (root / "!README.md").exists()


def test_apply_updates_skips_unchanged_existing_files(tmp_path: Path) -> None:
    root = tmp_path / "docs"
    root.mkdir()
    readme = root / "!README.md"
    readme.write_text("# Docs\n", encoding="utf-8")

    result = ReconcileResult(
        updates=[FileUpdate(path=readme, old_text="# Docs\n", new_text="# Docs\n")],
        messages=[],
    )

    assert apply_updates(result) == 0
    assert readme.read_text(encoding="utf-8") == "# Docs\n"
