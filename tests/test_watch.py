from __future__ import annotations

import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.config import DocLedgerConfig
from doc_ledger.config import FileConfig
from doc_ledger.config import WatchConfig
from doc_ledger.watch import WatchScheduler
from doc_ledger.watch import _is_relevant_watch_event
from doc_ledger.watch import watch_root


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
    root = Path("/docs")

    assert _is_relevant_watch_event(FakeEvent("/docs/guide.md"), root=root) is True
    assert _is_relevant_watch_event(FakeEvent("/docs/stubs", is_directory=True, event_type="created"), root=root) is True
    assert _is_relevant_watch_event(FakeEvent("/docs/temp.txt", dest_path="/docs/moved.md", event_type="moved"), root=root) is True


def test_watch_event_filter_default_png_event_is_not_relevant() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/diagram.png"), root=Path("/docs")) is False


def test_watch_event_filter_configured_png_include_is_relevant() -> None:
    config = DocLedgerConfig(file=FileConfig(include_patterns=["**/*.png"]))

    assert _is_relevant_watch_event(FakeEvent("/docs/diagram.png"), config, root=Path("/docs")) is True


def test_watch_event_filter_configured_tmp_exclude_is_ignored() -> None:
    config = DocLedgerConfig(file=FileConfig(include_patterns=["**/*"], exclude_patterns=["**/*.tmp"]))

    assert _is_relevant_watch_event(FakeEvent("/docs/scratch.tmp"), config, root=Path("/docs")) is False


def test_watch_event_filter_draft_files_follow_include_exclude_rules() -> None:
    root = Path("/docs")
    config = DocLedgerConfig(file=FileConfig(include_patterns=["**/*.png"], exclude_patterns=["stubs/*.tmp"]))

    assert _is_relevant_watch_event(FakeEvent("/docs/stubs/diagram.png"), config, root=root) is True
    assert _is_relevant_watch_event(FakeEvent("/docs/stubs/scratch.tmp"), config, root=root) is False


def test_watch_event_filter_ignores_temp_and_hidden_paths() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/.git/config")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/.cache/guide.md")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/__pycache__/guide.md")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/guide.md~")) is False
    assert _is_relevant_watch_event(FakeEvent("/docs/.#guide.md")) is False


def test_watch_event_filter_ignores_default_git_directory() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/.git/config")) is False


def test_watch_event_filter_ignores_default_cache_directory() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/.cache/guide.md")) is False


def test_watch_event_filter_ignores_default_swp_suffix() -> None:
    assert _is_relevant_watch_event(FakeEvent("/docs/guide.md.swp")) is False


def test_watch_event_filter_honors_custom_ignored_directory() -> None:
    config = DocLedgerConfig(watch=WatchConfig(ignored_dirs=["build"], ignored_suffixes=[]))

    assert _is_relevant_watch_event(FakeEvent("/docs/build/guide.md"), config) is False


def test_watch_event_filter_honors_custom_ignored_suffix() -> None:
    config = DocLedgerConfig(watch=WatchConfig(ignored_dirs=[], ignored_suffixes=[".draft"]))

    assert _is_relevant_watch_event(FakeEvent("/docs/guide.md.draft"), config) is False


def test_watch_root_once_runs_fix_without_observer(tmp_path: Path, monkeypatch) -> None:
    called: list[Path] = []

    monkeypatch.setattr("doc_ledger.watch._run_fix_and_report", lambda root, config=None: called.append(root) or 0)

    assert watch_root(tmp_path, once=True) == 0
    assert called == [tmp_path]


def test_watch_root_once_reports_startup_and_summary(tmp_path: Path, monkeypatch, capsys) -> None:
    class Result:
        messages = ["updated"]

    monkeypatch.setattr("doc_ledger.watch.reconcile_tree", lambda root, config=None: Result())
    monkeypatch.setattr("doc_ledger.watch.apply_updates", lambda result: 3)

    assert watch_root(tmp_path, once=True) == 0

    output = capsys.readouterr().out
    assert f"doc-ledger watch watching {tmp_path}" in output
    assert "doc-ledger watch updated 3 file(s)" in output


def test_watch_root_runs_initial_fix_before_observer_start(tmp_path: Path, monkeypatch) -> None:
    calls: list[str] = []

    class FakeObserver:
        def schedule(self, *_args, **_kwargs) -> None:
            calls.append("schedule")

        def start(self) -> None:
            calls.append("start")

        def stop(self) -> None:
            calls.append("stop")

        def join(self) -> None:
            calls.append("join")

    monkeypatch.setattr("doc_ledger.watch._run_fix_and_report", lambda root, config=None: calls.append("fix") or 0)
    monkeypatch.setattr("watchdog.observers.Observer", lambda: FakeObserver())
    monkeypatch.setattr("doc_ledger.watch.time.sleep", lambda _seconds: (_ for _ in ()).throw(KeyboardInterrupt()))

    assert watch_root(tmp_path, once=False) == 0
    assert calls[:3] == ["fix", "schedule", "start"]
    assert calls[-2:] == ["stop", "join"]


def test_watch_root_uses_configured_debounce_seconds(tmp_path: Path, monkeypatch) -> None:
    sleep_calls: list[float] = []

    class FakeObserver:
        def schedule(self, *_args, **_kwargs) -> None:
            pass

        def start(self) -> None:
            pass

        def stop(self) -> None:
            pass

        def join(self) -> None:
            pass

    def fake_sleep(seconds: float) -> None:
        sleep_calls.append(seconds)
        raise KeyboardInterrupt

    config = DocLedgerConfig(watch=WatchConfig(debounce_seconds=0.4))
    monkeypatch.setattr("doc_ledger.watch._run_fix_and_report", lambda root, config=None: 0)
    monkeypatch.setattr("watchdog.observers.Observer", lambda: FakeObserver())
    monkeypatch.setattr("doc_ledger.watch.time.sleep", fake_sleep)

    assert watch_root(tmp_path, config=config, once=False) == 0
    assert sleep_calls == [0.2]


def test_watch_root_explicit_debounce_seconds_wins(tmp_path: Path, monkeypatch) -> None:
    sleep_calls: list[float] = []

    class FakeObserver:
        def schedule(self, *_args, **_kwargs) -> None:
            pass

        def start(self) -> None:
            pass

        def stop(self) -> None:
            pass

        def join(self) -> None:
            pass

    def fake_sleep(seconds: float) -> None:
        sleep_calls.append(seconds)
        raise KeyboardInterrupt

    config = DocLedgerConfig(watch=WatchConfig(debounce_seconds=0.4))
    monkeypatch.setattr("doc_ledger.watch._run_fix_and_report", lambda root, config=None: 0)
    monkeypatch.setattr("watchdog.observers.Observer", lambda: FakeObserver())
    monkeypatch.setattr("doc_ledger.watch.time.sleep", fake_sleep)

    assert watch_root(tmp_path, config=config, debounce_seconds=0.1, once=False) == 0
    assert sleep_calls == [0.05]


def test_watch_root_once_still_runs_one_fix_and_exits(tmp_path: Path, monkeypatch) -> None:
    calls: list[Path] = []
    config = DocLedgerConfig(watch=WatchConfig(debounce_seconds=0.2))

    monkeypatch.setattr("doc_ledger.watch._run_fix_and_report", lambda root, config=None: calls.append(root) or 0)

    assert watch_root(tmp_path, config=config, once=True) == 0
    assert calls == [tmp_path]
