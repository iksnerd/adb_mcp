# Backlog & ideas

Open, unstarted work. Shipped history is in [CHANGELOG.md](CHANGELOG.md).

## XcodeBuildMCP parity gaps

Core driving/build/test/automate is at parity (and ahead on screen recording,
device-lock/Keystore, custom PIN pads, `tap_on_text`/`wait_for_text`,
`set_status_bar`). These are the remaining gaps vs XcodeBuildMCP:

- [ ] **`build_and_run`** — one-shot `gradle_build` → `install_app` → `launch_app` (optionally on a chosen variant). Currently composable from three calls; XcodeBuildMCP exposes it as a single tool. Highest value / lowest lift.
- [ ] **Deeper project discovery** — no analogue of "list schemes / dump build settings". Add `list_gradle_variants` (parse `./gradlew tasks`/`app:properties` or the `assemble*` task list) and a module/build-info dump, complementing `list_gradle_tasks` + `get_app_details`.
- [ ] **Project scaffolding** — no "create a new Android project from a template" tool (XcodeBuildMCP has `scaffold`). Biggest lift; would need a bundled template + Gradle wrapper generation.
- [ ] **Embedded runtime-crash telemetry (`last_crash`)** — XcodeBuildMCP bundles a lib that streams structured runtime errors; the Android analogue is `logcat` "Caused by:" triage. Concrete proposal from field feedback below: pull `adb shell dumpsys dropbox --print data_app_crash` (or a tombstone for native crashes) so the FATAL EXCEPTION header + message + frames come back together in one call, instead of being reconstructed from a filtered/spilled logcat capture.

## Enhancements

- [ ] **Multi-touch / pinch-zoom gestures.** The single-pointer half shipped as `drag` (`input draganddrop`). True two-finger pinch/rotate needs the `sendevent` multi-touch protocol, which is device/kernel-specific (the `input` command has no multi-pointer verb) — parked until there's a reliable cross-device approach.

## Field feedback (a partner app debugging sessions, 2026-07-15)

From council-hub `android-emulator-mcp-feedback` — real friction driving a
React Native/Expo dev-client app across two long debugging sessions. The
reporter flagged the first two as costing the most back-and-forth.

- [ ] **Bound `stop_logcat_capture` output by default.** Every capture in these sessions exceeded the token limit and force-spilled to a file, even with a filter (a broad term like "crash" still matches Monkey/WindowManager noise). Default to a tail cap (~500 lines) or return `{summary + file path + top matches}`.
- [ ] **Clearer `launch_app` failure output.** A missing LAUNCHER activity (wrong package / not installed) currently returns the raw `monkey` arg-dump with no clear message. Detect `No activities found to run, monkey aborted` and surface a plain error; on success, echo the resolved component.
- [ ] **`launch_app` dev-client awareness.** For Expo/RN dev builds, `launch_app` lands on the Dev Launcher (which then needs Metro), and neither the Dev Launcher UI nor a native PIN pad is visible to `describe_ui`/`tap_on_text`. Option to launch a dev-server URL directly (deep link), plus docs on which surfaces are drivable vs opaque.
- [ ] **`logcat` buffer-rotation hint.** A one-shot dump of an already-rotated-out crash silently returns "(no matching log lines)". A hint to use `start_logcat_capture`/`stop_logcat_capture` instead would save a cycle.
- [ ] **`screenshot`/`describe_ui` state-skew note.** The two can briefly disagree during app transitions (backgrounding, Dev Launcher hand-off) — looked like a timing race. Worth a short settle delay or a documented caveat.

## Conventions (read before adding a tool)

- Every device-facing tool takes an optional `serial`; single-device sessions can omit it.
- Keep `internal/android` pure/testable; `internal/tools` stays a thin MCP binding. Each `tools/<domain>.go` mirrors an `android/<domain>.go` — see [../ARCHITECTURE.md](../ARCHITECTURE.md).
- Add unit tests for any new pure logic (parsers, coordinate math, arg parsing).
- Open a [GitHub issue](https://github.com/iksnerd/adb_mcp/issues) for feedback, bugs, or tool requests.
