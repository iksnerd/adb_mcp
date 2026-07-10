package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"AndroidEmulatorMCP/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serialArg is embedded by every device-targeted tool's arguments.
type serialArg struct {
	Serial string `json:"serial,omitempty" jsonschema:"Target device serial (adb -s). Optional when exactly one device is attached."`
}

// text formats a plain-text tool result.
func text(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}}
}

// jsonResult renders v as indented JSON in a tool result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil
}

// resolve wraps android.ResolveSerial for the common serial-defaulting pattern.
func resolve(ctx context.Context, serial string) (string, error) {
	return android.ResolveSerial(ctx, serial)
}

// boolOr returns *p, or def when p is nil (for optional bool arguments).
func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}
