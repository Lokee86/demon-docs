from __future__ import annotations

import tomllib
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(PROJECT_ROOT))

from doc_ledger import __version__


PYPROJECT_PATH = PROJECT_ROOT / "pyproject.toml"


def _load_pyproject() -> dict:
    return tomllib.loads(PYPROJECT_PATH.read_text(encoding="utf-8"))


def test_pyproject_toml_exists() -> None:
    assert PYPROJECT_PATH.exists()


def test_pyproject_declares_project_name() -> None:
    pyproject = _load_pyproject()

    assert pyproject["project"]["name"] == "doc-ledger"


def test_pyproject_declares_console_script() -> None:
    pyproject = _load_pyproject()

    assert pyproject["project"]["scripts"]["doc-ledger"] == "doc_ledger.cli:main"


def test_pyproject_declares_license_metadata() -> None:
    pyproject = _load_pyproject()

    assert pyproject["project"]["license"] == "MIT"
    assert pyproject["project"]["license-files"] == ["LICENSE"]


def test_pyproject_version_matches_package_version() -> None:
    pyproject = _load_pyproject()

    assert pyproject["project"]["version"] == __version__


def test_pyproject_declares_runtime_dependencies() -> None:
    pyproject = _load_pyproject()

    dependencies = pyproject["project"]["dependencies"]

    assert "tomlkit" in dependencies
    assert "watchdog" in dependencies
