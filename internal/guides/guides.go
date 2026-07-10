// Package guides exposes the driving know-how — the "skill" behind this server —
// as MCP resources. A client can list and read them to learn the observe→act
// loop, native PIN/lock handling, and native crash triage, the same way it would
// consult a skill file. The content is embedded so the binary is self-contained.
package guides

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed getting-started.md
var gettingStarted string

//go:embed driving.md
var driving string

//go:embed pin-and-lock.md
var pinAndLock string

//go:embed crash-triage.md
var crashTriage string

type guide struct {
	uri, name, title, desc, body string
}

var all = []guide{
	{
		uri:   "android://guide/getting-started",
		name:  "getting-started",
		title: "Getting started: boot & connect",
		desc:  "How to list AVDs, boot an emulator, target a device with the serial argument, and run a first interaction.",
		body:  gettingStarted,
	},
	{
		uri:   "android://guide/driving",
		name:  "driving",
		title: "Driving a UI (observe→act loop)",
		desc:  "The core observe→locate→act→re-observe loop, the true-pixel coordinate rule, and the gotchas (overlays eating taps, keyboard covering buttons, settle delays) that waste turns.",
		body:  driving,
	},
	{
		uri:   "android://guide/pin-and-lock",
		name:  "pin-and-lock",
		title: "Native PIN pads & device lock",
		desc:  "Entering digits on native (non-IME) PIN pads, and setting/clearing a secure lock screen required by AndroidKeyStore / Keystore-backed crypto flows.",
		body:  pinAndLock,
	},
	{
		uri:   "android://guide/crash-triage",
		name:  "crash-triage",
		title: "Finding why a native call failed",
		desc:  "Using logcat to surface the real 'Caused by:' root cause hidden behind a generic UI error, plus app lifecycle tools for reproducing failures.",
		body:  crashTriage,
	},
}

// Register adds every guide to the server as a readable text/markdown resource.
func Register(s *mcp.Server) {
	for _, g := range all {
		g := g
		s.AddResource(&mcp.Resource{
			URI:         g.uri,
			Name:        g.name,
			Title:       g.title,
			Description: g.desc,
			MIMEType:    "text/markdown",
		}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			if req.Params.URI != g.uri {
				return nil, fmt.Errorf("unknown resource %q", req.Params.URI)
			}
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      g.uri,
					MIMEType: "text/markdown",
					Text:     g.body,
				}},
			}, nil
		})
	}
}
