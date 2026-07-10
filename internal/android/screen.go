package android

import (
	"context"
	"fmt"
)

// Screenshot captures the screen with `exec-out screencap -p` (avoids the
// CRLF corruption of `shell screencap`) and downscales it so its largest
// dimension is at most maxDim. It returns the PNG bytes and their dimensions.
func Screenshot(ctx context.Context, serial string, maxDim int) (png []byte, w, h int, err error) {
	raw, err := runAdbBytes(ctx, serial, "exec-out", "screencap", "-p")
	if err != nil {
		return nil, 0, 0, err
	}
	if len(raw) == 0 {
		return nil, 0, 0, fmt.Errorf("empty screenshot from device %s", serial)
	}
	out, w, h := downscalePNG(raw, maxDim)
	return out, w, h, nil
}
