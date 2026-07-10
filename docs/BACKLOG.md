# Backlog & ideas

Open, unstarted work. Shipped history is in [CHANGELOG.md](CHANGELOG.md).

## Open

- [ ] **Multi-touch / pinch-zoom gestures.** The single-pointer half shipped as `drag` (`input draganddrop`). True two-finger pinch/rotate needs the `sendevent` multi-touch protocol, which is device/kernel-specific (the `input` command has no multi-pointer verb) — parked until there's a reliable cross-device approach.
- [ ] **`set_status_bar` — richer demo controls.** Extend beyond clock/battery to mobile-signal type (`network -e mobile ...`), carrier/operator name, and notification-icon injection.
- [ ] **Deeper test-report insight.** Attach failure stack traces (not just the first message line) and per-suite timing; optionally emit machine-readable JSON alongside the human summary.
- [ ] **`input_key_combo` presets.** Named shortcuts (e.g. `select_all`, `copy`, `paste`) that expand to the right chord, so callers don't need to know keycodes.

## Conventions (read before adding a tool)

- Every device-facing tool takes an optional `serial`; single-device sessions can omit it.
- Keep `internal/android` pure/testable; `internal/tools` stays a thin MCP binding. Each `tools/<domain>.go` mirrors an `android/<domain>.go` — see [../ARCHITECTURE.md](../ARCHITECTURE.md).
- Add unit tests for any new pure logic (parsers, coordinate math, arg parsing).
- Feedback/bugs also live in the Council-Hub room `android-emulator-mcp-feedback`.
