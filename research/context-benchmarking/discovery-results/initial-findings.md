# Initial Candidate-Search Findings

Date: 2026-07-18

The first exploratory search examined active Go repositories with usable tests and historical pull-request data. It produced useful candidate tasks but did not implement the final benchmark design.

## Candidate observations

### `shazow/wifitui`

Potential role: **good code / poor documentation**, pending formal review.

Observed strengths:

- approximately 8,700 handwritten Go lines across several package boundaries;
- multiple backends, OS-specific composition, terminal UI state, CLI behavior, themes, and mocks;
- a current full Go test suite that passed during the exploratory scan; and
- multiple bounded, issue-linked, test-backed historical changes.

Observed documentation gap:

- primarily a user-facing README;
- no clear repository architecture, ownership, or implementation-flow documentation found during the scan.

This does not establish that the code is good or bad. It only makes the repository a plausible under-documented candidate.

### `crossplane-contrib/provider-sql`

Potential role: **harder ownership and generated-surface benchmark**, quadrant unclassified.

Observed characteristics:

- approximately 31,000 handwritten Go lines across many packages;
- mirrored cluster-scoped and namespaced implementations;
- APIs, controllers, examples, CRDs, tests, and generated artifacts that may need synchronized changes; and
- useful historical issue/PR pairs.

This repository may test whether context prevents agents from missing mirrored or generated consequences, but domain and dependency complexity make it a poor first harness target.

### Other explored candidates

- `openstack-exporter/openstack-exporter` had useful issue and test history, but its default full suite required live OpenStack configuration.
- `mercuretechnologies/expo-open-ota` had more nested documentation than the initial scanner detected and showed platform-specific test friction.

## Historical `wifitui` tasks retained

Three task manifests were identified at pinned pre-change commits:

| PR | Base commit | Task shape |
|---|---|---|
| 163 | `4583f965beaac68ed8de4cdfebd614645fcbac8a` | Width-aware terminal UI layout across multiple files and tests. |
| 167 | `86e3912f192617d5cbbc001f0ca059f710ddbe3d` | Access-point annotation and theme behavior. |
| 178 | `111f53dee103724c0bbafd155cdcb51f8ab2a731` | `NO_COLOR` behavior. |

All three pre-change snapshots passed `go test ./...` in WSL during the exploratory work.

## Lessons from the first scanner

- High-star bias is not a valid way to find poor code.
- Under-documentation is only one benchmark axis.
- Documentation discovery must inspect nested application/documentation trees.
- Code quality, documentation quality, task quality, and operational feasibility need separate evidence and separate classifications.
- A repository can be an excellent benchmark candidate without being bad code.

The next discovery pass should use the matrix and protocol in `docs/context-injection-benchmarking.md` rather than extending the original single-score ranking.
