from __future__ import annotations

from pathlib import Path
import re

from docs_index.model import IndexEntry

MANAGED_SECTION_NAMES = ("files", "stubs", "folders")

SECTION_TITLES = {
    "files": "## Direct Files",
    "stubs": "## Stub Files",
    "folders": "## Direct Folders",
}

MARKER_START = {
    "files": "<!-- docs-index:files:start -->",
    "stubs": "<!-- docs-index:stubs:start -->",
    "folders": "<!-- docs-index:folders:start -->",
}

MARKER_END = {
    "files": "<!-- docs-index:files:end -->",
    "stubs": "<!-- docs-index:stubs:end -->",
    "folders": "<!-- docs-index:folders:end -->",
}

ENTRY_PATTERN = re.compile(r"^\s*-\s+\[([^\]]+)\]\(([^)]+)\)\s+-\s+(.*)$")
HEADING_PATTERN = re.compile(r"^#{1,6}\s+")


def ensure_managed_sections(text: str) -> str:
    if _has_all_managed_sections(text):
        return text

    if _has_legacy_managed_sections(text) and not _has_any_managed_markers(text):
        return _wrap_legacy_managed_sections(text)

    section_block = "\n\n".join(
        (
            "## Direct Files",
            MARKER_START["files"],
            MARKER_END["files"],
            "## Stub Files",
            MARKER_START["stubs"],
            MARKER_END["stubs"],
            "## Direct Folders",
            MARKER_START["folders"],
            MARKER_END["folders"],
        )
    )

    anchor = _find_anchor(text)
    if anchor is None:
        if text:
            return f"{text}\n\n{section_block}"
        return section_block

    before, after = text[:anchor], text[anchor:]
    if before and not before.endswith("\n\n"):
        before = before.rstrip("\n") + "\n\n"
    return f"{before}{section_block}\n\n{after.lstrip()}"


def replace_managed_block(text: str, section: str, lines: list[str]) -> str:
    if section not in MANAGED_SECTION_NAMES:
        raise ValueError(f"unknown managed section: {section}")

    text = ensure_managed_sections(text)
    start_marker = MARKER_START[section]
    end_marker = MARKER_END[section]
    start = text.index(start_marker)
    end = text.find(end_marker, start)
    if end == -1:
        span_end = _find_managed_section_end(text, start, section)
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


def description_from_file(path: Path, is_stub: bool) -> str:
    stem = path.stem.replace("-", " ").replace("_", " ")
    title = " ".join(part.capitalize() for part in stem.split())
    description = f"{title} documentation."
    if is_stub:
        return f"Stub: {description}"
    return description


def description_from_folder(path: Path) -> str:
    title = " ".join(part.capitalize() for part in path.name.replace("-", " ").replace("_", " ").split())
    return f"{title} documentation."


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


def make_readme_template(folder: Path, root: Path, parent_title: str | None) -> str:
    folder_title = title_from_folder(folder)
    lines = [
        f"# {folder_title}",
        "",
        f"This index keeps the {folder_title.lower()} documentation organized and easy to scan.",
    ]
    if folder != root and parent_title is not None:
        lines.extend(
            [
                "",
                f"Parent index: [{parent_title}](../!README.md)",
            ]
        )

    lines.extend(
        [
            "",
            "## Ownership",
            "",
            f"The {folder_title.lower()} docs live here so readers can follow the ownership boundaries without guessing.",
            "",
            "## Does Not Belong",
            "",
            f"Content that describes a different folder should stay in that folder's own index or README.",
            "",
            "## Direct Files",
            MARKER_START["files"],
            MARKER_END["files"],
            "",
            "## Stub Files",
            MARKER_START["stubs"],
            MARKER_END["stubs"],
            "",
            "## Direct Folders",
            MARKER_START["folders"],
            MARKER_END["folders"],
            "",
            "## Related Docs",
            "",
            "Related documentation belongs here when it helps readers move to the next relevant index.",
            "",
            "## Notes",
            "",
            "Use this space for brief context that does not fit in the managed index sections.",
        ]
    )
    return "\n".join(lines)


def parse_managed_entries(readme_path: Path, text: str) -> list[IndexEntry]:
    entries: list[IndexEntry] = []
    current_section: str | None = None

    for line in text.splitlines():
        section = _section_from_marker_line(line)
        if section is not None:
            current_section = section
            continue

        if _is_marker_end_line(line):
            current_section = None
            continue

        section = _section_from_heading_line(line)
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


def _has_all_managed_sections(text: str) -> bool:
    return all(
        MARKER_START[name] in text and MARKER_END[name] in text
        for name in MANAGED_SECTION_NAMES
    )


def _has_any_managed_markers(text: str) -> bool:
    return any(marker in text for marker in MARKER_START.values()) or any(marker in text for marker in MARKER_END.values())


def _has_legacy_managed_sections(text: str) -> bool:
    return any(_section_from_heading_line(line) is not None for line in text.splitlines())


def _find_anchor(text: str) -> int | None:
    related = text.find("## Related Docs")
    if related != -1:
        return related
    notes = text.find("## Notes")
    if notes != -1:
        return notes
    return None


def _section_from_marker_line(line: str) -> str | None:
    for section, marker in MARKER_START.items():
        if line == marker:
            return section
    return None


def _section_from_heading_line(line: str) -> str | None:
    for section, heading in SECTION_TITLES.items():
        if line == heading:
            return section
    return None


def _is_any_heading_line(line: str) -> bool:
    return bool(HEADING_PATTERN.match(line))


def _wrap_legacy_managed_sections(text: str) -> str:
    lines = text.splitlines()
    output: list[str] = []
    index = 0

    while index < len(lines):
        line = lines[index]
        section = _section_from_heading_line(line)
        if section is None:
            output.append(line)
            index += 1
            continue

        output.append(line)
        index += 1
        body: list[str] = []
        while index < len(lines) and not _is_any_heading_line(lines[index]):
            body.append(lines[index])
            index += 1

        output.append(MARKER_START[section])
        if body:
            output.append("")
            output.extend(body)
        output.append(MARKER_END[section])

    return "\n".join(output)


def _find_managed_section_end(text: str, start: int, section: str) -> int:
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
        if any(marker in line for marker in MARKER_START.values()) and MARKER_START[section] not in line:
            return position
        position += len(line)
        line_index += 1

    return len(text)


def _is_marker_end_line(line: str) -> bool:
    return line in MARKER_END.values()
