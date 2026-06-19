from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index.readme_io import ensure_managed_sections
from docs_index.readme_io import description_from_file
from docs_index.readme_io import description_from_folder
from docs_index.readme_io import first_heading_title
from docs_index.readme_io import make_readme_template
from docs_index.readme_io import folder_title
from docs_index.readme_io import parse_managed_entries
from docs_index.readme_io import render_file_entry
from docs_index.readme_io import render_folder_entry
from docs_index.readme_io import replace_managed_block
from docs_index.readme_io import title_from_folder


def test_ensure_managed_sections_inserts_before_related_docs() -> None:
    text = """# Title

Intro

## Related Docs

Stuff
"""

    result = ensure_managed_sections(text)

    assert "## Direct Files" in result
    assert "## Stub Files" in result
    assert "## Direct Folders" in result
    assert result.index("## Direct Files") < result.index("## Related Docs")
    assert "Intro" in result
    assert "Stuff" in result


def test_ensure_managed_sections_leaves_marked_readme_unchanged() -> None:
    text = """# Title

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->

## Notes

Keep me.
"""

    assert ensure_managed_sections(text) == text


def test_ensure_managed_sections_wraps_legacy_direct_files_section() -> None:
    text = """# Title

## Direct Files
- [alpha.md](alpha.md) - Alpha documentation.

## Related Docs

Keep me.
"""

    result = ensure_managed_sections(text)

    assert "## Direct Files" in result
    assert "<!-- docs-index:files:start -->" in result
    assert "<!-- docs-index:files:end -->" in result
    assert "- [alpha.md](alpha.md) - Alpha documentation." in result
    assert "Keep me." in result


def test_replace_managed_block_preserves_unrelated_sections() -> None:
    text = """# Title

Intro

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->

## Notes

Keep me.
"""

    result = replace_managed_block(text, "files", ["- alpha.md", "- beta.md"])

    assert "## Notes" in result
    assert "Keep me." in result
    assert "- alpha.md" in result
    assert "- beta.md" in result
    assert result.index("## Notes") > result.index("## Direct Folders")


def test_replace_managed_block_uses_blank_line_before_content() -> None:
    text = """# Title

## Direct Files
<!-- docs-index:files:start -->
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
<!-- docs-index:folders:end -->
"""

    result = replace_managed_block(text, "stubs", ["- one.md"])

    assert "<!-- docs-index:stubs:start -->\n\n- one.md" in result


def test_parse_managed_entries_parses_files_stubs_and_folders() -> None:
    readme_path = Path("/tmp/docs/!README.md")
    text = """# Title

## Direct Files
<!-- docs-index:files:start -->
- [foo.md](foo.md) - Foo documentation.
bad line
<!-- docs-index:files:end -->

## Stub Files
<!-- docs-index:stubs:start -->
- [Foo](foo/!README.md) - Foo documentation.
<!-- docs-index:stubs:end -->

## Direct Folders
<!-- docs-index:folders:start -->
- [Bar](bar/!README.md) - Bar docs.
<!-- docs-index:folders:end -->
"""

    entries = parse_managed_entries(readme_path, text)

    assert [entry.section for entry in entries] == ["files", "stubs", "folders"]
    assert [entry.link_text for entry in entries] == ["foo.md", "Foo", "Bar"]
    assert [entry.link_target for entry in entries] == ["foo.md", "foo/!README.md", "bar/!README.md"]
    assert [entry.description for entry in entries] == [
        "Foo documentation.",
        "Foo documentation.",
        "Bar docs.",
    ]
    assert entries[0].original_line == "- [foo.md](foo.md) - Foo documentation."
    assert entries[0].readme_path == readme_path


def test_parse_managed_entries_ignores_malformed_lines() -> None:
    readme_path = Path("/tmp/docs/!README.md")
    text = """# Title

## Direct Files
<!-- docs-index:files:start -->
- missing link syntax
- [ok.md](ok.md)
- [also ok](also-ok.md) - Good.
<!-- docs-index:files:end -->
"""

    entries = parse_managed_entries(readme_path, text)

    assert [entry.link_text for entry in entries] == ["also ok"]
    assert [entry.description for entry in entries] == ["Good."]


def test_description_from_file_uses_title_case_fallbacks() -> None:
    assert description_from_file(Path("foo-bar_baz.md"), is_stub=False) == "Foo Bar Baz documentation."
    assert description_from_file(Path("foo-bar_baz.md"), is_stub=True) == "Stub: Foo Bar Baz documentation."


def test_description_from_folder_uses_title_case_fallback() -> None:
    assert description_from_folder(Path("foo-bar_baz")) == "Foo Bar Baz documentation."


def test_render_helpers_use_repo_style_bullets() -> None:
    assert render_file_entry("foo.md", "foo.md", "Foo documentation.") == "- [foo.md](foo.md) - Foo documentation."
    assert render_folder_entry("Foo", "foo/!README.md", "Foo documentation.") == "- [Foo](foo/!README.md) - Foo documentation."


def test_title_from_folder_uses_title_case_split() -> None:
    assert title_from_folder(Path("foo-bar_baz")) == "Foo Bar Baz"


def test_make_readme_template_omits_parent_line_for_root() -> None:
    root = Path("/tmp/docs")

    result = make_readme_template(root, root, parent_title=None)

    assert result.startswith("# Docs")
    assert "Parent index:" not in result
    assert "## Ownership" in result
    assert "## Does Not Belong" in result
    assert "## Direct Files" in result
    assert "## Stub Files" in result
    assert "## Direct Folders" in result
    assert "## Related Docs" in result
    assert "## Notes" in result


def test_make_readme_template_includes_parent_line_for_child() -> None:
    root = Path("/tmp/docs")
    folder = root / "guide"

    result = make_readme_template(folder, root, parent_title="Docs")

    assert "Parent index: [Docs](../!README.md)" in result
    assert result.startswith("# Guide")
    assert "This index keeps the guide documentation organized and easy to scan." in result


def test_first_heading_title_handles_single_hash_heading() -> None:
    assert first_heading_title("# Hello World\n\nBody") == "Hello World"


def test_first_heading_title_handles_double_hash_heading() -> None:
    assert first_heading_title("Intro\n\n## Hello World\n\nBody") == "Hello World"


def test_first_heading_title_returns_none_without_heading() -> None:
    assert first_heading_title("Intro\n\nBody") is None


def test_folder_title_falls_back_to_hyphenated_folder_name() -> None:
    folder = Path("foo-bar_baz")

    assert folder_title(folder) == "Foo Bar Baz"
