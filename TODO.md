# AndroidEmulatorMCP — roadmap

The Android counterpart to [XcodeBuildMCP](https://github.com/getsentry/XcodeBuildMCP).
This file is the lean hub — details live in linked docs so it stays readable.

**Current:** v0.5.0 · 46 tools + 4 guide resources · [tool reference in README](README.md#tools)

## Map

| Doc | What's in it |
|---|---|
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | Everything shipped, newest first (v0.1.0 → v0.5.0) |
| [docs/BACKLOG.md](docs/BACKLOG.md) | Open ideas + the conventions to follow when adding a tool |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Two-layer mirror layout + how to add a tool (with diagram) |

## Next up

Pulled from [docs/BACKLOG.md](docs/BACKLOG.md) — see there for the full context:

- [ ] Multi-touch / pinch-zoom (needs `sendevent`; single-pointer `drag` already shipped)
- [ ] `set_status_bar` richer demo controls (mobile signal, operator, notification icons)
- [ ] Deeper test-report insight (stack traces, per-suite timing, JSON output)
- [ ] `input_key_combo` named presets (`select_all`, `copy`, `paste`, …)

## Ground rules

- Every device-facing tool takes an optional `serial`; single-device sessions omit it.
- `internal/android` stays pure/testable; `internal/tools` stays a thin MCP binding (see [ARCHITECTURE.md](ARCHITECTURE.md)).
- Unit-test any new pure logic (parsers, coordinate math, arg parsing).
