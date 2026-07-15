package android

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// StatusBarOptions configures StatusBarDemo. Zero-value fields leave the
// corresponding status bar element unchanged (clock/battery) or fall back to
// the clean default (network/notifications) documented per field below.
type StatusBarOptions struct {
	// Clock is "HHMM" (e.g. "1200"); empty leaves it unchanged.
	Clock string
	// Battery is 0-100, or nil to leave it unchanged.
	Battery *int
	// NetworkType selects the network icon: "" or "wifi" (default, full
	// signal), "mobile" (signal/data-type/carrier below apply), or "none"
	// (hide both wifi and mobile icons).
	NetworkType string
	// MobileLevel is 0-4 signal bars, used only when NetworkType is "mobile".
	// Defaults to 4 (full signal) when nil.
	MobileLevel *int
	// DataType is the data-type icon shown next to the mobile signal, e.g.
	// "lte", "4g", "5g", "3g", "edge", "1x", "h", "h+", "roam". Used only
	// when NetworkType is "mobile".
	DataType string
	// Carrier is the operator name shown in the status bar. Used only when
	// NetworkType is "mobile".
	Carrier string
	// NotificationsVisible shows/hides notification icons. Defaults to false
	// (hidden, for clean screenshots) when nil.
	NotificationsVisible *bool
	// NotificationIcon is a best-effort AOSP SystemUI icon resource name shown
	// in the first notification slot. Support for this is an obscure,
	// version-dependent SystemUI internal (DemoStatusIcons) and may silently
	// no-op on some SDK images, unlike the other fields here.
	NotificationIcon string
}

// StatusBarDemo drives SystemUI "demo mode", which pins a clean, deterministic
// status bar — fixed clock, chosen signal/battery, no notification icons by
// default — so screenshots for docs don't leak the wall clock or a random
// signal state. enter=false exits demo mode and restores the live status bar
// (opts is ignored in that case).
//
// Requires the demo-mode broadcast to be allowed, which this enables first
// (settings global sysui_demo_allowed=1).
func StatusBarDemo(ctx context.Context, serial string, enter bool, opts StatusBarOptions) error {
	if _, err := runAdb(ctx, serial, "shell", "settings", "put", "global", "sysui_demo_allowed", "1"); err != nil {
		return err
	}
	if !enter {
		return demoBroadcast(ctx, serial, "exit")
	}
	if err := demoBroadcast(ctx, serial, "enter"); err != nil {
		return err
	}
	if opts.Clock != "" {
		if len(opts.Clock) != 4 || !isAllDigits(opts.Clock) {
			return fmt.Errorf("clock must be 4 digits HHMM (e.g. \"0930\"), got %q", opts.Clock)
		}
		if err := demoBroadcast(ctx, serial, "clock", "hhmm", opts.Clock); err != nil {
			return err
		}
	}
	if opts.Battery != nil {
		if *opts.Battery < 0 || *opts.Battery > 100 {
			return fmt.Errorf("battery must be 0-100, got %d", *opts.Battery)
		}
		if err := demoBroadcast(ctx, serial, "battery", "level", strconv.Itoa(*opts.Battery), "plugged", "false"); err != nil {
			return err
		}
	}
	if opts.MobileLevel != nil && (*opts.MobileLevel < 0 || *opts.MobileLevel > 4) {
		return fmt.Errorf("mobile_level must be 0-4, got %d", *opts.MobileLevel)
	}
	if err := networkBroadcast(ctx, serial, opts); err != nil {
		return err
	}
	visible := "false"
	if opts.NotificationsVisible != nil && *opts.NotificationsVisible {
		visible = "true"
	}
	kv := []string{"visible", visible}
	if opts.NotificationIcon != "" {
		kv = append(kv, "icon1", opts.NotificationIcon)
	}
	return demoBroadcast(ctx, serial, "notifications", kv...)
}

// networkBroadcast sends the "network" demo command for the selected
// NetworkType. Default ("" or "wifi") is full wifi signal; "mobile" applies
// signal level/data-type/carrier; "none" hides both network icons.
func networkBroadcast(ctx context.Context, serial string, opts StatusBarOptions) error {
	switch opts.NetworkType {
	case "", "wifi":
		return demoBroadcast(ctx, serial, "network", "wifi", "show", "level", "4")
	case "mobile":
		level := 4
		if opts.MobileLevel != nil {
			level = *opts.MobileLevel
		}
		kv := []string{"mobile", "show", "level", strconv.Itoa(level)}
		if opts.DataType != "" {
			kv = append(kv, "datatype", opts.DataType)
		}
		if opts.Carrier != "" {
			kv = append(kv, "carriername", opts.Carrier)
		}
		return demoBroadcast(ctx, serial, "network", kv...)
	case "none":
		return demoBroadcast(ctx, serial, "network", "wifi", "hide", "mobile", "hide")
	default:
		return fmt.Errorf("network_type must be \"wifi\", \"mobile\", or \"none\", got %q", opts.NetworkType)
	}
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
