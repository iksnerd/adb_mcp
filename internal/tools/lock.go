package tools

import (
	"context"

	"github.com/iksnerd/adb_mcp/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type setLockArgs struct {
	serialArg
	Type     string `json:"type,omitempty" jsonschema:"Lock type: pin (default), pattern, or password."`
	Value    string `json:"value" jsonschema:"Credential to set, e.g. \"1234\"."`
	OldValue string `json:"old_value,omitempty" jsonschema:"The CURRENT credential, required only when a lock is already set and you want to change it (locksettings refuses to overwrite otherwise)."`
}

type clearLockArgs struct {
	serialArg
	OldValue string `json:"old_value" jsonschema:"The current credential, needed to clear the lock."`
}

// ---- Handlers ----

func setDeviceLock(ctx context.Context, in setLockArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.SetDeviceLock(ctx, serial, in.Type, in.Value, in.OldValue); err != nil {
		return nil, err
	}
	return text("Lock screen set on %s.", serial), nil
}

func clearDeviceLock(ctx context.Context, in clearLockArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.ClearDeviceLock(ctx, serial, in.OldValue); err != nil {
		return nil, err
	}
	return text("Lock screen cleared on %s.", serial), nil
}

func isDeviceSecure(ctx context.Context, in serialArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	secure, err := android.IsDeviceSecure(ctx, serial)
	if err != nil {
		return nil, err
	}
	return text("Device secure: %v", secure), nil
}
