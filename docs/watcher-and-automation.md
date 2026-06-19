# Watcher and Automation

doc-ledger has a watch mode for long-running docs maintenance, but it is still a convenience layer. Use `check` when you want a clean verification gate before commit or CI.

## Watch Commands

- `doc-ledger fix --root docs`
- `doc-ledger check --root docs`
- `doc-ledger watch --root docs`
- `doc-ledger watch --root docs --once`

`doc-ledger watch --help` shows the watch-specific flags and examples.

`--once` runs a single reconciliation pass and exits. Regular watch mode runs one reconciliation pass immediately, then keeps observing the docs tree recursively.

## What Watch Mode Does

Watch mode is built to rerun reconciliation when the docs tree changes.

- It starts with an initial fix so the tree is reconciled before observation begins.
- It watches the configured root recursively.
- It reacts to relevant file events and directory create, delete, and move events.
- It debounces bursts of changes.
- It runs one fix at a time.
- If changes arrive during a fix, it schedules one more pass after the current run finishes.
- It ignores configured ignored directories and ignored filename suffixes.
- It applies the same include and exclude rules used by scanning when deciding whether a file event matters.

The watcher prints timestamped status lines and includes the current process ID in its startup line. That makes it easier to spot watcher/fix races in logs.

## Practical Usage

Watch mode is useful while iterating locally, but it is not a replacement for `check`.

- Use `watch` while editing docs and fixtures.
- Use `check` before commit or in CI.
- Expect `fix` to report `0` updated files when the watcher already reconciled the tree for you.

## Example `processes.env`

If you run doc-ledger under a lightweight process manager, keep the command direct and let the wrapper handle the guard logic:

```bash
export DOC_LEDGER_ROOT="${DOC_LEDGER_ROOT:-docs}"

pid_file="$PWD/.cache/doc-ledger-watch.pid"
log_file="$PWD/.cache/doc-ledger-watch.log"

mkdir -p .cache

start_doc_ledger_watch() {
  nohup doc-ledger watch --root "$DOC_LEDGER_ROOT" > "$log_file" 2>&1 &
  echo $! > "$pid_file"
}

if [ -s "$pid_file" ]; then
  pid="$(cat "$pid_file")"
  if kill -0 "$pid" 2>/dev/null && ps -p "$pid" -o args= | grep -q "doc-ledger watch"; then
    :
  else
    start_doc_ledger_watch
  fi
else
  start_doc_ledger_watch
fi
```

This pattern keeps the process start explicit, writes logs to `.cache/doc-ledger-watch.log`, and avoids relying on shell aliases during startup.

## Output

Watcher logs include timestamped status lines such as:

```text
2026-06-18T23:59:59 doc-ledger watch watching docs pid=12345
2026-06-18T23:59:59 doc-ledger watch updated 3 file(s)
```

Those timestamps make it easier to understand the order of events when a fix pass and a file change happen close together.

## Related Files

- `doc_ledger/watch.py`
- `doc_ledger/cli.py`
- `docs/make-dummy-docs.sh`
