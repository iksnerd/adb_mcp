# AndroidEmulatorMCP — TODO / Roadmap

The Android counterpart to [XcodeBuildMCP](https://github.com/getsentry/XcodeBuildMCP).
Tracks what's shipped and what's planned. Feedback/bugs also live in the
Council-Hub room `android-emulator-mcp-feedback`.

## Shipped

### v0.1.0 — core driving (21 tools + 4 guide resources)
- Emulator mgmt: `list_avds`, `boot_emulator`, `list_devices`, `wait_for_boot`, `shutdown_emulator`
- Observe: `screenshot` (auto-downscale), `describe_ui` (true-pixel centers)
- Interact: `tap`, `tap_on_text`, `swipe`, `input_text`, `press_key`, `enter_pin`
- Device lock: `set_device_lock`, `clear_device_lock`, `is_device_secure`
- Logs: `logcat`
- App: `list_packages`, `install_app`, `launch_app`, `stop_app`
- Resources: `android://guide/{getting-started,driving,pin-and-lock,crash-triage}`

### v0.2.0 — fixes from live E2E feedback
- [x] `boot_emulator`: launch emulator in its own session (`Setsid`) so it survives the server exit
- [x] `enter_pin`: `grid`/`coords` fallback for canvas-drawn (RN/Skia) pads invisible to uiautomator
- [x] `swipe`: accept `x`/`y` aliases for `x1`/`y1`; clearer missing-arg error
- [x] `describe_ui`: dump-twice-and-settle guard against stale mid-refresh trees
- [x] Guides: RN/Skia limitation + lock-then-restart + Google-Play-image notes
- [x] Install: `rm` + ad-hoc `codesign` in Makefile (fixes macOS "Killed: 9" SIGKILL)
- [x] Versioning: `-version` flag + build-time `-ldflags -X main.version` (git/VERSION)

### v0.3.0 — XcodeBuildMCP-parity tool batches (40 tools total)

**Batch A — app-lifecycle & interaction completeness**
- [x] `long_press` (input swipe same-point, long duration)
- [x] `uninstall_app` (`adb uninstall`)
- [x] `clear_app_data` (`pm clear`) — reset to clean state
- [x] `grant_permission` / `revoke_permission` (`pm grant`/`revoke`) — skip runtime dialogs
- [x] `open_url` (`am start -a android.intent.action.VIEW -d`) — deep links
- [x] `push_file` / `pull_file` (`adb push`/`pull`) — test data & artifacts

**Batch B — build & test (true XcodeBuildMCP parity)**
- [x] `gradle_build` (`./gradlew assembleDebug` → APK path)
- [x] `run_unit_tests` (`./gradlew test`)
- [x] `run_instrumented_tests` (`./gradlew connectedAndroidTest`)
- [x] `list_gradle_tasks` (`./gradlew tasks`)

**Batch C — environment & session polish**
- [x] `start_logcat_capture` / `stop_logcat_capture` (streaming session vs one-shot dump)
- [x] `set_dark_mode` (`cmd uimode night yes/no`)
- [x] `set_location` (`emu geo fix <lon> <lat>`)
- [x] `start_screen_record` / `stop_screen_record` (`screenrecord` → mp4 pull)
- [x] `doctor` (check adb/emulator/SDK/AVDs and report)

## Planned

## Backlog / ideas

### v0.4.0 — backlog cleared
- [x] `set_device_lock` optional `old_value` to change an existing lock in one call
- [x] `connect_wireless` (`adb pair`/`connect`) for wireless-adb devices
- [x] `get_app_details` (dumpsys package → version, launchable activity)
- [x] `wait_for_text` (poll describe_ui until a label appears — kills manual sleep+screenshot)
- [x] emulator factory reset — shipped as `boot_emulator {wipe_data:true}` (`-wipe-data`)

### Still open
- [ ] `swipe`/`drag` multi-touch & pinch gestures
- [ ] `set_status_bar` demo mode (clean screenshots for docs)
- [ ] `input_key_combo` (chorded keys)
- [ ] structured test-report parsing for `run_*_tests` (pass/fail counts)

## Notes
- Every device-facing tool takes an optional `serial`; single-device sessions can omit it.
- Keep `internal/android` pure/testable; `internal/tools` stays a thin MCP binding.
- Add unit tests for any new pure logic (parsers, coordinate math, arg parsing).
