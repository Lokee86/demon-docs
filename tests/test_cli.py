from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger.readme_io import make_readme_template
from doc_ledger import cli


def test_fix_accepts_root(tmp_path: Path) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0
    assert (tmp_path / "!README.md").exists()


def test_fix_without_config_uses_defaults(tmp_path: Path) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0

    readme_text = (tmp_path / "!README.md").read_text(encoding="utf-8")
    assert "<!-- doc-ledger:files:start -->" in readme_text


def test_fix_loads_discovered_dot_config(tmp_path: Path, monkeypatch) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (tmp_path / ".doc-ledger.toml").write_text('[markers]\nprefix = "discovered-ledger"\n', encoding="utf-8")
    monkeypatch.chdir(tmp_path)

    assert cli.main(["fix", "--root", str(docs_root)]) == 0

    readme_text = (docs_root / "!README.md").read_text(encoding="utf-8")
    assert "<!-- discovered-ledger:files:start -->" in readme_text


def test_fix_explicit_config_wins_over_discovered_config(tmp_path: Path, monkeypatch) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (tmp_path / ".doc-ledger.toml").write_text('[markers]\nprefix = "discovered-ledger"\n', encoding="utf-8")
    explicit_config = tmp_path / "explicit.toml"
    explicit_config.write_text('[markers]\nprefix = "explicit-ledger"\n', encoding="utf-8")
    monkeypatch.chdir(tmp_path)

    assert cli.main(["fix", "--config", str(explicit_config), "--root", str(docs_root)]) == 0

    readme_text = (docs_root / "!README.md").read_text(encoding="utf-8")
    assert "<!-- explicit-ledger:files:start -->" in readme_text
    assert "discovered-ledger" not in readme_text


def test_fix_reports_summary(tmp_path: Path, capsys) -> None:
    readme_path = tmp_path / "!README.md"
    marker_block = "<!-- doc-ledger:files:start -->\n<!-- doc-ledger:files:end -->"
    readme_text = make_readme_template(tmp_path, tmp_path, None).replace(
        marker_block,
        "<!-- doc-ledger:files:start -->\n\n- [ghost.md](ghost.md) - Ghost documentation.\n\n<!-- doc-ledger:files:end -->",
    )
    readme_path.write_text(readme_text, encoding="utf-8")

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0

    output = capsys.readouterr().out
    assert "doc-ledger fix updated 1 file(s)" in output
    assert "doc-ledger fix reconciliation messages: 1" in output


def test_check_accepts_root(tmp_path: Path, capsys) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["fix", "--root", str(tmp_path)]) == 0
    capsys.readouterr()

    assert cli.main(["check", "--root", str(tmp_path)]) == 0
    output = capsys.readouterr().out
    assert output.strip() == "doc-ledger check passed"


def test_check_reports_failure_output(tmp_path: Path, capsys) -> None:
    tmp_path.mkdir(exist_ok=True)

    assert cli.main(["check", "--root", str(tmp_path)]) == 1
    output = capsys.readouterr().out
    assert "doc-ledger check failed" in output
    assert str(tmp_path / "!README.md") in output


def test_fix_rejects_missing_root(tmp_path: Path) -> None:
    missing_root = tmp_path / "missing"

    try:
        cli.main(["fix", "--root", str(missing_root)])
    except SystemExit as exc:
        assert exc.code == 2
    else:  # pragma: no cover - defensive
        raise AssertionError("expected fix to fail for missing root")


def test_check_rejects_missing_root(tmp_path: Path) -> None:
    missing_root = tmp_path / "missing"

    try:
        cli.main(["check", "--root", str(missing_root)])
    except SystemExit as exc:
        assert exc.code == 2
    else:  # pragma: no cover - defensive
        raise AssertionError("expected check to fail for missing root")


def test_watch_accepts_root_and_once(tmp_path: Path) -> None:
    assert cli.main(["watch", "--root", str(tmp_path), "--once"]) == 0


def test_watch_passes_loaded_config_to_watch_root(tmp_path: Path, monkeypatch) -> None:
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text("[watch]\ndebounce_seconds = 0.2\n", encoding="utf-8")
    calls: list[float] = []

    def fake_watch_root(root, config, once=False):
        calls.append(config.watch.debounce_seconds)
        assert root == tmp_path
        assert once is True
        return 0

    monkeypatch.setattr("doc_ledger.cli.watch_root", fake_watch_root)

    assert cli.main(["watch", "--config", str(config_path), "--root", str(tmp_path), "--once"]) == 0
    assert calls == [0.2]


def test_fix_uses_config_root_when_root_is_omitted(tmp_path: Path) -> None:
    config_root = tmp_path / "configured-docs"
    config_root.mkdir()
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(f'root = "{config_root}"\n', encoding="utf-8")

    assert cli.main(["fix", "--config", str(config_path)]) == 0
    assert (config_root / "!README.md").exists()


def test_fix_resolves_config_root_relative_to_config_file(tmp_path: Path, monkeypatch) -> None:
    project = tmp_path / "project"
    project.mkdir()
    notes = project / "notes"
    notes.mkdir()
    outside = tmp_path / "notes"
    outside.mkdir()
    (project / ".doc-ledger.toml").write_text('root = "notes"\n', encoding="utf-8")
    monkeypatch.chdir(tmp_path)

    assert cli.main(["fix", "--config", str(project / ".doc-ledger.toml")]) == 0
    assert (notes / "!README.md").exists()
    assert not (outside / "!README.md").exists()


def test_fix_root_argument_overrides_config_root(tmp_path: Path, monkeypatch) -> None:
    project = tmp_path / "project"
    project.mkdir()
    configured = project / "notes"
    configured.mkdir()
    override = tmp_path / "override-docs"
    override.mkdir()
    (project / ".doc-ledger.toml").write_text('root = "notes"\n', encoding="utf-8")
    monkeypatch.chdir(tmp_path)

    assert cli.main(["fix", "--config", str(project / ".doc-ledger.toml"), "--root", str(override)]) == 0
    assert (override / "!README.md").exists()
    assert not (configured / "!README.md").exists()


def test_fix_absolute_root_argument_remains_absolute(tmp_path: Path, monkeypatch) -> None:
    project = tmp_path / "project"
    project.mkdir()
    configured = project / "notes"
    configured.mkdir()
    absolute_root = tmp_path / "absolute-docs"
    absolute_root.mkdir()
    (project / ".doc-ledger.toml").write_text('root = "notes"\n', encoding="utf-8")
    monkeypatch.chdir(project)

    assert cli.main(["fix", "--root", str(absolute_root)]) == 0
    assert (absolute_root / "!README.md").exists()
    assert not (configured / "!README.md").exists()


def test_fix_without_config_defaults_to_cwd_docs(tmp_path: Path, monkeypatch) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    monkeypatch.chdir(tmp_path)

    assert cli.main(["fix"]) == 0
    assert (docs_root / "!README.md").exists()


def test_fix_prefers_root_argument_over_config_root(tmp_path: Path) -> None:
    config_root = tmp_path / "configured-docs"
    override_root = tmp_path / "override-docs"
    config_root.mkdir()
    override_root.mkdir()
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(f'root = "{config_root}"\n', encoding="utf-8")

    assert cli.main(["fix", "--config", str(config_path), "--root", str(override_root)]) == 0
    assert (override_root / "!README.md").exists()
    assert not (config_root / "!README.md").exists()


def test_fix_rejects_missing_config(tmp_path: Path) -> None:
    missing_config = tmp_path / "missing.toml"

    try:
        cli.main(["fix", "--config", str(missing_config)])
    except SystemExit as exc:
        assert exc.code == 2
    else:  # pragma: no cover - defensive
        raise AssertionError("expected fix to fail for missing config")


def test_main_py_runs_without_pythonpath_via_subprocess(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide.md").write_text("# Guide\n", encoding="utf-8")
    relative_root = os.path.relpath(docs_root, TOOL_ROOT)
    env = os.environ.copy()
    env.pop("PYTHONPATH", None)

    fix_result = subprocess.run(
        ["python3", "main.py", "fix", "--root", relative_root],
        cwd=TOOL_ROOT,
        env=env,
        check=False,
        capture_output=True,
        text=True,
    )
    assert fix_result.returncode == 0

    check_result = subprocess.run(
        ["python3", "main.py", "check", "--root", relative_root],
        cwd=TOOL_ROOT,
        env=env,
        check=False,
        capture_output=True,
        text=True,
    )
    assert check_result.returncode == 0


def test_docs_index_compatibility_shim_imports_doc_ledger_cli() -> None:
    from docs_index import cli as legacy_cli
    from doc_ledger import cli as modern_cli

    assert legacy_cli is modern_cli
