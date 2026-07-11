package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iksnerd/adb_mcp/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type listAVDsArgs struct{}

type bootArgs struct {
	AVD         string `json:"avd" jsonschema:"AVD name to boot (see list_avds)."`
	WaitForBoot *bool  `json:"wait_for_boot,omitempty" jsonschema:"Wait until fully booted before returning. Default true."`
	NoSnapshot  *bool  `json:"no_snapshot,omitempty" jsonschema:"Cold boot without loading a saved snapshot. Default true."`
	WipeData    *bool  `json:"wipe_data,omitempty" jsonschema:"Factory-reset the AVD on this boot (-wipe-data). Default false. Use to start from a pristine device."`
	TimeoutS    int    `json:"timeout_s,omitempty" jsonschema:"Boot timeout in seconds. Default 180."`
}

type waitBootArgs struct {
	serialArg
	TimeoutS int `json:"timeout_s,omitempty" jsonschema:"Timeout in seconds. Default 120."`
}

type connectWirelessArgs struct {
	HostPort    string `json:"host_port" jsonschema:"Device address to connect to, host:port (e.g. 192.168.1.42:5555)."`
	PairAddress string `json:"pair_address,omitempty" jsonschema:"Pairing address host:port (Android 11+ Wireless debugging), if different from host_port. Only needed when pairing."`
	PairingCode string `json:"pairing_code,omitempty" jsonschema:"6-digit pairing code shown on the device. Provide to pair before connecting."`
}

// ---- Handlers ----

func listAVDs(ctx context.Context, _ listAVDsArgs) (*mcp.CallToolResult, error) {
	avds, err := android.ListAVDs(ctx)
	if err != nil {
		return nil, err
	}
	if len(avds) == 0 {
		return text("No AVDs found. Create one in Android Studio's Device Manager."), nil
	}
	return text("Available AVDs:\n%s", strings.Join(avds, "\n")), nil
}

func bootEmulator(ctx context.Context, in bootArgs) (*mcp.CallToolResult, error) {
	if strings.TrimSpace(in.AVD) == "" {
		return nil, fmt.Errorf("avd is required")
	}
	timeout := 180 * time.Second
	if in.TimeoutS > 0 {
		timeout = time.Duration(in.TimeoutS) * time.Second
	}
	serial, err := android.BootEmulator(ctx, in.AVD, boolOr(in.NoSnapshot, true), boolOr(in.WaitForBoot, true), boolOr(in.WipeData, false), timeout)
	if err != nil {
		if serial != "" {
			return nil, fmt.Errorf("emulator %s came up as %s but %w", in.AVD, serial, err)
		}
		return nil, err
	}
	return text("Booted %q as %s.", in.AVD, serial), nil
}

func listDevices(ctx context.Context, _ serialArg) (*mcp.CallToolResult, error) {
	devices, err := android.ListDevices(ctx)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return text("No devices attached."), nil
	}
	return jsonResult(devices)
}

func waitForBoot(ctx context.Context, in waitBootArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(in.TimeoutS) * time.Second
	if err := android.WaitForBoot(ctx, serial, timeout); err != nil {
		return nil, err
	}
	return text("%s is booted.", serial), nil
}

func shutdownEmulator(ctx context.Context, in serialArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.Shutdown(ctx, serial); err != nil {
		return nil, err
	}
	return text("Shutdown requested for %s.", serial), nil
}

func connectWireless(ctx context.Context, in connectWirelessArgs) (*mcp.CallToolResult, error) {
	out, err := android.ConnectWireless(ctx, in.HostPort, in.PairAddress, in.PairingCode)
	if err != nil {
		return nil, fmt.Errorf("%v\n%s", err, out)
	}
	return text("%s", out), nil
}
