# AndroidEmulatorMCP

An [MCP](https://modelcontextprotocol.io) server that lets an AI agent drive an
Android emulator or device over `adb` — boot an AVD, screenshot, read the UI
hierarchy, tap/swipe/type, set a device lock, read `logcat`, and manage app
lifecycle.

It is the Android counterpart to [XcodeBuildMCP](https://github.com/getsentry/XcodeBuildMCP):
where XcodeBuildMCP gives an agent first-class control of iOS simulators, this
gives the same for Android emulators. Built on the official
[Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk), speaks stdio.

## Why

Driving Android by hand means a long runbook of raw `adb` commands, and it is
easy to get wrong (stale tap coordinates, CRLF-corrupted screenshots, forgetting
`exec-out`, guessing coordinates off a downscaled image). This server encodes
that hard-won knowledge as tools so the agent can't repeat the mistakes:

- Screenshots use `exec-out screencap` (no CRLF corruption) and are auto-downscaled
  so the image reader accepts them.
- `describe_ui` returns each element's **center in true device pixels**, so taps
  land where you mean them to — no guessing off the image.
- `describe_ui` retries the transient "could not get idle state" failure.

The workflow itself is bundled as readable **resources** (see below), so the
agent can consult the "skill" the same way it would read a skill file.

## Requirements

- Go 1.26+ (to build)
- Android SDK with `platform-tools` (`adb`) and `emulator`. The server finds it
  via `$ANDROID_HOME` / `$ANDROID_SDK_ROOT`, else the platform default
  (`~/Library/Android/sdk` on macOS).
- At least one AVD (create one in Android Studio's Device Manager).

## Build & install

```bash
make install                 # builds ./bin/android-emulator-mcp and copies it to ~/.local/bin
# or:
go build -o bin/android-emulator-mcp .
```

## Register with an MCP client

**Claude Code** (from this repo, the bundled `.mcp.json` is picked up automatically),
or explicitly:

```bash
claude mcp add android -- android-emulator-mcp
```

Any other client: run `android-emulator-mcp` over stdio. Example config:

```json
{
  "mcpServers": {
    "android": { "command": "android-emulator-mcp" }
  }
}
```

## Tools

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
| `screenshot` | Capture the screen as a PNG (auto-downscaled) — to *see* state |
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
| `input_key_combo` | Press a chord together, e.g. `["ctrl","a"]`, `["alt","tab"]` |
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
| `clear_app_data` | Wipe data+cache → first-launch state |
| `grant_permission` / `revoke_permission` | Grant/revoke a runtime permission |
| `open_url` | Open a URL or deep link (ACTION_VIEW) |
| `push_file` / `pull_file` | Copy files to/from the device |

### Logs & capture
| Tool | Purpose |
|---|---|
| `logcat` | One-shot dump of recent log lines, filterable — find the native `Caused by:` |
| `start_logcat_capture` / `stop_logcat_capture` | Stream logs across a flow, then return them |
| `start_screen_record` / `stop_screen_record` | Record the screen to mp4 and pull it |

### Environment & diagnostics
| Tool | Purpose |
|---|---|
| `set_dark_mode` | Toggle the system dark theme |
| `set_location` | Set the mock GPS location |
| `set_status_bar` | Pin a clean status bar (SystemUI demo mode) for tidy screenshots |
| `doctor` | Report SDK/adb/emulator/AVD/device health |

### Build & test (Gradle)
| Tool | Purpose |
|---|---|
| `gradle_build` | `./gradlew assembleDebug` (or a given task) → APK path |
| `run_unit_tests` | `./gradlew test` → structured pass/fail/skip summary + failing tests |
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

## The core loop

**observe → locate → act → re-observe.** `screenshot` to see, `describe_ui` to
get true-pixel centers, `tap`/`tap_on_text`/`swipe`/`input_text` to act, then
`screenshot` again to confirm. Read `android://guide/driving` for the full loop
and the gotchas that waste turns.

## Development

```bash
make check     # go vet + go test (unit tests need no emulator)
make run       # run over stdio for manual JSON-RPC poking
```

Layout:

```
main.go                    entry: build server, register tools + resources, Run(stdio)
internal/android/          pure adb/emulator execution + uiautomator parsing (unit-tested)
internal/tools/            thin MCP tool bindings
internal/guides/           the skill guides, embedded and served as MCP resources
```

The two layers follow a **mirror convention** — each `internal/tools/<domain>.go`
adapter maps to an `internal/android/<domain>.go` execution file of the same
name. See [ARCHITECTURE.md](ARCHITECTURE.md) (with a Mermaid diagram) for the
full map and the rules for adding a tool.
