# Tool reference

← Back to the [README](../README.md)

Every device-facing tool takes an optional `serial` (adb `-s`). Omit it when a
single device is attached; with several, pass one from `list_devices`.

### Emulator / device management
| Tool | Purpose |
|---|---|
| `list_avds` | List AVDs available to boot |
| `boot_emulator` | Boot an AVD (detached), wait for boot, return its serial |
| `list_devices` | List attached devices and adb state |
| `wait_for_boot` | Block until `sys.boot_completed=1` |
| `shutdown_emulator` | Power off (`adb emu kill`) |
| `connect_wireless` | Connect/pair a device over Wi-Fi (`adb connect`/`pair`) |

### Observe
| Tool | Purpose |
|---|---|
| `screenshot` | Capture the screen as a PNG (auto-downscaled) — to *see* state; retries an all-black frame and flags why (FLAG_SECURE / screen off) |
| `describe_ui` | UI hierarchy as elements with text/desc/id + true-pixel `center` — to *aim* |

### Interact
| Tool | Purpose |
|---|---|
| `tap` | Tap true-pixel `(x,y)` (use a `describe_ui` center) |
| `tap_on_text` | Find an element by label/desc and tap its center |
| `long_press` | Press and hold `(x,y)` for a duration |
| `wait_for_text` | Poll until a label appears, then return its tappable center |
| `swipe` | Swipe/drag (scroll down = high y → low y); `x`/`y` alias `x1`/`y1` |
| `drag` | Press-hold-move-release drag (`draganddrop`) — for drag handles & reorder |
| `input_text` | Type into the focused field via the IME |
| `press_key` | Press a named key (`enter`,`back`,`home`,`escape`,…) or raw keycode |
| `input_key_combo` | Press a chord together — `keys=["ctrl","a"]` or `preset="select_all"`/`copy`/`paste`/… |
| `enter_pin` | Enter digits on a PIN pad — with `grid`/`coords` for canvas-drawn pads |

### Device lock / Keystore
| Tool | Purpose |
|---|---|
| `set_device_lock` | Set a pin/pattern/password (needed for Keystore-backed crypto) |
| `clear_device_lock` | Remove the lock (supply the current credential) |
| `is_device_secure` | Whether a secure lock is set |

### App lifecycle
| Tool | Purpose |
|---|---|
| `list_packages` | List installed packages (filterable) |
| `get_app_details` | Version name/code + launchable activity of an app |
| `install_app` / `uninstall_app` | Install/reinstall or remove an app |
| `launch_app` / `stop_app` | Launch the LAUNCHER activity / force-stop |
| `reload_app` | Best-effort Metro/JS reload via the RN `RELOAD_APP_ACTION` broadcast |
| `open_dev_menu` | Open the RN dev menu (`KEYCODE_MENU`) when `reload_app` doesn't apply |
| `clear_app_data` | Wipe data+cache → first-launch state |
| `grant_permission` / `revoke_permission` | Grant/revoke a runtime permission |
| `open_url` | Open a URL or deep link (ACTION_VIEW) |
| `push_file` / `pull_file` | Copy files to/from the device |

### Logs & capture
| Tool | Purpose |
|---|---|
| `logcat` | One-shot dump of recent log lines, filterable by substring/`priority`/`tags` — find the native `Caused by:` |
| `start_logcat_capture` / `stop_logcat_capture` | Stream logs across a flow, then return them (same substring/`priority`/`tags` filters on stop) |
| `start_screen_record` / `stop_screen_record` | Record the screen to mp4 and pull it |

### Environment & diagnostics
| Tool | Purpose |
|---|---|
| `set_dark_mode` | Toggle the system dark theme |
| `set_location` | Set the mock GPS location |
| `set_status_bar` | Pin a clean status bar (SystemUI demo mode) — clock/battery/mobile signal+carrier+data-type/notifications — for tidy screenshots |
| `doctor` | Report SDK/adb/emulator/AVD/device health |

### Build & test (Gradle)
| Tool | Purpose |
|---|---|
| `gradle_build` | `./gradlew assembleDebug` (or a given task) → APK path |
| `run_unit_tests` | `./gradlew test` → pass/fail/skip summary, per-suite timing, failure stack traces; `json=true` for structured output |
| `run_instrumented_tests` | `./gradlew connectedAndroidTest` (needs a device) → same summary |
| `list_gradle_tasks` | Discover available Gradle tasks |

## Resources (the bundled "skill")

The driving know-how ships as MCP resources the client can list and read:

| URI | What it covers |
|---|---|
| `android://guide/getting-started` | Boot & connect, the `serial` argument, a first interaction |
| `android://guide/driving` | The observe→act loop, the true-pixel coordinate rule, tap-eating gotchas |
| `android://guide/pin-and-lock` | Native PIN pads and Keystore-required device locks |
| `android://guide/crash-triage` | Using `logcat` to find why a native call really failed |
