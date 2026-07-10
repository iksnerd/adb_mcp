package android

import (
	"context"
	"fmt"
	"os"
)

// PushFile copies a local file onto the device.
func PushFile(ctx context.Context, serial, localPath, devicePath string) (string, error) {
	if _, err := os.Stat(localPath); err != nil {
		return "", fmt.Errorf("local file not found: %s", localPath)
	}
	return runAdb(ctx, serial, "push", localPath, devicePath)
}

// PullFile copies a file off the device to a local path.
func PullFile(ctx context.Context, serial, devicePath, localPath string) (string, error) {
	return runAdb(ctx, serial, "pull", devicePath, localPath)
}
