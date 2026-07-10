package tools

import (
	"context"
	"strings"

	"AndroidEmulatorMCP/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type logcatArgs struct {
	serialArg
	Lines  int    `json:"lines,omitempty" jsonschema:"Number of recent lines to dump. Default 400."`
	Filter string `json:"filter,omitempty" jsonschema:"Case-insensitive substring to keep (e.g. an app tag or \"Exception\")."`
}

type startLogcatArgs struct {
	serialArg
	Clear *bool `json:"clear,omitempty" jsonschema:"Clear the logcat buffer before capturing. Default true."`
}

type stopLogcatArgs struct {
	serialArg
	Filter string `json:"filter,omitempty" jsonschema:"Case-insensitive substring to keep."`
}

type stopRecordArgs struct {
	serialArg
	LocalPath string `json:"local_path" jsonschema:"Local path to save the pulled mp4, e.g. /tmp/rec.mp4."`
}

// ---- Handlers ----

func logcat(ctx context.Context, in logcatArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.Logcat(ctx, serial, in.Lines, in.Filter)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return text("(no matching log lines)"), nil
	}
	return text("%s", out), nil
}

func startLogcatCapture(ctx context.Context, in startLogcatArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.StartLogcatCapture(ctx, serial, boolOr(in.Clear, true)); err != nil {
		return nil, err
	}
	return text("Logcat capture started for %s. Drive your flow, then stop_logcat_capture.", serial), nil
}

func stopLogcatCapture(ctx context.Context, in stopLogcatArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	out, err := android.StopLogcatCapture(serial, in.Filter)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return text("(capture stopped; no matching lines)"), nil
	}
	return text("%s", out), nil
}

func startScreenRecord(ctx context.Context, in serialArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	if err := android.StartScreenRecord(ctx, serial); err != nil {
		return nil, err
	}
	return text("Recording %s (max ~180s). Drive your flow, then stop_screen_record.", serial), nil
}

func stopScreenRecord(ctx context.Context, in stopRecordArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	path, err := android.StopScreenRecord(ctx, serial, in.LocalPath)
	if err != nil {
		return nil, err
	}
	return text("Saved recording to %s.", path), nil
}
