# Security Policy

## Supported versions

Until Demon Docs reaches a stable `1.0` release, security fixes are applied to the latest published release and the current `main` branch. Older releases may not receive backported fixes.

| Version | Supported |
| --- | --- |
| Latest release | Yes |
| Current `main` branch | Yes |
| Older releases | No guaranteed fixes |

## Reporting a vulnerability

Do not open a public issue containing vulnerability details, proof-of-concept material, repository contents, or sensitive logs.

Use GitHub's private vulnerability reporting feature from the repository's **Security** tab when it is available. If private reporting is unavailable, open a public issue containing only a request for a private reporting channel. Do not include the vulnerability details in that issue.

A useful report includes:

- the affected Demon Docs version or commit;
- the operating system and relevant filesystem details;
- the smallest reproducible repository or fixture;
- exact reproduction steps;
- the expected and observed behavior;
- the potential impact; and
- any suggested mitigation.

Reports will be reviewed as soon as practical. Confirmed vulnerabilities will be handled privately until a fix or documented mitigation is available. No fixed response-time guarantee is currently offered.

## Relevant security areas

Demon Docs reads, moves, and rewrites repository files and maintains private state under `.ddocs/`. Reports are particularly relevant when they involve:

- repository-boundary or path-traversal escapes;
- unintended file modification or deletion;
- unsafe symbolic-link handling;
- arbitrary command execution;
- private-state corruption or cross-repository state access;
- daemon ownership or stale-owner failures that permit conflicting writers;
- source-hash or transaction-guard bypasses; or
- sensitive information written to logs or generated output.

## Disclosure

Please allow time for validation, remediation, release preparation, and downstream notification before public disclosure. Security advisories and patch releases will describe confirmed impact without exposing unnecessary private reporter information.
