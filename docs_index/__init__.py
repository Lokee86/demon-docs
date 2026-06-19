from __future__ import annotations

import importlib
import sys


_MODULE_ORDER = [
    "config",
    "model",
    "paths",
    "parent_index",
    "readme_io",
    "scan",
    "reconcile",
    "watch",
    "cli",
]


def _export_module(name: str) -> None:
    module = importlib.import_module(f"doc_ledger.{name}")
    sys.modules[f"{__name__}.{name}"] = module
    globals()[name] = module


for _module_name in _MODULE_ORDER:
    _export_module(_module_name)

