# Bulleted Described Targets

## Code map

### Primary implementation

* `client/scripts/session/controller.gd` - Creates `SessionState` and owns session transitions.
- `client/scripts/session/state.gd` — Stores durable session state.

### Symbol boundary

* `SessionController` - Owns orchestration, not transport.
