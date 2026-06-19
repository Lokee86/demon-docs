from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from docs_index.watch import WatchScheduler
from docs_index.watch import _is_relevant_watch_event
from docs_index.watch import watch_root


class FakeClock:
    def __init__(self, value: float = 0.0) -> None:
        self.value = value

    def __call__(self) -> float:
        return self.value

    def advance(self, amount: float) -> None:
        self.value += amount


class FakeEvent:
    def __init__(
        self,
        src_path: str,
        *,
        is_directory: bool = False,
        event_type: str = "modified",
        dest_path: str | None = None,
    ) -> None:
        self.src_path = src_path
        self.is_directory = is_directory
        self.event_type = event_type
        self.dest_path = dest_path


def test_watch_scheduler_runs_once_for_single_event() -> None:
    clock = FakeClock()
    runs = 0

    def run_fix() -> None:
        nonlocal runs
        runs += 1

    scheduler = WatchScheduler(run_fix, debounce_seconds=0.5, clock=clock)

    scheduler.mark_changed()
    clock.advance(0.5)

    assert scheduler.run_once_if_pending() is True
    assert runs == 1
    assert scheduler.run_once_if_pending() is False


def test_watch_scheduler_debounces_multiple_events() -> None:
    clock = FakeClock()
    runs = 0

    def run_fix() -> None:
        nonlocal runs
        runs += 1

    scheduler = WatchScheduler(run_fix, debounce_seconds=0.5, clock=clock)

    scheduler.mark_changed()
    clock.advance(0.1)
    scheduler.mark_changed()
    clock.advance(0.1)
    scheduler.mark_changed()

    assert scheduler.run_once_if_pending() is False
    clock.advance(0.5)

    assert scheduler.run_once_if_pending() is True
    assert runs == 1


def test_watch_scheduler_runs_again_after_change_during_fix() -> None:
    clock = FakeClock()
    scheduler: WatchScheduler | None = None
    runs: list[int] = []

    def run_fix() -> None:
        runs.append(len(runs))
        if len(runs) == 1:
            assert scheduler is not None
            scheduler.mark_changed()

    scheduler = WatchScheduler(run_fix, debounce_seconds=0.0, clock=clock)

    scheduler.mark_changed()

    assert scheduler.run_once_if_pending() is True
    assert runs == [0]
    assert scheduler.run_once_if_pending() is True
    assert runs == [0, 1]
    assert scheduler.run_once_if_pending() is False


def test_watch_scheduler_coalesces_multiple_changes_during_fix() -> None:
    clock = FakeClock()
    scheduler: WatchScheduler | None = None
    runs = 0

    def run_fix() -> None:
        nonlocal runs
        runs += 1
        if runs == 1:
            assert scheduler is not None
            scheduler.mark_changed()
            scheduler.mark_changed()

    scheduler = WatchScheduler(run_fix, debounce_seconds=0.0, clock=clock)

    scheduler.mark_changed()

    assert scheduler.run_once_if_pending() is True
    assert runs == 1
    assert scheduler.run_once_if_pending() is True
    assert runs == 2
    assert scheduler.run_once_if_pending() is False


def test_watch_event_filter_accepts_relevant_paths() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/guide.md")) is True
    assert _is_relevant_watch_event(FakeEvent("/docs/stubs", is_directory=True, event_type="created")) is True
    assert _is_relevant_watch_event(FakeEvent("/docs/temp.txt", dest_path="/docs/moved.md", event_type="moved")) is True


def test_watch_event_filter_ignores_temp_and_hidden_paths() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/.git/config")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/.cache/guide.md")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/__pycache__/guide.md")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/guide.md~")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/.#guide.md")) is False


def test_watch_root_once_runs_fix_without_observer(tmp_path: Path, monkeypatch) -> None:
    called: list[Path] = []

    monkeypatch.setattr("docs_index.watch._run_fix_and_report", lambda root: called.append(root) or 0)

    assert watch_root(tmp_path, once=True) == 0
    assert called == [tmp_path]


def test_watch_root_once_reports_startup_and_summary(tmp_path: Path, monkeypatch, capsys) -> None:
    class Result:
        messages = ["updated"]

    monkeypatch.setattr("docs_index.watch.reconcile_tree", lambda root: Result())
    monkeypatch.setattr("docs_index.watch.apply_updates", lambda result: 3)

    assert watch_root(tmp_path, once=True) == 0

    output = capsys.readouterr().out
    assert f"docs-index watch watching {tmp_path}" in output
    assert "docs-index watch updated 3 file(s)" in output
