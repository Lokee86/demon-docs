# Testing and Fixtures

Parent index: [Docs](./!README.md)

doc-ledger is covered by focused pytest tests plus a few end-to-end flows that exercise real docs trees.

## Test Command

From `tools/doc-ledger`, run:

```bash
python3 -m pytest tests
```

## What the Tests Cover

The doc-ledger tests are split across small, focused areas:

- CLI behavior
- config loading and discovery
- scan model construction
- README IO and managed section handling
- parent index behavior
- reconciliation planning and file updates
- watcher scheduling and filtering
- end-to-end flows
- public config examples

Those tests keep the implementation honest without depending on the full Space Rocks game runtime.

## Dummy Docs Fixture Generator

`tools/doc-ledger/docs/make-dummy-docs.sh` is a manual stress and fixture generator. It builds a nested docs tree with randomly shaped folders and files so you can exercise reconciliation against a larger input.

Run it with:

```bash
./tools/doc-ledger/docs/make-dummy-docs.sh
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

That output is not ignored by the repo’s `.gitignore`, so remove it when you are done:

```bash
rm -rf dummy-docs
```

## Manual Smoke Workflow

A simple end-to-end smoke test looks like this:

```bash
./docs/make-dummy-docs.sh
python3 main.py fix --root dummy-docs
python3 main.py check --root dummy-docs
```

After that, try a move or rename inside `dummy-docs/`, run `fix` and `check` again, and inspect the diff. That is a good way to verify description preservation, stale entry removal, and parent index updates on a realistic tree.

## Fixture Guidance

- Dummy docs are manual stress fixtures, not canonical docs.
- Keep them disposable.
- Use them when you want a noisy tree that exercises nesting, stubs, and cross-folder reconciliation.

## Related Files

- `tools/doc-ledger/docs/make-dummy-docs.sh`
- `tools/doc-ledger/tests/test_end_to_end.py`
- `tools/doc-ledger/tests/test_public_config_end_to_end.py`
- `tools/doc-ledger/tests/test_watch.py`
