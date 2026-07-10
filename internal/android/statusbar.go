package android

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// StatusBarDemo drives SystemUI "demo mode", which pins a clean, deterministic
// status bar — fixed clock, full signal, chosen battery level, no notification
// icons — so screenshots for docs don't leak the wall clock or a random signal
// state. enter=false exits demo mode and restores the live status bar.
//
// clock is "HHMM" (e.g. "1200"); empty leaves it unchanged. battery is 0-100 or
// nil to leave it unchanged. Requires the demo-mode broadcast to be allowed,
// which this enables first (settings global sysui_demo_allowed=1).
func StatusBarDemo(ctx context.Context, serial string, enter bool, clock string, battery *int) error {
	if _, err := runAdb(ctx, serial, "shell", "settings", "put", "global", "sysui_demo_allowed", "1"); err != nil {
		return err
	}
	if !enter {
		return demoBroadcast(ctx, serial, "exit")
	}
	if err := demoBroadcast(ctx, serial, "enter"); err != nil {
		return err
	}
	if clock != "" {
		if len(clock) != 4 || !isAllDigits(clock) {
			return fmt.Errorf("clock must be 4 digits HHMM (e.g. \"0930\"), got %q", clock)
		}
		if err := demoBroadcast(ctx, serial, "clock", "hhmm", clock); err != nil {
			return err
		}
	}
	if battery != nil {
		if *battery < 0 || *battery > 100 {
			return fmt.Errorf("battery must be 0-100, got %d", *battery)
		}
		if err := demoBroadcast(ctx, serial, "battery", "level", strconv.Itoa(*battery), "plugged", "false"); err != nil {
			return err
		}
	}
	// Best-effort clean defaults: full wifi signal, no notification icons.
	_ = demoBroadcast(ctx, serial, "network", "wifi", "show", "level", "4")
	_ = demoBroadcast(ctx, serial, "notifications", "visible", "false")
	return nil
}

// demoBroadcast sends one SystemUI demo-mode command. Extra args are key/value
// pairs appended as `-e key value` (e.g. command=clock, hhmm=1200).
func demoBroadcast(ctx context.Context, serial, command string, kv ...string) error {
	args := []string{"shell", "am", "broadcast", "-a", "com.android.systemui.demo", "-e", "command", command}
	for i := 0; i+1 < len(kv); i += 2 {
		args = append(args, "-e", kv[i], kv[i+1])
	}
	_, err := runAdb(ctx, serial, args...)
	return err
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	return strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' }) == -1
}
