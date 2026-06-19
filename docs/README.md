# doc-ledger Docs

The repo root [README.md](../README.md) is the starting point for doc-ledger.
This `docs/` folder holds deeper operational and maintenance references for the tool.

## References

- [Configuration](configuration.md): Config file shape, defaults, and supported overrides.
- [Reconciliation Model](reconciliation-model.md): How doc-ledger scans, plans, and applies index updates.
- [Watcher and Automation](watcher-and-automation.md): Watch mode behavior, timestamps, PID output, and automation guidance.
- [Testing and Fixtures](testing-and-fixtures.md): Test layout, fixture strategy, and regression coverage for doc-ledger.
- [Dummy Docs Fixture Generator](make-dummy-docs.sh): Manual fixture and stress generator for recursive docs-tree testing.

## Direct Files

<!-- doc-ledger:files:start -->

- [configuration.md](configuration.md) - Configuration documentation.
- [reconciliation-model.md](reconciliation-model.md) - Reconciliation Model documentation.
- [testing-and-fixtures.md](testing-and-fixtures.md) - Testing And Fixtures documentation.
- [watcher-and-automation.md](watcher-and-automation.md) - Watcher And Automation documentation.
<!-- doc-ledger:files:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Notes

- `docs/README.md` is the default docs-tree index.
- doc-ledger keeps Python cache files out of commits through `.gitignore` and the repo hygiene test.