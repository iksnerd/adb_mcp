// Command android-emulator-mcp is an MCP server that drives Android
// emulators/devices over adb: boot AVDs, screenshot, read the UI hierarchy,
// tap/swipe/type, manage the device lock, read logcat, and control app
// lifecycle. It is the Android counterpart to XcodeBuildMCP.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"AndroidEmulatorMCP/internal/android"
	"AndroidEmulatorMCP/internal/guides"
	"AndroidEmulatorMCP/internal/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is overridable at build time via -ldflags "-X main.version=...".
// The Makefile injects the value from the VERSION file / git.
var version = "0.5.0"

func main() {
	log.SetFlags(0)
	log.SetPrefix("android-emulator-mcp: ")

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("android-emulator-mcp %s\n", version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Tear down any running logcat/screen-record capture sessions on exit so
	// their detached adb processes and temp files don't leak.
	defer android.StopAllCaptures()

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "android-emulator-mcp",
		Version: version,
	}, nil)

	tools.Register(srv)
	guides.Register(srv)

	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
		log.Fatalf("server error: %v", err)
	}
}
