package android

import (
	"context"
	"strconv"
	"strings"
)

// Logcat dumps the last `lines` log lines (adb logcat -d -t), optionally keeping
// only lines that contain filter (case-insensitive), and dropping chatty spam.
func Logcat(ctx context.Context, serial string, lines int, filter string) (string, error) {
	if lines <= 0 {
		lines = 400
	}
	out, err := runAdb(ctx, serial, "logcat", "-d", "-t", strconv.Itoa(lines))
	if err != nil {
		return "", err
	}
	needle := strings.ToLower(strings.TrimSpace(filter))
	var kept []string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "chatty") {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(line), needle) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n"), nil
}
