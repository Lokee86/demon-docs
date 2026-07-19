# Link Performance Records

This directory records the historical Markdown-link performance results referenced by `docs/link-performance.md`.

## Measurement Environment

- Date: July 19, 2026
- Host: Windows development machine
- Repository: Demon Docs
- Parallel rewrite implementation: commit `12856e3a65a96528c2da2e2304aa30cc420c1824`
- Pre-parallel phase-timing implementation: commit `38228b5e25446f4e7cd112e9e0acee0c848f559a`
- Rewrite worker count: 16

Results are wall-clock engineering measurements and can vary with filesystem cache state, antivirus activity, storage load, and host scheduling.

## Synthetic 250-Source Target Move

The fixture creates 250 Markdown files that all link to `asset-a.bin`, establishes the baseline, renames the target to `asset-moved.bin`, then reconciles and applies every source rewrite.

### Before bounded parallel source writes

- filesystem rewrites: 817–881 ms;
- generated-source refresh: 30–32 ms;
- `.ddocs` publication: approximately 38 ms;
- complete apply phase: 885–954 ms; and
- complete operation: 1.025–1.095 s.

### After bounded parallel source writes

- filesystem rewrites: 261–299 ms;
- generated-source refresh: 11–12 ms;
- complete apply phase: 322–358 ms; and
- complete operation: 505–586 ms.

The implementation uses 16 workers. Planning remains deterministic, and every worker performs the same expected-hash check, same-directory temporary write, atomic replacement, and source verification required by the sequential implementation.

## Real Space Rocks Target Move

Copied-corpus target: `services/game-server/!INDEX.md`

- incoming link occurrences: 106;
- rewritten source files: 96;
- reconciliation state load: 57.1 ms;
- inventory build: 12.4 ms;
- reconciliation planning: 888.4 ms;
- total reconciliation: 957.9 ms;
- filesystem rewrites: 898.3 ms;
- source verification and refresh: 44.9 ms;
- `.ddocs` publication: 136.8 ms; and
- total application: 1.08 s.

This result predates the bounded parallel rewrite optimization.

## Mass Rename

The full correctness logs and five-run timing data are stored in sibling directories:

- `../mass-rename-results/`
- `../mass-rename-timing/`

Each mass-rename pass renamed 341 Markdown files, rewrote 340 Markdown sources, and repaired 3,717 link destinations. Median `ddocs fix -l` time was 1.928 seconds on the first pass and 1.980 seconds on the repeated pass.