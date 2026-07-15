# Backlog & ideas

Open, unstarted work. Shipped history is in [CHANGELOG.md](CHANGELOG.md).

## XcodeBuildMCP parity gaps

Core driving/build/test/automate is at parity (and ahead on screen recording,
device-lock/Keystore, custom PIN pads, `tap_on_text`/`wait_for_text`,
`set_status_bar`). These are the remaining gaps vs XcodeBuildMCP:

- [ ] **`build_and_run`** — one-shot `gradle_build` → `install_app` → `launch_app` (optionally on a chosen variant). Currently composable from three calls; XcodeBuildMCP exposes it as a single tool. Highest value / lowest lift.
- [ ] **Deeper project discovery** — no analogue of "list schemes / dump build settings". Add `list_gradle_variants` (parse `./gradlew tasks`/`app:properties` or the `assemble*` task list) and a module/build-info dump, complementing `list_gradle_tasks` + `get_app_details`.
- [ ] **Project scaffolding** — no "create a new Android project from a template" tool (XcodeBuildMCP has `scaffold`). Biggest lift; would need a bundled template + Gradle wrapper generation.
- [ ] **Embedded runtime-crash telemetry** — XcodeBuildMCP bundles a lib that streams structured runtime errors; the Android analogue is `logcat` "Caused by:" triage. A structured crash-extractor over logcat (parse the fatal-exception block into fields) would narrow this.

## Enhancements

- [ ] **Multi-touch / pinch-zoom gestures.** The single-pointer half shipped as `drag` (`input draganddrop`). True two-finger pinch/rotate needs the `sendevent` multi-touch protocol, which is device/kernel-specific (the `input` command has no multi-pointer verb) — parked until there's a reliable cross-device approach.

## Conventions (read before adding a tool)

- Every device-facing tool takes an optional `serial`; single-device sessions can omit it.
- Keep `internal/android` pure/testable; `internal/tools` stays a thin MCP binding. Each `tools/<domain>.go` mirrors an `android/<domain>.go` — see [../ARCHITECTURE.md](../ARCHITECTURE.md).
- Add unit tests for any new pure logic (parsers, coordinate math, arg parsing).
- Open a [GitHub issue](https://github.com/iksnerd/adb_mcp/issues) for feedback, bugs, or tool requests.
