# adb_mcp — roadmap

The Android counterpart to [XcodeBuildMCP](https://github.com/getsentry/XcodeBuildMCP).
This file is the lean hub — details live in linked docs so it stays readable.

**Current:** v0.9.0 · 48 tools + 4 guide resources · [tool reference in README](README.md#tools)
Installed + smoke-tested live on an emulator (screenshot black-frame detection verified; reload_app verified against a real Expo dev client; status bar + key-combo presets verified).
Core parity with [XcodeBuildMCP](https://github.com/getsentry/XcodeBuildMCP) reached; remaining gaps below.

## Map

| Doc | What's in it |
|---|---|
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | Everything shipped, newest first (v0.1.0 → v0.9.0) |
| [docs/BACKLOG.md](docs/BACKLOG.md) | Open ideas + the conventions to follow when adding a tool |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Two-layer mirror layout + how to add a tool (with diagram) |

## Next up

Pulled from [docs/BACKLOG.md](docs/BACKLOG.md) — see there for full context.

**Field feedback** (from real a partner app debugging sessions — see BACKLOG.md)
- [ ] `last_crash` — pull `dumpsys dropbox` so a full crash stack comes back in one call
- [ ] Bound `stop_logcat_capture` output by default (tail cap or summary+file)
- [ ] Clearer `launch_app` failure output + dev-client (Expo Dev Launcher) awareness
- [ ] `logcat` buffer-rotation hint; `screenshot`/`describe_ui` state-skew note

**XcodeBuildMCP parity gaps** (priority order)
- [ ] `build_and_run` — one-shot build → install → launch (highest value / lowest lift)
- [ ] Deeper project discovery — `list_gradle_variants` + module/build-info dump
- [ ] `last_crash` — structured crash extractor (`dumpsys dropbox`/tombstone → fields)
- [ ] Project scaffolding — new Android project from a template (biggest lift)

**Enhancements**
- [ ] Multi-touch / pinch-zoom (needs `sendevent`; single-pointer `drag` already shipped) — parked, no reliable cross-device approach yet

## Ground rules

- Every device-facing tool takes an optional `serial`; single-device sessions omit it.
- `internal/android` stays pure/testable; `internal/tools` stays a thin MCP binding (see [ARCHITECTURE.md](ARCHITECTURE.md)).
- Unit-test any new pure logic (parsers, coordinate math, arg parsing).
