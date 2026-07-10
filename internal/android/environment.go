package android

import (
	"context"
	"strconv"
)

// SetDarkMode toggles the system dark theme (cmd uimode night yes|no).
func SetDarkMode(ctx context.Context, serial string, on bool) error {
	mode := "no"
	if on {
		mode = "yes"
	}
	_, err := runAdb(ctx, serial, "shell", "cmd", "uimode", "night", mode)
	return err
}

// SetLocation sets the emulator's GPS fix (longitude, latitude), via the console
// `geo fix` command. Note adb's `emu geo fix` takes longitude first.
func SetLocation(ctx context.Context, serial string, lon, lat float64) error {
	_, err := runAdb(ctx, serial, "emu", "geo", "fix",
		strconv.FormatFloat(lon, 'f', -1, 64), strconv.FormatFloat(lat, 'f', -1, 64))
	return err
}
