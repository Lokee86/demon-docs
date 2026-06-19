from __future__ import annotations

import sys
from pathlib import Path
import pytest


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.config import default_config
from doc_ledger.config import discover_config
from doc_ledger.config import is_parent_link_editable
from doc_ledger.config import load_config


def test_default_config_preserves_current_behavior() -> None:
    config = default_config()

    assert config.root == "docs"
    assert config.index_file == "!README.md"
    assert config.markers.prefix == "doc-ledger"
    assert config.parent_link.label == "Parent index"
    assert config.parent_link.enabled is True
    assert config.sections.files_heading == "Direct Files"
    assert config.sections.stubs_heading == "Stub Files"
    assert config.sections.folders_heading == "Direct Folders"
    assert config.sections.legacy_files_headings == ["Top-Level Files"]
    assert config.sections.legacy_folders_headings == ["Top-Level Folders"]
    assert config.draft.folder == "stubs"
    assert config.draft.description_prefix == "Stub: "
    assert config.file.include_patterns == ["**/*.md"]
    assert config.file.exclude_patterns == []
    assert config.file.editable_parent_index_extensions == [".md"]
    assert config.description.file_template == "{title} documentation."
    assert config.description.folder_template == "{title} documentation."
    assert config.watch.debounce_seconds == 0.75
    assert config.watch.ignored_dirs == [".git", ".cache", "__pycache__"]
    assert config.watch.ignored_suffixes == ["~", ".swp", ".tmp", ".bak"]
    assert config.readme_template.managed_sections == ["files", "stubs", "folders"]
    assert config.readme_template.include_ownership is True
    assert config.readme_template.include_does_not_belong is True
    assert config.readme_template.include_related_docs is True
    assert config.readme_template.include_notes is True


def test_load_config_missing_path_raises(tmp_path: Path) -> None:
    missing = tmp_path / "doc-ledger.toml"

    with pytest.raises(FileNotFoundError):
        load_config(missing)


def test_discover_config_returns_none_when_no_config_exists(tmp_path: Path) -> None:
    nested = tmp_path / "docs" / "guide"
    nested.mkdir(parents=True)

    assert discover_config(nested) is None


def test_discover_config_finds_dot_config_in_current_directory(tmp_path: Path) -> None:
    config_path = tmp_path / ".doc-ledger.toml"
    config_path.write_text("", encoding="utf-8")

    assert discover_config(tmp_path) == config_path


def test_discover_config_finds_plain_config_in_parent_directory(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("", encoding="utf-8")
    nested = tmp_path / "docs" / "guide"
    nested.mkdir(parents=True)

    assert discover_config(nested) == config_path


def test_discover_config_prefers_nearer_config_over_farther_config(tmp_path: Path) -> None:
    farther = tmp_path / ".doc-ledger.toml"
    farther.write_text("", encoding="utf-8")
    nested = tmp_path / "docs"
    nested.mkdir()
    nearer = nested / "doc-ledger.toml"
    nearer.write_text("", encoding="utf-8")

    assert discover_config(nested) == nearer


def test_discover_config_prefers_dot_config_over_plain_config_in_same_directory(tmp_path: Path) -> None:
    plain = tmp_path / "doc-ledger.toml"
    plain.write_text("", encoding="utf-8")
    dot = tmp_path / ".doc-ledger.toml"
    dot.write_text("", encoding="utf-8")

    assert discover_config(tmp_path) == dot


def test_load_config_empty_file_keeps_defaults(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("", encoding="utf-8")

    config = load_config(config_path)

    assert config == default_config()


def test_load_config_overrides_marker_prefix(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("[markers]\nprefix = \"custom-ledger\"\n", encoding="utf-8")

    config = load_config(config_path)

    assert config.markers.prefix == "custom-ledger"


def test_load_config_overrides_index_file(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text('index_file = "INDEX.md"\n', encoding="utf-8")

    config = load_config(config_path)

    assert config.index_file == "INDEX.md"


def test_load_config_overrides_include_patterns(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("[files]\ninclude_patterns = [\"docs/**/*.md\", \"notes/**/*.md\"]\n", encoding="utf-8")

    config = load_config(config_path)

    assert config.file.include_patterns == ["docs/**/*.md", "notes/**/*.md"]


def test_load_config_overrides_exclude_patterns(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("[files]\nexclude_patterns = [\"**/*.tmp\"]\n", encoding="utf-8")

    config = load_config(config_path)

    assert config.file.exclude_patterns == ["**/*.tmp"]


def test_load_config_overrides_editable_extensions(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("[editable]\nparent_index_extensions = [\".md\", \".markdown\"]\n", encoding="utf-8")

    config = load_config(config_path)

    assert config.file.editable_parent_index_extensions == [".md", ".markdown"]


def test_load_config_overrides_section_headings_and_aliases(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
[sections.files]
heading = "Files"

[sections.stubs]
heading = "Draft Files"

[sections.folders]
heading = "Directories"

[aliases]
files = ["Pages"]
folders = ["Subdirectories"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    config = load_config(config_path)

    assert config.sections.files_heading == "Files"
    assert config.sections.stubs_heading == "Draft Files"
    assert config.sections.folders_heading == "Directories"
    assert config.sections.legacy_files_headings == ["Pages"]
    assert config.sections.legacy_folders_headings == ["Subdirectories"]


def test_load_config_overrides_readme_template_toggles(tmp_path: Path) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
[template]
include_ownership = false
include_does_not_belong = true
include_related_docs = false
include_notes = false
""".strip()
        + "\n",
        encoding="utf-8",
    )

    config = load_config(config_path)

    assert config.readme_template.include_ownership is False
    assert config.readme_template.include_does_not_belong is True
    assert config.readme_template.include_related_docs is False
    assert config.readme_template.include_notes is False


def test_is_parent_link_editable_defaults_to_md_only() -> None:
    config = default_config()

    assert is_parent_link_editable(Path("guide.md"), config) is True
    assert is_parent_link_editable(Path("image.png"), config) is False


def test_is_parent_link_editable_respects_configured_extensions() -> None:
    config = default_config()
    config.file.editable_parent_index_extensions = [".md", ".mdx"]

    assert is_parent_link_editable(Path("guide.mdx"), config) is True


def test_is_parent_link_editable_matches_exact_suffix_with_dot() -> None:
    config = default_config()
    config.file.editable_parent_index_extensions = ["md"]

    assert is_parent_link_editable(Path("guide.md"), config) is False
