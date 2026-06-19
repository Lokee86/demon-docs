from __future__ import annotations

from pathlib import Path
import re

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import default_config
from doc_ledger.model import IndexEntry

MANAGED_SECTION_NAMES = ("files", "stubs", "folders")

ENTRY_PATTERN = re.compile(r"^\s*-\s+\[([^\]]+)\]\(([^)]+)\)\s+-\s+(.*)$")
HEADING_PATTERN = re.compile(r"^#{1,6}\s+")


def ensure_managed_sections(text: str, config: DocLedgerConfig | None = None) -> str:
    config = config or default_config()
    marker_start = _marker_start_map(config)
    marker_end = _marker_end_map(config)
    section_titles = _section_titles_map(config)
    legacy_section_titles = _legacy_section_titles_map(config)
    legacy_present = _has_legacy_managed_sections(text, section_titles, legacy_section_titles)

    if legacy_present and not _has_any_managed_markers(text, marker_start, marker_end):
        return _wrap_legacy_managed_sections(text, marker_start, marker_end, section_titles, legacy_section_titles)

    if legacy_present:
        text = _normalize_legacy_heading_lines(text, section_titles, legacy_section_titles)

    missing_sections = [section for section in MANAGED_SECTION_NAMES if not _has_managed_section(text, section, marker_start, marker_end, section_titles)]
    if not missing_sections:
        return text

    anchor = _find_anchor(text)
    missing_block = _render_missing_sections(missing_sections, marker_start, marker_end, section_titles)

    if anchor is None:
        if text:
            return f"{text}\n\n{missing_block}"
        return missing_block

    before, after = text[:anchor], text[anchor:]
    if before and not before.endswith("\n\n"):
        before = before.rstrip("\n") + "\n\n"
    return f"{before}{missing_block}\n\n{after.lstrip()}"


def replace_managed_block(text: str, section: str, lines: list[str], config: DocLedgerConfig | None = None) -> str:
    config = config or default_config()
    if section not in MANAGED_SECTION_NAMES:
        raise ValueError(f"unknown managed section: {section}")

    marker_start = _marker_start_map(config)
    marker_end = _marker_end_map(config)
    text = ensure_managed_sections(text, config)
    start_marker = marker_start[section]
    end_marker = marker_end[section]
    start = text.index(start_marker)
    end = text.find(end_marker, start)
    if end == -1:
        span_end = _find_managed_section_end(text, start, section, marker_start)
    else:
        span_end = end + len(end_marker)

    block_lines = [start_marker]
    if lines:
        block_lines.append("")
        block_lines.extend(lines)
    block_lines.append(end_marker)

    replacement = "\n".join(block_lines)
    return f"{text[:start]}{replacement}{text[span_end:]}"


def render_file_entry(filename: str, target: str, description: str) -> str:
    return f"- [{filename}]({target}) - {description}"


def render_folder_entry(link_text: str, target: str, description: str) -> str:
    return f"- [{link_text}]({target}) - {description}"


def render_description_template(template: str, title: str) -> str:
    return template.replace("{title}", title)


def description_from_file(path: Path, is_stub: bool, config: DocLedgerConfig | None = None) -> str:
    config = config or default_config()
    stem = path.stem.replace("-", " ").replace("_", " ")
    title = " ".join(part.capitalize() for part in stem.split())
    description = render_description_template(config.description.file_template, title)
    if is_stub:
        return f"{config.draft.description_prefix}{description}"
    return description


def description_from_folder(path: Path, config: DocLedgerConfig | None = None) -> str:
    config = config or default_config()
    title = title_from_folder(path)
    return render_description_template(config.description.folder_template, title)


def title_from_folder(path: Path) -> str:
    return " ".join(part.capitalize() for part in path.name.replace("-", " ").replace("_", " ").split())


def first_heading_title(text: str) -> str | None:
    for line in text.splitlines():
        match = re.match(r"^#{1,6}\s+(.*\S)\s*$", line)
        if match is not None:
            return match.group(1)
    return None


def folder_title(folder: Path, readme_text: str | None = None) -> str:
    if readme_text is not None:
        heading = first_heading_title(readme_text)
        if heading is not None:
            return heading
    return title_from_folder(folder)


def managed_root_title(folder: Path, readme_text: str | None = None, child_parent_titles: list[str] | None = None) -> str:
    unique_child_titles: list[str] = []
    for title in child_parent_titles or []:
        if title and title not in unique_child_titles:
            unique_child_titles.append(title)

    if len(unique_child_titles) == 1:
        return unique_child_titles[0]

    if readme_text is not None:
        heading = first_heading_title(readme_text)
        if heading is not None:
            return heading

    return title_from_folder(folder)


def make_readme_template(
    folder: Path,
    root: Path,
    parent_title: str | None,
    index_file: str | None = None,
    config: DocLedgerConfig | None = None,
) -> str:
    config = config or default_config()
    index_file = index_file or config.index_file
    marker_start = _marker_start_map(config)
    marker_end = _marker_end_map(config)
    section_titles = _section_titles_map(config)
    folder_title = title_from_folder(folder)
    lines = [
        f"# {folder_title}",
        "",
        f"This index summarizes the {folder_title.lower()} docs.",
    ]
    if folder != root and parent_title is not None:
        if config.parent_link.enabled:
            lines.extend(
                [
                    "",
                    f"{config.parent_link.label}: [{parent_title}](../{index_file})",
                ]
            )

    if config.readme_template.include_ownership:
        lines.extend(
            [
                "",
                "## Ownership",
                "",
                "Describe who maintains these docs.",
            ]
        )

    if config.readme_template.include_does_not_belong:
        lines.extend(
            [
                "",
                "## Does Not Belong",
                "",
                "List content that belongs somewhere else.",
            ]
        )

    lines.extend(
        [
            "",
            section_titles["files"],
            marker_start["files"],
            marker_end["files"],
            "",
            section_titles["stubs"],
            marker_start["stubs"],
            marker_end["stubs"],
            "",
            section_titles["folders"],
            marker_start["folders"],
            marker_end["folders"],
        ]
    )

    if config.readme_template.include_related_docs:
        lines.extend(
            [
                "",
                "## Related Docs",
                "",
                "Add hand-picked links that help readers continue.",
            ]
        )

    if config.readme_template.include_notes:
        lines.extend(
            [
                "",
                "## Notes",
                "",
                "Add brief context that does not fit above.",
            ]
        )

    return "\n".join(lines)


def parse_managed_entries(
    readme_path: Path,
    text: str,
    config: DocLedgerConfig | None = None,
) -> list[IndexEntry]:
    config = config or default_config()
    marker_start = _marker_start_map(config)
    marker_end = _marker_end_map(config)
    section_titles = _section_titles_map(config)
    entries: list[IndexEntry] = []
    current_section: str | None = None

    for line in text.splitlines():
        section = _section_from_marker_line(line, marker_start)
        if section is not None:
            current_section = section
            continue

        if _is_marker_end_line(line, marker_end):
            current_section = None
            continue

        section = _section_from_heading_line(line, section_titles, _legacy_section_titles_map(config))
        if section is not None:
            current_section = section
            continue

        if _is_any_heading_line(line):
            current_section = None
            continue

        if current_section is None:
            continue

        match = ENTRY_PATTERN.match(line)
        if match is None:
            continue

        link_text, link_target, description = match.groups()
        entries.append(
            IndexEntry(
                readme_path=readme_path,
                section=current_section,
                link_text=link_text,
                link_target=link_target,
                description=description,
                original_line=line,
            )
        )

    return entries


def _has_any_managed_markers(text: str, marker_start: dict[str, str], marker_end: dict[str, str]) -> bool:
    return any(marker in text for marker in marker_start.values()) or any(marker in text for marker in marker_end.values())


def _has_managed_section(
    text: str,
    section: str,
    marker_start: dict[str, str],
    marker_end: dict[str, str],
    section_titles: dict[str, str],
) -> bool:
    return marker_start[section] in text and marker_end[section] in text or section_titles[section] in text


def _render_missing_sections(
    missing_sections: list[str],
    marker_start: dict[str, str],
    marker_end: dict[str, str],
    section_titles: dict[str, str],
) -> str:
    parts: list[str] = []
    for section in missing_sections:
        parts.extend(
            [
                section_titles[section],
                marker_start[section],
                marker_end[section],
            ]
        )
    return "\n\n".join(parts)


def _has_legacy_managed_sections(
    text: str,
    section_titles: dict[str, str],
    legacy_section_titles: dict[str, list[str]],
) -> bool:
    return any(
        _section_from_heading_line(line, section_titles, legacy_section_titles) is not None
        for line in text.splitlines()
    )


def _find_anchor(text: str) -> int | None:
    related = text.find("## Related Docs")
    if related != -1:
        return related
    notes = text.find("## Notes")
    if notes != -1:
        return notes
    return None


def _section_from_marker_line(line: str, marker_start: dict[str, str]) -> str | None:
    for section, marker in marker_start.items():
        if line == marker:
            return section
    return None


def _section_from_heading_line(line: str, section_titles: dict[str, str], legacy_section_titles: dict[str, list[str]]) -> str | None:
    for section, heading in section_titles.items():
        if line == heading:
            return section
    for section, aliases in legacy_section_titles.items():
        if line in aliases:
            return section
    return None


def _is_any_heading_line(line: str) -> bool:
    return bool(HEADING_PATTERN.match(line))


def _wrap_legacy_managed_sections(
    text: str,
    marker_start: dict[str, str],
    marker_end: dict[str, str],
    section_titles: dict[str, str],
    legacy_section_titles: dict[str, list[str]],
) -> str:
    lines = text.splitlines()
    output: list[str] = []
    index = 0

    while index < len(lines):
        line = lines[index]
        section = _section_from_heading_line(line, section_titles, legacy_section_titles)
        if section is None:
            output.append(line)
            index += 1
            continue

        output.append(section_titles[section])
        index += 1
        body: list[str] = []
        while index < len(lines) and not _is_any_heading_line(lines[index]):
            body.append(lines[index])
            index += 1

        output.append(marker_start[section])
        if body:
            output.append("")
            output.extend(body)
        output.append(marker_end[section])

    return "\n".join(output)


def _normalize_legacy_heading_lines(
    text: str,
    section_titles: dict[str, str],
    legacy_section_titles: dict[str, list[str]],
) -> str:
    lines = []
    for line in text.splitlines():
        section = _section_from_heading_line(line, section_titles, legacy_section_titles)
        if section in section_titles and any(line in aliases for aliases in legacy_section_titles.values()):
            lines.append(section_titles[section])
        else:
            lines.append(line)
    return "\n".join(lines)


def _find_managed_section_end(text: str, start: int, section: str, marker_start: dict[str, str]) -> int:
    lines = text.splitlines(keepends=True)
    position = 0
    line_index = 0
    while line_index < len(lines) and position < start:
        position += len(lines[line_index])
        line_index += 1

    while line_index < len(lines):
        line = lines[line_index]
        if _is_any_heading_line(line.lstrip()):
            return position
        if any(marker in line for marker in marker_start.values()) and marker_start[section] not in line:
            return position
        position += len(line)
        line_index += 1

    return len(text)


def _is_marker_end_line(line: str, marker_end: dict[str, str]) -> bool:
    return line in marker_end.values()


def _marker_start_map(config: DocLedgerConfig) -> dict[str, str]:
    prefix = config.markers.prefix
    return {
        "files": render_marker_start(prefix, "files"),
        "stubs": render_marker_start(prefix, "stubs"),
        "folders": render_marker_start(prefix, "folders"),
    }


def _marker_end_map(config: DocLedgerConfig) -> dict[str, str]:
    prefix = config.markers.prefix
    return {
        "files": render_marker_end(prefix, "files"),
        "stubs": render_marker_end(prefix, "stubs"),
        "folders": render_marker_end(prefix, "folders"),
    }


def render_marker_start(prefix: str, section: str) -> str:
    return f"<!-- {prefix}:{section}:start -->"


def render_marker_end(prefix: str, section: str) -> str:
    return f"<!-- {prefix}:{section}:end -->"


def _section_titles_map(config: DocLedgerConfig) -> dict[str, str]:
    return {
        "files": f"## {config.sections.files_heading}",
        "stubs": f"## {config.sections.stubs_heading}",
        "folders": f"## {config.sections.folders_heading}",
    }


def _legacy_section_titles_map(config: DocLedgerConfig) -> dict[str, list[str]]:
    return {
        "files": [f"## {heading}" for heading in config.sections.legacy_files_headings],
        "folders": [f"## {heading}" for heading in config.sections.legacy_folders_headings],
    }
