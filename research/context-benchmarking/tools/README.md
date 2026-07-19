# Context Benchmarking Tools

These scripts support future corpus preparation. They are research utilities, not Demon Docs product commands or release gates.

## Current Helpers

- `discover_candidates.py`: collect reviewable GitHub metadata for explicitly named repositories without ranking them by stars.
- `prepare_historical_fixture.py`: create a disposable pre-change source workspace from a pinned repository commit.
- `validate_fixture.py`: run and record a fixture's baseline verification command.

## Preserved Initial Scanner

`initial-scan/` contains the scripts used for the first broad Go-repository search:

- `scan_candidates.py`;
- `validate_candidates.py`;
- `prepare_benchmarks.py`; and
- `validate_benchmark_snapshots.py`.

Those scripts are retained for provenance and reproducibility. Their candidate search used star bands and a combined suitability score weighted toward documentation gaps. They must not be treated as the future quadrant-classification method.

Generated clone caches belong under `../clones/`, generated task source trees under `../fixtures/*/source/`, and temporary run workspaces under `../runs/` or `../workspaces/`; all are ignored by Git.
