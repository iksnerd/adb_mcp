package tools

import (
	"context"

	"AndroidEmulatorMCP/internal/android"

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

func doctor(ctx context.Context, _ doctorArgs) (*mcp.CallToolResult, error) {
	return text("%s", android.Doctor(ctx)), nil
}
