# Astra Relay Documentation

Astra Relay is a fictional space-operations platform used to demonstrate adopting Demon Docs in an existing documentation repository.

The authored material is useful, but its metadata, structure, navigation, and organization have drifted. This page lists the complete intentional starting condition so every tutorial result can be inspected manually.

## Repository-wide shortcomings

- The documentation tree has no local folder indexes. This is not invalid Markdown, but it reduces navigability across eleven indexed folders: `docs`, `concepts`, `concepts/deep`, `guides`, `guides/troubleshooting`, `notes`, `old-system`, `old-system/assets`, `old-system/storage`, `planning`, and `planning/phases`.
- The service documentation still lives under the legacy `docs/old-system` name, with generic filenames such as `api-notes.md` and `worker-notes.md`. The tutorial reorganizes this area without breaking its references.
- `docs/private-notes/unstructured-notes.md` is intentionally excluded through `.docignore` and is not a Demon Docs problem.

## Link and relationship problems

- `docs/home.md` contains `[[overview|project overview]]`, which is ambiguous because both `docs/concepts/overview.md` and `docs/guides/overview.md` match it.
- `docs/notes/launch-retrospective.md` is useful but orphaned: no managed document links to it.
- The remaining authored Markdown links, wiki links, aliases, heading fragments, image links, and wiki embeds resolve correctly at the start. They are included to verify that later moves and renames preserve them.

## Complete document-policy inventory

The initialized tutorial policy requires YAML fields `document_id`, `author`, `document_type`, `created`, `summary`, and `policy_exempt`. Document IDs must be UUIDs. The table below lists every intentional metadata and schema problem in the managed documents.

| Document | Metadata problems | Structure problems |
| --- | --- | --- |
| `docs/home.md` | All required fields missing. | Sections out of order; `Overview` missing. |
| `docs/getting-started.md` | Non-UUID ID; `author` and `summary` missing; unknown `team` field. | Sections out of order; `Notes` missing. |
| `docs/quick-start.md` | All required fields missing. | `Notes` and `Related docs` missing. |
| `docs/glossary.md` | Non-UUID ID duplicated by `concepts/deep/terminology.md`; `summary` missing. | Sections out of order; `Related docs` missing. |
| `docs/concepts/architecture.md` | All required fields missing. | Sections out of order; `Overview` missing. |
| `docs/concepts/configuration.md` | Non-UUID ID; `summary` missing; unknown `review_cycle` field. | Sections out of order; `Related docs` missing. |
| `docs/concepts/navigation.md` | All required fields missing. | `Overview` and `Related docs` missing. |
| `docs/concepts/overview.md` | All required fields missing. | Sections out of order; `Notes` missing. |
| `docs/concepts/deep/lifecycle.md` | All required fields missing. | Sections out of order; `Related docs` missing. |
| `docs/concepts/deep/terminology.md` | Non-UUID ID duplicated by `glossary.md`. | `Notes` and `Related docs` missing. |
| `docs/guides/local-setup.md` | All required fields missing. | Sections out of order; `Related docs` missing. |
| `docs/guides/deployment.md` | All required fields missing. | Sections out of order; `Related docs` missing; authored `Rollout Checklist` is not in the shared schema and requires an explicit preserve/delete decision. |
| `docs/guides/monitoring.md` | All required fields missing. | Sections out of order; `Related docs` missing. |
| `docs/guides/overview.md` | All required fields missing. | `Notes` missing. |
| `docs/guides/troubleshooting/common-errors.md` | Non-UUID ID; `author` missing. | `Related docs` missing. |
| `docs/guides/troubleshooting/networking.md` | All required fields missing. | Sections out of order; `Notes` and `Related docs` missing. |
| `docs/guides/troubleshooting/permissions.md` | All required fields missing. | `Related docs` missing. |
| `docs/old-system/api-notes.md` | Non-UUID ID duplicated by `worker-notes.md`; `author` and `summary` missing; unknown `legacy_status` field. | Sections out of order; `Code map`, `Data ownership`, `Protocol and API surfaces`, `Related docs`, and `Tests and verification` missing. |
| `docs/old-system/worker-notes.md` | Non-UUID ID duplicated by `api-notes.md`. | Sections out of order; duplicate `Responsibilities`; `Code map`, `Data ownership`, `Protocol and API surfaces`, `Related docs`, and `Tests and verification` missing. |
| `docs/old-system/storage/storage-notes.md` | Non-UUID ID; `author` and `summary` missing. | `Code map`, `Notes`, `Protocol and API surfaces`, `Responsibilities`, and `Tests and verification` missing. |
| `docs/old-system/storage/backups.md` | Non-UUID ID. | Sections out of order; `Code map`, `Data ownership`, `Notes`, `Protocol and API surfaces`, `Related docs`, and `Tests and verification` missing. |
| `docs/planning/roadmap.md` | Non-UUID ID. | Sections out of order; `Implementation Implications`, `Notes`, `Settled Product Model`, and `System Handoffs` missing. |
| `docs/planning/release-plan.md` | All required fields missing. | `Notes`, `Ownership Boundary`, `Related Docs`, and `System Handoffs` missing. |
| `docs/planning/migration.md` | Non-UUID ID; `author` and `summary` missing; unknown `owner` field. | Sections out of order; `Implementation Implications`, `Related Docs`, and `Settled Product Model` missing. |
| `docs/planning/phases/phase-one.md` | All required fields missing. | `Implementation Implications`, `Notes`, `Related Docs`, and `Settled Product Model` missing. |
| `docs/planning/phases/phase-two.md` | All required fields missing. | Sections out of order; `Notes`, `Ownership Boundary`, `Related Docs`, and `System Handoffs` missing. |
| `docs/notes/launch-retrospective.md` | All required fields missing. | `Related docs` missing. |
| `docs/stubs/future-integrations.md` | All required fields missing. | `Purpose`, `Overview`, `Notes`, and `Related docs` missing. |

Demon Docs should report this starting condition, repair deterministic issues, and stop for the two authored structural decisions: preserving `Rollout Checklist` and merging the duplicate `Responsibilities` sections.
