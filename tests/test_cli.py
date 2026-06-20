from __future__ import annotations

import sys
from pathlib import Path
from types import SimpleNamespace
import importlib

import pytest


TOOL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(TOOL_ROOT))

from doc_ledger import cli


def _capture_help_output(args: list[str], capsys: pytest.CaptureFixture[str]) -> str:
    with pytest.raises(SystemExit) as exc:
        cli.main(args)

    assert exc.value.code == 0
    return capsys.readouterr().out


def test_main_help_shows_top_level_description_and_examples(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["--help"], capsys)

    assert "doc-ledger reconciles local index files with a file tree." in output
    assert "fix" in output
    assert "check" in output
    assert "watch" in output
    assert "config" in output
    assert "doc-ledger fix" in output
    assert "doc-ledger check" in output
    assert "doc-ledger watch" in output
    assert "doc-ledger config paths" in output
    assert "doc-ledger config show" in output
    assert "doc-ledger fix --root docs" in output
    assert "doc-ledger check --root docs" in output
    assert "doc-ledger watch --root docs" in output
    assert "doc-ledger fix --config .doc-ledger.toml" in output
    assert "doc-ledger --version" in output


@pytest.mark.parametrize("flag", [["--version"], ["-v"]])
def test_main_version_flags_exit_zero_and_print_version(flag: list[str], capsys: pytest.CaptureFixture[str]) -> None:
    with pytest.raises(SystemExit) as exc:
        cli.main(flag)

    assert exc.value.code == 0
    output = capsys.readouterr().out
    assert output.strip() == "doc-ledger 0.1.1"


def test_fix_help_includes_flag_help(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["fix", "--help"], capsys)

    assert "Reconcile the docs tree and write any needed updates." in output
    assert "--root PATH" in output
    assert "--config PATH" in output
    assert "--no-local-config" in output
    assert "--no-global-config" in output
    assert "--index-file NAME" in output
    assert "--draft-folder NAME" in output
    assert "--draft-description-prefix TEXT" in output
    assert "--include PATTERN" in output
    assert "--exclude PATTERN" in output
    assert "--marker-prefix TEXT" in output
    assert "--parent-label TEXT" in output
    assert "--parent-link-folder-indexes" in output
    assert "--no-parent-link-folder-indexes" in output
    assert "--parent-link-indexed-files" in output
    assert "--no-parent-link-indexed-files" in output
    assert "1. --config PATH" in output
    assert "2. ./.doc-ledger.toml" in output
    assert "3. ./doc-ledger.toml" in output
    assert "4. global user config" in output
    assert "5. built-in defaults" in output
    assert "local config is current-directory only" in output
    assert "there is no upward parent-directory search" in output
    assert "local and global configs are not merged" in output
    assert "CLI flags override the selected config" in output
    assert "docs root directory to reconcile" in output
    assert "explicit doc-ledger config file" in output


def test_check_help_includes_flag_help(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["check", "--help"], capsys)

    assert "Verify that the docs tree is already reconciled." in output
    assert "--root PATH" in output
    assert "--config PATH" in output
    assert "--no-local-config" in output
    assert "--no-global-config" in output
    assert "--index-file NAME" in output
    assert "--draft-folder NAME" in output
    assert "--draft-description-prefix TEXT" in output
    assert "--include PATTERN" in output
    assert "--exclude PATTERN" in output
    assert "--marker-prefix TEXT" in output
    assert "--parent-label TEXT" in output
    assert "--parent-link-folder-indexes" in output
    assert "--no-parent-link-folder-indexes" in output
    assert "--parent-link-indexed-files" in output
    assert "--no-parent-link-indexed-files" in output
    assert "1. --config PATH" in output
    assert "2. ./.doc-ledger.toml" in output
    assert "3. ./doc-ledger.toml" in output
    assert "4. global user config" in output
    assert "5. built-in defaults" in output
    assert "local config is current-directory only" in output
    assert "there is no upward parent-directory search" in output
    assert "local and global configs are not merged" in output
    assert "CLI flags override the selected config" in output


def test_watch_help_includes_once_flag_help(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["watch", "--help"], capsys)

    assert "Watch runs in the foreground by default, runs one reconciliation immediately, and then watches for relevant filesystem changes." in output
    assert "--root PATH" in output
    assert "--config PATH" in output
    assert "--no-local-config" in output
    assert "--no-global-config" in output
    assert "--index-file NAME" in output
    assert "--draft-folder NAME" in output
    assert "--draft-description-prefix TEXT" in output
    assert "--include PATTERN" in output
    assert "--exclude PATTERN" in output
    assert "--marker-prefix TEXT" in output
    assert "--parent-label TEXT" in output
    assert "--parent-link-folder-indexes" in output
    assert "--no-parent-link-folder-indexes" in output
    assert "--parent-link-indexed-files" in output
    assert "--no-parent-link-indexed-files" in output
    assert "--once" in output
    assert "--debounce-seconds FLOAT" in output
    assert "run one reconciliation pass and exit" in output
    assert "1. --config PATH" in output
    assert "2. ./.doc-ledger.toml" in output
    assert "3. ./doc-ledger.toml" in output
    assert "4. global user config" in output
    assert "5. built-in defaults" in output
    assert "local config is current-directory only" in output
    assert "there is no upward parent-directory search" in output
    assert "local and global configs are not merged" in output
    assert "CLI flags override the selected config" in output


def test_config_help_includes_subcommands(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["config", "--help"], capsys)

    assert "paths" in output
    assert "show" in output
    assert "init" in output
    assert "Local config lookup is current-directory only." in output
    assert "There is no upward parent-directory search." in output
    assert "Local and global configs are not merged." in output
    assert "CLI flags override the selected config." in output


def test_config_paths_help_explains_config_paths(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["config", "paths", "--help"], capsys)

    assert "Print the current-directory local config, global user config, and selected config paths." in output
    assert "current-directory local config candidates" in output
    assert "global user config path" in output
    assert "selected config path" in output


def test_config_show_help_explains_resolved_config_output(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["config", "show", "--help"], capsys)

    assert "resolved selected config" in output


def test_config_init_help_includes_flags(capsys: pytest.CaptureFixture[str]) -> None:
    output = _capture_help_output(["config", "init", "--help"], capsys)

    assert "--local" in output
    assert "--global" in output
    assert "--force" in output
    assert "global user config location" in output


def test_python_module_entry_point_import_has_no_cli_side_effects(monkeypatch: pytest.MonkeyPatch) -> None:
    import doc_ledger.cli as cli_module

    called = {"value": False}

    def fake_main() -> int:
        called["value"] = True
        return 0

    monkeypatch.setattr(cli_module, "main", fake_main)
    module = importlib.import_module("doc_ledger.__main__")

    assert module.main is fake_main
    assert called["value"] is False


def test_fix_parent_link_flags_override_loaded_config(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (docs_root / "stubs").mkdir()
    (docs_root / "stubs" / "draft.md").write_text("# Draft\n", encoding="utf-8")
    (docs_root / "guide").mkdir()
    (docs_root / "guide" / "topic.md").write_text("# Topic\n", encoding="utf-8")

    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = false
indexed_files = false
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(
        [
            "fix",
            "--config",
            str(config_path),
            "--root",
            str(docs_root),
            "--parent-link-folder-indexes",
            "--parent-link-indexed-files",
        ]
    ) == 0

    assert "Parent index: [Docs](./README.md)" in (docs_root / "page.md").read_text(encoding="utf-8")
    assert "Parent index: [Docs](../README.md)" in (docs_root / "stubs" / "draft.md").read_text(encoding="utf-8")
    assert "Parent index: [Docs](../README.md)" in (docs_root / "guide" / "README.md").read_text(encoding="utf-8")


def test_check_parent_link_flags_override_loaded_config(tmp_path: Path, monkeypatch) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = true
indexed_files = true
""".strip()
        + "\n",
        encoding="utf-8",
    )
    seen: dict[str, bool] = {}

    def fake_reconcile_tree(root, config):
        seen["folder_indexes"] = config.parent_link.folder_indexes
        seen["indexed_files"] = config.parent_link.indexed_files
        return SimpleNamespace(updates=[], messages=[])

    monkeypatch.setattr("doc_ledger.cli.reconcile_tree", fake_reconcile_tree)

    assert cli.main(
        [
            "check",
            "--config",
            str(config_path),
            "--root",
            str(docs_root),
            "--no-parent-link-folder-indexes",
            "--no-parent-link-indexed-files",
        ]
    ) == 0

    assert seen == {"folder_indexes": False, "indexed_files": False}


def test_watch_parent_link_flags_override_loaded_config(tmp_path: Path, monkeypatch) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = true
indexed_files = true
""".strip()
        + "\n",
        encoding="utf-8",
    )
    seen: dict[str, bool] = {}

    def fake_watch_root(root, config, debounce_seconds=None, once=False):
        seen["folder_indexes"] = config.parent_link.folder_indexes
        seen["indexed_files"] = config.parent_link.indexed_files
        seen["debounce_seconds"] = debounce_seconds
        seen["once"] = once
        return 0

    monkeypatch.setattr("doc_ledger.cli.watch_root", fake_watch_root)

    assert cli.main(
        [
            "watch",
            "--config",
            str(config_path),
            "--root",
            str(docs_root),
            "--no-parent-link-folder-indexes",
            "--no-parent-link-indexed-files",
            "--once",
        ]
    ) == 0

    assert seen == {"folder_indexes": False, "indexed_files": False, "debounce_seconds": None, "once": True}


def test_fix_parent_link_indexed_files_flag_enables_file_links(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    stubs = docs_root / "stubs"
    stubs.mkdir()
    (stubs / "draft.md").write_text("# Draft\n", encoding="utf-8")

    assert cli.main(
        [
            "fix",
            "--root",
            str(docs_root),
            "--parent-link-indexed-files",
        ]
    ) == 0

    assert "Parent index: [Docs](./README.md)" in (docs_root / "page.md").read_text(encoding="utf-8")
    assert "Parent index: [Docs](../README.md)" in (stubs / "draft.md").read_text(encoding="utf-8")


def test_fix_no_parent_link_indexed_files_flag_disables_file_links(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    stubs = docs_root / "stubs"
    stubs.mkdir()
    (stubs / "draft.md").write_text("# Draft\n", encoding="utf-8")
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"
index_file = "README.md"

[parent_link]
indexed_files = true
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(
        [
            "fix",
            "--config",
            str(config_path),
            "--root",
            str(docs_root),
            "--no-parent-link-indexed-files",
        ]
    ) == 0

    assert "Parent index:" not in (docs_root / "page.md").read_text(encoding="utf-8")
    assert "Parent index:" not in (stubs / "draft.md").read_text(encoding="utf-8")


def test_fix_no_parent_link_folder_indexes_flag_disables_child_readme_links(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide").mkdir()

    assert cli.main(
        [
            "fix",
            "--root",
            str(docs_root),
            "--no-parent-link-folder-indexes",
        ]
    ) == 0

    assert "Parent index:" not in (docs_root / "guide" / "README.md").read_text(encoding="utf-8")


def test_fix_parent_link_folder_indexes_flag_enables_child_readme_links(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "guide").mkdir()
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"
index_file = "README.md"

[parent_link]
folder_indexes = false
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(
        [
            "fix",
            "--config",
            str(config_path),
            "--root",
            str(docs_root),
            "--parent-link-folder-indexes",
        ]
    ) == 0

    assert "Parent index: [Docs](../README.md)" in (docs_root / "guide" / "README.md").read_text(encoding="utf-8")


def test_fix_uses_current_directory_local_config(tmp_path: Path, monkeypatch) -> None:
    monkeypatch.chdir(tmp_path)
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (tmp_path / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix"]) == 0

    assert (docs_root / "!README.md").exists()
    assert not (docs_root / "README.md").exists()


def test_fix_ignores_parent_directory_config(tmp_path: Path, monkeypatch) -> None:
    parent = tmp_path / "parent"
    child = parent / "child"
    child.mkdir(parents=True)
    monkeypatch.chdir(child)
    docs_root = child / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (parent / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix"]) == 0

    assert (docs_root / "README.md").exists()
    assert not (docs_root / "!README.md").exists()


def test_fix_uses_global_config_when_cwd_has_no_local_config(tmp_path: Path, monkeypatch) -> None:
    cwd = tmp_path / "cwd"
    cwd.mkdir()
    monkeypatch.chdir(cwd)
    docs_root = cwd / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    xdg = tmp_path / "xdg"
    config_path = xdg / "doc-ledger" / "config.toml"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(
        f"""
root = "{docs_root}"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["fix"]) == 0

    assert (docs_root / "!README.md").exists()
    assert not (docs_root / "README.md").exists()


def test_fix_uses_built_in_defaults_when_no_local_or_global_config_exists(tmp_path: Path, monkeypatch) -> None:
    cwd = tmp_path / "cwd"
    cwd.mkdir()
    monkeypatch.chdir(cwd)
    monkeypatch.delenv("XDG_CONFIG_HOME", raising=False)
    docs_root = cwd / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")

    assert cli.main(["fix"]) == 0

    assert (docs_root / "README.md").exists()
    assert not (docs_root / "!README.md").exists()


def test_fix_config_path_bypasses_local_and_global_config(tmp_path: Path, monkeypatch) -> None:
    cwd = tmp_path / "cwd"
    cwd.mkdir()
    monkeypatch.chdir(cwd)
    docs_root = cwd / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (cwd / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    xdg = tmp_path / "xdg"
    global_config = xdg / "doc-ledger" / "config.toml"
    global_config.parent.mkdir(parents=True)
    global_config.write_text(
        f"""
root = "{docs_root}"
index_file = "README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))
    explicit = tmp_path / "explicit.toml"
    explicit.write_text(
        f"""
root = "{docs_root}"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(explicit)]) == 0

    assert (docs_root / "!README.md").exists()
    assert not (docs_root / "README.md").exists()


def test_fix_no_local_config_skips_cwd_config_and_falls_back_to_global(tmp_path: Path, monkeypatch) -> None:
    cwd = tmp_path / "cwd"
    cwd.mkdir()
    monkeypatch.chdir(cwd)
    (cwd / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    xdg = tmp_path / "xdg"
    global_config = xdg / "doc-ledger" / "config.toml"
    global_config.parent.mkdir(parents=True)
    docs_root = cwd / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    global_config.write_text(
        f"""
root = "{docs_root}"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["fix", "--no-local-config"]) == 0

    assert (docs_root / "!README.md").exists()
    assert not (docs_root / "README.md").exists()


def test_fix_no_global_config_uses_defaults_when_no_cwd_config_exists(tmp_path: Path, monkeypatch) -> None:
    cwd = tmp_path / "cwd"
    cwd.mkdir()
    monkeypatch.chdir(cwd)
    xdg = tmp_path / "xdg"
    global_config = xdg / "doc-ledger" / "config.toml"
    global_config.parent.mkdir(parents=True)
    docs_root = cwd / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    global_config.write_text(
        f"""
root = "{docs_root}"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["fix", "--no-global-config"]) == 0

    assert (docs_root / "README.md").exists()
    assert not (docs_root / "!README.md").exists()


def test_fix_root_override_still_applies_after_config_selection(tmp_path: Path, monkeypatch) -> None:
    cwd = tmp_path / "cwd"
    cwd.mkdir()
    monkeypatch.chdir(cwd)
    docs_root = cwd / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (cwd / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    override_root = tmp_path / "override"
    override_root.mkdir()
    (override_root / "page.md").write_text("# Override\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(override_root)]) == 0

    assert (override_root / "!README.md").exists()
    assert not (docs_root / "!README.md").exists()


def test_config_paths_exits_zero_and_shows_path_candidates(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    (tmp_path / ".doc-ledger.toml").write_text("", encoding="utf-8")
    (tmp_path / "doc-ledger.toml").write_text("", encoding="utf-8")
    xdg = tmp_path / "xdg"
    (xdg / "doc-ledger").mkdir(parents=True)
    (xdg / "doc-ledger" / "config.toml").write_text("", encoding="utf-8")
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["config", "paths"]) == 0

    output = capsys.readouterr().out
    assert f"cwd = {tmp_path}" in output
    assert f"local dot config = {tmp_path / '.doc-ledger.toml'} exists=True" in output
    assert f"local plain config = {tmp_path / 'doc-ledger.toml'} exists=True" in output
    assert f"selected local config = {tmp_path / '.doc-ledger.toml'}" in output
    assert f"global config = {xdg / 'doc-ledger' / 'config.toml'} exists=True" in output
    assert f"selected config = {tmp_path / '.doc-ledger.toml'}" in output


def test_config_show_defaults_when_no_config_exists(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)

    assert cli.main(["config", "show"]) == 0

    output = capsys.readouterr().out
    assert "selected_config_path = <built-in defaults>" in output
    assert "root = 'docs'" in output
    assert "index_file = 'README.md'" in output


def test_config_show_uses_cwd_local_config(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    (tmp_path / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["config", "show"]) == 0

    output = capsys.readouterr().out
    assert f"selected_config_path = {tmp_path / '.doc-ledger.toml'}" in output
    assert "index_file = '!README.md'" in output


def test_config_show_uses_global_config_when_no_local_exists(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    xdg = tmp_path / "xdg"
    config_path = xdg / "doc-ledger" / "config.toml"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["config", "show"]) == 0

    output = capsys.readouterr().out
    assert f"selected_config_path = {config_path}" in output
    assert "index_file = '!README.md'" in output


def test_config_show_honors_config_argument(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    config_path = tmp_path / "explicit.toml"
    config_path.write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["config", "show", "--config", str(config_path)]) == 0

    output = capsys.readouterr().out
    assert f"selected_config_path = {config_path}" in output
    assert "index_file = '!README.md'" in output


def test_config_show_honors_no_local_config(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    (tmp_path / ".doc-ledger.toml").write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    xdg = tmp_path / "xdg"
    config_path = xdg / "doc-ledger" / "config.toml"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(
        """
root = "docs"
index_file = "README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["config", "show", "--no-local-config"]) == 0

    output = capsys.readouterr().out
    assert f"selected_config_path = {config_path}" in output
    assert "index_file = 'README.md'" in output


def test_config_show_honors_no_global_config(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    xdg = tmp_path / "xdg"
    config_path = xdg / "doc-ledger" / "config.toml"
    config_path.parent.mkdir(parents=True)
    config_path.write_text(
        """
root = "docs"
index_file = "!README.md"
""".strip()
        + "\n",
        encoding="utf-8",
    )
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["config", "show", "--no-global-config"]) == 0

    output = capsys.readouterr().out
    assert "selected_config_path = <built-in defaults>" in output
    assert "index_file = 'README.md'" in output


def test_config_init_local_creates_dot_config(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)

    assert cli.main(["config", "init", "--local"]) == 0

    written = tmp_path / ".doc-ledger.toml"
    output = capsys.readouterr().out
    assert written.exists()
    assert str(written) in output
    assert 'root = "docs"' in written.read_text(encoding="utf-8")
    assert 'index_file = "README.md"' in written.read_text(encoding="utf-8")


def test_config_init_global_creates_global_config_under_xdg(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    xdg = tmp_path / "xdg"
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["config", "init", "--global"]) == 0

    written = xdg / "doc-ledger" / "config.toml"
    output = capsys.readouterr().out
    assert written.exists()
    assert str(written) in output
    assert (xdg / "doc-ledger").is_dir()


def test_config_init_does_not_overwrite_without_force(tmp_path: Path, monkeypatch) -> None:
    monkeypatch.chdir(tmp_path)
    written = tmp_path / ".doc-ledger.toml"
    written.write_text('root = "existing"\n', encoding="utf-8")

    with pytest.raises(SystemExit) as exc:
        cli.main(["config", "init", "--local"])

    assert exc.value.code == 2
    assert written.read_text(encoding="utf-8") == 'root = "existing"\n'


def test_config_init_overwrites_with_force(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    written = tmp_path / ".doc-ledger.toml"
    written.write_text('root = "existing"\n', encoding="utf-8")

    assert cli.main(["config", "init", "--local", "--force"]) == 0

    output = capsys.readouterr().out
    assert str(written) in output
    assert 'root = "docs"' in written.read_text(encoding="utf-8")


def test_config_init_global_creates_parent_directories(tmp_path: Path, monkeypatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.chdir(tmp_path)
    xdg = tmp_path / "xdg"
    monkeypatch.setenv("XDG_CONFIG_HOME", str(xdg))

    assert cli.main(["config", "init", "--global"]) == 0

    assert (xdg / "doc-ledger").is_dir()
    assert (xdg / "doc-ledger" / "config.toml").exists()
    assert str(xdg / "doc-ledger" / "config.toml") in capsys.readouterr().out


def test_fix_index_file_flag_changes_generated_index_filename(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root), "--index-file", "!README.md"]) == 0

    assert (docs_root / "!README.md").exists()
    assert not (docs_root / "README.md").exists()


def test_fix_draft_folder_flag_changes_stub_folder_behavior(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    drafts = docs_root / "_drafts"
    drafts.mkdir()
    (drafts / "draft.md").write_text("# Draft\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root), "--draft-folder", "_drafts"]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    assert "- [draft.md](_drafts/draft.md) - Stub: Draft documentation." in root_readme
    assert not (drafts / "README.md").exists()


def test_fix_include_flag_indexes_non_default_file_type(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (docs_root / "diagram.png").write_text("png body\n", encoding="utf-8")
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"

[files]
include_patterns = ["**/*.md"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root), "--include", "**/*.png"]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    assert "- [diagram.png](diagram.png) - Diagram documentation." in root_readme
    assert "page.md" not in root_readme


def test_fix_exclude_flag_excludes_matching_file(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")
    (docs_root / "diagram.png").write_text("png body\n", encoding="utf-8")
    config_path = tmp_path / "doc-ledger.toml"
    config_path.write_text(
        """
root = "docs"

[files]
include_patterns = ["**/*.md", "**/*.png"]
exclude_patterns = ["**/*.png"]
""".strip()
        + "\n",
        encoding="utf-8",
    )

    assert cli.main(["fix", "--config", str(config_path), "--root", str(docs_root), "--exclude", "**/*.md"]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    assert "page.md" not in root_readme
    assert "- [diagram.png](diagram.png) - Diagram documentation." in root_readme


def test_fix_marker_prefix_flag_changes_managed_markers(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")

    assert cli.main(["fix", "--root", str(docs_root), "--marker-prefix", "nav-ledger"]) == 0

    root_readme = (docs_root / "README.md").read_text(encoding="utf-8")
    assert "<!-- nav-ledger:files:start -->" in root_readme
    assert "<!-- nav-ledger:stubs:start -->" in root_readme
    assert "<!-- nav-ledger:folders:start -->" in root_readme


def test_fix_parent_label_flag_changes_enabled_parent_links(tmp_path: Path) -> None:
    docs_root = tmp_path / "docs"
    docs_root.mkdir()
    (docs_root / "page.md").write_text("# Page\n", encoding="utf-8")

    assert cli.main(
        [
            "fix",
            "--root",
            str(docs_root),
            "--parent-link-indexed-files",
            "--parent-label",
            "Back to Index",
        ]
    ) == 0

    assert "Back to Index: [Docs](./README.md)" in (docs_root / "page.md").read_text(encoding="utf-8")
