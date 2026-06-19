from __future__ import annotations

from collections.abc import Callable
from datetime import datetime
import os
from pathlib import Path
import threading
import time

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import default_config
from doc_ledger.reconcile import apply_updates
from doc_ledger.reconcile import reconcile_tree
from doc_ledger.scan import _is_indexable_file


class WatchScheduler:
    def __init__(
        self,
        run_fix: Callable[[], None],
        debounce_seconds: float,
        clock: Callable[[], float] = time.monotonic,
    ) -> None:
        self._run_fix = run_fix
        self._debounce_seconds = debounce_seconds
        self._clock = clock
        self._lock = threading.Lock()
        self._pending_changes = 0
        self._running = False
        self._last_change_at = 0.0

    def mark_changed(self) -> None:
        with self._lock:
            self._pending_changes += 1
            self._last_change_at = self._clock()

    def run_once_if_pending(self) -> bool:
        with self._lock:
            if self._running or self._pending_changes == 0:
                return False

            if self._debounce_seconds > 0:
                elapsed = self._clock() - self._last_change_at
                if elapsed + 1e-9 < self._debounce_seconds:
                    return False

            self._running = True
            self._pending_changes = 0
        try:
            self._run_fix()
        finally:
            with self._lock:
                self._running = False

        return True


def _status_line(message: str, now: Callable[[], datetime] = datetime.now) -> str:
    return f"{now().isoformat(timespec='seconds')} {message}"


def watch_root(
    root: Path,
    config: DocLedgerConfig | None = None,
    debounce_seconds: float | None = None,
    once: bool = False,
) -> int:
    config = config or default_config()
    debounce_seconds = config.watch.debounce_seconds if debounce_seconds is None else debounce_seconds
    print(_status_line(f"doc-ledger watch watching {root} pid={os.getpid()}"))
    if once:
        _run_fix_and_report(root, config)
        return 0

    _run_fix_and_report(root, config)

    from watchdog.events import FileSystemEventHandler
    from watchdog.observers import Observer

    scheduler = WatchScheduler(lambda: _run_fix_and_report(root, config), debounce_seconds=debounce_seconds)

    class DocsIndexEventHandler(FileSystemEventHandler):
        def on_any_event(self, event) -> None:  # type: ignore[override]
            if _is_relevant_watch_event(event, config, root=root):
                scheduler.mark_changed()

    observer = Observer()
    observer.schedule(DocsIndexEventHandler(), str(root), recursive=True)
    observer.start()
    try:
        while True:
            if scheduler.run_once_if_pending():
                continue
            time.sleep(_sleep_interval(debounce_seconds))
    except KeyboardInterrupt:
        return 0
    finally:
        observer.stop()
        observer.join()


def _run_fix_and_report(root: Path, config: DocLedgerConfig | None = None) -> int:
    result = reconcile_tree(root, config)
    changed = apply_updates(result)
    print(_status_line(f"doc-ledger watch updated {changed} file(s)"))
    if result.messages:
        print(_status_line(f"doc-ledger watch reconciliation messages: {len(result.messages)}"))
    return changed


def _sleep_interval(debounce_seconds: float) -> float:
    if debounce_seconds <= 0:
        return 0.1
    return min(debounce_seconds / 2, 0.25)


def _is_relevant_watch_event(event, config: DocLedgerConfig | None = None, root: Path | None = None) -> bool:
    config = config or default_config()
    root = root or _watch_root_from_event(event)
    paths = [Path(getattr(event, "src_path", ""))]
    dest_path = getattr(event, "dest_path", None)
    if dest_path:
        paths.append(Path(dest_path))

    is_directory = bool(getattr(event, "is_directory", False))
    return any(_is_relevant_watch_path(path, is_directory=is_directory, config=config, root=root) for path in paths)


def _is_relevant_watch_path(
    path: Path,
    is_directory: bool,
    config: DocLedgerConfig | None = None,
    root: Path | None = None,
) -> bool:
    config = config or default_config()
    if _is_ignored_watch_path(path, config):
        return False
    if is_directory:
        return True
    root = root or path.parent
    return _is_indexable_file(root.resolve(strict=False), path.resolve(strict=False), config)


def _watch_root_from_event(event) -> Path:
    path = Path(getattr(event, "src_path", ""))
    return path.parent


def _is_ignored_watch_path(path: Path, config: DocLedgerConfig) -> bool:
    if any(part in config.watch.ignored_dirs for part in path.parts):
        return True

    name = path.name
    if name.startswith(".#"):
        return True
    if any(name.endswith(suffix) for suffix in config.watch.ignored_suffixes):
        return True
    return False
