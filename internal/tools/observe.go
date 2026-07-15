package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/iksnerd/adb_mcp/internal/android"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Arguments ----

type screenshotArgs struct {
	serialArg
	MaxDim *int `json:"max_dim,omitempty" jsonschema:"Max width/height of the returned image in pixels. Omit for the default 760; pass 0 (or a negative) to disable downscaling and get the full-resolution image."`
}

type waitForTextArgs struct {
	serialArg
	Text     string `json:"text" jsonschema:"Text or content-description to wait for."`
	Partial  *bool  `json:"partial,omitempty" jsonschema:"Substring match instead of exact. Default true."`
	TimeoutS int    `json:"timeout_s,omitempty" jsonschema:"How long to wait, in seconds. Default 15."`
}

// ---- Handlers ----

func screenshot(ctx context.Context, in screenshotArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	maxDim := 760 // default when omitted; an explicit 0/negative disables downscaling
	if in.MaxDim != nil {
		maxDim = *in.MaxDim
	}
	cap, err := android.CaptureScreen(ctx, serial, maxDim)
	if err != nil {
		return nil, err
	}
	caption := fmt.Sprintf("Screenshot of %s (%dx%d).", serial, cap.Width, cap.Height)
	if cap.AllBlack {
		caption += " " + blackCaptureNote(cap)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: caption},
			&mcp.ImageContent{Data: cap.PNG, MIMEType: "image/png"},
		},
	}, nil
}

// blackCaptureNote explains an all-black frame and points at describe_ui, so
// the caller doesn't misread "black" as "screen asleep" when it isn't. It ends
// with a compact machine-readable status the model can branch on.
func blackCaptureNote(c android.ScreenCapture) string {
	var reason string
	switch {
	case c.SecureWindow:
		reason = "a FLAG_SECURE window is on screen — the OS blanks secure content (e.g. a native PIN pad) to black in screenshots"
	case c.ScreenOff:
		reason = "the display appears to be asleep/dozing"
	default:
		reason = fmt.Sprintf("the frame came back all black after %d attempt(s) — likely an intermittent capture glitch, not a real black screen", c.Attempts)
	}
	return fmt.Sprintf("WARNING: image is all black — %s. Fall back to describe_ui to read the UI (it works even when a screenshot is blanked). "+
		"status: {\"all_black\":true,\"secure_window\":%t,\"screen_off\":%t,\"attempts\":%d}",
		reason, c.SecureWindow, c.ScreenOff, c.Attempts)
}

func describeUI(ctx context.Context, in serialArg) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	elems, err := android.DescribeUI(ctx, serial)
	if err != nil {
		return nil, err
	}
	return jsonResult(elems)
}

func waitForText(ctx context.Context, in waitForTextArgs) (*mcp.CallToolResult, error) {
	serial, err := resolve(ctx, in.Serial)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(in.TimeoutS) * time.Second
	e, err := android.WaitForText(ctx, serial, in.Text, boolOr(in.Partial, true), timeout)
	if err != nil {
		return nil, err
	}
	return jsonResult(e)
}
