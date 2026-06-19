# Testing and Fixtures

doc-ledger is covered by focused pytest tests plus a few end-to-end flows that exercise real docs trees.

## Test Command

For an install smoke check, run:

```bash
python3 -m pip install -e ".[dev]"
doc-ledger --help
doc-ledger --version
```

From the repo root, run:

```bash
python3 -m pytest tests
```

## What the Tests Cover

The doc-ledger tests are split across small, focused areas:

- CLI behavior
- config loading and selection
- scan model construction
- README IO and managed section handling
- parent index behavior
- reconciliation planning and file updates
- watcher scheduling and filtering
- end-to-end flows
- public config examples

Those tests keep the implementation honest without depending on a larger application runtime.

## Dummy Docs Fixture Generator

`docs/make-dummy-docs.sh` is a manual stress and fixture generator. It builds a nested docs tree with randomly shaped folders and files so you can exercise reconciliation against a larger input.

Run it with:

```bash
./docs/make-dummy-docs.sh
```

Useful knobs exposed by the script:

- `ROOT_DIR` sets the output directory name
- `RECREATE=1` deletes the output directory before generating
- `RECREATE=0` adds into an existing tree
- `EXTENSIONS` controls the generated file extensions
- `MIN_AREAS` and `MAX_AREAS` control top-level area folders
- `MIN_SUBAREAS_PER_AREA` and `MAX_SUBAREAS_PER_AREA` control subarea nesting
- `MIN_TOPICS_PER_SUBAREA` and `MAX_TOPICS_PER_SUBAREA` control topic nesting
- `MIN_ROOT_FILES` and `MAX_ROOT_FILES` control files at the root
- `MIN_AREA_FILES` and `MAX_AREA_FILES` control files in area folders
- `MIN_SUBAREA_FILES` and `MAX_SUBAREA_FILES` control files in subarea folders
- `MIN_TOPIC_FILES` and `MAX_TOPIC_FILES` control files in topic folders

The default output directory is `dummy-docs/` in the current working directory.
The repo’s `.gitignore` already ignores that output directory.

## Manual Smoke Workflow

A simple end-to-end smoke test prefers the packaged CLI:

```bash
./docs/make-dummy-docs.sh
doc-ledger fix --root dummy-docs
doc-ledger check --root dummy-docs
```

If you are working from the repo checkout, `python3 main.py` can still be used as a development fallback. After that, try a move or rename inside `dummy-docs/`, run `fix` and `check` again, and inspect the diff. That is a good way to verify description preservation, stale entry removal, and parent index updates on a realistic tree.

## Fixture Guidance

- Dummy docs are manual stress fixtures, not canonical docs.
- Keep them disposable.
- Use them when you want a noisy tree that exercises nesting, stubs, and cross-folder reconciliation.

## Related Files

- `docs/make-dummy-docs.sh`
- `tests/test_end_to_end.py`
- `tests/test_public_config_end_to_end.py`
- `tests/test_watch.py`
