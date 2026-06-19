from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import DescriptionConfig
from doc_ledger.config import DraftConfig
from doc_ledger.config import MarkerConfig
from doc_ledger.config import ReadmeTemplateConfig
from doc_ledger.config import SectionConfig
from doc_ledger.readme_io import ensure_managed_sections
from doc_ledger.readme_io import description_from_file
from doc_ledger.readme_io import description_from_folder
from doc_ledger.readme_io import make_readme_template
from doc_ledger.readme_io import managed_root_title
from doc_ledger.readme_io import parse_managed_entries
from doc_ledger.readme_io import render_marker_end
from doc_ledger.readme_io import render_marker_start
from doc_ledger.readme_io import render_description_template
from doc_ledger.readme_io import replace_managed_block


def test_ensure_managed_sections_inserts_doc_ledger_markers() -> None:
    text = "# Title\n\n## Related Docs\n"

    result = ensure_managed_sections(text)

    assert "<!-- doc-ledger:files:start -->" in result
    assert "<!-- doc-ledger:stubs:start -->" in result
    assert "<!-- doc-ledger:folders:start -->" in result


def test_marker_helpers_render_configured_prefix() -> None:
    assert render_marker_start("nav-ledger", "files") == "<!-- nav-ledger:files:start -->"
    assert render_marker_end("nav-ledger", "files") == "<!-- nav-ledger:files:end -->"


def test_ensure_managed_sections_uses_configured_marker_prefix() -> None:
    text = "# Title\n\n## Related Docs\n"

    result = ensure_managed_sections(text, DocLedgerConfig(markers=MarkerConfig(prefix="nav-ledger")))

    assert "<!-- nav-ledger:files:start -->" in result
    assert "<!-- nav-ledger:stubs:start -->" in result
    assert "<!-- nav-ledger:folders:start -->" in result


def test_replace_managed_block_uses_doc_ledger_markers() -> None:
    text = """# Title

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
"""

    result = replace_managed_block(text, "files", ["- [guide.md](guide.md) - Guide documentation."])

    assert "<!-- doc-ledger:files:start -->" in result
    assert "- [guide.md](guide.md) - Guide documentation." in result


def test_parse_managed_entries_uses_configured_marker_prefix() -> None:
    readme_path = Path("/tmp/docs/README.md")
    text = """# Title

## Direct Files
<!-- nav-ledger:files:start -->
- [guide.md](guide.md) - Guide documentation.
<!-- nav-ledger:files:end -->
"""

    entries = parse_managed_entries(readme_path, text, DocLedgerConfig(markers=MarkerConfig(prefix="nav-ledger")))

    assert len(entries) == 1
    assert entries[0].section == "files"
    assert entries[0].link_text == "guide.md"
    assert entries[0].link_target == "guide.md"


def test_make_readme_template_uses_doc_ledger_markers() -> None:
    folder = Path("/tmp/docs")

    result = make_readme_template(folder, folder, None)

    assert "<!-- doc-ledger:files:start -->" in result
    assert "<!-- doc-ledger:stubs:start -->" in result
    assert "<!-- doc-ledger:folders:start -->" in result


def test_make_readme_template_uses_configured_index_file() -> None:
    folder = Path("/tmp/docs/guide")
    root = Path("/tmp/docs")

    result = make_readme_template(folder, root, "Docs", index_file="README.md")

    assert "Parent index: [Docs](../README.md)" in result


def test_make_readme_template_defaults_include_optional_sections() -> None:
    folder = Path("/tmp/docs")

    result = make_readme_template(folder, folder, None)

    assert "## Ownership" in result
    assert "## Does Not Belong" in result
    assert "## Related Docs" in result
    assert "## Notes" in result


def test_make_readme_template_omits_disabled_optional_sections() -> None:
    folder = Path("/tmp/docs")
    config = DocLedgerConfig(
        readme_template=ReadmeTemplateConfig(
            include_ownership=False,
            include_does_not_belong=False,
            include_related_docs=False,
            include_notes=False,
        )
    )

    result = make_readme_template(folder, folder, None, config=config)

    assert "## Ownership" not in result
    assert "## Does Not Belong" not in result
    assert "## Related Docs" not in result
    assert "## Notes" not in result


def test_make_readme_template_keeps_managed_sections_when_optional_sections_are_disabled() -> None:
    folder = Path("/tmp/docs")
    config = DocLedgerConfig(
        readme_template=ReadmeTemplateConfig(
            include_ownership=False,
            include_does_not_belong=False,
            include_related_docs=False,
            include_notes=False,
        )
    )

    result = make_readme_template(folder, folder, None, config=config)

    assert "## Direct Files" in result
    assert "<!-- doc-ledger:files:start -->" in result
    assert "## Stub Files" in result
    assert "<!-- doc-ledger:stubs:start -->" in result
    assert "## Direct Folders" in result
    assert "<!-- doc-ledger:folders:start -->" in result


def test_description_from_file_uses_filename_for_non_markdown_extensions() -> None:
    assert description_from_file(Path("architecture.png"), False) == "Architecture documentation."
    assert description_from_file(Path("draft-report.pdf"), True) == "Stub: Draft Report documentation."


def test_render_description_template_replaces_title() -> None:
    assert render_description_template("File: {title}.", "Foo Bar") == "File: Foo Bar."


def test_description_from_file_uses_default_file_description() -> None:
    assert description_from_file(Path("foo-bar.md"), False) == "Foo Bar documentation."


def test_description_from_file_uses_default_stub_description() -> None:
    assert description_from_file(Path("foo-bar.md"), True) == "Stub: Foo Bar documentation."


def test_description_from_file_uses_configured_file_template() -> None:
    config = DocLedgerConfig(description=DescriptionConfig(file_template="File: {title}."))

    assert description_from_file(Path("foo-bar.md"), False, config) == "File: Foo Bar."


def test_description_from_file_uses_configured_draft_prefix() -> None:
    config = DocLedgerConfig(draft=DraftConfig(description_prefix="Draft: "))

    assert description_from_file(Path("foo-bar.md"), True, config) == "Draft: Foo Bar documentation."


def test_description_from_folder_uses_default_folder_description() -> None:
    assert description_from_folder(Path("foo-bar")) == "Foo Bar documentation."


def test_description_from_folder_uses_configured_folder_template() -> None:
    config = DocLedgerConfig(description=DescriptionConfig(folder_template="Folder: {title}."))

    assert description_from_folder(Path("foo-bar"), config) == "Folder: Foo Bar."


def test_description_from_folder_title_cases_hyphenated_folder_names() -> None:
    assert description_from_folder(Path("service-runbooks")) == "Service Runbooks documentation."


def test_ensure_managed_sections_migrates_top_level_headings() -> None:
    text = """# Docs

## Top-Level Files
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Alpha description.
<!-- doc-ledger:files:end -->

## Rulebook
Keep this rulebook.

## Top-Level Folders
<!-- doc-ledger:folders:start -->
- [Guide](guide/README.md) - Guide description.
<!-- doc-ledger:folders:end -->

## Related Docs
Still here.

## Notes
More notes.
"""

    result = ensure_managed_sections(text)

    assert "## Top-Level Files" not in result
    assert "## Top-Level Folders" not in result
    assert "## Direct Files" in result
    assert result.count("## Direct Files") == 1
    assert result.count("## Stub Files") == 1
    assert "## Direct Folders" in result
    assert result.count("## Direct Folders") == 1
    assert "- [alpha.md](alpha.md) - Alpha description." in result
    assert "- [Guide](guide/README.md) - Guide description." in result
    assert "## Rulebook" in result
    assert "## Related Docs" in result
    assert "## Notes" in result


def test_ensure_managed_sections_uses_configured_headings_and_aliases() -> None:
    config = DocLedgerConfig(
        sections=SectionConfig(
            files_heading="Files",
            stubs_heading="Draft Files",
            folders_heading="Directories",
            legacy_files_headings=["Pages"],
            legacy_folders_headings=["Subdirectories"],
        )
    )
    text = """# Docs

## Pages
<!-- doc-ledger:files:start -->
- [alpha.md](alpha.md) - Alpha description.
<!-- doc-ledger:files:end -->

## Subdirectories
<!-- doc-ledger:folders:start -->
- [guide](guide/README.md) - Guide description.
<!-- doc-ledger:folders:end -->
"""

    result = ensure_managed_sections(text, config)

    assert "## Files" in result
    assert "## Draft Files" in result
    assert "## Directories" in result
    assert result.count("## Files") == 1
    assert result.count("## Draft Files") == 1
    assert result.count("## Directories") == 1
    assert "## Pages" not in result
    assert "## Subdirectories" not in result
    assert "- [alpha.md](alpha.md) - Alpha description." in result
    assert "- [guide](guide/README.md) - Guide description." in result


def test_managed_root_title_prefers_consistent_child_titles() -> None:
    root = Path("/tmp/docs")

    result = managed_root_title(root, "# Documentation\n", ["Docs", "Docs"])

    assert result == "Docs"


def test_managed_root_title_falls_back_to_root_heading_when_no_child_titles() -> None:
    root = Path("/tmp/docs")

    result = managed_root_title(root, "# Documentation\n", [])

    assert result == "Documentation"
