from __future__ import annotations

from collections.abc import Callable
from pathlib import Path
import threading
import time

from docs_index.reconcile import apply_updates
from docs_index.reconcile import reconcile_tree

IGNORED_DIRECTORIES = {".git", ".cache", "__pycache__"}
IGNORED_SUFFIXES = ("~", ".swp", ".tmp", ".bak")


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


def watch_root(root: Path, debounce_seconds: float = 0.75, once: bool = False) -> int:
    print(f"docs-index watch watching {root}")
    if once:
        _run_fix_and_report(root)
        return 0

    from watchdog.events import FileSystemEventHandler
    from watchdog.observers import Observer

    scheduler = WatchScheduler(lambda: _run_fix_and_report(root), debounce_seconds=debounce_seconds)

    class DocsIndexEventHandler(FileSystemEventHandler):
        def on_any_event(self, event) -> None:  # type: ignore[override]
            if _is_relevant_watch_event(event):
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


def _run_fix_and_report(root: Path) -> int:
    result = reconcile_tree(root)
    changed = apply_updates(result)
    print(f"docs-index watch updated {changed} file(s)")
    if result.messages:
        print(f"docs-index watch reconciliation messages: {len(result.messages)}")
    return changed


def _sleep_interval(debounce_seconds: float) -> float:
    if debounce_seconds <= 0:
        return 0.1
    return min(debounce_seconds / 2, 0.25)


def _is_relevant_watch_event(event) -> bool:
    paths = [Path(getattr(event, "src_path", ""))]
    dest_path = getattr(event, "dest_path", None)
    if dest_path:
        paths.append(Path(dest_path))

    is_directory = bool(getattr(event, "is_directory", False))
    return any(_is_relevant_watch_path(path, is_directory=is_directory) for path in paths)


def _is_relevant_watch_path(path: Path, is_directory: bool) -> bool:
    if _is_ignored_watch_path(path):
        return False
    if is_directory:
        return True
    return path.suffix == ".md"


def _is_ignored_watch_path(path: Path) -> bool:
    if any(part in IGNORED_DIRECTORIES for part in path.parts):
        return True

    name = path.name
    if name.startswith(".#"):
        return True
    if any(name.endswith(suffix) for suffix in IGNORED_SUFFIXES):
        return True
    return False
