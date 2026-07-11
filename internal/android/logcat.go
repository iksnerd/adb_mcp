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
		if isChattyNoise(line) {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(line), needle) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n"), nil
}

// isChattyNoise reports whether a logcat line is Android "chatty" dedup spam,
// emitted under the chatty tag as "uid=… expire N lines" / "… identical N
// lines". It requires both the chatty marker and the dedup phrasing so a real
// log line that merely mentions the word "chatty" (an app tag, package name, or
// message text) is not silently dropped.
func isChattyNoise(line string) bool {
	if !strings.Contains(line, "chatty") {
		return false
	}
	return strings.Contains(line, "expire ") || strings.Contains(line, "identical ")
}
