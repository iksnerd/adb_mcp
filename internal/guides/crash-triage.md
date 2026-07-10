# Finding out WHY a native call failed

A JS-level or UI-level "X failed" alert almost always hides the real cause. The
truth is in the native log. Use `logcat`.

## Workflow

1. Reproduce the failure.
2. Call `logcat` with a `filter` — a substring that narrows to the relevant
   lines. Good filters: your app's log tag, `Exception`, `Caused by`,
   `keystore`, `StrongBox`, the failing class name.
3. Read from the bottom up. The **`Caused by:`** line is the root error — that is
   the one that matters, not the top-level wrapper exception.

`logcat` dumps the last N lines (default 400) and exits; it does not stream. It
also drops `chatty` dedup spam automatically. Increase `lines` if the stack
trace is long and gets truncated.

## Things the root cause often reveals

- An obfuscated/renamed class breaking JNA/JNI struct layout at runtime.
- `No StrongBox available` → the device fell back from hardware-backed to
  software keys (relevant to Keystore attestation flows).
- A missing permission, a `ClassNotFoundException`, or a native `.so` that failed
  to load.

## App lifecycle while triaging

- `list_packages` with a filter confirms the app is installed and gives its exact
  package name.
- `stop_app` then `launch_app` restarts it cleanly to reproduce from a known
  state.
- `install_app` (re)installs an APK build you want to test.
