# doc-ledger Docs

Parent index: [Docs](./!README.md)

The repo root [README.md](/mnt/d/!bin/space-rocks/README.md) is the starting point for doc-ledger.
This `docs/` folder holds deeper operational and maintenance references for the tool.

## References

- [Configuration](configuration.md): Config file shape, defaults, and supported overrides.
- [Reconciliation Model](reconciliation-model.md): How doc-ledger scans, plans, and applies index updates.
- [Watcher and Automation](watcher-and-automation.md): Watch mode behavior, timestamps, PID output, and automation guidance.
- [Testing and Fixtures](testing-and-fixtures.md): Test layout, fixture strategy, and regression coverage for doc-ledger.
- [Dummy Docs Fixture Generator](../tools/doc-ledger/docs/make-dummy-docs.sh): Manual fixture and stress generator for recursive docs-tree testing.

## Notes

- `docs/!README.md` remains the primary docs-tree index for the wider Space Rocks documentation set.
- doc-ledger keeps Python cache files out of commits through `.gitignore` and the repo hygiene test.