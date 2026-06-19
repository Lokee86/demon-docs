from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index.parent_index import parent_index_for_file
from docs_index.parent_index import update_parent_index_line


def test_parent_index_for_root_readme_returns_none() -> None:
    root = Path("/tmp/docs")

    assert parent_index_for_file(root / "!README.md", root, lambda _: "ignored") is None


def test_parent_index_for_normal_markdown_uses_same_folder_readme() -> None:
    root = Path("/tmp/docs")
    path = root / "guide" / "intro.md"

    result = parent_index_for_file(path, root, lambda folder: "Guide" if folder == root / "guide" else "Root")

    assert result == "Parent index: [Guide](./!README.md)"


def test_parent_index_for_stub_markdown_uses_parent_folder_title() -> None:
    root = Path("/tmp/docs")
    path = root / "guide" / "stubs" / "intro.md"

    result = parent_index_for_file(
        path,
        root,
        lambda folder: "Guide" if folder == root / "guide" else "Root",
    )

    assert result == "Parent index: [Guide](../!README.md)"


def test_parent_index_for_child_folder_readme_uses_parent_folder_title() -> None:
    root = Path("/tmp/docs")
    path = root / "guide" / "!README.md"

    result = parent_index_for_file(
        path,
        root,
        lambda folder: "Root" if folder == root else "Guide",
    )

    assert result == "Parent index: [Root](../!README.md)"


def test_update_parent_index_line_inserts_after_first_heading() -> None:
    text = """# Title

Intro
"""

    result = update_parent_index_line(text, "Parent index: [Docs](./!README.md)")

    assert result == """# Title

Parent index: [Docs](./!README.md)

Intro"""


def test_update_parent_index_line_inserts_after_double_hash_heading() -> None:
    text = """## Title

Intro
"""

    result = update_parent_index_line(text, "Parent index: [Docs](./!README.md)")

    assert result == """## Title

Parent index: [Docs](./!README.md)

Intro"""


def test_update_parent_index_line_replaces_existing_line() -> None:
    text = """# Title

Parent index: [Old](./!README.md)

Intro
"""

    result = update_parent_index_line(text, "Parent index: [New](./!README.md)")

    assert result == """# Title

Parent index: [New](./!README.md)

Intro
"""


def test_update_parent_index_line_removes_existing_line() -> None:
    text = """# Title

Parent index: [Docs](./!README.md)

Intro
"""

    result = update_parent_index_line(text, None)

    assert result == """# Title

Intro
"""


def test_update_parent_index_line_inserts_at_top_without_heading() -> None:
    text = "Intro\n"

    result = update_parent_index_line(text, "Parent index: [Docs](./!README.md)")

    assert result == "Parent index: [Docs](./!README.md)\n\nIntro\n"
