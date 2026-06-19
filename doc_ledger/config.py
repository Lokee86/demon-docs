from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path

from tomlkit import parse as parse_toml

DEFAULT_INDEX_FILE = "README.md"


@dataclass
class MarkerConfig:
    prefix: str = "doc-ledger"


@dataclass
class ParentLinkConfig:
    label: str = "Parent index"
    folder_indexes: bool = True
    indexed_files: bool = False

    def __init__(
        self,
        label: str = "Parent index",
        folder_indexes: bool = True,
        indexed_files: bool = False,
        enabled: bool | None = None,
    ) -> None:
        self.label = label
        self.folder_indexes = folder_indexes
        self.indexed_files = indexed_files
        if enabled is not None:
            self.enabled = enabled

    @property
    def enabled(self) -> bool:
        return self.folder_indexes or self.indexed_files

    @enabled.setter
    def enabled(self, value: bool) -> None:
        enabled = bool(value)
        self.folder_indexes = enabled
        self.indexed_files = enabled


@dataclass
class SectionConfig:
    files_heading: str = "Direct Files"
    stubs_heading: str = "Stub Files"
    folders_heading: str = "Direct Folders"
    legacy_files_headings: list[str] = field(default_factory=lambda: ["Top-Level Files"])
    legacy_folders_headings: list[str] = field(default_factory=lambda: ["Top-Level Folders"])


@dataclass
class DraftConfig:
    folder: str = "stubs"
    description_prefix: str = "Stub: "


@dataclass
class FileConfig:
    index_file: str = DEFAULT_INDEX_FILE
    include_patterns: list[str] = field(default_factory=lambda: ["**/*.md"])
    exclude_patterns: list[str] = field(default_factory=list)
    editable_parent_index_extensions: list[str] = field(default_factory=lambda: [".md"])


@dataclass
class DescriptionConfig:
    file_template: str = "{title} documentation."
    folder_template: str = "{title} documentation."


@dataclass
class WatchConfig:
    debounce_seconds: float = 0.75
    ignored_dirs: list[str] = field(default_factory=lambda: [".git", ".cache", "__pycache__"])
    ignored_suffixes: list[str] = field(default_factory=lambda: ["~", ".swp", ".tmp", ".bak"])


@dataclass
class ReadmeTemplateConfig:
    managed_sections: list[str] = field(default_factory=lambda: ["files", "stubs", "folders"])
    include_ownership: bool = True
    include_does_not_belong: bool = True
    include_related_docs: bool = True
    include_notes: bool = True


@dataclass
class DocLedgerConfig:
    root: str = "docs"
    index_file: str = DEFAULT_INDEX_FILE
    markers: MarkerConfig = field(default_factory=MarkerConfig)
    parent_link: ParentLinkConfig = field(default_factory=ParentLinkConfig)
    sections: SectionConfig = field(default_factory=SectionConfig)
    draft: DraftConfig = field(default_factory=DraftConfig)
    file: FileConfig = field(default_factory=FileConfig)
    description: DescriptionConfig = field(default_factory=DescriptionConfig)
    watch: WatchConfig = field(default_factory=WatchConfig)
    readme_template: ReadmeTemplateConfig = field(default_factory=ReadmeTemplateConfig)


def default_config() -> DocLedgerConfig:
    return DocLedgerConfig()


def starter_config_text() -> str:
    return (
        'root = "docs"\n'
        'index_file = "README.md"\n'
        "\n"
        "[parent_link]\n"
        "folder_indexes = true\n"
        "indexed_files = false\n"
        "\n"
        "[drafts]\n"
        'folder = "stubs"\n'
        'description_prefix = "Stub: "\n'
        "\n"
        "[watch]\n"
        "debounce_seconds = 0.75\n"
        'ignored_dirs = [".git", ".cache", "__pycache__"]\n'
        'ignored_suffixes = ["~", ".swp", ".tmp", ".bak"]\n'
    )


def discover_config(start: Path) -> Path | None:
    current = start.resolve(strict=False)
    if current.is_file():
        current = current.parent

    while True:
        dot_config = current / ".doc-ledger.toml"
        if dot_config.exists():
            return dot_config

        plain_config = current / "doc-ledger.toml"
        if plain_config.exists():
            return plain_config

        if current == current.parent:
            return None
        current = current.parent


def local_config_path(cwd: Path) -> Path | None:
    dot_config = cwd / ".doc-ledger.toml"
    if dot_config.exists():
        return dot_config

    plain_config = cwd / "doc-ledger.toml"
    if plain_config.exists():
        return plain_config

    return None


def global_config_path(env: dict[str, str] | None = None, home: Path | None = None) -> Path:
    env = env or {}
    if "XDG_CONFIG_HOME" in env:
        return Path(env["XDG_CONFIG_HOME"]) / "doc-ledger" / "config.toml"

    home = home or Path.home()
    return home / ".config" / "doc-ledger" / "config.toml"


def selected_config_path(
    cwd: Path,
    explicit_config: Path | None,
    no_local: bool,
    no_global: bool,
    env: dict[str, str] | None = None,
    home: Path | None = None,
) -> Path | None:
    if explicit_config is not None:
        return explicit_config

    if not no_local:
        local_config = local_config_path(cwd)
        if local_config is not None:
            return local_config

    if not no_global:
        global_config = global_config_path(env=env, home=home)
        if global_config.exists():
            return global_config

    return None


def load_config(config_path: Path | None = None) -> DocLedgerConfig:
    if config_path is None:
        return default_config()

    if not config_path.exists():
        raise FileNotFoundError(config_path)

    config = default_config()
    data = parse_toml(config_path.read_text(encoding="utf-8"))

    if "root" in data:
        config.root = str(data["root"])
    if "index_file" in data:
        config.index_file = str(data["index_file"])

    markers = data.get("markers")
    if markers is not None:
        _apply_marker_config(config.markers, markers)

    parent_link = data.get("parent_link")
    if parent_link is not None:
        _apply_parent_link_config(config.parent_link, parent_link)

    sections = data.get("sections")
    if sections is not None:
        _apply_section_config(config.sections, sections)

    drafts = data.get("drafts")
    if drafts is not None:
        _apply_draft_config(config.draft, drafts)

    files = data.get("files")
    if files is not None:
        _apply_file_config(config.file, files)

    editable = data.get("editable")
    if editable is not None:
        _apply_editable_config(config.file, editable)

    descriptions = data.get("descriptions")
    if descriptions is not None:
        _apply_description_config(config.description, descriptions)

    watch = data.get("watch")
    if watch is not None:
        _apply_watch_config(config.watch, watch)

    aliases = data.get("aliases")
    if aliases is not None:
        _apply_alias_config(config.sections, aliases)

    template = data.get("template")
    if template is not None:
        _apply_template_config(config.readme_template, template)

    return config


def is_parent_link_editable(path: Path, config: DocLedgerConfig) -> bool:
    return path.suffix in config.file.editable_parent_index_extensions


def _apply_marker_config(config: MarkerConfig, data: dict) -> None:
    if "prefix" in data:
        config.prefix = str(data["prefix"])


def _apply_parent_link_config(config: ParentLinkConfig, data: dict) -> None:
    if "label" in data:
        config.label = str(data["label"])

    has_folder_indexes = "folder_indexes" in data
    has_indexed_files = "indexed_files" in data

    if "enabled" in data:
        enabled = bool(data["enabled"])
        if not has_folder_indexes:
            config.folder_indexes = enabled
        if not has_indexed_files:
            config.indexed_files = enabled

    if has_folder_indexes:
        config.folder_indexes = bool(data["folder_indexes"])
    if has_indexed_files:
        config.indexed_files = bool(data["indexed_files"])


def _apply_section_config(config: SectionConfig, data: dict) -> None:
    files = data.get("files")
    if files is not None:
        _apply_section_heading(config, "files_heading", files)

    stubs = data.get("stubs")
    if stubs is not None:
        _apply_section_heading(config, "stubs_heading", stubs)

    folders = data.get("folders")
    if folders is not None:
        _apply_section_heading(config, "folders_heading", folders)


def _apply_section_heading(config: SectionConfig, attribute: str, data: dict) -> None:
    for key in ("heading", "title", "name"):
        if key in data:
            setattr(config, attribute, str(data[key]))
            return


def _apply_draft_config(config: DraftConfig, data: dict) -> None:
    if "folder" in data:
        config.folder = str(data["folder"])
    if "description_prefix" in data:
        config.description_prefix = str(data["description_prefix"])


def _apply_file_config(config: FileConfig, data: dict) -> None:
    if "include_patterns" in data:
        config.include_patterns = [str(pattern) for pattern in data["include_patterns"]]
    if "exclude_patterns" in data:
        config.exclude_patterns = [str(pattern) for pattern in data["exclude_patterns"]]


def _apply_editable_config(config: FileConfig, data: dict) -> None:
    if "parent_index_extensions" in data:
        config.editable_parent_index_extensions = [str(extension) for extension in data["parent_index_extensions"]]
    elif "extensions" in data:
        config.editable_parent_index_extensions = [str(extension) for extension in data["extensions"]]


def _apply_description_config(config: DescriptionConfig, data: dict) -> None:
    if "file_template" in data:
        config.file_template = str(data["file_template"])
    if "folder_template" in data:
        config.folder_template = str(data["folder_template"])


def _apply_watch_config(config: WatchConfig, data: dict) -> None:
    if "debounce_seconds" in data:
        config.debounce_seconds = float(data["debounce_seconds"])
    if "ignored_dirs" in data:
        config.ignored_dirs = [str(entry) for entry in data["ignored_dirs"]]
    if "ignored_suffixes" in data:
        config.ignored_suffixes = [str(entry) for entry in data["ignored_suffixes"]]


def _apply_alias_config(config: SectionConfig, data: dict) -> None:
    if "files" in data:
        config.legacy_files_headings = [str(entry) for entry in data["files"]]
    if "folders" in data:
        config.legacy_folders_headings = [str(entry) for entry in data["folders"]]


def _apply_template_config(config: ReadmeTemplateConfig, data: dict) -> None:
    if "managed_sections" in data:
        config.managed_sections = [str(entry) for entry in data["managed_sections"]]
    if "include_ownership" in data:
        config.include_ownership = bool(data["include_ownership"])
    if "include_does_not_belong" in data:
        config.include_does_not_belong = bool(data["include_does_not_belong"])
    if "include_related_docs" in data:
        config.include_related_docs = bool(data["include_related_docs"])
    if "include_notes" in data:
        config.include_notes = bool(data["include_notes"])
