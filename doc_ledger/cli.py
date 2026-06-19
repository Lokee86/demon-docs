from __future__ import annotations

import argparse
from pathlib import Path
from typing import Sequence

from doc_ledger.config import default_config
from doc_ledger.config import discover_config
from doc_ledger.config import load_config
from doc_ledger.reconcile import apply_updates
from doc_ledger.reconcile import reconcile_tree
from doc_ledger.paths import resolve_root
from doc_ledger.watch import watch_root


def _add_root_argument(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("--root", default=None, help="docs root directory")


def _add_config_argument(parser: argparse.ArgumentParser) -> None:
    parser.add_argument("--config", default=None, help="doc-ledger config file")


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="doc-ledger")
    subparsers = parser.add_subparsers(dest="command", required=True)

    fix_parser = subparsers.add_parser("fix", help="fix doc-ledger issues")
    _add_root_argument(fix_parser)
    _add_config_argument(fix_parser)
    fix_parser.set_defaults(command="fix")

    check_parser = subparsers.add_parser("check", help="check doc-ledger issues")
    _add_root_argument(check_parser)
    _add_config_argument(check_parser)
    check_parser.set_defaults(command="check")

    watch_parser = subparsers.add_parser("watch", help="watch doc-ledger changes")
    _add_root_argument(watch_parser)
    _add_config_argument(watch_parser)
    watch_parser.add_argument("--once", action="store_true", help="run once and exit")
    watch_parser.set_defaults(command="watch")

    return parser


def _load_cli_config(config_path: str | None, parser: argparse.ArgumentParser):
    if config_path is None:
        discovered_config = discover_config(Path.cwd())
        if discovered_config is None:
            return default_config(), None
        config_path = discovered_config
    else:
        config_path = Path(config_path)

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
    config, config_path = _load_cli_config(args.config, parser)

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
