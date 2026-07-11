package tools

import (
	"context"

	"github.com/iksnerd/adb_mcp/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type darkModeArgs struct {
	serialArg
	Enabled bool `json:"enabled" jsonschema:"true = dark theme on, false = off."`
}

type locationArgs struct {
	serialArg
	Longitude float64 `json:"longitude" jsonschema:"Longitude of the mock GPS fix."`
	Latitude  float64 `json:"latitude" jsonschema:"Latitude of the mock GPS fix."`
}

type statusBarArgs struct {
	serialArg
	Enabled bool   `json:"enabled" jsonschema:"true = enter demo mode (clean, fixed status bar); false = exit and restore the live status bar."`
	Clock   string `json:"clock,omitempty" jsonschema:"Fixed clock as 4 digits HHMM, e.g. \"1200\". Only when enabled=true; omit to leave unchanged."`
	Battery *int   `json:"battery,omitempty" jsonschema:"Fixed battery level 0-100 (shown unplugged). Only when enabled=true; omit to leave unchanged."`
}

type doctorArgs struct{}

// ---- Handlers ----

func setDarkMode(ctx context.Context, in darkModeArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.SetDarkMode(ctx, serial, in.Enabled); err != nil {
		return nil, err
	}
	return text("Dark mode: %v", in.Enabled), nil
}

func setLocation(ctx context.Context, in locationArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.SetLocation(ctx, serial, in.Longitude, in.Latitude); err != nil {
		return nil, err
	}
	return text("Location set to (lon %v, lat %v).", in.Longitude, in.Latitude), nil
}

func setStatusBar(ctx context.Context, in statusBarArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.StatusBarDemo(ctx, serial, in.Enabled, in.Clock, in.Battery); err != nil {
		return nil, err
	}
	if in.Enabled {
		return text("Status bar demo mode on for %s (clean bar for screenshots). Call again with enabled=false to restore.", serial), nil
	}
	return text("Status bar demo mode off for %s (live bar restored).", serial), nil
}

func doctor(ctx context.Context, _ doctorArgs) (*mcp.CallToolResult, error) {
	return text("%s", android.Doctor(ctx)), nil
}
