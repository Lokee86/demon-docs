# Architecture

## Related docs

- [Configuration](configuration.md)
- [Navigation](navigation.md)
- [Concept overview](docs/concepts/overview.md)
- [Lifecycle](lifecycle.md)
- [Terminology](terminology.md)
- [[api-notes|API service notes]]
- [Deployment path][deployment-guide]

[deployment-guide]: ../guides/deployment.md

## Purpose

Describe the major project boundaries and how telemetry flows between them.

## Notes

The current design favors explicit handoffs over shared mutable state.
