// Package tools binds the pure android execution layer to MCP tools. Each
// handler is a thin adapter: resolve the target device, call internal/android,
// and format the result. Keeping it thin means the real logic stays testable in
// internal/android without any MCP dependency.
//
// This file is the tool CATALOG. Handlers and their argument types live in
// domain files that mirror internal/android: emulator.go, observe.go,
// interact.go, lock.go, logs.go, apps.go, environment.go, gradle.go. Shared
// adapter helpers live in helpers.go.
package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Register adds every Android tool to the server.
//
// Tool descriptions are written for the model that will call them: each says
// what the tool does, when to reach for it, and the one gotcha that most often
// wastes a turn. The overarching workflow lives in the android://guide/*
// resources — read android://guide/driving for the observe→act loop.
func Register(s *mcp.Server) {
	// --- Emulator / device management ---
	add(s, "list_avds",
		"List the Android Virtual Devices (AVDs) installed on this machine that can be booted. Start here when no emulator is running yet; the returned names feed boot_emulator.",
		listAVDs)
	add(s, "boot_emulator",
		"Boot an AVD by name and return its device serial (e.g. emulator-5554). Launches the emulator detached so it outlives this call, and waits for full boot (sys.boot_completed) by default. Use the returned serial for later tools if you boot more than one device. Booting can take 30-120s on a cold start.",
		bootEmulator)
	add(s, "list_devices",
		"List attached emulators/devices and their adb state (device = ready, offline, unauthorized). Use it to confirm a device is up before driving it, or to get the serial when several are attached.",
		listDevices)
	add(s, "wait_for_boot",
		"Block until a device finishes booting (sys.boot_completed=1). Only needed if you started an emulator elsewhere; boot_emulator already waits by default.",
		waitForBoot)
	add(s, "shutdown_emulator",
		"Power off a running emulator (adb emu kill). Use when you are done with a device you booted.",
		shutdownEmulator)

	// --- Observe ---
	add(s, "screenshot",
		"Capture the current screen as a PNG so you can SEE the UI state. Call it after every action to confirm the screen changed before acting again — driving blind chains taps onto the wrong screen. The image is auto-downscaled (default max 760px) so it is accepted by the image reader; this is for seeing only — derive tap coordinates from describe_ui, not from this image.",
		screenshot)
	add(s, "describe_ui",
		"Read the on-screen UI hierarchy as a list of elements, each with its text, content_desc, resource_id, class, clickable flag, pixel bounds, and a precomputed center in TRUE DEVICE PIXELS. This is your source of truth for AIMING: pass an element's center straight to tap. Never guess coordinates from the screenshot (it is downscaled and you will miss). Pure-layout containers are filtered out to keep the list focused on actionable elements.",
		describeUI)

	// --- Interact ---
	add(s, "tap",
		"Tap a single coordinate in true device pixels. Use a center value from describe_ui. If a tap seems to do nothing, the coordinate is almost always stale/misscaled — re-run describe_ui and use a fresh center. Prefer tap_on_text when you know the element's label.",
		tap)
	add(s, "tap_on_text",
		"Find an element by its visible text or content-description and tap its center — the one-shot way to press a labelled button/row without computing coordinates yourself. Runs describe_ui internally and prefers a clickable match. Use exact match (partial=false) to avoid hitting the wrong item when labels overlap.",
		tapOnText)
	add(s, "swipe",
		"Swipe/drag from a start point to an end point. Params: x1,y1 (start) and x2,y2 (end) in true device pixels — x and y are accepted aliases for x1 and y1. To SCROLL DOWN a list, swipe from a HIGH y to a LOW y (drag the content up); reverse to scroll up. A longer duration_ms gives a slower, controlled drag; a short one flings.",
		swipe)
	add(s, "input_text",
		"Type text into the currently focused input field via the IME. Tap the field first so it has focus. Afterwards the soft keyboard may cover buttons lower on screen — dismiss it with press_key escape (or back) before tapping them. For native non-IME PIN pads this does nothing; use enter_pin instead.",
		inputText)
	add(s, "press_key",
		"Press a hardware/navigation key by name (enter, back, home, menu, tab, del, escape, up, down, left, right, dpad_center, app_switch, search, power, volume_up, volume_down, ...) or a raw Android keycode number. Handy to submit a form (enter), dismiss the keyboard (escape), or go back (back).",
		pressKey)
	add(s, "long_press",
		"Press and hold a coordinate (true device pixels) for a duration — for context menus, drag handles, and long-press actions.",
		longPress)
	add(s, "enter_pin",
		"Enter digits on a numeric PIN pad by tapping each key with a settle delay. Use when input_text does nothing because the pad renders its own key views. By default it finds each digit in the UI hierarchy — but CUSTOM-DRAWN pads (React Native / Skia SDK pads) draw their keys on a canvas that is invisible to describe_ui, so that lookup fails. For those, pass 'grid' (the pad's bounding box; digits are placed on a standard 3x4 dialpad) or 'coords' (explicit per-digit x,y). Read the pad's bounds/coordinates off a screenshot.",
		enterPIN)
	add(s, "wait_for_text",
		"Poll the UI until an element with the given text/content-description appears (or times out), then return it. Use this after an async action (network load, navigation, animation) instead of a blind wait-then-screenshot — it returns as soon as the element is present, with its tappable center. Note: canvas-drawn (RN/Skia) text never enters the hierarchy, so it will time out on those — screenshot instead.",
		waitForText)

	// --- Device lock / Keystore ---
	add(s, "set_device_lock",
		"Set a secure lock screen (type: pin [default], pattern, or password). REQUIRED before AndroidKeyStore / Keystore-backed crypto flows, which fail with 'A secure lock screen is required' on a fresh emulator that has no lock. Follow with is_device_secure to confirm.",
		setDeviceLock)
	add(s, "clear_device_lock",
		"Remove the secure lock screen, supplying the current credential as old_value. Use to restore a clean state after testing a Keystore flow.",
		clearDeviceLock)
	add(s, "is_device_secure",
		"Report whether a secure lock screen is set (KeyguardManager.isDeviceSecure). Use it to verify set_device_lock worked before running a Keystore-gated flow.",
		isDeviceSecure)

	// --- Logs & capture ---
	add(s, "logcat",
		"Dump recent native log lines (last N, default 400), optionally filtered by a case-insensitive substring. This is how you find the REAL reason a native call failed when the UI only shows a generic 'X failed' alert: filter by your app tag or 'Exception'/'Caused by' and read the 'Caused by:' line — that is the root cause. Dumps and exits (does not stream); chatty spam is stripped.",
		logcat)
	add(s, "start_logcat_capture",
		"Begin streaming logcat into a buffer for this device (optionally clearing first). Pair with stop_logcat_capture to get everything logged DURING a flow — use this instead of the one-shot 'logcat' when you need logs across an interaction.",
		startLogcatCapture)
	add(s, "stop_logcat_capture",
		"Stop the running logcat capture and return everything collected since start, optionally filtered (case-insensitive).",
		stopLogcatCapture)
	add(s, "start_screen_record",
		"Start recording the screen to an mp4 on the device (Android caps a single recording at ~180s). Pair with stop_screen_record.",
		startScreenRecord)
	add(s, "stop_screen_record",
		"Stop the screen recording, finalize the mp4, and pull it to a local path.",
		stopScreenRecord)

	// --- App lifecycle ---
	add(s, "list_packages",
		"List installed package names, optionally filtered by substring — to confirm an app is installed and get its exact package name for launch_app/stop_app.",
		listPackages)
	add(s, "install_app",
		"Install (or reinstall, -r) an APK from a local file path onto the device. Use to deploy a build you want to test.",
		installApp)
	add(s, "launch_app",
		"Launch an app by package name (starts its LAUNCHER activity). Combine with stop_app to restart an app cleanly from a known state.",
		launchApp)
	add(s, "stop_app",
		"Force-stop an app by package name. Pair with launch_app to reset an app to a clean start when reproducing a bug.",
		stopApp)
	add(s, "uninstall_app",
		"Uninstall an app by package name (adb uninstall). Use to remove a build before a clean install, or to verify first-run behavior after reinstalling. To keep the app but reset it, prefer clear_app_data.",
		uninstallApp)
	add(s, "clear_app_data",
		"Wipe an app's data and cache (pm clear) to reset it to a first-launch state — the fastest way to reproduce onboarding/permission flows from scratch.",
		clearAppData)
	add(s, "grant_permission",
		"Grant a runtime permission to an app (pm grant), e.g. android.permission.CAMERA — skips the in-app permission dialog so you can drive straight to the feature.",
		grantPermission)
	add(s, "revoke_permission",
		"Revoke a runtime permission from an app (pm revoke) — to test the denied path or re-trigger the permission-request dialog on next use. Pairs with grant_permission. Note: revoking some permissions kills the app process.",
		revokePermission)
	add(s, "open_url",
		"Open a URL or deep link via an ACTION_VIEW intent (am start) — the way to jump straight to a deep-linked screen. Optionally target a specific package.",
		openURL)
	add(s, "get_app_details",
		"Report an installed app's version name/code and its launchable activity (dumpsys package + resolve-activity) — to confirm what build is installed and find the activity to launch.",
		getAppDetails)
	add(s, "push_file",
		"Copy a local file onto the device (adb push) — e.g. seed test data or a file to import.",
		pushFile)
	add(s, "pull_file",
		"Copy a file off the device to a local path (adb pull) — e.g. retrieve a generated file, database, or screenshot.",
		pullFile)

	// --- Device discovery ---
	add(s, "connect_wireless",
		"Connect to a device over Wi-Fi/TCP (adb connect), optionally pairing first (adb pair) with the 6-digit code from Android 11+ Wireless debugging. Pass host:port; for pairing also pass the pairing address + code shown on the device.",
		connectWireless)

	// --- Environment ---
	add(s, "set_dark_mode",
		"Turn the system dark theme on or off (cmd uimode night) — to test light/dark appearances.",
		setDarkMode)
	add(s, "set_location",
		"Set the emulator's mock GPS location (longitude, latitude) — for location-gated features.",
		setLocation)
	add(s, "doctor",
		"Diagnose the local Android tooling: SDK path, adb/emulator availability, known AVDs, and attached devices. Run this first when something isn't working.",
		doctor)

	// --- Build & test (Gradle) ---
	add(s, "gradle_build",
		"Build the app with Gradle (default task assembleDebug) in project_dir, and report the produced APK path(s). project_dir must contain the Gradle wrapper (gradlew). Runs on the host, not a device.",
		gradleBuild)
	add(s, "run_unit_tests",
		"Run Gradle JVM unit tests (default task 'test') in project_dir and return the result summary.",
		runUnitTests)
	add(s, "run_instrumented_tests",
		"Run Gradle instrumented (on-device) tests (default task 'connectedAndroidTest') in project_dir — requires a booted device/emulator.",
		runInstrumentedTests)
	add(s, "list_gradle_tasks",
		"List the available Gradle tasks in project_dir (gradlew tasks) — to discover build/test/install targets.",
		listGradleTasks)
}

// add is a small generic wrapper over mcp.AddTool to cut boilerplate.
func add[In any](s *mcp.Server, name, desc string, h func(context.Context, In) (*mcp.CallToolResult, error)) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			res, err := h(ctx, in)
			return res, nil, err
		})
}
