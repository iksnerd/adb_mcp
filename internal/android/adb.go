package android

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// runAdb runs an adb command (optionally targeting serial) and returns combined
// trimmed stdout. Stderr is folded into the error on failure.
func runAdb(ctx context.Context, serial string, args ...string) (string, error) {
	out, err := runAdbBytes(ctx, serial, args...)
	return strings.TrimRight(string(out), "\r\n"), err
}

// runAdbBytes is like runAdb but returns raw stdout bytes (used for screencap,
// where trimming or text conversion would corrupt the PNG).
func runAdbBytes(ctx context.Context, serial string, args ...string) ([]byte, error) {
	full := args
	if serial != "" {
		full = append([]string{"-s", serial}, args...)
	}
	cmd := exec.CommandContext(ctx, adbPath(), full...)
	cmd.Env = commandEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg != "" {
			return stdout.Bytes(), fmt.Errorf("adb %s: %s", strings.Join(full, " "), msg)
		}
		return stdout.Bytes(), fmt.Errorf("adb %s: %w", strings.Join(full, " "), err)
	}
	return stdout.Bytes(), nil
}
