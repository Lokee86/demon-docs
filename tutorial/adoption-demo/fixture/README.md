# Astra Relay Documentation

Astra Relay is a fictional space-operations platform used to demonstrate adopting Demon Docs in an existing documentation repository.

The authored material is useful, but its metadata, structure, navigation, and organization have drifted. This page lists the complete intentional starting condition so every tutorial result can be inspected manually.

## Repository-wide shortcomings

- The documentation tree has no local folder indexes. This is not invalid Markdown, but it reduces navigability across eleven indexed folders: `docs`, `concepts`, `concepts/deep`, `guides`, `guides/troubleshooting`, `notes`, `old-system`, `old-system/assets`, `old-system/storage`, `planning`, and `planning/phases`.
- The service documentation still lives under the legacy `docs/old-system` name, with generic filenames such as `api-notes.md` and `worker-notes.md`. The tutorial reorganizes this area without breaking its references.
- The managed corpus contains 175 authored local link occurrences across Markdown links, wiki links, aliases, heading fragments, one reference-style link, an image link, and a wiki image embed. This is intentionally dense enough for the move and rename demonstrations to affect a real document graph.
- Twenty-three of the twenty-eight managed Markdown documents already contain YAML frontmatter. Some are valid, while others contain missing or empty values, an invalid date, weak or duplicated IDs, and unknown fields. Five documents have no frontmatter at all.
- `docs/private-notes/unstructured-notes.md` is intentionally excluded through `.docignore` and is not a Demon Docs problem.

## Link and relationship problems

- `docs/getting-started.md` links to the stale path `concepts/archive/configuration.md`; the only matching `configuration.md` document is `docs/concepts/configuration.md`.
- `docs/old-system/api-notes.md` links to the stale path `storage/archive/storage-notes.md#retention`; the only matching `storage-notes.md` document is `storage/storage-notes.md#retention`.
- `docs/home.md` contains `[[overview|project overview]]`, which is ambiguous because both `docs/concepts/overview.md` and `docs/guides/overview.md` match it.
- `docs/notes/launch-retrospective.md` is intentionally orphaned: it links outward to related work, but no managed document links back to it.
- The other 172 authored local references resolve correctly at the start. They are included to verify that later moves and renames preserve and rewrite a substantial graph rather than one isolated example.

## Complete document-policy inventory

The initialized tutorial policy requires YAML fields `document_id`, `author`, `document_type`, `created`, `summary`, and `policy_exempt`. Document IDs must be UUIDs. The table below lists every intentional metadata and schema problem in the managed documents, including documents whose frontmatter is already valid.

| Document | Metadata condition | Structure condition |
| --- | --- | --- |
| `docs/home.md` | No frontmatter; all required fields missing. | Sections out of order; `Overview` missing. |
| `docs/getting-started.md` | Non-UUID ID; `author` and `summary` missing; unknown `team` field. | Sections out of order; `Notes` missing. |
| `docs/quick-start.md` | Partial frontmatter; `document_id`, `author`, `created`, and `summary` missing. | `Notes` missing. |
| `docs/glossary.md` | Non-UUID ID; `summary` missing. | Sections out of order. |
| `docs/concepts/architecture.md` | No frontmatter; all required fields missing. | Sections out of order; `Overview` missing. |
| `docs/concepts/configuration.md` | Non-UUID ID; `summary` missing; unknown `review_cycle` field. | Sections out of order. |
| `docs/concepts/navigation.md` | Non-UUID ID; empty `author`; impossible `created` date (`2026-02-30`). | Sections out of order; `Overview` missing. |
| `docs/concepts/overview.md` | Non-UUID ID `overview`, also used by `docs/guides/overview.md`. | Sections out of order; `Notes` missing. |
| `docs/concepts/deep/lifecycle.md` | Valid UUID and otherwise complete frontmatter, but `author` is empty. | Sections out of order. |
| `docs/concepts/deep/terminology.md` | Complete valid YAML frontmatter. | `Notes` missing. |
| `docs/guides/local-setup.md` | Valid UUID; `created` uses `June 4, 2026` instead of `YYYY-MM-DD`; `summary` missing. | Sections out of order. |
| `docs/guides/deployment.md` | No frontmatter; all required fields missing. | Sections out of order; authored `Rollout Checklist` is not in the shared schema and requires an explicit preserve/delete decision. |
| `docs/guides/monitoring.md` | Non-UUID ID; remaining required fields are present. | Sections out of order. |
| `docs/guides/overview.md` | Non-UUID ID `overview`, also used by `docs/concepts/overview.md`; empty `summary`. | `Notes` missing. |
| `docs/guides/troubleshooting/common-errors.md` | Non-UUID ID; `author` missing. | Sections out of order. |
| `docs/guides/troubleshooting/networking.md` | Complete valid YAML frontmatter. | Sections out of order; `Notes` missing. |
| `docs/guides/troubleshooting/permissions.md` | Valid UUID, but `author` is empty. | Sections out of order. |
| `docs/old-system/api-notes.md` | Non-UUID ID `service-node`, also used by `worker-notes.md`; `author` and `summary` missing; unknown `legacy_status` field. | Sections out of order; `Code map`, `Data ownership`, `Protocol and API surfaces`, and `Tests and verification` missing. |
| `docs/old-system/worker-notes.md` | Non-UUID ID `service-node`, also used by `api-notes.md`; other frontmatter fields are present. | Sections out of order; duplicate `Responsibilities`; `Code map`, `Data ownership`, `Protocol and API surfaces`, and `Tests and verification` missing. |
| `docs/old-system/storage/storage-notes.md` | Non-UUID ID; `author` and `summary` missing. | `Code map`, `Notes`, `Protocol and API surfaces`, `Responsibilities`, and `Tests and verification` missing. |
| `docs/old-system/storage/backups.md` | Complete valid YAML frontmatter. | Sections out of order; `Code map`, `Data ownership`, `Notes`, `Protocol and API surfaces`, and `Tests and verification` missing. |
| `docs/planning/roadmap.md` | Complete valid YAML frontmatter. | Sections out of order; `Implementation Implications`, `Notes`, `Settled Product Model`, and `System Handoffs` missing. |
| `docs/planning/release-plan.md` | Complete valid YAML frontmatter. | `Notes`, `Ownership Boundary`, and `System Handoffs` missing. |
| `docs/planning/migration.md` | Non-UUID ID; `author` and `summary` missing; unknown `owner` field. | Sections out of order; `Implementation Implications` and `Settled Product Model` missing. |
| `docs/planning/phases/phase-one.md` | Non-UUID ID; remaining required fields are present. | `Implementation Implications`, `Notes`, and `Settled Product Model` missing. |
| `docs/planning/phases/phase-two.md` | No frontmatter; all required fields missing. | Sections out of order; `Notes`, `Ownership Boundary`, and `System Handoffs` missing. |
| `docs/notes/launch-retrospective.md` | Valid UUID; `author` and `summary` missing. | Sections out of order. |
| `docs/stubs/future-integrations.md` | No frontmatter; all required fields missing. | `Purpose`, `Overview`, and `Notes` missing. |

Demon Docs should report this starting condition, repair deterministic issues, and stop for the two authored structural decisions: preserving `Rollout Checklist` and merging the duplicate `Responsibilities` sections.
