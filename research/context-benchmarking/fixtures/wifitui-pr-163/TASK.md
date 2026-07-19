# Benchmark Task: `shazow/wifitui` PR #163

## Task

### UI: Span border full window width

The window border currently shrinks to fit the content rather than using the full terminal width. It should span the full window, and columns should adjust dynamically to take advantage of additional width where available.

An earlier work-in-progress pull request existed, but the benchmark agent must work only from the pinned pre-change repository and this task statement.

## Verification

```bash
go test ./...
```
