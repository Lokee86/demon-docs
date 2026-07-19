# Demon Docs Benchmark Search

This workspace discovers open-source repositories suitable for evaluating deterministic context injection.

The scanner:

- searches GitHub for active, moderately sized Go repositories;
- shallow-clones a bounded candidate set;
- measures code, package, test, CI, and documentation characteristics;
- inspects merged pull-request history for realistic benchmark tasks;
- ranks repositories by testability, documentation gap, structural complexity, and historical-task quality;
- writes machine-readable data and a human-readable shortlist under `results/`.

## Run

```bash
python scan_candidates.py
```

The script requires authenticated `gh`, `git`, and Python 3. It only writes beneath this directory.

## Outputs

- `results/candidates.json` — complete scored candidate data.
- `results/shortlist.md` — ranked shortlist and selection notes.
- `results/task_inventory.json` / `.md` — issue/PR task records with pinned base commits.
- `results/recommendation.md` — validated repository decision and experiment sequence.
- `results/run.log` — command and failure log.
- `clones/<run-id>/` — isolated shallow clones used for static analysis.
- `benchmarks/` — prepared pre-change source snapshots, task text, oracle metadata, and baseline results.

## Supporting commands

```bash
python validate_candidates.py
python prepare_benchmarks.py
python validate_benchmark_snapshots.py
```

The baseline validator uses WSL because historical `wifitui` snapshots contain Linux-specific NetworkManager tests that are not expected to compile as Windows test binaries.
