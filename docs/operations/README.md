---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-7e24-b520-215f4eac33af
document_type: general
policy_exempt: false
summary: Running behavior, watcher and demon ownership, runtime state, recovery, troubleshooting, and operational verification.
---
# Operations

Parent index: [Demon Docs Documentation](../README.md)

Running behavior, watcher and demon ownership, runtime state, recovery, troubleshooting, and operational verification.

## Direct Files

<!-- doc-ledger:files:start -->

- [dynamic-watch-scope.md](dynamic-watch-scope.md) - Feature-selected watch roots, directory admission, `.docignore` reload, external targets, and reverse-root refresh.
- [host-adapters.md](host-adapters.md) - Integrate MCP, agent, editor, or other hosts through acquire, heartbeat, and release feeder commands.
- [recovery-and-troubleshooting.md](recovery-and-troubleshooting.md) - Diagnose configuration, stale automation, state corruption, ambiguous links, and unexpected reconciliation.
- [repository-demon.md](repository-demon.md) - Single-owner repository watcher lifecycle, feeders, worktrees, runtime state, shutdown, and logs.
- [watcher-and-automation.md](watcher-and-automation.md) - Foreground watch behavior and its relationship to the repository demon.
<!-- doc-ledger:files:end -->

## Direct Folders

<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Notes

Operational automation is optional. Static `check` and `fix` commands remain the authoritative correctness and recovery surfaces.
