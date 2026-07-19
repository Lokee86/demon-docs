# Mixed Paths And Boundaries

## Code map

Primary files:

```text
client/scripts/devtools/window.gd
client/scripts/devtools/window_controller.gd
```

Important non-ownership boundaries:

* `client/scripts/devtools/window.gd` - Owns presentation, not gameplay mutation.
* `WindowController` - Owns window lifecycle.
