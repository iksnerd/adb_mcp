package tools

import (
	"context"
	"fmt"
	"time"

	"AndroidEmulatorMCP/internal/android"

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
	png, w, h, err := android.Screenshot(ctx, serial, maxDim)
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Screenshot of %s (%dx%d).", serial, w, h)},
			&mcp.ImageContent{Data: png, MIMEType: "image/png"},
		},
	}, nil
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
