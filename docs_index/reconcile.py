from __future__ import annotations

from pathlib import Path

from docs_index.model import FileUpdate, FolderInfo, IndexEntry, ReconcileResult
from docs_index.parent_index import parent_index_for_file
from docs_index.parent_index import update_parent_index_line
from docs_index.readme_io import description_from_file
from docs_index.readme_io import description_from_folder
from docs_index.readme_io import ensure_managed_sections
from docs_index.readme_io import folder_title
from docs_index.readme_io import make_readme_template
from docs_index.readme_io import parse_managed_entries
from docs_index.readme_io import render_file_entry
from docs_index.readme_io import render_folder_entry
from docs_index.readme_io import replace_managed_block
from docs_index.scan import scan_docs_tree


def reconcile_tree(root: Path) -> ReconcileResult:
    tree = scan_docs_tree(root)
    folders = sorted(tree.folders.values(), key=_folder_sort_key)
    readme_text_by_folder: dict[Path, str] = {}
    existing_entries_by_folder: dict[Path, list[IndexEntry]] = {}
    cross_folder_entries_by_name: dict[str, list[IndexEntry]] = {}
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]] = {}
    updates: list[FileUpdate] = []

    for folder in folders:
        if folder.readme_path is None:
            continue

        if folder.readme_path.exists():
            current_text = folder.readme_path.read_text(encoding="utf-8")
            readme_text_by_folder[folder.path] = current_text
            existing_entries_by_folder[folder.path] = parse_managed_entries(folder.readme_path, current_text)
            continue

        parent_title = None
        if folder.path != tree.root:
            parent_title = folder_title(folder.path.parent, readme_text_by_folder.get(folder.path.parent))

        planned_text = make_readme_template(folder.path, tree.root, parent_title)
        readme_text_by_folder[folder.path] = planned_text

    title_lookup = _title_lookup(readme_text_by_folder)
    cross_folder_entries_by_name = _cross_folder_entries_by_name(
        existing_entries_by_folder,
    )
    cross_folder_folder_entries_by_name = _cross_folder_folder_entries_by_name(
        existing_entries_by_folder,
    )
    matched_entry_ids: set[int] = set()

    for folder in folders:
        if folder.readme_path is not None:
            current_text = readme_text_by_folder[folder.path]
            desired_line = parent_index_for_file(folder.readme_path, tree.root, title_lookup)
            new_text = _update_managed_sections(
                folder,
                folder.readme_path,
                update_parent_index_line(current_text, desired_line),
                existing_entries_by_folder.get(folder.path, []),
                cross_folder_entries_by_name,
                cross_folder_folder_entries_by_name,
                matched_entry_ids,
            )
            if not folder.readme_path.exists() or new_text != current_text:
                updates.append(
                    FileUpdate(
                        path=folder.readme_path,
                        old_text=current_text if folder.readme_path.exists() else None,
                        new_text=new_text,
                    )
                )

        for markdown_path in folder.direct_markdown_files + folder.stub_markdown_files:
            current_text = markdown_path.read_text(encoding="utf-8")
            desired_line = parent_index_for_file(markdown_path, tree.root, title_lookup)
            new_text = update_parent_index_line(current_text, desired_line)
            if new_text != current_text:
                updates.append(
                    FileUpdate(
                        path=markdown_path,
                        old_text=current_text,
                        new_text=new_text,
                    )
                )

    messages = _stale_entry_messages(existing_entries_by_folder, matched_entry_ids)
    return ReconcileResult(updates=updates, messages=messages)


def apply_updates(result: ReconcileResult) -> int:
    changed = 0
    for update in result.updates:
        if update.old_text is not None and update.old_text == update.new_text:
            continue

        update.path.parent.mkdir(parents=True, exist_ok=True)
        update.path.write_text(update.new_text, encoding="utf-8")
        changed += 1

    return changed


def _title_lookup(readme_text_by_folder: dict[Path, str]):
    def lookup(folder: Path) -> str:
        readme_text = readme_text_by_folder.get(folder)
        if readme_text is not None:
            return folder_title(folder, readme_text)
        return folder_title(folder)

    return lookup


def _update_managed_sections(
    folder: FolderInfo,
    readme_path: Path,
    readme_text: str,
    existing_entries: list[IndexEntry],
    cross_folder_entries_by_name: dict[str, list[IndexEntry]],
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]],
    matched_entry_ids: set[int],
) -> str:
    ensured = ensure_managed_sections(readme_text)
    existing_by_section_and_target = _existing_entries_by_target(readme_path, existing_entries)

    file_lines = [
        _render_file_line(
            readme_path,
            markdown_path,
            False,
            existing_by_section_and_target,
            existing_entries,
            cross_folder_entries_by_name,
            matched_entry_ids,
        )
        for markdown_path in folder.direct_markdown_files
    ]
    stub_lines = [
        _render_file_line(
            readme_path,
            markdown_path,
            True,
            existing_by_section_and_target,
            existing_entries,
            cross_folder_entries_by_name,
            matched_entry_ids,
        )
        for markdown_path in folder.stub_markdown_files
    ]
    folder_lines = [
        _render_folder_line(
            readme_path,
            child_folder,
            existing_by_section_and_target,
            cross_folder_folder_entries_by_name,
            matched_entry_ids,
        )
        for child_folder in folder.direct_subfolders
    ]

    updated = replace_managed_block(ensured, "files", file_lines)
    updated = replace_managed_block(updated, "stubs", stub_lines)
    updated = replace_managed_block(updated, "folders", folder_lines)
    return updated


def _folder_sort_key(folder: FolderInfo) -> tuple[int, str]:
    folder_path = folder.path
    return (len(folder_path.parts), str(folder_path))


def _existing_entries_by_target(
    readme_path: Path,
    entries: list[IndexEntry],
) -> dict[tuple[str, Path], IndexEntry]:
    result: dict[tuple[str, Path], IndexEntry] = {}
    for entry in entries:
        target_path = (readme_path.parent / entry.link_target).resolve(strict=False)
        result[(entry.section, target_path)] = entry
    return result


def _render_file_line(
    readme_path: Path,
    markdown_path: Path,
    is_stub: bool,
    existing_by_section_and_target: dict[tuple[str, Path], IndexEntry],
    existing_entries: list[IndexEntry],
    cross_folder_entries_by_name: dict[str, list[IndexEntry]],
    matched_entry_ids: set[int],
) -> str:
    section = "stubs" if is_stub else "files"
    existing = existing_by_section_and_target.get((section, markdown_path.resolve(strict=False)))
    if existing is not None:
        matched_entry_ids.add(id(existing))
        return render_file_entry(
            existing.link_text,
            _canonical_target(readme_path, markdown_path),
            existing.description,
        )

    if not is_stub:
        graduated_stub = _existing_stub_entry_for_direct_file(readme_path, markdown_path, existing_entries)
        if graduated_stub is not None:
            matched_entry_ids.add(id(graduated_stub))
            return render_file_entry(
                graduated_stub.link_text,
                _canonical_target(readme_path, markdown_path),
                _promote_stub_description(graduated_stub.description),
            )
    else:
        canonical_direct = _existing_direct_entry_for_stub_file(readme_path, markdown_path, existing_entries)
        if canonical_direct is not None:
            matched_entry_ids.add(id(canonical_direct))
            return render_file_entry(
                canonical_direct.link_text,
                _canonical_target(readme_path, markdown_path),
                _ensure_stub_description(canonical_direct.description),
            )

    moved_entry = _cross_folder_entry_for_file(markdown_path, cross_folder_entries_by_name)
    if moved_entry is not None:
        matched_entry_ids.add(id(moved_entry))
        return render_file_entry(
            markdown_path.name,
            _canonical_target(readme_path, markdown_path),
            _moved_description(moved_entry.description, is_stub=is_stub),
        )

    return render_file_entry(
        markdown_path.name,
        _canonical_target(readme_path, markdown_path),
        description_from_file(markdown_path, is_stub=is_stub),
    )


def _render_folder_line(
    readme_path: Path,
    child_folder: Path,
    existing_by_section_and_target: dict[tuple[str, Path], IndexEntry],
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]],
    matched_entry_ids: set[int],
) -> str:
    target_path = child_folder / "!README.md"
    existing = existing_by_section_and_target.get(("folders", target_path.resolve(strict=False)))
    if existing is not None:
        matched_entry_ids.add(id(existing))
        return render_folder_entry(
            existing.link_text,
            _canonical_target(readme_path, target_path),
            existing.description,
        )

    moved_entry = _cross_folder_entry_for_folder(child_folder, cross_folder_folder_entries_by_name)
    if moved_entry is not None:
        matched_entry_ids.add(id(moved_entry))
        return render_folder_entry(
            moved_entry.link_text,
            _canonical_target(readme_path, target_path),
            moved_entry.description,
        )

    return render_folder_entry(
        child_folder.name,
        _canonical_target(readme_path, target_path),
        description_from_folder(child_folder),
    )


def _canonical_target(readme_path: Path, target_path: Path) -> str:
    return str(target_path.relative_to(readme_path.parent))


def _existing_stub_entry_for_direct_file(
    readme_path: Path,
    markdown_path: Path,
    existing_entries: list[IndexEntry],
) -> IndexEntry | None:
    expected_target = (readme_path.parent / "stubs" / markdown_path.name).resolve(strict=False)
    for entry in existing_entries:
        if entry.section != "stubs":
            continue
        if (readme_path.parent / entry.link_target).resolve(strict=False) != expected_target:
            continue
        return entry
    return None


def _promote_stub_description(description: str) -> str:
    if not description.startswith("Stub: "):
        return description

    promoted = description[len("Stub: ") :]
    if promoted and promoted[0].islower():
        promoted = promoted[0].upper() + promoted[1:]
    return promoted


def _existing_direct_entry_for_stub_file(
    readme_path: Path,
    markdown_path: Path,
    existing_entries: list[IndexEntry],
) -> IndexEntry | None:
    expected_target = (readme_path.parent / markdown_path.name).resolve(strict=False)
    for entry in existing_entries:
        if entry.section != "files":
            continue
        if (readme_path.parent / entry.link_target).resolve(strict=False) != expected_target:
            continue
        return entry
    return None


def _ensure_stub_description(description: str) -> str:
    if description.startswith("Stub: "):
        return description
    return f"Stub: {description}"


def _cross_folder_entries_by_name(
    existing_entries_by_folder: dict[Path, list[IndexEntry]],
) -> dict[str, list[IndexEntry]]:
    result: dict[str, list[IndexEntry]] = {}
    for folder_entries in existing_entries_by_folder.values():
        for entry in folder_entries:
            if entry.section not in {"files", "stubs"}:
                continue
            target_path = (entry.readme_path.parent / entry.link_target).resolve(strict=False)
            if target_path.exists():
                continue
            result.setdefault(target_path.name, []).append(entry)
    return result


def _cross_folder_folder_entries_by_name(
    existing_entries_by_folder: dict[Path, list[IndexEntry]],
) -> dict[str, list[IndexEntry]]:
    result: dict[str, list[IndexEntry]] = {}
    for folder_entries in existing_entries_by_folder.values():
        for entry in folder_entries:
            if entry.section != "folders":
                continue
            target_path = (entry.readme_path.parent / entry.link_target).resolve(strict=False)
            if target_path.exists():
                continue
            result.setdefault(target_path.parent.name, []).append(entry)
    return result


def _cross_folder_entry_for_file(
    markdown_path: Path,
    cross_folder_entries_by_name: dict[str, list[IndexEntry]],
) -> IndexEntry | None:
    entries = cross_folder_entries_by_name.get(markdown_path.name)
    if entries is None or len(entries) != 1:
        return None
    return entries[0]


def _cross_folder_entry_for_folder(
    child_folder: Path,
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]],
) -> IndexEntry | None:
    entries = cross_folder_folder_entries_by_name.get(child_folder.name)
    if entries is None or len(entries) != 1:
        return None
    return entries[0]


def _moved_description(description: str, is_stub: bool) -> str:
    if is_stub:
        return _ensure_stub_description(description)
    return _promote_stub_description(description)


def _stale_entry_messages(
    existing_entries_by_folder: dict[Path, list[IndexEntry]],
    matched_entry_ids: set[int],
) -> list[str]:
    messages: list[str] = []
    for folder_path, entries in existing_entries_by_folder.items():
        for entry in entries:
            if id(entry) in matched_entry_ids:
                continue
            messages.append(f"Removed stale {entry.section} entry from {folder_path / '!README.md'}: {entry.original_line}")
    return messages
