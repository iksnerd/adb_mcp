# Changelog

Shipped work, newest first. Roadmap and open ideas live in
[BACKLOG.md](BACKLOG.md); the code layout is described in
[../ARCHITECTURE.md](../ARCHITECTURE.md).

## v0.5.0 — architecture split, bug-fix pass, gesture/status/report tools (46 tools)

**Refactor — world-class layout**
- Split the two monolith files (`tools/register.go` 997 lines, `android/device.go` 552 lines) into a domain-mirrored layout: each `tools/<domain>.go` adapter maps to an `android/<domain>.go` execution file, and `register.go` is now just the tool catalog. See [ARCHITECTURE.md](../ARCHITECTURE.md) and [architecture.mmd](architecture.mmd).

**Bug fixes (from code review)**
- `screenshot`: `max_dim:0` now actually disables downscaling (was silently remapped to the 760 default, making the documented full-res path unreachable). Arg is `*int`: omit → 760, `0`/negative → full resolution.
- `commandEnv`: match the `PATH` key case-insensitively and preserve its value — the old uppercase-only check appended a duplicate `PATH=` on Windows (case-insensitive keys), clobbering the system path.
- `boot_emulator`: honor the caller's `timeout_s` for the boot-wait phase instead of passing `WaitForBoot` a `<=0` remainder, which it silently treated as its own 120s default.
- `input_text`: single-quote the argument for the device shell so `$`, backtick, quotes, and other metacharacters type literally (the old escaper handled only a subset). Unit-tested.
- Capture sessions: `StopAllCaptures()` drains running logcat/screen-record sessions on shutdown so detached adb processes and temp files don't leak.
- `downscalePNG`: average/store in a consistent premultiplied color space (correctness for translucent PNGs; no change for opaque screenshots).

**New tools**
- `drag` — press-hold-move-release drag (`input draganddrop`, Android 11+), distinct from the fling of `swipe`.
- `input_key_combo` — chorded keys (`input keycombination`, Android 11+), e.g. `["ctrl","a"]`; added modifier + a-z key names.
- `set_status_bar` — SystemUI demo mode for clean doc screenshots (fixed clock, full signal, chosen battery, no notification icons).
- `run_unit_tests` / `run_instrumented_tests` now parse the JUnit XML and report a structured pass/fail/error/skipped summary with the failing tests — on both success and failure.

## v0.4.0 — backlog cleared

- `set_device_lock` optional `old_value` to change an existing lock in one call
- `connect_wireless` (`adb pair`/`connect`) for wireless-adb devices
- `get_app_details` (dumpsys package → version, launchable activity)
- `wait_for_text` (poll describe_ui until a label appears — kills manual sleep+screenshot)
- emulator factory reset — shipped as `boot_emulator {wipe_data:true}` (`-wipe-data`)

## v0.3.0 — XcodeBuildMCP-parity tool batches (40 tools total)

**Batch A — app-lifecycle & interaction completeness**
- `long_press` (input swipe same-point, long duration)
- `uninstall_app` (`adb uninstall`)
- `clear_app_data` (`pm clear`) — reset to clean state
- `grant_permission` / `revoke_permission` (`pm grant`/`revoke`) — skip runtime dialogs
- `open_url` (`am start -a android.intent.action.VIEW -d`) — deep links
- `push_file` / `pull_file` (`adb push`/`pull`) — test data & artifacts

**Batch B — build & test (true XcodeBuildMCP parity)**
- `gradle_build` (`./gradlew assembleDebug` → APK path)
- `run_unit_tests` (`./gradlew test`)
- `run_instrumented_tests` (`./gradlew connectedAndroidTest`)
- `list_gradle_tasks` (`./gradlew tasks`)

**Batch C — environment & session polish**
- `start_logcat_capture` / `stop_logcat_capture` (streaming session vs one-shot dump)
- `set_dark_mode` (`cmd uimode night yes/no`)
- `set_location` (`emu geo fix <lon> <lat>`)
- `start_screen_record` / `stop_screen_record` (`screenrecord` → mp4 pull)
- `doctor` (check adb/emulator/SDK/AVDs and report)

## v0.2.0 — fixes from live E2E feedback

- `boot_emulator`: launch emulator in its own session (`Setsid`) so it survives the server exit
- `enter_pin`: `grid`/`coords` fallback for canvas-drawn (RN/Skia) pads invisible to uiautomator
- `swipe`: accept `x`/`y` aliases for `x1`/`y1`; clearer missing-arg error
- `describe_ui`: dump-twice-and-settle guard against stale mid-refresh trees
- Guides: RN/Skia limitation + lock-then-restart + Google-Play-image notes
- Install: `rm` + ad-hoc `codesign` in Makefile (fixes macOS "Killed: 9" SIGKILL)
- Versioning: `-version` flag + build-time `-ldflags -X main.version` (git/VERSION)

## v0.1.0 — core driving (21 tools + 4 guide resources)

- Emulator mgmt: `list_avds`, `boot_emulator`, `list_devices`, `wait_for_boot`, `shutdown_emulator`
- Observe: `screenshot` (auto-downscale), `describe_ui` (true-pixel centers)
- Interact: `tap`, `tap_on_text`, `swipe`, `input_text`, `press_key`, `enter_pin`
- Device lock: `set_device_lock`, `clear_device_lock`, `is_device_secure`
- Logs: `logcat`
- App: `list_packages`, `install_app`, `launch_app`, `stop_app`
- Resources: `android://guide/{getting-started,driving,pin-and-lock,crash-triage}`
