# AndroidEmulatorMCP — MCP server for driving Android emulators

## Context

There is no MCP UI-automation server for Android the way [XcodeBuildMCP](https://github.com/getsentry/XcodeBuildMCP)
exists for iOS. Today the only way to drive an Android emulator from Claude is the
`android-emulator-drive` skill (`~/.claude/skills/android-emulator-drive/SKILL.md`), which is a
prose runbook of raw `adb`/`emulator` commands the model has to type by hand each time — slow, error-prone
(stale/misscaled tap coordinates, CRLF-corrupted screenshots, forgetting `exec-out`), and not reusable
across MCP clients.

This project turns that skill's hard-won knowledge into a proper **MCP server** (`AndroidEmulatorMCP`,
already scaffolded as a Go 1.26 module) so any MCP client (Claude Code, Cursor, etc.) gets first-class
Android tools: boot an AVD, screenshot, read the UI hierarchy as structured elements with real pixel
bounds, tap/swipe/type, set a device lock, read logcat, and manage app lifecycle. The server encodes the
skill's gotchas (use `exec-out`, downscale large PNGs, derive coordinates from `uiautomator` bounds) so the
model can't repeat them.

Local environment confirmed: `go1.26.0`, `adb` at `~/Library/Android/sdk/platform-tools/adb`,
`emulator` at `~/Library/Android/sdk/emulator/emulator`, AVDs available:
`Medium_Phone_API_36.1`, `Pixel_9a`, `Small_Phone`.

## Decisions (confirmed with user)

- **SDK:** official `github.com/modelcontextprotocol/go-sdk` v1.6.1 (already added to `go.mod`), stdio transport.
- **Tool set:** full parity with the skill + convenience helpers (`tap_on_text`, `enter_pin`).
- **Ship with:** `.mcp.json` config snippet, README with per-tool docs, Go unit tests for pure logic, Makefile.

## Architecture

```
AndroidEmulatorMCP/
  main.go                     # entry: build server, register all tools, Run(stdio)
  internal/android/
    sdk.go                    # locate Android SDK, build PATH, resolve adb/emulator binaries
    device.go                 # run adb (with optional -s serial), run emulator, screencap, logcat, lock
    uiauto.go                 # uiautomator dump + XML parse -> []Element{Text,Desc,Class,Bounds,Clickable}
    uiauto_test.go            # parse tests (fixture XML) — no emulator needed
    keyevent.go               # named-key -> keycode map (enter/back/home/...)
    keyevent_test.go          # mapping tests
    image.go                  # downscale large PNGs (sips or pure-Go image/png resize)
  internal/tools/
    register.go               # AddTool(...) for every tool; each handler calls internal/android
  README.md
  Makefile
  .mcp.json                   # sample Claude Code registration
```

Rationale: `internal/android` is the pure execution/parse layer (unit-testable, no MCP deps);
`internal/tools` is the thin MCP binding. Keeps SDK surface isolated so tool logic stays testable.

## SDK API (verified against v1.6.1 source)

- `srv := mcp.NewServer(&mcp.Implementation{Name:"android-emulator-mcp", Version:"0.1.0"}, nil)`
- `mcp.AddTool(srv, &mcp.Tool{Name, Description}, handler)` where
  `handler func(ctx, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)` — input JSON schema is
  auto-generated from the `In` struct; struct tags (`json:"..."` + `jsonschema:"..."` for descriptions) drive it.
- Return unstructured content via `&mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text:...}}}`.
- Screenshots: `&mcp.ImageContent{Data: pngBytes, MIMEType: "image/png"}` (SDK base64-encodes `Data`).
- Errors: return a normal `error`; SDK sets `IsError` + text content so the model can self-correct.
- `srv.Run(ctx, &mcp.StdioTransport{})`.

## Tools (full set)

Every tool takes an optional `serial` arg (adb `-s`); when omitted and exactly one device is attached, use it,
else return a clear "specify serial" error.

**Emulator / device management**
- `list_avds` — `emulator -list-avds`.
- `boot_emulator` {avd, wait_for_boot=true, no_snapshot=true} — launch detached, poll `getprop sys.boot_completed`.
- `list_devices` — parse `adb devices` → serial + state.
- `wait_for_boot` {serial, timeout_s=120} — poll boot_completed.
- `shutdown_emulator` {serial} — `adb -s <serial> emu kill`.

**Observe**
- `screenshot` {serial, max_dim=760} — `adb exec-out screencap -p`, downscale if larger than max_dim, return ImageContent.
- `describe_ui` {serial} — `uiautomator dump` → parsed elements (text, content-desc, class, clickable,
  bounds `[x1,y1][x2,y2]`, computed center) as structured JSON. Retries once on
  "could not get idle state". This is the model's source of truth for tap coordinates.

**Interact**
- `tap` {serial, x, y}
- `tap_on_text` {serial, text, partial=true} — describe_ui, find element whose text/content-desc matches, tap its center.
- `swipe` {serial, x1, y1, x2, y2, duration_ms=300}
- `input_text` {serial, text} — via `input text` (documents the space/`%s` caveat).
- `press_key` {serial, key} — named key → keycode (enter, back, home, menu, tab, del, escape, …) or raw int.
- `enter_pin` {serial, digits} — reads pad layout via describe_ui once, taps digit-by-digit with settle delay.

**Device lock**
- `set_device_lock` {serial, type=pin, value} — `locksettings set-pin|set-pattern|set-password`.
- `clear_device_lock` {serial, old_value} — `locksettings clear --old <value>`.
- `is_device_secure` {serial} — `locksettings get-disabled` → bool.

**Logs**
- `logcat` {serial, lines=400, filter} — `adb logcat -d -t <lines>`, optional case-insensitive grep, strips `chatty`.

**App lifecycle**
- `list_packages` {serial, filter} — `pm list packages`.
- `install_app` {serial, apk_path} — `adb install -r`.
- `launch_app` {serial, package} — `monkey -p <pkg> -c android.intent.category.LAUNCHER 1`.
- `stop_app` {serial, package} — `am force-stop`.

## Key implementation notes (encode the skill's gotchas)

- Always `adb exec-out screencap -p` (never `shell screencap`) to avoid CRLF corruption.
- Coordinates from `uiautomator` bounds are true device pixels — `describe_ui` returns centers precomputed so
  the model never guesses from the downscaled image.
- Downscale screenshots > `max_dim` before returning (large PNGs get rejected by image readers).
- `describe_ui` retries on transient "could not get idle state" (animation in progress).
- Resolve the SDK via `ANDROID_HOME`/`ANDROID_SDK_ROOT` env, else default `~/Library/Android/sdk`; prepend
  `platform-tools` + `emulator` to the command's PATH so `adb`/`emulator` resolve regardless of caller env.

## Files to create / modify

- Replace placeholder `main.go`.
- New: `internal/android/{sdk,device,uiauto,keyevent,image}.go` + `{uiauto,keyevent}_test.go`.
- New: `internal/tools/register.go`.
- New: `README.md`, `Makefile`, `.mcp.json`.
- `go.mod` already has the SDK dep; `go mod tidy` will pin `go.sum`.

## Verification

1. `make build` (or `go build ./...`) compiles cleanly; `go test ./...` passes the parse/keymap unit tests.
2. `go vet ./...` clean.
3. Live smoke test end-to-end against a real AVD:
   - Register the built binary in this Claude Code session via `.mcp.json` (or `claude mcp add`).
   - `list_avds` → shows the three AVDs.
   - `boot_emulator{avd:"Pixel_9a"}` → device reaches boot_completed; `list_devices` shows it.
   - `screenshot` → returns a viewable image of the home screen.
   - `describe_ui` → returns elements with bounds; `tap_on_text` on a visible label changes the screen
     (confirm via a follow-up `screenshot`).
   - `set_device_lock{value:"1234"}` then `is_device_secure` → true; `clear_device_lock{old_value:"1234"}`.
   - `logcat{filter:"ActivityManager"}` → returns recent lines.
4. Sanity-check raw MCP handshake without a client: pipe an `initialize` + `tools/list` JSON-RPC request into
   the binary over stdio and confirm all tools are advertised.
```
```

## Out of scope (first version)

- Physical USB devices over adb (should mostly work but untested), wireless adb pairing, app build/gradle,
  video recording, and Play Store automation. Easy to add later as new tools.
