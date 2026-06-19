from __future__ import annotations

import argparse
import os
from pathlib import Path
from typing import Sequence

from doc_ledger import __version__
from doc_ledger.config import default_config
from doc_ledger.config import global_config_path
from doc_ledger.config import local_config_path
from doc_ledger.config import selected_config_path
from doc_ledger.config import load_config
from doc_ledger.config import starter_config_text
from doc_ledger.reconcile import apply_updates
from doc_ledger.reconcile import reconcile_tree
from doc_ledger.paths import resolve_root
from doc_ledger.watch import watch_root


def _add_root_argument(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("--root", default=None, help="docs root directory to reconcile")


def _add_config_argument(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("--config", default=None, help="explicit doc-ledger config file")


def _add_config_selection_arguments(parser: argparse.ArgumentParser) -> None:
    parser.add_argument(
        "--no-local-config",
        action="store_true",
        default=False,
        help="skip the local config in the current directory",
    )
    parser.add_argument(
        "--no-global-config",
        action="store_true",
        default=False,
        help="skip the global config fallback",
    )


def _add_config_override_arguments(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("--index-file", default=None, help="override the folder index filename")
    parser.add_argument("--draft-folder", default=None, help="override the draft folder name")
    parser.add_argument(
        "--draft-description-prefix",
        default=None,
        help="override the draft file description prefix",
    )
    parser.add_argument(
        "--include",
        dest="include_patterns",
        action="append",
        default=None,
        help="add an include pattern for indexed files",
    )
    parser.add_argument(
        "--exclude",
        dest="exclude_patterns",
        action="append",
        default=None,
        help="add an exclude pattern for indexed files",
    )
    parser.add_argument("--marker-prefix", default=None, help="override the managed marker prefix")
    parser.add_argument("--parent-label", default=None, help="override the parent link label")

    boolean_optional_action = getattr(argparse, "BooleanOptionalAction", None)
    if boolean_optional_action is None:  # pragma: no cover - compatibility fallback
        parser.add_argument(
            "--parent-link-folder-indexes",
            dest="parent_link_folder_indexes",
            action="store_true",
            default=None,
            help="enable parent links in folder indexes",
        )
        parser.add_argument(
            "--no-parent-link-folder-indexes",
            dest="parent_link_folder_indexes",
            action="store_false",
            help=argparse.SUPPRESS,
        )
        parser.add_argument(
            "--parent-link-indexed-files",
            dest="parent_link_indexed_files",
            action="store_true",
            default=None,
            help="enable parent links in indexed files",
        )
        parser.add_argument(
            "--no-parent-link-indexed-files",
            dest="parent_link_indexed_files",
            action="store_false",
            help=argparse.SUPPRESS,
        )
        return

    parser.add_argument(
        "--parent-link-folder-indexes",
        dest="parent_link_folder_indexes",
        action=boolean_optional_action,
        default=None,
        help="enable parent links in folder indexes",
    )
    parser.add_argument(
        "--parent-link-indexed-files",
        dest="parent_link_indexed_files",
        action=boolean_optional_action,
        default=None,
        help="enable parent links in indexed files",
    )


def _apply_config_overrides(config, args: argparse.Namespace) -> None:
    if getattr(args, "index_file", None) is not None:
        config.index_file = str(args.index_file)
        config.file.index_file = str(args.index_file)
    if getattr(args, "draft_folder", None) is not None:
        config.draft.folder = str(args.draft_folder)
    if getattr(args, "draft_description_prefix", None) is not None:
        config.draft.description_prefix = str(args.draft_description_prefix)
    if getattr(args, "include_patterns", None) is not None:
        config.file.include_patterns = [str(pattern) for pattern in args.include_patterns]
    if getattr(args, "exclude_patterns", None) is not None:
        config.file.exclude_patterns = [str(pattern) for pattern in args.exclude_patterns]
    if getattr(args, "marker_prefix", None) is not None:
        config.markers.prefix = str(args.marker_prefix)
    if getattr(args, "parent_label", None) is not None:
        config.parent_link.label = str(args.parent_label)
    if getattr(args, "parent_link_folder_indexes", None) is not None:
        config.parent_link.folder_indexes = bool(args.parent_link_folder_indexes)
    if getattr(args, "parent_link_indexed_files", None) is not None:
        config.parent_link.indexed_files = bool(args.parent_link_indexed_files)


def _format_config_value(value) -> str:
    if isinstance(value, str):
        return repr(value)
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, float):
        return repr(value)
    if isinstance(value, list):
        return "[" + ", ".join(_format_config_value(item) for item in value) + "]"
    return repr(value)


def _format_config_show(config, config_path: Path | None) -> str:
    selected_path = str(config_path) if config_path is not None else "<built-in defaults>"
    lines = [
        f"selected_config_path = {selected_path}",
        f"root = {_format_config_value(config.root)}",
        f"index_file = {_format_config_value(config.index_file)}",
        "[markers]",
        f"prefix = {_format_config_value(config.markers.prefix)}",
        "[parent_link]",
        f"label = {_format_config_value(config.parent_link.label)}",
        f"folder_indexes = {_format_config_value(config.parent_link.folder_indexes)}",
        f"indexed_files = {_format_config_value(config.parent_link.indexed_files)}",
        "[drafts]",
        f"folder = {_format_config_value(config.draft.folder)}",
        f"description_prefix = {_format_config_value(config.draft.description_prefix)}",
        "[files]",
        f"include_patterns = {_format_config_value(config.file.include_patterns)}",
        f"exclude_patterns = {_format_config_value(config.file.exclude_patterns)}",
    ]
    return "\n".join(lines)


def _print_config_paths(parser: argparse.ArgumentParser) -> int:
    cwd = Path.cwd()
    local_dot = cwd / ".doc-ledger.toml"
    local_plain = cwd / "doc-ledger.toml"
    selected_local = local_config_path(cwd)
    global_path = global_config_path(env=os.environ, home=Path.home())
    selected_config = selected_config_path(
        cwd=cwd,
        explicit_config=None,
        no_local=False,
        no_global=False,
        env=os.environ,
        home=Path.home(),
    )

    print(f"cwd = {cwd}")
    print(f"local dot config = {local_dot} exists={local_dot.exists()}")
    print(f"local plain config = {local_plain} exists={local_plain.exists()}")
    print(f"selected local config = {selected_local if selected_local is not None else '<none>'}")
    print(f"global config = {global_path} exists={global_path.exists()}")
    print(f"selected config = {selected_config if selected_config is not None else '<none>'}")
    return 0


def _init_config_file(
    parser: argparse.ArgumentParser,
    *,
    is_local: bool,
    force: bool,
    env: os._Environ[str],
    home: Path,
) -> int:
    target_path = Path.cwd() / ".doc-ledger.toml" if is_local else global_config_path(env=env, home=home)
    target_path.parent.mkdir(parents=True, exist_ok=True)

    if target_path.exists() and not force:
        parser.exit(2, f"doc-ledger error: config file already exists: {target_path}\n")

    target_path.write_text(starter_config_text(), encoding="utf-8")
    print(target_path)
    return 0


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="doc-ledger",
        description="doc-ledger reconciles index files with a folder tree.",
        epilog=(
            "Examples:\n"
            "  doc-ledger fix --root docs\n"
            "  doc-ledger check --root docs\n"
            "  doc-ledger watch --root docs\n"
            "  doc-ledger fix --config .doc-ledger.toml\n"
            "  doc-ledger --version"
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("-v", "--version", action="version", version=f"doc-ledger {__version__}")
    subparsers = parser.add_subparsers(dest="command", required=True)

    fix_parser = subparsers.add_parser(
        "fix",
        help="reconcile and write updated files",
        description="Reconcile the docs tree and write any needed updates.",
    )
    _add_root_argument(fix_parser)
    _add_config_argument(fix_parser)
    _add_config_selection_arguments(fix_parser)
    _add_config_override_arguments(fix_parser)
    fix_parser.set_defaults(command="fix")

    check_parser = subparsers.add_parser(
        "check",
        help="reconcile without writing files",
        description="Verify that the docs tree is already reconciled.",
    )
    _add_root_argument(check_parser)
    _add_config_argument(check_parser)
    _add_config_selection_arguments(check_parser)
    _add_config_override_arguments(check_parser)
    check_parser.set_defaults(command="check")

    watch_parser = subparsers.add_parser(
        "watch",
        help="watch the tree and rerun reconciliation",
        description="Watch the docs tree and rerun reconciliation when relevant files change.",
    )
    _add_root_argument(watch_parser)
    _add_config_argument(watch_parser)
    _add_config_selection_arguments(watch_parser)
    _add_config_override_arguments(watch_parser)
    watch_parser.add_argument("--once", action="store_true", help="run one reconciliation pass and exit")
    watch_parser.set_defaults(command="watch")

    config_parser = subparsers.add_parser(
        "config",
        help="inspect config path selection and resolved config",
        description="Inspect config path selection and show the resolved selected config.",
    )
    config_subparsers = config_parser.add_subparsers(dest="config_command", required=True)

    config_paths_parser = config_subparsers.add_parser(
        "paths",
        help="show config path candidates",
        description="Show current-directory, local, and global config paths.",
    )
    config_paths_parser.set_defaults(command="config", config_command="paths")

    config_show_parser = config_subparsers.add_parser(
        "show",
        help="show the resolved selected config",
        description="Show the resolved selected config after config-file selection.",
    )
    _add_config_argument(config_show_parser)
    _add_config_selection_arguments(config_show_parser)
    config_show_parser.set_defaults(command="config", config_command="show")

    config_init_parser = config_subparsers.add_parser(
        "init",
        help="write a starter config file",
        description="Write a starter config file in the current directory or global config location.",
    )
    config_init_target = config_init_parser.add_mutually_exclusive_group(required=True)
    config_init_target.add_argument(
        "--local",
        action="store_true",
        default=False,
        help="write .doc-ledger.toml in the current directory",
    )
    config_init_target.add_argument(
        "--global",
        dest="global_config",
        action="store_true",
        default=False,
        help="write the global config file",
    )
    config_init_parser.add_argument("--force", action="store_true", default=False, help="overwrite an existing config file")
    config_init_parser.set_defaults(command="config", config_command="init")

    return parser


def _load_cli_config(args: argparse.Namespace, parser: argparse.ArgumentParser):
    config_path = selected_config_path(
        cwd=Path.cwd(),
        explicit_config=Path(args.config) if args.config is not None else None,
        no_local=bool(args.no_local_config),
        no_global=bool(args.no_global_config),
        env=os.environ,
        home=Path.home(),
    )

    if config_path is None:
        return default_config(), None

    try:
        return load_config(config_path), config_path
    except FileNotFoundError:
        parser.exit(2, f"doc-ledger error: config file does not exist: {config_path}\n")
    except Exception as exc:  # pragma: no cover - defensive CLI boundary
        parser.exit(2, f"doc-ledger error: {exc}\n")


def _resolve_cli_root(root_arg: str | None, config_root: str, config_path: Path | None) -> Path:
    if root_arg is not None:
        return resolve_root(root_arg)

    if config_path is not None:
        return resolve_root(config_root, cwd=config_path.parent)

    return resolve_root(config_root)


def main(argv: Sequence[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)
    if args.command == "config" and args.config_command == "paths":
        return _print_config_paths(parser)
    if args.command == "config" and args.config_command == "init":
        return _init_config_file(
            parser,
            is_local=bool(args.local),
            force=bool(args.force),
            env=os.environ,
            home=Path.home(),
        )
    config, config_path = _load_cli_config(args, parser)

    if args.command == "config" and args.config_command == "show":
        print(_format_config_show(config, config_path))
        return 0

    _apply_config_overrides(config, args)

    if args.command in {"fix", "check"}:
        try:
            root = _resolve_cli_root(args.root, config.root, config_path)
            if not root.exists():
                parser.error(f"docs root does not exist: {root}")

            result = reconcile_tree(root, config)
        except SystemExit:
            raise
        except Exception as exc:  # pragma: no cover - defensive CLI boundary
            parser.exit(2, f"doc-ledger error: {exc}\n")

        if args.command == "fix":
            changed = apply_updates(result)
            print(f"doc-ledger fix updated {changed} file(s)")
            if result.messages:
                for message in result.messages:
                    print(f"message: {message}")
            return 0

        if result.updates:
            print("doc-ledger check failed")
            for update in result.updates:
                print(str(update.path))
            for message in result.messages:
                print(f"message: {message}")
            return 1

        print("doc-ledger check passed")
        return 0
    elif args.command == "watch":
        try:
            root = _resolve_cli_root(args.root, config.root, config_path)
            if not root.exists():
                parser.error(f"docs root does not exist: {root}")

            return watch_root(root, config, once=args.once)
        except SystemExit:
            raise
        except Exception as exc:  # pragma: no cover - defensive CLI boundary
            parser.exit(2, f"doc-ledger error: {exc}\n")

    return 0
