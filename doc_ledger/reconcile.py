from __future__ import annotations

from pathlib import Path

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import default_config
from doc_ledger.config import is_parent_link_editable
from doc_ledger.model import FileUpdate, FolderInfo, IndexEntry, ReconcileResult
from doc_ledger.path_format import posix_relative_path
from doc_ledger.parent_index import parent_index_for_file
from doc_ledger.parent_index import update_parent_index_line
from doc_ledger.readme_io import description_from_file
from doc_ledger.readme_io import description_from_folder
from doc_ledger.readme_io import ensure_managed_sections
from doc_ledger.readme_io import folder_title
from doc_ledger.readme_io import make_readme_template
from doc_ledger.readme_io import managed_root_title
from doc_ledger.readme_io import parse_managed_entries
from doc_ledger.readme_io import render_file_entry
from doc_ledger.readme_io import render_folder_entry
from doc_ledger.readme_io import replace_managed_block
from doc_ledger.scan import scan_docs_tree

def reconcile_tree(root: Path, config: DocLedgerConfig | None = None) -> ReconcileResult:
    config = config or default_config()
    tree = scan_docs_tree(root, config)
    folders = sorted(tree.folders.values(), key=_folder_sort_key)
    readme_text_by_folder: dict[Path, str] = {}
    existing_entries_by_folder: dict[Path, list[IndexEntry]] = {}
    cross_folder_entries_by_name: dict[str, list[IndexEntry]] = {}
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]] = {}
    current_unmatched_file_counts_by_name: dict[str, int] = {}
    current_unmatched_folder_counts_by_name: dict[str, int] = {}
    updates: list[FileUpdate] = []

    for folder in folders:
        if folder.readme_path is None or not folder.readme_path.exists():
            continue

        current_text = folder.readme_path.read_text(encoding="utf-8")
        readme_text_by_folder[folder.path] = current_text
        existing_entries_by_folder[folder.path] = parse_managed_entries(folder.readme_path, current_text, config)

    root_child_parent_titles = _root_child_parent_titles(
        tree.root,
        folders,
        readme_text_by_folder,
        config.index_file,
        config.parent_link.label,
        config,
    )
    root_display_title = managed_root_title(
        tree.root,
        readme_text_by_folder.get(tree.root),
        root_child_parent_titles,
    )

    title_lookup = _title_lookup(readme_text_by_folder, tree.root, root_display_title)

    for folder in folders:
        if folder.readme_path is None or folder.readme_path.exists():
            continue

        parent_title = None
        if folder.path != tree.root:
            parent_title = title_lookup(folder.path.parent)

        planned_text = make_readme_template(folder.path, tree.root, parent_title, config.index_file, config)
        readme_text_by_folder[folder.path] = planned_text
    cross_folder_entries_by_name = _cross_folder_entries_by_name(
        existing_entries_by_folder,
    )
    cross_folder_folder_entries_by_name = _cross_folder_folder_entries_by_name(
        existing_entries_by_folder,
    )
    current_unmatched_file_counts_by_name = _current_unmatched_file_counts_by_name(
        folders,
        existing_entries_by_folder,
        config.draft.folder,
    )
    current_unmatched_folder_counts_by_name = _current_unmatched_folder_counts_by_name(
        folders,
        existing_entries_by_folder,
        config.index_file,
    )
    matched_entry_ids: set[int] = set()

    for folder in folders:
        if folder.readme_path is not None:
            current_text = readme_text_by_folder[folder.path]
            desired_line = parent_index_for_file(folder.readme_path, tree.root, title_lookup, config)
            new_text = _update_managed_sections(
                folder,
                folder.readme_path,
                update_parent_index_line(current_text, desired_line, config.parent_link.label),
                existing_entries_by_folder.get(folder.path, []),
                cross_folder_entries_by_name,
                current_unmatched_file_counts_by_name,
                cross_folder_folder_entries_by_name,
                current_unmatched_folder_counts_by_name,
                matched_entry_ids,
                config.index_file,
                config.draft.folder,
                config,
            )
            if not folder.readme_path.exists() or new_text != current_text:
                updates.append(
                    FileUpdate(
                        path=folder.readme_path,
                        old_text=current_text if folder.readme_path.exists() else None,
                        new_text=new_text,
                    )
                )

        if folder.is_stubs:
            continue

        for markdown_path in folder.direct_markdown_files + folder.stub_markdown_files:
            if not is_parent_link_editable(markdown_path, config):
                continue
            current_text = markdown_path.read_text(encoding="utf-8")
            desired_line = parent_index_for_file(markdown_path, tree.root, title_lookup, config)
            new_text = update_parent_index_line(current_text, desired_line, config.parent_link.label)
            if new_text != current_text:
                updates.append(
                    FileUpdate(
                        path=markdown_path,
                        old_text=current_text,
                        new_text=new_text,
                    )
                )

    messages = _stale_entry_messages(existing_entries_by_folder, matched_entry_ids, config.index_file)
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


def _title_lookup(readme_text_by_folder: dict[Path, str], root: Path, root_display_title: str):
    def lookup(folder: Path) -> str:
        if folder == root:
            return root_display_title
        readme_text = readme_text_by_folder.get(folder)
        if readme_text is not None:
            return folder_title(folder, readme_text)
        return folder_title(folder)

    return lookup


def _root_child_parent_titles(
    root: Path,
    folders: list[FolderInfo],
    readme_text_by_folder: dict[Path, str],
    index_file: str,
    label: str,
    config: DocLedgerConfig,
) -> list[str]:
    titles: list[str] = []
    root_readme = root / index_file

    for folder in folders:
        if folder.readme_path is not None and folder.readme_path.exists():
            text = readme_text_by_folder.get(folder.path)
            if text is not None:
                title = _parent_index_title_for_root(folder.readme_path, text, root_readme, label)
                if title is not None:
                    titles.append(title)

        for markdown_path in folder.direct_markdown_files + folder.stub_markdown_files:
            if not is_parent_link_editable(markdown_path, config):
                continue
            text = markdown_path.read_text(encoding="utf-8")
            title = _parent_index_title_for_root(markdown_path, text, root_readme, label)
            if title is not None:
                titles.append(title)

    return titles


def _parent_index_title_for_root(file_path: Path, text: str, root_readme: Path, label: str) -> str | None:
    for line in text.splitlines():
        title, target = _parent_index_line_parts(line, label)
        if title is None or target is None:
            continue
        if (file_path.parent / target).resolve(strict=False) == root_readme.resolve(strict=False):
            return title
    return None


def _update_managed_sections(
    folder: FolderInfo,
    readme_path: Path,
    readme_text: str,
    existing_entries: list[IndexEntry],
    cross_folder_entries_by_name: dict[str, list[IndexEntry]],
    current_unmatched_file_counts_by_name: dict[str, int],
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]],
    current_unmatched_folder_counts_by_name: dict[str, int],
    matched_entry_ids: set[int],
    index_file: str,
    draft_folder: str,
    config: DocLedgerConfig,
) -> str:
    ensured = ensure_managed_sections(readme_text, config)
    existing_by_section_and_target = _existing_entries_by_target(readme_path, existing_entries)

    file_lines = [
        _render_file_line(
            readme_path,
            markdown_path,
            False,
            existing_by_section_and_target,
            existing_entries,
            cross_folder_entries_by_name,
            current_unmatched_file_counts_by_name,
            matched_entry_ids,
            config,
            draft_folder,
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
            current_unmatched_file_counts_by_name,
            matched_entry_ids,
            config,
            draft_folder,
        )
        for markdown_path in folder.stub_markdown_files
    ]
    folder_lines = [
        _render_folder_line(
            readme_path,
            child_folder,
            existing_by_section_and_target,
            cross_folder_folder_entries_by_name,
            current_unmatched_folder_counts_by_name,
            matched_entry_ids,
            index_file,
            config,
        )
        for child_folder in folder.direct_subfolders
    ]

    updated = replace_managed_block(ensured, "files", file_lines, config)
    updated = replace_managed_block(updated, "stubs", stub_lines, config)
    updated = replace_managed_block(updated, "folders", folder_lines, config)
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


def _current_unmatched_file_counts_by_name(
    folders: list[FolderInfo],
    existing_entries_by_folder: dict[Path, list[IndexEntry]],
    draft_folder: str,
) -> dict[str, int]:
    counts: dict[str, int] = {}
    for folder in folders:
        if folder.readme_path is None:
            continue

        existing_entries = existing_entries_by_folder.get(folder.path, [])
        existing_by_section_and_target = _existing_entries_by_target(folder.readme_path, existing_entries)

        for markdown_path in folder.direct_markdown_files:
            if _has_stable_or_local_match(
                folder.readme_path,
                markdown_path,
                False,
                existing_entries,
                existing_by_section_and_target,
                draft_folder,
            ):
                continue
            counts[markdown_path.name] = counts.get(markdown_path.name, 0) + 1

        for markdown_path in folder.stub_markdown_files:
            if _has_stable_or_local_match(
                folder.readme_path,
                markdown_path,
                True,
                existing_entries,
                existing_by_section_and_target,
                draft_folder,
            ):
                continue
            counts[markdown_path.name] = counts.get(markdown_path.name, 0) + 1

    return counts


def _current_unmatched_folder_counts_by_name(
    folders: list[FolderInfo],
    existing_entries_by_folder: dict[Path, list[IndexEntry]],
    index_file: str,
) -> dict[str, int]:
    counts: dict[str, int] = {}
    for folder in folders:
        if folder.readme_path is None:
            continue
        if folder.is_stubs:
            continue

        existing_entries = existing_entries_by_folder.get(folder.path, [])
        existing_by_section_and_target = _existing_entries_by_target(folder.readme_path, existing_entries)
        for child_folder in folder.direct_subfolders:
            if _has_stable_or_local_folder_match(
                folder.readme_path,
                child_folder,
                existing_entries,
                existing_by_section_and_target,
                index_file,
            ):
                continue
            counts[child_folder.name] = counts.get(child_folder.name, 0) + 1

    return counts


def _has_stable_or_local_match(
    readme_path: Path,
    markdown_path: Path,
    is_stub: bool,
    existing_entries: list[IndexEntry],
    existing_by_section_and_target: dict[tuple[str, Path], IndexEntry],
    draft_folder: str,
) -> bool:
    section = "stubs" if is_stub else "files"
    if existing_by_section_and_target.get((section, markdown_path.resolve(strict=False))) is not None:
        return True
    if not is_stub:
        return _existing_stub_entry_for_direct_file(readme_path, markdown_path, existing_entries, draft_folder) is not None
    return _existing_direct_entry_for_stub_file(readme_path, markdown_path, existing_entries) is not None


def _has_stable_or_local_folder_match(
    readme_path: Path,
    child_folder: Path,
    existing_entries: list[IndexEntry],
    existing_by_section_and_target: dict[tuple[str, Path], IndexEntry],
    index_file: str,
) -> bool:
    target_path = (child_folder / index_file).resolve(strict=False)
    if existing_by_section_and_target.get(("folders", target_path)) is not None:
        return True
    return _existing_folder_entry_for_child_folder(readme_path, child_folder, existing_entries, index_file) is not None


def _render_file_line(
    readme_path: Path,
    markdown_path: Path,
    is_stub: bool,
    existing_by_section_and_target: dict[tuple[str, Path], IndexEntry],
    existing_entries: list[IndexEntry],
    cross_folder_entries_by_name: dict[str, list[IndexEntry]],
    current_unmatched_file_counts_by_name: dict[str, int],
    matched_entry_ids: set[int],
    config: DocLedgerConfig,
    draft_folder: str,
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
        graduated_stub = _existing_stub_entry_for_direct_file(readme_path, markdown_path, existing_entries, draft_folder)
        if graduated_stub is not None:
            matched_entry_ids.add(id(graduated_stub))
            return render_file_entry(
                graduated_stub.link_text,
                _canonical_target(readme_path, markdown_path),
                _promote_stub_description(graduated_stub.description, config),
            )
    else:
        canonical_direct = _existing_direct_entry_for_stub_file(readme_path, markdown_path, existing_entries)
        if canonical_direct is not None:
            matched_entry_ids.add(id(canonical_direct))
            return render_file_entry(
                canonical_direct.link_text,
                _canonical_target(readme_path, markdown_path),
                _ensure_stub_description(canonical_direct.description, config),
            )

    moved_entry = _cross_folder_entry_for_file(
        markdown_path,
        cross_folder_entries_by_name,
        current_unmatched_file_counts_by_name,
    )
    if moved_entry is not None:
        matched_entry_ids.add(id(moved_entry))
        return render_file_entry(
            markdown_path.name,
            _canonical_target(readme_path, markdown_path),
            _moved_description(moved_entry.description, is_stub=is_stub, config=config),
        )

    return render_file_entry(
        markdown_path.name,
        _canonical_target(readme_path, markdown_path),
        description_from_file(markdown_path, is_stub=is_stub, config=config),
    )


def _render_folder_line(
    readme_path: Path,
    child_folder: Path,
    existing_by_section_and_target: dict[tuple[str, Path], IndexEntry],
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]],
    current_unmatched_folder_counts_by_name: dict[str, int],
    matched_entry_ids: set[int],
    index_file: str,
    config: DocLedgerConfig,
) -> str:
    target_path = child_folder / index_file
    existing = existing_by_section_and_target.get(("folders", target_path.resolve(strict=False)))
    if existing is not None:
        matched_entry_ids.add(id(existing))
        return render_folder_entry(
            existing.link_text,
            _canonical_target(readme_path, target_path),
            existing.description,
        )

    moved_entry = _cross_folder_entry_for_folder(
        child_folder,
        cross_folder_folder_entries_by_name,
        current_unmatched_folder_counts_by_name,
    )
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
        description_from_folder(child_folder, config),
    )


def _canonical_target(readme_path: Path, target_path: Path) -> str:
    return posix_relative_path(target_path, readme_path.parent)


def _existing_stub_entry_for_direct_file(
    readme_path: Path,
    markdown_path: Path,
    existing_entries: list[IndexEntry],
    draft_folder: str,
) -> IndexEntry | None:
    expected_target = (readme_path.parent / draft_folder / markdown_path.name).resolve(strict=False)
    for entry in existing_entries:
        if entry.section != "stubs":
            continue
        if (readme_path.parent / entry.link_target).resolve(strict=False) != expected_target:
            continue
        return entry
    return None


def _promote_stub_description(description: str, config: DocLedgerConfig) -> str:
    prefix = config.draft.description_prefix
    if not description.startswith(prefix):
        return description

    promoted = description[len(prefix) :]
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


def _ensure_stub_description(description: str, config: DocLedgerConfig) -> str:
    prefix = config.draft.description_prefix
    if description.startswith(prefix):
        return description
    return f"{prefix}{description}"


def _existing_folder_entry_for_child_folder(
    readme_path: Path,
    child_folder: Path,
    existing_entries: list[IndexEntry],
    index_file: str,
) -> IndexEntry | None:
    expected_target = (child_folder / index_file).resolve(strict=False)
    for entry in existing_entries:
        if entry.section != "folders":
            continue
        if (readme_path.parent / entry.link_target).resolve(strict=False) != expected_target:
            continue
        return entry
    return None


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
    current_unmatched_file_counts_by_name: dict[str, int],
) -> IndexEntry | None:
    if current_unmatched_file_counts_by_name.get(markdown_path.name, 0) != 1:
        return None
    entries = cross_folder_entries_by_name.get(markdown_path.name)
    if entries is None or len(entries) != 1:
        return None
    return entries[0]


def _cross_folder_entry_for_folder(
    child_folder: Path,
    cross_folder_folder_entries_by_name: dict[str, list[IndexEntry]],
    current_unmatched_folder_counts_by_name: dict[str, int],
) -> IndexEntry | None:
    if current_unmatched_folder_counts_by_name.get(child_folder.name, 0) != 1:
        return None
    entries = cross_folder_folder_entries_by_name.get(child_folder.name)
    if entries is None or len(entries) != 1:
        return None
    return entries[0]


def _moved_description(description: str, is_stub: bool, config: DocLedgerConfig) -> str:
    if is_stub:
        return _ensure_stub_description(description, config)
    return _promote_stub_description(description, config)


def _stale_entry_messages(
    existing_entries_by_folder: dict[Path, list[IndexEntry]],
    matched_entry_ids: set[int],
    index_file: str,
) -> list[str]:
    messages: list[str] = []
    for folder_path, entries in existing_entries_by_folder.items():
        for entry in entries:
            if id(entry) in matched_entry_ids:
                continue
            messages.append(f"Removed stale {entry.section} entry from {folder_path / index_file}: {entry.original_line}")
    return messages


def _parent_index_line_parts(line: str, label: str | None = None) -> tuple[str | None, str | None]:
    import re

    if label is None:
        label = default_config().parent_link.label

    pattern = rf"^{re.escape(label)}:\s+\[([^\]]+)\]\(([^)]+)\)\s*$"
    match = re.match(pattern, line)
    if match is None:
        return None, None
    return match.group(1), match.group(2)
