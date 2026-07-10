//go:build windows

package android

import "os/exec"

// detach is a no-op on Windows; the emulator is launched without a controlling
// session by default.
func detach(cmd *exec.Cmd) {}
