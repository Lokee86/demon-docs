# Legacy Indented Map

## Code map

Primary implementation files:

  `services/example/session.go`

    Defines `SessionState`.
    Owns the session cooldown.

  `services/example/input.go` routes `PacketTypeResume` into the session owner.
