// Package android is the pure execution/parse layer that wraps the adb and
// emulator command-line tools. It has no dependency on the MCP SDK so its logic
// stays unit-testable. The internal/tools package binds these functions to MCP
// tools.
package android

import (
	"os"
	"path/filepath"
	"runtime"
)

// sdkRoot resolves the Android SDK location, checking the standard environment
// variables first and falling back to the per-platform default install path.
func sdkRoot() string {
	for _, env := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Android", "sdk")
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Android", "Sdk")
	default: // linux
		return filepath.Join(home, "Android", "Sdk")
	}
}

// adbPath returns the path to the adb binary. If it cannot be found under the
// resolved SDK root, it returns the bare name so a PATH lookup can still work.
func adbPath() string {
	if root := sdkRoot(); root != "" {
		p := filepath.Join(root, "platform-tools", exe("adb"))
		if fileExists(p) {
			return p
		}
	}
	return exe("adb")
}

// emulatorPath returns the path to the emulator binary, mirroring adbPath.
func emulatorPath() string {
	if root := sdkRoot(); root != "" {
		p := filepath.Join(root, "emulator", exe("emulator"))
		if fileExists(p) {
			return p
		}
	}
	return exe("emulator")
}

// commandEnv returns an environment with the SDK tool directories prepended to
// PATH, so adb/emulator resolve their own helper binaries regardless of the
// caller's environment.
func commandEnv() []string {
	env := os.Environ()
	root := sdkRoot()
	if root == "" {
		return env
	}
	extra := filepath.Join(root, "platform-tools") + string(os.PathListSeparator) +
		filepath.Join(root, "emulator")
	for i, kv := range env {
		if len(kv) >= 5 && kv[:5] == "PATH=" {
			env[i] = "PATH=" + extra + string(os.PathListSeparator) + kv[5:]
			return env
		}
	}
	return append(env, "PATH="+extra)
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
