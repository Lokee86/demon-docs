from __future__ import annotations

from collections.abc import Callable
import re
from pathlib import Path


def parent_index_for_file(
    path: Path,
    root: Path,
    title_lookup: Callable[[Path], str],
) -> str | None:
    if path == root / "!README.md":
        return None

    if path.name == "!README.md":
        parent_folder = path.parent.parent
        parent_title = title_lookup(parent_folder)
        return f"Parent index: [{parent_title}](../!README.md)"

    if path.suffix != ".md":
        return None

    if path.parent.name == "stubs":
        parent_folder = path.parent.parent
        parent_title = title_lookup(parent_folder)
        return f"Parent index: [{parent_title}](../!README.md)"

    parent_title = title_lookup(path.parent)
    return f"Parent index: [{parent_title}](./!README.md)"


def update_parent_index_line(text: str, desired_line: str | None) -> str:
    lines = text.splitlines()
    parent_index_pattern = re.compile(r"^Parent index:\s+.*$")
    parent_index_line_index = next(
        (index for index, line in enumerate(lines) if parent_index_pattern.match(line)),
        None,
    )

    if parent_index_line_index is not None:
        trailing_newline = text.endswith("\n")
        if desired_line is None:
            del lines[parent_index_line_index]
            if (
                parent_index_line_index > 0
                and parent_index_line_index < len(lines)
                and lines[parent_index_line_index - 1] == ""
                and lines[parent_index_line_index] == ""
            ):
                del lines[parent_index_line_index]
        else:
            lines[parent_index_line_index] = desired_line
        rewritten = "\n".join(lines)
        if trailing_newline and not rewritten.endswith("\n"):
            rewritten += "\n"
        return rewritten

    if desired_line is None:
        return text

    heading_index = next(
        (
            index
            for index, line in enumerate(lines)
            if re.match(r"^#{1,6}\s", line)
        ),
        None,
    )

    if heading_index is None:
        if text:
            return f"{desired_line}\n\n{text}"
        return desired_line

    prefix = lines[: heading_index + 1]
    suffix = lines[heading_index + 1 :]
    if suffix and suffix[0] == "":
        suffix = suffix[1:]
    rewritten = prefix + ["", desired_line, ""] + suffix
    return "\n".join(rewritten)
