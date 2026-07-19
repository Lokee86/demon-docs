# Initial Discovery Results

This directory retains reports from the first OSS candidate search performed on 2026-07-18.

The reports are exploratory evidence only. Their scores estimated initial benchmark usefulness and must not be interpreted as software-quality judgments or final quadrant assignments.

## Retained Files

- `initial-candidates.json`: complete mechanically scored candidate data.
- `initial-shortlist.md`: the original ranked shortlist.
- `initial-task-inventory.json` and `initial-task-inventory.md`: historical issue/PR task records and pinned base commits.
- `initial-recommendation.md`: the initial repository recommendation and validation notes.
- `initial-findings.md`: corrected qualitative conclusions after reviewing the search bias.
- `initial-search-workspace.md`: the original standalone-workspace instructions.
- `initial-run.txt`: retained command/failure log from the scan.

## Known Limitations

- emphasis on under-documented Go repositories;
- candidate-pool star bands that biased discovery toward visible projects;
- static heuristics that could miss nested documentation sites;
- incomplete separation of code quality, documentation quality, task quality, and operational feasibility; and
- limited manual validation.

Future searches should use the matrix and protocol in [Context-Injection Benchmarking](../../../docs/context-injection-benchmarking.md). The historical scanner is retained under `../tools/initial-scan/` only so the initial results can be reproduced and audited.
