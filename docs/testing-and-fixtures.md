# Testing and Fixtures

doc-ledger is covered by focused Go package tests, watcher integration coverage, and Go CLI fixture regression gates. Go is the sole implementation and supported runtime.

## Test Commands

For an install smoke check, run:

```bash
go install ./cmd/doc-ledger
doc-ledger --help
doc-ledger --version
```

From the repo root, run the complete local release gate:

```bash
make release-check
```

The individual gates are:

```bash
make test-go
make regression
make vet
make build
make smoke
```

`make regression` runs the Go CLI fixture regression matrix. It builds the binary once, then runs the ten retained fixture scenarios through `fix`, verifies that `check` succeeds on the clean result, runs `fix` again, and requires the complete fixture tree to be byte-identical after the first and second fixes.

The ten scenarios cover defaults; custom index headings, markers, drafts, and non-Markdown editable files; direct-to-stub transition; stub graduation; unique and ambiguous file moves; unique and ambiguous folder moves; stale entry removal; and malformed managed blocks. Focused Go specification tests cover Goldmark ignoring fenced headings and exact final-newline preservation.

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

## Continuous Integration

`.github/workflows/ci.yml` runs:

- the complete Go suite, including `./tests`, on Linux and Windows;
- `go vet ./...`;
- an executable build and basic CLI smoke tests.

## Release Requirements

A release is eligible only when all CI jobs pass. In particular:

- Linux and Windows Go tests are green;
- the ten-fixture Go CLI regression matrix is green;
- focused specification tests cover the intentional compatibility corrections;
- `go vet`, the executable build, and CLI smoke checks are green;
- repeated reconciliation is byte-identical and check mode remains non-mutating.

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

A simple end-to-end smoke test uses the Go CLI:

```bash
./docs/make-dummy-docs.sh
doc-ledger fix --root dummy-docs
doc-ledger check --root dummy-docs
```

If you are working from the repo checkout, `go run ./cmd/doc-ledger` is the primary fallback. After that, try a move or rename inside `dummy-docs/`, run `fix` and `check` again, and inspect the diff.

## Fixture Guidance

- Dummy docs are manual stress fixtures, not canonical docs.
- Keep them disposable.
- Use them when you want a noisy tree that exercises nesting, stubs, and cross-folder reconciliation.

## Related Files

- `docs/make-dummy-docs.sh`
- `tests/regression_test.go`
- `tests/regression_fixtures_test.go`
- `internal/reconcile/reconcile_test.go`
- `internal/watch/watch_test.go`
