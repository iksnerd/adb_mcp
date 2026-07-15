package android

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Logcat dumps the last `lines` log lines (adb logcat -d -t) and applies
// LogFilter (substring/priority/tags), dropping chatty spam.
func Logcat(ctx context.Context, serial string, lines int, f LogFilter) (string, error) {
	if lines <= 0 {
		lines = 400
	}
	if err := f.validate(); err != nil {
		return "", err
	}
	out, err := runAdb(ctx, serial, "logcat", "-d", "-t", strconv.Itoa(lines))
	if err != nil {
		return "", err
	}
	return f.apply(out), nil
}

// LogFilter narrows a raw logcat dump. All set fields must match (AND); Tags
// itself is an OR across its entries. Zero value keeps everything (minus
// chatty noise, which is always stripped).
type LogFilter struct {
	// Substring is a case-insensitive substring to keep anywhere in the line.
	Substring string
	// Priority is a minimum level to keep: V, D, I, W, E, or F — matching
	// adb's own "*:E"-style filterspec (E keeps Error and Fatal). Empty means
	// no priority filtering.
	Priority string
	// Tags keeps only lines whose log tag contains one of these
	// (case-insensitive, OR'd). Empty means no tag filtering.
	Tags []string
}

// priorityRank orders Android log priorities low to high, matching adb's
// filterspec semantics: a minimum priority keeps that level and everything
// more severe.
var priorityRank = map[string]int{"V": 0, "D": 1, "I": 2, "W": 3, "E": 4, "F": 5}

// logLineRe matches the `-v time` / default dump header:
// "MM-DD HH:MM:SS.mmm  PID  TID X TAG   : message" — capturing the priority
// letter and tag (padding around the tag is trimmed by the caller).
var logLineRe = regexp.MustCompile(`^\S+\s+\S+\s+\d+\s+\d+\s+([VDIWEF])\s+([^:]*):`)

func (f LogFilter) validate() error {
	if f.Priority == "" {
		return nil
	}
	if _, ok := priorityRank[strings.ToUpper(strings.TrimSpace(f.Priority))]; !ok {
		return fmt.Errorf("priority must be one of V, D, I, W, E, F, got %q", f.Priority)
	}
	return nil
}

// apply filters raw multi-line logcat output, keeping lines that pass every
// set criterion. Lines whose header doesn't parse (rare: adb status lines, a
// wrapped continuation without its own header) are kept rather than dropped,
// since priority/tag filtering can't be evaluated for them.
func (f LogFilter) apply(raw string) string {
	needle := strings.ToLower(strings.TrimSpace(f.Substring))
	minRank, hasPriority := priorityRank[strings.ToUpper(strings.TrimSpace(f.Priority))]
	tags := make([]string, 0, len(f.Tags))
	for _, t := range f.Tags {
		if t = strings.ToLower(strings.TrimSpace(t)); t != "" {
			tags = append(tags, t)
		}
	}

	var kept []string
	for _, line := range strings.Split(raw, "\n") {
		if isChattyNoise(line) {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(line), needle) {
			continue
		}
		if (hasPriority || len(tags) > 0) && !matchesPriorityAndTags(line, minRank, hasPriority, tags) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func matchesPriorityAndTags(line string, minRank int, hasPriority bool, tags []string) bool {
	prio, tag, ok := parseLogLine(line)
	if !ok {
		return true // can't evaluate; don't drop
	}
	if hasPriority {
		if rank, known := priorityRank[prio]; known && rank < minRank {
			return false
		}
	}
	if len(tags) > 0 {
		tagLower := strings.ToLower(tag)
		matched := false
		for _, t := range tags {
			if strings.Contains(tagLower, t) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// parseLogLine extracts the priority letter and tag from one `-v time` /
// default-format logcat line.
func parseLogLine(line string) (priority, tag string, ok bool) {
	m := logLineRe.FindStringSubmatch(line)
	if m == nil {
		return "", "", false
	}
	return m[1], strings.TrimSpace(m[2]), true
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
