# Watcher and Automation

doc-ledger has a watch mode for long-running docs maintenance, but it is still a convenience layer. Use `check` when you want a clean verification gate before commit or CI.

## Watch Commands

- `doc-ledger fix --root docs`
- `doc-ledger check --root docs`
- `doc-ledger watch --root docs`
- `doc-ledger watch --root docs --once`

`doc-ledger watch --help` shows the watch-specific flags and examples.

`--once` runs a single reconciliation pass and exits. Regular watch mode runs one reconciliation pass immediately, then keeps observing the docs tree recursively.
Watch mode runs in the foreground by default. `doc-ledger` does not daemonize or own background lifecycle.

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
- It adds watches for newly created nested directories and removes deleted or renamed watched directories from its tracked set.
- Observer errors are surfaced instead of silently terminating observation.

The watcher prints timestamped status lines and includes the current process ID in its startup line. That makes it easier to spot watcher/fix races in logs.

## Practical Usage

Watch mode is useful while iterating locally, but it is not a replacement for `check`.

- Use `watch` while editing docs and fixtures.
- Use `check` before commit or in CI.
- Expect `fix` to report `0` updated files when the watcher already reconciled the tree for you.

## Safer Detached Example

If you run doc-ledger from a shell startup file, keep the launch explicit and let the wrapper handle the guard logic:

```bash
DOC_LEDGER_ROOT="${DOC_LEDGER_ROOT:-docs}"

doc_ledger_pid_file="$PWD/.cache/doc-ledger-watch.pid"
doc_ledger_log_file="$PWD/.cache/doc-ledger-watch.log"

mkdir -p "$PWD/.cache"

doc_ledger_watch_is_running() {
  [ -s "$doc_ledger_pid_file" ] || return 1

  local watcher_pid
  watcher_pid="$(cat "$doc_ledger_pid_file" 2>/dev/null)" || return 1

  case "$watcher_pid" in
    ''|*[!0-9]*) return 1 ;;
  esac

  kill -0 "$watcher_pid" 2>/dev/null || return 1
  ps -p "$watcher_pid" -o args= 2>/dev/null | grep -Fq "doc-ledger watch"
}

start_doc_ledger_watch() {
  setsid bash -c '
    cd "$1" || exit 1
    exec doc-ledger watch --root "$2" </dev/null >>"$3" 2>&1
  ' _ "$PWD" "$DOC_LEDGER_ROOT" "$doc_ledger_log_file" >/dev/null 2>&1 &

  echo $! > "$doc_ledger_pid_file"
}

if ! doc_ledger_watch_is_running; then
  rm -f "$doc_ledger_pid_file"
  start_doc_ledger_watch
fi

unset doc_ledger_pid_file
unset doc_ledger_log_file
```

When using `direnv`, source process startup files outside any `set -a` block so helper variables such as PID and log paths are not exported.

This pattern keeps the process start explicit, writes logs to `.cache/doc-ledger-watch.log`, and avoids relying on shell aliases during startup.

## Output

Watcher logs include timestamped status lines such as:

```text
2026-06-18T23:59:59 doc-ledger watch watching docs pid=12345
2026-06-18T23:59:59 doc-ledger watch updated 3 file(s)
```

Those timestamps make it easier to understand the order of events when a fix pass and a file change happen close together.

Watcher unit and temporary-filesystem integration tests cover source and destination rename events, nested directory creation, watched-directory deletion, configured filtering, events queued during reconciliation, explicit debounce overrides, observer errors, clean cancellation, and self-write convergence. The CI matrix runs the watcher package as part of the Linux and Windows Go suites.

## Related Files

- `internal/watch/watch.go`
- `internal/app/app.go`
- `cmd/doc-ledger/main.go`
- `docs/make-dummy-docs.sh`
